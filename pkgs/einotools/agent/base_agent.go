package agent

import (
	"context"
	"io"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	ub "github.com/cloudwego/eino/utils/callbacks"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/insmtx/corekg/pkgs/einotools/printer"
	"github.com/insmtx/corekg/pkgs/einotools/utils"
	"github.com/ygpkg/yg-go/logs"
)

type AgentState int

const (
	IDLE AgentState = iota
	RUNNING
	FINISHED
	ERROR
)

type BaseAgent struct {
	Name         string
	AgentContext *AgentContext

	State          AgentState
	ChatModel      model.ToolCallingChatModel
	SystemPrompt   string
	NextStepPrompt string
	MaxStep        int
	Memory         *models.Memory
	Printer        printer.Printer

	Stats models.AgentStats

	// runnable compose.Runnable[[]*schema.Message, *schema.Message]
	// graph    *compose.Graph[[]*schema.Message, *schema.Message]
}

func (r *BaseAgent) Run(ctx context.Context, query string) (string, error) {
	return "", nil
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

func (agent *BaseAgent) newModelHandler(msgType string) *ub.ModelCallbackHandler {
	return &ub.ModelCallbackHandler{
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *model.CallbackOutput) context.Context {
			msg := output.Message
			msgId := agent.Memory.CreateMessageId(msgType)
			logModelToolCalls(ctx, "on_end", msg)

			// agent.tempMsgs.MessageType = msgType
			// agent.tempMsgs.Payload = append(agent.tempMsgs.Payload, msg)
			// agent.Flush()
			if msg.Content != "" {
				agent.AgentContext.Printer.Send(ctx, msgId, msgType, msg.Content, true)
			}

			agent.Memory.AppendTempPayload(msgId, msg)
			agent.FlushMsg(msgId)
			return ctx
		},
		OnEndWithStreamOutput: func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
			msgId := agent.Memory.CreateMessageId(msgType)

			handleStreamOutput(ctx, output,
				func(chunk *model.CallbackOutput) {
					msg := chunk.Message
					if msg == nil {
						return
					}
					logModelToolCalls(ctx, "stream_chunk", msg)
					agent.Memory.AppendTempPayload(msgId, msg)
					if msg.Content != "" {
						agent.AgentContext.Printer.Send(ctx, msgId, msgType, msg.Content, false)
					}
				},
				func(full string) {
					concatedMsg := agent.Memory.FlushMsg(msgId)
					if concatedMsg == nil {
						return
					}
					logModelToolCalls(ctx, "stream_final", concatedMsg.Payload)
					if concatedMsg.Payload.Content != "" {
						agent.AgentContext.Printer.Send(ctx, msgId, msgType, concatedMsg.Payload.Content, true)
					}
				},
			)
			return ctx
		},
	}
}

func (agent *BaseAgent) newToolHandler() *ub.ToolCallbackHandler {
	return &ub.ToolCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
			toolName := info.Name
			toolCallID := compose.GetToolCallID(ctx)

			msgId := agent.Memory.CreateMessageId(toolName)
			shellStr := utils.ConvertToToolShell(&toolName, &input.ArgumentsInJSON)
			agent.Memory.CallbackExecInfo[toolCallID] = &models.CallbackExecution{
				MessageId:   msgId,
				MessageType: toolName,
				Name:        shellStr,
			}

			agent.AgentContext.Printer.Send(ctx, msgId, toolName, &models.ToolResponse{ToolShell: shellStr}, false)
			return ctx
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
			toolName := info.Name
			toolCallID := compose.GetToolCallID(ctx)
			msgId := getMessageIdByCallToolId(agent, toolCallID)

			agent.Memory.AppendTempPayload(
				msgId,
				schema.ToolMessage(output.Response, toolCallID, schema.WithToolName(toolName)))
			agent.FlushMsg(msgId)
			agent.AgentContext.Printer.Send(ctx, msgId, toolName, output.Response, true)
			return ctx
		},
		OnEndWithStreamOutput: func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[*tool.CallbackOutput]) context.Context {
			toolName := info.Name
			toolCallID := compose.GetToolCallID(ctx)
			msgId := getMessageIdByCallToolId(agent, toolCallID)

			handleStreamOutput(ctx, output,
				func(chunk *tool.CallbackOutput) {
					if chunk.Response != "" {
						agent.AgentContext.Printer.Send(ctx, msgId, toolName, chunk.Response, false)
					}
				},
				func(full string) {
					if full != "" {
						agent.Memory.AppendTempPayload(
							msgId,
							schema.ToolMessage(full, toolCallID, schema.WithToolName(toolName)),
						)
						agent.FlushMsg(msgId)
						// agent.Memory.AddMessage(msgId, toolName, schema.ToolMessage(full, toolCallID, schema.WithToolName(toolName)))
						agent.AgentContext.Printer.Send(ctx, msgId, toolName, full, true)
					}
				},
			)
			return ctx
		},
	}
}

func handleStreamOutput[T any](
	ctx context.Context,
	output *schema.StreamReader[*T],
	onChunk func(chunk *T),
	onEnd func(full string),
) {
	// go func() {
	var full string
	for {
		chunk, err := output.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logs.ErrorContextf(ctx, "[handleStreamOutput] read stream failed: %v", err)
			onEnd(full)
			return
		}
		onChunk(chunk)
		switch v := any(chunk).(type) {
		case *model.CallbackOutput:
			if v.Message != nil {
				full += v.Message.Content
			}
		case *tool.CallbackOutput:
			full += v.Response
		}
	}
	onEnd(full)
	// }()
}

func getMessageIdByCallToolId(agent *BaseAgent, callToolId string) string {
	execInfo := agent.Memory.CallbackExecInfo[callToolId]
	if execInfo == nil {
		return ""
	}
	return execInfo.MessageId
}
