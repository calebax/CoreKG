package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoperm"
	"github.com/insmtx/corekg/apps/kecore/services/svcperm"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// SetResourcePerm 设置资源权限
// @Tags 资源权限
// @Summary 设置资源权限
// @Description 设置资源权限
// @Router /forest.SetResourcePerm [post]
// @Param request body dtoperm.SetResourcePermRequest true "request"
// @Success 200 {object} dtoperm.SetResourcePermResponse "response"
func SetResourcePerm(ctx *gin.Context, req *dtoperm.SetResourcePermRequest, resp *dtoperm.SetResourcePermResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[SetResourcePerm] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcperm.SetResourcePerm(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[SetResourcePerm] svcperm.SetResourcePerm failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_set_resource_perm_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetResourcePerm 获取资源权限
// @Tags 资源权限
// @Summary 获取资源权限
// @Description 获取资源权限
// @Router /forest.GetResourcePerm [post]
// @Param request body dtoperm.GetResourcePermRequest true "request"
// @Success 200 {object} dtoperm.GetResourcePermResponse "response"
func GetResourcePerm(ctx *gin.Context, req *dtoperm.GetResourcePermRequest, resp *dtoperm.GetResourcePermResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetResourcePerm] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcperm.GetResourcePerm(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetResourcePerm] svcperm.GetResourcePerm failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_resource_perm_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
