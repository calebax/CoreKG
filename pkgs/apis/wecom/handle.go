package wecom

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/pkgs/utils/encryptor"
	"github.com/insmtx/corekg/pkgs/wecoms"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
)

// HandleVerifyServerMessage
func HandleVerifyServerMessage(ctx *gin.Context, app config.WecomApp) {
	var (
		vfyMsg VerifyMessage
		logger = runtime.Logger(ctx)
	)
	if err := ctx.BindQuery(&vfyMsg); err != nil {
		runtime.BadRequest(ctx, "query failed, %s", err.Error())
		return
	}
	logger.Infof("[wecom] request: %+v", vfyMsg)

	if !vfyMsg.Verify(ctx, app.Token) {
		runtime.BadRequest(ctx, "verified failed.")
		return
	}

	msgData, err := vfyMsg.Decrypt(ctx, app)
	if err != nil {
		return
	}
	ctx.Writer.Write(msgData)
	ctx.AbortWithStatus(http.StatusOK)
}

type VerifyMessage struct {
	MsgSign   string `form:"msg_signature"`
	Timestamp int64  `form:"timestamp"`
	Nonce     int64  `form:"nonce"`

	EchoStr string `form:"echostr"`
}

func (v VerifyMessage) signStrs(token string) []string {
	sli := []string{
		token,
		fmt.Sprint(v.Timestamp),
		fmt.Sprint(v.Nonce),
		v.EchoStr,
	}
	sort.Strings(sli)
	return sli
}

func (v VerifyMessage) makeDevMsgSignature(ctx context.Context, token string) string {
	state := sha1.New()
	for _, str := range v.signStrs(token) {
		_, err := state.Write([]byte(str))
		if err != nil {
			logs.ErrorContextf(ctx, "[wecom] write bash writer failed, %s", err)
			return ""
		}
	}
	result := state.Sum(nil)
	return hex.EncodeToString(result)
}

func (v VerifyMessage) Verify(ctx context.Context, token string) bool {
	gotSign := v.makeDevMsgSignature(ctx, token)
	if gotSign == v.MsgSign {
		return true
	}
	logs.WarnContextf(ctx, "message signature not equal, got: %s, exp: %s", gotSign, v.MsgSign)
	return false
}

func (v VerifyMessage) Decrypt(ctx *gin.Context, app config.WecomApp) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(v.EchoStr)
	if err != nil {
		runtime.BadRequest(ctx, "decode base64 msg failed, %s, msg: %s", err, v.EchoStr)
		return nil, err
	}

	ret, err := encryptor.AesDecrypt(data, app.AESKey())
	if err != nil {
		runtime.BadRequest(ctx, "decrypt msg failed, %s", err)
		return nil, err
	}

	msgLenData := ret[16:20]
	msgLen := binary.BigEndian.Uint32(msgLenData)
	msgData := ret[20 : 20+msgLen]
	msgId := ret[20+msgLen:]
	runtime.Logger(ctx).Debugf("[wecom] decrypt data: %s, len: %v, msg: %s, id: %s", ret, msgLen, msgData, msgId)

	return msgData, nil
}

type xmlRxEnvelope struct {
	ToUserName string `xml:"ToUserName"`
	AgentID    string `xml:"AgentID"`
	Encrypt    string `xml:"Encrypt"`
}

type ReceiveMessageHandler func(ctx *gin.Context, app config.WecomApp, rcvMsg *wecoms.ReceiveMessage)

func HandleReceiveServerMessage(ctx *gin.Context, app config.WecomApp, handler ReceiveMessageHandler) {
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

	if !vfyMsg.Verify(ctx, app.Token) {
		runtime.BadRequest(ctx, "verified failed.")
		return
	}

	msgData, err := vfyMsg.Decrypt(ctx, app)
	if err != nil {
		return
	}
	logger.Debugf("decrypt: %s", msgData)

	rcvMsg := &wecoms.ReceiveMessage{}
	if err := xml.Unmarshal(msgData, rcvMsg); err != nil {
		runtime.BadRequest(ctx, "decode decrypted xml failed, %s", err.Error())
		return
	}

	handler(ctx, app, rcvMsg)
}

var DefaultReceiveMessageHandler ReceiveMessageHandler = func(ctx *gin.Context, app config.WecomApp, rcvMsg *wecoms.ReceiveMessage) {

	var logger = runtime.Logger(ctx)
	switch rcvMsg.MsgType {
	case wecoms.MessageTypeText:
		logs.InfoContextf(ctx, "receive message(%s): %s", rcvMsg.MsgType, rcvMsg.Content)
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
