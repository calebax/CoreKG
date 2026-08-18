package qachatnodes

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/compose"
	"github.com/insmtx/corekg/apps/einonodes/nodebase"
	"github.com/ygpkg/yg-go/logs"
)

func ChunkChatBuilder[I, O, S any](ctx context.Context, genFunc compose.GenLocalState[S]) (compose.Runnable[I, O], error) {
	g := compose.NewGraph[I, O](
		compose.WithGenLocalState(genFunc),
	)

	if err := g.AddLambdaNode(NodeQAPairDataLoader, compose.InvokableLambdaWithOption(newDataLoader().QAPairNode)); err != nil {
		logs.ErrorContextf(ctx, "[ChunkChatBuilder] add basic data loader node error: %v", err)
		return nil, err
	}
	if err := g.AddLambdaNode(NodeChunkChatIntentRecognizer, compose.InvokableLambdaWithOption(newIntent().NewReferencesIntentLambdaNode)); err != nil {
		logs.ErrorContextf(ctx, "[ChunkChatBuilder] add intent recognizer node error: %v", err)
		return nil, err
	}
	if err := g.AddLambdaNode(NodeChunkChatReferencesChunkDataLoader, compose.InvokableLambdaWithOption(newDataLoader().QueryChunkReferenceNode)); err != nil {
		logs.ErrorContextf(ctx, "[ChunkChatBuilder] add references chunk loader node error: %v", err)
		return nil, err
	}
	if err := g.AddLambdaNode(NodeChunkChatLLMExecutor, compose.InvokableLambdaWithOption(newExecutor().NewChunkChatNode)); err != nil {
		logs.ErrorContextf(ctx, "[ChunkChatBuilder] add llm chat node error: %v", err)
		return nil, err
	}

	if err := g.AddEdge(compose.START, NodeQAPairDataLoader); err != nil {
		logs.ErrorContextf(ctx, "[ChunkChatBuilder] add edge error: %v", err)
		return nil, err
	}
	if err := g.AddEdge(NodeQAPairDataLoader, NodeChunkChatIntentRecognizer); err != nil {
		logs.ErrorContextf(ctx, "[ChunkChatBuilder] add edge error: %v", err)
		return nil, err
	}
	if err := g.AddEdge(NodeChunkChatIntentRecognizer, NodeChunkChatReferencesChunkDataLoader); err != nil {
		logs.ErrorContextf(ctx, "[ChunkChatBuilder] add edge error: %v", err)
		return nil, err
	}
	if err := g.AddEdge(NodeChunkChatReferencesChunkDataLoader, NodeChunkChatLLMExecutor); err != nil {
		logs.ErrorContextf(ctx, "[ChunkChatBuilder] add edge error: %v", err)
		return nil, err
	}

	if err := g.AddEdge(NodeChunkChatLLMExecutor, compose.END); err != nil {
		logs.ErrorContextf(ctx, "[ChunkChatBuilder] add edge error: %v", err)
		return nil, err
	}

	r, err := g.Compile(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "[ChunkChatBuilder] compile error: %v", err)
		return nil, err
	}
	return r, nil
}

