package dtochunk

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type ErrorResponse struct {
	Code    uint32 `json:"code"`    // 错误码，
	Message string `json:"message"` // 错误信息，
}

type GetChunkBySequenceRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileID        uint `json:"file_id"`
		ChunkSequence int  `json:"chunk_sequence"`
	}
}

func (req *GetChunkBySequenceRequest) Validity(resp *GetChunkBySequenceResponse) {
	if req.Request.ChunkSequence == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_invalid_parameter" // 参数错误
		return
	}
	if req.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_invalid_parameter" // 参数错误
		return
	}
}

func (req *GetChunkBySequenceRequest) ValidityDetail(resp *GetChunkDetailResponse) {
	if req.Request.ChunkSequence == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_invalid_parameter" // 参数错误
		return
	}
	if req.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_invalid_parameter" // 参数错误
		return
	}
}
