package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoforest"
	"github.com/insmtx/corekg/apps/kecore/services/svcforest"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// UpdateForestDescription 更新知识库描述
// @Tags 知识库
// @Summary 更新知识库描述
// @Description 更新知识库描述
// @Router /forest.UpdateForestDescription [post]
// @Param request body dtoforest.UpdateForestDescriptionRequest true "request"
// @Success 200 {object} dtoforest.UpdateForestDescriptionResponse "response"
func UpdateForestDescription(ctx *gin.Context, req *dtoforest.UpdateForestDescriptionRequest, resp *dtoforest.UpdateForestDescriptionResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[UpdateForestDescription] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcforest.UpdateForestDescription(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[UpdateForestDescription] svcforest.UpdateForestDescription failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = errcode.GetMessage(errcode.ErrCode_InternalError)
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetOriginResource 获取远程资源
// @Tags 知识库
// @Summary 获取远程资源
// @Description 获取远程资源
// @Router /forest.GetOriginResource [post]
// @Param request body dtoforest.GetOriginResourceRequest true "request"
// @Success 200 {object} dtoforest.GetOriginResourceResponse "response"
func GetOriginResource(ctx *gin.Context, req *dtoforest.GetOriginResourceRequest, resp *dtoforest.GetOriginResourceResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetOriginResource] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	
	res, err := svcforest.GetOriginResource(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetOriginResource] svcforest.GetOriginResource failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_origin_resource_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
