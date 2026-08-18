package dtoagent

import (
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type GetLatestAgentsResponse struct {
	apiobj.BaseResponse
	Response GetLatestAgentsEmbedResponse
}

type AgentItem struct {
	// 轻应用 ID
	ID uint `json:"id"`
	// 轻应用名称
	Name string `json:"name"`
	// 轻应用名称
	ShowName string `json:"show_name"`
	// 轻应用图标
	AvatarURL string `json:"avatar_url"`
	// 轻应用类型
	AgentType chattype.AgentType `json:"agent_type"`
}

type GetLatestAgentsEmbedResponse struct {
	Data []AgentItem `json:"data"`
	apiobj.QueryResponse
}
