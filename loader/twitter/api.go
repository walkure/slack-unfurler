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
	"time"

	"github.com/slack-go/slack"
)

// authHeader contains Bearer Token
const authHeader = "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"

// csrfToken should be [a-f0-9]{32}
const csrfToken = "12345678901234567890123456789012"

var authTokenList = os.Getenv("TWITTER_AUTH_TOKENS_FROM_BROWSER")
var apiID = os.Getenv("TWITTER_API_ID")

func fetchFromAPI(idStr string) (*slack.Attachment, error) {

	if authTokenList == "" {
		return nil, errors.New("cannot get auth_token. disable apicall")
	}

	if apiID == "" {
		return nil, errors.New("cannot get api_id. disable apicall")
	}

	authTokens := strings.Split(authTokenList, ",")
	rand.Shuffle(len(authTokens), func(i, j int) { authTokens[i], authTokens[j] = authTokens[j], authTokens[i] })

	endpoint, err := url.Parse(fmt.Sprintf("https://twitter.com/i/api/graphql/%s/TweetResultByRestId", apiID))
	if err != nil {
		return nil, fmt.Errorf("url parse error: %w", err)
	}

	q := endpoint.Query()
	q.Set("variables", fmt.Sprintf(`{"tweetId":"%s","withCommunity":false,"includePromotedContent":false,"withVoice":false}`, idStr))
	q.Set("features", `{"creator_subscriptions_tweet_preview_api_enabled":true,"tweetypie_unmention_optimization_enabled":true,"responsive_web_edit_tweet_api_enabled":true,"graphql_is_translatable_rweb_tweet_is_translatable_enabled":true,"view_counts_everywhere_api_enabled":true,"longform_notetweets_consumption_enabled":true,"responsive_web_twitter_article_tweet_consumption_enabled":false,"tweet_awards_web_tipping_enabled":false,"freedom_of_speech_not_reach_fetch_enabled":true,"standardized_nudges_misinfo":true,"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled":true,"longform_notetweets_rich_text_read_enabled":true,"longform_notetweets_inline_media_enabled":true,"responsive_web_graphql_exclude_directive_enabled":true,"verified_phone_label_enabled":false,"responsive_web_media_download_video_enabled":false,"responsive_web_graphql_skip_user_profile_image_extensions_enabled":false,"responsive_web_graphql_timeline_navigation_enabled":true,"responsive_web_enhance_cards_enabled":false}`)
	q.Set("fieldToggles", `{"withArticleRichContentState":false,"withAuxiliaryUserLabels":false}`)
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
			return nil, fmt.Errorf("http/api transport error: %w", err)
		}

		if res.StatusCode == http.StatusOK {
			defer func() {
				io.Copy(io.Discard, res.Body)
				res.Body.Close()
			}()
			return extractStatus(res.Body)
		}

		io.Copy(io.Discard, res.Body)
		res.Body.Close()
		fmt.Printf("HTTP status:%d Token:%s\n", res.StatusCode, authToken)

		switch res.StatusCode {
		case http.StatusUnauthorized:
		case http.StatusTooManyRequests:
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
	statusContainer := statusContainer{}

	//tr := io.TeeReader(responseBody, os.Stdout)

	if err := json.NewDecoder(responseBody).Decode(&statusContainer); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}
	result := statusContainer.Data.TweetResult.Result
	tweet := result.Legacy
	user := result.Core.UserResults.Result.Legacy

	if result.RestID == "" {
		tweet = result.Tweet.Legacy
		user = result.Tweet.Core.UserResults.Result.Legacy
	}

	blocks := []slack.Block{
		getUserBlock(user),
		getTweetBlock(tweet.FullText, append(tweet.Entities.Media, tweet.Entities.Urls...)),
	}

	for _, p := range tweet.ExtendedEntities.Media {
		if block := getMediaBlocks(p); block != nil {
			blocks = append(blocks, block)
		}
	}

	blocks = append(blocks, getCreatedAtBlock(time.Time(tweet.CreatedAt)))

	if tweet.QuotedStatusIDStr != "" {

		qtresult := result.QuotedStatusResult.Result
		qtstatus := qtresult.Legacy
		qtuser := qtresult.Core.UserResults.Result.Legacy
		if qtresult.RestID == "" {
			qtstatus = qtresult.Tweet.Legacy
			qtuser = qtresult.Tweet.Core.UserResults.Result.Legacy
		}

		blocks = append(blocks,
			getUserBlock(qtuser),
			getTweetBlock(qtstatus.FullText,
				append(qtstatus.Entities.Media,
					qtstatus.Entities.Urls...)),
		)

		for _, p := range qtstatus.ExtendedEntities.Media {
			blocks = append(blocks, &slack.ImageBlock{
				Type:     slack.MBTImage,
				ImageURL: p.MediaURLHTTPS,
				AltText:  p.DisplayURL,
			})
		}

		blocks = append(blocks, getCreatedAtBlock(time.Time(qtstatus.CreatedAt)))
	}

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

type statusLegacyEntity struct {
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
	SelfThread struct {
		ID    int64  `json:"id"`
		IDStr string `json:"id_str"`
	} `json:"self_thread"`
	RetweetCount          int    `json:"retweet_count"`
	FavoriteCount         int    `json:"favorite_count"`
	ReplyCount            int    `json:"reply_count"`
	QuotedStatusIDStr     string `json:"quoted_status_id_str"`
	QuotedStatusPermalink struct {
		URL      string `json:"url"`
		Expanded string `json:"expanded"`
		Display  string `json:"display"`
	} `json:"quoted_status_permalink"`
}

type statusResultEntity struct {
	Typename string `json:"__typename"`
	statusResultCommonEntity
	QuotedRefResult struct {
		Result struct {
			Typename string `json:"__typename"`
			RestID   string `json:"rest_id"`
		} `json:"result"`
	} `json:"quotedRefResult"`
	Tweet statusResultCommonEntity `json:"tweet"`
}

type statusResultCommonEntity struct {
	RestID string             `json:"rest_id"`
	Legacy statusLegacyEntity `json:"legacy"`
	Core   struct {
		UserResults struct {
			Result struct {
				Typename       string     `json:"__typename"`
				ID             string     `json:"id"`
				RestID         string     `json:"rest_id"`
				IsBlueVerified bool       `json:"is_blue_verified"`
				Legacy         userEntity `json:"legacy"`
			} `json:"result"`
		} `json:"user_results"`
	} `json:"core"`
}

type statusResultWrapper struct {
	statusResultEntity
	QuotedStatusResult struct {
		Result statusResultEntity `json:"result"`
	} `json:"quoted_status_result"`
}

type statusContainer struct {
	Data struct {
		TweetResult struct {
			Result statusResultWrapper `json:"result"`
		} `json:"tweetResult"`
	} `json:"data"`
}
