package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtochart"
	"github.com/insmtx/corekg/apps/kechat/services/svcchart"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// BatchDeleteChart 批量删除图表
// @Tags 图表
// @Summary 批量删除图表
// @Description 批量删除图表
// @Router /chat.BatchDeleteChart [post]
// @Param request body dtochart.BatchDeleteChartRequest true "request"
// @Success 200 {object} dtochart.BatchDeleteChartResponse "response"
func BatchDeleteChart(ctx *gin.Context, req *dtochart.BatchDeleteChartRequest, resp *dtochart.BatchDeleteChartResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[BatchDeleteChart] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	// TODO: 需要手动注册路由
	res, err := svcchart.BatchDeleteChart(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[BatchDeleteChart] svcecharts.BatchDeleteChart failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_delete_chart_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
