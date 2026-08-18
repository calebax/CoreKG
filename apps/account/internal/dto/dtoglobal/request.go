package dtoglobal

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type GetGlobalInfoRequest struct {
	apiobj.BaseRequest
	Request GetGlobalInfoEmbedRequest
}

type GetGlobalInfoEmbedRequest struct {
}

func (opt *GetGlobalInfoRequest) Validity(resp *GetGlobalInfoResponse) {
}
