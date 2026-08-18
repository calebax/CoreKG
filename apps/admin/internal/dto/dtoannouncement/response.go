package dtoannouncement

import (
	"github.com/insmtx/corekg/apps/admin/models/admin"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type ListAnnouncementResponse struct {
	apiobj.BaseResponse
	Response ListAnnouncementEmbedResponse
}

type ListAnnouncementEmbedResponse struct {
	apiobj.QueryResponse
	Data admin.AnnouncementList `json:"data"`
}

type GetAnnouncementResponse struct {
	apiobj.BaseResponse
	Response GetAnnouncementEmbedResponse
}
type GetAnnouncementEmbedResponse struct {
	admin.Announcement
}

type ModifyAnnouncementResponse struct {
	apiobj.BaseResponse
	Response ModifyAnnouncementEmbedResponse
}
type ModifyAnnouncementEmbedResponse struct {
}

type DeleteAnnouncementResponse struct {
	apiobj.BaseResponse
	Response DeleteAnnouncementEmbedResponse
}
type DeleteAnnouncementEmbedResponse struct {
}

type CreateAnnouncementResponse struct {
	apiobj.BaseResponse
	Response CreateAnnouncementEmbedResponse
}
type CreateAnnouncementEmbedResponse struct {
}
