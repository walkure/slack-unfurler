package twitter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/slack-go/slack"
)

func fetchTwitterProfile(screenName string) (*slack.Attachment, error) {
	return fetchAPI("UserByScreenName", loginApiID, func(graphQuery url.Values) {
		graphQuery.Set("variables", fmt.Sprintf(`{"screen_name":"%s","withSafetyModeUserFields":true}`, screenName))
		graphQuery.Set("features", `{"hidden_profile_likes_enabled":false,"hidden_profile_subscriptions_enabled":true,"responsive_web_graphql_exclude_directive_enabled":true,"verified_phone_label_enabled":false,"subscriptions_verification_info_is_identity_verified_enabled":false,"subscriptions_verification_info_verified_since_enabled":true,"highlights_tweets_tab_ui_enabled":true,"creator_subscriptions_tweet_preview_api_enabled":true,"responsive_web_graphql_skip_user_profile_image_extensions_enabled":false,"responsive_web_graphql_timeline_navigation_enabled":true}`)
		graphQuery.Set("fieldToggles", `{"withAuxiliaryUserLabels":false}`)

	}, func(responseBody io.Reader) (*slack.Attachment, error) {
		return extractProfile(responseBody)
	})
}

func extractProfile(responseBody io.Reader) (*slack.Attachment, error) {

	userContainer := userContainer{}

	//tr := io.TeeReader(responseBody, os.Stdout)

	if err := json.NewDecoder(responseBody).Decode(&userContainer); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}
	fmt.Println("")

	userLegacy := userContainer.Data.User.Result.Legacy

	blocks := []slack.Block{
		getProfileSummaryBlock(userLegacy),
		getProfileDescriptionBlock(userLegacy),
		getProfileParamsBlock(userLegacy),
	}

	return &slack.Attachment{Blocks: slack.Blocks{BlockSet: blocks}}, nil
}

func getProfileSummaryBlock(userLegacy userLegacyEntity) *slack.ContextBlock {

	protected := ""
	if userLegacy.Protected {
		protected = " 🔒"
	}

	return &slack.ContextBlock{
		Type: slack.MBTContext,
		ContextElements: slack.ContextElements{
			Elements: []slack.MixedElement{
				slack.ImageBlockElement{
					Type:     slack.METImage,
					ImageURL: userLegacy.ProfileImageURLHTTPS,
					AltText:  userLegacy.ScreenName,
				},
				slack.TextBlockObject{
					Type: "mrkdwn",
					Text: fmt.Sprintf("<https://twitter.com/%s|*%s*> <https://twitter.com/%s|@%s>%s",
						userLegacy.ScreenName, userLegacy.Name,
						userLegacy.ScreenName, userLegacy.ScreenName,
						protected,
					),
				},
			},
		},
	}
}

func getProfileDescriptionBlock(userLegacy userLegacyEntity) *slack.SectionBlock {

	txt := extractShortenURL(userLegacy.Description, userLegacy.Entities.Description.Urls)
	p := strings.Replace(userLegacy.ProfileImageURLHTTPS, "_normal", "", 1)

	return &slack.SectionBlock{
		Type: slack.MBTSection,
		Text: &slack.TextBlockObject{
			Type: "mrkdwn",
			Text: txt,
		},
		Accessory: &slack.Accessory{
			ImageElement: &slack.ImageBlockElement{
				Type:     slack.METImage,
				ImageURL: p,
				AltText:  "profile image",
			},
		},
	}
}

func getProfileParamsBlock(userLegacy userLegacyEntity) *slack.ContextBlock {

	elms := []slack.MixedElement{}

	url := extractShortenURL(userLegacy.Url, userLegacy.Entities.Url.Urls)
	if url != "" {
		url = "🔗 " + url
		elms = append(elms, slack.TextBlockObject{
			Type: "mrkdwn",
			Text: url,
		})
	}
	location := userLegacy.Location
	if location != "" {
		location = "📍 " + location
		elms = append(elms, slack.TextBlockObject{
			Type: "plain_text",
			Text: location,
		})
	}

	return &slack.ContextBlock{
		Type: slack.MBTContext,
		ContextElements: slack.ContextElements{
			Elements: elms,
		},
	}
}

