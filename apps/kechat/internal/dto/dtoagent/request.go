package dtoagent

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type GetLatestAgentsRequest struct {
	apiobj.BaseRequest
	Request apiobj.PageQuery
}

func (opt *GetLatestAgentsRequest) Validity(resp *GetLatestAgentsResponse) {
}
