package dtoannouncement

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type ListAnnouncementRequest struct {
	apiobj.BaseRequest
	Request ListAnnouncementEmbedRequest
}

type ListAnnouncementEmbedRequest struct {
	apiobj.PageQuery
}

func (opt *ListAnnouncementRequest) Validity(_ *ListAnnouncementResponse) {
}
