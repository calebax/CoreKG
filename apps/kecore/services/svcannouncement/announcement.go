package svcannouncement

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoannouncement"
	"github.com/insmtx/corekg/pkgs/platform/announcement"
	"github.com/ygpkg/yg-go/logs"
)

func ListAnnouncement(ctx *gin.Context, req *dtoannouncement.ListAnnouncementRequest) (res *dtoannouncement.ListAnnouncementResponse, err error) {
	res = &dtoannouncement.ListAnnouncementResponse{}

	as, i, err := admin.NewAdminAnnouncementDao().GetPageListByCond(ctx, &admin.AnnouncementCond{
		BaseCond: admin.BaseCond{
			Limit:     req.Request.Limit,
			Offset:    req.Request.Offset,
			BeginTime: req.Request.BeginTime,
			EndTime:   req.Request.EndTime,
			OrderBy:   req.Request.OrderBy,
		},
		Filters: req.Request.Filters,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "ListAnnouncement failed err: %v", err)
		return res, err
	}
	res.Response.Limit = req.Request.Limit
	res.Response.Offset = req.Request.Offset
	res.Response.Total = i
	res.Response.Data = as

	return res, nil
}
