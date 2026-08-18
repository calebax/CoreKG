package dtochart

import (
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type SaveChartCanvasRequest struct {
	apiobj.BaseRequest
	Request SaveChartCanvasEmbedRequest
}

type SaveChartCanvasEmbedRequest struct {
	// SubjectID 主体 id，如项目 ID 或会话 ID
	SubjectID uint `json:"subject_id" validate:"required"`
	// SubjectType 主体类型,session：会话，project：项目
	SubjectType chattype.SessionSubjectType `json:"subject_type" validate:"required"`
	// Content 画布内容
	Content string `json:"content" validate:"required"`
}

func (opt *SaveChartCanvasRequest) Validity(resp *SaveChartCanvasResponse) {
	if opt.Request.SubjectID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_chartcanvas_subject_id_required"
	}
	if opt.Request.Content == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_chartcanvas_content_required"
	}
	if opt.Request.SubjectType == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_chartcanvas_subject_type_required"
	}
}

type GetChartCanvasRequest struct {
	apiobj.BaseRequest
	Request GetEChartsCanvasEmbedRequest
}
type GetEChartsCanvasEmbedRequest struct {
	// SubjectID 主体 id，如项目 ID 或会话 ID
	SubjectID uint `json:"subject_id" validate:"required"`
	// SubjectType 主体类型,session：会话，project：项目
	SubjectType chattype.SessionSubjectType `json:"subject_type" validate:"required"`
}

func (opt *GetChartCanvasRequest) Validity(resp *GetChartCanvasResponse) {
	if opt.Request.SubjectID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_chartcanvas_subject_id_required"
	}
	if opt.Request.SubjectType == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_chartcanvas_subject_type_required"
	}
}

type BatchDeleteChartRequest struct {
	apiobj.BaseRequest
	Request BatchDeleteChartEmbedRequest
}
type BatchDeleteChartEmbedRequest struct {
	// ChartIDs 图表 id 列表
	ChartIDs []uint `json:"chart_ids"`
}

func (opt *BatchDeleteChartRequest) Validity(resp *BatchDeleteChartResponse) {
	if len(opt.Request.ChartIDs) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_echarts_ids_required"
	}
}
