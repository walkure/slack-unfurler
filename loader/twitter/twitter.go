package twitter

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

var unfurlVideo = os.Getenv("UNFURL_TWITTER_VIDEO")

func FetchTwitter(uri *url.URL) (*slack.Attachment, error) {
	path := uri.Path
	if path == "" {
		return nil, errors.New("does not contains path")
	}

	params := strings.Split(path, "/")

	if len(params) >= 4 && params[3] != "" {
		return fetchTweet(params[3])
	}

	if len(params) >= 2 && params[1] != "" {
		return fetchTwitterProfile(params[1])
	}

	return nil, errors.New("not tweet uri")

}

func fetchTweet(idStr string) (*slack.Attachment, error) {
	atch, err := fetchFromAPI(idStr)
	if err == nil {
		return atch, nil
	}

	fmt.Printf("fetchFromAPI failed: %v\n", err)

	return fetchFromSyndication(idStr)
}

type urlShortenEntity struct {
	DisplayURL  string `json:"display_url"`
	ExpandedURL string `json:"expanded_url"`
	Indices     []int  `json:"indices"`
	URL         string `json:"url"`
}

func filterShortenURLs(urls []urlShortenEntity) []urlShortenEntity {
	var filtered []urlShortenEntity
	urlMap := make(map[string]bool)
	for _, it := range urls {
		if _, exists := urlMap[it.URL]; !exists {
			filtered = append(filtered, it)
			urlMap[it.URL] = true
		}
	}
	return filtered
}

func extractShortenURL(text string, targetUrls []urlShortenEntity) string {

	urls := filterShortenURLs(targetUrls)
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
			Type:     "mrkdwn",
			Text:     txt,
			Verbatim: true,
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
					Verbatim: true,
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
	VideoInfo     struct {
		AspectRatio    []int `json:"aspect_ratio"`
		DurationMillis int   `json:"duration_millis"`
		Variants       []struct {
			Bitrate     int    `json:"bitrate,omitempty"`
			ContentType string `json:"content_type"`
			URL         string `json:"url"`
		} `json:"variants"`
	} `json:"video_info"`
}

func getMediaBlocks(media mediaEntity) slack.Block {
	if media.Type == "photo" {
		return slack.ImageBlock{
			Type:     slack.MBTImage,
			ImageURL: media.MediaURLHTTPS,
			AltText:  media.DisplayURL,
		}
	}

	if (media.Type == "video" || media.Type == "animated_gif") && unfurlVideo != "" {
		videoURL := ""
		bitrate := 0
		for _, v := range media.VideoInfo.Variants {
			if v.ContentType != "video/mp4" {
				continue
			}

			// use best bitrate
			if bitrate <= v.Bitrate {
				bitrate = v.Bitrate
				videoURL = v.URL
				//fmt.Printf("use:%d %s\n", bitrate, videoURL)
			}
		}
		return slack.VideoBlock{
			Type:         slack.MBTVideo,
			VideoURL:     videoURL,
			ThumbnailURL: media.MediaURLHTTPS,
			AltText:      media.DisplayURL,
			TitleURL:     videoURL,
			Title: &slack.TextBlockObject{
				Type: "plain_text",
				Text: media.DisplayURL,
			},
		}
	}

	return nil
}
