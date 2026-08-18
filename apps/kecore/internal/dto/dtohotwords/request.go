package dtohotwords

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type GetHotWordsRequest struct {
	apiobj.BaseRequest
	Request GetHotWordsEmbedRequest `json:"request"`
}

type GetHotWordsEmbedRequest struct {
}

func (opt *GetHotWordsRequest) Validity(resp *GetHotWordsResponse) {
}
