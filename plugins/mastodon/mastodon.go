package mastodon

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/jbaikge/micro-bot/irc"
	"github.com/mattn/go-mastodon"
	"golang.org/x/exp/slog"
)

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
			content := re.ReplaceAllString(event.Status.Content, "")
			message := fmt.Sprintf("<%s> %s", event.Status.Account.Username, content)
			if len(message) > 500 {
				slog.Warn("message too long, not sending", "msg", message)
				continue
			}
			m.client.Privmsg(m.config.Channel, message)
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
