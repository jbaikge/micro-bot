package irc

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slog"
)

func (c *Client) Join(channel string, key string) (err error) {
	if channel[0] == '#' {
		channel = channel[1:]
	}
	msg := fmt.Sprintf("JOIN #%s", channel)
	logMsg := msg
	if key != "" {
		msg += " " + key
		logMsg += " [REDACTED]"
	}
	slog.Debug("sending JOIN", "msg", logMsg)
	fmt.Fprint(c, msg)
	return
}

func (c *Client) Nick(nick string) (err error) {
	if strings.IndexByte(nick, ' ') > -1 {
		return fmt.Errorf("nick cannot contain spaces: %s", nick)
	}
	msg := fmt.Sprintf("NICK %s", nick)
	slog.Debug("sending NICK", "msg", msg)
	_, err = fmt.Fprint(c, msg)
	return
}

func (c *Client) Pass(pass string) (err error) {
	slog.Debug("sending PASS", "msg", "PASS [REDACTED]")
	_, err = fmt.Fprintf(c, "PASS %s", pass)
	return
}

func (c *Client) Ping(token string) (err error) {
	msg := fmt.Sprintf("PING %s", token)
	slog.Debug("sending PING", "msg", msg)
	_, err = fmt.Fprint(c, msg)
	return
}

func (c *Client) Pong(token string) (err error) {
	msg := fmt.Sprintf("PONG %s", token)
	slog.Debug("sending PONG", "msg", msg)
	_, err = fmt.Fprint(c, msg)
	return
}

func (c *Client) Privmsg(target string, message string) (err error) {
	msg := fmt.Sprintf("PRIVMSG %s :%s", target, message)
	slog.Debug("sending PRIVMSG", "msg", msg)
	_, err = fmt.Fprint(c, msg)
	return
}

func (c *Client) Quit(message string) (err error) {
	msg := fmt.Sprintf("QUIT :%s", message)
	slog.Debug("sending QUIT", "msg", msg)
	_, err = fmt.Fprint(c, msg)
	return
}

func (c *Client) User(username string, realname string) (err error) {
	if strings.IndexByte(username, ' ') > -1 {
		return fmt.Errorf("username cannot contain spaces: %s", username)
	}
	msg := fmt.Sprintf("USER %s 0 * :%s", username, realname)
	slog.Debug("sending USER", "msg", msg)
	_, err = fmt.Fprint(c, msg)
	return
}
