package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtomembership"
	"github.com/insmtx/corekg/apps/kecore/services/svcmembership"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// ListPackage 套餐列表
// @Tags 套餐管理
// @Summary 套餐列表
// @Description 套餐列表
// @Router /forest.ListPackage [post]
// @Param request body dtomembership.ListPackageRequest true "request"
// @Success 200 {object} dtomembership.ListPackageResponse "response"
func ListPackage(ctx *gin.Context, req *dtomembership.ListPackageRequest, resp *dtomembership.ListPackageResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListPackage] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcmembership.ListPackage(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListPackage] svcmembership.ListPackage failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_list_package_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
