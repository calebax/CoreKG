package chunkctl

import (
	"github.com/insmtx/corekg/apps/kesearch/models/chunk"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type ListFileChunkRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint `json:"forest_id"`
		FileID   uint `json:"file_id"`
	}
}

func (req *ListFileChunkRequest) Validity(resp *ListFileChunkResponse) {
	if req.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_select_forest" // 请选择知识森林
		return
	}
	if req.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_select_file" // 请选择文件
		return
	}
}

type ListFileChunkResponse struct {
	apiobj.BaseResponse
	Response struct {
		Chunks []*chunk.Chunk `json:"chunks"`
	}
}

type GetChunkByIDRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileID  uint   `json:"file_id"`
		ChunkID string `json:"chunk_id"`
	}
}

func (req *GetChunkByIDRequest) Validity(resp *GetChunkByIDResponse) {
	if req.Request.ChunkID == "" {
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

type GetChunkByIDResponse struct {
	apiobj.BaseResponse
	Response struct {
		Chunk *chunk.Chunk `json:"chunk"`
	}
}

type UpdateChunkRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileID      uint   `json:"file_id"`
		ChunkID     string `json:"chunk_id"`
		Description string `json:"description"`
		Table       string `json:"table"`
	}
}

func (req *UpdateChunkRequest) Validity(resp *UpdateChunkResponse) {
	if req.Request.ChunkID == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_select_chunk" // 请选择chunk
		return
	}
	if req.Request.Description == "" && req.Request.Table == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_enter_description" // 请输入描述
		return
	}
	if req.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_invalid_parameter" // 参数错误
		return
	}
}

type UpdateChunkResponse struct {
	apiobj.BaseResponse
	Response struct {
		Chunk *chunk.Chunk `json:"chunk"`
	}
}

type DeleteChunkRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileID  uint   `json:"file_id"`
		ChunkID string `json:"chunk_id"`
	}
}

func (req *DeleteChunkRequest) Validity(resp *DeleteChunkResponse) {
	if req.Request.ChunkID == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_select_chunk" // 请选择chunk
		return
	}
	if req.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_invalid_parameter" // 参数错误
		return
	}
}

type DeleteChunkResponse struct {
	apiobj.BaseResponse
	Response struct {
	}
}

type DisableFileChunkRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileID    uint `json:"file_id"`
		IsDisable bool `json:"is_disable"`
	}
}

func (req *DisableFileChunkRequest) Validity(resp *DisableFileChunkResponse) {
	if req.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_select_file" // 请选择文件
		return
	}
}

type DisableFileChunkResponse struct {
	apiobj.BaseResponse
	Response struct {
	}
}

type MigrateRequest struct {
	apiobj.BaseRequest
	Request struct {
		Key string `json:"key"`
	}
}

// Validity 验证有效性
func (req *MigrateRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Key != "ajhsdkajdssajkzb" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_invalid_key" // key错误
		return
	}
}

// CreateAgentResponse 创建指令型机器人命令响应
type MigrateResponse struct {
	apiobj.BaseResponse
	Response struct{}
}
