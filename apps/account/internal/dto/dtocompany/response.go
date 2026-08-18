package dtocompany

import (
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type ErrorResponse struct {
	Code    uint32 `json:"code"`    // 错误码，
	Message string `json:"message"` // 错误信息，
}

type CreateCompanyResponse struct {
	apiobj.BaseResponse
	Response *CompanyInfo
}

type CompanyInfo struct {
	Uin           accounttype.UserIdentification `json:"uin"`
	Name          string                         `json:"name,omitempty"`
	CompanyLogo   string                         `json:"company_logo,omitempty"`
	CompanyName   string                         `json:"company_name,omitempty"`
	Role          accounttype.SysRole            `json:"role,omitempty"`
	CompanyStatus accounttype.CompanyStatus      `json:"company_status,omitempty"`
	LastLoginAt   *time.Time                     `json:"last_login_at,omitempty"`
}
