package dtofilesession

import (
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type GetFileSessionResponse struct {
	apiobj.BaseResponse
	Response GetFileSessionEmbedResponse
}

type GetFileSessionEmbedResponse struct {
	Session *chattype.ChatSession `json:"Session"`
}