func ExcelChatBuilder[I, O, S any](ctx context.Context, genFunc compose.GenLocalState[S]) (compose.Runnable[I, O], error) {
	g := compose.NewGraph[I, O](
		compose.WithGenLocalState(genFunc),
	)

	branchMap := map[string]bool{
		NodeBlackExecutor:  true,
		NodeNoDataReporter: true,
		compose.END:        true,
	}

	// 添加节点
	dataLoader := newDataLoader()
	executor := newExecutor()
	reporter := newReporter()

	if err := g.AddLambdaNode(NodeExcelDDLDataLoader, compose.InvokableLambdaWithOption(dataLoader.ExcelDDLNode)); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add excel ddl loader node error: %v", err)
		return nil, err
	}
	if err := g.AddLambdaNode(NodeExcelRunSQLExecutor, compose.InvokableLambdaWithOption(executor.RunMySQLStatementNode)); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add run sql node error: %v", err)
		return nil, err
	}
	if err := g.AddLambdaNode(NodeExcelTransferSQLToAnswerExecutor, compose.InvokableLambdaWithOption(executor.TransferSQLNode)); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add transfer sql to answer node error: %v", err)
		return nil, err
	}
	if err := g.AddLambdaNode(NodeNoDataReporter, compose.InvokableLambdaWithOption(reporter.NoDataReportNode)); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add transfer sql to answer node error: %v", err)
		return nil, err
	}

	if err := g.AddLambdaNode(NodeBlackExecutor, compose.InvokableLambdaWithOption(executor.BlackNode)); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add black executor node error: %v", err)
		return nil, err
	}

	if err := g.AddLambdaNode(NodeBlackExecutor+"2", compose.InvokableLambdaWithOption(executor.BlackNodeCompose)); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add black executor node 2 error: %v", err)
		return nil, err
	}

	if err := g.AddLambdaNode(NodeMysqlGenerateECharts, compose.InvokableLambdaWithOption(executor.GenerateMysqlChart)); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add generate echarts by chart type node error: %v", err)
		return nil, err
	}
	if err := g.AddLambdaNode(NodeEChartsReporter, compose.InvokableLambdaWithOption(reporter.EChartReportNode)); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add echarts reporter node error: %v", err)
		return nil, err
	}

	// 添加边
	if err := g.AddEdge(compose.START, NodeExcelDDLDataLoader); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add edge %s -> %s fail, error: %v", compose.START, NodeExcelDDLDataLoader, err)
		return nil, err
	}
	if err := g.AddEdge(NodeExcelDDLDataLoader, NodeExcelRunSQLExecutor); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add edge %s -> %s fail, error: %v", NodeExcelDDLDataLoader, NodeExcelRunSQLExecutor, err)
		return nil, err
	}
	// 使用标准的NewGraphBranch函数创建分支
	if err := g.AddBranch(NodeExcelRunSQLExecutor, compose.NewGraphBranch(func(ctx context.Context, input nodebase.RecordList) (next string, err error) {
		next = NodeBlackExecutor
		_ = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
			sqlResValue := input.Get(RecordKeyQueryStatementRes).Value
			var sqlResList []map[string]any
			if err := json.Unmarshal([]byte(sqlResValue), &sqlResList); err != nil {
				logs.ErrorContextf(state.Ctx, "[CheckStatementRes] unmarshal sql res error: %v", err)
				return err
			}
			if len(sqlResList) == 0 {
				next = NodeNoDataReporter
				return nil
			}
			return nil
		})
		return next, nil
	}, branchMap)); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add branch error: %v", err)
		return nil, err
	}
	if err := g.AddEdge(NodeNoDataReporter, compose.END); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add edge %s -> %s fail, error: %v", NodeNoDataReporter, compose.END, err)
		return nil, err
	}

	if err := g.AddEdge(NodeBlackExecutor, NodeExcelTransferSQLToAnswerExecutor); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add edge %s -> %s fail, error: %v", NodeBlackExecutor, NodeExcelTransferSQLToAnswerExecutor, err)
		return nil, err
	}
	if err := g.AddEdge(NodeBlackExecutor, NodeMysqlGenerateECharts); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add edge %s -> %s fail, error: %v", NodeBlackExecutor, NodeMysqlGenerateECharts, err)
		return nil, err
	}

	if err := g.AddEdge(NodeExcelTransferSQLToAnswerExecutor, NodeBlackExecutor+"2"); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add edge %s -> %s2 fail, error: %v", NodeExcelTransferSQLToAnswerExecutor, NodeBlackExecutor+"2", err)
		return nil, err
	}
	if err := g.AddEdge(NodeMysqlGenerateECharts, NodeEChartsReporter); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add edge %s -> %s2 fail, error: %v", NodeMysqlGenerateECharts, NodeBlackExecutor+"2", err)
		return nil, err
	}
	if err := g.AddEdge(NodeEChartsReporter, NodeBlackExecutor+"2"); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add edge %s2 -> %s fail, error: %v", NodeEChartsReporter, NodeBlackExecutor+"2", err)
		return nil, err
	}
	if err := g.AddEdge(NodeBlackExecutor+"2", compose.END); err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] add edge %s -> %s fail, error: %v", NodeBlackExecutor+"2", compose.END, err)
		return nil, err
	}

	r, err := g.Compile(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "[ExcelChatBuilder] compile error: %v", err)
		return nil, err
	}
	return r, nil
}

var AvalibaleAgentList []string

// ExternalDataToolsBuilder 外部数据源搜索
func ExternalDataToolsBuilder[I, O, S any](ctx context.Context, genFunc compose.GenLocalState[S]) (compose.Runnable[I, O], error) {
	g, err := ExternalDataToolsGraph[I, O](ctx, genFunc)
	if err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataToolsBuilder] compile error: %v", err)
		return nil, err
	}
	r, err := g.Compile(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] compile error: %v", err)
		return nil, err
	}
	return r, nil
}

// ExternalChatBuilder 外部数据源问答
func ExternalChatBuilder[I, O, S any](ctx context.Context, genFunc compose.GenLocalState[S]) (compose.Runnable[I, O], error) {
	g := compose.NewGraph[I, O](
		compose.WithGenLocalState(genFunc),
	)
	externalToolGraph, err := ExternalDataToolsGraph[I, O](ctx, genFunc)
	if err != nil {
		logs.ErrorContextf(ctx, "[ExternalChatBuilder] compile error: %v", err)
		return nil, err
	}
	if err := g.AddGraphNode(NodeExternalSearchNode, externalToolGraph); err != nil {
		logs.ErrorContextf(ctx, "[ExternalChatBuilder] add gmail loader node error: %v", err)
		return nil, err
	}
	if err := g.AddLambdaNode(NodeExternalChat, compose.InvokableLambdaWithOption(newExecutor().ExternalChatNode)); err != nil {
		logs.ErrorContextf(ctx, "[ExternalChatBuilder] add excel ddl loader node error: %v", err)
		return nil, err
	}
	// 添加边
	if err := g.AddEdge(compose.START, NodeExternalSearchNode); err != nil {
		logs.ErrorContextf(ctx, "[ExternalChatBuilder] add edge %s -> %s fail, error: %v", compose.START, NodeExternalSearchNode, err)
		return nil, err
	}
	// 添加边
	if err := g.AddEdge(NodeExternalSearchNode, NodeExternalChat); err != nil {
		logs.ErrorContextf(ctx, "[ExternalChatBuilder] add edge %s -> %s fail, error: %v", NodeExternalSearchNode, NodeExternalChat, err)
		return nil, err
	}
	// 添加边
	if err := g.AddEdge(NodeExternalChat, compose.END); err != nil {
		logs.ErrorContextf(ctx, "[ExternalChatBuilder] add edge %s -> %s fail, error: %v", NodeExternalChat, compose.END, err)
		return nil, err
	}
	r, err := g.Compile(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "[ExternalChatBuilder] compile error: %v", err)
		return nil, err
	}
	return r, nil
}
