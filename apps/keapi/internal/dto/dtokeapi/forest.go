package dtokeapi

import (
	"time"

	kecoreforestmodel "github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// ListForestRequest 知识库列表请求
type ListForestRequest struct {
	apiobj.BaseRequest
	Request apiobj.PageQuery `json:"request"`
}

func (req *ListForestRequest) ValidListForest(resp *apiobj.BaseResponse) bool {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_offset_limit"
		return false
	}
	NormalizePageQuery(&req.Request, "created_at desc")
	return true
}

// BatchGetForestRequest 批量查询知识库请求
type BatchGetForestRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestIDs []uint `json:"forest_ids"`
	} `json:"request"`
}

func (req *BatchGetForestRequest) ValidBatchGetForest(resp *apiobj.BaseResponse) bool {
	if len(req.Request.ForestIDs) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_forest_ids"
		return false
	}
	return true
}

type ForestItem struct {
	ForestID        uint                               `json:"forest_id"`
	Name            string                             `json:"name"`
	KnowledgeStatus foresttype.KnownowForestTaskStatus `json:"knowledge_status"`
	AvatarURL       string                             `json:"avatar_url"`
	Description     string                             `json:"description"`
	FileCount       int64                              `json:"file_count"`
	TotalSize       int64                              `json:"total_size"`
	CreatedAt       time.Time                          `json:"created_at,omitempty"`
}

func NewForestItem(item *kecoreforestmodel.ForestInfoItemstruct) *ForestItem {
	if item == nil {
		return nil
	}
	return &ForestItem{
		ForestID:        item.ID,
		Name:            item.Name,
		KnowledgeStatus: item.KnowledgeStatus,
		AvatarURL:       item.AvatarUrl,
		Description:     item.Description,
		FileCount:       item.FileCount,
		TotalSize:       item.TotalSize,
		CreatedAt:       item.CreatedAt,
	}
}

// ListForestResponse 知识库列表响应
type ListForestResponse struct {
	apiobj.BaseResponse
	Response struct {
		apiobj.QueryResponse
		Data []*ForestItem `json:"data"`
	} `json:"response"`
}

// BatchGetForestResponse 批量查询知识库响应
type BatchGetForestResponse struct {
	apiobj.BaseResponse
	Response struct {
		apiobj.QueryResponse
		Data []*ForestItem `json:"data"`
	} `json:"response"`
}

// CreateForestRequest 创建知识库请求
type CreateForestRequest struct {
	apiobj.BaseRequest
	Request struct {
		Name        string                `json:"name"`
		AvatarURL   string                `json:"avatar_url"`
		Description string                `json:"description"`
		ForestType  foresttype.ForestType `json:"forest_type"`
	} `json:"request"`
}

func (req *CreateForestRequest) ValidCreateForest(resp *apiobj.BaseResponse) bool {
	if req.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_forest_name_empty"
		return false
	}

	if req.Request.ForestType == "" {
		req.Request.ForestType = foresttype.ForestTypeFile
	}
	switch req.Request.ForestType {
	case foresttype.ForestTypeFile, foresttype.ForestTypeData:
		return true
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_forest_type"
		return false
	}
}

// CreateForestResponse 创建知识库响应
type CreateForestResponse struct {
	apiobj.BaseResponse
	Response struct {
		ForestID uint `json:"forest_id"`
	} `json:"response"`
}

// UpdateForestRequest 更新知识库请求
type UpdateForestRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID    uint   `json:"forest_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"request"`
}

func (req *UpdateForestRequest) ValidUpdateForest(resp *apiobj.BaseResponse) bool {
	if req.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_forest_id"
		return false
	}
	if req.Request.Name == "" && req.Request.Description == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_empty_update_fields"
		return false
	}
	return true
}

// DeleteForestRequest 删除知识库请求
type DeleteForestRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint `json:"forest_id"`
	} `json:"request"`
}

func (req *DeleteForestRequest) ValidDeleteForest(resp *apiobj.BaseResponse) bool {
	if req.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_forest_id"
		return false
	}
	return true
}

// ListFileRequest 文档列表请求
type ListFileRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint `json:"forest_id"`
		apiobj.PageQuery
	} `json:"request"`
}

func (req *ListFileRequest) ValidListFile(resp *apiobj.BaseResponse) bool {
	if req.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_forest_id"
		return false
	}
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_offset_limit"
		return false
	}
	NormalizePageQuery(&req.Request.PageQuery, "created_at desc")
	return true
}

