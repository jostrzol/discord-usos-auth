package usos

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/dghubble/oauth1"
)

func usosURL(key string) string {
	var baseURL = "https://apps.usos.pw.edu.pl/"

	var urls = map[string]string{
		"":             "",
		"requestToken": "services/oauth/request_token",
		"authorize":    "services/oauth/authorize",
		"accessToken":  "services/oauth/access_token",
		"user":         "services/users/user?fields=%s",
		"groups":       "services/groups/user?fields=%s&active_terms=%s",
	}
	return baseURL + urls[key]
}

var config = oauth1.Config{
	ConsumerKey:    "774c544Rjd7R3hevEzkg",
	ConsumerSecret: "hFH6hFfEqJmbvHn7VcrPqfchKn357U6mErGN7F2F",
	CallbackURL:    "oob",
	Endpoint: oauth1.Endpoint{
		RequestTokenURL: usosURL("requestToken"),
		AuthorizeURL:    usosURL("authorize"),
		AccessTokenURL:  usosURL("accessToken"),
	},
}

// RequestToken represents oauth1 request token
type RequestToken struct {
	Token            string
	Secret           string
	AuthorizationURL *url.URL
}

// GetAccessToken returns an access token from the request token and verifier
func (rt *RequestToken) GetAccessToken(verifier string) (*oauth1.Token, error) {
	token, secret, err := config.AccessToken(rt.Token, rt.Secret, verifier)
	if err != nil {
		return nil, err
	}
	return oauth1.NewToken(token, secret), nil
}

// NewRequestToken returns an usos unauthorized
func NewRequestToken() (*RequestToken, error) {
	token, secret, err := config.RequestToken()
	if err != nil {
		return nil, err
	}
	authorizationURL, err := config.AuthorizationURL(token)
	if err != nil {
		return nil, err
	}
	return &RequestToken{token, secret, authorizationURL}, nil
}

func makeCall(client *http.Client, key string, a ...interface{}) ([]byte, error) {
	url := fmt.Sprintf(usosURL(key), a...)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}
