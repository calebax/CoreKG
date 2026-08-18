package models

import (
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
)

type AgentRequest struct {
	SessionID  uint             `json:"sessionID"`
	RequestID  string           `json:"requestID"`
	Query      string           `json:"query"`
	InputFiles []File           `json:"inputFiles"`
	IsStream   bool             `json:"isStream"`
	Messages   []schema.Message `json:"messages"`
	Options    *AgentOptions    `json:"options,omitempty"`
}

type AgentOptions struct {
	// 是否开启联网搜索
	EnableWebSearch bool `json:"enable_web_search,omitempty"`
}

func (a *AgentRequest) GetInitialMessage() string {
	var historyBuilder strings.Builder
	for _, msg := range a.Messages {
		fmt.Fprintf(&historyBuilder, "role:%s content:%s\n", msg.Role, msg.Content)
	}
	return historyBuilder.String()
}
