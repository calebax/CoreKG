package svcannouncement

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/admin/internal/dto/dtoannouncement"
	"github.com/insmtx/corekg/apps/admin/models/admin"
	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/services/messagecenter"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
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

func GetAnnouncement(ctx *gin.Context, req *dtoannouncement.GetAnnouncementRequest) (res *dtoannouncement.GetAnnouncementResponse, err error) {
	res = &dtoannouncement.GetAnnouncementResponse{}
	as, err := admin.NewAdminAnnouncementDao().GetByID(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetByID(%v) faild err: %v", req.Request.ID, err)
		return res, err
	}
	if as.ID == 0 {
		logs.ErrorContextf(ctx, "GetByID(%v) faild : this record not exist", req.Request.ID)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "admin_get_announcement_failed"
		return res, nil
	}
	res.Response.Announcement = admin.Announcement{
		ID:        as.ID,
		CreatedAt: as.CreatedAt,
		Uin:       as.Uin,
		CompanyID: as.CompanyID,
		Creator:   as.Creator,
		Tag:       as.Tag,
		Content:   as.Content,
	}
	return res, nil
}

func ModifyAnnouncement(ctx *gin.Context, req *dtoannouncement.ModifyAnnouncementRequest) (res *dtoannouncement.ModifyAnnouncementResponse, err error) {
	res = &dtoannouncement.ModifyAnnouncementResponse{}
	as, err := admin.NewAdminAnnouncementDao().GetByID(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetByID(%v) faild err: %v", req.Request.ID, err)
		return res, err
	}
	if as.ID == 0 {
		logs.ErrorContextf(ctx, "GetByUD(%v) faild : this record not exist", req.Request.ID)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "admin_get_announcement_failed"
		return res, nil
	}

	as.Tag = req.Request.Tag
	as.Content = req.Request.Content

	if err = dbutil.Account().Save(&as).Error; err != nil {
		logs.ErrorContextf(ctx, "ModifyAnnouncement save announcement faild err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "admin_save_announcement_failed"
		return res, err
	}

	return res, nil
}

func DeleteAnnouncement(ctx *gin.Context, req *dtoannouncement.DeleteAnnouncementRequest) (res *dtoannouncement.DeleteAnnouncementResponse, err error) {
	res = &dtoannouncement.DeleteAnnouncementResponse{}
	if err = admin.NewAdminAnnouncementDao().Delete(ctx, req.Request.ID); err != nil {
		logs.ErrorContextf(ctx, "DeleteAnnouncement(id:%v) faild", req.Request.ID)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "admin_delete_announcement_failed"
		return res, err
	}
	return res, nil
}

type UserUin struct {
	UserID    uint `json:"user_id"`
	Uin       uint `json:"uin"`
	CompanyID uint `json:"company_id"`
}

func CreateAnnouncement(ctx *gin.Context, req *dtoannouncement.CreateAnnouncementRequest) (res *dtoannouncement.CreateAnnouncementResponse, err error) {
	res = &dtoannouncement.CreateAnnouncementResponse{}
	uin := runtime.Uin(ctx)
	ui, err := user.GetUserIdentificationByUIN(ctx, uin)
	if err != nil {
		logs.ErrorContextf(ctx, "GetUserIdentificationByUIN(id:%v) failed err: %v", uin, err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "account_get_uin_failed"
		return res, err
	}

	ans := &admintype.AdminAnnouncement{
		Uin:     ui.ID,
		Creator: ui.Name,
		Tag:     req.Request.Tag,
		Content: req.Request.Content,
	}

	if err = dbutil.Account().WithContext(ctx).Create(ans).Error; err != nil {
		logs.ErrorContextf(ctx, "CreateAnnouncement failed err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "admin_create_announcement_failed"
		return res, err
	}
	// send message to all users

	var us []UserUin
	if err := dbutil.Account().Table(accounttype.TableNameUserIdentification+" ui").
		Unscoped().
		Select("u.id as user_id, ui.id as uin,ui.subject_id as company_id").
		Where("ui.deleted_at is null").
		Where("ui.uin_status = ?", accounttype.UinStatusNormal).
		Where("ui.subject_type = ?", accounttype.SubjectTypeCompany).
		Where("ui.issuer = ?", global.IssuerYYGU).
		Joins("INNER JOIN user u on u.id = ui.user_id AND u.deleted_at is null ").
		Find(&us).
		Error; err != nil {
		logs.ErrorContextf(ctx, "GetUserIdentificationByUIN(id:%v) failed err: %v", uin, err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "account_get_uin_failed"
		return res, err
	}

	if len(us) == 0 {
		logs.WarnContextf(ctx, "CreateAnnouncement target no user found")
		return res, nil
	}

	var reqs []*messagecenter.SendMessageReq
	for _, v := range us {
		reqs = append(reqs, &messagecenter.SendMessageReq{
			CompanyID:    v.CompanyID,
			UserID:       v.UserID,
			Uin:          v.Uin,
			TemplateName: foresttype.MessageTemplateNameAnnouncementNewRelease,
			SourceType:   foresttype.MessageSourceTypeAnnouncement,
			SourceID:     ans.ID,
			MessageParams: map[string]string{
				"tag":             ans.Tag,
				"announcement_id": strconv.Itoa(int(ans.ID)),
			},
		})
	}

	
	go func(c context.Context) {
		_, err = messagecenter.NewMessage().BatchSendMessage(c, reqs)
		if err != nil {
			logs.ErrorContextf(c, "BatchSendMessage failed err: %v", err)
		}
	}(ctx)

	return res, nil
}
