package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtographnode"
	"github.com/insmtx/corekg/apps/kecore/services/svcgraphnode"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// CreateNode 创建实体
// @Tags 图节点
// @Summary 创建实体
// @Description 创建实体。入参：实体类型/实体名称/实体属性/实体关系，返回值：节点id
// @Router /forest.CreateNode [post]
// @Param request body dtographnode.CreateNodeRequest true "request"
// @Success 200 {object} dtographnode.CreateNodeResponse "response"
func CreateNode(ctx *gin.Context, req *dtographnode.CreateNodeRequest, resp *dtographnode.CreateNodeResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreateNode] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcgraphnode.CreateNode(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateNode] svcgraphnode.CreateNode failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_node_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetNodeEdges 获取实体的边
// @Tags 图节点
// @Summary 获取实体的边
// @Description 获取实体的边
// @Router /forest.GetNodeEdges [post]
// @Param request body dtographnode.GetNodeEdgesRequest true "request"
// @Success 200 {object} dtographnode.GetNodeEdgesResponse "response"
func GetNodeEdges(ctx *gin.Context, req *dtographnode.GetNodeEdgesRequest, resp *dtographnode.GetNodeEdgesResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetNodeEdges] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcgraphnode.GetNodeEdges(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetNodeEdges] svcgraphnode.GetNodeEdges failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_node_edges_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// EditNode 编辑实体
// @Tags 图节点
// @Summary 编辑实体
// @Description 编辑实体。入参：id以及创建所需参数（全量）
// @Router /forest.EditNode [post]
// @Param request body dtographnode.EditNodeRequest true "request"
// @Success 200 {object} dtographnode.EditNodeResponse "response"
func EditNode(ctx *gin.Context, req *dtographnode.EditNodeRequest, resp *dtographnode.EditNodeResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[EditNode] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcgraphnode.EditNode(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[EditNode] svcgraphnode.EditNode failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_edit_node_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// CreateNodeEdge
// @Tags 图节点
// @Summary
// @Description 编辑实体之间的关系。入参：名称/起点id/终点id
// @Router /forest.CreateNodeEdge [post]
// @Param request body dtographnode.CreateNodeEdgeRequest true "request"
// @Success 200 {object} dtographnode.CreateNodeEdgeResponse "response"
func CreateNodeEdge(ctx *gin.Context, req *dtographnode.CreateNodeEdgeRequest, resp *dtographnode.CreateNodeEdgeResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreateNodeEdge] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcgraphnode.CreateNodeEdge(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateNodeEdge] svcgraphnode.CreateNodeEdge failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_node_edge_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// DeleteNode 删除实体
// @Tags 图节点
// @Summary 删除实体
// @Description 删除实体。入参：id
// @Router /forest.DeleteNode [post]
// @Param request body dtographnode.DeleteNodeRequest true "request"
// @Success 200 {object} dtographnode.DeleteNodeResponse "response"
func DeleteNode(ctx *gin.Context, req *dtographnode.DeleteNodeRequest, resp *dtographnode.DeleteNodeResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[DeleteNode] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcgraphnode.DeleteNode(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[DeleteNode] svcgraphnode.DeleteNode failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_delete_node_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetNodeReference 获取包含节点的文件以及相关chunk
// @Tags 图节点
// @Summary 获取包含节点的文件以及相关chunk
// @Description 获取包含节点的文件以及相关chunk
// @Router /forest.GetNodeReference [post]
// @Param request body dtographnode.GetNodeReferenceRequest true "request"
// @Success 200 {object} dtographnode.GetNodeReferenceResponse "response"
func GetNodeReference(ctx *gin.Context, req *dtographnode.GetNodeReferenceRequest, resp *dtographnode.GetNodeReferenceResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetNodeReference] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcgraphnode.GetNodeReference(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetNodeReference] svcgraphnode.GetNodeReference failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_node_reference_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetGraphEdges 获取图谱所有边类型
// @Tags 图节点
// @Summary 获取图谱所有边类型
// @Description 获取图谱所有边类型
// @Router /forest.GetGraphEdges [post]
// @Param request body dtographnode.GetGraphEdgesRequest true "request"
// @Success 200 {object} dtographnode.GetGraphEdgesResponse "response"
func GetGraphEdges(ctx *gin.Context, req *dtographnode.GetGraphEdgesRequest, resp *dtographnode.GetGraphEdgesResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetGraphEdges] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcgraphnode.GetGraphEdges(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetGraphEdges] svcgraphnode.GetGraphEdges failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_edges_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// RenameNode 重命名节点
// @Tags 图节点
// @Summary 重命名节点
// @Description 重命名节点
// @Router /forest.RenameNode [post]
// @Param request body dtographnode.RenameNodeRequest true "request"
// @Success 200 {object} dtographnode.RenameNodeResponse "response"
func RenameNode(ctx *gin.Context, req *dtographnode.RenameNodeRequest, resp *dtographnode.RenameNodeResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[RenameNode] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcgraphnode.RenameNode(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[RenameNode] svcgraphnode.RenameNode failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_rename_node_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
