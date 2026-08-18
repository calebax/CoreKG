package wechatmp

import (
	"github.com/silenceper/wechat/v2/officialaccount/message"
)

// NewLoginVerificationCodeTemlateMessage 新建登录验证码模板消息
func NewLoginVerificationCodeTemlateMessage(
	devName string,
	loginTime string,
	loginPlace string,
	userName string,
	verificationCode string,
) *message.TemplateMessage {
	return &message.TemplateMessage{
		Data: map[string]*message.TemplateDataItem{
			// 设备名称 {{thing5.DATA}}
			"thing5": {Value: devName},
			// 登录时间 {{time6.DATA}}
			"time6": {Value: loginTime},
			// 登录地点 {{thing8.DATA}}
			"thing8": {Value: loginPlace},
			// 用户名称 {{thing2.DATA}}
			"thing2": {Value: userName},
			// 登录验证 {{character_string14.DATA}}
			"character_string14": {Value: verificationCode},
		},
	}
}
