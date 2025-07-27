package twitter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// authHeader contains Bearer Token
const authHeader = "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"

// csrfToken should be [a-f0-9]{32}
const csrfToken = "12345678901234567890123456789012"

var authTokenList = os.Getenv("TWITTER_AUTH_TOKENS_FROM_BROWSER")
var guestApiID = os.Getenv("TWITTER_GUEST_API_ID")
var loginApiID = os.Getenv("TWITTER_LOGIN_API_ID")

func invokeGraphQL(apiName string, apiID string, queryFactory func(graphQuery url.Values), response any) error {
	if authTokenList == "" {
		return errors.New("cannot get auth_token. disable apicall")
	}

	if apiID == "" {
		return errors.New("cannot get api_id. disable apicall")
	}

	authTokens := strings.Split(authTokenList, ",")
	rand.Shuffle(len(authTokens), func(i, j int) { authTokens[i], authTokens[j] = authTokens[j], authTokens[i] })

	endpoint, err := url.Parse(fmt.Sprintf("https://twitter.com/i/api/graphql/%s/%s", apiID, apiName))
	if err != nil {
		return fmt.Errorf("url parse error: %w", err)
	}

	q := endpoint.Query()
	queryFactory(q)
	endpoint.RawQuery = q.Encode()

	for _, authToken := range authTokens {
		req, _ := http.NewRequest("GET", endpoint.String(), nil)
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Cookie", fmt.Sprintf("auth_token=%s; ct0=%s; ", authToken, csrfToken))
		req.Header.Set("x-twitter-active-user", "yes")
		req.Header.Set("x-twitter-auth-type", "OAuth2Session")
		req.Header.Set("x-twitter-client-language", "en")
		req.Header.Set("x-csrf-token", csrfToken)
		req.Header.Set("User-Agent", userAgent)

		res, err := http.DefaultClient.Do(req)

		if err != nil {
			return fmt.Errorf("http/api transport error: %w", err)
		}

		if res.StatusCode == http.StatusOK {
			defer func() {
				io.Copy(io.Discard, res.Body)
				res.Body.Close()
			}()
			//tr := io.TeeReader(res.Body, os.Stdout)
			return json.NewDecoder(res.Body).Decode(response)
		}

		io.Copy(io.Discard, res.Body)
		res.Body.Close()
		fmt.Printf("HTTP status:%d Token:%s\n", res.StatusCode, authToken)

		switch res.StatusCode {
		case http.StatusUnauthorized:
		case http.StatusTooManyRequests:
			continue
		case http.StatusNotFound:
			return errors.New("status not found or deleted")
		default:
			return fmt.Errorf("http/api status error:%d", res.StatusCode)
		}
	}
	return errors.New("no valid auth_token remained")

}
