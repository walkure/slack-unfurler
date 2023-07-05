package twitter

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

type tweetEntity struct {
	Typename          string     `json:"__typename"`
	Lang              string     `json:"lang"`
	PossiblySensitive bool       `json:"possibly_sensitive"`
	CreatedAt         twJSONTime `json:"created_at"`
	DisplayTextRange  []int      `json:"display_text_range"`
	Entities          struct {
		Urls  []urlShortenEntity `json:"urls"`
		Media []urlShortenEntity `json:"media"`
	} `json:"entities"`
	IDStr    string `json:"id_str"`
	Text     string `json:"text"`
	FullText string `json:"full_text"`
	User     struct {
		IDStr                string `json:"id_str"`
		Name                 string `json:"name"`
		ProfileImageURLHTTPS string `json:"profile_image_url_https"`
		ScreenName           string `json:"screen_name"`
		Verified             bool   `json:"verified"`
	} `json:"user"`
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

func fetchFromSyndication(id_str string) (io.ReadCloser, error) {
	target := fmt.Sprintf("https://cdn.syndication.twimg.com/tweet-result?id=%s&lang=ja", id_str)

	resp, err := http.Get(target)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, errorNotFoundOrNSFW
	case http.StatusTooManyRequests:
		return nil, errors.New("http:too many requests")
	}

	return resp.Body, nil
}
