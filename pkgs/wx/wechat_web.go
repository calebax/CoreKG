package wx

import (
	"context"

	"github.com/silenceper/wechat/v2"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/oauth"
	"github.com/ygpkg/yg-go/cache"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

// GetWechatWebOAuth 获取微信网页授权
func GetWechatWebOAuth(ctx context.Context, group, key string) (*oauth.Oauth, error) {
	cfg := &offConfig.Config{}
	if err := settings.GetYaml(group, key, cfg); err != nil {
		logs.ErrorContextf(ctx, "GetWechatWebOAuth: get config failed, %s", err)
		return nil, err
	}
	logs.InfoContextf(ctx, "GetWechatWebOAuth(%s/%s): %+v", group, key, cfg.AppID)

	wApp := wechat.NewWechat()

	wApp.SetCache(cache.WechatCache())
	oa := wApp.GetOfficialAccount(cfg).GetOauth()
	return oa, nil
}
