# slack-unfurler

SlackのURL展開を[go](https://github.com/slack-go/slack)でやってみた。

## 対応

- [秋月電子](https://akizukidenshi.com/)
  - [goquery](https://github.com/PuerkitoBio/goquery)でスクレイピング
- [Twitter](https://twitter.com/)
  - APIから取得を試み、失敗すると埋め込みTweetから引用を試みます。

## 起動方法

1. SlackのAppを[作成](https://api.slack.com/apps?new_app=1)
2. Socket-modeを有効にする(まだWebhookは作ってないので)
3. Event Subscriptionsで「App unfurl domains」を選んでよしなにドメインを追加
   1. Twitter videoを展開する場合は、`twitter.com`だけでなく`video.twimg.com`も追加する。
   2. その上で、環境変数`UNFURL_TWITTER_VIDEO`に何かしらの値を設定。
4. OAuth & PermissionsでBot Tokenに `links:write` を追加
   1. `links:read`は「App unfurl domains」した際に追加されるっぽい。
   2. Twitter Videoを展開する場合は`links.embed:write`も必要。忘れると展開時に`cannot_parse_attachment`と出る。
5. 環境変数 `SLACK_BOT_TOKEN` と `SLACK_APP_TOKEN` によしなにトークンを設定する。
6. Twitter APIで展開する場合は以下の設定を追加
   1. ブラウザでログインした際にCookieへセットされた`auth_token`をいくつか取得して`TWITTER_AUTH_TOKENS_FROM_BROWSER`にカンマ区切りで設定する。
      1. この値は1年間有効の模様。
      2. 念のために日頃使ってるアカウントの値は使わないほうがいいでしょう。
   2. `auth_token`を送って呼ばれる`TweetResultByRestId`を呼び出すURIのうちID部分を`TWITTER_API_ID`に設定
      1. `https://twitter.com/i/api/graphql/0hWvDhmW8YQ-S_ib3azIrw/TweetResultByRestId?variables(以下略)`でリクエストを送っている場合`0hWvDhmW8YQ-S_ib3azIrw`の部分
      2. どうやら時々変わる様子。
   3. API呼び出しでの展開に失敗する場合 or 設定がない場合は埋め込みTweet展開を試します。このときNSFW Tweetは展開できません。
7. 起動

## 展開情報のキャッシュについて

参照元の負荷を下げるため、LRUで128件キャッシュします。

### Tweet展開について

Twitterの展開は、埋め込みTweetを取得する方法とAPI経由する方法の２つを使っています。それぞれpros/consがあるので、お好みに応じて設定してください。

### 埋め込みTweetを取得

- pros
  - token不要
  - 複数の動画が添付されたTweetに対応
- cons
  - NSFW Tweet非対応
    - 全く存在しないものとする扱いをする
    - NSFW TweetをQTしたTweetを展開する場合も、QTしたNSFW Tweetが全く存在しないとみなすので存在すらわからない。

### API経由

- pros
  - NSFW Tweet対応
  - RT数も取れる
    - 現状は展開していません
- cons
  - `auth_token`が必要
  - rate limitがある

## 参考文献

[Unfurling links in messages](https://api.slack.com/reference/messaging/link-unfurling)

## 言語

go 1.20

## todo

- caching
- webhook

## License

MIT

## Author

walkure at 3pf.jp
