package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoresource"
	"github.com/insmtx/corekg/apps/kecore/services/svcresource"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// SetResourceScope 设置资源权限范围
// @Tags 资源权限管理
// @Summary 设置资源权限范围
// @Description 设置资源权限范围
// @Router /forest.SetResourceScope [post]
// @Param request body dtoresource.SetResourceScopeRequest true "request"
// @Success 200 {object} dtoresource.SetResourceScopeResponse "response"
func SetResourceScope(ctx *gin.Context, req *dtoresource.SetResourceScopeRequest, resp *dtoresource.SetResourceScopeResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[SetResourceScope] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcresource.SetResourceScope(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[SetResourceScope] svcresource.SetResourceScope failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_set_resource_scope_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetResourceScope 获取资源权限范围
// @Tags 资源权限管理
// @Summary 获取资源权限范围
// @Description 获取资源权限范围
// @Router /forest.GetResourceScope [post]
// @Param request body dtoresource.GetResourceScopeRequest true "request"
// @Success 200 {object} dtoresource.GetResourceScopeResponse "response"
func GetResourceScope(ctx *gin.Context, req *dtoresource.GetResourceScopeRequest, resp *dtoresource.GetResourceScopeResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetResourceScope] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcresource.GetResourceScope(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetResourceScope] svcresource.GetResourceScope failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_resource_scope_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
