package chatagent

import (
	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type AgentWithVersion struct {
	chattype.ChatAgent
	Description     string                    `json:"description"`
	GreetingMessage string                    `json:"greeting_message"`
	Params          chattype.ParamsList       `json:"params"`
	ChatModelIDs    types.UintArray           `json:"chat_model_ids"`
	IsCollected     bool                      `json:"is_collected"`
	IsAdmin         bool                      `json:"is_admin"`
	PromptTemplate  string                    `json:"prompt_template"`
	Temperature     float32                   `json:"temperature"`
	ForestOption    chattype.ForestChatOption `json:"forest_option"`
	IsSynced        bool                      `json:"is_synced"`
}

// QueryChatAgentListResponse 查询机器人列表响应
type QueryChatAgentListResponse struct {
	apiobj.QueryResponse
	Data []*AgentWithVersion
}

type UpdateChatAgentItem struct {
	ID          uint   `json:"id"`
	AvatarURL   string `json:"avatar_url"`
	ShowName    string `json:"show_name"`
	Description string `json:"description"`
}

type AgentApp struct {
	AvatarURL   string             `json:"avatar_url"`
	ShowName    string             `json:"show_name"`
	Description string             `json:"description"`
	AgentType   chattype.AgentType `json:"agent_type"`
	WorkflowID  string             `json:"workflow_id"`
	SpaceID     string             `json:"space_id"`
}

// CreatePromptAgentItem 创建prompt机器人请求
type CreateAgentItem struct {
	chattype.ChatAgent
	// prompt模板
	PromptTemplate string `json:"prompt_template"`
	// 描述
	Description string `json:"description"`
	// 通用大模型
	ChatModelIDs types.UintArray `json:"chat_model_ids"`
	// 自定义温度
	Temperature float32 `json:"temperature"`
	// 公开范围
	PublicScope chattype.PublicScope `json:"public_scope"`
	// 标签
	Tag types.StringArray `json:"tag"`
	// prompt参数解释
	Params chattype.ParamsList `json:"params"`
	// GreetingMessage 问候语
	GreetingMessage string `json:"greeting_message"`
	// 知识库id
	ForestIDs []uint `json:"doc_forest_ids"`
}

type UpdateAgentItem struct {
	ID              uint                 `json:"id"`
	AvatarURL       string               `json:"avatar_url"`
	ShowName        string               `json:"show_name"`
	PublicScope     chattype.PublicScope `json:"public_scope"`
	Description     string               `json:"description"`
	ChatModelIDs    types.UintArray      `json:"chat_model_ids"`
	Temperature     float32              `json:"temperature"`
	IsEnableAIGC    types.Bool           `json:"is_enable_aigc"`
	PromptTemplate  string               `json:"prompt_template"`
	Tag             types.StringArray    `json:"tag"`
	Params          chattype.ParamsList  `json:"params"`
	ManagerIDs      types.UintArray      `json:"manager_ids"`
	ScopeIDs        types.UintArray      `json:"scope_ids"`
	GreetingMessage string               `json:"greeting_message"`
	AgentType       chattype.AgentType   `json:"agent_type"`
	// 知识库id
	ForestIDs []uint `json:"doc_forest_ids"`
	CompanyID uint   `json:"-"`
}

// PromptAgentItem 指令型机器人信息
type AgentItemInfo struct {
	Description     string                      `json:"description"`
	ChatModels      []chatmodel.LLmModel        `json:"chat_models"`
	Temperature     float32                     `json:"temperature"`
	AgentType       chattype.AgentType          `json:"agent_type"`
	PromptTemplate  string                      `json:"prompt_template"`
	Params          chattype.ParamsList         `json:"params"`
	GreetingMessage string                      `json:"greeting_message"`
	Forests         []*foresttype.KnownowForest `json:"forests"`
}

type AgentWithPerm struct {
	chattype.ChatAgent
	ManagerIDs types.UintArray `json:"manager_ids"`
	ScopeIDs   types.UintArray `json:"scope_ids"`
}

type ForestPromptConfig struct {
	PromptTemplate string `yaml:"prompt_template"`
}
