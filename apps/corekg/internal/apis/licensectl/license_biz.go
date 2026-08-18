package licensectl

import (
	"github.com/insmtx/corekg/apps/admin/models/admintype"
	corekglicense "github.com/insmtx/corekg/apps/corekg/models/license"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type RegisterLicenseRequest struct {
	apiobj.BaseRequest
	Request struct {
		License     string      `json:"license"`
		CompanyName string      `json:"company_name"`
		CompanyLogo string      `json:"company_logo"`
		WebsiteInfo WebsiteInfo `json:"website_info"`
	}
}

type WebsiteInfo struct {
	WebsiteLogo string `json:"website_logo"`
	WebsiteName string `json:"website_name"`
}

func (r *RegisterLicenseRequest) Valid(p *apiobj.BaseResponse) {
	if len(r.Request.License) <= 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "corekg_license_empty" // License不可为空
		return
	}
	if len(r.Request.CompanyName) <= 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "corekg_company_name_empty" // 公司名称不可为空
		return
	}
	if len(r.Request.CompanyLogo) <= 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "corekg_company_logo_empty" // 公司logo不可为空
		return
	}
	if len(r.Request.WebsiteInfo.WebsiteLogo) <= 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "corekg_website_logo_empty" // 网站logo不可为空
		return
	}
	if len(r.Request.WebsiteInfo.WebsiteName) <= 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "corekg_website_name_empty" // 网站名称不可为空
		return
	}
}

type GetLicenseInfoResponse struct {
	apiobj.BaseResponse
	Response struct {
		Meta      *admintype.Meta                `json:"meta"`
		Modules   []global.Module                `json:"modules"`
		Status    corekglicense.ValidationStatus `json:"status"`
		ValidDays int                            `json:"valid_days"`

		WebsiteInfo

		CompanyLogo string `json:"company_logo"`
		CompanyName string `json:"company_name"`
	}
}
