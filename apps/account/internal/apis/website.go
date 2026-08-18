package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/internal/dto/dtowebsite"
	"github.com/insmtx/corekg/apps/account/services/svcwebsite"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// UpdateWebsiteInfo 更新网站信息
// @Tags 网站信息
// @Summary 更新网站信息
// @Description 更新网站信息
// @Router /account.UpdateWebsiteInfo [post]
// @Param request body dtowebsite.UpdateWebsiteInfoRequest true "request"
// @Success 200 {object} dtowebsite.UpdateWebsiteInfoResponse "response"
func UpdateWebsiteInfo(ctx *gin.Context, req *dtowebsite.UpdateWebsiteInfoRequest, resp *dtowebsite.UpdateWebsiteInfoResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[UpdateWebsiteInfo] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcwebsite.UpdateWebsiteInfo(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[UpdateWebsiteInfo] svcwebsite.UpdateWebsiteInfo failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = errcode.GetMessage(errcode.ErrCode_InternalError)
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
