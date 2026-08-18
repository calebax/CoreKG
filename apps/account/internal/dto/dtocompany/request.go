package dtocompany

import (
	"strings"

	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// CreateCompany创建公司
type CreateCompanyRequest struct {
	apiobj.BaseRequest
	Request struct {
		DomainName      string `json:"domain_name"`
		RefreshToken    string `json:"refresh_token"`
		UserID          uint   `json:"user_id"`
		CompanyName     string `json:"company_name"`
		UserDisplayName string `json:"user_display_name"`
	}
}

func (req *CreateCompanyRequest) Validity(resp *CreateCompanyResponse) {
	req.Request.CompanyName = strings.TrimSpace(req.Request.CompanyName)
	if len(req.Request.CompanyName) <= 0 {
		// 参数错误
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters"
		return
	}
	req.Request.UserDisplayName = strings.TrimSpace(req.Request.UserDisplayName)
	if len(req.Request.UserDisplayName) <= 0 {
		// 参数错误
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters"
		return
	}
}
