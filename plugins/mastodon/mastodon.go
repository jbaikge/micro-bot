package mastodon

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jbaikge/micro-bot/irc"
	"github.com/mattn/go-mastodon"
	"golang.org/x/exp/slog"
	"golang.org/x/net/html"
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

const (
	StreamFederated = "federated"
	StreamLocal     = "local"
	StreamTimeline  = "timeline"
)

// https://modern.ircdocs.horse/formatting.html
var colors = []int{
	// Default colors in the first-16 space. Skipping the blues because they are
	// hard to see on a black background
	3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 14, 15,
	// Some more specific shades of colors
	// 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51
}

type Config struct {
	Stream       string
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

	var events chan mastodon.Event
	var err error
	switch m.config.Stream {
	case StreamFederated:
		events, err = client.StreamingPublic(ctx, false)
	case StreamLocal:
		events, err = client.StreamingPublic(ctx, true)
	case StreamTimeline:
		events, err = client.StreamingUser(ctx)
	default:
		slog.Error("unable to determine stream", fmt.Errorf("unknown stream type: %s", m.config.Stream))
		return
	}

	if err != nil {
		slog.Error("unable to get stream", err)
		return
	}

	m.client.Join(m.config.Channel, m.config.Password)

	// re := regexp.MustCompile(`<.*?>`)

	for e := range events {
		switch event := e.(type) {
		case *mastodon.UpdateEvent:
			slog.Debug("mastodon update", "url", event.Status.URL, "id", event.Status.Account.ID, "username", event.Status.Account.Username)
			// // Break up ending paragraph tags into newlines
			// content := strings.ReplaceAll(event.Status.Content, "</p>", "</p>\n")
			// // Strip HTML
			// content = re.ReplaceAllString(content, "")
			// // Break content up into lines
			// lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
			// Caclulate username color
			color := m.color(event.Status.Account)
			var text string
			if event.Status.Reblog != nil {
				text = textContent(event.Status.Reblog.Content)
			} else {
				text = textContent(event.Status.Content)
			}
			for _, line := range strings.Split(text, "\n") {
				message := fmt.Sprintf(
					"%s%02d%s%s %s",
					FormatColor,
					color,
					event.Status.Account.Username,
					FormatReset,
					line,
				)
				if len(message) > 500 {
					slog.Warn("message too long, not sending", "msg", message)
					continue
				}
				m.client.Privmsg(m.config.Channel, message)
			}
			link := fmt.Sprintf(
				"%s%02d%s%s \u00bb %s",
				FormatColor,
				color,
				event.Status.Account.Username,
				FormatReset,
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

func (m *Mastodon) color(account mastodon.Account) int {
	id, _ := strconv.ParseInt(string(account.ID), 10, 64)
	idx := id % int64(len(colors))
	return colors[idx]
}

// Shamelessly ripped from go-mastodon's main.go
func textContent(s string) string {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return s
	}
	var buf bytes.Buffer
	extractText(doc, &buf)
	return strings.TrimRight(buf.String(), "\n")
}

func extractText(node *html.Node, w *bytes.Buffer) {
	if node.Type == html.TextNode {
		if data := strings.Trim(node.Data, "\r\n"); data != "" {
			w.WriteString(data)
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, w)
	}
	if node.Type == html.ElementNode {
		if name := strings.ToLower(node.Data); name == "br" {
			w.WriteRune('\n')
		}
	}
}
