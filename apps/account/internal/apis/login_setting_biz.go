package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type GetLoginSettingRequest struct {
	apiobj.BaseRequest

	Request struct {
		DomainName string `json:"domain_name"`
		Path       string `json:"path"`
	}
}

func (req *GetLoginSettingRequest) Validity(ctx *gin.Context, resp *apiobj.BaseResponse) {
	if req.Request.DomainName == "" {
		req.Request.DomainName = ctx.Request.Host
	}
}

type settingData struct {
	WeChat struct {
		Enable bool   `json:"enable"`
		AppID  string `json:"appid"`
	} `json:"wechat"`
	WeChatCom struct {
		Enable bool `json:"enable"`
		// 企微登录appid
		AppIDCom string `json:"appid_com"`
		// 企微agentid，自建项目使用
		AgentID string `json:"agentid"`
	} `json:"wechat_com"`
	Phone struct {
		Enable bool `json:"enable"`
	} `json:"phone"`
	Emali struct {
		Enable bool `json:"enable"`
	} `json:"email"`
	Password struct {
		Enable bool `json:"enable"`
	} `json:"password"`
	Title         string `json:"title"`
	ImageUrl      string `json:"image_url"`
	AllowRegister bool   `json:"allow_register"`
	Issuer        string `json:"issuer"`
	AuthKey       string `json:"auth_key"`
	LoginURL      string `json:"login_url"`
}

type GetLoginSettingResponse struct {
	apiobj.BaseResponse
	Response settingData
}

type CheckPassWordRequest struct {
	apiobj.BaseRequest
	Request struct {
		Password string `json:"password"`
	}
}
