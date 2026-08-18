package dtostatistics

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type GetAgentQuestionExcelResponse struct {
	apiobj.BaseResponse
	Response GetAgentQuestionExcelEmbedResponse
}

type GetAgentQuestionExcelEmbedResponse struct {
}

type GetAgentQuestionCountResponse struct {
	apiobj.BaseResponse
	Response GetAgentQuestionCountEmbedResponse
}
type GetAgentQuestionCountEmbedResponse struct {
	Count int64 `json:"count"`
}
