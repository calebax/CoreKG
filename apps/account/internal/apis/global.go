package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/internal/dto/dtoglobal"
	"github.com/insmtx/corekg/apps/account/services/svcglobal"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// GetGlobalInfo 获取全局通用数据
// @Tags 全局通用
// @Summary 获取全局通用数据
// @Description 获取全局通用数据
// @Router /account.GetGlobalInfo [post]
// @Param request body dtoglobal.GetGlobalInfoRequest true "request"
// @Success 200 {object} dtoglobal.GetGlobalInfoResponse "response"
func GetGlobalInfo(ctx *gin.Context, req *dtoglobal.GetGlobalInfoRequest, resp *dtoglobal.GetGlobalInfoResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetGlobalInfo] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcglobal.GetGlobalInfo(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetGlobalInfo] svcglobal.GetGlobalInfo failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_global_info_failed"
		return
	}
	resp.Code = errcode.CodeOK
	resp.Message = "account_get_global_info_success"
	resp.Response = res.Response
}
