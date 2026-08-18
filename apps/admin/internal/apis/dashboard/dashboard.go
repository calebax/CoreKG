package dashboard

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/models/dashboard"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// GetDashboardData 获取数据看板
// @Tags 数据看板
// @Summary 获取数据看板
// @Description 获取数据看板
// @Router /admin.GetDashboardData [post]
// @Param user body GetDashboardDataRequest true "入参"
// @Success 200 {object} GetDashboardDataResponse "返回值"
func GetDashboardData(ctx *gin.Context, req *GetDashboardDataRequest, resp *GetDashboardDataResponse) {
	if req.Validate(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "GetDashboardData valid faild: %v", resp.Message)
		return
	}

	frs, err := dashboard.GetForestData(ctx, req.Request.CompanyID)
	if err != nil {
		logs.ErrorContextf(ctx, "dashboard.GetForestData failed: %v", err)
		runtime.InternalError(ctx, "获取知识库数据失败")
		return
	}
	doc, err := dashboard.GetDocData(ctx, req.Request.CompanyID)
	if err != nil {
		logs.ErrorContextf(ctx, "dashboard.GetDocData failed: %v", err)
		runtime.InternalError(ctx, "获取上传文件数据失败")
		return
	}

	parse, err := dashboard.GetParseData(ctx, req.Request.CompanyID)
	if err != nil {
		logs.ErrorContextf(ctx, "dashboard.GetParseData failed: %v", err)
		runtime.InternalError(ctx, "获取解析记录失败")
		return
	}

	session, err := dashboard.GetSessionData(ctx, req.Request.CompanyID)
	if err != nil {
		logs.ErrorContextf(ctx, "dashboard.GetSessionData failed: %v", err)
		runtime.InternalError(ctx, "获取会话列表失败")
		return
	}
	resp.Response.Forest = *frs
	resp.Response.File = *doc
	resp.Response.Parse = *parse
	resp.Response.Session = *session
}
