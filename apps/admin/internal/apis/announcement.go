package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/internal/dto/dtoannouncement"
	"github.com/insmtx/corekg/apps/admin/services/svcannouncement"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// ListAnnouncement 查询公告列表
// @Tags 系统公告
// @Summary 查询公告列表
// @Description 查询公告列表
// @Router /admin.ListAnnouncement [post]
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
		resp.Message = "admin_list_announcement_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetAnnouncement 获取公告详情
// @Tags 系统公告
// @Summary 获取公告详情
// @Description 获取公告详情
// @Router /admin.GetAnnouncement [post]
// @Param request body dtoannouncement.GetAnnouncementRequest true "request"
// @Success 200 {object} dtoannouncement.GetAnnouncementResponse "response"
func GetAnnouncement(ctx *gin.Context, req *dtoannouncement.GetAnnouncementRequest, resp *dtoannouncement.GetAnnouncementResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetAnnouncement] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcannouncement.GetAnnouncement(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetAnnouncement] svcannouncement.GetAnnouncement failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "admin_get_announcement_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ModifyAnnouncement 修改公告
// @Tags 系统公告
// @Summary 修改公告
// @Description 修改公告
// @Router /admin.ModifyAnnouncement [post]
// @Param request body dtoannouncement.ModifyAnnouncementRequest true "request"
// @Success 200 {object} dtoannouncement.ModifyAnnouncementResponse "response"
func ModifyAnnouncement(ctx *gin.Context, req *dtoannouncement.ModifyAnnouncementRequest, resp *dtoannouncement.ModifyAnnouncementResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ModifyAnnouncement] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcannouncement.ModifyAnnouncement(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ModifyAnnouncement] svcannouncement.ModifyAnnouncement failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "admin_modify_announcement_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// DeleteAnnouncement 删除公告
// @Tags 系统公告
// @Summary 删除公告
// @Description 删除公告
// @Router /admin.DeleteAnnouncement [post]
// @Param request body dtoannouncement.DeleteAnnouncementRequest true "request"
// @Success 200 {object} dtoannouncement.DeleteAnnouncementResponse "response"
func DeleteAnnouncement(ctx *gin.Context, req *dtoannouncement.DeleteAnnouncementRequest, resp *dtoannouncement.DeleteAnnouncementResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[DeleteAnnouncement] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcannouncement.DeleteAnnouncement(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[DeleteAnnouncement] svcannouncement.DeleteAnnouncement failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "admin_delete_announcement_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// CreateAnnouncement 创建公告
// @Tags 系统公告
// @Summary 创建公告
// @Description 创建公告
// @Router /admin.CreateAnnouncement [post]
// @Param request body dtoannouncement.CreateAnnouncementRequest true "request"
// @Success 200 {object} dtoannouncement.CreateAnnouncementResponse "response"
func CreateAnnouncement(ctx *gin.Context, req *dtoannouncement.CreateAnnouncementRequest, resp *dtoannouncement.CreateAnnouncementResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreateAnnouncement] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcannouncement.CreateAnnouncement(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateAnnouncement] svcannouncement.CreateAnnouncement failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "admin_create_announcement_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
