package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	enagent "github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	ub "github.com/cloudwego/eino/utils/callbacks"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	sysPrompt "github.com/insmtx/corekg/pkgs/einotools/prompt"
	entools "github.com/insmtx/corekg/pkgs/einotools/tools"
)

type ReActAgent struct {
	BaseAgent

	executor                      *react.Agent
	debugModelInputSeq            atomic.Int64
	debugModelInputLogMu          sync.Mutex
	debugModelInputLastSignatures []string
}

const defaultMaxStep = 20

func NewReactAgent(ctx context.Context, agentContext *AgentContext) (*ReActAgent, error) {
	agent := &ReActAgent{}
	agent.Name = "reactAgent"
	agent.Memory = models.NewMemory()
	agent.MaxStep = agentContext.MaxStep
	if agent.MaxStep == 0 {
		agent.MaxStep = defaultMaxStep
	}
	agent.AgentContext = agentContext
	agent.ChatModel = agentContext.ChatModel
	agent.Printer = agentContext.Printer
	agent.Stats = models.AgentStats{}

	files, err := json.Marshal(agentContext.AgentRequest.InputFiles)
	if err != nil {
		return nil, err
	}
	systemPrompt := agentContext.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = sysPrompt.ReactSystemPrompt
	}
	nextStepPrompt := agentContext.NextStepPrompt
	if nextStepPrompt == "" {
		nextStepPrompt = sysPrompt.ReactNextStepPrompt
	}
	sysPromptTemplate := prompt.FromMessages(schema.GoTemplate,
		schema.SystemMessage(systemPrompt+sysPrompt.BasePrompt),
		schema.UserMessage(nextStepPrompt),
	)
	params := map[string]any{
		"roleName":         agentContext.ModelRoleName,
		"date":             agentContext.DateInfo,
		"files":            string(files),
		"history_dialogue": agentContext.AgentRequest.GetInitialMessage(),
		"query":            agentContext.Query,
	}
	messages, err := sysPromptTemplate.Format(ctx, params)
	if err != nil {
		return nil, err
	}
	agent.SystemPrompt = messages[0].Content
	agent.NextStepPrompt = messages[1].Content

	baseTools, err := entools.GetTools(ctx, agentContext.ChatModel, agentContext.Tools, agentContext.SaveChartFunc)
	if err != nil {
		return nil, err
	}
	if len(agentContext.AvailableTools) != 0 {
		baseTools = append(baseTools, agentContext.AvailableTools...)
	}
	logReActToolInfos(ctx, baseTools)
	tools := compose.ToolsNodeConfig{
		Tools: baseTools,
		UnknownToolsHandler: func(ctx context.Context, toolName, callInput string) (string, error) {
			return fmt.Sprintf("Tool '%s' invocation failed. Please verify that the tool exists and the parameters are correct.", toolName), nil
		},
	}

	einoAgent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: agentContext.ChatModel,
		ToolsConfig:      tools,
		MessageModifier: func(ctx context.Context, input []*schema.Message) []*schema.Message {
			// res := make([]*schema.Message, 0, 2+len(agent.Memory.Messages))
			res := make([]*schema.Message, 0, 2+len(input))
			res = append(res, schema.SystemMessage(agent.SystemPrompt))
			// res = append(res, agent.Memory.Messages...)
			res = append(res, input...)

			if agent.NextStepPrompt != "" {
				lastMsg := agent.Memory.GetLastMessage()
				if lastMsg != nil && lastMsg.Payload.Role != schema.User {
					res = append(res, schema.UserMessage(agent.NextStepPrompt))
				}
			}

			round := agent.debugModelInputSeq.Add(1)
			startIndex := agent.nextModelInputLogStart(res)
			logModelInputMessages(ctx, round, res, startIndex)
			return res
		},
		StreamToolCallChecker: func(_ context.Context, sr *schema.StreamReader[*schema.Message]) (bool, error) {
			defer sr.Close()
			hasTool := false
			for {
				msg, e := sr.Recv()
				if e != nil {
					if e == io.EOF {
						break
					}
					return false, e
				}
				if len(msg.ToolCalls) > 0 {
					hasTool = true
					// 不立刻返回，继续读到EOF以保持一致行为
				}
			}
			return hasTool, nil
		},
		MaxStep: agent.MaxStep,
	})
	if err != nil {
		return nil, err
	}

	agent.executor = einoAgent
	return agent, nil
}

func (agent *ReActAgent) nextModelInputLogStart(messages []*schema.Message) int {
	signatures := buildModelInputMessageSignatures(messages)

	agent.debugModelInputLogMu.Lock()
	defer agent.debugModelInputLogMu.Unlock()

	start := 0
	for start < len(signatures) &&
		start < len(agent.debugModelInputLastSignatures) &&
		signatures[start] == agent.debugModelInputLastSignatures[start] {
		start++
	}
	agent.debugModelInputLastSignatures = signatures
	return start
}

func (agent *ReActAgent) Run(ctx context.Context, query string, options models.HandlerOptions) (string, error) {
	agent.State = IDLE
	agent.Stats.StartTimestamp = time.Now().UnixMilli()
	defer agent.Stats.Stop()

	if len(query) != 0 {
		agent.Memory.AddMessageWithType("user", schema.UserMessage(query))
	}

	chatMsgMode := models.MsgTypeResult
	if options.SummaryMode {
		chatMsgMode = models.MsgTypeTaskThought
	}
	modelHandlerCalls := ub.NewHandlerHelper().
		ChatModel(agent.newModelHandler(chatMsgMode)).
		Handler()

	toolHandlerCalls := ub.NewHandlerHelper().
		Tool(agent.newToolHandler()).
		Handler()

	agentOptions := enagent.WithComposeOptions(
		// 只触发react agent内部node，复杂情况可采用DesignateNodeWithPath
		compose.WithCallbacks(modelHandlerCalls).DesignateNode("chat"),
		compose.WithCallbacks(toolHandlerCalls).DesignateNode("tools"),
	)

	if agent.AgentContext.IsStream {
		stream, err := agent.executor.Stream(ctx,
			agent.Memory.GetLlmMessages(),
			agentOptions)
		if err != nil {
			agent.State = ERROR
			return "", err
		}
		defer stream.Close()

		for {
			_, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				agent.State = ERROR
				return "", fmt.Errorf("react agent stream receive failed: %w", err)
			}
		}
	} else {
		_, err := agent.executor.Generate(ctx,
			agent.Memory.GetLlmMessages(),
			agentOptions)
		if err != nil {
			agent.State = ERROR
			return "", err
		}
	}

	agent.State = FINISHED
	return agent.Memory.GetLastMessage().Payload.Content, nil
}
