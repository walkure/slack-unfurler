package twitter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

	id_str := params[3]

	resp, err := fetchFromSyndication(id_str)

	if errors.Is(err, errorNotFoundOrNSFW) {
		resp, err = fetchFromAPI(id_str)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	defer resp.Close()

	return expandTwitter(resp)
}

type twJSONTime time.Time

func (tt *twJSONTime) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}

	var err error
	t, err := time.Parse(`"`+time.RFC3339+`"`, string(data))

	if err != nil {
		t, err = time.Parse(`"`+time.RubyDate+`"`, string(data))
	}

	*tt = twJSONTime(t)

	return err
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

func expandTwitter(body io.Reader) (attachment *slack.Attachment, err error) {
	tweet := &tw{}
	if err := json.NewDecoder(body).Decode(tweet); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}

	txt := tweet.Text
	if tweet.Text == "" {
		txt = tweet.FullText
	}

	txt = extractShortenURL(txt, append(tweet.Entities.Media, tweet.Entities.Urls...))

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
				Type: "mrkdwn",
				Text: txt,
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
					Text: time.Time(tweet.CreatedAt).Local().Format(time.UnixDate),
				},
			},
		},
	})

	attachment = &slack.Attachment{Blocks: slack.Blocks{BlockSet: blocks}}
	return attachment, nil
}
