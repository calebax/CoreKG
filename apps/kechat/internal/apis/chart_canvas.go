package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtochart"
	"github.com/insmtx/corekg/apps/kechat/services/svcchart"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// SaveChartCanvas 保存图表画布
// @Tags 图表
// @Summary 保存图表画布
// @Description 保存图表画布
// @Router /chat.SaveChartCanvas [post]
// @Param request body dtochart.SaveChartCanvasRequest true "request"
// @Success 200 {object} dtochart.SaveChartCanvasResponse "response"
func SaveChartCanvas(ctx *gin.Context, req *dtochart.SaveChartCanvasRequest, resp *dtochart.SaveChartCanvasResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[SaveChartCanvas] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	// TODO: 需要手动注册路由
	res, err := svcchart.SaveChartCanvas(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[SaveEChartsCanvas] svcchart.SaveEChartsCanvas failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_save_chart_canvas_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetChartCanvas 获取图表画布
// @Tags 图表
// @Summary 获取图表画布
// @Description 获取图表画布
// @Router /chat.GetChartCanvas [post]
// @Param request body dtochart.GetChartCanvasRequest true "request"
// @Success 200 {object} dtochart.GetChartCanvasResponse "response"
func GetChartCanvas(ctx *gin.Context, req *dtochart.GetChartCanvasRequest, resp *dtochart.GetChartCanvasResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetChartCanvas] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	// TODO: 需要手动注册路由
	res, err := svcchart.GetChartCanvas(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetChartCanvas] svcchart.GetChartCanvas failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_chart_canvas_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
