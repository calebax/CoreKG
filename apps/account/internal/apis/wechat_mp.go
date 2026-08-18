package apis

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/wechatmp"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"github.com/ygpkg/yg-go/logs"
)

var (
	// syncWechatSubscriptInfoOnce 同步微信订阅信息
	syncWechatSubscriptInfoOnce sync.Once
)

// HandleWechatMpMessage 处理微信公众号消息
func HandleWechatMpMessage(ctx *gin.Context) {
	mp, err := wechatmp.GetWechatOfficialAccount(ctx, "account", "mp_wechat_service")
	if err != nil {
		logs.ErrorContextf(ctx, "HandleWechatMpMessage: %v", err)
		return
	}
	appid := mp.GetBasic().AppID

	go syncWechatSubscriptInfoOnce.Do(func() {
		wechatmp.SyncWechatSubscriptionInfo(ctx, mp)
	})
	// mp.GetTemplate()
	mpServer := mp.GetServer(ctx.Request, ctx.Writer)
	mpServer.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
		logs.InfoContextf(ctx, "HandleWechatMpMessage: %+v", msg)
		switch msg.MsgType {
		case message.MsgTypeText:
		case message.MsgTypeEvent:
			switch msg.Event {
			case message.EventSubscribe:
				openid := string(msg.FromUserName)
				userinfo, err := mp.GetUser().GetUserInfo(openid)
				if err != nil {
					logs.ErrorContextf(ctx, "HandleWechatMpMessage: get userinfo failed %s", err)
					return &message.Reply{}
				}
				err = wechatmp.UpsertWechatOfficialSubscription(ctx, appid, userinfo)
				if err != nil {
					logs.ErrorContextf(ctx, "HandleWechatMpMessage: upsert useraccount failed %s", err)
					return &message.Reply{}
				}
				return &message.Reply{
					MsgType: message.MsgTypeText,
					MsgData: message.NewText("欢迎关注"),
				}
			case message.EventUnsubscribe:
				err := wechatmp.UnsubscribeToWeChatOfficial(ctx, appid, string(msg.FromUserName))
				if err != nil {
					logs.ErrorContextf(ctx, "HandleWechatMpMessage: unsubscribe failed %s", err)
					return &message.Reply{}
				}
			case message.EventScan:
			}
		}
		return &message.Reply{}
	})

	if err := mpServer.Serve(); err != nil {
		logs.ErrorContextf(ctx, "HandleWechatMpMessage: %+v", err)
		return
	}
	mpServer.Send()
}
