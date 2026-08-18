package dashboard

import (
	"github.com/insmtx/corekg/apps/admin/models/dashboard"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type GetDashboardDataResponse struct {
	apiobj.BaseResponse
	Response struct {
		Forest  dashboard.DataSet `json:"forest"`
		File    dashboard.DataSet `json:"file"`
		Parse   dashboard.DataSet `json:"parse"`
		Session dashboard.DataSet `json:"session"`
	}
}

type GetDashboardDataRequest struct {
	apiobj.BaseRequest
	Request struct {
		CompanyID uint `json:"company_id"`
	}
}

func (r *GetDashboardDataRequest) Validate(o *apiobj.BaseResponse) {
	if r.Request.CompanyID < 0 {
		o.Code = errcode.ErrCode_BadRequest
		o.Message = "非法公司id"
	}
}
