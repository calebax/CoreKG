package dtohotwords

import (
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type GetHotWordsResponse struct {
	apiobj.BaseResponse
	Response GetHotWordsEmbedResponse `json:"response"`
}

type GetHotWordsEmbedResponse struct {
	Words types.StringArray `json:"words"`
}
