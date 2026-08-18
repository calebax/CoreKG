package dtographnode

import (
	"fmt"

	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type CreateNodeRequest struct {
	apiobj.BaseRequest
	Request CreateNodeEmbedRequest
}

type CreateNodeEmbedRequest struct {
	GraphID uint `json:"graph_id"`
	// 实体名称
	NodeName string `json:"node_name"`
	// 实体类型
	Tags []TagObject `json:"tags"`
	// 实体关系
	Edges []EdgeObject `json:"edges"`
}

const MaxPropertieNameLen = 50

func (opt *CreateNodeRequest) Validity(resp *CreateNodeResponse) {
	if opt.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_id_empty"
		return
	}
	if opt.Request.NodeName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_name_empty"
		return
	}
	if len(opt.Request.Tags) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tags_empty"
		return
	}
	// 校验每条边的起点或终点必须包含当前节点名
	for i, edge := range opt.Request.Edges {
		if edge.SrcNodeName != opt.Request.NodeName && edge.DstNodeName != opt.Request.NodeName {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = fmt.Sprintf("kecore_edge_%d_not_contains_current_node", i)
			return
		}
	}

	for _, tag := range opt.Request.Tags {
		for _, property := range tag.Properties {
			if len([]rune(property.Name)) > MaxPropertieNameLen {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = fmt.Sprintf("属性名超长: %s", property.Name)
				return
			}
		}
	}
}

type GetNodeEdgesRequest struct {
	apiobj.BaseRequest
	Request GetNodeEdgesEmbedRequest
}
type GetNodeEdgesEmbedRequest struct {
	// 图id
	GraphID uint `json:"graph_id"`
	// 节点id
	NodeID uint `json:"node_id"`
	// 节点名称
	NodeName string `json:"node_name"`
	// TagID
	TagID uint `json:"tag_id"`
}

func (opt *GetNodeEdgesRequest) Validity(resp *GetNodeEdgesResponse) {
	if len(opt.Request.NodeName) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_name_empty"
		return
	}
	if opt.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_id_empty"
		return
	}
	if opt.Request.TagID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_id_empty"
		return
	}
}

type EditNodeRequest struct {
	apiobj.BaseRequest
	Request EditNodeEmbedRequest
}
type EditNodeEmbedRequest struct {
	GraphID uint `json:"graph_id"`
	// 旧实体名称
	OldNodeName string `json:"old_node_name"`
	// 实体类型
	Tags []TagObject `json:"tags"`
	// 实体关系
	Edges []EdgeObject `json:"edges"`
}

func (opt *EditNodeRequest) Validity(resp *EditNodeResponse) {
	if opt.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_id_empty"
		return
	}
	if len(opt.Request.OldNodeName) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_old_node_name_empty"
		return
	}
	for i, edge := range opt.Request.Edges {
		if edge.SrcNodeName != opt.Request.OldNodeName {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = fmt.Sprintf("kecore_edge_%d_src_not_current_node", i)
			return
		}
	}

	for _, tag := range opt.Request.Tags {
		for _, property := range tag.Properties {
			if len([]rune(property.Name)) > MaxPropertieNameLen {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = fmt.Sprintf("属性名超长: %s", property.Name)
				return
			}
		}
	}
}

type CreateNodeEdgeRequest struct {
	apiobj.BaseRequest
	Request CreateNodeEdgeEmbedRequest
}
type CreateNodeEdgeEmbedRequest struct {
	GraphID uint       `json:"graph_id"`
	Edge    EdgeObject `json:"edge"`
}

func (opt *CreateNodeEdgeRequest) Validity(resp *CreateNodeEdgeResponse) {
	if opt.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_id_empty"
		return
	}
	if len(opt.Request.Edge.EdgeName) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_name_empty"
		return
	}
	if opt.Request.Edge.SrcNodeName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_src_node_name_empty"
		return
	}
	if opt.Request.Edge.DstNodeName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_dst_node_name_empty"
		return
	}
}

type DeleteNodeRequest struct {
	apiobj.BaseRequest
	Request DeleteNodeEmbedRequest
}
type DeleteNodeEmbedRequest struct {
	// 图id
	GraphID uint `json:"graph_id"`
	// 节点名称
	NodeName string `json:"node_name"`
	// TagID
	TagID uint `json:"tag_id"`
}

func (opt *DeleteNodeRequest) Validity(resp *DeleteNodeResponse) {
	if opt.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_id_empty"
		return
	}
	if len(opt.Request.NodeName) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_name_empty"
		return
	}
	// TagID 可选，如果为0则删除该节点名称的所有tag节点
}

type GetNodeReferenceRequest struct {
	apiobj.BaseRequest
	Request GetNodeReferenceEmbedRequest
}
type GetNodeReferenceEmbedRequest struct {
	// GraphID 图谱id
	GraphID uint `json:"graph_id"`
	// NodeName 节点名
	NodeName string `json:"node_name"`
	// TagID
	TagID uint `json:"tag_id"`
}

func (opt *GetNodeReferenceRequest) Validity(resp *GetNodeReferenceResponse) {
	if opt.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_id_empty"
		return
	}

	if len(opt.Request.NodeName) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_name_empty"
		return
	}
	if opt.Request.TagID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_id_empty"
		return
	}
}

type GetGraphEdgesRequest struct {
	apiobj.BaseRequest
	Request GetGraphEdgesEmbedRequest
}
type GetGraphEdgesEmbedRequest struct {
	GraphID uint `json:"graph_id"`
}

func (opt *GetGraphEdgesRequest) Validity(resp *GetGraphEdgesResponse) {
	if opt.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_id_empty"
		return
	}
}

type RenameNodeRequest struct {
	apiobj.BaseRequest
	Request RenameNodeEmbedRequest
}
type RenameNodeEmbedRequest struct {
	GraphID uint `json:"graph_id"`
	// 旧节点名称
	OldNodeName string `json:"old_node_name"`
	// TagID
	TagID uint `json:"tag_id"`
	// 新节点名称
	NodeName string `json:"node_name"`
}

func (opt *RenameNodeRequest) Validity(resp *RenameNodeResponse) {
	if opt.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_id_empty"
		return
	}
	if len(opt.Request.OldNodeName) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_old_node_name_empty"
		return
	}
	if len(opt.Request.NodeName) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_name_empty"
		return
	}
	if opt.Request.TagID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_id_empty"
		return
	}
	if opt.Request.OldNodeName == opt.Request.NodeName {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_name_same_as_old_name"
		return
	}
}
