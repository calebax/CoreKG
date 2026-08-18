package qachat

import (
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
	"github.com/ygpkg/yg-go/logs"
)

// LLmChat  单纯模型问答
func (w *ChatWapper) LLmChat(stream bool) error {
	// 历史记录
	questions, err := chatquestion.ListSessionQuestionsByUin(w.ctx, w.session.Uin, w.session.ID)
	if err != nil {
		logs.ErrorContextf(w.ctx, "get session questions error: %v", err)
		return err
	}
	messages := GetMessages(questions)
	request := &llmchat.ChatReqBody{
		Messages: messages,
		Stream:   stream,
	}
	if stream {
		request.StreamOptions = llmchat.NewStreamOptions()
	}
	wrapper := llmchat.NewLLmChatWrapper(w.ctx, request, w.model)
	res, err := wrapper.ChatResponseFont(nil)
	if err != nil {
		logs.ErrorContextf(w.ctx, "LLmChtat failed ,err %s", err)
		// return err
	}
	if res != nil {
		w.question.Source.Answer = res.Content
		w.question.Source.Reasoning = res.Reasoning
		w.question.Source.ReasoningSeconds = res.ReasoningTime
		w.question.Source.CostSeconds = res.CostSeconds
		w.question.Source.OutToken = res.Usage.CompletionTokens
		w.question.Source.CacheHitToken = res.Usage.PromptCacheHitTokens
		w.question.Source.CacheMissToken = res.Usage.PromptCacheMissTokens
		w.question.Source.TotalTokens = res.Usage.TotalTokens
		w.question.Source.Status = chattype.QuestionStatusAnswered
	}
	return err
}

// GetMessages 根据历史会话获取发给模型的message
func GetMessages(questions []*chattype.ChatQuestion) []*llmchat.Message {
	var messages []*llmchat.Message
	if len(questions) > 1 {
		for i := 0; i < len(questions)-1; i++ {
			messages = append(messages, &llmchat.Message{
				Role:    llmchat.MessageRoleUser,
				Content: questions[i].Source.Question,
			})
			messages = append(messages, &llmchat.Message{
				Role:    llmchat.MessageRoleAssistant,
				Content: questions[i].Source.Answer,
			})
		}
	}
	messages = append(messages, &llmchat.Message{
		Role:    llmchat.MessageRoleUser,
		Content: questions[len(questions)-1].Source.Question,
	})
	return messages
}