// BatchGetFileRequest 批量查询文档请求
type BatchGetFileRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestFileIDs []uint `json:"forest_file_ids"`
	} `json:"request"`
}

func (req *BatchGetFileRequest) ValidBatchGetFile(resp *apiobj.BaseResponse) bool {
	if len(req.Request.ForestFileIDs) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_forest_file_ids"
		return false
	}
	return true
}

type ForestFileItem struct {
	ForestFileID uint      `json:"forest_file_id"`
	ForestID     uint      `json:"forest_id"`
	IsDir        int8      `json:"is_dir"`
	ParentID     uint      `json:"parent_id"`
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	Ext          string    `json:"ext"`
	FileStatus   string    `json:"file_status"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
}

func NewForestFileItem(item *kecoreforestmodel.File) *ForestFileItem {
	if item == nil {
		return nil
	}
	return &ForestFileItem{
		ForestFileID: item.ID,
		ForestID:     item.ForestID,
		IsDir:        int8(item.IsDir),
		ParentID:     item.ParentID,
		Name:         item.Name,
		Size:         item.Size,
		Ext:          item.Ext,
		FileStatus:   item.FileStatus,
		CreatedAt:    item.CreatedAt,
	}
}

// ListFileResponse 文档列表响应
type ListFileResponse struct {
	apiobj.BaseResponse
	Response struct {
		apiobj.QueryResponse
		Data []*ForestFileItem `json:"data"`
	} `json:"response"`
}

// BatchGetFileResponse 批量查询文档响应
type BatchGetFileResponse struct {
	apiobj.BaseResponse
	Response struct {
		apiobj.QueryResponse
		Data []*ForestFileItem `json:"data"`
	} `json:"response"`
}

// UploadFileResponse 上传文档响应
type UploadFileResponse struct {
	apiobj.BaseResponse
	Response struct {
		ForestFileID uint `json:"forest_file_id"`
	} `json:"response"`
}

// DeleteFileRequest 删除文档请求
type DeleteFileRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestFileID uint `json:"forest_file_id"`
	} `json:"request"`
}

func (req *DeleteFileRequest) ValidDeleteFile(resp *apiobj.BaseResponse) bool {
	if req.Request.ForestFileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_forest_file_id"
		return false
	}
	return true
}

// PreviewFileByURLRequest 下载文档请求
type PreviewFileByURLRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestFileID uint `json:"forest_file_id"`
		IsDownload   bool `json:"is_download,omitempty"`
	} `json:"request"`
}

func (req *PreviewFileByURLRequest) ValidPreviewFile(resp *PreviewFileByURLResponse) bool {
	if req.Request.ForestFileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_forest_file_id"
		return false
	}
	return true
}

// PreviewFileByURLResponse 下载文档响应
type PreviewFileByURLResponse struct {
	apiobj.BaseResponse
	Response struct {
		URL string `json:"url"`
	} `json:"response"`
}

// CreateDirRequest 创建文件夹请求
type CreateDirRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint   `json:"forest_id"`
		ParentID uint   `json:"parent_id"`
		Name     string `json:"name"`
	} `json:"request"`
}

func (req *CreateDirRequest) ValidCreateDir(resp *apiobj.BaseResponse) bool {
	if req.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_forest_id"
		return false
	}
	if req.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_dir_name_empty"
		return false
	}
	return true
}

// CreateDirResponse 创建文件夹响应
type CreateDirResponse struct {
	apiobj.BaseResponse
	Response struct {
		ForestFileID uint `json:"forest_file_id"`
	} `json:"response"`
}

// RenamePathRequest 重命名请求
type RenamePathRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestFileID uint   `json:"forest_file_id"`
		Name         string `json:"name"`
	} `json:"request"`
}

func (req *RenamePathRequest) ValidRenamePath(resp *apiobj.BaseResponse) bool {
	if req.Request.ForestFileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_forest_file_id"
		return false
	}
	if req.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_name_empty"
		return false
	}
	return true
}

// DeletePathRequest 删除文件夹请求
type DeletePathRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestFileIDs []uint `json:"forest_file_id"`
	} `json:"request"`
}

func (req *DeletePathRequest) ValidDeletePath(resp *apiobj.BaseResponse) bool {
	if len(req.Request.ForestFileIDs) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_forest_file_ids"
		return false
	}
	return true
}
