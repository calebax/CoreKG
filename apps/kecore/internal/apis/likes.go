package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtolikes"
	"github.com/insmtx/corekg/apps/kecore/services/svclikes"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// ListLikes 获取点赞列表
// @Tags 点赞管理
// @Summary 获取点赞列表
// @Description 获取点赞列表
// @Router /kecore.ListLikes [post]
// @Param request body dtolikes.ListLikesRequest true "request"
// @Success 200 {object} dtolikes.ListLikesResponse "response"
func ListLikes(ctx *gin.Context, req *dtolikes.ListLikesRequest, resp *dtolikes.ListLikesResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListLikes] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svclikes.ListLikes(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListLikes] svclikes.ListLikes failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_list_likes_fail"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// MarkResourceLike 标记资源点赞
// @Tags 点赞管理
// @Summary 标记资源点赞
// @Description 标记资源点赞
// @Router /kecore.MarkResourceLike [post]
// @Param request body dtolikes.MarkResourceLikeRequest true "request"
// @Success 200 {object} dtolikes.MarkResourceLikeResponse "response"
func MarkResourceLike(ctx *gin.Context, req *dtolikes.MarkResourceLikeRequest, resp *dtolikes.MarkResourceLikeResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[MarkResourceLike] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svclikes.MarkResourceLike(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[MarkResourceLike] svclikes.MarkResourceLike failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_mark_resource_like_fail"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
