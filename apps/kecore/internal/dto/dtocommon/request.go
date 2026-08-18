package dtocommon

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type GetCommonInfoRequest struct {
	apiobj.BaseRequest
	Request GetCommonInfoEmbedRequest
}

type GetCommonInfoEmbedRequest struct {
}

func (opt *GetCommonInfoRequest) Validity(resp *GetCommonInfoResponse) {
}
