package dtoannouncement

import (
	"github.com/insmtx/corekg/apps/admin/models/admin"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type ListAnnouncementRequest struct {
	apiobj.BaseRequest
	Request ListAnnouncementEmbedRequest
}

type ListAnnouncementEmbedRequest struct {
	apiobj.PageQuery
}

func (opt *ListAnnouncementRequest) Validity(resp *ListAnnouncementResponse) {
}

type GetAnnouncementRequest struct {
	apiobj.BaseRequest
	Request GetAnnouncementEmbedRequest
}
type GetAnnouncementEmbedRequest struct {
	//公告id
	ID uint `json:"id"`
}

func (opt *GetAnnouncementRequest) Validity(resp *GetAnnouncementResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_id_empty"
		return
	}
}

type ModifyAnnouncementRequest struct {
	apiobj.BaseRequest
	Request ModifyAnnouncementEmbedRequest
}
type ModifyAnnouncementEmbedRequest struct {
	admin.Announcement
}

func (opt *ModifyAnnouncementRequest) Validity(resp *ModifyAnnouncementResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_id_empty"
		return
	}
	if len(opt.Request.Tag) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "admin_tag_empty"
		return
	}
}

type DeleteAnnouncementRequest struct {
	apiobj.BaseRequest
	Request DeleteAnnouncementEmbedRequest
}
type DeleteAnnouncementEmbedRequest struct {
	ID uint `json:"id"`
}

func (opt *DeleteAnnouncementRequest) Validity(resp *DeleteAnnouncementResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_id_empty"
		return
	}
}

type CreateAnnouncementRequest struct {
	apiobj.BaseRequest
	Request CreateAnnouncementEmbedRequest
}
type CreateAnnouncementEmbedRequest struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

func (opt *CreateAnnouncementRequest) Validity(resp *CreateAnnouncementResponse) {
	if len(opt.Request.Tag) <= 0 {
		if len(opt.Request.Tag) == 0 {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "admin_tag_empty"
			return
		}
	}
}
