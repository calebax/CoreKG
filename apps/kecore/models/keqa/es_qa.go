package keqa

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/agentclient"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
)

var (
	defalutLLM = agentclient.DefaultLLMConfig{
		BaseURL:   "https://tapi.example.com/v3/chat.Agent/chat/completions",
		ModelName: "3B2DqQT",
	}
	llmImageParseCfg = agentclient.DefaultLLMConfig{
		BaseURL:   "https://api.example.com/v3/llm.chat/chat/completions",
		ModelName: "qwen2.5-vl-72b-instruct",
	}
)

func DefaultChat(wrapper *searchReferenceWrapper, sseClient *sseclient.SSEClient, qs *foresttype.KnownowForestQA, history string, session *foresttype.KnownowQASession) (string, string, error) {
	preSearchResult, files, err := wrapper.PreSearchQuestionChunk()
	if err != nil {
		return "", "", err
	}
	WriteReferenceFile(wrapper.ctx, sseClient, files, qs.ID)

	refList, err := wrapper.SupSearchQuestionChunk(preSearchResult)
	if err != nil {
		return "", "", err
	}
	logs.InfoContextf(wrapper.ctx, "[ForestChat] SupSearchQuestionChunk result: %v", len(refList))
	qs.QueryReferenceList = refList
	searchStr, err := TransformChatReferenceList(wrapper.ctx, refList)
	if err != nil {
		return "", "", err
	}

	answer, reason, err := ESChat(wrapper.ctx, sseClient, string(searchStr), history, qs, session)
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "ESChat error: %v", err)
	}
	return answer, reason, err
}

// ESChat 根据es查到的chunk来进行问答
func ESChat(ctx *gin.Context, sseClient *sseclient.SSEClient, searchstr, history string, qs *foresttype.KnownowForestQA, session *foresttype.KnownowQASession) (string, string, error) {
	cfg, err := agentclient.GetLLMConfig(ctx, global.SettingGroupKnowledge, global.SettingKeyAgentEsChat)
	if err != nil {
		logs.ErrorContextf(ctx, "get llm config failed: %v", err)
		return "", "", err
	}
	client := agentclient.NewChatClientWithConfig(nil, cfg)
	logs.DebugContextf(ctx, "用户问题:\n%v\n", qs.Question)
	//logs.DebugContextf(ctx, "知识库检索内容\n%v\n", searchstr)
	logs.DebugContextf(ctx, "对话历史\n%v\n", history)
	req := &agentclient.ChatRequestBody{
		//replace with r1(aliyun) model
		Model:      cfg.ModelName,
		LLMModelID: session.LLMModelID,
		ChatOptions: agentclient.ChatOptions{
			Input: []chattype.Input{
				{Name: "input1", Value: qs.Question}, // 用户问题
				{Name: "input2", Value: searchstr},   // 知识库检索内容
				{Name: "input3", Value: history},     // 对话历史
			},
		},
	}
	var (
		contentBuilder   strings.Builder
		reasoningBuilder strings.Builder
	)
	err = client.SendChatStreamWithCallback(ctx, req, func(chunk *agentclient.ChatStreamResponseBody) error {
		for _, v := range chunk.Choices {
			contentBuilder.WriteString(v.Delta.Content)
			reasoningBuilder.WriteString(v.Delta.ReasoningContent)
			writeResult := WriteResult{
				ReasoningContent: v.Delta.ReasoningContent,
				Content:          v.Delta.Content,
			}
			if stoped, err := sseClient.WriteMessage(ctx, ctx.Writer, fmt.Sprintf("%v", qs.ID), writeResult.String()); err != nil {
				if strings.Contains(err.Error(), "broken pipe") {
					continue
				}
				logs.ErrorContextf(ctx, "[forestqa] Failed to write Answering response to KEQA: %v", err)
				return err
			} else if stoped {
				logs.InfoContextf(ctx, "[forestqa] stream Stoped by KEQA")
				return nil
			}
			print(v.Delta.ReasoningContent)
			print(v.Delta.Content)
		}
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "SendChatStreamWithCallback error: %v", err)
		return contentBuilder.String(), reasoningBuilder.String(), err
	}
	return contentBuilder.String(), reasoningBuilder.String(), nil
}

// WriteReferenceFile 写入检索到的文件
func WriteReferenceFile(ctx *gin.Context, sseClient *sseclient.SSEClient, files []*foresttype.KnownowForestFile, question_id uint) {
	for _, file := range files {
		if stoped, err := sseClient.WriteMessage(ctx, ctx.Writer, fmt.Sprintf("%v", question_id), WriteResult{
			Reference: Reference{
				ForestID: file.ForestID,
				FileID:   file.ID,
				Name:     file.Name,
			},
			Flag: FlagFound,
		}.String()); err != nil {
			logs.ErrorContextf(ctx, "[forestqa] Failed to write Answering response to KEQA: %v", err)
			continue
		} else if stoped {
			logs.ErrorContextf(ctx, "[forestqa] stream Stoped by KEQA")
			return
		}
	}
}

// SumDescAgent 根据es查到的desc总结
func SumDescAgent(ctx context.Context, cfg config.LLMModelConfig, searchstr, history, question string) (string, error) {

	client := agentclient.NewChatClientWithConfig(nil, cfg)
	req := &agentclient.ChatRequestBody{
		Model: cfg.ModelName,
		ChatOptions: agentclient.ChatOptions{
			Input: []chattype.Input{
				{Name: "input1", Value: question},  // 用户问题
				{Name: "input2", Value: searchstr}, // 知识库检索内容
				{Name: "input3", Value: history},   // 对话历史
			},
		},
	}
	resp, err := client.SendChat(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "IntentionRecognition SendChat error: %v", err)
		return "", err
	}
	if len(resp.Choices) == 0 {
		logs.ErrorContextf(ctx, "IntentionRecognition no choices found")
		return "", fmt.Errorf("no choices found")
	}
	content := resp.Choices[0].Message.Content
	if content == "" {
		logs.ErrorContextf(ctx, "IntentionRecognition no content found")
		return "", fmt.Errorf("no content found")
	}
	return content, nil
}
