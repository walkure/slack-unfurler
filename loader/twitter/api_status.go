package twitter

import (
	"fmt"
	"net/url"
	"time"

	"github.com/slack-go/slack"
)

func FetchTweetStatus(idStr string) (*StatusResultWrapper, error) {

	statusContainer := statusContainer{}

	if err := invokeGraphQL("TweetResultByRestId", guestApiID, func(graphQuery url.Values) {
		graphQuery.Set("variables", fmt.Sprintf(`{"tweetId":"%s","withCommunity":false,"includePromotedContent":false,"withVoice":false}`, idStr))
		graphQuery.Set("features", `{"creator_subscriptions_tweet_preview_api_enabled":true,"tweetypie_unmention_optimization_enabled":true,"responsive_web_edit_tweet_api_enabled":true,"graphql_is_translatable_rweb_tweet_is_translatable_enabled":true,"view_counts_everywhere_api_enabled":true,"longform_notetweets_consumption_enabled":true,"responsive_web_twitter_article_tweet_consumption_enabled":false,"tweet_awards_web_tipping_enabled":false,"freedom_of_speech_not_reach_fetch_enabled":true,"standardized_nudges_misinfo":true,"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled":true,"longform_notetweets_rich_text_read_enabled":true,"longform_notetweets_inline_media_enabled":true,"responsive_web_graphql_exclude_directive_enabled":true,"verified_phone_label_enabled":false,"responsive_web_media_download_video_enabled":false,"responsive_web_graphql_skip_user_profile_image_extensions_enabled":false,"responsive_web_graphql_timeline_navigation_enabled":true,"responsive_web_enhance_cards_enabled":false}`)
		graphQuery.Set("fieldToggles", `{"withArticleRichContentState":false,"withAuxiliaryUserLabels":false}`)
	}, &statusContainer); err != nil {
		return nil, fmt.Errorf("fetch tweet from API: %w", err)
	}

	return &statusContainer.Data.TweetResult.Result, nil
}

func extendTweetByAPI(idStr string) (*slack.Attachment, error) {

	result, err := FetchTweetStatus(idStr)
	if err != nil {
		return nil, fmt.Errorf("fetch tweet by API: %w", err)
	}

	return extractSlackStatus(result)
}

func extractSlackStatus(result *StatusResultWrapper) (*slack.Attachment, error) {

	legacyTweet := result.Legacy
	noteTweet := result.NoteTweet.NoteTweetResults.Result
	user := result.Core.UserResults.Result.Legacy
	qtResult := result.QuotedStatusResult.Result

	if result.RestID == "" {
		legacyTweet = result.Tweet.Legacy
		noteTweet = result.Tweet.NoteTweet.NoteTweetResults.Result
		user = result.Tweet.Core.UserResults.Result.Legacy
		if result.Tweet.QuotedStatusResult.Result != nil {
			qtResult = *result.Tweet.QuotedStatusResult.Result
		}
	}

	var tweetText string
	var entities []urlShortenEntity

	if noteTweet.ID != "" {
		tweetText = noteTweet.Text
		// note tweet has no explicit screen_name in conversation
		entities = noteTweet.EntitySet.getShortenURLs("", "")
	} else {
		tweetText = legacyTweet.FullText
		entities = legacyTweet.Entities.getShortenURLs(legacyTweet.ConversationIDStr, legacyTweet.InReplyToUserIDStr)
	}

	blocks := []slack.Block{
		getUserBlock(user),
		getTweetBlock(tweetText, entities),
	}
	for _, p := range legacyTweet.ExtendedEntities.Media {
		if block := getMediaBlocks(p); block != nil {
			blocks = append(blocks, block)
		}
	}
	blocks = append(blocks, getCreatedAtBlock(time.Time(legacyTweet.CreatedAt)))

	if legacyTweet.QuotedStatusIDStr != "" {
		qtLegacy := qtResult.Legacy
		qtNote := qtResult.NoteTweet.NoteTweetResults.Result
		qtuser := qtResult.Core.UserResults.Result.Legacy
		if qtResult.RestID == "" {
			qtLegacy = qtResult.Tweet.Legacy
			qtuser = qtResult.Tweet.Core.UserResults.Result.Legacy
			qtNote = qtResult.Tweet.NoteTweet.NoteTweetResults.Result
		}

		if qtLegacy.IDStr == "" && qtNote.ID == "" {

			blocks = append(blocks, &slack.DividerBlock{Type: slack.MBTDivider},
				&slack.SectionBlock{
					Type: slack.MBTSection,
					Text: &slack.TextBlockObject{
						Type: "mrkdwn",
						Text: fmt.Sprintf("<%s|%s> (deleted)", legacyTweet.QuotedStatusPermalink.Expanded,
							legacyTweet.QuotedStatusPermalink.Display),
						Verbatim: true,
					},
				})

		} else {
			if qtNote.ID != "" {
				tweetText = qtNote.Text
				entities = qtNote.EntitySet.getShortenURLs("", "")
			} else {
				tweetText = qtLegacy.FullText
				entities = qtLegacy.Entities.getShortenURLs(qtLegacy.ConversationIDStr, qtLegacy.InReplyToUserIDStr)
			}

			blocks = append(blocks,
				&slack.DividerBlock{
					Type: slack.MBTDivider,
				},
				getUserBlock(qtuser),
				getTweetBlock(tweetText, entities),
			)

			for _, p := range qtLegacy.ExtendedEntities.Media {
				if block := getMediaBlocks(p); block != nil {
					blocks = append(blocks, block)
				}
			}
			blocks = append(blocks, getCreatedAtBlock(time.Time(qtLegacy.CreatedAt)))
		}
	}

	return &slack.Attachment{Blocks: slack.Blocks{BlockSet: blocks}}, nil
}

