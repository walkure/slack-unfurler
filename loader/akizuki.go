package loader

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/slack-go/slack"
)

func fetchAkizuki(uri *url.URL) (attachment *slack.Attachment, err error) {

	doc, err := fetchDocument(uri)
	if err != nil {
		return nil, fmt.Errorf("akizuki:fetch err:%w", err)
	}

	// get Title
	title := doc.Find("body > div.wrapper > div.pane-contents > div > main > div > div.pane-goods-header > div.block-goods-name > h1").Text()
	if title == "" {
		return nil, errors.New("akizuki:cannot get title")
	}

	// get Image URL (lazy loaded)
	href, ok := doc.Find("#gallery > div.block-src-l > a > figure > img").Attr("data-src")
	if !ok {
		return nil, errors.New("akizuki:cannot get image uri")
	}
	imageUrl, err := resolveRelativeURI(uri, href)

	if err != nil {
		return nil, fmt.Errorf("akizuki:image_url err:%w", err)
	}

	// get description.
	desc := doc.Find("body > div.wrapper > div.pane-contents > div > main > div > div.pane-goods-center > div.block-goods-overview").Text()
	if desc == "" {
		return nil, errors.New("akizuki:cannot get description")
	}

	// get price (include tax)
	price := trimDescription(doc.Find("#SalesArea > div > div:nth-child(1) > div.block-goods-price--price.price.js-enhanced-ecommerce-goods-price").Text())
	if price == "" {
		return nil, errors.New("akizuki:cannot get price")
	}

	blocks := []slack.Block{
		slack.SectionBlock{
			Type: slack.MBTSection,
			Text: &slack.TextBlockObject{
				Type: "plain_text",
				Text: title,
			},
		},
		slack.DividerBlock{
			Type: slack.MBTDivider,
		},
		slack.SectionBlock{
			Type: slack.MBTSection,
			Text: &slack.TextBlockObject{
				Type: "plain_text",
				Text: price,
			},
		},
		slack.DividerBlock{
			Type: slack.MBTDivider,
		},
		slack.SectionBlock{
			Type: slack.MBTSection,
			Text: &slack.TextBlockObject{
				Type: "plain_text",
				Text: trimDescription(desc),
			},
		},
		&slack.ImageBlock{
			Type:     slack.MBTImage,
			ImageURL: imageUrl,
			AltText:  imageUrl,
		},
	}

	return &slack.Attachment{
		Blocks: slack.Blocks{BlockSet: blocks},
	}, nil
}
