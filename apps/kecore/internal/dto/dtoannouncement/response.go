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
	Data []admin.Announcement `json:"data"`
}
