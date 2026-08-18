package dtodashboard

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type GetDashboardOverviewRequest struct {
	apiobj.BaseRequest
	Request GetDashboardOverviewEmbedRequest
}

type GetDashboardOverviewEmbedRequest struct {
	// CompanyID 公司ID
	CompanyID uint `json:"company_id"`
	// BeginAt 开始时间（Unix 时间戳-秒）
	BeginAt int64 `json:"begin_at"`
	// EndAt 结束时间（Unix 时间戳-秒）
	EndAt int64 `json:"end_at"`
}

func (opt *GetDashboardOverviewRequest) Validity(_ *GetDashboardOverviewResponse) {
}
