package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtotag"
	"github.com/insmtx/corekg/apps/kecore/services/svctag"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// CreateTagGroup 创建标签分组
// @Tags 标签管理
// @Summary 创建标签分组
// @Description 创建标签分组
// @Router /kecore.CreateTagGroup [post]
// @Param request body dtotag.CreateTagGroupRequest true "request"
// @Success 200 {object} dtotag.CreateTagGroupResponse "response"
func CreateTagGroup(ctx *gin.Context, req *dtotag.CreateTagGroupRequest, resp *dtotag.CreateTagGroupResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreateTagGroup] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svctag.CreateTagGroup(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateTagGroup] svctag.CreateTagGroup failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_tag_group_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ModifyTagGroup 修改标签分组
// @Tags 标签管理
// @Summary 修改标签分组
// @Description 修改标签分组
// @Router /kecore.ModifyTagGroup [post]
// @Param request body dtotag.ModifyTagGroupRequest true "request"
// @Success 200 {object} dtotag.ModifyTagGroupResponse "response"
func ModifyTagGroup(ctx *gin.Context, req *dtotag.ModifyTagGroupRequest, resp *dtotag.ModifyTagGroupResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ModifyTagGroup] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svctag.ModifyTagGroup(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ModifyTagGroup] svctag.ModifyTagGroup failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_modify_tag_group_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// DeleteTagGroup 删除标签分组
// @Tags 标签管理
// @Summary 删除标签分组
// @Description 删除标签分组
// @Router /kecore.DeleteTagGroup [post]
// @Param request body dtotag.DeleteTagGroupRequest true "request"
// @Success 200 {object} dtotag.DeleteTagGroupResponse "response"
func DeleteTagGroup(ctx *gin.Context, req *dtotag.DeleteTagGroupRequest, resp *dtotag.DeleteTagGroupResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[DeleteTagGroup] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svctag.DeleteTagGroup(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[DeleteTagGroup] svctag.DeleteTagGroup failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_delete_tag_group_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ListTagGroup 标签分组列表
// @Tags 标签管理
// @Summary 标签分组列表
// @Description 标签分组列表
// @Router /kecore.ListTagGroup [post]
// @Param request body dtotag.ListTagGroupRequest true "request"
// @Success 200 {object} dtotag.ListTagGroupResponse "response"
func ListTagGroup(ctx *gin.Context, req *dtotag.ListTagGroupRequest, resp *dtotag.ListTagGroupResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListTagGroup] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svctag.ListTagGroup(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListTagGroup] svctag.ListTagGroup failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_list_tag_group_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
