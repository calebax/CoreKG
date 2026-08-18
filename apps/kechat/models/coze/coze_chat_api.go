package coze

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

const (
	publicAgentExternalTokenPath = "/api/public/agent/external_token"
	personalAccessTokenPath      = "/api/permission_api/pat/get_personal_access_token"
	createConversationPath       = "/v1/conversation/create"
	chatV3Path                   = "/v3/chat"
	chatV3RetrievePath           = "/v3/chat/retrieve"
	chatV3MessageListPath        = "/v3/chat/message/list"
)

type CreateConversationResponse struct {
	ConversationID string `json:"id"`
}

type ChatV3Message struct {
	Role        string `json:"role"`
	Content     string `json:"content"`
	ContentType string `json:"content_type,omitempty"`
}

type ChatV3Request struct {
	BotID              string          `json:"bot_id"`
	Role               string          `json:"role,omitempty"`
	UserID             string          `json:"user_id,omitempty"`
	Stream             bool            `json:"stream"`
	AutoSaveHistory    *bool           `json:"auto_save_history,omitempty"`
	AdditionalMessages []ChatV3Message `json:"additional_messages,omitempty"`
}

type ChatV3Response struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	Status         string `json:"status"`
	CreatedAt      int64  `json:"created_at"`
}

type ChatV3ResponseEnvelope struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data ChatV3Response `json:"data"`
}

type ChatV3RetrieveData struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	BotID          string `json:"bot_id"`
	Status         string `json:"status"`
	CreatedAt      int64  `json:"created_at"`
	CompletedAt    int64  `json:"completed_at"`
	FailedAt       int64  `json:"failed_at"`
}

type ChatV3MessageItem struct {
	ID               string                 `json:"id"`
	ConversationID   string                 `json:"conversation_id"`
	BotID            string                 `json:"bot_id"`
	Role             string                 `json:"role"`
	Type             string                 `json:"type"`
	Content          string                 `json:"content"`
	ContentType      string                 `json:"content_type"`
	MetaData         map[string]interface{} `json:"meta_data,omitempty"`
	ChatID           string                 `json:"chat_id"`
	SectionID        string                 `json:"section_id"`
	CreatedAt        int64                  `json:"created_at"`
	ReasoningContent string                 `json:"reasoning_content"`
}

type chatV3MessageListEnvelope struct {
	Messages []ChatV3MessageItem `json:"Messages"`
	Code     int                 `json:"code"`
	Msg      string              `json:"msg"`
}

type CreateMessageRequest struct {
	ConversationID string                 `json:"conversation_id"`
	Role           string                 `json:"role"`
	Content        string                 `json:"content"`
	ContentType    string                 `json:"content_type,omitempty"`
	MetaData       map[string]interface{} `json:"meta_data,omitempty"`
	BotID          string                 `json:"bot_id"`
	UserID         string                 `json:"user_id"`
}

type CreateMessageResponse struct {
	MessageID string `json:"id"`
}

type PersonalAccessTokenData struct {
	APIKey   string `json:"api_key"`
	ExpireAt int64  `json:"expire_at"`
}

// GetCozeAPIKeyByAgentCode 通过 agentCode 获取 cozeApiKey(auth_token) 和 agentId
func GetCozeAPIKeyByAgentCode(ctx *gin.Context, agentCode string) (string, string, error) {
	if agentCode == "" {
		return "", "", errors.New("agentCode is empty")
	}

	payload := map[string]interface{}{
		"short_code": agentCode,
	}

	var data PublicAgentExternalTokenData
	if err := CozeRequest(ctx, publicAgentExternalTokenPath, payload, &data, WithCozeDataOnly()); err != nil {
		logs.ErrorContextf(ctx, "request coze api key by agent code failed: %v", err)
		return "", "", err
	}

	return data.CozeApiKey, data.AgentID, nil
}

// GetPersonalAccessToken 获取个人 access token
func GetPersonalAccessToken(ctx *gin.Context) (PersonalAccessTokenData, error) {
	var data PersonalAccessTokenData
	if err := CozeRequest(ctx, personalAccessTokenPath, nil, &data, WithCozeDataOnly()); err != nil {
		logs.ErrorContextf(ctx, "get coze personal access token failed: %v", err)
		return PersonalAccessTokenData{}, err
	}
	return data, nil
}

