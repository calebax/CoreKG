package accountctl

import (
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/company"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/pkgs/utils/validate"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type GetCompanyQuotaResponse struct {
	apiobj.BaseResponse
	Response struct {
		Quota company.Quota `json:"quota"`
	}
}

type VersionUpgradeSendCodeRequest struct {
	apiobj.BaseRequest
	Request struct {
		Phone string               `json:"phone"`
		Type  accounttype.FormType `json:"type"`
	}
}

func (req *VersionUpgradeSendCodeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Phone == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_phone_empty" // 手机号不能为空
		return
	}
	if err := validate.IsPhone(req.Request.Phone); err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_phone_format" // 手机号格式错误
		return
	}
	switch req.Request.Type {
	case accounttype.FormTypeUpgrade, accounttype.FormTypeContact, accounttype.FormTypeDotpenContact:
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_form_type" // 非法表单类型
	}
}

type VersionUpgradeVerifyCodeRequest struct {
	apiobj.BaseRequest
	Request struct {
		accounttype.CompanyUpgradeApply
		Code string `json:"code"`
	}
}

func (req *VersionUpgradeVerifyCodeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Phone == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_phone_empty" // 手机号不能为空
		return
	}
	if err := validate.IsPhone(req.Request.Phone); err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_phone_format" // 手机号格式错误
		return
	}
	if len(req.Request.Code) != user.SMSCodeLen {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_verification_code" // 验证码非法
		return
	}
	if len(req.Request.Name) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_name_empty" // 姓名不可为空
		return
	}

	if len(req.Request.CompanyName) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_company_name_empty" // 公司名不可为空
		return
	}
	switch req.Request.Type {
	case accounttype.FormTypeContact, accounttype.FormTypeUpgrade, accounttype.FormTypeDotpenContact:
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_form_type" // 非法表单类型
	}
}
