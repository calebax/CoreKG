package apis

import (
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/pkgs/utils/validate"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
)

// DetailPersonalCenterRequest 获取等待认证的列表
type DetailPersonalCenterRequest struct {
	apiobj.BaseRequest
}

func (req *DetailPersonalCenterRequest) Validity(resp *DetailPersonalCenterResponse) {

}

// DetailPersonalCenterResponse 获取等待认证的列表
type DetailPersonalCenterResponse struct {
	apiobj.BaseResponse
	Response struct {
		// UserInfo 用户信息
		UserInfo *user.UserInfo `json:"user_info,omitempty"`
		// CompanyInfo 公司信息
		CompanyInfo *accounttype.Company `json:"company_info,omitempty"`
		// EmployeeDetail 员工信息
		EmployeeDetail *employee.EmployeeDetail `json:"employee_detail,omitempty"`
	}
}

// DetailPersonalCenterRequest 获取等待认证的列表
type ListUinRequest struct {
	apiobj.BaseRequest
}

func (req *ListUinRequest) Validity(resp *ListUinResponse) {

}

// ListUinResponse 获取等待认证的列表
type ListUinResponse struct {
	apiobj.BaseResponse
	Response struct {
		// Uin 分类uin
		Uin []*LoginUin `json:"uin,omitempty"`
	}
}

// SwitchLoginRequest 更换登录身份信息
type SwitchLoginRequest struct {
	apiobj.BaseRequest
	Request struct {
		Uin uint `json:"uin"`
		// 登录方式
		LoginWay auth.LoginWay `json:"login_way"`
	}
}

func (req *SwitchLoginRequest) Validity(resp *SwitchLoginResponse) {
	if req.Request.Uin == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_select_login_identity" // 请选择登录身份
		return
	}
	if req.Request.LoginWay == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_select_login_method" // 请选择登录方式
		return
	}
}

// SwitchLoginResponse 更换登录身份信息
type SwitchLoginResponse struct {
	apiobj.BaseResponse
	Response struct {
		JwtToken string `json:"jwt_token,omitempty"`
	}
}

// UpdatePhoneVerifyCodeRequest 更新手机号发送验证码请求体
type UpdatePhoneVerifyCodeRequest struct {
	apiobj.BaseRequest
	Request struct {
		Phone     string `json:"phone"`
		PhoneCode string `json:"phone_code"`
	}
}

// Validity
func (req *UpdatePhoneVerifyCodeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Phone == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_phone_empty" // 手机号不能为空
		return
	}
	if req.Request.PhoneCode == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_verify_code_empty" // 验证码不能为空
		return
	}
	if err := validate.IsPhone(req.Request.Phone); err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_phone_format" // 手机号格式错误
		return
	}
}

// UpdatePhoneVerifyCodeResponse 更新手机号发送验证码响应体
type UpdatePhoneVerifyCodeResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

// UpdatePhoneSendCodeRequest 更新手机号发送验证码请求体
type UpdatePhoneSendCodeRequest struct {
	apiobj.BaseRequest
	Request struct {
		Phone string `json:"phone"`
	}
}

func (req *UpdatePhoneSendCodeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Phone == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_phone_empty" // 手机号不能为空
		return
	}
	if err := validate.IsPhone(req.Request.Phone); err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_phone_format" // 手机号格式错误
		return
	}
}

type UpdatePhoneSendCodeResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

// BindUserWechatRequest 绑定第三方微信请求体
type BindUserWechatRequest struct {
	apiobj.BaseRequest
	Request struct {
		Code string `json:"code"`
	}
}

func (req *BindUserWechatRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Code == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_code_empty" // code不能为空
		return
	}
}

type BindUserWechatResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

// UpdateAccountPasswordRequest 更新账号密码请求体
type UpdateAccountPasswordRequest struct {
	apiobj.BaseRequest
	Request struct {
		// 旧密码
		OldPassword string `json:"old_password"`
		// 新密码
		NewPassword string `json:"new_password"`
	}
}

func (req *UpdateAccountPasswordRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.NewPassword == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_new_password_empty" // 新密码不能为空
		return
	}
	if len(req.Request.NewPassword) < 8 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_password_too_short" // 密码长度最短8位
		return
	}
	if len(req.Request.NewPassword) > 36 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_password_too_long" // 密码长度最长36位
		return
	}
	if validate.HasCNChar(req.Request.NewPassword) {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_password_no_chinese" // 密码不能包含中文
		return
	}
}

// UpdateAccountPasswordResponse 更新账号密码响应体
type UpdateAccountPasswordResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

// UpdateUserInfoRequest 更新用户信息请求体
type UpdateUserInfoRequest struct {
	apiobj.BaseRequest
	Request struct {
		// 用户名
		Name string `json:"name"`
		// 头像
		AvatarURL string `json:"avatar_url"`
		// 邮箱
		Email *string `json:"email"`
	}
}

func (req *UpdateUserInfoRequest) Validity(resp *apiobj.BaseResponse) {
	// 校验邮箱
	if req.Request.Email != nil && len(*req.Request.Email) > 0 {
		if err := validate.IsEmail(*req.Request.Email); err != nil {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "邮箱格式错误"
			return
		}
	}
}

// UpdateUserInfoResponse 更新用户信息响应体
type UpdateUserInfoResponse struct {
	apiobj.BaseResponse
	Response struct{}
}
