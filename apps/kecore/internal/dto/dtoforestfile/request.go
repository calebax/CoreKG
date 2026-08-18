package dtoforestfile

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// validateHashFormat 校验哈希格式 "算法协议名:hash值"
func validateHashFormat(hash string) bool {
	if hash == "" {
		return false
	}
	parts := strings.Split(hash, ":")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

// ----------------------------
// 预上传文件请求
// ----------------------------
type PreUploadFileRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint             `json:"forest_id"`
		ParentID uint             `json:"parent_id"`
		Files    []UploadFileItem `json:"files"`
	}
}

type UploadFileItem struct {
	Filename    string               `json:"filename"`
	Hash        string               `json:"hash"`
	Size        int64                `json:"size"`
	SplitConfig *ragtask.SplitConfig `json:"split_config"`
	ContentType string               `json:"content_type"`
}

func (opt *PreUploadFileRequest) Validity(resp *PreUploadFileResponse) {
	// 1️⃣ 检查 forestId
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest" // 请选择知识森林
		return
	}

	// 2️⃣ 检查文件列表
	if len(opt.Request.Files) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_list_empty" // 文件列表为空
		return
	}

	// 3️⃣ 校验每个文件
	for _, f := range opt.Request.Files {
		if f.Filename == "" {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_enter_file_name" // 请输入文件名
			return
		}

		if f.Size == 0 {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_file_size_empty" // 文件大小为空
			return
		}

		if filepath.Ext(f.Filename) == "" {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_file_type_empty" // 文件类型为空
			return
		}

		if !validateHashFormat(f.Hash) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_file_hash_format_invalid" // 文件哈希格式无效
			return
		}

		// SplitConfig 必须存在
		if f.SplitConfig == nil {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_split_config_empty" // 分片配置为空
			return
		}

		cfg := f.SplitConfig

		// 校验分片模式
		if cfg.SplitMode != ragtask.SplitAuto && cfg.SplitMode != ragtask.SplitRule {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_split_mode_invalid" // 分片模式错误
			return
		}

		// 校验分片大小
		if cfg.ChunkSize > 1024 || cfg.ChunkSize < 256 {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_chunk_size_invalid" // 分片大小无效
			return
		}

		// 校验分片标识符
		if len(cfg.SplitMark) == 0 {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_split_mark_empty" // 分片标识符为空
			return
		}

		// 校验分片重叠度
		if cfg.SplitOverlap < 0 || cfg.SplitOverlap > 1 {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_split_overlap_invalid" // 请输入正确的分片重叠度
			return
		}
	}

}

// ----------------------------
// 文件上传回调请求
// ----------------------------
type UploadFileCallBackRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint   `json:"forest_id"`
		UploadID uint   `json:"upload_id"` // 上传ID
		Filename string `json:"filename"`  // 文件名
		Hash     string `json:"hash"`      // 文件hash
		Parts    []struct {
			PartNumber int    `json:"part_number"` // 分片序号
			ETag       string `json:"etag"`        // 分片ETag
		} `json:"parts"` // 分片上传专属
	} `json:"request"`
}

func (opt *UploadFileCallBackRequest) Validity(resp *UploadFileCallBackResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest" // 请选择知识森林
		return
	}
	// 校验 Filename 是否为空
	if opt.Request.Filename == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "filename_empty" // 文件名为空
		return
	}

	// 校验 Hash 是否为空
	if !validateHashFormat(opt.Request.Hash) {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_hash_format_invalid" // 文件哈希格式无效
		return
	}

	// 校验分片信息（可以为空，但如果有元素，ETag 必须非空）
	for i, part := range opt.Request.Parts {
		if part.ETag == "" {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = fmt.Sprintf("empty_etag at index %d", i)
			return
		}
	}

}

// ----------------------------
// 文件上传取消请求
// ----------------------------
type AbortUploadRequest struct {
	apiobj.BaseRequest
	Request struct {
		UploadID uint `json:"upload_id"` // 上传ID
	} `json:"request"`
}

func (opt *AbortUploadRequest) Validity(resp *AbortUploadResponse) {
	if opt.Request.UploadID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "upload_id_empty" // 上传ID为空
		return
	}
}

// ----------------------------
// 文件重新上传请求
// ----------------------------
type RenewUploadUrlRequest struct {
	apiobj.BaseRequest
	Request struct {
		UploadID       uint   `json:"upload_id"`       // 上传ID
		Filename       string `json:"filename"`        // 文件名
		Hash           string `json:"hash"`            // 文件hash
		ExpiredParts   []int  `json:"expired_parts"`   // 需要重新上传的分片序号
		CompletedParts []int  `json:"completed_parts"` // 已完成上传的分片序号
	} `json:"request"`
}

func (opt *RenewUploadUrlRequest) Validity(resp *RenewUploadUrlResponse) {
	if opt.Request.UploadID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "upload_id_empty" // 上传ID为空
		return
	}
	if opt.Request.Hash != "" && !validateHashFormat(opt.Request.Hash) {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_hash_format_invalid" // 文件哈希格式无效
		return
	}
}
