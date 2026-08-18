package lkxctl

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/internal/dto/dtolkx"
	"github.com/insmtx/corekg/apps/admin/services/svrlkx"

	"github.com/ygpkg/yg-go/logs"
)

// SendVerifyCode 发送验证码
// @Tags 丽科星
// @Summary 发送验证码
// @Description 发送验证码
// @Router /lkxadmin.SendVerifyCode [post]
// @Param request body dtolkx.SendVerifyCodeRequest true "参数"
// @Success 200 {object} dtolkx.SendVerifyCodeResponse "成功响应"
func SendVerifyCode(ctx *gin.Context, req *dtolkx.SendVerifyCodeRequest, resp *dtolkx.SendVerifyCodeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}

	// 发送验证码
	err := svrlkx.SendVerifyCode(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "send verify code failed, %s", err)
		resp.Code = 500
		resp.Message = "短信发送失败，请稍后再试"
		return
	}
	resp.Code = 0
	resp.Message = "短信发送成功"
}

// VerifyCodeAndSave 验证验证码成功保存
// @Tags 丽科星
// @Summary 验证验证码成功保存
// @Description 验证验证码成功保存
// @Router /lkxadmin.VerifyCodeAndSave [post]
// @Param request body dtolkx.VerifyVerifyCodeRequest true "参数"
// @Success 200 {object} dtolkx.VerifyVerifyCodeResponse "成功响应"
func VerifyCodeAndSave(ctx *gin.Context, req *dtolkx.VerifyVerifyCodeRequest, resp *dtolkx.VerifyVerifyCodeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	// 验证验证码
	err := svrlkx.VerifyCodeAndSave(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "verify code and save failed, %s", err)
		resp.Code = 500
		resp.Message = "验证码验证失败"
		return
	}
	resp.Code = 0
	resp.Message = "验证码验证成功"
}
