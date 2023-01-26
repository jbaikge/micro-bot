package config

import (
	"encoding/json"
	"os"

	"github.com/jbaikge/micro-bot/irc"
	"github.com/jbaikge/micro-bot/plugins/mastodon"
	"github.com/jbaikge/micro-bot/plugins/twitter"
)

type Config struct {
	IRC      irc.Config        `json:"irc"`
	Mastodon []mastodon.Config `json:"mastodon"`
	Twitter  []twitter.Config  `json:"twitter"`
}

func ParseConfig(filename string) (c Config, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&c)
	return
}
