package chatclient

import (
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/apikey"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

type InternalChat struct {
	ctx    *gin.Context
	step   int
	reqid  string
	apiKey string
	req    *chattype.ChatRequestBody
}

func NewInternalChat(ctx *gin.Context, reqid, apikey string, step int, req *chattype.ChatRequestBody) (*InternalChat, error) {
	if apikey == "" {
		sysAPIKey, err := settings.GetText(global.SettingGroupKnowledge, global.SettingKeySystemLlmAPIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get system api key: %v", err)
		}
		apikey = sysAPIKey
	}
	return &InternalChat{
		ctx:    ctx,
		step:   step,
		reqid:  reqid,
		apiKey: apikey,
		req:    req,
	}, nil
}

// AgentChatInternal 通过内部调用获取机器人问答
func (w *InternalChat) AgentChatInternal(onMessage func(*chattype.ChatStreamResponseBody) error) (*llmchat.QaRes, error) {
	// 获取机器人信息
	agentInfo, err := chatagent.GetAgentDetailByName(w.ctx, w.req.Model)
	if err != nil {
		logs.ErrorContextf(w.ctx, "failed to fetch agent info: %s", err)
		return nil, err
	}
	var modelID uint
	if w.req.LLMModelID != 0 {
		modelID = w.req.LLMModelID
	} else {
		modelIDs := agentInfo.ChatModelIDs.Slice()
		if len(modelIDs) == 0 {
			logs.ErrorContextf(w.ctx, "AgentChat agent has no model ,agent id %d", agentInfo.ID)
			return nil, fmt.Errorf("agent has no model")
		}
		modelID = modelIDs[0]
	}
	// 获取模型信息
	model, err := chatmodel.GetModelByID(w.ctx, modelID)
	if err != nil {
		logs.ErrorContextf(w.ctx, "AgentChat GetModelByID failed ,err %s", err)
		return nil, err
	}
	apikeyInfo, err := apikey.GetAPIKeyInfo(w.ctx, w.apiKey)
	if err != nil {
		logs.ErrorContextf(w.ctx, "failed to validate API key: %v", err)
		return nil, err
	}
	ques := &chattype.ChatQuestion{
		Source: &chattype.Question{
			CompanyID:    apikeyInfo.CompanyID,
			Uin:          agentInfo.Uin,
			AgentStep:    w.step,
			ReqID:        w.reqid,
			Status:       chattype.QuestionStatusPending,
			ApiKeyID:     apikeyInfo.ID,
			BaseAgentID:  agentInfo.ID,
			AgentVersion: agentInfo.Version,
			ModelID:      model.ID,
			UserInput:    w.req,
			AgentName:    agentInfo.Name,
		},
	}
	messages := getAgentMessages(agentInfo, ques)
	request := &llmchat.ChatReqBody{
		Messages:    messages,
		Stream:      ques.Source.UserInput.Stream,
		Temperature: &agentInfo.Temperature,
	}
	if ques.Source.UserInput.Stream {
		request.StreamOptions = llmchat.NewStreamOptions()
	}
	wrapper := llmchat.NewLLmChatWrapper(w.ctx, request, model)
	// 直接返回模型原始结果到调用方
	res, err := wrapper.InternalChatResponse(onMessage)

	if res != nil {
		ques.Source.Answer = res.Content
		ques.Source.Reasoning = res.Reasoning
		ques.Source.ReasoningSeconds = res.ReasoningTime
		ques.Source.CostSeconds = res.CostSeconds
		ques.Source.OutToken = res.Usage.CompletionTokens
		ques.Source.CacheHitToken = res.Usage.PromptCacheHitTokens
		ques.Source.CacheMissToken = res.Usage.PromptCacheMissTokens
		ques.Source.TotalTokens = res.Usage.TotalTokens
		ques.Source.Status = chattype.QuestionStatusAnswered
	}
	if err != nil {
		logs.ErrorContextf(w.ctx, "AgentChat AgentAPIChat failed ,err %s", err)
		// return
		ques.Source.Status = chattype.QuestionStatusError
	}

	err = chatquestion.CreateQuestion(w.ctx, ques)
	if err != nil {
		logs.ErrorContextf(w.ctx, "AgentChat CreateQuestion err: %v", err)
		return nil, err
	}
	return res, err
}

func getAgentMessages(agentInfo *chatagent.AgentWithVersion, question *chattype.ChatQuestion) []*llmchat.Message {
	messages := []*llmchat.Message{}
	if agentInfo.AgentType == chattype.AgentTypeRolePlay {
		messages = append(messages, &llmchat.Message{
			Role:    llmchat.MessageRoleSystem,
			Content: agentInfo.PromptTemplate,
		})
		messages = append(messages, &llmchat.Message{
			Role:    llmchat.MessageRoleAssistant,
			Content: agentInfo.GreetingMessage,
		})
		messages = append(messages, &llmchat.Message{
			Role:    llmchat.MessageRoleUser,
			Content: question.Source.Question,
		})
	}
	if agentInfo.AgentType == chattype.AgentTypePrompt || len(question.Source.UserInput.ChatOptions.Input) != 0 {
		updatedTemplate := replaceInputPlaceholders(agentInfo.PromptTemplate, &question.Source.UserInput.ChatOptions.Input)
		messages = append(messages, &llmchat.Message{
			Role:    llmchat.MessageRoleUser,
			Content: updatedTemplate,
		})
	}
	for _, v := range question.Source.UserInput.Messages {
		messages = append(messages, &llmchat.Message{
			Role:    llmchat.MessageRole(v.Role),
			Content: v.Content.Text,
		})
	}
	return messages
}

// ReplaceInputPlaceholders 替换模板中的占位符 inputN 为对应的 input.Value
func replaceInputPlaceholders(template string, inputs *chattype.InputList) string {
	// 将 inputs 转换为 map，方便快速查找
	inputMap := make(map[string]string)
	for _, input := range *inputs {
		inputMap[input.Name] = input.Value
	}

	// 正则表达式匹配所有 {{inputN}}
	re := regexp.MustCompile(`\{\{input\d+\}\}`)

	// 替换所有匹配的占位符
	template = re.ReplaceAllStringFunc(template, func(match string) string {
		// 提取占位符名称（如 input1）
		name := match[2 : len(match)-2]
		if value, ok := inputMap[name]; ok {
			return value
		}
		// 如果未找到对应的 input.Value，保留占位符
		return match
	})

	return template
}
