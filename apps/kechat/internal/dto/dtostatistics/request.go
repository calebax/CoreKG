package dtostatistics

import (
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type GetAgentQuestionExcelRequest struct {
	apiobj.BaseRequest
	Request GetAgentQuestionExcelEmbedRequest
}

type GetAgentQuestionExcelEmbedRequest struct {
	chatquestion.StatisticsReq
}

func (opt *GetAgentQuestionExcelRequest) Validity(resp *GetAgentQuestionExcelResponse) {
	// if opt.Request.AgentID == 0 {
	// 	resp.Code = errcode.ErrCode_BadRequest
	// 	resp.Message = "agent_id不能为空"
	// }
}

type GetAgentQuestionCountRequest struct {
	apiobj.BaseRequest
	Request GetAgentQuestionCountEmbedRequest
}
type GetAgentQuestionCountEmbedRequest struct {
	chatquestion.StatisticsReq
}

func (opt *GetAgentQuestionCountRequest) Validity(resp *GetAgentQuestionCountResponse) {
	if opt.Request.AgentID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "agent_id不能为空"
	}
}
