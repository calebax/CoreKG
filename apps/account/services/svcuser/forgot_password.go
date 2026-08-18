package svcuser

import (
	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"

	"github.com/insmtx/corekg/apps/account/internal/dto/dtouser"
	"github.com/insmtx/corekg/apps/account/models/user"
)

// RequestPasswordResetCode 发送重置密码验证码到手机号
func RequestPasswordResetCode(ctx *gin.Context, phone, key string) *dtouser.ErrorResponse {
	usr, err := user.GetUserByPhone(phone)
	if err != nil || usr == nil || usr.ID == 0 {
		logs.InfoContextf(ctx, "RequestPasswordResetCode: user not found, phone: %v", phone)
		return &dtouser.ErrorResponse{
			Code:    errcode.ErrCode_BadRequest,
			Message: "account_user_not_found",
		}
	}

	sendCodeResponse := &apiobj.BaseResponse{}
	if err := user.CustomerSendSms(ctx, usr.ID, phone, key, sendCodeResponse); err != nil {
		logs.ErrorContextf(ctx, "RequestPasswordResetCode: send sms failed, %s", err)
		return &dtouser.ErrorResponse{
			Code:    sendCodeResponse.Code,
			Message: sendCodeResponse.Message,
		}
	}
	return nil
}

// ForgotPassword 校验验证码并重置密码
func ForgotPassword(ctx *gin.Context, phone, code, password string) *dtouser.ErrorResponse {
	usr, err := user.GetUserByPhone(phone)
	if err != nil || usr == nil || usr.ID == 0 {
		logs.InfoContextf(ctx, "ForgotPassword: user not found, phone: %v", phone)
		return &dtouser.ErrorResponse{
			Code:    errcode.ErrCode_BadRequest,
			Message: "account_user_not_found",
		}
	}

	// 验证短信验证码
	verifyCodeResponse := &apiobj.BaseResponse{}
	if err := user.CustomerVerifySms(usr.ID, phone, string(dtouser.PasswordKeyForgot), code, verifyCodeResponse); err != nil {
		logs.InfoContextf(ctx, "ForgotPassword: verify sms failed, %s", err)
		return &dtouser.ErrorResponse{
			Code:    verifyCodeResponse.Code,
			Message: verifyCodeResponse.Message,
		}
	}

	// 更新密码
	if err := user.UpdateAccountPassword(usr.ID, password); err != nil {
		logs.ErrorContextf(ctx, "ForgotPassword: update password failed, %s", err)
		return &dtouser.ErrorResponse{
			Code:    errcode.ErrCode_InternalError,
			Message: "account_update_password_failed",
		}
	}

	return nil
}
