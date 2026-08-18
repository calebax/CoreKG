package dtosession

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type MoveSessionRequest struct {
	apiobj.BaseRequest
	Request MoveSessionEmbedRequest
}

type MoveSessionEmbedRequest struct {
	// ID 会话ID
	ID uint `json:"id"`
	// SubjectID 目标projectID若移出项目则为0
	SubjectID uint `json:"subject_id"`
}

func (opt *MoveSessionRequest) Validity(resp *MoveSessionResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_id_empty"
		return
	}
}

type ListFreeSessionsRequest struct {
	apiobj.BaseRequest
	Request ListFreeSessionsEmbedRequest
}
type ListFreeSessionsEmbedRequest struct {
}

func (opt *ListFreeSessionsRequest) Validity(resp *ListFreeSessionsResponse) {
}
