package core

import (
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	agentservice "github.com/insmtx/corekg/pkgs/einotools/service"
)

type ChatResult struct {
	Answer      string                  `json:"answer,omitempty"`
	Reasoning   string                  `json:"reasoning,omitempty"`
	Usage       TokenUsage              `json:"usage,omitempty"`
	Performance Performance             `json:"performance,omitempty"`
	Status      chattype.QuestionStatus `json:"status,omitempty"`
	Meta        ChatResultMeta          `json:"meta,omitempty"`
}

type ChatResultMetaKey string

const (
	MetaKeyAgentService    ChatResultMetaKey = "agentService"
	MetaKeyQueryReferences ChatResultMetaKey = "queryReferences"
	MetaKeyChatReferences  ChatResultMetaKey = "chatReferences"
)

type ChatResultMeta map[ChatResultMetaKey]interface{}

func (m ChatResultMeta) AgentService() *agentservice.ReactAgentService {
	if m == nil {
		return nil
	}
	if agentSvc, ok := m[MetaKeyAgentService]; ok {
		if svc, ok := agentSvc.(*agentservice.ReactAgentService); ok {
			return svc
		}
	}
	return nil
}

func (m ChatResultMeta) QueryReferences() *chattype.QueryReferenceList {
	if m == nil {
		return nil
	}
	if refs, ok := m[MetaKeyQueryReferences]; ok {
		switch v := refs.(type) {
		case *chattype.QueryReferenceList:
			return v
		case chattype.QueryReferenceList:
			return &v
		}
	}
	return nil
}

func (m ChatResultMeta) ChatReferences() *chattype.ChatReferenceList {
	if m == nil {
		return nil
	}
	if refs, ok := m[MetaKeyChatReferences]; ok {
		switch v := refs.(type) {
		case *chattype.ChatReferenceList:
			return v
		case chattype.ChatReferenceList:
			return &v
		}
	}
	return nil
}

// Token用量统计
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	CacheHitTokens   int `json:"cache_hit_tokens,omitempty"`
	CacheMissTokens  int `json:"cache_miss_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

// 性能统计
type Performance struct {
	ReasoningSeconds int `json:"reasoning_seconds,omitempty"`
	CostSeconds      int `json:"cost_seconds,omitempty"`
}
