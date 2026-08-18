package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/internal/dto/dtodashboard"
	"github.com/insmtx/corekg/apps/admin/services/svcdashboard"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// GetDashboardOverview 获取运营概览数据
// @Tags 运营概览
// @Summary 获取运营概览数据
// @Description 获取运营概览数据
// @Router /admin.GetDashboardOverview [post]
// @Param request body dtodashboard.GetDashboardOverviewRequest true "request"
// @Success 200 {object} dtodashboard.GetDashboardOverviewResponse "response"
func GetDashboardOverview(ctx *gin.Context, req *dtodashboard.GetDashboardOverviewRequest, resp *dtodashboard.GetDashboardOverviewResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetDashboardOverview] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcdashboard.GetDashboardOverview(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetDashboardOverview] svcdashboard.GetDashboardOverview failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "admin_get_dashboard_overview_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
