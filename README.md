# slack-unfurler

SlackのURL展開を[go](https://github.com/slack-go/slack)でやってみた。

## 対応

- [秋月電子](https://akizukidenshi.com/)
  - [goquery](https://github.com/PuerkitoBio/goquery)でスクレイピング

## 起動方法

1. SlackのAppを[作成](https://api.slack.com/apps?new_app=1)
2. Socket-modeを有効にする(まだWebhookは作ってないので)
3. Event Subscriptionsで「App unfurl domains」を選んでよしなにドメインを追加
4. OAuth & PermissionsでBot Tokenに `links:write` を追加
   1. `links:read`は「App unfurl domains」した際に追加されるっぽい。
5. 環境変数 `SLACK_BOT_TOKEN` と `SLACK_APP_TOKEN` によしなにトークンを設定する。
6. 起動

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
