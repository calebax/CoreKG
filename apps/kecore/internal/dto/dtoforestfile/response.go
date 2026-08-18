package dtoforestfile

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type ErrorResponse struct {
	Code    uint32 `json:"code"`    // 错误码，
	Message string `json:"message"` // 错误信息，
}

// ----------------------------
// 预上传文件响应
// ----------------------------
type PreUploadFileResponse struct {
	apiobj.BaseResponse
	Response struct {
		Files []*PreUploadFileResponseItem `json:"files"`
	}
}

// 单个文件上传信息
type PreUploadFileResponseItem struct {
	UploadID   uint                      `json:"upload_id"`
	Hash       string                    `json:"hash"`
	Exists     bool                      `json:"exists"`
	Multipart  *PreUploadMultipartConfig `json:"multipart"`
	UploadURLs map[int]string            `json:"upload_urls"`
}

// 分片配置
type PreUploadMultipartConfig struct {
	Enabled   bool  `json:"enabled"`
	ChunkSize int64 `json:"chunk_size"`
	PartCount int   `json:"part_count"`
}

// ----------------------------
// 文件上传回调响应
// ----------------------------
type UploadFileCallBackResponse struct {
	apiobj.BaseResponse
	Response struct {
		FileID *uint `json:"file_id"`
	}
}

// ----------------------------
// 文件上传取消响应
// ----------------------------
type AbortUploadResponse struct {
	apiobj.BaseResponse
}

// ----------------------------
// 文件重新上传响应
// ----------------------------
type RenewUploadUrlResponse struct {
	apiobj.BaseResponse
	Response struct {
		// key: 分片号，value: 新的上传 URL
		RenewedUrls *map[int]string `json:"renewed_urls"`
	}
}
