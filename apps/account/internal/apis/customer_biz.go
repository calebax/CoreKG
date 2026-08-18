package apis

import (
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/company"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
)

type LoginThirdRequest struct {
	apiobj.BaseRequest

	Request struct {
		// Way github, work_wechat, wechat_web
		Way        string      `json:"way"`
		Code       string      `json:"code"`
		DomainName string      `json:"domain_name"`
		Option     LoginOption `json:"option,omitempty"`
	}
}

type LoginThirdResponse struct {
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
	CompanyLogo   string                         `json:"company_logo,omitempty"`
	CompanyName   string                         `json:"company_name,omitempty"`
	Role          accounttype.SysRole            `json:"role,omitempty"`
	CompanyStatus accounttype.CompanyStatus      `json:"company_status,omitempty"`
	LastLoginAt   *time.Time                     `json:"last_login_at,omitempty"`
	CompanyUserID uint                           `json:"company_user_id,omitempty"`
}

// RegisterThirdRequest 第三方注册
type RegisterThirdRequest struct {
	apiobj.BaseRequest
	Request struct {
		Way string `json:"way"`
		// Code     string         `json:"code"`
		UserInfo    *user.UserInfo       `json:"user_info"`
		Issuer      string               `json:"issuer"`
		CompanyInfo *company.CompanyInfo `json:"company_info"`
		Option      LoginOption          `json:"option,omitempty"`
	}
}

// CheckUserIdentifyExistRequest 检查用户是否存在
type CheckUserIdentifyExistRequest struct {
	apiobj.BaseRequest
	Request struct {
		Identify string `json:"identify"`
	}
}

// CheckUserIdentifyExistResponse 检查用户是否存在
type CheckUserIdentifyExistResponse struct {
	apiobj.BaseResponse
	Response struct {
		Identify string `json:"identify"`
		Exist    bool   `json:"exist"`
	}
}

// ChooseUinRequest 选择UIN
type ChooseUinRequest struct {
	apiobj.BaseRequest
	Request struct {
		RefreshToken string `json:"refresh_token"`
		UinID        uint   `json:"uin_id"`
		UserID       uint   `json:"user_id"`
		// 登录方式
		LoginWay auth.LoginWay `json:"login_way"`
	}
}

// ChooseUinResponse 选择UIN响应
type ChooseUinResponse struct {
	apiobj.BaseResponse
	Response struct {
		// JwtToken jwt token
		JwtToken string `json:"jwt_token,omitempty"`
	}
}

// GetOBOTokenRequest 内部服务 OBO 获取调用凭证
type GetOBOTokenRequest struct {
	apiobj.BaseRequest
	Request struct {
		// Uin 用户身份 ID
		Uin uint `json:"uin"`
		// Audience 调用方服务标识
		Audience string `json:"audience"`
		// GrantType 授权类型（预留）
		GrantType string `json:"grant_type"`
		// Scope 权限范围（预留）
		Scope string `json:"scope"`
	}
}

func (req *GetOBOTokenRequest) Validity(resp *GetOBOTokenResponse) {
	if req.Request.Uin == 0 || req.Request.Audience == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters"
		return
	}
}

// GetOBOTokenResponse OBO 调用凭证响应
type GetOBOTokenResponse struct {
	apiobj.BaseResponse
	Response struct {
		// JwtToken jwt token
		JwtToken string `json:"jwt_token,omitempty"`
		// ExpiredAt token 有效截止时间（Unix 秒）
		ExpiredAt int64 `json:"expired_at,omitempty"`
	}
}

// LoginOption 登录选项,用于适配与逻辑拓展
type LoginOption struct {
	// 单UIN是否直接登录
	SingleUinDirective bool `json:"single_uin_direct"`
	//SetDefaultUin 是否为无uin用户直接设置默认UIN
	SetDefaultUin bool `json:"set_default_uin"`
}
