package dtosession

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type MoveSessionResponse struct {
	apiobj.BaseResponse
	Response MoveSessionEmbedResponse
}

type MoveSessionEmbedResponse struct {
}

type ListFreeSessionsResponse struct {
	apiobj.BaseResponse
	Response ListFreeSessionsEmbedResponse
}
type ListFreeSessionsEmbedResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
