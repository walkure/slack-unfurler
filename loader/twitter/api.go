package twitter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

// authHeader contains Bearer Token
const authHeader = "Bearer AAAAAAAAAAAAAAAAAAAAAPYXBAAAAAAACLXUNDekMxqa8h%2F40K4moUkGsoc%3DTYfbDKbT3jJPCEVnMYqilB28NHfOPqkca3qaAxGfsyKCs0wRbw"
const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"

// csrfToken should be [a-f0-9]{32}
const csrfToken = "12345678901234567890123456789012"

var authTokenList = os.Getenv("TWITTER_AUTH_TOKENS_FROM_BROWSER")

func fetchFromAPI(idStr string) (*slack.Attachment, error) {

	if authTokenList == "" {
		return nil, errors.New("cannot get auth_token. disable apicall")
	}

	authTokens := strings.Split(authTokenList, ",")
	rand.Shuffle(len(authTokens), func(i, j int) { authTokens[i], authTokens[j] = authTokens[j], authTokens[i] })

	for _, authToken := range authTokens {
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/1.1/statuses/show/%s.json?tweet_mode=extended&cards_platform=Web-12&include_cards=1&include_reply_count=1&include_user_entities=0", idStr), nil)
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
			defer res.Body.Close()
			return extractStatus(res.Body)
		case http.StatusUnauthorized:
			continue
		case http.StatusNotFound:
			return nil, errors.New("status not found or deleted")
		default:
			return nil, fmt.Errorf("http/api status error:%d", res.StatusCode)
		}
	}
	return nil, errors.New("no valid auth_token remained")
}

func extractStatus(responseBody io.Reader) (*slack.Attachment, error) {
	status := statusEntity{}

	if err := json.NewDecoder(responseBody).Decode(&status); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}

	blocks := []slack.Block{
		getUserBlock(status.User),
		getTweetBlock(status.FullText, append(status.Entities.Media, status.Entities.Urls...)),
	}

	for _, p := range status.ExtendedEntities.Media {
		blocks = append(blocks, &slack.ImageBlock{
			Type:     slack.MBTImage,
			ImageURL: p.MediaURLHTTPS,
			AltText:  p.DisplayURL,
		})
	}

	blocks = append(blocks, getCreatedAtBlock(time.Time(status.CreatedAt)))

	return &slack.Attachment{Blocks: slack.Blocks{BlockSet: blocks}}, nil
}

type rubyDateTime time.Time

func (rdt *rubyDateTime) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}

	tt, err := time.Parse(`"`+time.RubyDate+`"`, string(data))
	*rdt = rubyDateTime(tt)

	return err
}

type statusInternalEntity struct {
	CreatedAt rubyDateTime `json:"created_at"`
	ID        int64        `json:"id"`
	IDStr     string       `json:"id_str"`
	FullText  string       `json:"full_text"`
	Entities  struct {
		Urls  []urlShortenEntity `json:"urls"`
		Media []urlShortenEntity `json:"media"`
	} `json:"entities"`
	ExtendedEntities struct {
		Media []mediaEntity `json:"media"`
	} `json:"extended_entities"`
	User       userEntity `json:"user"`
	SelfThread struct {
		ID    int64  `json:"id"`
		IDStr string `json:"id_str"`
	} `json:"self_thread"`
	RetweetCount          int    `json:"retweet_count"`
	FavoriteCount         int    `json:"favorite_count"`
	ReplyCount            int    `json:"reply_count"`
	QuotedStatusID        int64  `json:"quoted_status_id"`
	QuotedStatusIDStr     string `json:"quoted_status_id_str"`
	QuotedStatusPermalink struct {
		URL      string `json:"url"`
		Expanded string `json:"expanded"`
		Display  string `json:"display"`
	} `json:"quoted_status_permalink"`
}

type statusEntity struct {
	statusInternalEntity
	QuotedStatus statusInternalEntity `json:"quoted_status"`
}
