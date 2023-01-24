package mastodon

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"regexp"
	"strings"

	"github.com/jbaikge/micro-bot/irc"
	"github.com/mattn/go-mastodon"
	"golang.org/x/exp/slog"
)

const (
	FormatBold   = "\x02"
	FormatItalic = "\x1d"
	FormatColor  = "\x03"
	FormatReset  = "\x0f"
)

// https://www.w3schools.com/charsets/ref_utf_box.asp
const (
	DrawStart    = "\u250f"
	DrawContinue = "\u2503"
	DrawEnd      = "\u2503"
)

// https://modern.ircdocs.horse/formatting.html
var colors = []int{40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51}

type Config struct {
	Channel      string
	Password     string
	Server       string
	ClientID     string
	ClientSecret string
	AccessToken  string
}

type Mastodon struct {
	client *irc.Client
	config Config
}

func NewMastodon(client *irc.Client, config Config) *Mastodon {
	if config.Channel[0] != '#' {
		config.Channel = "#" + config.Channel
	}
	return &Mastodon{
		client: client,
		config: config,
	}
}

func (m *Mastodon) Run(ctx context.Context) {
	config := &mastodon.Config{
		Server:       m.config.Server,
		ClientID:     m.config.ClientID,
		ClientSecret: m.config.ClientSecret,
		AccessToken:  m.config.AccessToken,
	}
	client := mastodon.NewClient(config)
	events, err := client.StreamingPublic(ctx, false)
	if err != nil {
		slog.Error("unable to get stream", err)
		return
	}

	m.client.Join(m.config.Channel, m.config.Password)

	re := regexp.MustCompile(`<.*?>`)

	for e := range events {
		switch event := e.(type) {
		case *mastodon.UpdateEvent:
			slog.Debug("mastodon update", "url", event.Status.URL, "username", event.Status.Account.Username)
			content := strings.ReplaceAll(event.Status.Content, "</p>", "</p>\n")
			content = re.ReplaceAllString(content, "")
			lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
			sep := "\u250f"
			for _, line := range lines {
				message := fmt.Sprintf(
					"%s%02d%s%s %s %s",
					FormatColor,
					m.color(event.Status.Account.Username),
					event.Status.Account.Username,
					FormatReset,
					sep,
					line,
				)
				if len(message) > 500 {
					slog.Warn("message too long, not sending", "msg", message)
					continue
				}
				m.client.Privmsg(m.config.Channel, message)
				sep = "\u2503"
			}
			link := fmt.Sprintf(
				"%s%02d%s%s %s %s",
				FormatColor,
				m.color(event.Status.Account.Username),
				event.Status.Account.Username,
				FormatReset,
				"\u2517",
				event.Status.URL,
			)
			m.client.Privmsg(m.config.Channel, link)
		case *mastodon.NotificationEvent:
			slog.Debug("mastodon notification", "type", event.Notification.Type, "username", event.Notification.Account.Username)
		case *mastodon.ErrorEvent:
			slog.Error("mastodon failure", errors.New(event.Error()))
			return
		default:
			// Probably a delete event or something
			slog.Info("not sure how to handle this event type", "event", event)
		}
	}
}

func (m *Mastodon) color(username string) int {
	sum := crc32.ChecksumIEEE([]byte(username))
	idx := int(sum) % len(colors)
	return colors[idx]
}
