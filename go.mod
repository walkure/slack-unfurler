module github.com/walkure/slack-unfurler

go 1.20

require (
	github.com/PuerkitoBio/goquery v1.8.0
	github.com/saintfish/chardet v0.0.0-20120816061221-3af4cd4741ca
	github.com/slack-go/slack v0.12.2
	golang.org/x/net v0.0.0-20210916014120-12bc252f5db8
)

require (
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	golang.org/x/text v0.3.6 // indirect
)

replace github.com/slack-go/slack v0.12.2 => github.com/walkure/slack-go v0.12.3-pre1
