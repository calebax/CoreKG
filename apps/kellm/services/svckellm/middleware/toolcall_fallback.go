package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/insmtx/corekg/apps/kellm/drivers"
	"github.com/insmtx/corekg/apps/kellm/models/functioncall/parser"
	"github.com/insmtx/corekg/apps/kellm/models/functioncall/renderer"
	"github.com/insmtx/corekg/apps/kellm/models/kellmtype"
	"io"
	"net/http"
	"strings"
	"time"
)

func ToolCallFallback() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, chatCtx *drivers.ChatContext) (*drivers.ProxyResult, error) {
			if needsToolCallFallback(chatCtx) {
				upstreamReq, err := prepareToolCallFallbackRequest(chatCtx.OriginalRequest)
				if err != nil {
					return nil, err
				}
				chatCtx.UpstreamRequest = upstreamReq
				chatCtx.NeedToolCallFallback = true
			}

			result, err := next(ctx, chatCtx)
			if err != nil || !chatCtx.NeedToolCallFallback {
				return result, err
			}

			if chatCtx.OriginalRequest != nil && chatCtx.OriginalRequest.Stream {
				return AdaptToolCallFallbackStreamResponse(result)
			}
			return AdaptToolCallFallbackResponse(result)
		}
	}
}

func AdaptToolCallFallbackResponse(result *drivers.ProxyResult) (*drivers.ProxyResult, error) {
	body, err := readProxyResultBody(result)
	if err != nil {
		return nil, err
	}

	adapted := &drivers.ProxyResult{
		StatusCode: result.StatusCode,
		Header:     cloneHeader(result.Header),
		Body:       body,
		Stream:     false,
	}
	if result.StatusCode != http.StatusOK {
		return adapted, nil
	}

	var chatResp kellmtype.ChatResponseBody
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return adapted, nil
	}
	if len(chatResp.Choices) == 0 {
		return adapted, nil
	}
	if chatResp.Object == "" {
		chatResp.Object = "chat.completion"
	}
	if chatResp.ID == "" {
		chatResp.ID = generateChatCompletionID()
	}
	if chatResp.Created == 0 {
		chatResp.Created = time.Now().Unix()
	}

	content := chatResp.Choices[0].Message.Content
	if strings.TrimSpace(content) == "" {
		return adapted, nil
	}

	p := parser.NewParser("json")
	toolCalls, remainingText, parseErr := p.Parse(content)
	if parseErr != nil || len(toolCalls) == 0 {
		return adapted, nil
	}

	chatResp.Choices[0].Message.ToolCalls = convertToOpenAIToolCalls(toolCalls)
	chatResp.Choices[0].Message.Content = remainingText
	chatResp.Choices[0].FinishReason = "tool_calls"

	adaptedBody, err := json.Marshal(chatResp)
	if err != nil {
		return nil, err
	}
	adapted.Body = adaptedBody
	adapted.Header.Del("Content-Length")
	adapted.Header.Set("Content-Type", "application/json")
	return adapted, nil
}

