package testutils

type AppName string

const (
	AppNameKecore   AppName = "kecore"
	AppNameKechat   AppName = "kechat"
	AppNameKesale   AppName = "kesale"
	AppNameAdmin    AppName = "admin"
	AppNameKesearch AppName = "kesearch"
)

var appNameMap = map[AppName]struct{}{
	AppNameKecore:   {},
	AppNameKechat:   {},
	AppNameKesale:   {},
	AppNameAdmin:    {},
	AppNameKesearch: {},
}
