package chatclient

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/apikey"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
	"github.com/ygpkg/yg-go/logs"
)

type LightApp struct {
	reqid  string
	apiKey string
	step   int
	req    *chattype.ChatRequestBody
}

func NewLightApp(reqid, apiKey string, step int, req *chattype.ChatRequestBody) *LightApp {
	return &LightApp{
		reqid:  reqid,
		apiKey: apiKey,
		step:   step,
		req:    req,
	}
}

type LightAppMetadata struct {
	LightAppVersion *chatagent.AgentWithVersion
	ChatModelEntity *chattype.ChatModel
	ApiKeyEntity    *accounttype.APIKey
}

func (c *LightApp) InvokeChat(ctx context.Context) (*llmchat.QaRes, error) {
	lightAppMetadata, err := loadMetadata(ctx, c.req.Model, c.apiKey, c.req.LLMModelID)
	if err != nil {
		logs.ErrorContextf(ctx, "[lightApp.InvokeChat] failed to load metadata, err: %v", err)
		return nil, err
	}
	lightAppInfo := lightAppMetadata.LightAppVersion
	model := lightAppMetadata.ChatModelEntity
	apiKeyInfo := lightAppMetadata.ApiKeyEntity
	questionEntity := &chattype.ChatQuestion{
		Source: &chattype.Question{
			CompanyID:    apiKeyInfo.CompanyID,
			Uin:          lightAppInfo.Uin,
			AgentStep:    c.step,
			ReqID:        c.reqid,
			Status:       chattype.QuestionStatusPending,
			ApiKeyID:     apiKeyInfo.ID,
			BaseAgentID:  lightAppInfo.ID,
			AgentVersion: lightAppInfo.Version,
			ModelID:      model.ID,
			UserInput:    c.req,
			AgentName:    lightAppInfo.Name,
		},
	}
	inputMessages := getAgentMessages(lightAppMetadata.LightAppVersion, questionEntity)

	llmReq := &llmchat.ChatReqBody{
		Messages:    inputMessages,
		Stream:      questionEntity.Source.UserInput.Stream,
		Temperature: &lightAppInfo.Temperature,
	}
	if questionEntity.Source.UserInput.Stream {
		llmReq.StreamOptions = llmchat.NewStreamOptions()
	}

	wrapper := llmchat.NewLLmChatWrapperV2(ctx, llmReq, model)
	res, err := wrapper.InvokeChat()

	if res != nil {
		questionEntity.Source.Answer = res.Content
		questionEntity.Source.Reasoning = res.Reasoning
		questionEntity.Source.ReasoningSeconds = res.ReasoningTime
		questionEntity.Source.CostSeconds = res.CostSeconds
		questionEntity.Source.OutToken = res.Usage.CompletionTokens
		questionEntity.Source.CacheHitToken = res.Usage.PromptCacheHitTokens
		questionEntity.Source.CacheMissToken = res.Usage.PromptCacheMissTokens
		questionEntity.Source.TotalTokens = res.Usage.TotalTokens
		questionEntity.Source.Status = chattype.QuestionStatusAnswered
	}
	if err != nil {
		logs.ErrorContextf(ctx, "[lightApp.InvokeChat] AgentAPIChat failed ,err %s", err)
		// return
		questionEntity.Source.Status = chattype.QuestionStatusError
	}

	err = chatquestion.CreateQuestion(ctx, questionEntity)
	if err != nil {
		logs.ErrorContextf(ctx, "[lightApp.InvokeChat] CreateQuestion err: %v", err)
		return nil, err
	}
	return res, err

}

func loadMetadata(ctx context.Context, lightAppName, apiKey string, chatModelID uint) (*LightAppMetadata, error) {
	lightAppInfo, err := chatagent.GetAgentDetailByName(ctx, lightAppName)
	if err != nil {
		logs.ErrorContextf(ctx, "[loadMetadata] failed to get light app info, light app name: %s, err: %v", lightAppName, err)
		return nil, err
	}
	var modelID uint
	if chatModelID != 0 {
		modelID = chatModelID
	} else {
		modelIDs := lightAppInfo.ChatModelIDs.Slice()
		if len(modelIDs) == 0 {
			logs.ErrorContextf(ctx, "[loadMetadata] light app has no model ,light app id: %d", lightAppInfo.ID)
			return nil, fmt.Errorf("agent has no model")
		}
		modelID = modelIDs[0]
	}
	// 获取模型信息
	chatModelEntity, err := chatmodel.GetModelByID(ctx, modelID)
	if err != nil {
		logs.ErrorContextf(ctx, "[loadMetadata] GetModelByID failed, chat model id: %d, err %v", modelID, err)
		return nil, err
	}
	apiKeyEntity, err := apikey.GetAPIKeyInfo(ctx, apiKey)
	if err != nil {
		logs.ErrorContextf(ctx, "[loadMetadata] GetAPIKeyInfo failed, api key: %s, err %v", apiKey, err)
		return nil, err
	}

	metadata := &LightAppMetadata{
		LightAppVersion: lightAppInfo,
		ChatModelEntity: chatModelEntity,
		ApiKeyEntity:    apiKeyEntity,
	}
	return metadata, nil
}