func AdaptToolCallFallbackStreamResponse(result *drivers.ProxyResult) (*drivers.ProxyResult, error) {
	body, err := readProxyResultBody(result)
	if err != nil {
		return nil, err
	}

	adapted := &drivers.ProxyResult{
		StatusCode: result.StatusCode,
		Header:     cloneHeader(result.Header),
		Stream:     true,
	}
	if result.StatusCode != http.StatusOK {
		adapted.Body = body
		return adapted, nil
	}

	streamState, err := parseStreamContent(body)
	if err != nil {
		adapted.Body = body
		return adapted, nil
	}
	p := parser.NewParser("json")
	toolCalls, remainingText, parseErr := p.Parse(streamState.content.String())
	if parseErr != nil || len(toolCalls) == 0 {
		adapted.BodyReader = io.NopCloser(bytes.NewReader(body))
		return adapted, nil
	}

	chunk := kellmtype.ChatStreamResponseBody{
		ID:                defaultString(streamState.id, generateChatCompletionID()),
		Created:           defaultInt64(streamState.created, time.Now().Unix()),
		Model:             streamState.model,
		SystemFingerprint: streamState.systemFingerprint,
		Object:            defaultString(streamState.object, "chat.completion.chunk"),
		Usage:             streamState.usage,
		Choices: []kellmtype.ChoiceStream{{
			Index:        0,
			FinishReason: "tool_calls",
			Delta: kellmtype.Delta{
				Content:   remainingText,
				ToolCalls: convertToOpenAIToolCalls(toolCalls),
			},
		}},
	}

	chunkData, err := json.Marshal(chunk)
	if err != nil {
		return nil, err
	}
	adapted.Body = []byte("data: " + string(chunkData) + "\n\n" + "data: [DONE]\n\n")
	adapted.Header.Del("Content-Length")
	adapted.Header.Set("Content-Type", "text/event-stream")
	adapted.Header.Set("Cache-Control", "no-cache")
	adapted.Header.Set("Connection", "keep-alive")
	return adapted, nil
}

func needsToolCallFallback(chatCtx *drivers.ChatContext) bool {
	if chatCtx == nil || chatCtx.OriginalRequest == nil || chatCtx.ModelConfig == nil {
		return false
	}
	if len(chatCtx.OriginalRequest.Tools) == 0 {
		return false
	}
	return !chatCtx.ModelConfig.Capabilities.ToolCall
}

func prepareToolCallFallbackRequest(req *kellmtype.ChatRequestBody) (*kellmtype.ChatRequestBody, error) {
	r := renderer.NewRenderer()
	prompt, err := r.Render(req.Messages, req.Tools)
	if err != nil {
		return nil, fmt.Errorf("render prompt failed: %w", err)
	}

	modifiedReq := *req
	modifiedReq.Tools = nil
	modifiedReq.ToolChoice = nil
	modifiedReq.Messages = []kellmtype.Message{{
		Role: "user",
		Content: kellmtype.MessageContent{
			Text: prompt,
		},
	}}
	return &modifiedReq, nil
}

type streamState struct {
	id                string
	created           int64
	model             string
	systemFingerprint string
	object            string
	usage             *kellmtype.Usage
	content           strings.Builder
}

func parseStreamContent(body []byte) (*streamState, error) {
	state := &streamState{}
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		var data string
		switch {
		case strings.HasPrefix(line, "data: "):
			data = strings.TrimPrefix(line, "data: ")
		case strings.HasPrefix(line, "{"):
			data = line
		default:
			continue
		}
		if data == "[DONE]" || data == "" {
			continue
		}
		var chunk kellmtype.ChatStreamResponseBody
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if chunk.ID != "" {
			state.id = chunk.ID
		}
		if chunk.Created > 0 {
			state.created = chunk.Created
		}
		if chunk.Model != "" {
			state.model = chunk.Model
		}
		if chunk.SystemFingerprint != "" {
			state.systemFingerprint = chunk.SystemFingerprint
		}
		if chunk.Object != "" {
			state.object = chunk.Object
		}
		if chunk.Usage != nil {
			state.usage = chunk.Usage
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				state.content.WriteString(choice.Delta.Content)
			}
		}
	}
	return state, nil
}

func convertToOpenAIToolCalls(calls []parser.ToolCall) []kellmtype.ToolCall {
	result := make([]kellmtype.ToolCall, len(calls))
	for i, call := range calls {
		argsJSON, _ := json.Marshal(call.Arguments)
		result[i] = kellmtype.ToolCall{
			Index: i,
			ID:    generateToolCallID(),
			Type:  "function",
			Function: kellmtype.ToolCallFunction{
				Name:      call.Name,
				Arguments: string(argsJSON),
			},
		}
	}
	return result
}

func cloneHeader(header http.Header) http.Header {
	dst := make(http.Header, len(header))
	for key, values := range header {
		copied := make([]string, len(values))
		copy(copied, values)
		dst[key] = copied
	}
	return dst
}
