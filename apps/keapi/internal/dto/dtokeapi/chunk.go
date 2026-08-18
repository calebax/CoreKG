package dtokeapi

import (
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// GetFileChunksBySequencesRequest 按知识库文件 ID 和 chunk 序号列表查询 chunk 信息。
type GetFileChunksBySequencesRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestFileID   uint  `json:"forest_file_id"`
		ChunkSequences []int `json:"chunk_sequences"`
		ChunkSequence  []int `json:"chunk_sequence,omitempty"`
	} `json:"request"`
}

func (req *GetFileChunksBySequencesRequest) EffectiveChunkSequences() []int {
	if len(req.Request.ChunkSequences) > 0 {
		return req.Request.ChunkSequences
	}
	return req.Request.ChunkSequence
}

func (req *GetFileChunksBySequencesRequest) ValidGetFileChunksBySequences(resp *apiobj.BaseResponse) bool {
	if req.Request.ForestFileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_forest_file_id"
		return false
	}
	sequences := req.EffectiveChunkSequences()
	if len(sequences) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_chunk_sequences"
		return false
	}
	for _, sequence := range sequences {
		if sequence <= 0 {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "keapi_invalid_chunk_sequences"
			return false
		}
	}
	return true
}

// GetFileChunksBySequencesResponse 按知识库文件 ID 和 chunk 序号列表查询 chunk 信息响应。
type GetFileChunksBySequencesResponse struct {
	apiobj.BaseResponse
	Response *chattype.QueryReference `json:"response"`
}
