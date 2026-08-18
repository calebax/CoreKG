package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kesearch/internal/dto/dtoreranksearch"
	"github.com/insmtx/corekg/apps/kesearch/services/svcreranksearch"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// RerankSearchChunk 知识库检索rerank版本
// @Tags 知识库检索
// @Summary 知识库检索rerank版本
// @Description 知识库检索rerank版本
// @Router /kesearch.RerankSearchChunk [post]
// @Param request body dtoreranksearch.RerankSearchChunkRequest true "request"
// @Success 200 {object} dtoreranksearch.RerankSearchChunkResponse "response"
func RerankSearchChunk(ctx *gin.Context, req *dtoreranksearch.RerankSearchChunkRequest, resp *dtoreranksearch.RerankSearchChunkResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[RerankSearchChunk] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	res, err := svcreranksearch.RerankSearchChunk(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[RerankSearchChunk] svcreranksearch.RerankSearchChunk failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_internal_error"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
