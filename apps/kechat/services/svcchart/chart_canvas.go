package svcchart

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtochart"
	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

func SaveChartCanvas(ctx *gin.Context, req *dtochart.SaveChartCanvasRequest) (res *dtochart.SaveChartCanvasResponse, err error) {
	res = &dtochart.SaveChartCanvasResponse{}
	existEntity, err := chat.NewChatChartCanvasDao().GetByCond(ctx, &chat.ChatChartCanvasCond{
		SubjectID:   req.Request.SubjectID,
		SubjectType: req.Request.SubjectType,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[SaveChartCanvas] Failed to get chat chart canvas, subjectID: %d, subjectType: %s, err: %v", req.Request.SubjectID, req.Request.SubjectType, err)
		return res, err
	}
	if existEntity.ID > 0 {
		updateMap := map[string]any{
			"content": req.Request.Content,
		}
		if err := chat.NewChatChartCanvasDao().UpdateMap(ctx, existEntity.ID, updateMap); err != nil {
			logs.ErrorContextf(ctx, "[SaveChartCanvas] Failed to update chat chart canvas, id: %d, err: %v", existEntity.ID, err)
			return res, err
		}
		res.Response.CanvasID = existEntity.ID
		return res, nil
	}
	canvasEntity := &chattype.ChatChartCanvas{
		CompanyID:   runtime.CompanyID(ctx),
		Uin:         runtime.Uin(ctx),
		SubjectID:   req.Request.SubjectID,
		SubjectType: req.Request.SubjectType,
		Content:     req.Request.Content,
	}
	if err := chat.NewChatChartCanvasDao().Insert(ctx, canvasEntity); err != nil {
		logs.ErrorContextf(ctx, "[SaveChartCanvas] Failed to insert chat chart canvas, canvasEntity: %s, err: %v", logs.JSON(canvasEntity), err)
		return res, err
	}
	res.Response.CanvasID = canvasEntity.ID
	return res, nil
}

func GetChartCanvas(ctx *gin.Context, req *dtochart.GetChartCanvasRequest) (res *dtochart.GetChartCanvasResponse, err error) {
	res = &dtochart.GetChartCanvasResponse{}
	canvasEntity, err := chat.NewChatChartCanvasDao().GetByCond(ctx, &chat.ChatChartCanvasCond{
		SubjectType: req.Request.SubjectType,
		SubjectID:   req.Request.SubjectID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[GetChartCanvas] Failed to get chat chart canvas, subjectID: %d, subjectType: %s, err: %v", req.Request.SubjectID, req.Request.SubjectType, err)
		return res, err
	}
	if canvasEntity.ID == 0 {
		// 没有找到对应的画布，返回空值
		return res, nil
	}
	res.Response = dtochart.GetChartCanvasEmbedResponse{
		CanvasID:    canvasEntity.ID,
		SubjectType: canvasEntity.SubjectType,
		SubjectID:   canvasEntity.SubjectID,
		Content:     canvasEntity.Content,
	}
	return res, nil
}
