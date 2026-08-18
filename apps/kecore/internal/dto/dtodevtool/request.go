package dtodevtool

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type RewriteMarkdownURLRequest struct {
	apiobj.BaseRequest
	Request RewriteMarkdownURLEmbedRequest
}

type RewriteMarkdownURLEmbedRequest struct {
	Path      string `json:"path"`       // S3 路径前缀，例如 "algo-lke/434/41"，为空时使用默认路径
	OldPrefix string `json:"old_prefix"` // 需要替换的 URL 前缀，例如 "http://ip:port"
	NewPrefix string `json:"new_prefix"` // 新的 URL 前缀，例如 "https://new-domain.com"
	BizBucket string `json:"biz_bucket"` // 业务 bucket 名称，用于构建动态路径
	Service   string `json:"service"`    // 服务名称，用于构建动态路径
}

func (opt *RewriteMarkdownURLRequest) Validity(resp *RewriteMarkdownURLResponse) {
	if opt.Request.OldPrefix == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_require_old_prefix"
		return
	}
	if opt.Request.NewPrefix == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_require_new_prefix"
		return
	}
	if opt.Request.Path == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_require_storage_path"
		return
	}
	if opt.Request.OldPrefix == opt.Request.NewPrefix {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_old_new_prefix_same"
		return
	}
	if opt.Request.BizBucket == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_require_biz_bucket"
		return
	}
	if opt.Request.Service == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_require_service"
		return
	}

}

type StatAlgoMarkdownRequest struct {
	apiobj.BaseRequest
	Request StatAlgoMarkdownEmbedRequest
}
type StatAlgoMarkdownEmbedRequest struct {
	Path string `json:"path"` // minio 中的路径前缀，例如 "algo-lke/434/41"
}

func (opt *StatAlgoMarkdownRequest) Validity(resp *StatAlgoMarkdownResponse) {
	if opt.Request.Path == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_require_storage_path"
	}
}
