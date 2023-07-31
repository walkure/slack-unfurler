package twitter

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/slack-go/slack"
)

func fetchFromSyndication(idStr string) (*slack.Attachment, error) {
	target := fmt.Sprintf("https://cdn.syndication.twimg.com/tweet-result?id=%s&lang=ja", idStr)

	resp, err := http.Get(target)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, errors.New("http:not found or NSFW tweet")
	case http.StatusTooManyRequests:
		return nil, errors.New("http:too many requests")
	}
	defer resp.Body.Close()

	tweet := tweetEntity{}
	if err := json.NewDecoder(resp.Body).Decode(&tweet); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}

	blocks := []slack.Block{
		getUserBlock(tweet.User),
		getTweetBlock(tweet.Text, append(tweet.Entities.Media, tweet.Entities.Urls...)),
	}

	for _, p := range tweet.MediaDetails {
		if block := getMediaBlocks(p); block != nil {
			blocks = append(blocks, block)
		}
	}

	blocks = append(blocks, getCreatedAtBlock(tweet.CreatedAt))

	if tweet.QuotedTweet.IDStr != "" {
		blocks = append(blocks,
			getUserBlock(tweet.QuotedTweet.User),
			getTweetBlock(tweet.QuotedTweet.Text,
				append(tweet.QuotedTweet.Entities.Media,
					tweet.QuotedTweet.Entities.Urls...)),
		)

		for _, p := range tweet.QuotedTweet.Photos {
			blocks = append(blocks, &slack.ImageBlock{
				Type:     slack.MBTImage,
				ImageURL: p.URL,
				AltText:  p.URL,
			})
		}

		blocks = append(blocks, getCreatedAtBlock(tweet.QuotedTweet.CreatedAt))
	}

	return &slack.Attachment{Blocks: slack.Blocks{BlockSet: blocks}}, nil
}

type tweetInternalEntity struct {
	CreatedAt time.Time `json:"created_at"`
	Entities  struct {
		Urls  []urlShortenEntity `json:"urls"`
		Media []urlShortenEntity `json:"media"`
	} `json:"entities"`
	IDStr        string        `json:"id_str"`
	Text         string        `json:"text"`
	User         userEntity    `json:"user"`
	MediaDetails []mediaEntity `json:"mediaDetails"`
	Photos       []struct {
		ExpandedURL string `json:"expandedUrl"`
		URL         string `json:"url"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
	} `json:"photos"`
	IsEdited      bool `json:"isEdited"`
	IsStaleEdit   bool `json:"isStaleEdit"`
	FavoriteCount int  `json:"favorite_count"`
}
type tweetEntity struct {
	tweetInternalEntity
	QuotedTweet tweetInternalEntity `json:"quoted_tweet"`
}
