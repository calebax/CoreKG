package wecom

import (
	"encoding/xml"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/pkgs/wecoms"
	"github.com/ygpkg/yg-go/apis/runtime"
)

// VerifyServerMessage 验证企业微信消息
// @Tags 企业微信
// @Summary 验证接口
// @Produce json
// @Param namespace formData string false "命名空间,企业id"
// @Success 200
// @Param namespace formData string false "命名空间,企业id"
// @Param appid formData string false "appid"
// Success 200 {object} account.User
func ReceiveServerMessage(ctx *gin.Context) {
	var (
		vfyMsg VerifyMessage
		logger = runtime.Logger(ctx)
	)
	if err := ctx.BindQuery(&vfyMsg); err != nil {
		runtime.BadRequest(ctx, "query failed, %s", err.Error())
		return
	}
	logger.Infof("[wecom] request: %+v", vfyMsg)

	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		runtime.BadRequest(ctx, "read body failed, %s", err.Error())
		return
	}
	x := &xmlRxEnvelope{}
	if err := xml.Unmarshal(body, x); err != nil {
		runtime.BadRequest(ctx, "decode xml failed, %s", err.Error())
		return
	}
	vfyMsg.EchoStr = x.Encrypt

	var (
		companyID = ctx.Param("namespace")
		appID     = ctx.Param("appid")
	)
	app, err := wecoms.GetApp(companyID, appID)
	if err != nil {
		runtime.BadRequest(ctx, "not found company secret for %s, %s", companyID, err.Error())
		return
	}

	if !vfyMsg.Verify(ctx, app.Token) {
		runtime.BadRequest(ctx, "verified failed.")
		return
	}

	msgData, err := vfyMsg.Decrypt(ctx, app.ToConfig())
	if err != nil {
		return
	}
	logger.Debugf("decrypt: %s", msgData)

	rcvMsg := &wecoms.ReceiveMessage{}
	if err := xml.Unmarshal(msgData, rcvMsg); err != nil {
		runtime.BadRequest(ctx, "decode decrypted xml failed, %s", err.Error())
		return
	}

	switch rcvMsg.MsgType {
	case wecoms.MessageTypeText:
		logger.Infof("receive message(%s): %s", rcvMsg.MsgType, rcvMsg.Content)
		// app.WxCli().SendTextMessage(&workwx.Recipient{
		// 	UserIDs: []string{rcvMsg.FromUserName},
		// }, string(rcvMsg.Content), false)
	case wecoms.MessageTypeLocation:
		logger.Infow("receive message",
			"type", rcvMsg.MsgType,
			"lat", rcvMsg.LatX,
			"lon", rcvMsg.LonY,
			"scale", rcvMsg.Scale,
			"label", rcvMsg.Label,
		)
		// app.WxCli().SendTextMessage(&workwx.Recipient{
		// 	UserIDs: []string{rcvMsg.FromUserName},
		// }, fmt.Sprintf("(%v, %v)[%s]", rcvMsg.LatX, rcvMsg.LonY, rcvMsg.Label), false)
	case wecoms.MessageTypeImage:
		logger.Infow("receive message",
			"MsgType", rcvMsg.MsgType,
			"MediaID", rcvMsg.MediaID,
			"PicURL", rcvMsg.PicURL,
		)
		// app.WxCli().SendImageMessage(&workwx.Recipient{
		// 	UserIDs: []string{rcvMsg.FromUserName},
		// }, rcvMsg.MediaID, false)
	case wecoms.MessageTypeEvent:
		switch rcvMsg.Event {
		case wecoms.EventTypeLocation:
			logger.Infow("receive event",
				"Event", rcvMsg.Event,
				"lat", rcvMsg.Lat,
				"lon", rcvMsg.Lon,
			)
			// app.WxCli().SendTextMessage(&workwx.Recipient{
			// 	UserIDs: []string{rcvMsg.FromUserName},
			// }, fmt.Sprintf("(%v, %v)[%s]", rcvMsg.Lat, rcvMsg.Lon, "来自事件"), false)
		}

	default:
		logger.Infof("receive message(%s): %+v", rcvMsg.MsgType, rcvMsg)
	}

}

// VerifyServerMessage 验证企业微信消息
// @Tags 企业微信
// @Summary 验证接口
// @Produce json
// @Param namespace formData string false "命名空间,企业id"
// @Param appid formData string false "appid"
// @Success 200
// @Router /apis/wecom/v1/namespaces/:namespace/app/:appid/serve [get]
// Success 200 {object} account.User
func VerifyServerMessage(ctx *gin.Context) {
	var (
		companyID = ctx.Param("namespace")
		appID     = ctx.Param("appid")
	)
	app, err := wecoms.GetApp(companyID, appID)
	if err != nil {
		runtime.BadRequest(ctx, "not found company secret for %s, %s", companyID, err.Error())
		return
	}
	HandleVerifyServerMessage(ctx, app.ToConfig())
}
