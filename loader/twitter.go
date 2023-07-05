package loader

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

var errorNotFoundOrNSFW = errors.New("http:not found or NSFW content")

func fetchTwitter(uri *url.URL) (attachment *slack.Attachment, err error) {
	path := uri.Path
	if path == "" {
		return nil, errors.New("does not contains path")
	}

	params := strings.Split(path, "/")
	if len(params) < 3 {
		return nil, errors.New("not tweet uri")
	}

	id_str := params[3]

	target := fmt.Sprintf("https://cdn.syndication.twimg.com/tweet-result?id=%s&lang=ja", id_str)

	resp, err := http.Get(target)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, errorNotFoundOrNSFW
	case http.StatusTooManyRequests:
		return nil, errors.New("http:too many requests")
	}

	return expandTwitter(resp.Body)
}

type tweetEntity struct {
	Typename          string    `json:"__typename"`
	Lang              string    `json:"lang"`
	FavoriteCount     int       `json:"favorite_count"`
	PossiblySensitive bool      `json:"possibly_sensitive"`
	CreatedAt         time.Time `json:"created_at"`
	DisplayTextRange  []int     `json:"display_text_range"`
	Entities          struct {
		Urls []struct {
			DisplayURL  string `json:"display_url"`
			ExpandedURL string `json:"expanded_url"`
			Indices     []int  `json:"indices"`
			URL         string `json:"url"`
		} `json:"urls"`
		Media []struct {
			DisplayURL  string `json:"display_url"`
			ExpandedURL string `json:"expanded_url"`
			Indices     []int  `json:"indices"`
			URL         string `json:"url"`
		} `json:"media"`
	} `json:"entities"`
	IDStr string `json:"id_str"`
	Text  string `json:"text"`
	User  struct {
		IDStr                string `json:"id_str"`
		Name                 string `json:"name"`
		ProfileImageURLHTTPS string `json:"profile_image_url_https"`
		ScreenName           string `json:"screen_name"`
		Verified             bool   `json:"verified"`
		VerifiedType         string `json:"verified_type"`
		IsBlueVerified       bool   `json:"is_blue_verified"`
	} `json:"user"`
	EditControl struct {
		EditTweetIds       []string `json:"edit_tweet_ids"`
		EditableUntilMsecs string   `json:"editable_until_msecs"`
		IsEditEligible     bool     `json:"is_edit_eligible"`
		EditsRemaining     string   `json:"edits_remaining"`
	} `json:"edit_control"`
	MediaDetails []struct {
		AdditionalMediaInfo struct {
			Description string `json:"description"`
			Embeddable  bool   `json:"embeddable"`
			Title       string `json:"title"`
		} `json:"additional_media_info"`
		DisplayURL           string `json:"display_url"`
		ExpandedURL          string `json:"expanded_url"`
		ExtMediaAvailability struct {
			Status string `json:"status"`
		} `json:"ext_media_availability"`
		Indices       []int  `json:"indices"`
		MediaURLHTTPS string `json:"media_url_https"`
		Type          string `json:"type"`
		URL           string `json:"url"`
		VideoInfo     struct {
			AspectRatio    []int `json:"aspect_ratio"`
			DurationMillis int   `json:"duration_millis"`
			Variants       []struct {
				Bitrate     int    `json:"bitrate,omitempty"`
				ContentType string `json:"content_type"`
				URL         string `json:"url"`
			} `json:"variants"`
		} `json:"video_info"`
	} `json:"mediaDetails"`
	Photos []struct {
		ExpandedURL string `json:"expandedUrl"`
		URL         string `json:"url"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
	} `json:"photos"`
	Video struct {
		ContentType       string `json:"contentType"`
		DurationMs        int    `json:"durationMs"`
		MediaAvailability struct {
			Status string `json:"status"`
		} `json:"mediaAvailability"`
		Poster   string `json:"poster"`
		Variants []struct {
			Type string `json:"type"`
			Src  string `json:"src"`
		} `json:"variants"`
		VideoID struct {
			Type string `json:"type"`
			ID   string `json:"id"`
		} `json:"videoId"`
		ViewCount int `json:"viewCount"`
	} `json:"video"`
	ConversationCount int    `json:"conversation_count"`
	NewsActionType    string `json:"news_action_type"`
	IsEdited          bool   `json:"isEdited"`
	IsStaleEdit       bool   `json:"isStaleEdit"`
}
type tw struct {
	tweetEntity
	QuotedTweet tweetEntity `json:"quoted_tweet"`
}

func expandTwitter(body io.Reader) (attachment *slack.Attachment, err error) {
	tweet := &tw{}
	if err := json.NewDecoder(body).Decode(tweet); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}

	blocks := []slack.Block{
		&slack.ContextBlock{
			Type: slack.MBTContext,
			ContextElements: slack.ContextElements{
				Elements: []slack.MixedElement{
					slack.ImageBlockElement{
						Type:     slack.METImage,
						ImageURL: tweet.User.ProfileImageURLHTTPS,
						AltText:  tweet.User.ScreenName,
					},
					slack.TextBlockObject{
						Type: "mrkdwn",
						Text: fmt.Sprintf("<https://twitter.com/%s|*%s*> <https://twitter.com/%s|@%s>",
							tweet.User.ScreenName, tweet.User.Name,
							tweet.User.ScreenName, tweet.User.ScreenName,
						),
					},
				},
			},
		},
		&slack.SectionBlock{
			Type: slack.MBTSection,
			Text: &slack.TextBlockObject{
				Type: "plain_text",
				Text: tweet.Text,
			},
		},
	}

	for _, p := range tweet.Photos {
		blocks = append(blocks, &slack.ImageBlock{
			Type:     slack.MBTImage,
			ImageURL: p.URL,
			AltText:  p.URL,
		})
	}

	blocks = append(blocks, &slack.ContextBlock{
		Type: slack.MBTContext,
		ContextElements: slack.ContextElements{
			Elements: []slack.MixedElement{
				slack.TextBlockObject{
					Type: "plain_text",
					Text: tweet.CreatedAt.Local().Format(time.UnixDate),
				},
			},
		},
	})

	attachment = &slack.Attachment{Blocks: slack.Blocks{BlockSet: blocks}}
	return attachment, nil
}
