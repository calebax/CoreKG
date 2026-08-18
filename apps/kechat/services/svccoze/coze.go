package svccoze

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/coze"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

const cozeEventConversationDelta = "conversation.message.delta"

type incomingChunk struct {
	Role             string `json:"role"`
	Type             string `json:"type"`
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
}

type outgoingChunk struct {
	ReasoningContent string      `json:"reasoning_content"`
	Content          string      `json:"content"`
	ReasoningSeconds int         `json:"reasoning_seconds"`
	Reference        interface{} `json:"reference"`
	Flag             string      `json:"flag"`
}

// CreateExternalToken 获取会话ID并拼接外部 token
func CreateExternalToken(ctx *gin.Context, accessToken string) (string, error) {
	if accessToken == "" {
		return "", errors.New("accessToken is empty")
	}

	cozeApiKey, agentID, err := coze.GetCozeAPIKeyByAgentCode(ctx, accessToken)
	if err != nil {
		logs.ErrorContextf(ctx, "get coze api key by agent code failed: %v", err)
		return "", err
	}
	if agentID == "" {
		return "", errors.New("agentID is empty")
	}

	conversationID, err := coze.CreateConversation(ctx, cozeApiKey, agentID, "")
	if err != nil {
		logs.ErrorContextf(ctx, "create coze conversation failed: %v", err)
		return "", err
	}

	return fmt.Sprintf("%s-%s", accessToken, conversationID), nil
}

// CreateExternalMessage stores a message payload and returns message ID.
func CreateExternalMessage(ctx *gin.Context, cozeApiKey, conversationID, botID, userID, content string) (string, error) {
	if cozeApiKey == "" {
		return "", errors.New("cozeApiKey is empty")
	}
	if conversationID == "" {
		return "", errors.New("conversationID is empty")
	}
	if botID == "" {
		return "", errors.New("botID is empty")
	}
	if content == "" {
		return "", errors.New("content is empty")
	}

	req := coze.CreateMessageRequest{
		ConversationID: conversationID,
		Role:           "user",
		Content:        content,
		ContentType:    "text",
		BotID:          botID,
		UserID:         userID,
	}
	messageID, err := coze.CreateMessage(ctx, cozeApiKey, req)
	if err != nil {
		logs.ErrorContextf(ctx, "create coze message failed: %v", err)
		return "", err
	}
	return messageID, nil
}

// CreateExternalChatStream starts a chat stream based on a stored message.
func CreateExternalChatStream(ctx *gin.Context, cozeApiKey, messageID string) error {
	if cozeApiKey == "" {
		return errors.New("cozeApiKey is empty")
	}
	if messageID == "" {
		return errors.New("messageID is empty")
	}

	resp, err := coze.CreateChat(ctx, cozeApiKey, messageID)
	if err != nil {
		logs.ErrorContextf(ctx, "create coze chat failed: %v", err)
		return err
	}
	defer resp.Body.Close()
	return streamCozeChatResponse(ctx, resp)
}

// GetPersonalAccessToken 获取 coze 个人 access token
func GetPersonalAccessToken(ctx *gin.Context) (coze.PersonalAccessTokenData, error) {
	data, err := coze.GetPersonalAccessToken(ctx)
	if err != nil {
		return coze.PersonalAccessTokenData{}, err
	}
	if data.APIKey == "" {
		return coze.PersonalAccessTokenData{}, errors.New("personal access token is empty")
	}
	return data, nil
}

