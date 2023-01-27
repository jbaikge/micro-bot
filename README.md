# Micro Bot

An IRC bot to read micro-blogging feeds. Currently there is support for Mastodon and Twitter.

## Installation

Make sure Go is installed, at least version 1.18, then:

```
go get github.com/jbaikge/micro-bot
```

## Configuration

By default, `micro-bot` looks in `$HOME/.config/micro-bot/micro-bot.json` for the configuration. It will not automatically create the directory or the file if you want the configuration stored there. Alternatively, there is a `-config` flag that will take any path to a JSON file.

The main configuration is split into three sections: IRC, Mastodon, and Twitter. A sample configuration is available in `micro-bot.sample.json`. Configurations for Mastodon and Twitter are arrays to allow for multiple usernames, or in the case of Mastodon: servers or streams.

### IRC Configuration

```json
{
  "server":  "localhost:6667",
  "nick":     "bot",
  "password": "",
  "realname": "Micro Bot",
  "username": "micro-bot"
}
```

It is _highly_ recommended to run a local IRC server to connect the bot to, then add the local server to your IRC client. I recommend and tested the bot with [ergo](https://github.com/ergochat/ergo). I am not responsible for any server bans related to using this bot on a large network.

* `server` must include the port
* `nick` cannot contain spaces or strange characters (I tried the mu character (μ) and got booted)
* `password` is optional or can be blank
* `realname` can be anything
* `username` same rules as `nick` above

### Mastodon Configuration

```json
{
  "stream":        "timeline",
  "channel":       "#m_username",
  "password":      "password",
  "server":        "mastodon.server.domain",
  "client_id":     "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "client_secret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "access_token":  "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}
```

* `stream` may be one of:
  * `federated` - your server's federated feed
  * `local` - your server's public feed
  * `timeline` - posts from people you follow
* `channel` is the IRC channel to join, the leading pound sign is required
* `password` is the password used to access the IRC channel, it may be blank
* `server` is your server's domain name, please include the protocol (https)

Set up `client_id`, `client_secret`, and `access_token`:
1. Login to your Mastodon server, preferrably on a computer
2. Click Preferences in the lower right
3. Click Development in the left menu
4. Click New Application in the upper right
5. Set the Application Name to Micro Bot. The URL can be blank or point to this repository. Leave the Redirect URI alone
6. Uncheck all permissions except *read*
7. Click Submit
8. The three values should be on the next page

### Twitter Configuration

```json
{
  "channel":             "#t_username",
  "password":            "password",
  "api_key":             "xxxxxxxxxxxxxxxxxxxxxxxxx",
  "api_key_secret":      "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "access_token":        "123456789-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "access_token_secret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}
```

* `channel` is the IRC channel to join, the leading pound sign is required
* `password` is the password used to access the IRC channel, it may be blank

Set up `api_key`, `api_key_secret`, `access_token`, and `access_tokne_secret`:
1. Set up a Twitter developer account
2. Create a new application for the v2 API
3. Set up each of the keys on the application portal
4. Copy those keys into the correct spots in the configuration

## Running

Once all the configuration is established, running is as simple as

```
micro-bot
```

If the configuration is in a different spot:

```
micro-bot -config /path/to/config.json
```
