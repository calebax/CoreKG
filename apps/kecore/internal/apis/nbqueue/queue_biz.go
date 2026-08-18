package nbqueue

import (
	"github.com/insmtx/corekg/apps/kecore/models/nbgraph"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// GetForestWordCloudRequest 获取知识森林词云图请求
type GetForestWordCloudRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint `json:"forest_id"`
	}
}

// Validity 校验获取知识森林词云图请求
func (opt *GetForestWordCloudRequest) Validity(resp *GetForestWordCloudResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest" // 请选择知识森林
		return
	}
}

// GetForestWordCloudResponse 获取知识森林词云图响应
type GetForestWordCloudResponse struct {
	apiobj.BaseResponse
	Response struct {
		WordCloud []nbgraph.WordsCloud `json:"word_cloud"`
	}
}

// GetNodesByIDRequest 根据ID获取相连图请求
type GetNodesByIDRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint   `json:"forest_id"`
		NodeID   string `json:"node_id"`
	}
}

// Validity 校验根据ID获取相连图请求
func (opt *GetNodesByIDRequest) Validity(resp *GetNodesByIDResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest" // 请选择知识森林
		return
	}
	if opt.Request.NodeID == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_node_id" // 请传递正确节点ID
		return
	}
}

// GetNodesByIDResponse 根据ID获取相连图响应
type GetNodesByIDResponse struct {
	apiobj.BaseResponse
	Response struct {
		Graph *nbgraph.Graph `json:"graph"`
	}
}

// GetForestWordCloudGraphRequest 获取知识森林词云图对应知识图谱请求
type GetForestWordCloudGraphRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint `json:"forest_id"`
	}
}

// Validity 校验获取知识森林词云图对应知识图谱请求
func (opt *GetForestWordCloudGraphRequest) Validity(resp *GetForestWordCloudGraphResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest" // 请选择知识森林
		return
	}
}

// GetForestWordCloudGraphResponse 获取知识森林词云图对应知识图谱响应
type GetForestWordCloudGraphResponse struct {
	apiobj.BaseResponse
	Response struct {
		Graph *nbgraph.Graph `json:"graph"`
	}
}
