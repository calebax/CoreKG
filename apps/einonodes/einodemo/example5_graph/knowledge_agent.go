package example5graph

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// TODO

type KnowledgeRequest struct {
	ChatModel    model.BaseChatModel
	Question     string   `json:"question"`
	UserId       string   `json:"user_id"`
	KnowledgeIds []string `json:"knowledge_ids"`
	// 外部检索源
	RetrievalSources []string `json:"retrieval_sources"`
}

const (
	nodeKeyIntent          = "intent"           // 意图识别
	nodeKeyQuestionExpand  = "question_expand"  // 问题扩写
	nodeKeyVectorSearch    = "vector_search"    // 向量检索
	nodeKeyKeywordSearch   = "keyword_search"   // 关键词检索
	nodeKeyAggregate       = "aggregate"        // 结果聚合
	nodeKeyRank            = "rank"             // 结果排序
	nodeKeyContextCompress = "context_compress" // 上下文压缩
	nodeKeyLLMGenerate     = "llm_generate"     // LLM生成
	nodeKeyValidate        = "validate"         // 结果校验
)

type ragState struct {
	question     string
	userId       string   `json:"user_id"`
	knowledgeIds []string `json:"knowledge_ids"`
	// 外部检索源
	retrievalSources []string
}

// ========================
// KnowledgeGraph 构建
// ========================

func KnowledgeGraph(ctx context.Context, req KnowledgeRequest) *compose.Graph[KnowledgeRequest, *schema.Message] {

	gen := func(ctx context.Context) *ragState {
		return &ragState{
			question:         req.Question,
			userId:           req.UserId,
			knowledgeIds:     req.KnowledgeIds,
			retrievalSources: req.RetrievalSources,
		}
	}
	g := compose.NewGraph[KnowledgeRequest, *schema.Message](compose.WithGenLocalState(gen))

	_ = g.AddLambdaNode(nodeKeyIntent, compose.InvokableLambda(IntentNodeLambda), compose.WithNodeName(nodeKeyIntent))
	_ = g.AddLambdaNode(nodeKeyQuestionExpand, compose.InvokableLambda(QuestionExpandNodeLambda), compose.WithNodeName(nodeKeyQuestionExpand))
	_ = g.AddLambdaNode(nodeKeyVectorSearch, compose.InvokableLambda(VectorSearchNodeLambda), compose.WithNodeName(nodeKeyVectorSearch))
	_ = g.AddLambdaNode(nodeKeyKeywordSearch, compose.InvokableLambda(KeywordSearchNodeLambda), compose.WithNodeName(nodeKeyKeywordSearch))
	_ = g.AddLambdaNode(nodeKeyAggregate, compose.InvokableLambda(AggregateNodeLambda), compose.WithNodeName(nodeKeyAggregate))
	_ = g.AddLambdaNode(nodeKeyRank, compose.InvokableLambda(RankNodeLambda), compose.WithNodeName(nodeKeyRank))
	_ = g.AddLambdaNode(nodeKeyContextCompress, compose.InvokableLambda(ContextCompressNodeLambda), compose.WithNodeName(nodeKeyContextCompress))

	_ = g.AddChatModelNode(nodeKeyLLMGenerate, req.ChatModel, compose.WithNodeName(nodeKeyLLMGenerate))

	_ = g.AddLambdaNode(nodeKeyValidate, compose.InvokableLambda(ValidateNodeLambda), compose.WithNodeName(nodeKeyValidate))

	// 扇入合并逻辑
	// ~ 在 AddNode 时，可以通过添加 WithOutputKey 这个 Option 来把节点的输出转成 Map
	// ~ 或使用RegisterValuesMergeFunc
	compose.RegisterValuesMergeFunc(func(values [][]string) ([]string, error) {
		var merged []string
		for _, arr := range values {
			merged = append(merged, arr...)
		}
		return merged, nil
	})

	// 构建边
	_ = g.AddEdge(compose.START, nodeKeyIntent)
	g.AddEdge(nodeKeyIntent, nodeKeyQuestionExpand)
	g.AddEdge(nodeKeyQuestionExpand, nodeKeyVectorSearch)
	g.AddEdge(nodeKeyQuestionExpand, nodeKeyKeywordSearch)
	g.AddEdge(nodeKeyVectorSearch, nodeKeyAggregate)
	g.AddEdge(nodeKeyKeywordSearch, nodeKeyAggregate)
	g.AddEdge(nodeKeyAggregate, nodeKeyRank)
	g.AddEdge(nodeKeyRank, nodeKeyContextCompress)
	g.AddEdge(nodeKeyContextCompress, nodeKeyLLMGenerate)
	g.AddEdge(nodeKeyLLMGenerate, nodeKeyValidate)
	_ = g.AddEdge(nodeKeyValidate, compose.END)

	return g
}

