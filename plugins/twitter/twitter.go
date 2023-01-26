package twitter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jbaikge/micro-bot/irc"
	"github.com/mrjones/oauth"
	"golang.org/x/exp/slog"
)

// https://developer.twitter.com/en/docs/twitter-api/rate-limits
// The product portal shows 180 requests / 15 minutes
// RateLimit is definied as:
// <window (in min)> * <seconds> / <Max requests per window> * <Second duration>
const APIRateLimit = 15 * 60 / 15 * time.Second

type Config struct {
	Channel           string
	Password          string
	ApiKey            string
	ApiKeySecret      string
	AccessToken       string
	AccessTokenSecret string
}

type Tweet struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	AuthorID string `json:"author_id"`
	User     User
}

type Twitter struct {
	config      Config
	irc         *irc.Client
	http        *http.Client
	userId      string
	lastTweetId string
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

func NewTwitter(client *irc.Client, config Config) (twitter *Twitter, err error) {
	twitter = &Twitter{
		irc:    client,
		config: config,
	}

	// Set up Twitter OAuth goofiness
	serviceProvider := oauth.ServiceProvider{
		RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
		AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
		AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
	}
	consumer := oauth.NewConsumer(config.ApiKey, config.ApiKeySecret, serviceProvider)
	consumer.Debug(false)

	accessToken := &oauth.AccessToken{
		Token:  config.AccessToken,
		Secret: config.AccessTokenSecret,
	}

	twitter.http, err = consumer.MakeHttpClient(accessToken)
	if err != nil {
		return nil, err
	}

	return
}

func (t *Twitter) Run(ctx context.Context) {
	var err error

	t.userId, err = t.getUserId()
	if err != nil {
		slog.Error("unable to get user id", err)
		return
	}

	t.irc.Join(t.config.Channel, t.config.Password)

	ticker := time.NewTicker(APIRateLimit)
	for range ticker.C {
		tweets, err := t.getLatest()
		if err != nil {
			slog.Error("getLatest", err)
			continue
		}
		for _, tweet := range tweets {
			accountColor := color(tweet.User)
			for _, line := range strings.Split(tweet.Text, "\n") {
				message := fmt.Sprintf(
					"%s%02d%s%s %s",
					irc.FormatColor,
					accountColor,
					tweet.User.Username,
					irc.FormatReset,
					line,
				)
				t.irc.Privmsg(t.config.Channel, message)
			}
		}
	}
}

func (t *Twitter) getUserId() (id string, err error) {
	response, err := t.http.Get("https://api.twitter.com/2/users/me")
	if err != nil {
		return
	}
	defer response.Body.Close()

	user := struct {
		Data struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"data"`
	}{}
	err = json.NewDecoder(response.Body).Decode(&user)
	if err != nil {
		return
	}
	slog.Debug("me", "user", user.Data)
	return user.Data.ID, nil
}

func (t *Twitter) getLatest() (tweets []Tweet, err error) {
	apiUrl, err := url.Parse("https://api.twitter.com/2/users/" + t.userId + "/timelines/reverse_chronological")
	if err != nil {
		return
	}

	query := apiUrl.Query()
	query.Add("expansions", "author_id")
	if t.lastTweetId == "" {
		query.Add("max_results", "1")
	} else {
		query.Add("since_id", t.lastTweetId)
	}
	apiUrl.RawQuery = query.Encode()
	slog.Debug("constructed api url", "url", apiUrl.String())

	response, err := t.http.Get(apiUrl.String())
	if err != nil {
		return
	}
	defer response.Body.Close()

	// {
	// 	"data": [
	// 		{
	// 			"text": "RT @ShmooConPuzzle: The link to our 2023 puzzle solution slides. #ShmooCon  https://t.co/vmiFe0fthr",
	// 			"author_id": "15803690",
	// 			"edit_history_tweet_ids":["1618618308844191744"],
	// 			"id":"1618618308844191744"
	// 		}
	// 	],
	// 	"includes": {
	// 		"users": [
	// 			{
	// 				"id":"15803690",
	// 				"name":"ShmooCon",
	// 				"username":"shmoocon"
	// 			}
	// 		]
	// 	},
	// 	"meta": {
	// 		"next_token":"7140dibdnow9c7btw450r26sg5t2jgvsj5wnxa9rgf584",
	// 		"result_count":1,
	// 		"newest_id":"1618618308844191744",
	// 		"oldest_id":"1618618308844191744"
	// 	}
	// }
	data := struct {
		Tweets   []Tweet `json:"data"`
		Includes struct {
			Users []User `json:"users"`
		} `json:"includes"`
		Meta struct {
			NewestID string `json:"newest_id"`
			OldestID string `json:"oldest_id"`
		} `json:"meta"`
	}{}
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return
	}
	fmt.Printf("%+v\n", data)

	// No sense in doing anything if there is nothing new
	if len(data.Tweets) == 0 {
		return
	}

	// Set the latest ID for the next run
	t.lastTweetId = data.Meta.NewestID

	tweets = make([]Tweet, 0, len(data.Tweets))
	for _, tweet := range data.Tweets {
		for _, user := range data.Includes.Users {
			if user.ID == tweet.AuthorID {
				tweet.User = user
				break
			}
		}
		tweets = append(tweets, tweet)
	}

	return
}

func color(user User) int {
	id, _ := strconv.ParseInt(user.ID, 10, 64)
	idx := id % int64(len(irc.Colors))
	return irc.Colors[idx]
}
