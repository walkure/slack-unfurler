package common

import (
	"context"
	"time"

	"github.com/slack-go/slack"
)

func postMessageWithRetry(ctx context.Context, api *slack.Client, channelID string, options ...slack.MsgOption) (string, string, error) {
	for {
		respChannel, respTimestamp, err := api.PostMessageContext(ctx, channelID, options...)
		if err == nil {
			return respChannel, respTimestamp, nil
		} else if rateLimitedError, ok := err.(*slack.RateLimitedError); ok {
			select {
			case <-ctx.Done():
				return "", "", ctx.Err()
			case <-time.After(rateLimitedError.RetryAfter):
				continue
			}
		} else {
			return "", "", err
		}
	}
}
