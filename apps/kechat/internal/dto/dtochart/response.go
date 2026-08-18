package dtochart

import (
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type SaveChartCanvasResponse struct {
	apiobj.BaseResponse
	Response SaveChartCanvasEmbedResponse
}

type SaveChartCanvasEmbedResponse struct {
	// CanvasID 画布 ID
	CanvasID uint `json:"canvas_id"`
}

type GetChartCanvasResponse struct {
	apiobj.BaseResponse
	Response GetChartCanvasEmbedResponse
}
type GetChartCanvasEmbedResponse struct {
	// CanvasID 画布 ID
	CanvasID uint `json:"canvas_id"`
	// SubjectID 主体 id，如项目 ID
	SubjectID uint `json:"subject_id" validate:"required"`
	// SubjectType 主体类型
	SubjectType chattype.SessionSubjectType `json:"subject_type"`
	// Content 画布内容
	Content string `json:"content" validate:"required"`
}

type BatchDeleteChartResponse struct {
	apiobj.BaseResponse
	Response BatchDeleteChartEmbedResponse
}
type BatchDeleteChartEmbedResponse struct {
	// ChartIDs 图表 id 列表
	ChartIDs []uint `json:"chart_ids"`
}
