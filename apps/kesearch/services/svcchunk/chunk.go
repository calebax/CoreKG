package svcchunk

import (
	"context"

	"github.com/insmtx/corekg/apps/kesearch/internal/dto/dtochunk"
	"github.com/insmtx/corekg/apps/kesearch/models/chunk"
	"github.com/ygpkg/yg-go/apis/errcode"
)

func GetChunkBySequence(ctx context.Context, req *dtochunk.GetChunkBySequenceRequest) (chunkInfo *chunk.Chunk, errInfo *dtochunk.ErrorResponse) {
	chunkInfo, err := chunk.GetChunkBySequence(ctx, req.Request.FileID, req.Request.ChunkSequence)
	if err != nil {
		return nil, &dtochunk.ErrorResponse{
			// 获取分片失败
			Code: errcode.ErrCode_InternalError, Message: "kesearch_get_chunk_failed",
		}
	}
	return chunkInfo, nil
}
