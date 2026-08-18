package dtographnode

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type CreateNodeResponse struct {
	apiobj.BaseResponse
	Response CreateNodeEmbedResponse
}

type CreateNodeEmbedResponse struct {
	NodeID uint `json:"node_id"`
}

type GetNodeEdgesResponse struct {
	apiobj.BaseResponse
	Response GetNodeEdgesEmbedResponse
}
type GetNodeEdgesEmbedResponse struct {
	Edges []EdgeObject `json:"edges"`
}

type EditNodeResponse struct {
	apiobj.BaseResponse
	Response EditNodeEmbedResponse
}
type EditNodeEmbedResponse struct {
}

type CreateNodeEdgeResponse struct {
	apiobj.BaseResponse
	Response CreateNodeEdgeEmbedResponse
}
type CreateNodeEdgeEmbedResponse struct {
	EdgeID uint `json:"edge_id"`
}

type DeleteNodeResponse struct {
	apiobj.BaseResponse
	Response DeleteNodeEmbedResponse
}
type DeleteNodeEmbedResponse struct {
}

type GetNodeReferenceResponse struct {
	apiobj.BaseResponse
	Response GetNodeReferenceEmbedResponse
}
type GetNodeReferenceEmbedResponse struct {
	NodeName string       `json:"node_name"`
	Tags     []TagObject  `json:"tags,omitempty"`
	Edges    []EdgeObject `json:"edges,omitempty"`
	Files    []FileObject `json:"files,omitempty"`
}

type GetGraphEdgesResponse struct {
	apiobj.BaseResponse
	Response GetGraphEdgesEmbedResponse
}
type GetGraphEdgesEmbedResponse struct {
	Edges []EdgeObject `json:"edges"`
}

type RenameNodeResponse struct {
	apiobj.BaseResponse
	Response RenameNodeEmbedResponse
}
type RenameNodeEmbedResponse struct {
}