// ========================
// Lambda Node 函数
// ========================

func IntentNodeLambda(ctx context.Context, req KnowledgeRequest) (KnowledgeRequest, error) {
	intent := "knowledge_query"
	fmt.Println("Intent result:", intent)
	_ = compose.ProcessState[*ragState](ctx, func(_ context.Context, state *ragState) error {
		state.retrievalSources = append(state.retrievalSources, "gmail", "notion")
		return nil
	})
	return req, nil
}

func QuestionExpandNodeLambda(ctx context.Context, req KnowledgeRequest) (string, error) {
	expandedQuestion := "如何使用Eino框架构建RAG应用？它有哪些优势？"
	fmt.Println("Expanded question:", expandedQuestion)
	return expandedQuestion, nil
}

func VectorSearchNodeLambda(ctx context.Context, question string) ([]string, error) {
	results := []string{
		"Eino是一个高效的AI应用开发框架，支持多种模型和工具的集成。",
		"Eino框架提供了Graph、Chain等多种组合方式，方便开发者构建复杂的AI应用。",
	}
	fmt.Println("Vector search results:", results)
	return results, nil
}

func KeywordSearchNodeLambda(ctx context.Context, question string) ([]string, error) {
	results := []string{
		"使用Eino可以快速实现RAG（检索增强生成）等应用模式。",
		"Eino的优势在于高效、灵活和易于扩展。",
	}
	fmt.Println("Keyword search results:", results)
	return results, nil
}

func AggregateNodeLambda(ctx context.Context, inputs []string) ([]string, error) {
	fmt.Println("Aggregated results:", inputs)
	return inputs, nil
}

func RankNodeLambda(ctx context.Context, results []string) ([]string, error) {
	var msg string
	_ = compose.ProcessState[*ragState](ctx, func(_ context.Context, state *ragState) error {
		msg = state.question
		return nil
	})
	fmt.Printf("%s \n Ranked results: %v\n", msg, results)
	return results, nil
}

func ContextCompressNodeLambda(ctx context.Context, results []string) ([]*schema.Message, error) {
	var msg string
	_ = compose.ProcessState[*ragState](ctx, func(_ context.Context, state *ragState) error {
		msg = state.question
		return nil
	})

	context := fmt.Sprintf("%s \n相关知识：\n", msg)
	for i, r := range results {
		context += fmt.Sprintf("%d. %s\n", i+1, r)
	}
	context += "\n\n请仅根据上述上下文回答问题，不要提及来源或上下文本身。"
	fmt.Println("Compressed context:", context)
	return []*schema.Message{
		{
			Role:    schema.System,
			Content: "你是一个专业的知识问答助手，只能根据提供的上下文中的相关知识，或通过搜索真实数据回答问题，不可以虚假或编造信息。",
		},
		{
			Role:    schema.User,
			Content: context,
		},
	}, nil
}

func ValidateNodeLambda(ctx context.Context, response *schema.Message) (*schema.Message, error) {
	fmt.Println("Validated response:", response.Content)
	return response, nil
}
