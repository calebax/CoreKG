package svcmodel

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const cozeModelTypeLLM int8 = 0

// SyncCozeModelInstance 将 ChatModel 同步到 coze model_instance 表
// 如果 cozeModelID 不存在则新建，存在则更新，同时回写 ChatModel 的 coze_model_id。
// 当模型不支持 function call 时，直接删除 coze 中对应 coze_model_id 的模型。
func SyncCozeModelInstance(ctx context.Context, modelID, cozeModelID uint) (uint, error) {
	chatModelDao := chat.NewChatModelDao()
	cozeModelDao := chat.NewCozeModelInstanceDao()

	modelEntity, err := chatModelDao.GetByID(ctx, modelID)
	if err != nil {
		logs.ErrorContextf(ctx, "[svcmodel] SyncCozeModelInstance failed to get chat model, modelID:%d err:%v", modelID, err)
		return 0, err
	}
	if modelEntity == nil || modelEntity.ID == 0 {
		return 0, fmt.Errorf("chat model %d not found", modelID)
	}

	if modelEntity.SupportFunctionCall == chattype.SupportFunctionCallUnsupported {
		deleteCozeID := modelEntity.CozeModelID
		if deleteCozeID == 0 {
			logs.InfoContextf(ctx, "[svcmodel] SyncCozeModelInstance skip sync for unsupported function call model, modelID:%d no bound coze model", modelEntity.ID)
			return 0, nil
		}

		err = dbutil.Coze().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txCozeDao := cozeModelDao.WithTx(tx)
			if err := txCozeDao.UpdateMap(ctx, deleteCozeID, map[string]interface{}{
				"deleted_at": time.Now().AddDate(0, 0, -30),
			}); err != nil {
				return fmt.Errorf("mark coze model %d deleted_at failed: %w", deleteCozeID, err)
			}
			return nil
		})
		if err != nil {
			logs.ErrorContextf(ctx, "[svcmodel] SyncCozeModelInstance failed to delete unsupported coze model, modelID:%d cozeModelID:%d err:%v", modelID, cozeModelID, err)
			return 0, err
		}

		logs.InfoContextf(ctx, "[svcmodel] SyncCozeModelInstance skip sync for unsupported function call model, modelID:%d deletedCozeID:%d", modelEntity.ID, deleteCozeID)
		return deleteCozeID, nil
	}

	cozeEntity := buildCozeModelInstance(modelEntity)
	cozeEntity.DeletedAt = gorm.DeletedAt{}

	var syncedCozeID uint
	err = dbutil.Coze().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCozeDao := cozeModelDao.WithTx(tx)
		if cozeModelID > 0 {
			// Always work on fresh DB sessions so a previous ErrRecordNotFound doesn't poison later writes.
			modelDB := txCozeDao.DB(ctx).Model(&chattype.CozeModelInstance{}).Unscoped()
			updateFields := []string{"type", "provider", "display_info", "connection", "capability", "parameters", "extra", "deleted_at"}

			var existing chattype.CozeModelInstance
			queryDB := modelDB.Where("id = ?", cozeModelID)
			err := queryDB.First(&existing).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("query coze model %d failed: %w", cozeModelID, err)
			}

			if errors.Is(err, gorm.ErrRecordNotFound) || existing.ID == 0 {
				cozeEntity.ID = cozeModelID
				if err := txCozeDao.DB(ctx).Model(&chattype.CozeModelInstance{}).Unscoped().Create(cozeEntity).Error; err != nil {
					return fmt.Errorf("create coze model %d failed: %w", cozeModelID, err)
				}
				syncedCozeID = cozeEntity.ID
				return nil
			}

			cozeEntity.ID = cozeModelID
			if err := txCozeDao.DB(ctx).Model(&chattype.CozeModelInstance{}).Unscoped().
				Select(updateFields).Where("id = ?", cozeModelID).Updates(cozeEntity).Error; err != nil {
				return fmt.Errorf("update coze model %d failed: %w", cozeModelID, err)
			}
			syncedCozeID = cozeModelID
			return nil
		}

		if err := txCozeDao.Insert(ctx, cozeEntity); err != nil {
			return fmt.Errorf("create coze model failed: %w", err)
		}
		syncedCozeID = cozeEntity.ID
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[svcmodel] SyncCozeModelInstance failed on coze db, modelID:%d cozeModelID:%d err:%v", modelID, cozeModelID, err)
		return 0, err
	}

	if syncedCozeID == 0 {
		return 0, fmt.Errorf("sync coze model failed, empty coze id")
	}

	if err := chatModelDao.UpdateMap(ctx, modelEntity.ID, map[string]interface{}{
		"coze_model_id": syncedCozeID,
	}); err != nil {
		logs.ErrorContextf(ctx, "[svcmodel] SyncCozeModelInstance failed to update chat model, modelID:%d cozeID:%d err:%v", modelEntity.ID, syncedCozeID, err)
		return 0, err
	}

	logs.InfoContextf(ctx, "[svcmodel] SyncCozeModelInstance success, modelID:%d cozeID:%d", modelEntity.ID, syncedCozeID)
	return syncedCozeID, nil
}

