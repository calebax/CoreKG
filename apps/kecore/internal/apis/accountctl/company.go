package accountctl

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/company"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// GetCompanyQuota 获取公司资源配额
// @Tags 企业版本
// @Summary 公司资源配额
// @Description 获取公司资源配额
// @Router /forest.GetCompanyQuota [post]
// @Param user body apiobj.BaseRequest true "入参"
// @Success 200 {object} GetCompanyQuotaResponse "返回值"
func GetCompanyQuota(ctx *gin.Context, _ *apiobj.BaseRequest, resp *GetCompanyQuotaResponse) {
	cmpID := runtime.CompanyID(ctx)
	if cmpID == 0 {
		return
	}
	//calculate all quota
	quota, err := company.GetCompanyQuota(ctx, cmpID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCompanyQuota: get company quota failed, %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_company_quota_failed" // 获取公司资源配额失败
		return
	}
	resp.Response.Quota = *quota
}

// VersionUpgradeVerify 企业版本升级申请校验验证码与表单
// @Tags 企业版本
// @Summary 企业版本升级申请校验验证码与表单
// @Description 企业版本升级申请校验验证码与表单
// @Router /forest.VersionUpgradeVerify [post]
// @Param request body VersionUpgradeVerifyCodeRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func VersionUpgradeVerify(ctx *gin.Context, req *VersionUpgradeVerifyCodeRequest, resp *apiobj.BaseResponse) {
	// 校验参数
	if req.Validity(resp); resp.Code != 0 {
		return
	}

	var position string
	switch req.Request.Type {
	case accounttype.FormTypeUpgrade:
		position = user.UpgradeFormPosition
	case accounttype.FormTypeContact:
		position = user.ContactFormPosition
	case accounttype.FormTypeDotpenContact:
		position = user.DotpenContactFormPosition
	}
	if err := user.CustomerVerifySms(runtime.Uin(ctx), req.Request.Phone, position, req.Request.Code, resp); err != nil {
		logs.ErrorContextf(ctx, "VersionUpgrade verify phone failed, %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_verify_phone_failed" // 手机号验证码失败
		return
	}

	if err := dbutil.Account().Create(&req.Request.CompanyUpgradeApply).Error; err != nil {
		logs.ErrorContextf(ctx, "create version upgrade apply faild, %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_upgrade_apply_failed" // 创建申请失败
		return
	}

	err := company.WechatBotWebhook(ctx, &req.Request.CompanyUpgradeApply, ctx.ClientIP())
	if err != nil {
		logs.ErrorContextf(ctx, "WechatBotWebhook verify phone failed, %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_upgrade_apply_failed" // 创建申请失败
		return
	}
}

// VersionUpgradeSendCode 企业版本升级申请发送验证码
// @Tags 企业版本
// @Summary 企业版本升级申请发送验证码
// @Description 企业版本升级申请发送验证码
// @Router /forest.VersionUpgradeSendCode [post]
// @Param request body VersionUpgradeSendCodeRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func VersionUpgradeSendCode(ctx *gin.Context, req *VersionUpgradeSendCodeRequest, resp *apiobj.BaseResponse) {
	if req.Validity(resp); resp.Code != 0 {
		return
	}
	var position string
	switch req.Request.Type {
	case accounttype.FormTypeUpgrade:
		position = user.UpgradeFormPosition
	case accounttype.FormTypeContact:
		position = user.ContactFormPosition
	case accounttype.FormTypeDotpenContact:
		position = user.DotpenContactFormPosition
	}
	if err := user.CustomerSendSms(ctx, runtime.Uin(ctx), req.Request.Phone, position, resp); err != nil {
		logs.ErrorContextf(ctx, "send sms failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_send_sms_failed" // 发送验证码失败
		return
	}
}
