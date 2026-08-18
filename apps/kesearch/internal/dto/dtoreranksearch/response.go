package dtoreranksearch

import (
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type RerankSearchChunkResponse struct {
	apiobj.BaseResponse
	Response RerankSearchChunkEmbedResponse
}

type RerankSearchChunkEmbedResponse struct {
	SearchResult chattype.QueryReferenceList `json:"search_result"`
}
