# slack-unfurler

SlackのURL展開を[go](https://github.com/slack-go/slack)でやってみた。

## 対応

- [秋月電子](https://akizukidenshi.com/)
  - [goquery](https://github.com/PuerkitoBio/goquery)でスクレイピング
- [Twitter](https://twitter.com/)
  - 埋め込みTweetから引用を試み、出来なかったらAPIを叩きます。
  - 動画埋め込みの展開には非対応です。

## 起動方法

1. SlackのAppを[作成](https://api.slack.com/apps?new_app=1)
2. Socket-modeを有効にする(まだWebhookは作ってないので)
3. Event Subscriptionsで「App unfurl domains」を選んでよしなにドメインを追加
4. OAuth & PermissionsでBot Tokenに `links:write` を追加
   1. `links:read`は「App unfurl domains」した際に追加されるっぽい。
5. 環境変数 `SLACK_BOT_TOKEN` と `SLACK_APP_TOKEN` によしなにトークンを設定する。
6. Twitter APIで展開する場合は、ブラウザでログインした際にCookieへセットされた`auth_token`をいくつか取得して`TWITTER_AUTH_TOKENS_FROM_BROWSER`に設定する。
   1. この値は1年間有効の模様。
   2. 念のために日頃使ってるアカウントの値は使わないほうがいいでしょう。
   3. これを設定した状態で`USE_TWITTER_SYNDICATION`に何かしらの値を入れると、APIを叩く前に埋め込みTweet引用を試します。
7. 起動

## 参考文献

[Unfurling links in messages](https://api.slack.com/reference/messaging/link-unfurling)

## 言語

go 1.17

## todo

- caching
- webhook

## License

MIT

## Author

walkure at 3pf.jp
