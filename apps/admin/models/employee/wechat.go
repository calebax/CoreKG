package employee

import (
	"context"
	"fmt"
	"strings"

	"github.com/mozillazg/go-pinyin"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/oauth"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

// UserInfo 微信用户信息
type UserInfo struct {
	WechatUnionID   string `json:"wechat_union_id,omitempty"`
	WechatWebOpenID string `json:"wechat_web_open_id,omitempty"`

	Identify  string `json:"identify,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Bio       string `json:"bio,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Name      string `json:"name,omitempty"`
	Sex       uint8  `json:"sex,omitempty"`
}

func getOpsWechatWebConfig(ctx context.Context) (*offConfig.Config, error) {
	cfg := &offConfig.Config{}
	if err := settings.GetYaml("account", "wechat_web_oauth", cfg); err != nil {
		logs.ErrorContextf(ctx, "getOpsWechatWebConfig: get config failed, %s", err)
		return nil, err
	}
	cfg.Cache = cache.NewMemory()
	logs.InfoContextf(ctx, "getOpsWechatWebConfig: %+v", cfg.AppID)
	return cfg, nil
}

// getOpsWechatWebOAuth 获取微信网页授权
func getOpsWechatWebOAuth(ctx context.Context) (*oauth.Oauth, error) {
	cfg, err := getOpsWechatWebConfig(ctx)
	if err != nil {
		return nil, err
	}
	wApp := wechat.NewWechat()
	oa := wApp.GetOfficialAccount(cfg).GetOauth()
	return oa, nil
}

// GetOpsWechatWebLoginUserInfo 获取运营端网页扫码登录用户的信息
func GetOpsWechatWebLoginUserInfo(ctx context.Context, code string, resp *apiobj.BaseResponse) (*UserInfo, error) {
	wxapi, err := getOpsWechatWebOAuth(ctx)
	if err != nil {
		resp.Code = errcode.ErrCode_NotFound
		resp.Message = fmt.Sprintf("获取微信开放平台配置失败: %v", err)
		logs.ErrorContextf(ctx, "getOpsWechatWebOAuth failed ,err %s", err)
		return nil, err
	}

	tkn, err := wxapi.GetUserAccessToken(code)
	if err != nil {
		resp.Code = errcode.ErrCode_NotFound
		resp.Message = fmt.Sprintf("获取用户授权信息失败: %v", err)
		logs.ErrorContextf(ctx, "GetUserAccessToken failed ,err %s", err)
		return nil, err
	}

	user, err := wxapi.GetUserInfo(tkn.AccessToken, tkn.OpenID, "")
	if err != nil {
		resp.Code = errcode.ErrCode_NotFound
		resp.Message = fmt.Sprintf("获取微信用户信息失败: %v", err)
		logs.ErrorContextf(ctx, "GetUserInfo failed ,err %s", err)
		return nil, err
	}

	logs.InfoContextf(ctx, "get tkn webopenid %s ,user %+v", user.OpenID, user)
	userInfo := &UserInfo{
		WechatUnionID:   user.Unionid,
		WechatWebOpenID: user.OpenID,
		Identify:        strings.Join(pinyin.LazyPinyin(user.Nickname, pinyin.NewArgs()), ""),
		AvatarURL:       user.HeadImgURL,
		Sex:             uint8(user.Sex),
		Name:            user.Nickname,
	}
	return userInfo, nil
}
