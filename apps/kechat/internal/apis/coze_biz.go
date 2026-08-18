package apis

import (
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type CreateCozePluginRequest struct {
	apiobj.BaseRequest
	Request struct {
		AgentID uint `json:"agent_id"`
	}
}

type DeleteCozeMappingByCozeIDRequest struct {
	apiobj.BaseRequest
	Request struct {
		CozeID   string            `json:"coze_id"`
		ChatType chattype.ChatType `json:"chat_type"`
	}
}

type GetAgentMappingRequest struct {
	apiobj.BaseRequest
	Request struct {
		WorkflowID string            `json:"workflow_id"`
		CozeType   chattype.ChatType `json:"coze_type"`
	} `json:"request"`
}

type GetAgentMappingResponse struct {
	apiobj.BaseResponse
	Response struct {
		AgentID uint `json:"agent_id"`
	}
}
