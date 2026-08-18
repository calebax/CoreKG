package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtocollections"
	"github.com/insmtx/corekg/apps/kecore/services/svccollections"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// MarkResourceCollection 标记资源收藏
// @Tags 收藏管理
// @Summary 标记资源收藏
// @Description 标记资源收藏
// @Router /kecore.MarkResourceCollection [post]
// @Param request body dtocollections.MarkResourceCollectionRequest true "request"
// @Success 200 {object} dtocollections.MarkResourceCollectionResponse "response"
func MarkResourceCollection(ctx *gin.Context, req *dtocollections.MarkResourceCollectionRequest, resp *dtocollections.MarkResourceCollectionResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[MarkResourceCollection] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svccollections.MarkResourceCollection(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[MarkResourceCollection] svccollections.MarkResourceCollection failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_mark_resource_collection_fail"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ListCollection 获取收藏列表
// @Tags 收藏管理
// @Summary 获取收藏列表
// @Description 获取收藏列表
// @Router /kecore.ListCollection [post]
// @Param request body dtocollections.ListCollectionRequest true "request"
// @Success 200 {object} dtocollections.ListCollectionResponse "response"
func ListCollection(ctx *gin.Context, req *dtocollections.ListCollectionRequest, resp *dtocollections.ListCollectionResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListCollection] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svccollections.ListCollection(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListCollection] svccollections.ListCollection failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_list_collection_fail"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
