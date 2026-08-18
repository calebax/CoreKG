package essearch

import (
	"fmt"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/agentclient"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/logs"
)

var (
	intentionLLM = agentclient.DefaultLLMConfig{
		BaseURL:   "https://api.example.com/v3/chat.Agent/chat/completions",
		ModelName: "oSMfiUt",
	}
)

// IntentionRecognition 用户意图识别
func (w *EsSearchWrapper) IntentionRecognition() (string, error) {
	cfg, err := agentclient.GetLLMConfig(w.ctx, global.SettingGroupKnowledge, global.SettingKeyAgentIntentionRecognition)
	if err != nil {
		logs.ErrorContextf(w.ctx, "IntentionRecognition GetLLMConfig error: %v", err)
		return "", err
	}
	cli := agentclient.NewChatClient(nil, cfg.BaseURL, cfg.APIKEY)
	req := &agentclient.ChatRequestBody{
		Model: cfg.ModelName,
		ChatOptions: agentclient.ChatOptions{
			Input: []chattype.Input{
				{Name: "input1", Value: w.question},
			},
		},
		Stream: false,
	}
	resp, err := cli.SendChat(w.ctx, req)
	if err != nil {
		logs.ErrorContextf(w.ctx, "IntentionRecognition SendChat error: %v", err)
		return "", err
	}
	if len(resp.Choices) == 0 {
		logs.ErrorContextf(w.ctx, "IntentionRecognition no choices found")
		return "", fmt.Errorf("no choices found")
	}
	content := resp.Choices[0].Message.Content
	if content == "" {
		logs.ErrorContextf(w.ctx, "IntentionRecognition no content found")
		return "", fmt.Errorf("no content found")
	}
	return content, nil
}