type rubyDateTime time.Time

func (rdt *rubyDateTime) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}

	tt, err := time.Parse(`"`+time.RubyDate+`"`, string(data))
	*rdt = rubyDateTime(tt)

	return err
}

type statusLegacyEntity struct {
	CreatedAt        rubyDateTime         `json:"created_at"`
	ID               int64                `json:"id"`
	IDStr            string               `json:"id_str"`
	FullText         string               `json:"full_text"`
	Entities         externalLinkEntities `json:"entities"`
	ExtendedEntities struct {
		Media []mediaEntity `json:"media"`
	} `json:"extended_entities"`
	SelfThread struct {
		ID    int64  `json:"id"`
		IDStr string `json:"id_str"`
	} `json:"self_thread"`
	RetweetCount          int    `json:"retweet_count"`
	FavoriteCount         int    `json:"favorite_count"`
	ReplyCount            int    `json:"reply_count"`
	ConversationIDStr     string `json:"conversation_id_str"`
	InReplyToUserIDStr    string `json:"in_reply_to_user_id_str"`
	QuotedStatusIDStr     string `json:"quoted_status_id_str"`
	QuotedStatusPermalink struct {
		URL      string `json:"url"`
		Expanded string `json:"expanded"`
		Display  string `json:"display"`
	} `json:"quoted_status_permalink"`
}

type statusResultEntity struct {
	Typename string `json:"__typename"`
	statusResultCommonEntity
	QuotedRefResult struct {
		Result struct {
			Typename string `json:"__typename"`
			RestID   string `json:"rest_id"`
		} `json:"result"`
	} `json:"quotedRefResult"`
	Tweet statusResultCommonEntity `json:"tweet"`
}

type noteTweetEntity struct {
	IsExpandable     bool `json:"is_expandable"`
	NoteTweetResults struct {
		Result struct {
			ID        string               `json:"id"`
			Text      string               `json:"text"`
			EntitySet externalLinkEntities `json:"entity_set"`
		} `json:"result"`
	} `json:"note_tweet_results"`
}

type statusResultCommonEntity struct {
	RestID    string             `json:"rest_id"`
	Legacy    statusLegacyEntity `json:"legacy"`
	NoteTweet noteTweetEntity    `json:"note_tweet"`
	Core      struct {
		UserResults struct {
			Result struct {
				Typename       string     `json:"__typename"`
				ID             string     `json:"id"`
				RestID         string     `json:"rest_id"`
				IsBlueVerified bool       `json:"is_blue_verified"`
				Legacy         userEntity `json:"legacy"`
			} `json:"result"`
		} `json:"user_results"`
	} `json:"core"`
	QuotedStatusResult struct {
		Result *statusResultEntity `json:"result"`
	} `json:"quoted_status_result"`
}

type StatusResultWrapper struct {
	statusResultEntity
	QuotedStatusResult struct {
		Result statusResultEntity `json:"result"`
	} `json:"quoted_status_result"`
}

type statusContainer struct {
	Data struct {
		TweetResult struct {
			Result StatusResultWrapper `json:"result"`
		} `json:"tweetResult"`
	} `json:"data"`
}
