package irc

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"golang.org/x/exp/slog"
)

const MaxMessageLength = 500

const (
	FormatBold   = "\x02"
	FormatItalic = "\x1d"
	FormatColor  = "\x03"
	FormatReset  = "\x0f"
)

// https://modern.ircdocs.horse/formatting.html
var Colors = []int{
	// Default colors in the first-16 space. Skipping the blues because they are
	// hard to see on a black background
	3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 14, 15,
	// Some more specific shades of colors
	// 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51
}

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
	ticker     *time.Ticker
}

// Creates a new client and connects to the IRC server
func NewClient(ctx context.Context, config Config) (client *Client, err error) {
	client = &Client{
		ctx:    ctx,
		config: config,
		ticker: time.NewTicker(250 * time.Millisecond),
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
	for i := 0; i < len(p)/MaxMessageLength+1; i++ {
		<-c.ticker.C
		low := i * MaxMessageLength
		high := low + MaxMessageLength
		if high > len(p) {
			high = len(p)
		}
		fmt.Fprint(c.connection, string(p[low:high]), "\r\n")
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
