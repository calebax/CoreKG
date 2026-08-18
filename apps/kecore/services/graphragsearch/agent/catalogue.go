package agent

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/insmtx/corekg/apps/kecore/services/graphragsearch/search"
	"github.com/ygpkg/yg-go/logs"
)

const (
	cataloguePrompt = `# 你是一个专业的目录节点选择助手，能够根据用户提供的目录结构，选择出最相关的关键节点以回答用户的问题。
# 节点图图谱中找到与用户问题相关目录结构节点返回我节点名词,节点名返回时通过英文逗号分隔,返回我完整的目录名称，严禁省略任何一个字符。
# 目录结构图如下：
` + "```json\n{tree}\n```" + `
# 用户问题如下：
{query}

# 注意 
只需要返回节点名词,不要多余的描述。
例如
节点1,节点2,节点3,P16E016/P16E116/P16E216-- 故障诊断,节点5
不要在节点中添加任何字符，不要捏造任何节点，返回我完整的目录名称，严禁省略任何一个字符
`
	nebulaSpace = "a_car_test"
)

// 根据目录获取关键节点agent
func NewCatalpgueAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tree []string) adk.Agent {
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "TreeNodeSelector",
		Description: "根据目录获取关键节点",
		Model:       chatModel,
		Instruction: `你是一个专业的目录节点选择助手，能够根据用户提供的目录结构，选择出最相关的关键节点以回答用户的问题。`,
		GenModelInput: func(ctx context.Context, instruction string, input *adk.AgentInput) ([]adk.Message, error) {
			logs.InfoContextf(ctx, "NewCatalpgueAgent GenModelInput : %s", logs.JSON(input))
			ct := prompt.FromMessages(schema.FString,
				schema.SystemMessage(instruction),
				schema.UserMessage(cataloguePrompt),
			)
			msgs, err := ct.Format(ctx, map[string]any{
				"tree":  tree,
				"query": input.Messages[0].Content,
			})
			if err != nil {
				logs.ErrorContextf(ctx, "NewCatalpgueAgent GenModelInput ct.Format error: %v", err)
				return nil, err
			}

			return msgs, nil
		},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "NewCatalpgueAgent NewChatModelAgent error: %v", err)
		return nil
	}
	return a
}

// NewCatalpgueParallelAgent 并行目录节点选择代理
func NewCatalpgueParallelAgent(ctx context.Context, chatModel model.ToolCallingChatModel) (adk.Agent, error) {

	cli, err := nebulagraph.NewNebulaCLI(ctx, nebulaSpace)
	if err != nil {
		logs.ErrorContextf(ctx, "NewCatalpgueParallelAgent NewNebulaCLI error: %v", err)
		return nil, err
	}
	defer cli.Release()

	zhenduantitle, err := search.GetTitleList(ctx, cli, "诊断手册")
	if err != nil {
		logs.ErrorContextf(ctx, "NewCatalpgueParallelAgent GetTitleList error: %v", err)
		return nil, err
	}
	zhenduan := NewCatalpgueAgent(ctx, chatModel, zhenduantitle)

	weixiutitle, err := search.GetTitleList(ctx, cli, "维修手册")
	if err != nil {
		logs.ErrorContextf(ctx, "NewCatalpgueParallelAgent GetTitleList error: %v", err)
		return nil, err
	}
	weixiu := NewCatalpgueAgent(ctx, chatModel, weixiutitle)

	dianlutitle, err := search.GetTitleList(ctx, cli, "电路图")
	if err != nil {
		logs.ErrorContextf(ctx, "NewCatalpgueParallelAgent GetTitleList error: %v", err)
		return nil, err
	}
	dianlu := NewCatalpgueAgent(ctx, chatModel, dianlutitle)

	parallelAgent, err := adk.NewParallelAgent(ctx, &adk.ParallelAgentConfig{
		Name:        "MultiPerspectiveAnalyzer",
		Description: "获取多个目录节点选择结果",
		SubAgents:   []adk.Agent{zhenduan, weixiu, dianlu},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "NewCatalpgueParallelAgent NewParallelAgent error: %v", err)
		return nil, err
	}
	return parallelAgent, err
}

// ExecuteCatalogueAgent 执行目录节点选择agent
func ExecuteCatalogueAgent(ctx context.Context, chatModel model.ToolCallingChatModel, query string) ([]search.DirectoryInfo, error) {
	ag, err := NewCatalpgueParallelAgent(ctx, chatModel)
	if err != nil {
		logs.ErrorContextf(ctx, "ExecuteCatalogueAgent NewAnalystAgent error: %v", err)
		return nil, err
	}
	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: ag,
	})
	iter := runner.Query(ctx, query)

	var results []string
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			logs.ErrorContextf(ctx, "ExecuteCatalogueAgent iter.Next error: %v", event.Err)
			continue
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			if m := event.Output.MessageOutput.Message; m != nil {
				if len(m.Content) > 0 {
					logs.InfoContextf(ctx, "ExecuteCatalogueAgent answer: %s", m.Content)
					results = append(results, strings.Split(m.Content, ",")...)
				}
			}
		}
	}
	cli, err := nebulagraph.NewNebulaCLI(ctx, nebulaSpace)
	if err != nil {
		logs.ErrorContextf(ctx, "NewCatalpgueParallelAgent NewNebulaCLI error: %v", err)
		return nil, err
	}
	defer cli.Release()
	res, err := search.GetNodesInfo(ctx, cli, results)
	if err != nil {
		logs.ErrorContextf(ctx, "NewCatalpgueParallelAgent GetNodesInfo error: %v", err)
		return nil, err
	}

	return res, nil
}
