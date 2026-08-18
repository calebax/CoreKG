package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtotag"
	"github.com/insmtx/corekg/apps/kecore/services/svctag"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// SetResourceTag 设置资源标签
// @Tags 标签管理
// @Summary 设置资源标签
// @Description 设置资源标签
// @Router /kecore.SetResourceTag [post]
// @Param request body dtotag.SetResourceTagRequest true "request"
// @Success 200 {object} dtotag.SetResourceTagResponse "response"
func SetResourceTag(ctx *gin.Context, req *dtotag.SetResourceTagRequest, resp *dtotag.SetResourceTagResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[SetResourceTag] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svctag.SetResourceTag(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[SetResourceTag] svctag.SetResourceTag failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_set_resource_tag_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
