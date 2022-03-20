package common

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/walkure/slack-unfurler/loader"
)

func CallbackEventHandler(ctx context.Context, api *slack.Client, eventsAPIEvent slackevents.EventsAPIEvent) error {
	innerEvent := eventsAPIEvent.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.LinkSharedEvent:
		return handleLinkSharedEvent(ctx, api, ev)
	default:
		return fmt.Errorf("unsupported Callback Event received: %T", ev)
	}
}

func handleLinkSharedEvent(ctx context.Context, api *slack.Client, ev *slackevents.LinkSharedEvent) error {

	data := make(map[string]slack.Attachment)
	for _, target := range ev.Links {
		blocks, err := loader.GetUnfluredBlocks(target.URL)
		if err != nil {
			fmt.Printf("cannot unfurl %s : %v", target.URL, err)
			continue
		}
		data[target.URL] = slack.Attachment{
			Blocks: slack.Blocks{BlockSet: *blocks},
		}
	}
	_, _, err := postMessageWithRetry(ctx, api, ev.Channel, slack.MsgOptionUnfurl(ev.MessageTimeStamp, data))

	return err
}