func buildCozeModelInstance(model *chattype.ChatModel) *chattype.CozeModelInstance {
	providerName := strings.TrimSpace(model.ModelProvider)
	modelName := strings.TrimSpace(model.ModelName)
	displayName := strings.TrimSpace(model.ShowName)
	if providerName == "" {
		providerName = deriveProviderFromModelName(modelName)
	}
	if displayName == "" {
		displayName = modelName
	}

	return &chattype.CozeModelInstance{
		Type: cozeModelTypeLLM,
		Provider: &chattype.ModelProvider{
			Name:       &chattype.I18nText{ZhCn: providerName, EnUs: providerName},
			IconURL:    model.HeadURL,
			ModelClass: mapProviderToClass(providerName),
		},
		DisplayInfo: &chattype.DisplayInfo{
			Name:         displayName,
			OutputTokens: 128000,
			MaxTokens:    128000,
			Description: &chattype.I18nText{
				ZhCn: model.ModelName,
				EnUs: model.ModelName,
			},
		},
		Connection: &chattype.Connection{
			BaseConnInfo: &chattype.BaseConnectionInfo{
				BaseURL:      sanitizeBaseURL(model.ModelUrl),
				APIKey:       model.APIKey,
				Model:        modelName,
				ThinkingType: chattype.ThinkingType_Default,
			},
		},
		Capability: buildCapability(model),
		Parameters: buildDefaultParameters(),
		Extra:      datatypes.JSON([]byte("{}")),
	}
}

func buildCapability(model *chattype.ChatModel) *chattype.ModelAbility {
	ability := &chattype.ModelAbility{}
	switch model.SupportFunctionCall {
	case chattype.SupportFunctionCallSupported:
		ability.FunctionCall = boolPtr(true)
	case chattype.SupportFunctionCallUnsupported:
		ability.FunctionCall = boolPtr(false)
	}
	return ability
}

type providerClassMapping struct {
	provider string
	class    chattype.ModelClass
	tokens   []string
}

