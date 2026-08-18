package qachat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/keqa"
	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
	"github.com/ygpkg/yg-go/logs"
)

func (w *ChatWapper) ForestChat() error {
	history, err := GetForestChatHistory(w.ctx, w.session)
	if err != nil {
		logs.ErrorContextf(w.ctx, "ForestChat GetForestChatHistory error: %v", err)
		return err
	}
	if len(w.question.Source.ImageUrlList) > 0 {
		// 解析多模态
		w.question.Source.ImageContent = fmt.Sprintf("\n用户上传了%v张图片，根据描述信息找出与问题相关的可用信息并用于对话中，每条图片的具体描述信息如下：\n", len(w.question.Source.ImageUrlList))
		for i, url := range w.question.Source.ImageUrlList {
			res, err := keqa.DoImageParseRequest(w.ctx, url)
			if err != nil {
				logs.ErrorContextf(w.ctx, "[ForestChat] Failed to parse image: %v", err)
			}
			w.question.Source.ImageContent += fmt.Sprintf("第%v张\n%v\n", i+1, res)
		}
	}
	// 写入searching
	keqa.WriteFlag(w.ctx, w.question.Source.ReqID, llmchat.FlagSearching)
	wrapper, err := keqa.HandelSearchReference(w.ctx, w.session.ForestIDList.Slice(), w.session.FileIDList.Slice(), w.session.EsIndex, w.question.Source.Question+w.question.Source.ImageContent)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ForestChat] Failed to HandelSearchReference: %v", err)
		return err
	}
	// 查找问答对
	fqa, err := wrapper.SearchWrapper.FindFQAByQuestion()
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ForestChat] Failed to FindFQAByQuestion: %v", err)
		return err
	}
	if len(fqa.Hits.Hits) != 0 {
		logs.InfoContextf(w.ctx, "[ForestChat] FindFQAByQuestion result: %v", len(fqa.Hits.Hits))
		keqa.WriteSearchQA(w.ctx, fqa.Hits.Hits[0].Source.QAAnswer, w.question.Source.ReqID)
		w.question.Source.Answer = fqa.Hits.Hits[0].Source.QAAnswer
		w.question.Source.Status = chattype.QuestionStatusAnswered
		return nil
	}
	// 开始意图识别，正常搜索还是总结性问题
	intention, err := keqa.IntentionRecognition(w.ctx, w.question)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ForestChat] Failed to IntentionRecognition: %v", err)
		return err
	}
	forestChatWapper := keqa.NewForestWrapper(w.ctx, w.question, wrapper, history)
	desc := false
	switch intention {
	case "C", "c":
		_, err = forestChatWapper.DescriptionChat(true)
		desc = true
	default:
		_, err = forestChatWapper.DefaultRerankChat(true)
	}
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ForestChat] Failed to DefaultChat: %v", err)
		return err
	}
	res, err := forestChatWapper.PromptChat(w.model.ID, desc, w.session.PromptMode)
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

// GetForestChatHistory 获取chat历史记录
func GetForestChatHistory(ctx context.Context, session *chattype.ChatSession) (string, error) {
	quesrions, err := chatquestion.ListSessionQuestionsByUin(ctx, session.Uin, session.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestChatHistory ListSessionQuestions error: %v", err)
		return "", err
	}
	chats := []*ForestChatHistory{}
	for _, qa := range quesrions {
		if qa.Source.Status != chattype.QuestionStatusAnswered {
			continue
		}
		chats = append(chats, &ForestChatHistory{
			Question: qa.Source.Question,
			Answer:   qa.Source.Answer,
		})
	}
	chatsJSON, err := json.Marshal(chats)
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestChatHistory json.Marshal error: %v", err)
		return "", err
	}
	return string(chatsJSON), nil
}

type ForestChatHistory struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

func getMountForestMessage(ctx *gin.Context, esIndex string, opt chattype.ForestChatOption, question *chattype.ChatQuestion, writeRef bool) (string, error) {
	wrapper, err := keqa.HandelSearchReference(ctx, opt.ForestIDs, nil, esIndex, question.Source.Question)
	if err != nil {
		logs.ErrorContextf(ctx, "[ChatWapper] GetAgentMessages Failed to HandelSearchReference: %v", err)
		return "", err
	}
	// 查找问答对
	fqa, err := wrapper.SearchWrapper.FindFQAByQuestion()
	if err != nil {
		logs.ErrorContextf(ctx, "[ChatWapper] GetAgentMessages Failed to FindFQAByQuestion: %v", err)
		return "", err
	}
	if len(fqa.Hits.Hits) != 0 {
		logs.InfoContextf(ctx, "[ChatWapper] GetAgentMessages FindFQAByQuestion result: %v", len(fqa.Hits.Hits))
		keqa.WriteSearchQA(ctx, fqa.Hits.Hits[0].Source.QAAnswer, question.Source.ReqID)
		question.Source.Status = chattype.QuestionStatusAnswered
		question.Source.Answer = fqa.Hits.Hits[0].Source.QAAnswer
		return "", err
	}
	// 开始意图识别，正常搜索还是总结性问题
	intention, err := keqa.IntentionRecognition(ctx, question)
	if err != nil {
		logs.ErrorContextf(ctx, "[ForestChat] Failed to IntentionRecognition: %v", err)
		return "", err
	}
	forestChatWapper := keqa.NewForestWrapper(ctx, question, wrapper, "")
	searchStr := ""
	switch intention {
	case "C", "c":
		searchStr, err = forestChatWapper.DescriptionChat(writeRef)
	default:
		searchStr, err = forestChatWapper.DefaultRerankChat(writeRef)
	}
	if err != nil {
		logs.ErrorContextf(ctx, "[ForestChat] Failed to DefaultChat: %v", err)
		return "", err
	}
	input := chattype.InputList{
		{Name: "input1", Value: searchStr},
	}
	// TODO 先写死后期大改
	updatedTemplate := replaceInputPlaceholders(agentPrompt, &input)
	return updatedTemplate, nil
}
