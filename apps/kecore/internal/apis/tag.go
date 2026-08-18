package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtotag"
	"github.com/insmtx/corekg/apps/kecore/services/svctag"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// ListTag 标签列表
// @Tags 标签管理
// @Summary 标签列表
// @Description 标签列表
// @Router /kecore.ListResourceTag [post]
// @Param request body dtotag.ListTagRequest true "request"
// @Success 200 {object} dtotag.ListTagResponse "response"
func ListTag(ctx *gin.Context, req *dtotag.ListTagRequest, resp *dtotag.ListTagResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListTag] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svctag.ListTag(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListTag] svctag.ListTag failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_list_tag_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// CreateTag 创建标签
// @Tags 标签管理
// @Summary 创建标签
// @Description 创建标签
// @Router /kecore.CreateResourceTag [post]
// @Param request body dtotag.CreateTagRequest true "request"
// @Success 200 {object} dtotag.CreateTagResponse "response"
func CreateTag(ctx *gin.Context, req *dtotag.CreateTagRequest, resp *dtotag.CreateTagResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreateTag] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svctag.CreateTag(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateTag] svctag.CreateTag failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_tag_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ModifyTag 修改标签
// @Tags 标签管理
// @Summary 修改标签
// @Description 修改标签
// @Router /kecore.ModifyResourceTag [post]
// @Param request body dtotag.ModifyTagRequest true "request"
// @Success 200 {object} dtotag.ModifyTagResponse "response"
func ModifyTag(ctx *gin.Context, req *dtotag.ModifyTagRequest, resp *dtotag.ModifyTagResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ModifyTag] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svctag.ModifyTag(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ModifyTag] svctag.ModifyTag failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_modify_tag_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// DeleteTag 删除标签
// @Tags 标签管理
// @Summary 删除标签
// @Description 删除标签
// @Router /kecore.DeleteResourceTag [post]
// @Param request body dtotag.DeleteTagRequest true "request"
// @Success 200 {object} dtotag.DeleteTagResponse "response"
func DeleteTag(ctx *gin.Context, req *dtotag.DeleteTagRequest, resp *dtotag.DeleteTagResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[DeleteTag] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svctag.DeleteTag(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[DeleteTag] svctag.DeleteTag failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_delete_tag_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetTagTree 获取标签树
// @Tags 标签管理
// @Summary 获取标签树
// @Description 获取标签树
// @Router /kecore.GetTagTree [post]
// @Param request body dtotag.GetTagTreeRequest true "request"
// @Success 200 {object} dtotag.GetTagTreeResponse "response"
func GetTagTree(ctx *gin.Context, req *dtotag.GetTagTreeRequest, resp *dtotag.GetTagTreeResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetTagTree] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svctag.GetTagTree(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetTagTree] svctag.GetTagTree failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_tag_tree_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
