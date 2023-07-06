package twitter

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

var errorNotFoundOrNSFW = errors.New("http:not found or NSFW content")

func FetchTwitter(uri *url.URL) (*slack.Attachment, error) {
	path := uri.Path
	if path == "" {
		return nil, errors.New("does not contains path")
	}

	params := strings.Split(path, "/")
	if len(params) < 3 {
		return nil, errors.New("not tweet uri")
	}

	idStr := params[3]

	atch, err := fetchFromSyndication(idStr)

	if err == nil {
		return atch, nil
	}

	if !errors.Is(err, errorNotFoundOrNSFW) {
		return nil, err
	}

	return fetchFromAPI(idStr)
}

type urlShortenEntity struct {
	DisplayURL  string `json:"display_url"`
	ExpandedURL string `json:"expanded_url"`
	Indices     []int  `json:"indices"`
	URL         string `json:"url"`
}

func extractShortenURL(text string, urls []urlShortenEntity) string {

	// replace from tail
	sort.Slice(urls, func(i, j int) bool {
		return urls[i].Indices[0] > urls[j].Indices[0]
	})

	// indices are unit by rune.
	proceed := []rune(text)

	for _, it := range urls {

		swapped := append([]rune{}, proceed[:it.Indices[0]]...)

		link := fmt.Sprintf("<%s|%s>", it.ExpandedURL, it.DisplayURL)

		swapped = append(swapped, []rune(link)...)
		swapped = append(swapped, proceed[it.Indices[1]:]...)

		proceed = swapped
	}

	return string(proceed)
}

func getTweetBlock(tweet string, shortenEntities []urlShortenEntity) *slack.SectionBlock {
	txt := extractShortenURL(tweet, shortenEntities)
	return &slack.SectionBlock{
		Type: slack.MBTSection,
		Text: &slack.TextBlockObject{
			Type: "mrkdwn",
			Text: txt,
		},
	}
}

type userEntity struct {
	IDStr                string `json:"id_str"`
	Name                 string `json:"name"`
	ProfileImageURLHTTPS string `json:"profile_image_url_https"`
	ScreenName           string `json:"screen_name"`
}

func getUserBlock(user userEntity) *slack.ContextBlock {
	return &slack.ContextBlock{
		Type: slack.MBTContext,
		ContextElements: slack.ContextElements{
			Elements: []slack.MixedElement{
				slack.ImageBlockElement{
					Type:     slack.METImage,
					ImageURL: user.ProfileImageURLHTTPS,
					AltText:  user.ScreenName,
				},
				slack.TextBlockObject{
					Type: "mrkdwn",
					Text: fmt.Sprintf("<https://twitter.com/%s|*%s*> <https://twitter.com/%s|@%s>",
						user.ScreenName, user.Name,
						user.ScreenName, user.ScreenName,
					),
				},
			},
		},
	}
}

func getCreatedAtBlock(createdAt time.Time) *slack.ContextBlock {
	return &slack.ContextBlock{
		Type: slack.MBTContext,
		ContextElements: slack.ContextElements{
			Elements: []slack.MixedElement{
				slack.TextBlockObject{
					Type: "plain_text",
					Text: time.Time(createdAt).Local().Format(time.UnixDate),
				},
			},
		},
	}
}

type mediaEntity struct {
	ID            int64  `json:"id"`
	IDStr         string `json:"id_str"`
	Indices       []int  `json:"indices"`
	MediaURL      string `json:"media_url"`
	MediaURLHTTPS string `json:"media_url_https"`
	URL           string `json:"url"`
	DisplayURL    string `json:"display_url"`
	ExpandedURL   string `json:"expanded_url"`
	Type          string `json:"type"`
}
