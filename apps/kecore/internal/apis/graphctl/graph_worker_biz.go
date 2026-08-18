package graphctl

import (
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// ParseGraphRequest 解析图谱请求
type ParseGraphRequest struct {
	apiobj.BaseRequest
	Request struct {
		GraphID uint `json:"graph_id"`
	}
}

// Validity 校验解析图谱请求
func (req *ParseGraphRequest) Validity(resp *ParseGraphResponse) {
	if req.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择图谱
		return
	}
}

// ParseGraphResponse 解析图谱响应
type ParseGraphResponse struct {
	apiobj.BaseResponse
}

// GraphTaskCallbackRequest 图谱任务回调请求
type GraphTaskCallbackRequest struct {
	apiobj.BaseRequest
	Request *graph.GraphAlgoResp
}

// Validity 校验图谱任务回调请求
func (opt *GraphTaskCallbackRequest) Validity(resp *GraphTaskCallbackResponse) {
	if opt.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 知识图谱id不能为0
		return
	}
	if opt.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_file_id" // 文件id不能为0
		return
	}
	opt.Request.ReplaceStr()
}

// GraphTaskCallbackResponse 图谱任务回调响应
type GraphTaskCallbackResponse struct {
	apiobj.BaseResponse
}
