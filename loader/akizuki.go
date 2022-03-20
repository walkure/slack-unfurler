package loader

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/slack-go/slack"
)

func fetchAkizuki(uri *url.URL) (blocks *[]slack.Block, err error) {

	doc, err := fetchDocument(uri)
	if err != nil {
		return nil, fmt.Errorf("akizuki:fetch err:%w", err)
	}

	// get Title
	title := doc.Find("#maincontents > div:nth-child(6)").Contents().Get(0).Data
	fmt.Printf("title:%s\n", title)
	//title := doc.Find("#maincontents > table > tbody > tr:nth-child(1) > td > table > tbody > tr > td:nth-child(2) > table > tbody > tr:nth-child(1) > td > h6").Text()
	if title == "" {
		return nil, errors.New("akizuki:cannot get image uri")
	}

	// get Image URL
	href, ok := doc.Find("#imglink > img").Attr("src")
	if !ok {
		return nil, errors.New("akizuki:cannot get image uri")
	}
	imageUrl, err := resolveRelativeURI(uri, href)
	if err != nil {
		return nil, fmt.Errorf("akizuki:image_url err:%w", err)
	}

	// get description.
	desc := ""
	doc.Find("#maincontents > table > tbody > tr:nth-child(1) > td > table > tbody > tr > td:nth-child(2) > table > tbody > tr:nth-child(3) > td").Contents().EachWithBreak(func(i int, s *goquery.Selection) bool {
		if !s.Is("br") {
			desc = s.Text()
			return false
		}
		return true
	})
	if desc == "" {
		return nil, errors.New("akizuki:cannot get description")
	}

	// get price
	price := trimDescription(doc.Find("#maincontents > div:nth-child(6) > table > tbody > tr > td:nth-child(1)").Text())
	if price == "" {
		return nil, errors.New("akizuki:cannot get price")
	}

	blocks = &[]slack.Block{
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

	return blocks, nil
}
