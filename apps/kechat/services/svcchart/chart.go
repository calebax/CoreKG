package svcchart

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtochart"
	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/ygpkg/yg-go/logs"
)

func BatchDeleteChart(ctx *gin.Context, req *dtochart.BatchDeleteChartRequest) (res *dtochart.BatchDeleteChartResponse, err error) {
	res = &dtochart.BatchDeleteChartResponse{}
	if err := chat.NewChatChartDao().DeleteByIDs(ctx, req.Request.ChartIDs); err != nil {
		logs.ErrorContextf(ctx, "[BatchDeleteChart] Failed to delete chat chart, chartIDs: %v, err: %v", req.Request.ChartIDs, err)
		return res, err
	}
	res.Response = dtochart.BatchDeleteChartEmbedResponse{
		ChartIDs: req.Request.ChartIDs,
	}
	return res, nil
}