// CreateCozeChatAPI 发起 coze /v3/chat 请求并将响应转发给客户端。
func CreateCozeChatAPI(ctx *gin.Context, cozeApiKey, botID, userID string, additionalMessages []coze.ChatV3Message, stream bool) error {
	if cozeApiKey == "" {
		return errors.New("cozeApiKey is empty")
	}
	if botID == "" {
		return errors.New("botID is empty")
	}
	if userID == "" {
		return errors.New("userID is empty")
	}
	if len(additionalMessages) == 0 {
		return errors.New("additionalMessages is empty")
	}

	autoSaveHistory := true
	payload := coze.ChatV3Request{
		BotID:              botID,
		UserID:             userID,
		Stream:             stream,
		AutoSaveHistory:    &autoSaveHistory,
		AdditionalMessages: additionalMessages,
	}

	resp, err := coze.CreateChatV3(ctx, cozeApiKey, payload, stream, "")
	if err != nil {
		logs.ErrorContextf(ctx, "create coze v3 chat failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if stream {
		return streamCozeChatResponseOpenAI(ctx, resp, botID)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "read coze v3 chat response failed: %v", err)
		return err
	}
	var chatResp coze.ChatV3ResponseEnvelope
	if err := json.Unmarshal(body, &chatResp); err != nil {
		logs.ErrorContextf(ctx, "unmarshal coze v3 chat response failed: %v", err)
		return err
	}
	if chatResp.Code != 0 {
		return fmt.Errorf("coze chat code: %d, msg: %s", chatResp.Code, chatResp.Msg)
	}
	chatID := chatResp.Data.ID
	conversationID := chatResp.Data.ConversationID
	if chatID == "" || conversationID == "" {
		return errors.New("coze chat response missing chat_id or conversation_id")
	}
	status := chatResp.Data.Status
	for !isChatTerminalStatus(status) {
		select {
		case <-ctx.Request.Context().Done():
			return ctx.Request.Context().Err()
		case <-time.After(time.Second):
		}
		latest, err := coze.RetrieveChatV3(ctx, cozeApiKey, conversationID, chatID)
		if err != nil {
			logs.ErrorContextf(ctx, "retrieve coze chat failed: %v", err)
			return err
		}
		status = latest.Status
	}

	msgResp, err := coze.ListChatV3Messages(ctx, cozeApiKey, conversationID, chatID)
	if err != nil {
		logs.ErrorContextf(ctx, "list coze chat messages failed: %v", err)
		return err
	}

	ctx.Writer.Header().Set("Content-Type", "application/json")
	ctx.Status(http.StatusOK)
	if err := json.NewEncoder(ctx.Writer).Encode(msgResp); err != nil {
		return err
	}
	return nil
}

type cozeStreamContext struct {
	RequestID string
	CreatedAt int64
	Model     string
	RoleSent  bool
}

