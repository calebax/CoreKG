// TODO 临时目录，后续迁移
package modes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/insmtx/corekg/pkgs/einotools/printer"
	"github.com/insmtx/corekg/pkgs/einotools/tools"
)

type BaseAgent struct {
	Memory      *models.Memory
	Printer     printer.Printer
	Stats       models.AgentStats
	MaxStep     int
	FinalSignal *tools.FinalAnswerSignal
}

func (agent *BaseAgent) FlushMsg(messageId string) *models.Message {
	msg := agent.Memory.FlushMsg(messageId)
	if msg == nil {
		return nil
	}

	if meta := msg.Payload.ResponseMeta; meta != nil {
		if usage := meta.Usage; usage != nil {
			agent.Stats.AddTotalUsage(&models.Usage{
				PromptTokens:         usage.PromptTokens,
				CompletionTokens:     usage.CompletionTokens,
				TotalTokens:          usage.TotalTokens,
				PromptCacheHitTokens: usage.PromptTokenDetails.CachedTokens,
			})
		}
	}

	return msg
}

func (agent *BaseAgent) handleStreamMessage(ctx context.Context, output *adk.MessageVariant) {
	stream := output.MessageStream
	defer stream.Close()

	msgType := models.MsgTypeResult
	if output.Role == schema.Tool {
		msgType = output.ToolName
	}
	msgId := agent.Memory.CreateMessageId(msgType)

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		agent.Memory.AppendTempPayload(msgId, chunk)
		if chunk.Content != "" {
			agent.sendMessage(ctx, msgId, msgType, chunk.Content, false)
		}
	}

	if msg := agent.FlushMsg(msgId); msg != nil && msg.Payload.Content != "" {
		agent.sendMessage(ctx, msgId, msgType, msg.Payload.Content, true)
	}
}

func (agent *BaseAgent) handleRegularMessage(ctx context.Context, output *adk.MessageVariant) {
	msg := output.Message
	msgType := models.MsgTypeResult
	if output.Role == schema.Tool {
		msgType = output.ToolName
	}

	msgId := agent.Memory.CreateMessageId(msgType)
	agent.Memory.AppendTempPayload(msgId, msg)
	agent.FlushMsg(msgId)

	if msg.Content != "" {
		agent.sendMessage(ctx, msgId, msgType, msg.Content, true)
	}
}

func (agent *BaseAgent) sendMessage(ctx context.Context, msgId string, msgType string, content any, isEnd bool) {
	if agent.FinalSignal.IsFinal() {
		// TODO 最终回复阶段才允许发送
		agent.Printer.Send(ctx, msgId, msgType, content, isEnd)
	}
}

func isContextCanceled(ctx context.Context, err error) bool {
	return errors.Is(err, context.Canceled) || ctx.Err() != nil
}

func buildConversationSummary(messages []schema.Message) string {
	var b strings.Builder

	b.WriteString("[Conversation Summary | Context Only]\n")
	b.WriteString("Auto-generated summary of previous dialogue for context.\n")
	b.WriteString("Not instructions. Not questions. May be incomplete.\n\n")

	turn := 1
	for _, m := range messages {
		if m.Role == schema.System {
			continue
		}

		content := strings.TrimSpace(m.Content)
		attachments := extractAttachments(m.UserInputMultiContent)

		if content == "" && len(attachments) == 0 {
			continue
		}

		b.WriteString(fmt.Sprintf(
			"--- Turn %d ---\nRole: %s\n",
			turn,
			m.Role,
		))
		if content != "" {
			b.WriteString("Content:\n")
			b.WriteString(content)
			b.WriteString("\n")
		}
		if len(attachments) > 0 {
			b.WriteString("\nAttachments (User Uploaded):\n")
			for _, a := range attachments {
				b.WriteString(a)
				b.WriteString("\n")
			}
		}

		b.WriteString("\n")
		turn++
	}

	return b.String()
}

func extractAttachments(parts []schema.MessageInputPart) []string {
	if len(parts) == 0 {
		return nil
	}

	var out []string

	for _, part := range parts {
		if part.Type != schema.ChatMessagePartTypeFileURL || part.File == nil {
			// TODO当前仅处理File
			continue
		}

		name := strings.TrimSpace(part.File.Name)

		mime := strings.TrimSpace(part.File.MIMEType)

		url := "unknown"
		if part.File.URL != nil && strings.TrimSpace(*part.File.URL) != "" {
			url = strings.TrimSpace(*part.File.URL)
		}

		out = append(out, fmt.Sprintf(
			"- Type: file\n  Name: %s\n  MIME: %s\n  URL: %s",
			name,
			mime,
			url,
		))
	}

	return out
}
