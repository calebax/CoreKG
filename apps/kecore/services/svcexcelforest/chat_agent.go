package svcexcelforest

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatclient"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// RequestIsFirstColumnRowTitleAgent 发送数据给agent获取是否第一列为行标题
func RequestIsFirstColumnRowTitleAgent(ctx *gin.Context, data string) (bool, error) {
	model := chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentExcelIsFirstColumnRowTitle)
	req := &chattype.ChatRequestBody{
		Stream: false,
		Model:  model,
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{Name: "input1", Value: data},
			},
		},
	}
	w, err := chatclient.NewInternalChat(ctx, runtime.RequestID(ctx), "", 1, req)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to create internal chat: %v", err)
		return false, err
	}
	res, err := w.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(ctx, "agent chat fail, name: %s, error: %v", model, err)
		return false, err
	}
	logs.InfoContextf(ctx, "original Agent response: %s", res)
	return res.Content == "true", nil
}

// RequestHeaderRowAgent 发送数据给agent获取正确的表头行以及英文列名数组
// 返回: (列名数组, 表头行号, 错误)
func RequestHeaderRowAgent(ctx *gin.Context, data string) ([]string, int, []string, error) {
	model := chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentExcelHeaderRow)
	req := &chattype.ChatRequestBody{
		Stream: false,
		Model:  model,
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{Name: "input1", Value: data},
			},
		},
	}
	w, err := chatclient.NewInternalChat(ctx, runtime.RequestID(ctx), "", 1, req)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to create internal chat: %v", err)
		return nil, 0, nil, err
	}
	res, err := w.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(ctx, "agent chat fail, name: %s, error: %v", model, err)
		return nil, 0, nil, err
	}

	// 清理Markdown代码块标记
	cleanedContent := strings.TrimSpace(res.Content)
	cleanedContent = strings.TrimPrefix(cleanedContent, "```json")
	cleanedContent = strings.TrimPrefix(cleanedContent, "```")
	cleanedContent = strings.TrimSuffix(cleanedContent, "```")
	cleanedContent = strings.TrimSpace(cleanedContent)

	logs.InfoContextf(ctx, "cleaned response: %s", cleanedContent)

	var headerAgentResp HeaderAgentResp
	if err := json.Unmarshal([]byte(cleanedContent), &headerAgentResp); err != nil {
		logs.ErrorContextf(ctx, "unmarshal Agent data failed: %v, content: %s", err, cleanedContent)
		return nil, 0, nil, fmt.Errorf("unmarshal Agent data failed: %v", err)
	}

	// 验证返回数据
	if len(headerAgentResp.ColumnNames) == 0 {
		return nil, 0, nil, fmt.Errorf("agent did not return column names")
	}

	if headerAgentResp.HeaderRow <= 0 {
		logs.WarnContextf(ctx, "agent return unvalid headerRow: %d", headerAgentResp.HeaderRow)
	}
	return headerAgentResp.ColumnNames, headerAgentResp.HeaderRow, headerAgentResp.ColumnRealNames, nil
}

// HeaderAgentResp 解析实际数据
type HeaderAgentResp struct {
	// 转换后的列名数组
	ColumnNames []string `json:"columnNames"`
	// 表头行
	HeaderRow int `json:"headerRow"`
	// 转换前的列名数组
	ColumnRealNames []string `json:"columnRealNames"`
}

// RequestGetHeaderNumbers 发送数据给agent获取Excel表头行号列表
func RequestGetHeaderNumbers(ctx *gin.Context, rows []string) ([]int, error) {
	inputBytes, _ := json.Marshal(rows)
	model := chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentExcelHeaderRowNumberListAnalysis)
	req := &chattype.ChatRequestBody{
		Stream: false,
		Model:  model,
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{Name: "input1", Value: string(inputBytes)},
			},
		},
	}
	w, err := chatclient.NewInternalChat(ctx, runtime.RequestID(ctx), "", 1, req)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to create internal chat: %v", err)
		return nil, err
	}
	res, err := w.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(ctx, "agent chat fail, name: %s, error: %v", model, err)
		return nil, err
	}
	var headerNumbers []int
	if err := json.Unmarshal([]byte(res.Content), &headerNumbers); err != nil {
		logs.ErrorContextf(ctx, "unmarshal Agent data failed: %v, content: %s", err, res.Content)
		return nil, nil
	}
	return headerNumbers, nil
}
