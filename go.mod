module github.com/gibsn/duty_bot

go 1.14

require (
	github.com/anatoliyfedorenko/isdayoff v1.0.2
	github.com/emersion/go-ical v0.0.0-20200224201310-cd514449c39e
	github.com/emersion/go-webdav v0.3.1
	github.com/kr/pretty v0.2.0 // indirect
	github.com/mail-ru-im/bot-golang v0.0.0-20200509193603-2c56a20fca87
	github.com/mitchellh/mapstructure v1.4.3
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.7.1
	golang.org/x/sys v0.0.0-20211210111614-af8b64212486 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/emersion/go-webdav => github.com/gibsn/go-webdav v0.3.2-0.20220511212135-85f98a968374
