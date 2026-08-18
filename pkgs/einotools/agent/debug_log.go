package agent

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"unicode/utf8"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/ygpkg/yg-go/logs"
)

const debugContentChunkRuneSize = 1800

func marshalDebugJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "marshal_error:" + err.Error()
	}
	return string(b)
}

func logReActToolInfos(ctx context.Context, tools []tool.BaseTool) {
	if len(tools) == 0 {
		logs.InfoContextf(ctx, "[DEBUG][tool-question-source] ReActAgent available tools: count=0")
		return
	}

	type debugToolInfo struct {
		Name       string `json:"name"`
		Desc       string `json:"desc"`
		Parameters any    `json:"parameters,omitempty"`
	}

	infos := make([]debugToolInfo, 0, len(tools))
	for _, tl := range tools {
		info, err := tl.Info(ctx)
		if err != nil {
			logs.WarnContextf(ctx, "[DEBUG][tool-question-source] ReActAgent tool info error: %v", err)
			continue
		}
		var parameters any
		if info.ParamsOneOf != nil {
			js, err := info.ParamsOneOf.ToJSONSchema()
			if err != nil {
				logs.WarnContextf(ctx, "[DEBUG][tool-question-source] ReActAgent tool params schema error: tool=%s, err=%v", info.Name, err)
			} else {
				parameters = js
			}
		}
		infos = append(infos, debugToolInfo{
			Name:       info.Name,
			Desc:       info.Desc,
			Parameters: parameters,
		})
	}

	logs.InfoContextf(ctx, "[DEBUG][tool-question-source] ReActAgent available tools: count=%d, tools=%s",
		len(infos), marshalDebugJSON(infos))
}

func logModelInputMessages(ctx context.Context, round int64, messages []*schema.Message, startIndex int) {
	if startIndex < 0 || startIndex > len(messages) {
		startIndex = 0
	}
	loggedCount := len(messages) - startIndex
	logs.InfoContextf(ctx, "[DEBUG][tool-question-source] ReActAgent model input begin: round=%d, message_count=%d, logged_from=%d, logged_count=%d, skipped_unchanged=%d",
		round, len(messages), startIndex, loggedCount, startIndex)

	for msgIndex := startIndex; msgIndex < len(messages); msgIndex++ {
		msg := messages[msgIndex]
		if msg == nil {
			logs.InfoContextf(ctx, "[DEBUG][tool-question-source] ReActAgent model input message nil: round=%d, message_index=%d",
				round, msgIndex)
			continue
		}

		logs.InfoContextf(ctx, "[DEBUG][tool-question-source] ReActAgent model input message meta: round=%d, message_index=%d, role=%s, name=%s, content_runes=%d, tool_call_id=%s, tool_name=%s",
			round, msgIndex, msg.Role, msg.Name, utf8.RuneCountInString(msg.Content), msg.ToolCallID, msg.ToolName)
		if len(msg.ToolCalls) > 0 {
			logs.InfoContextf(ctx, "[DEBUG][tool-question-source] ReActAgent model input message tool_calls: round=%d, message_index=%d, tool_calls=%s",
				round, msgIndex, marshalDebugJSON(msg.ToolCalls))
		}
		if msg.ResponseMeta != nil {
			logs.InfoContextf(ctx, "[DEBUG][tool-question-source] ReActAgent model input message response_meta: round=%d, message_index=%d, response_meta=%s",
				round, msgIndex, marshalDebugJSON(msg.ResponseMeta))
		}
		logContentChunks(ctx, "ReActAgent model input message content", round, msgIndex, msg.Content)
	}

	logs.InfoContextf(ctx, "[DEBUG][tool-question-source] ReActAgent model input end: round=%d, message_count=%d, logged_from=%d, logged_count=%d",
		round, len(messages), startIndex, loggedCount)
}

func buildModelInputMessageSignatures(messages []*schema.Message) []string {
	signatures := make([]string, 0, len(messages))
	for _, msg := range messages {
		signatures = append(signatures, modelInputMessageSignature(msg))
	}
	return signatures
}

func modelInputMessageSignature(msg *schema.Message) string {
	if msg == nil {
		return "<nil>"
	}
	contentHash := sha256.Sum256([]byte(msg.Content))
	return fmt.Sprintf("role=%s|name=%s|tool_call_id=%s|tool_name=%s|content_runes=%d|content_sha256=%x|tool_calls=%s|response_meta=%s",
		msg.Role,
		msg.Name,
		msg.ToolCallID,
		msg.ToolName,
		utf8.RuneCountInString(msg.Content),
		contentHash[:],
		marshalDebugJSON(msg.ToolCalls),
		marshalDebugJSON(msg.ResponseMeta),
	)
}

func logContentChunks(ctx context.Context, label string, round int64, msgIndex int, content string) {
	if content == "" {
		logs.InfoContextf(ctx, "[DEBUG][tool-question-source] %s empty: round=%d, message_index=%d",
			label, round, msgIndex)
		return
	}

	runes := []rune(content)
	chunkTotal := int(math.Ceil(float64(len(runes)) / float64(debugContentChunkRuneSize)))
	for chunkIndex := 0; chunkIndex < chunkTotal; chunkIndex++ {
		start := chunkIndex * debugContentChunkRuneSize
		end := start + debugContentChunkRuneSize
		if end > len(runes) {
			end = len(runes)
		}
		logs.InfoContextf(ctx, "[DEBUG][tool-question-source] %s chunk: round=%d, message_index=%d, chunk_index=%d, chunk_total=%d, content=%s",
			label, round, msgIndex, chunkIndex+1, chunkTotal, string(runes[start:end]))
	}
}

func logModelToolCalls(ctx context.Context, source string, msg *schema.Message) {
	if msg == nil || len(msg.ToolCalls) == 0 {
		return
	}
	logs.InfoContextf(ctx, "[DEBUG][tool-question-source] ReActAgent model output tool_calls: source=%s, content=%s, tool_calls=%s",
		source, msg.Content, marshalDebugJSON(msg.ToolCalls))
}
