package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoannouncement"
	"github.com/insmtx/corekg/apps/kecore/services/svcannouncement"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// ListAnnouncement 获取公告列表
// @Tags 系统公告
// @Summary 获取公告列表
// @Description 获取公告列表
// @Router /forest.ListAnnouncement [post]
// @Param request body dtoannouncement.ListAnnouncementRequest true "request"
// @Success 200 {object} dtoannouncement.ListAnnouncementResponse "response"
func ListAnnouncement(ctx *gin.Context, req *dtoannouncement.ListAnnouncementRequest, resp *dtoannouncement.ListAnnouncementResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListAnnouncement] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}


	res, err := svcannouncement.ListAnnouncement(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListAnnouncement] svcannouncement.ListAnnouncement failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = errcode.GetMessage(errcode.ErrCode_InternalError)
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
