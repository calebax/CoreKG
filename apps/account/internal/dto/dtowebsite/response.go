package dtowebsite

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type UpdateWebsiteInfoResponse struct {
	apiobj.BaseResponse
	Response UpdateWebsiteInfoEmbedResponse
}

type UpdateWebsiteInfoEmbedResponse struct {
}
