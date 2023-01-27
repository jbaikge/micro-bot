package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/jbaikge/micro-bot/config"
	"github.com/jbaikge/micro-bot/irc"
	"github.com/jbaikge/micro-bot/plugins/mastodon"
	"github.com/jbaikge/micro-bot/plugins/twitter"
	"golang.org/x/exp/slog"
)

var (
	configFile string
	debugMode  bool
)

func init() {
	configFile = filepath.Join(os.Getenv("HOME"), ".config", "micro-bot", "micro-bot.json")
	flag.StringVar(&configFile, "config", configFile, "Location of config file")
	flag.BoolVar(&debugMode, "debug", debugMode, "Enable debug messages")
}

func main() {
	flag.Parse()

	// Parse configuration
	parsed, err := config.ParseConfig(configFile)
	if err != nil {
		slog.Error("unable to parse config", err)
	}

	// Set up logging
	opts := slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if debugMode {
		opts.Level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(opts.NewTextHandler(os.Stderr)))

	// Neat way to set up a context that listens for interrupts
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	// Connect to IRC server
	client, err := irc.NewClient(ctx, parsed.IRC)
	if err != nil {
		slog.Error("failed to initialize client", err)
	}
	defer client.Disconnect()

	// Set up plugins
	plugins := make([]irc.Plugin, 0, len(parsed.Mastodon)+len(parsed.Twitter))
	for _, cfg := range parsed.Mastodon {
		plugins = append(plugins, mastodon.NewMastodon(client, cfg))
	}
	for _, cfg := range parsed.Twitter {
		plugin, err := twitter.NewTwitter(client, cfg)
		if err != nil {
			slog.Error("error creating twitter client", err, "config", cfg)
			return
		}
		plugins = append(plugins, plugin)
	}

	// Run plugins
	for _, plugin := range plugins {
		go plugin.Run(ctx)
	}

	// When an interrupt occurs, shut everything down cleanly
	<-ctx.Done()
}
