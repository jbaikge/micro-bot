package twitter

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/jbaikge/micro-bot/irc"
	"golang.org/x/exp/slog"
)

const APIBaseUrl = "https://api.twitter.com/2"

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
	BearerToken       string
	AccessToken       string
	AccessTokenSecret string
}

type Tweet struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Twitter struct {
	client      *irc.Client
	config      Config
	userId      string
	lastTweetId string
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

func NewTwitter(client *irc.Client, config Config) *Twitter {
	tw := &Twitter{
		client: client,
		config: config,
	}
	return tw
}

func (t *Twitter) Run(ctx context.Context) {
	t.client.Join(t.config.Channel, t.config.Password)

	user, err := t.getUser()
	if err != nil {
		slog.Error("problem getting my user ID", err)
	}
	slog.Info("got user", "user", user)

	ticker := time.NewTicker(APIRateLimit)
	for range ticker.C {
		slog.Debug("twitter tick", "now", time.Now().Format(time.Kitchen))
	}
}

// Access token === Token === resulting oauth_token
// Access token secret === Token Secret === resulting oauth_token_secret
func (t *Twitter) apiGet(apiUrl *url.URL) (response *http.Response, err error) {
	request, err := http.NewRequest(http.MethodGet, apiUrl.String(), nil)
	if err != nil {
		return
	}
	// request.Header.Add("Authorization", "Bearer "+t.config.BearerToken)
	nonce := md5.Sum([]byte(time.Now().Format(time.RFC3339Nano)))
	timestamp := time.Now().Unix()
	signature := ""
	value := fmt.Sprintf(
		`OAuth oauth_consumer_key="%s", oauth_nonce="%s", oauth_signature="%s", oauth_signature_method="HMAC-SHA1", oauth_timestamp="%d", oauth_token="%s", oauth_version="1.0"`,
		t.config.ApiKey,
		nonce,
		signature,
		timestamp,
		t.config.AccessToken,
	)
	request.Header.Add("Authorization", value)
	return http.DefaultClient.Do(request)
}

func (t *Twitter) getUser() (user User, err error) {
	rawUrl, err := url.JoinPath(APIBaseUrl, "users/me")
	if err != nil {
		return
	}
	apiUrl, err := url.Parse(rawUrl)
	if err != nil {
		return
	}
	response, err := t.apiGet(apiUrl)
	if err != nil {
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	slog.Debug("response from twitter", "body", string(body))
	return

	result := struct {
		Data User `json:"data"`
	}{}
	if err = json.NewDecoder(response.Body).Decode(&result); err != nil {
		return
	}
	return result.Data, nil
}

// func (t *Twitter) getLatest() (tweets []Tweet, err error) {
// 	// rawUrl := fmt.Sprintf("https://api.twitter.com/2/users/%s/timelines/reverse_chronological", t.userId)
// 	// apiUrl := url.Parse()
// 	// http.NewRequest(http.MethodGet, apiUrl, nil)
// 	return
// }

// curl "https://api.twitter.com/2/tweets?ids=1261326399320715264,1278347468690915330" \
//   -H "Authorization: Bearer AAAAAAAAAAAAAAAAAAAAAFnz2wAAAAAAxTmQbp%2BIHDtAhTBbyNJon%2BA72K4%3DeIaigY0QBrv6Rp8KZQQLOTpo9ubw5Jt?WRE8avbi"
