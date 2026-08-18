package llmchat

import (
	"time"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/logs"
)

// ChatResponseFont 流式或非流式都返回前端
func (w *LLMChatWrapper) ChatResponseFont(onMessage func(*chattype.ChatStreamResponseBody) error) (*QaRes, error) {
	now := time.Now()
	res := &QaRes{}
	request, err := w.ChatRequest()
	if err != nil {
		logs.ErrorContextf(w.ctx, "failed to chat: %v", err)
		return res, err
	}
	if w.req.Stream {
		res, err = w.ChatStreamResult(request, onMessage)
		if err != nil {
			logs.ErrorContextf(w.ctx, "failed to chat stream: %v", err)
			return res, err
		}
	} else {
		res, err = w.ChatResponse(request)
		if err != nil {
			logs.ErrorContextf(w.ctx, "failed to chat: %v", err)
			return res, err
		}
	}
	duration := time.Since(now)
	res.CostSeconds = int(duration.Seconds())
	return res, nil
}

// InternalChatResponse 内部调用函数，流式返回时直接返回前端，非流式时返回内部结构体
func (w *LLMChatWrapper) InternalChatResponse(onMessage func(*chattype.ChatStreamResponseBody) error) (*QaRes, error) {
	now := time.Now()
	res := &QaRes{}
	request, err := w.ChatRequest()
	if err != nil {
		logs.ErrorContextf(w.ctx, "failed to chat: %v", err)
		return res, err
	}
	if w.req.Stream {
		res, err = w.ChatStreamResult(request, onMessage)
		if err != nil {
			logs.ErrorContextf(w.ctx, "failed to chat stream: %v", err)
			return res, err
		}
	} else {
		res, err = w.ChatResult(request)
		if err != nil {
			logs.ErrorContextf(w.ctx, "failed to chat: %v", err)
			return res, err
		}
	}
	duration := time.Since(now)
	res.CostSeconds = int(duration.Seconds())
	return res, nil
}

// APIChatResponse 直接返回openai api规范结果到调用方
func (w *LLMChatWrapper) APIChatResponse(onMessage func(*chattype.ChatStreamResponseBody) error) (*QaRes, error) {
	now := time.Now()
	res := &QaRes{}
	request, err := w.ChatRequest()
	if err != nil {
		logs.ErrorContextf(w.ctx, "failed to chat: %v", err)
		return res, err
	}
	if w.req.Stream {
		res, err = w.ChatStreamResultSSE(request, onMessage)
		if err != nil {
			logs.ErrorContextf(w.ctx, "failed to chat stream: %v", err)
			return res, err
		}
	} else {
		res, err = w.ChatResponse(request)
		if err != nil {
			logs.ErrorContextf(w.ctx, "failed to chat: %v", err)
			return res, err
		}
	}
	duration := time.Since(now)
	res.CostSeconds = int(duration.Seconds())
	return res, nil
}
