package dtochat

import "github.com/ygpkg/yg-go/apis/apiobj"

type ChatQuestionStreamResponse struct {
	apiobj.BaseResponse
	Response ChatQuestionStreamEmbedResponse
}

type ChatQuestionStreamEmbedResponse struct {
}

type ExpansionQuestionResponse struct {
	apiobj.BaseResponse
	Response ExpansionQuestionEmbedResponse `json:"response"`
}
type ExpansionQuestionEmbedResponse struct {
	ExpandedQuestion string `json:"expanded_question"`
}
