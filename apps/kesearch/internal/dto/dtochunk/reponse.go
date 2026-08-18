package dtochunk

import (
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kesearch/models/chunk"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type GetChunkBySequenceResponse struct {
	apiobj.BaseResponse
	Response struct {
		Chunk *chunk.Chunk `json:"chunk"`
	}
}

type GetChunkDetailResponse struct {
	apiobj.BaseResponse
	Response struct {
		Detail *chattype.QueryReference `json:"detail"`
	}
}
