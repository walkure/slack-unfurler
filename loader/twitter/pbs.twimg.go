package twitter

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"

	"github.com/slack-go/slack"
)

func MakePbsBlock(target *url.URL) (attachment *slack.Attachment, err error) {

	targetUnescaped := html.UnescapeString(target.String())

	res, err := http.Head(targetUnescaped)
	if err != nil {
		return nil, fmt.Errorf("head err: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http not ok:%d", res.StatusCode)
	}

	contentType := res.Header.Get("content-type")
	if !strings.HasPrefix(contentType, "image/") {
		return nil, fmt.Errorf("content-type not image :%s", contentType)
	}

	return &slack.Attachment{
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.ImageBlock{
					Type:     slack.MBTImage,
					ImageURL: targetUnescaped,
					AltText:  targetUnescaped,
				},
			},
		},
	}, nil
}