// CreateConversation 创建会话，返回 conversation_id
func CreateConversation(ctx *gin.Context, cozeApiKey string, botID string, userID string) (string, error) {
	if cozeApiKey == "" {
		return "", errors.New("cozeApiKey is empty")
	}
	if botID == "" {
		return "", errors.New("botID is empty")
	}

	payload := map[string]interface{}{
		"bot_id": botID,
	}
	if userID != "" {
		payload["user_id"] = userID
	}

	var data CreateConversationResponse
	if err := CozeRequest(ctx, createConversationPath, payload, &data, WithCozeToken(cozeApiKey), WithCozeDataOnly()); err != nil {
		logs.ErrorContextf(ctx, "create coze conversation failed: %v", err)
		return "", err
	}
	return data.ConversationID, nil
}

// CreateChatV3 发起对话(chat v3)
func CreateChatV3(ctx *gin.Context, cozeApiKey string, payload ChatV3Request, stream bool, conversationID string) (*http.Response, error) {
	if cozeApiKey == "" {
		return nil, errors.New("cozeApiKey is empty")
	}
	if payload.AutoSaveHistory == nil {
		autoSaveHistory := true
		payload.AutoSaveHistory = &autoSaveHistory
	}

	baseURL, err := settings.GetText("corekg", "coze_url")
	if err != nil {
		logs.ErrorContextf(ctx, "get coze url error, %s", err.Error())
		return nil, err
	}

	apiURL, err := url.JoinPath(baseURL, chatV3Path)
	if err != nil {
		logs.ErrorContextf(ctx, "join coze url err %s", err.Error())
		return nil, err
	}
	if conversationID != "" {
		parsedURL, err := url.Parse(apiURL)
		if err != nil {
			logs.ErrorContextf(ctx, "parse coze url err %s", err.Error())
			return nil, err
		}
		query := parsedURL.Query()
		query.Set("conversation_id", conversationID)
		parsedURL.RawQuery = query.Encode()
		apiURL = parsedURL.String()
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		logs.ErrorContextf(ctx, "marshal request body error, %s", err.Error())
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %s", err.Error())
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cozeApiKey)
	if stream {
		httpReq.Header.Set("Accept", "text/event-stream")
	} else {
		httpReq.Header.Set("Accept", "application/json")
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %s", err.Error())
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		logs.ErrorContextf(ctx, "coze chat status %d body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("coze chat failed with status: %d", resp.StatusCode)
	}

	return resp, nil
}

// CreateChat 发起对话(chat v3)，根据 messageID 获取消息内容并返回流式响应(调用方负责关闭 Body)
func CreateChat(ctx *gin.Context, cozeApiKey, messageID string) (*http.Response, error) {
	if cozeApiKey == "" {
		return nil, errors.New("cozeApiKey is empty")
	}
	if messageID == "" {
		return nil, errors.New("messageID is empty")
	}
	stored, ok, err := getMessage(messageID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("message not found")
	}
	req := stored.Req
	if req.ConversationID == "" {
		return nil, errors.New("conversationID is empty")
	}
	if req.BotID == "" {
		return nil, errors.New("botID is empty")
	}
	if req.Role == "" {
		req.Role = "user"
	}
	if req.Content == "" {
		return nil, errors.New("content is empty")
	}
	if req.ContentType == "" {
		req.ContentType = "text"
	}

	chatReq := ChatV3Request{
		BotID:  req.BotID,
		Role:   req.Role,
		UserID: req.UserID,
		Stream: true,
		AdditionalMessages: []ChatV3Message{
			{
				Role:        req.Role,
				Content:     req.Content,
				ContentType: req.ContentType,
			},
		},
	}
	resp, err := CreateChatV3(ctx, cozeApiKey, chatReq, true, req.ConversationID)
	if err != nil {
		return nil, err
	}

	if err := deleteMessage(messageID); err != nil {
		logs.WarnContextf(ctx, "delete coze message from redis failed: %v", err)
	}
	return resp, nil
}

func RetrieveChatV3(ctx *gin.Context, cozeApiKey, conversationID, chatID string) (ChatV3RetrieveData, error) {
	if cozeApiKey == "" {
		return ChatV3RetrieveData{}, errors.New("cozeApiKey is empty")
	}
	if conversationID == "" {
		return ChatV3RetrieveData{}, errors.New("conversationID is empty")
	}
	if chatID == "" {
		return ChatV3RetrieveData{}, errors.New("chatID is empty")
	}

	query := map[string]string{
		"conversation_id": conversationID,
		"chat_id":         chatID,
	}
	var data ChatV3RetrieveData
	if err := CozeRequest(
		ctx,
		chatV3RetrievePath,
		nil,
		&data,
		WithCozeToken(cozeApiKey),
		WithCozeMethod(http.MethodGet),
		WithCozeQuery(query),
		WithCozeDataOnly(),
	); err != nil {
		return ChatV3RetrieveData{}, err
	}
	return data, nil
}

func ListChatV3Messages(ctx *gin.Context, cozeApiKey, conversationID, chatID string) (chattype.ChatResponseBody, error) {
	if cozeApiKey == "" {
		return chattype.ChatResponseBody{}, errors.New("cozeApiKey is empty")
	}
	if conversationID == "" {
		return chattype.ChatResponseBody{}, errors.New("conversationID is empty")
	}
	if chatID == "" {
		return chattype.ChatResponseBody{}, errors.New("chatID is empty")
	}

	query := map[string]string{
		"conversation_id": conversationID,
		"chat_id":         chatID,
	}
	var resp chatV3MessageListEnvelope
	if err := CozeRequest(
		ctx,
		chatV3MessageListPath,
		nil,
		&resp,
		WithCozeToken(cozeApiKey),
		WithCozeMethod(http.MethodGet),
		WithCozeQuery(query),
	); err != nil {
		return chattype.ChatResponseBody{}, err
	}

	var (
		contentBuilder   strings.Builder
		reasoningBuilder strings.Builder
		createdAt        int64
		model            string
	)
	for _, msg := range resp.Messages {
		if msg.Role != "assistant" || msg.Type != "answer" {
			continue
		}
		contentType := msg.ContentType
		if contentType == "" {
			contentType = "text"
		}
		if contentType != "text" {
			continue
		}
		if model == "" {
			model = msg.BotID
		}
		if msg.CreatedAt > createdAt {
			createdAt = msg.CreatedAt
		}
		if msg.Content != "" {
			contentBuilder.WriteString(msg.Content)
		}
		if msg.ReasoningContent != "" {
			reasoningBuilder.WriteString(msg.ReasoningContent)
		}
	}

	if model == "" {
		model = "coze"
	}
	if createdAt == 0 {
		createdAt = time.Now().Unix()
	}

	message := chattype.ChatMessage{
		Content: contentBuilder.String(),
		Role:    chattype.RoleAssistant,
	}
	if reasoningBuilder.Len() > 0 {
		message.ReasoningContent = reasoningBuilder.String()
	}

	out := chattype.ChatResponseBody{
		ID:      chatID,
		Created: createdAt,
		Model:   model,
		Object:  "chat.completion",
		Choices: []chattype.Choice{
			{
				Index:        0,
				FinishReason: "stop",
				Message:      message,
			},
		},
		Usage: chattype.Usage{},
	}
	return out, nil
}

// CreateMessage 创建消息，返回 message_id
func CreateMessage(ctx *gin.Context, cozeApiKey string, req CreateMessageRequest) (string, error) {
	if cozeApiKey == "" {
		return "", errors.New("cozeApiKey is empty")
	}
	if req.ConversationID == "" {
		return "", errors.New("conversationID is empty")
	}
	if req.BotID == "" {
		return "", errors.New("botID is empty")
	}
	if req.Role == "" {
		req.Role = "user"
	}
	if req.Content == "" {
		return "", errors.New("content is empty")
	}
	if req.ContentType == "" {
		req.ContentType = "text"
	}

	messageID, err := storeMessage(req)
	if err != nil {
		logs.ErrorContextf(ctx, "store coze message to redis failed: %v", err)
		return "", err
	}
	return messageID, nil
}
