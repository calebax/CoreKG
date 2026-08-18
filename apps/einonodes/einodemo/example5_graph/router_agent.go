package example5graph

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"
	"github.com/ygpkg/yg-go/logs"
)

// 节点常量定义
const (
	NodeIntentRecognizer   = "intent_recognizer"
	NodeKnowledgeGraph     = "Knowledge_graph"
	NodeIntentToLLM        = "intent_to_llm"
	NodeIntentToRagRequest = "intent_to_rag_request"
	NodeSummarizer         = "summarizer"
)

// 记录键常量定义
const (
	RecordKeyQuestion      = "question"
	RecordKeyIntent        = "intent"
	RecordKeySearchResults = "search_results"
	RecordKeyAnswer        = "answer"
)

// 意图类型常量
const (
	IntentDirect = "direct"
	IntentQA     = "qa"
)

// Agent 问答代理结构体
type TtAgent struct {
	runnable compose.Runnable[[]*schema.Message, *schema.Message]
}

type Config struct {
	ChatModel model.BaseChatModel
}

type state struct {
	messages []*schema.Message
}

// NewQAAgent 创建新的问答代理
func NewQAAgent(ctx context.Context, config *Config) (*TtAgent, error) {
	// 创建图
	gen := func(ctx context.Context) *state {
		return &state{}
	}
	g := compose.NewGraph[[]*schema.Message, *schema.Message](compose.WithGenLocalState(gen))

	modelPreHandle := func(prompt string) compose.StatePreHandler[[]*schema.Message, *state] {
		return func(ctx context.Context, input []*schema.Message, state *state) ([]*schema.Message, error) {
			state.messages = append(state.messages, input...)
			return append([]*schema.Message{schema.SystemMessage(prompt)}, state.messages...), nil
		}
	}

	// 添加意图识别节点
	_ = g.AddChatModelNode(NodeIntentRecognizer, config.ChatModel, compose.WithNodeName(NodeIntentRecognizer),
		compose.WithStatePreHandler(modelPreHandle(`你是一个意图识别助手。你需要判断用户的问题是需要直接回答，还是需要通过检索知识库后回答。
如果问题是一般性的、常识性的，或者是不需要特定知识库的，请返回"direct"。
如果问题可能需要特定知识库或文档的支持，请返回"qa"。
只返回"direct"或"qa"，不要有其他内容。`)))

	_ = g.AddChatModelNode(NodeSummarizer, config.ChatModel, compose.WithNodeName(NodeSummarizer), compose.WithStatePreHandler(modelPreHandle("你是一个有用的AI助手, 善于帮助人类解决问题")))

	_ = g.AddLambdaNode(NodeIntentToRagRequest, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (KnowledgeRequest, error) {
		var msg string
		_ = compose.ProcessState[*state](ctx, func(_ context.Context, state *state) error {
			for i := len(state.messages) - 1; i >= 0; i-- {
				if state.messages[i].Role == "user" {
					msg = state.messages[i].Content
					break
				}
			}
			return nil
		})

		return KnowledgeRequest{
			ChatModel:    config.ChatModel,
			Question:     msg,
			UserId:       "123",
			KnowledgeIds: []string{"知识库id1"},
		}, nil
	}))
	_ = g.AddLambdaNode(NodeIntentToLLM, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) ([]*schema.Message, error) {
		var messages []*schema.Message
		_ = compose.ProcessState[*state](ctx, func(_ context.Context, state *state) error {
			messages = state.messages
			state.messages = []*schema.Message{}
			return nil
		})
		return messages, nil
	}))

	// 添加内容检索节点
	g.AddGraphNode(NodeKnowledgeGraph, KnowledgeGraph(ctx, KnowledgeRequest{
		ChatModel: config.ChatModel,
	}), compose.WithNodeName(NodeKnowledgeGraph))

	// 添加边：开始 -> 意图识别
	_ = g.AddEdge(compose.START, NodeIntentRecognizer)
	// 添加分支：根据意图决定走直接回答还是问答子图
	_ = g.AddBranch(NodeIntentRecognizer, compose.NewGraphBranch(intentBranchCondition, map[string]bool{
		NodeIntentToLLM:        true,
		NodeIntentToRagRequest: true,
	}))
	_ = g.AddEdge(NodeIntentToRagRequest, NodeKnowledgeGraph)
	// 添加边：知识库问答
	_ = g.AddEdge(NodeKnowledgeGraph, compose.END)
	// 添加边：用户问题 -> 直接回答
	_ = g.AddEdge(NodeIntentToLLM, NodeSummarizer)
	// llm 处理回复
	_ = g.AddEdge(NodeSummarizer, compose.END)

	// 编译图
	runnable, err := g.Compile(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "[NewQAAgent] compile graph failed: %v", err)
		return nil, err
	}

	return &TtAgent{
		runnable: runnable,
	}, nil
}

// 意图分支条件函数
func intentBranchCondition(ctx context.Context, msg *schema.Message) (string, error) {
	if msg.Content == IntentQA {
		return NodeIntentToRagRequest, nil // 问答子图
	}
	return NodeIntentToLLM, nil // 直接LLM回答
}

// Generate 以非流式的方式调用多智能体.
func (r *TtAgent) Generate(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (output *schema.Message, err error) {
	output, err = r.runnable.Invoke(ctx, input, agent.GetComposeOptions(opts...)...)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// Stream 以流式的方式调用多智能体.
func (r *TtAgent) Stream(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (
	output *schema.StreamReader[*schema.Message], err error) {
	res, err := r.runnable.Stream(ctx, input, agent.GetComposeOptions(opts...)...)
	if err != nil {
		return nil, err
	}

	return res, nil
}
