package chunkctl

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kesearch/models/chunk"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// ListFileChunk 获取文件chunk
// @Tags chunk 管理
// @Summary 获取文件chunk
// @Description 获取文件chunk
// @Router /kesearch.ListFileChunk [post]
// @Param user body ListFileChunkRequest true "入参"
// @Success 200 {object} ListFileChunkResponse "返回值"
func ListFileChunk(ctx *gin.Context, req *ListFileChunkRequest, resp *ListFileChunkResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	_, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_file_failed" // 获取文件失败
		return
	}
	chunks, err := chunk.ListChunksByFileID(ctx, req.Request.FileID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_chunks_failed" // 获取文件分片失败
		return
	}
	resp.Response.Chunks = chunks
}

// GetChunkByID 根据id获取分片
// @Tags chunk 管理
// @Summary 根据id获取分片
// @Description 根据id获取分片
// @Router /kesearch.GetChunkByID [post]
// @Param user body GetChunkByIDRequest true "入参"
// @Success 200 {object} GetChunkByIDResponse "返回值"
func GetChunkByID(ctx *gin.Context, req *GetChunkByIDRequest, resp *GetChunkByIDResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	chunk, err := chunk.GetChunkByID(ctx, req.Request.ChunkID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_chunk_failed" // 获取分片失败
		return
	}
	resp.Response.Chunk = chunk
}

// UpdateChunk 修改chunk
// @Tags chunk 管理
// @Summary 修改chunk
// @Description 修改chunk
// @Router /kesearch.UpdateChunk [post]
// @Param user body UpdateChunkRequest true "入参"
// @Success 200 {object} UpdateChunkResponse "返回值"
func UpdateChunk(ctx *gin.Context, req *UpdateChunkRequest, resp *UpdateChunkResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	chunkInfo, err := chunk.GetChunkByID(ctx, req.Request.ChunkID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_chunk_failed" // 获取分片失败
		return
	}
	chunkInfo.Source.Description = req.Request.Description
	chunkInfo.Source.Table = req.Request.Table
	chunkInfo.Source.DescriptionHash = chunk.GetSHA256Hash(req.Request.Description)
	eb, err := essearch.GetEmbedding(req.Request.Description)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_embedding_failed" // 获取分片失败
		return
	}
	chunkInfo.Source.Embedding = eb
	err = chunk.UpdateChunk(ctx, chunkInfo)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_update_chunk_failed" // 更新分片失败
		return
	}
	resp.Response.Chunk = chunkInfo
}

// DeleteChunk 删除chunk
// @Tags chunk 管理
// @Summary 删除chunk
// @Description 删除chunk
// @Router /kesearch.DeleteChunk [post]
// @Param user body DeleteChunkRequest true "入参"
// @Success 200 {object} DeleteChunkResponse "返回值"
func DeleteChunk(ctx *gin.Context, req *DeleteChunkRequest, resp *DeleteChunkResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	chunkInfo, err := chunk.GetChunkByID(ctx, req.Request.ChunkID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_chunk_failed" // 获取分片失败
		return
	}
	err = chunk.MinusSequence(ctx, chunkInfo.Source.FileID, chunkInfo.Source.Sequence)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_update_chunk_failed" // 更新分片失败
		return
	}
	err = chunk.DeleteChunk(ctx, chunkInfo.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_delete_chunk_failed" // 删除分片失败
		return
	}
}

// DisableFileChunk 设置文件chunk状态
// @Tags chunk 管理
// @Summary 设置文件chunk状态
// @Description 设置文件chunk状态
// @Router /kesearch.DisableFileChunk [post]
// @Param user body DisableFileChunkRequest true "入参"
// @Success 200 {object} DisableFileChunkResponse "返回值"
func DisableFileChunk(ctx *gin.Context, req *DisableFileChunkRequest, resp *DisableFileChunkResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	err := chunk.DisableFileChunk(ctx, req.Request.FileID, req.Request.IsDisable)
	if err != nil {
		logs.ErrorContextf(ctx, "DisableFileChunk: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_disable_chunk_failed" // 设置文件chunk状态失败
		return
	}
}

// MigrateChunkFileName 迁移chunk文件名数据
// @Tags chunk 管理
// @Summary 迁移chunk文件名数据
// @Description 迁移chunk文件名数据
// @Router /kesearch.MigrateChunkFileName [post]
// @Param user body MigrateRequest true "入参"
// @Success 200 {object} MigrateResponse "返回值"
func MigrateChunkFileName(ctx *gin.Context, req *MigrateRequest, resp *MigrateResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != errcode.CodeOK {
		return
	}
	var files []*foresttype.KnownowForestFile
	err := dbutil.Knownow().Model(&foresttype.KnownowForestFile{}).Find(&files).Error
	if err != nil {
		logs.ErrorContextf(ctx, "MigrateChunkFileName: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_files_failed" // 获取文件失败
		return
	}

	for _, file := range files {
		logs.InfoContextf(ctx, "MigrateChunkFileName: %+v", file)
		err := chunk.UpdateChunkFileName(ctx, file.ID, file.Name)
		if err != nil {
			logs.ErrorContextf(ctx, "MigrateChunkFileName: %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kesearch_update_chunk_failed" // 修改chunk失败
			return
		}
	}
}