type openAIChatStreamResponse struct {
	ID      string               `json:"id"`
	Choices []openAIChoiceStream `json:"choices"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Object  string               `json:"object"`
	Usage   *chattype.Usage      `json:"usage,omitempty"`
}

type openAIChoiceStream struct {
	FinishReason *string        `json:"finish_reason"`
	Index        int            `json:"index"`
	Delta        chattype.Delta `json:"delta"`
}

func streamCozeChatResponse(ctx *gin.Context, resp *http.Response) error {
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")

	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		return errors.New("response writer does not support flushing")
	}

	reader := bufio.NewReader(resp.Body)
	currentEvent := ""

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			logs.ErrorContextf(ctx, "stream coze chat failed: %v", err)
			return err
		}
		if len(line) == 0 {
			if err == io.EOF {
				break
			}
			continue
		}

		done := err == io.EOF
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "" {
			currentEvent = ""
			if done {
				break
			}
			continue
		}
		if strings.HasPrefix(trimmed, "event:") {
			currentEvent = strings.TrimSpace(strings.TrimPrefix(trimmed, "event:"))
			if done {
				break
			}
			continue
		}
		if strings.HasPrefix(trimmed, "data:") && currentEvent == cozeEventConversationDelta {
			data := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
			if err := handleCozeStreamData(ctx, flusher, data); err != nil {
				return err
			}
		}

		if done {
			break
		}
	}

	return nil
}

func isChatTerminalStatus(status string) bool {
	switch status {
	case "completed", "required_action", "canceled", "failed":
		return true
	default:
		return false
	}
}

func streamCozeChatResponseOpenAI(ctx *gin.Context, resp *http.Response, model string) error {
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")

	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		return errors.New("response writer does not support flushing")
	}

	reader := bufio.NewReader(resp.Body)
	currentEvent := ""
	var dataBuilder strings.Builder
	streamCtx := &cozeStreamContext{
		RequestID: runtime.RequestID(ctx),
		CreatedAt: time.Now().Unix(),
		Model:     model,
		RoleSent:  false,
	}
	if streamCtx.Model == "" {
		streamCtx.Model = "coze"
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			logs.ErrorContextf(ctx, "stream coze chat failed: %v", err)
			return err
		}
		if len(line) == 0 {
			if err == io.EOF {
				break
			}
			continue
		}

		done := err == io.EOF
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "" {
			if dataBuilder.Len() > 0 {
				if err := dispatchCozeStreamEventOpenAI(ctx, flusher, currentEvent, dataBuilder.String(), streamCtx); err != nil {
					return err
				}
				dataBuilder.Reset()
			}
			currentEvent = ""
			if done {
				break
			}
			continue
		}
		if strings.HasPrefix(trimmed, ":") {
			if done {
				break
			}
			continue
		}
		if strings.HasPrefix(trimmed, "event:") {
			currentEvent = strings.TrimSpace(strings.TrimPrefix(trimmed, "event:"))
			if done {
				break
			}
			continue
		}
		if strings.HasPrefix(trimmed, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
			if dataBuilder.Len() > 0 {
				dataBuilder.WriteByte('\n')
			}
			dataBuilder.WriteString(data)
		}

		if done {
			if dataBuilder.Len() > 0 {
				if err := dispatchCozeStreamEventOpenAI(ctx, flusher, currentEvent, dataBuilder.String(), streamCtx); err != nil {
					return err
				}
			}
			break
		}
	}

	if err := writeOpenAIFinishChunk(ctx, flusher, streamCtx); err != nil {
		return err
	}
	if _, err := ctx.Writer.Write([]byte("data: [DONE]\n\n")); err != nil {
		logs.ErrorContextf(ctx, "stream coze chat write [DONE] failed: %v", err)
		return err
	}
	flusher.Flush()

	return nil
}

func dispatchCozeStreamEventOpenAI(ctx *gin.Context, flusher http.Flusher, event, data string, streamCtx *cozeStreamContext) error {
	switch event {
	case cozeEventConversationDelta, "":
		return handleCozeStreamDataOpenAI(ctx, flusher, data, streamCtx)
	default:
		return nil
	}
}

func handleCozeStreamData(ctx *gin.Context, flusher http.Flusher, data string) error {
	if data == "" || data == "[DONE]" {
		return nil
	}

	var chunk incomingChunk
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		logs.ErrorContextf(ctx, "stream coze chat unmarshal failed: %v", err)
		return nil
	}

	switch {
	case chunk.Role == "assistant" && chunk.Type == "answer":
		out := outgoingChunk{
			ReasoningContent: chunk.ReasoningContent,
			Content:          chunk.Content,
			ReasoningSeconds: 0,
			Reference:        nil,
			Flag:             "",
		}
		if err := writeJSONLine(ctx, flusher, out); err != nil {
			logs.ErrorContextf(ctx, "stream coze chat write chunk failed: %v", err)
			return err
		}
	case chunk.Role == "assistant" && chunk.Type == "follow_up":
		// TODO: handle follow_up type stream chunk
	}

	return nil
}

func handleCozeStreamDataOpenAI(ctx *gin.Context, flusher http.Flusher, data string, streamCtx *cozeStreamContext) error {
	if data == "" || data == "[DONE]" {
		return nil
	}

	var chunk incomingChunk
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		logs.ErrorContextf(ctx, "stream coze chat unmarshal failed: %v", err)
		return nil
	}

	switch {
	case chunk.Role == "assistant" && chunk.Type == "answer":
		if chunk.ReasoningContent == "" && chunk.Content == "" {
			return nil
		}
		delta := chattype.Delta{
			Content:          chunk.Content,
			ReasoningContent: chunk.ReasoningContent,
		}
		if !streamCtx.RoleSent {
			role := "assistant"
			if chunk.Role != "" {
				role = chunk.Role
			}
			delta.Role = role
			streamCtx.RoleSent = true
		}
		out := openAIChatStreamResponse{
			ID:      streamCtx.RequestID,
			Created: streamCtx.CreatedAt,
			Model:   streamCtx.Model,
			Object:  "chat.completion.chunk",
			Choices: []openAIChoiceStream{
				{
					Index:        0,
					FinishReason: nil,
					Delta:        delta,
				},
			},
			Usage: nil,
		}
		if err := writeSSEJSON(ctx, flusher, out); err != nil {
			if isBrokenPipeError(err) {
				return nil
			}
			logs.ErrorContextf(ctx, "stream coze chat write sse chunk failed: %v", err)
			return err
		}
	case chunk.Role == "assistant" && chunk.Type == "follow_up":
		// TODO: handle follow_up type stream chunk
	}

	return nil
}

func writeJSONLine(ctx *gin.Context, flusher http.Flusher, v interface{}) error {
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := ctx.Writer.Write(payload); err != nil {
		return err
	}
	if _, err := ctx.Writer.Write([]byte("\n")); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func writeOpenAIFinishChunk(ctx *gin.Context, flusher http.Flusher, streamCtx *cozeStreamContext) error {
	finishReason := "stop"
	out := openAIChatStreamResponse{
		ID:      streamCtx.RequestID,
		Created: streamCtx.CreatedAt,
		Model:   streamCtx.Model,
		Object:  "chat.completion.chunk",
		Choices: []openAIChoiceStream{
			{
				Index:        0,
				FinishReason: &finishReason,
				Delta:        chattype.Delta{},
			},
		},
		Usage: nil,
	}
	return writeSSEJSON(ctx, flusher, out)
}

func writeSSEJSON(ctx *gin.Context, flusher http.Flusher, v interface{}) error {
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := ctx.Writer.Write([]byte("data: ")); err != nil {
		return err
	}
	if _, err := ctx.Writer.Write(payload); err != nil {
		return err
	}
	if _, err := ctx.Writer.Write([]byte("\n\n")); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func isBrokenPipeError(err error) bool {
	return strings.Contains(err.Error(), "broken pipe") ||
		strings.Contains(err.Error(), "connection reset by peer")
}
