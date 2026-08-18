package llmchat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/logs"
)

type LLMChatWrapperV2 struct {
	ctx   context.Context
	req   *ChatReqBody
	model *chattype.ChatModel
}

func NewLLmChatWrapperV2(ctx context.Context, req *ChatReqBody, model *chattype.ChatModel) *LLMChatWrapperV2 {
	req.Model = model.ModelName
	return &LLMChatWrapperV2{
		ctx:   ctx,
		req:   req,
		model: model,
	}
}

// InternalChatResponse 内部调用函数，流式返回时直接返回前端，非流式时返回内部结构体
func (w *LLMChatWrapperV2) InvokeChat() (*QaRes, error) {
	now := time.Now()
	res := &QaRes{}
	request, err := w.ChatRequest()
	if err != nil {
		logs.ErrorContextf(w.ctx, "failed to chat: %v", err)
		return res, err
	}
	res, err = w.ChatResult(request)
	if err != nil {
		logs.ErrorContextf(w.ctx, "failed to chat: %v", err)
		return res, err
	}
	duration := time.Since(now)
	res.CostSeconds = int(duration.Seconds())
	return res, nil
}

// ChatRequest 发请求到模型 最底层发送临门一脚
func (w *LLMChatWrapperV2) ChatRequest() (*http.Response, error) {
	jsonPayload, err := w.req.ToString()
	if err != nil {
		logs.ErrorContext(w.ctx, "ChatRequest::ToString Failed to marshal: %v", err)
		return nil, err
	}
	logs.InfoContextf(w.ctx, "chat request with json:\n%s\n", jsonPayload)
	// 创建 HTTP
	req, err := http.NewRequestWithContext(w.ctx, "POST", w.model.ModelUrl, strings.NewReader(jsonPayload))
	if err != nil {
		logs.ErrorContextf(w.ctx, "ChatRequest Failed to create HTTP request: %s", err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.model.APIKey)

	// 创建 HTTP 客户端
	client := &http.Client{}
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(w.ctx, "ChatRequest Failed to make HTTP request: %v", err)
		return nil, fmt.Errorf("ChatRequest Failed to make HTTP request: %w", err)
	}
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(w.ctx, "ChatRequest Error status: %s,err: %s", resp.Status, string(body))
		return nil, fmt.Errorf("ChatRequest Error status: %s,err: %s", resp.Status, string(body))
	}

	// 返回响应
	return resp, nil
}

// ChatResult 内部调用不返回前端结果 非流式
func (w *LLMChatWrapperV2) ChatResult(resp *http.Response) (*QaRes, error) {
	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(w.ctx, "ChatResponse read response body failed: %v", err)
		return nil, fmt.Errorf("read response failed: %w", err)
	}
	// 解析响应
	var response chattype.ChatResponseBody
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		logs.ErrorContextf(w.ctx, "ChatResponse error unmarshalling json: %v", err)
		return nil, fmt.Errorf("response Parsing Error: %w", err)
	}
	// 验证响应有效性
	if len(response.Choices) == 0 && response.Error.Code == 0 {
		logs.ErrorContextf(w.ctx, "ChatResponse response is empty")
		return nil, fmt.Errorf("invalid API Response")
	}
	resault := &QaRes{}
	resault.Content = response.Choices[0].Message.Content
	resault.Reasoning = response.Choices[0].Message.ReasoningContent
	resault.Usage = Usage(response.Usage)
	return resault, nil
}