type userContainer struct {
	Data struct {
		User struct {
			Result userResultEntity `json:"result"`
		} `json:"user"`
	} `json:"data"`
}

type userResultEntity struct {
	Typename                   string `json:"__typename"`
	ID                         string `json:"id"`
	RestID                     string `json:"rest_id"`
	AffiliatesHighlightedLabel struct {
	} `json:"affiliates_highlighted_label"`
	HasGraduatedAccess              bool                      `json:"has_graduated_access"`
	IsBlueVerified                  bool                      `json:"is_blue_verified"`
	ProfileImageShape               string                    `json:"profile_image_shape"`
	Legacy                          userLegacyEntity          `json:"legacy"`
	SmartBlockedBy                  bool                      `json:"smart_blocked_by"`
	SmartBlocking                   bool                      `json:"smart_blocking"`
	LegacyExtendedProfile           userLegacyExtendedProfile `json:"legacy_extended_profile"`
	IsProfileTranslatable           bool                      `json:"is_profile_translatable"`
	HasHiddenSubscriptionsOnProfile bool                      `json:"has_hidden_subscriptions_on_profile"`
	VerificationInfo                struct {
	} `json:"verification_info"`
	HighlightsInfo struct {
		CanHighlightTweets bool   `json:"can_highlight_tweets"`
		HighlightedTweets  string `json:"highlighted_tweets"`
	} `json:"highlights_info"`
	BusinessAccount struct {
	} `json:"business_account"`
	CreatorSubscriptionsCount int `json:"creator_subscriptions_count"`
}

type userLegacyEntity struct {
	Protected           bool   `json:"protected"`
	Following           bool   `json:"following"`
	CanDm               bool   `json:"can_dm"`
	CanMediaTag         bool   `json:"can_media_tag"`
	CreatedAt           string `json:"created_at"`
	DefaultProfile      bool   `json:"default_profile"`
	DefaultProfileImage bool   `json:"default_profile_image"`
	Description         string `json:"description"`
	Entities            struct {
		Description struct {
			Urls []urlShortenEntity `json:"urls"`
		} `json:"description"`
		Url struct {
			Urls []urlShortenEntity `json:"urls"`
		} `json:"url"`
	} `json:"entities"`
	FastFollowersCount      int      `json:"fast_followers_count"`
	FavouritesCount         int      `json:"favourites_count"`
	FollowersCount          int      `json:"followers_count"`
	FriendsCount            int      `json:"friends_count"`
	HasCustomTimelines      bool     `json:"has_custom_timelines"`
	IsTranslator            bool     `json:"is_translator"`
	ListedCount             int      `json:"listed_count"`
	Location                string   `json:"location"`
	MediaCount              int      `json:"media_count"`
	Name                    string   `json:"name"`
	NormalFollowersCount    int      `json:"normal_followers_count"`
	PinnedTweetIdsStr       []string `json:"pinned_tweet_ids_str"`
	PossiblySensitive       bool     `json:"possibly_sensitive"`
	ProfileImageURLHTTPS    string   `json:"profile_image_url_https"`
	ProfileInterstitialType string   `json:"profile_interstitial_type"`
	ScreenName              string   `json:"screen_name"`
	StatusesCount           int      `json:"statuses_count"`
	TranslatorType          string   `json:"translator_type"`
	Url                     string   `json:"url"`
	Verified                bool     `json:"verified"`
	WantRetweets            bool     `json:"want_retweets"`
	WithheldInCountries     []any    `json:"withheld_in_countries"`
}

type userLegacyExtendedProfile struct {
	Birthdate struct {
		Day            int    `json:"day"`
		Month          int    `json:"month"`
		Visibility     string `json:"visibility"`
		YearVisibility string `json:"year_visibility"`
	} `json:"birthdate"`
}
