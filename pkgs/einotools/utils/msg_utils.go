package utils

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/ygpkg/yg-go/logs"
)

var ignoredMsgTypes = map[string]struct{}{
	"user":               {},
	"code_executor_tool": {},
}

var convertToolNameMap = map[string]string{
	"code_generator_tool": models.MsgTypeCodeExec,
	"code_processor_tool": models.MsgTypeCodeExec,
	"file_read_tool":      models.MsgTypeFileView,
	"file_inspect_tool":   models.MsgTypeFileView,
}

func ConvertMsg2WriteResult(ctx context.Context, agentResponse []*models.Message) []*models.WriteResult {
	var res []*models.WriteResult
	for _, msg := range agentResponse {

		if msg.MessageType == models.MsgTypeResult {
			res = append(res, &models.WriteResult{
				Content: msg.Payload.Content,
				Flag:    models.FlagFinalResult,
			})
			continue
		}

		response := models.AgentResponse{
			MessageID:   msg.MessageId,
			MessageTime: msg.MessageTime,
			MessageType: msg.MessageType,
			IsFinal:     true,
		}

		msgPayload := msg.Payload

		resultMap := ConvertMsgResultMap(ctx, &response, msg.MessageType, msgPayload.Content)
		if resultMap == nil {
			continue
		}

		response.ResultMap = resultMap

		res = append(res, &models.WriteResult{
			Content: response,
			Flag:    models.FlagAgent,
		})
	}
	return res
}

func ConvertMsgResultMap(ctx context.Context, agentResponse *models.AgentResponse, msgType string, content any) map[string]any {
	if _, ok := ignoredMsgTypes[msgType]; ok {
		return nil
	}

	if msgType == models.MsgTypeTaskThought {
		agentResponse.TaskThought = content.(string)
		return map[string]any{}
	}

	convertMsgTypeName, ok := convertToolNameMap[msgType]
	if ok {
		agentResponse.MessageType = convertMsgTypeName
	}

	var chartTypeRegex = regexp.MustCompile(`^create_.*_chart_option$`)
	if chartTypeRegex.MatchString(msgType) {
		agentResponse.MessageType = models.MsgTypeChartGenerate
	}

	var result map[string]any = map[string]any{}

	if tr, ok := content.(*models.ToolResponse); ok {
		// tool 调用前置消息推送，可处理回显内容，比如正在检索关键词 或 调用参数回显
		result["shell"] = tr.ToolShell
		return result
	}

	if chartTypeRegex.MatchString(msgType) {
		var contentMap map[string]interface{}
		if err := json.Unmarshal([]byte(content.(string)), &contentMap); err != nil {
			logs.ErrorContextf(ctx, "Error unmarshaling content: %v", err)
			return nil
		} else {
			result = contentMap
		}

		return result
	}

	switch msgType {
	default:
		resultContent := content
		if str, ok := content.(string); ok {
			err := json.Unmarshal([]byte(content.(string)), &result)
			if err == nil {
				return result
			}
			resultContent = &models.ToolResponse{ToolResult: str}
		}
		data, err := json.Marshal(resultContent)
		if err != nil {
			return nil
		} else if err := json.Unmarshal(data, &result); err != nil {
			return nil
		}
	}

	return result
}

func ConvertToToolShell(toolName *string, argumentsInJSON *string) string {
	return *argumentsInJSON
}
