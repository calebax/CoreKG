package wechatmp

import (
	"context"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/officialaccount"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/user"
	"github.com/ygpkg/yg-go/cache"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

// GetWechatOffConfig 获取微信配置
func GetWechatOffConfig(ctx context.Context, group, key string) (*offConfig.Config, error) {
	cfg := &offConfig.Config{}
	if err := settings.GetYaml(group, key, cfg); err != nil {
		logs.ErrorContextf(ctx, "GetWechatOffConfig: get config failed, %s", err)
		return nil, err
	}
	return cfg, nil
}

// GetWechatOfficialAccount 获取微信公众号实例
func GetWechatOfficialAccount(ctx context.Context, group, key string) (*officialaccount.OfficialAccount, error) {
	cfg, err := GetWechatOffConfig(ctx, group, key)
	if err != nil {
		logs.ErrorContextf(ctx, "GetWechatMP: get config failed, %s", err)
		return nil, err
	}
	logs.InfoContextf(ctx, "GetWechatMP: %v", cfg.AppID)

	wApp := wechat.NewWechat()

	wApp.SetCache(cache.WechatCache())
	return wApp.GetOfficialAccount(cfg), nil
}

// UpsertWechatOfficialSubscription 更新微信公众号订阅
func UpsertWechatOfficialSubscription(ctx context.Context, appid string, userInfo *user.Info) error {
	u := &accounttype.WechatBinding{
		AppID:         appid,
		OpenID:        userInfo.OpenID,
		UnionID:       userInfo.UnionID,
		Nickname:      userInfo.Nickname,
		Sex:           userInfo.Sex,
		City:          userInfo.City,
		Country:       userInfo.Country,
		Province:      userInfo.Province,
		Headimgurl:    userInfo.Headimgurl,
		SubscribeTime: userInfo.SubscribeTime,
		Subscribe:     userInfo.Subscribe,
	}
	err := dbtools.InsertOrUpdate(dbutil.Account(), u,
		"subscribe", "nickname", "sex", "city", "country", "province", "headimgurl", "subscribe_time")
	if err != nil {
		logs.ErrorContextf(ctx, "UpsertWechatOfficialSubscription: %s", err)
		return err
	}

	return nil
}

// UnsubscribeToWeChatOfficial 取消订阅微信公众号
func UnsubscribeToWeChatOfficial(ctx context.Context, appid, openid string) error {
	rst := dbutil.Account().Table(accounttype.TableNameWechatBinding).
		Where("app_id = ? AND open_id = ?", appid, openid).
		Update("subscribe", 0)
	if rst.Error != nil {
		logs.ErrorContextf(ctx, "UnsubscribeToWeChatPublicPlatform: %s", rst.Error)
		return rst.Error
	}
	return nil
}

// SyncWechatSubscriptionInfo 同步微信订阅信息
func SyncWechatSubscriptionInfo(ctx context.Context, mp *officialaccount.OfficialAccount) {
	appid := mp.GetBasic().AppID
	mpUser := mp.GetUser()

	nextOpenID := ""
	cnt := 0
	for {
		userList, err := mpUser.ListUserOpenIDs(nextOpenID)
		if err != nil {
			logs.ErrorContextf(ctx, "SyncWechatSubscriptionInfo: list user openids failed %s", err)
			return
		}
		for _, openid := range userList.Data.OpenIDs {
			userinfo, err := mpUser.GetUserInfo(openid)
			if err != nil {
				logs.ErrorContextf(ctx, "SyncWechatSubscriptionInfo: get userinfo failed %s", err)
				continue
			}
			err = UpsertWechatOfficialSubscription(ctx, appid, userinfo)
			if err != nil {
				logs.ErrorContextf(ctx, "SyncWechatSubscriptionInfo: upsert useraccount failed %s", err)
				continue
			}
			cnt++
		}
		if userList.Count < userList.Total && userList.NextOpenID != "" {
			nextOpenID = userList.NextOpenID
		} else {
			break
		}
	}
	logs.InfoContextf(ctx, "SyncWechatSubscriptionInfo: total %d", cnt)
}

// GetWechatBindingByUnionID 根据UnionID获取微信绑定信息
func GetWechatBindingByUnionID(ctx context.Context, appid, unionID string) (*accounttype.WechatBinding, error) {
	var wechatBinding accounttype.WechatBinding
	rst := dbutil.Account().Table(accounttype.TableNameWechatBinding).
		Where("app_id = ? AND union_id = ?", appid, unionID).
		First(&wechatBinding)
	if rst.Error != nil {
		logs.ErrorContextf(ctx, "GetWechatBindingByUnionID: %s", rst.Error)
		return nil, rst.Error
	}
	return &wechatBinding, nil
}

// GetOpenIDByUsername 根据用户名获取OpenID
func GetOpenIDByUsername(ctx context.Context, appid, username string) (string, error) {
	emp, err := employee.GetUserByUsername(ctx, username)
	if err != nil {
		logs.WarnContextf(ctx, "get employee by username failed, %s", err)
		return "", err
	}
	// if emp.Status != accounttype.UserStatusNormal {
	// 	logs.WarnContextf(ctx, "user status is not normal, %s", emp.Status)
	// 	return "", err
	// }
	if emp.WechatUnionID == nil {
		logs.WarnContextf(ctx, "user unionid is nil")
		return "", err
	}

	wechatBinding, err := GetWechatBindingByUnionID(ctx, appid, *emp.WechatUnionID)
	if err != nil {
		logs.WarnContextf(ctx, "get wechat binding by unionid failed, %s", err)
		return "", err
	}
	if wechatBinding.Subscribe != 1 {
		logs.WarnContextf(ctx, "wechat binding is not subscribe")
		return "", err
	}
	return wechatBinding.OpenID, nil
}
