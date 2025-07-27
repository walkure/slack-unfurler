package twitter

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/slack-go/slack"
)

func fetchGrokStatus(id string) (*slack.Attachment, error) {

	grokShareContainer := &grokEntity{}

	if err := invokeGraphQL("GrokShare", loginApiID, func(graphQuery url.Values) {
		graphQuery.Set("variables", fmt.Sprintf(`{"grok_share_id":"%s","withSafetyModeUserFields":true}`, id))
		graphQuery.Set("features", `{"creator_subscriptions_tweet_preview_api_enabled":true,"communities_web_enable_tweet_community_results_fetch":true,"c9s_tweet_anatomy_moderator_badge_enabled":true,"articles_preview_enabled":true,"responsive_web_edit_tweet_api_enabled":true,"graphql_is_translatable_rweb_tweet_is_translatable_enabled":true,"view_counts_everywhere_api_enabled":true,"longform_notetweets_consumption_enabled":true,"responsive_web_twitter_article_tweet_consumption_enabled":true,"tweet_awards_web_tipping_enabled":false,"creator_subscriptions_quote_tweet_preview_enabled":false,"freedom_of_speech_not_reach_fetch_enabled":true,"standardized_nudges_misinfo":true,"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled":true,"rweb_video_timestamps_enabled":true,"longform_notetweets_rich_text_read_enabled":true,"longform_notetweets_inline_media_enabled":true,"rweb_tipjar_consumption_enabled":true,"responsive_web_graphql_exclude_directive_enabled":true,"verified_phone_label_enabled":false,"responsive_web_graphql_skip_user_profile_image_extensions_enabled":false,"responsive_web_graphql_timeline_navigation_enabled":true,"responsive_web_enhance_cards_enabled":false}`)

	}, grokShareContainer); err != nil {
		return nil, fmt.Errorf("fetch grok share: %w", err)
	}

	return extractGrokShare(grokShareContainer)
}

func extractGrokShare(grokShareContainer *grokEntity) (*slack.Attachment, error) {
	grokShare := grokShareContainer.Data.GrokShare.Items

	if len(grokShare) == 0 {
		return nil, fmt.Errorf("no grok share found")
	}

	blocks := make([]slack.Block, 0, len(grokShare))

	for id, item := range grokShare {
		if item.Message == "" {
			continue
		}

		blocks = append(blocks,
			&slack.SectionBlock{
				Type: slack.MBTSection,
				Text: &slack.TextBlockObject{
					Type:     "mrkdwn",
					Text:     markdown2mrkdwn(item.Message),
					Verbatim: true,
				},
			},
		)

		if len(item.PublicMediaUrls) > 0 {
			for _, mediaURL := range item.PublicMediaUrls {
				blocks = append(blocks,
					&slack.ImageBlock{
						Type:     slack.MBTImage,
						ImageURL: mediaURL,
						AltText:  "media",
					},
				)
			}
		}

		if id != len(grokShare)-1 {
			blocks = append(blocks, &slack.DividerBlock{Type: slack.MBTDivider})
		}

	}

	return &slack.Attachment{Blocks: slack.Blocks{BlockSet: blocks}}, nil
}

func markdown2mrkdwn(markdown string) string {

	// replace markdown italic to mrkdwn italic
	markdown = regexp.MustCompile(`\*(.*?)\*`).ReplaceAllString(markdown, "_${1}_")

	// replace markdown bold to mrkdwn bold
	markdown = regexp.MustCompile(`__(.*?)__`).ReplaceAllString(markdown, "*$1*")

	return markdown
}

type grokEntity struct {
	Data struct {
		GrokShare struct {
			Items []struct {
				Message         string   `json:"message"`
				PublicMediaUrls []string `json:"public_media_urls"`
				Sender          string   `json:"sender"`
			} `json:"items"`
		} `json:"grokShare"`
	} `json:"data"`
}
