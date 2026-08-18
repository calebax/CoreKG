package dtofilesession

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type GetFileSessionRequest struct {
	apiobj.BaseRequest
	Request GetFileSessionEmbedRequest
}

type GetFileSessionEmbedRequest struct {
	FileID uint `json:"file_id"`
}

func (opt *GetFileSessionRequest) Validity(resp *GetFileSessionResponse) {
	if opt.Request.FileID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_file_id_required"
		return
	}
}
