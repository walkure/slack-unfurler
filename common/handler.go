package common

import (
	"context"
	"fmt"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/walkure/slack-unfurler/loader"
)

func CallbackEventHandler(ctx context.Context, api *slack.Client, eventsAPIEvent slackevents.EventsAPIEvent, cache *lru.Cache[string, *slack.Attachment]) error {
	innerEvent := eventsAPIEvent.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.LinkSharedEvent:
		return handleLinkSharedEvent(ctx, api, ev, cache)
	default:
		return fmt.Errorf("unsupported Callback Event received: %T", ev)
	}
}

func handleLinkSharedEvent(ctx context.Context, api *slack.Client, ev *slackevents.LinkSharedEvent, cache *lru.Cache[string, *slack.Attachment]) error {

	data := make(map[string]slack.Attachment)
	for _, target := range ev.Links {
		attachment, err := loader.GetUnfluredAttachment(ctx, target.Domain, target.URL, cache)
		if err != nil {
			fmt.Printf("cannot unfurl %s : %v\n", target.URL, err)
			continue
		}
		data[target.URL] = *attachment
	}
	_, _, err := postMessageWithRetry(ctx, api, ev.Channel, slack.MsgOptionUnfurl(ev.MessageTimeStamp, data))

	if err != nil {
		urls := make([]string, 0, len(data))
		for _, k := range ev.Links {
			urls = append(urls, k.URL)
		}
		fmt.Printf("unfurl failure[%s]: %v\n", urls, err)
	}

	return err
}
