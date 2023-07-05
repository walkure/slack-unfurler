package twitter

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const authHeader = "Bearer AAAAAAAAAAAAAAAAAAAAAPYXBAAAAAAACLXUNDekMxqa8h%2F40K4moUkGsoc%3DTYfbDKbT3jJPCEVnMYqilB28NHfOPqkca3qaAxGfsyKCs0wRbw"
const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"

// csrfToken should be [a-f0-9]{32}
const csrfToken = "12345678901234567890123456789012"

var authTokens = os.Getenv("TWITTER_AUTH_TOKENS_FROM_BROWSER")

func fetchFromAPI(id_str string) (io.ReadCloser, error) {

	for _, authToken := range strings.Split(authTokens, ",") {
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/1.1/statuses/show/%s.json?tweet_mode=extended&cards_platform=Web-12&include_cards=1&include_reply_count=1&include_user_entities=0", id_str), nil)
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Cookie", fmt.Sprintf("auth_token=%s; ct0=%s; ", authToken, csrfToken))
		req.Header.Set("x-twitter-active-user", "yes")
		req.Header.Set("x-twitter-auth-type", "OAuth2Session")
		req.Header.Set("x-twitter-client-language", "en")
		req.Header.Set("x-csrf-token", csrfToken)
		req.Header.Set("User-Agent", userAgent)

		res, err := http.DefaultClient.Do(req)

		if err != nil {
			return nil, fmt.Errorf("http/api transport error: %w", err)
		}

		fmt.Printf("HTTP status:%d Token:%s\n", res.StatusCode, authToken)

		switch res.StatusCode {
		case http.StatusOK:
			return res.Body, nil
		case http.StatusUnauthorized:
			continue
		case http.StatusNotFound:
			return nil, errors.New("tweet not found or deleted")
		default:
			return nil, fmt.Errorf("http/api status error:%d", res.StatusCode)
		}
	}
	return nil, errors.New("no valid auth_token remained")
}
