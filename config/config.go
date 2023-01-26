package config

import (
	"encoding/json"
	"os"

	"github.com/jbaikge/micro-bot/irc"
	"github.com/jbaikge/micro-bot/plugins/mastodon"
	"github.com/jbaikge/micro-bot/plugins/twitter"
)

type Config struct {
	IRC      irc.Config
	Mastodon []mastodon.Config
	Twitter  []twitter.Config
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
