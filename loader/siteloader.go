package loader

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/saintfish/chardet"
	"github.com/slack-go/slack"
	"golang.org/x/net/html/charset"
)

func GetUnfluredBlocks(target string) (blocks *[]slack.Block, err error) {
	uri, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(target, "https://akizukidenshi.com/catalog/g/g") {
		return fetchAkizuki(uri)
	}
	return nil, fmt.Errorf("%s not supported site", target)

}

func trimDescription(desc string) string {

	normalizedDesc := strings.NewReplacer(
		"\r\n", "\n",
		"\r", "\n",
	).Replace(desc)

	var sb strings.Builder
	for _, line := range strings.Split(normalizedDesc, "\n") {
		sb.WriteString(strings.TrimSpace(line))
	}

	return sb.String()
}

func resolveRelativeURI(baseUri *url.URL, relative string) (string, error) {
	relativeUri, err := baseUri.Parse(relative)
	if err != nil {
		return "", fmt.Errorf("relative:%w", err)
	}

	return relativeUri.String(), nil
}

func fetchDocument(target *url.URL) (*goquery.Document, error) {

	// HTTP Get
	res, err := http.Get(target.String())
	if err != nil {
		return nil, fmt.Errorf("HTTP/GET error:%w", err)
	}
	defer res.Body.Close()

	// Read
	bytesRead, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read error:%w", err)
	}

	// detect charset
	detector := chardet.NewTextDetector()
	deetctResult, err := detector.DetectBest(bytesRead)
	if err != nil {
		return nil, fmt.Errorf("charset detect error:%w", err)
	}

	// convert charset
	bytesReader := bytes.NewReader(bytesRead)
	reader, err := charset.NewReaderLabel(deetctResult.Charset, bytesReader)
	if err != nil {
		return nil, fmt.Errorf("charset convert error:%w", err)
	}

	// create document
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("cannot create goquery document:%w", err)
	}

	return doc, nil
}
