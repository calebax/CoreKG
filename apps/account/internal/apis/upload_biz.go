package apis

import "github.com/ygpkg/yg-go/apis/apiobj"

type UploadImageResponse struct {
	apiobj.BaseResponse

	Response FileInfo
}

// FileInfo file info
type FileInfo struct {
	FileID    uint `json:"file_id,omitempty"`
	Uin       uint `json:"uin,omitempty"`
	CompanyID uint `json:"company_id,omitempty"`
	// URL 文件访问地址
	URL string `json:"url,omitempty"`
	// Filename 原始文件名
	Filename string `json:"filename,omitempty"`
}