var providerClassMappings = []providerClassMapping{
	{provider: "openai", class: chattype.ModelClass_GPT, tokens: []string{"openai", "gpt", "chatgpt"}},
	{provider: "claude", class: chattype.ModelClass_Claude, tokens: []string{"claude", "anthropic"}},
	{provider: "deepseek", class: chattype.ModelClass_DeekSeek, tokens: []string{"deepseek"}},
	{provider: "qwen", class: chattype.ModelClass_QWen, tokens: []string{"qwen", "aliyun", "tongyi"}},
	{provider: "ernie", class: chattype.ModelClass_Ernie, tokens: []string{"ernie", "wenxin"}},
	{provider: "baidu", class: chattype.ModelClass_QianFan, tokens: []string{"qianfan", "baidu"}},
	{provider: "gemini", class: chattype.ModelClass_Gemini, tokens: []string{"gemini", "google"}},
	{provider: "cohere", class: chattype.ModelClass_Cohere, tokens: []string{"cohere"}},
	{provider: "minimax", class: chattype.ModelClass_MiniMax, tokens: []string{"minimax"}},
	{provider: "moonshot", class: chattype.ModelClass_Moonshot, tokens: []string{"moonshot", "kimi"}},
	{provider: "glm", class: chattype.ModelClass_GLM, tokens: []string{"glm", "chatglm", "zhipu"}},
	{provider: "baichuan", class: chattype.ModelClass_Baichuan, tokens: []string{"baichuan"}},
	{provider: "llama", class: chattype.ModelClass_Llama, tokens: []string{"llama", "llama2", "llama-2", "llama-3"}},
	{provider: "stepfun", class: chattype.ModelClass_StepFun, tokens: []string{"stepfun"}},
	{provider: "maas-auto", class: chattype.ModelClass_MaaSAutoSync, tokens: []string{"maas_auto", "maas-auto", "volcengine", "volcano", "ark"}},
	{provider: "maas", class: chattype.ModelClass_Maas, tokens: []string{"maas"}},
	{provider: "seed", class: chattype.ModelClass_SEED, tokens: []string{"seed"}},
	{provider: "plugin", class: chattype.ModelClass_Plugin, tokens: []string{"plugin"}},
	{provider: "stablediffusion", class: chattype.ModelClass_StableDiffusion, tokens: []string{"stablediffusion", "stable-diffusion", "stable diffusion", "stability", "sdxl"}},
	{provider: "byteartist", class: chattype.ModelClass_ByteArtist, tokens: []string{"byteartist", "byte-artist"}},
}

func mapProviderToClass(provider string) chattype.ModelClass {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	for _, mapping := range providerClassMappings {
		for _, token := range mapping.tokens {
			if normalized == token {
				return mapping.class
			}
		}
		if normalized == mapping.provider {
			return mapping.class
		}
	}
	return chattype.ModelClass_Other
}

func deriveProviderFromModelName(modelName string) string {
	lowerName := strings.ToLower(strings.TrimSpace(modelName))

	for _, mapping := range providerClassMappings {
		for _, token := range mapping.tokens {
			if strings.Contains(lowerName, token) {
				return mapping.provider
			}
		}
	}
	return ""
}

func boolPtr(val bool) *bool {
	return &val
}

func strPtr(val string) *string {
	return &val
}

func sanitizeBaseURL(url string) string {
	u := strings.TrimSpace(url)
	u = strings.TrimSuffix(u, "/chat/completions")
	u = strings.TrimSuffix(u, "/chat/completions/")
	return strings.TrimSuffix(u, "/")
}

func buildDefaultParameters() []*chattype.ModelParameter {
	return []*chattype.ModelParameter{
		{
			Name:      "temperature",
			Label:     "生成随机性",
			Desc:      "- **temperature**: 调高温度会使得模型的输出更多样性和创新性，反之，降低温度会使输出内容更加遵循指令要求但减少多样性。建议不要与“Top p”同时调整。",
			Type:      chattype.ModelParamType_Float,
			Min:       "0",
			Max:       "1",
			Precision: 1,
			DefaultVal: &chattype.ModelParamDefaultValue{
				DefaultVal: "1.0",
				Creative:   strPtr("1"),
				Balance:    strPtr("0.8"),
				Precise:    strPtr("0.3"),
			},
			Options: []*chattype.Option{},
			ParamClass: &chattype.ModelParamClass{
				ClassID: 1,
				Label:   "生成随机性",
			},
		},
		{
			Name:      "max_tokens",
			Label:     "最大回复长度",
			Desc:      "控制模型输出的Tokens 长度上限。通常 100 Tokens 约等于 150 个中文汉字。",
			Type:      chattype.ModelParamType_Int,
			Min:       "1",
			Max:       "4096",
			Precision: 0,
			DefaultVal: &chattype.ModelParamDefaultValue{
				DefaultVal: "4096",
			},
			Options: []*chattype.Option{},
			ParamClass: &chattype.ModelParamClass{
				ClassID: 2,
				Label:   "输入及输出设置",
			},
		},
	}
}

