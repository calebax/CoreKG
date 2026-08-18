package dtouser

import (
	"github.com/insmtx/corekg/pkgs/utils/validate"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type ResetPasswordKey string

const (
	PasswordKeyForgot ResetPasswordKey = "ForgotPassword"
	PasswordKeyReset  ResetPasswordKey = "ResetPassword"
	PasswordKeyChange ResetPasswordKey = "PasswordChange"
)

var ResetPasswordKeys = []ResetPasswordKey{
	PasswordKeyForgot,
	PasswordKeyReset,
	PasswordKeyChange,
}

func isValidResetPasswordKey(key string) bool {
	for _, validKey := range ResetPasswordKeys {
		if string(validKey) == key {
			return true
		}
	}
	return false
}

type RequestPasswordResetCodeRequest struct {
	apiobj.BaseRequest
	Request struct {
		// 手机号
		Phone string `json:"phone"`
		// 验证码key，用于区分场景（可选，ForgotPassword）
		Key string `json:"key"`
	}
}

func (req *RequestPasswordResetCodeRequest) Validity(resp *apiobj.BaseResponse) {
	if !isValidResetPasswordKey(req.Request.Key) {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "code_key_invalid" // 验证码key无效
		return
	}
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

// ForgotPasswordRequest 使用验证码重置密码
type ForgotPasswordRequest struct {
	apiobj.BaseRequest
	Request struct {
		// 手机号
		Phone string `json:"phone"`
		// 验证码
		Code string `json:"code"`
		// 新密码
		Password string `json:"password"`
	}
}

func (req *ForgotPasswordRequest) Validity(resp *apiobj.BaseResponse) {
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
	if req.Request.Code == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_verify_code_empty" // 验证码不能为空
		return
	}
	if len(req.Request.Password) < 6 {
		resp.Code = errcode.ErrCode_PasswordTooShort
		resp.Message = "account_password_too_short" // 密码长度不能小于6位
		return
	}
}
