package qachatnodes

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/ygpkg/yg-go/logs"
)

// ExternalDataToolsGraph 外部数据源搜索
func ExternalDataToolsGraph[I, O, S any](ctx context.Context, genFunc compose.GenLocalState[S]) (*compose.Graph[I, O], error) {
	g := compose.NewGraph[I, O](
		compose.WithGenLocalState(genFunc),
	)
	testNode := "test_search"
	if err := g.AddLambdaNode(NodeGetESKeyWords, compose.InvokableLambdaWithOption(newDataLoader().GetESAnalyzeNode)); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add gmail loader node error: %v", err)
		return nil, err
	}
	if err := g.AddLambdaNode(NodeGetSessionTools, compose.InvokableLambdaWithOption(newDataLoader().GetSessionToolsNode)); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add gmail loader node error: %v", err)
		return nil, err
	}
	if err := g.AddLambdaNode(NodeGmailSearch, compose.InvokableLambdaWithOption(newDataLoader().GmailDataNode)); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add gmail loader node error: %v", err)
		return nil, err
	}
	if err := g.AddLambdaNode(NodeSlackSearch, compose.InvokableLambdaWithOption(newDataLoader().SlackDataNode)); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add gmail loader node error: %v", err)
		return nil, err
	}
	if err := g.AddLambdaNode(NodeGoogleDriveSearch, compose.InvokableLambdaWithOption(newDataLoader().GoogleDriveDataNode)); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add gmail loader node error: %v", err)
		return nil, err
	}
	if err := g.AddLambdaNode(NodeConfluenceSearch, compose.InvokableLambdaWithOption(newDataLoader().ConfluenceDataNode)); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add confluence loader node error: %v", err)
		return nil, err
	}
	if err := g.AddLambdaNode(testNode, compose.InvokableLambdaWithOption(newDataLoader().TestNode)); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add test loader node error: %v", err)
		return nil, err
	}

	// 添加边
	if err := g.AddEdge(compose.START, NodeGetESKeyWords); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add edge %s -> %s fail, error: %v", compose.START, NodeGetESKeyWords, err)
		return nil, err
	}
	// 添加边
	if err := g.AddEdge(NodeGetESKeyWords, NodeGetSessionTools); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add edge %s -> %s fail, error: %v", NodeGetESKeyWords, NodeGetSessionTools, err)
		return nil, err
	}

	// 添加边
	if err := g.AddEdge(NodeGetSessionTools, NodeGmailSearch); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add edge %s -> %s fail, error: %v", NodeGetSessionTools, NodeGmailSearch, err)
		return nil, err
	}
	// 添加边
	if err := g.AddEdge(NodeGetSessionTools, NodeSlackSearch); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add edge %s -> %s fail, error: %v", NodeGetSessionTools, NodeSlackSearch, err)
		return nil, err
	}
	// 添加边
	if err := g.AddEdge(NodeGetSessionTools, NodeGoogleDriveSearch); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add edge %s -> %s fail, error: %v", NodeGetSessionTools, NodeGoogleDriveSearch, err)
		return nil, err
	}
	// 添加边
	if err := g.AddEdge(NodeGetSessionTools, NodeConfluenceSearch); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add edge %s -> %s fail, error: %v", NodeGetSessionTools, NodeConfluenceSearch, err)
		return nil, err
	}
	// 添加边
	if err := g.AddEdge(NodeGetSessionTools, testNode); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add edge %s -> %s fail, error: %v", NodeGetSessionTools, testNode, err)
		return nil, err
	}
	if err := g.AddEdge(NodeSlackSearch, compose.END); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add edge %s -> %s fail, error: %v", NodeSlackSearch, compose.END, err)
		return nil, err
	}
	if err := g.AddEdge(NodeGmailSearch, compose.END); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add edge %s -> %s fail, error: %v", NodeGmailSearch, compose.END, err)
		return nil, err
	}
	if err := g.AddEdge(NodeGoogleDriveSearch, compose.END); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add edge %s -> %s fail, error: %v", NodeGoogleDriveSearch, compose.END, err)
		return nil, err
	}
	if err := g.AddEdge(NodeConfluenceSearch, compose.END); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add edge %s -> %s fail, error: %v", NodeConfluenceSearch, compose.END, err)
		return nil, err
	}
	if err := g.AddEdge(testNode, compose.END); err != nil {
		logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] add edge %s -> %s fail, error: %v", testNode, compose.END, err)
		return nil, err
	}
	return g, nil
}
