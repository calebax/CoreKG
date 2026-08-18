package nbqueue

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/nbgraph"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// GetForestWordCloud 获取知识森林词云图
// @Tags 知识森林知识图谱
// @Summary 获取知识森林词云图
// @Description 获取知识森林词云图
// @Router /forest.GetForestWordCloud [post]
// @Param user body GetForestWordCloudRequest true "入参"
// @Success 200 {object} GetForestWordCloudResponse "返回值"
func GetForestWordCloud(ctx *gin.Context, req *GetForestWordCloudRequest, resp *GetForestWordCloudResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	w, err := nbgraph.NewWrapper(ctx, req.Request.ForestID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_wordcloud_init_failed" // 词云初始化失败
		return
	}
	defer w.Close()
	if err = w.DoWordCloudNql(); err != nil {
		if errors.Is(err, nbgraph.ErrEmptyResult) {
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_wordcloud_query_failed" // 执行词云查询失败
		return
	}
	if err = w.BuildWordCloud(); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_wordcloud_build_failed" // 构建词云操作失败
		return
	}

	resp.Response.WordCloud = w.Wc
}

// GetNodesByID 根据id获取获取相连的图
// @Tags 知识森林知识图谱
// @Summary 根据id获取获取相连的图
// @Description 根据id获取获取相连的图
// @Router /forest.GetNodesByID [post]
// @Param user body GetNodesByIDRequest true "入参"
// @Success 200 {object} GetNodesByIDResponse "返回值"
func GetNodesByID(ctx *gin.Context, req *GetNodesByIDRequest, resp *GetNodesByIDResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	if req.Request.ForestID == 109 {
		old, err := nbgraph.GetNodesByIDOld(ctx, 19, 109, req.Request.NodeID)
		if err != nil {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_nodes_query_empty" // 109查询结果为空
			return
		}
		resp.Response.Graph = old
		return
	}

	w, err := nbgraph.NewWrapper(ctx, req.Request.ForestID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_subgraph_init_failed" // 子图初始化失败
		return
	}
	defer w.Close()
	req.Request.NodeID = fmt.Sprintf("%v_%s", req.Request.ForestID, req.Request.NodeID)
	if err = w.DoGoFromIDNql(req.Request.NodeID); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_subgraph_query_failed" // 执行子图查询失败
		return
	}

	if err = w.BuildGraph(); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_subgraph_build_failed" // 构建子图失败
		return
	}
	resp.Response.Graph = w.G
}

// GetForestWordCloudGraph 获取知识森林词云图对应知识图谱
// @Tags 知识森林知识图谱
// @Summary 获取知识森林词云图对应知识图谱
// @Description 获取知识森林词云图对应知识图谱
// @Router /forest.GetForestWordCloudGraph [post]
// @Param user body GetForestWordCloudGraphRequest true "入参"
// @Success 200 {object} GetForestWordCloudGraphResponse "返回值"
func GetForestWordCloudGraph(ctx *gin.Context, req *GetForestWordCloudGraphRequest, resp *GetForestWordCloudGraphResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	w, err := nbgraph.NewWrapper(ctx, req.Request.ForestID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_graph_init_failed" // 图谱初始化失败
		return
	}
	defer w.Close()
	if err = w.GetWordCloudGraph(); err != nil {
		if errors.Is(err, nbgraph.ErrEmptyResult) {
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_graph_query_failed" // 执行图谱查询失败
		return
	}

	resp.Response.Graph = w.G
}
