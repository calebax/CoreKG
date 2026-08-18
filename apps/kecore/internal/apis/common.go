package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtocommon"
	"github.com/insmtx/corekg/apps/kecore/services/svccommon"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// GetCommonInfo 获取公共信息
// @Tags 通用功能
// @Summary 获取公共信息
// @Description 获取公共信息
// @Router /forest.GetCommonInfo [post]
// @Param request body dtocommon.GetCommonInfoRequest true "request"
// @Success 200 {object} dtocommon.GetCommonInfoResponse "response"
func GetCommonInfo(ctx *gin.Context, req *dtocommon.GetCommonInfoRequest, resp *dtocommon.GetCommonInfoResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetCommonInfo] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svccommon.GetCommonInfo(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetCommonInfo] svccommon.GetCommonInfo failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_common_info_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
