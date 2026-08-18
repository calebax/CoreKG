package dtokeapi

import (
	"strings"

	"github.com/insmtx/corekg/apps/kellm/models/kellmtype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// ChatCompletionsRequest 对话请求
type ChatCompletionsRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestFileIDs   []uint                    `json:"forest_file_id"`
		SessionID       uint                      `json:"session_id,omitempty"`
		Messages        []kellmtype.Message       `json:"messages"`
		Stream          bool                      `json:"stream"`
		Temperature     *float32                  `json:"temperature,omitempty"`
		TopP            *float32                  `json:"top_p,omitempty"`
		PresencePenalty *float32                  `json:"presence_penalty,omitempty"`
		ResponseFormat  *kellmtype.ResponseFormat `json:"response_format,omitempty"`
		ExtraBody       *ChatCompletionsExtraBody `json:"extra_body,omitempty"`
	} `json:"request"`
}

type ChatCompletionsExtraBody struct {
	EnableReference *bool `json:"enable_reference,omitempty"`
}

func (req *ChatCompletionsRequest) ValidChatCompletions(resp *apiobj.BaseResponse) bool {
	if len(req.Request.Messages) == 0 {
		if req.Request.SessionID > 0 {
			return true
		}
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_empty_messages"
		return false
	}
	hasContent := false
	for _, msg := range req.Request.Messages {
		if msg.Role != "user" {
			continue
		}
		if strings.TrimSpace(msg.Content.Text) != "" || len(msg.Content.Items) > 0 {
			hasContent = true
			break
		}
	}
	if !hasContent {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_empty_messages"
		return false
	}
	return true
}

type OpenAIChatMessage struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type OpenAIChatChoice struct {
	Index        int               `json:"index"`
	Message      OpenAIChatMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

type OpenAIChatCompletion struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []OpenAIChatChoice `json:"choices"`
	Usage   kellmtype.Usage    `json:"usage"`
}

type OpenAIChatDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type OpenAIChatChunkChoice struct {
	Index        int             `json:"index"`
	Delta        OpenAIChatDelta `json:"delta"`
	FinishReason *string         `json:"finish_reason"`
}

type OpenAIChatChunk struct {
	ID      string                  `json:"id"`
	Object  string                  `json:"object"`
	Created int64                   `json:"created"`
	Model   string                  `json:"model"`
	Choices []OpenAIChatChunkChoice `json:"choices"`
}
