package loginctl

import (
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
)

// LoginPasswordRequest 企业微信登录
type LoginPasswordRequest struct {
	apiobj.BaseRequest

	Request struct {
		Username string `json:"username"`
		// Password 原密码md5加密后的密码
		Password string `json:"password"`
	}
}

func (req *LoginPasswordRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Username == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "用户名不能为空"
		return
	}
	if req.Request.Password == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "密码不能为空"
		return
	}
	return
}

type LoginResponse struct {
	apiobj.BaseResponse

	Response struct {
		UserID uint `json:"user_id"`

		// LoginStatus 登陆状态
		// success: 登陆成功
		// register: 需要注册
		// failed: 登陆失败
		LoginStatus string `json:"login_status"`
		// UserInfo 用户信息
		UserInfo *user.UserInfo `json:"user_info,omitempty"`
		// JwtToken jwt token
		JwtToken string `json:"jwt_token,omitempty"`
		// FailedReason 登陆失败原因
		FailedReason string `json:"failed_reason,omitempty"`
		// Uin 分类uin
		Uin []*LoginUin `json:"uin,omitempty"`
		// Issuer 颁发者
		Issuer string `json:"issuer,omitempty"`
		// 是否允许注册
		AllowRegister bool `json:"allow_register"`
		// RefreshToken 用来继续选择用户后的登录
		RefreshToken string `json:"refresh_token,omitempty"`
		// 登录方式
		LoginWay auth.LoginWay `json:"login_way"`
	}
}

type LoginUin struct {
	Uin           accounttype.UserIdentification `json:"uin"`
	Name          string                         `json:"name,omitempty"`
	CompanyName   string                         `json:"company_name,omitempty"`
	Role          accounttype.SysRole            `json:"role,omitempty"`
	CompanyStatus accounttype.CompanyStatus      `json:"company_status,omitempty"`
}

type LoginThirdRequest struct {
	apiobj.BaseRequest

	Request struct {
		// Way github, work_wechat, wechat_web
		Way        string `json:"way"`
		Code       string `json:"code"`
		DomainName string `json:"domain_name"`
	}
}
