package dtograph

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type CreateForestGraphResponse struct {
	apiobj.BaseResponse
	Response CreateForestGraphEmbedResponse
}

type CreateForestGraphEmbedResponse struct {
	Data *foresttype.ForestGraphInfo `json:"data"`
}
