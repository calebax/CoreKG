package devkeywords

import (
	"context"
	"testing"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

func TestReplace(t *testing.T) {
	text := "你的后倒车雷达\n的维修步骤说明"
	// keywords := []string{"后倒车雷达", "维修步骤"}
	replacements := map[string]string{
		"维修步骤":  "维修流程",
		"后倒车雷达": "倒车雷达",
	}
	result, err := replaceByKeywords(text, replacements)
	if err != nil {
		t.Errorf("replaceByKeywords fail, err: %v", err)
	}
	println(result, err == nil)
}

func TestAA(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := context.Background()
	sysAPIKey, err := settings.GetText(global.SettingGroupKnowledge, global.SettingKeySystemLlmAPIKey)
	if err != nil {
		logs.ErrorContextf(ctx, "get system api key from setting fail, err: %w", err)
		return
	}
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  sysAPIKey,
		Model:   "deepseek/deepseek-v3",
		BaseURL: global.LLMBaseURL,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[newToolCallingChatModel] failed to create OpenAiChatModel: %v", err)
		return
	}
	println(ReplaceMajorKeywords(context.Background(), chatModel, 2, "熊出没和奥特曼在干什么"))
}
