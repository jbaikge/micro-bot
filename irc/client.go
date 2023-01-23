package irc

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"strings"

	"golang.org/x/exp/slog"
)

var _ io.Writer = &Client{}

type Config struct {
	Server   string
	Nick     string
	Username string
	Realname string
	Password string
}

// Client is based mostly on Zeebo's IRC implementation, just modified for my
// needs: https://github.com/zeebo/irc/
type Client struct {
	config     Config
	connection *net.TCPConn
	ctx        context.Context
}

// Creates a new client and connects to the IRC server
func NewClient(ctx context.Context, config Config) (client *Client, err error) {
	client = &Client{
		ctx:    ctx,
		config: config,
	}

	if client.connection, err = client.connect(); err != nil {
		slog.Error("failed to connect", err)
		return
	}
	slog.Info("connected")

	go client.readLoop()

	if err = client.login(); err != nil {
		slog.Error("failed to login", err)
		return
	}
	slog.Info("logged in", "nick", client.config.Nick)

	return
}

func (c *Client) Disconnect() {
	c.Quit("Disconnecting")
	c.connection.Close()
}

func (c *Client) Write(p []byte) (n int, err error) {
	scanner := bufio.NewScanner(bytes.NewReader(p))
	for scanner.Scan() {
		_, err = fmt.Fprint(c.connection, scanner.Text(), "\r\n")
		if err != nil {
			return
		}
	}
	return len(p), nil
}

func (c *Client) connect() (conn *net.TCPConn, err error) {
	addr, err := net.ResolveTCPAddr("tcp", c.config.Server)
	if err != nil {
		return
	}
	slog.Debug("resolved server", "from", c.config.Server, "to", addr.String())

	conn, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		return
	}
	slog.Debug("connected")

	return
}

func (c *Client) login() (err error) {
	if c.config.Password != "" {
		c.Pass(c.config.Password)
	}
	c.Nick(c.config.Nick)
	c.User(c.config.Username, c.config.Realname)
	return nil
}

func (c *Client) readLoop() {
	scanner := bufio.NewScanner(c.connection)
	for scanner.Scan() {
		message := scanner.Text()
		fmt.Println(message)
		parts := strings.Fields(message)
		switch strings.ToUpper(parts[0]) {
		case "PING":
			c.Pong(parts[1])
			continue
		case "ERROR":
			errMsg := message[6:]
			if errMsg[0] == ':' {
				errMsg = errMsg[1:]
			}
			slog.Error("server error", fmt.Errorf(errMsg))
			c.Disconnect()
			return
		}
	}
}
