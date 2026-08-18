package apis

import "github.com/ygpkg/yg-go/apis/apiobj"

// 登陆相关

// LoginWechatRequest 微信用户登录
type LoginWechatRequest struct {
	apiobj.BaseRequest

	Request struct {
		Code    string `json:"code"`
		AppID   string `json:"appid"`
		AgentID string `json:"agentid"`
	}
}

// LoginResponse 登录返回
type LoginResponse struct {
	apiobj.BaseResponse

	Response struct {
		// UserInfo 用户信息
		UserInfo LoginResponseUserInfo `json:"user_info"`

		// JwtToken 平台token，用于访问平台
		JwtToken string `json:"jwt_token,omitempty"`
	}
}

// LoginResponseUserInfo 登录返回用户信息
type LoginResponseUserInfo struct {
	// ID 用户ID
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}
