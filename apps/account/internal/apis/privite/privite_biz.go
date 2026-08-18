package privite

import (
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/pkgs/utils/validate"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
)

// CreateEmployeeRequest 创建员工请求
type CreateEmployeeRequest struct {
	apiobj.BaseRequest
	Request struct {
		EmpInfo
	}
}

// Valid 校验创建员工请求
func (r *CreateEmployeeRequest) Valid(p *apiobj.BaseResponse) {
	if err := validate.IsUsername(r.Request.UserName); err != nil {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_invalid_username" // 用户名无效
		return
	}
	if len(r.Request.Password) == 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_new_password_empty" // 密码为空
		return
	}
	if len(r.Request.Password) < 8 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_password_too_short_less8" // 密码长度最低为8位
		return
	}
	if len(r.Request.Password) > 36 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_password_too_long" // 密码最大长度36位
		return
	}
	if validate.HasCNChar(r.Request.Password) {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_password_no_chinese" // 密码不能包含中文
		return
	}
	if err := validate.IsEmail(r.Request.Email); err != nil {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_invalid_email_format" // 非法邮箱格式
		return
	}
	if len(r.Request.Email) > 50 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_email_too_long" // 邮箱最大长度为50
		return
	}
	if len(r.Request.Phone) > 0 {
		if err := validate.IsPhone(r.Request.Phone); err != nil {
			p.Code = errcode.ErrCode_BadRequest
			p.Message = "account_invalid_phone_format" // 非法手机号
			return
		}
	}
}

// EmpInfo 员工信息
type EmpInfo struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

// EditEmployeeRequest 编辑员工请求
type EditEmployeeRequest struct {
	apiobj.BaseRequest
	Request struct {
		Uin uint `json:"uin"`
		EmpInfo
	}
}

// Valid 校验编辑员工请求
func (r *EditEmployeeRequest) Valid(p *apiobj.BaseResponse) {
	if r.Request.Uin <= 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_invalid_uin" // 非法员工Uin
		return
	}
	if err := validate.IsUsername(r.Request.UserName); err != nil {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_invalid_username" // 用户名无效
		return
	}
	if err := validate.IsEmail(r.Request.Email); err != nil {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_invalid_email_format" // 非法邮箱格式
		return
	}
	if len(r.Request.Email) > 50 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_email_too_long" // 邮箱最大长度为50
		return
	}
	if len(r.Request.Password) < 8 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_password_too_short_less8" // 密码长度不可低于8位
		return
	}
	if len(r.Request.Password) > 36 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_password_too_long" // 密码最大长度36位
		return
	}
	if validate.HasCNChar(r.Request.Password) {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_password_no_chinese" // 密码不能包含中文
		return
	}
}

// LoginPasswordRequest 企业微信登录请求
type LoginPasswordRequest struct {
	apiobj.BaseRequest
	Request struct {
		DomainName string `json:"domain_name"`
		// Username 用户名 不能为空, 邮箱或者用户名
		Username string `json:"username"`
		// Password 原密码md5加密后的密码
		Password string `json:"password"`
	}
}

// LoginUin 登录Uin信息
type LoginUin struct {
	Uin           accounttype.UserIdentification `json:"uin"`
	Name          string                         `json:"name,omitempty"`
	CompanyName   string                         `json:"company_name,omitempty"`
	CompanyLogo   string                         `json:"company_logo,omitempty"`
	Role          accounttype.SysRole            `json:"role,omitempty"`
	CompanyStatus accounttype.CompanyStatus      `json:"company_status,omitempty"`
}

// LoginThirdResponse 登录响应
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

// DeleteEmployeeRequest 删除用户请求
type DeleteEmployeeRequest struct {
	apiobj.BaseRequest
	Request struct {
		EmployeeID uint `json:"employee_id"`
		// DeleteReason 删除原因
		DeleteReason string `json:"delete_reason"`
	}
}
