package agent

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	ub "github.com/cloudwego/eino/utils/callbacks"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	sysPrompt "github.com/insmtx/corekg/pkgs/einotools/prompt"
)

type SummaryAgent struct {
	BaseAgent

	runnable compose.Runnable[[]*schema.Message, *schema.Message]
	graph    *compose.Graph[[]*schema.Message, *schema.Message]
}

func NewSummaryAgent(ctx context.Context, agentContext *AgentContext) (*SummaryAgent, error) {
	agent := &SummaryAgent{}
	agent.Name = "summaryAgent"
	agent.Memory = models.NewMemory()
	agent.MaxStep = 2
	agent.AgentContext = agentContext
	agent.ChatModel = agentContext.ChatModel
	agent.SystemPrompt = agentContext.SummarySystemPrompt
	if agent.SystemPrompt == "" {
		agent.SystemPrompt = sysPrompt.SummarySystemPrompt
	}
	agent.Printer = agentContext.Printer
	agent.Stats = models.AgentStats{}

	err := agent.buildRunnable(ctx)
	if err != nil {
		return nil, err
	}

	return agent, nil
}

func (agent *SummaryAgent) buildRunnable(ctx context.Context) error {
	type state struct{}
	g := compose.NewGraph[[]*schema.Message, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *state {
			return &state{}
		}),
	)

	const (
		nodePrepare    = "prepare_input"
		nodeSummarizer = "summarizer_model"
	)

	preBuildMsg := func(ctx context.Context, input []*schema.Message, st *state) ([]*schema.Message, error) {
		sysPromptTemplate := prompt.FromMessages(schema.GoTemplate,
			schema.UserMessage(agent.SystemPrompt),
		)

		var historyBuilder strings.Builder
		for _, msg := range input {
			fmt.Fprintf(&historyBuilder, "role:%s content:%s\n", msg.Role, msg.Content)
		}

		params := map[string]any{
			"date":             agent.AgentContext.DateInfo,
			"roleName":         agent.AgentContext.ModelRoleName,
			"query":            agent.AgentContext.Query,
			"taskHistory":      historyBuilder.String(),
			"history_dialogue": agent.AgentContext.AgentRequest.GetInitialMessage(),
		}
		messages, err := sysPromptTemplate.Format(ctx, params)
		if err != nil {
			return nil, err
		}
		return messages, nil
	}

	_ = g.AddLambdaNode(nodePrepare, compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) ([]*schema.Message, error) {
		if len(input) == 0 {
			return []*schema.Message{schema.UserMessage("")}, nil
		}
		return input, nil
	}), compose.WithNodeName(nodePrepare))

	_ = g.AddChatModelNode(nodeSummarizer, agent.ChatModel,
		compose.WithNodeName(nodeSummarizer),
		compose.WithStatePreHandler(preBuildMsg),
	)

	_ = g.AddEdge(compose.START, nodePrepare)
	_ = g.AddEdge(nodePrepare, nodeSummarizer)
	_ = g.AddEdge(nodeSummarizer, compose.END)

	runnable, err := g.Compile(ctx, compose.WithNodeTriggerMode(compose.AnyPredecessor), compose.WithMaxRunSteps(agent.MaxStep))
	if err != nil {
		return err
	}

	agent.graph = g
	agent.runnable = runnable
	return nil
}

func (agent *SummaryAgent) RunSummarizeResult(ctx context.Context, msgs []*schema.Message) (string, error) {
	agent.State = IDLE
	defer agent.Stats.Stop()

	modelHandlerCalls := ub.NewHandlerHelper().
		ChatModel(agent.newModelHandler(models.MsgTypeResult)).
		Handler()

	if agent.AgentContext.IsStream {
		stream, err := agent.runnable.Stream(ctx, msgs, compose.WithCallbacks(modelHandlerCalls))
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
				return "", fmt.Errorf("summary agent stream receive failed: %w", err)
			}
		}
	} else {
		_, err := agent.runnable.Invoke(ctx, msgs, compose.WithCallbacks(modelHandlerCalls))
		if err != nil {
			agent.State = ERROR
			return "", err
		}
	}

	agent.State = FINISHED
	lastMsg := agent.Memory.GetLastMessage()
	if lastMsg == nil {
		return "", nil
	}
	return lastMsg.Payload.Content, nil
}