// DeleteModel 删除 chat_model 以及对应的 coze model_instance
func DeleteModel(ctx context.Context, modelID uint) error {
	chatModelDao := chat.NewChatModelDao()
	cozeModelDao := chat.NewCozeModelInstanceDao()

	modelEntity, err := chatModelDao.GetByID(ctx, modelID)
	if err != nil {
		logs.ErrorContextf(ctx, "[svcmodel] DeleteModel failed to get chat model, modelID:%d err:%v", modelID, err)
		return err
	}
	if modelEntity == nil || modelEntity.ID == 0 {
		return fmt.Errorf("chat model %d not found", modelID)
	}

	if modelEntity.CozeModelID > 0 {
		logs.InfoContextf(ctx, "[svcmodel] DeleteModel delete coze model, modelID:%d cozeID:%d", modelID, modelEntity.CozeModelID)
		if err := dbutil.Coze().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return cozeModelDao.WithTx(tx).Delete(ctx, modelEntity.CozeModelID)
		}); err != nil {
			logs.ErrorContextf(ctx, "[svcmodel] DeleteModel delete coze model failed, modelID:%d cozeID:%d err:%v", modelID, modelEntity.CozeModelID, err)
			return err
		}
	}

	if err := chatModelDao.Delete(ctx, modelID); err != nil {
		logs.ErrorContextf(ctx, "[svcmodel] DeleteModel delete chat model failed, modelID:%d err:%v", modelID, err)
		return err
	}

	logs.InfoContextf(ctx, "[svcmodel] DeleteModel success, modelID:%d cozeID:%d", modelID, modelEntity.CozeModelID)
	return nil
}

// BindCozeModel 删除现有关联的 coze model，然后按传入的 cozeModelID 创建或更新新的关联
func BindCozeModel(ctx context.Context, modelID, cozeModelID uint) (uint, error) {
	chatModelDao := chat.NewChatModelDao()
	cozeModelDao := chat.NewCozeModelInstanceDao()

	modelEntity, err := chatModelDao.GetByID(ctx, modelID)
	if err != nil {
		logs.ErrorContextf(ctx, "[svcmodel] BindCozeModel failed to get chat model, modelID:%d err:%v", modelID, err)
		return 0, err
	}
	if modelEntity == nil || modelEntity.ID == 0 {
		return 0, fmt.Errorf("chat model %d not found", modelID)
	}

	targetCozeID := cozeModelID
	if modelEntity.CozeModelID > 0 && modelEntity.CozeModelID != cozeModelID {
		logs.InfoContextf(ctx, "[svcmodel] BindCozeModel delete old coze model, modelID:%d oldCozeID:%d", modelID, modelEntity.CozeModelID)
		if err := dbutil.Coze().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return cozeModelDao.WithTx(tx).Delete(ctx, modelEntity.CozeModelID)
		}); err != nil {
			logs.ErrorContextf(ctx, "[svcmodel] BindCozeModel delete old coze model failed, modelID:%d cozeID:%d err:%v", modelID, modelEntity.CozeModelID, err)
			return 0, err
		}
	}

	newCozeID, err := SyncCozeModelInstance(ctx, modelID, targetCozeID)
	if err != nil {
		logs.ErrorContextf(ctx, "[svcmodel] BindCozeModel sync failed, modelID:%d cozeID:%d err:%v", modelID, targetCozeID, err)
		return 0, err
	}

	logs.InfoContextf(ctx, "[svcmodel] BindCozeModel success, modelID:%d newCozeID:%d", modelID, newCozeID)
	return newCozeID, nil
}
