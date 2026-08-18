package einodemo

import (
	"context"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/insmtx/corekg/apps/einonodes/nodebase"
	"github.com/ygpkg/yg-go/logs"
)

func SingleGraph(ctx context.Context) (compose.Runnable[nodebase.RecordList, nodebase.RecordList], error) {
	const (
		node1 = "node_1"
		node2 = "node_2"
		node3 = "node_3"
		node4 = "node_4"
		node5 = "node_5"
		node6 = "node_6"
	)
	// 创建图
	g := compose.NewGraph[nodebase.RecordList, nodebase.RecordList]()

	// 添加节点
	g.AddLambdaNode(node1, compose.InvokableLambda(func(ctx context.Context, input nodebase.RecordList) (output nodebase.RecordList, err error) {
		return input, nil
	}))
	g.AddLambdaNode(node2, compose.InvokableLambda(func(ctx context.Context, input nodebase.RecordList) (output nodebase.RecordList, err error) {
		return input, nil
	}))
	g.AddLambdaNode(node3, compose.InvokableLambda(func(ctx context.Context, input nodebase.RecordList) (output nodebase.RecordList, err error) {
		time.Sleep(3 * time.Second)
		return input, nil
	}))
	g.AddLambdaNode(node4, compose.InvokableLambda(func(ctx context.Context, input nodebase.RecordList) (output nodebase.RecordList, err error) {
		return input, nil
	}))
	g.AddLambdaNode(node5, compose.InvokableLambda(func(ctx context.Context, input nodebase.RecordList) (output string, err error) {
		return "summary", nil
	}))
	g.AddLambdaNode(node6, compose.InvokableLambda(func(ctx context.Context, input string) (output nodebase.RecordList, err error) {
		return nodebase.RecordList{
			&nodebase.Record{
				Key:   "question",
				Value: "234",
			}}, nil
	}))

	if err := g.AddEdge(compose.START, node1); err != nil {
		logs.ErrorContextf(ctx, "[SingleGraph] add edge %s -> %s failed, error: %v", compose.START, node1, err)
		return nil, err
	}
	if err := g.AddEdge(node1, node2); err != nil {
		logs.ErrorContextf(ctx, "[SingleGraph] add edge %s -> %s failed, error: %v", node1, node2, err)
		return nil, err
	}
	if err := g.AddEdge(node1, node3); err != nil {
		logs.ErrorContextf(ctx, "[SingleGraph] add edge %s -> %s failed, error: %v", node1, node2, err)
		return nil, err
	}
	// if err := g.AddEdge(node2, compose.END); err != nil {
	// 	logs.ErrorContextf(ctx, "[SingleGraph] add edge %s -> %s failed, error: %v", node2, compose.END, err)
	// 	return nil, err
	// }
	// if err := g.AddEdge(node3, compose.END); err != nil {
	// 	logs.ErrorContextf(ctx, "[SingleGraph] add edge %s -> %s failed, error: %v", node2, compose.END, err)
	// 	return nil, err
	// }
	if err := g.AddEdge(node2, node4); err != nil {
		logs.ErrorContextf(ctx, "[SingleGraph] add edge %s -> %s failed, error: %v", node2, compose.END, err)
		return nil, err
	}
	if err := g.AddEdge(node3, node4); err != nil {
		logs.ErrorContextf(ctx, "[SingleGraph] add edge %s -> %s failed, error: %v", node2, compose.END, err)
		return nil, err
	}
	if err := g.AddEdge(node4, node5); err != nil {
		logs.ErrorContextf(ctx, "[SingleGraph] add edge %s -> %s failed, error: %v", node2, compose.END, err)
		return nil, err
	}
	if err := g.AddEdge(node5, node6); err != nil {
		logs.ErrorContextf(ctx, "[SingleGraph] add edge %s -> %s failed, error: %v", node2, compose.END, err)
		return nil, err
	}
	if err := g.AddEdge(node6, compose.END); err != nil {
		logs.ErrorContextf(ctx, "[SingleGraph] add edge %s -> %s failed, error: %v", node6, compose.END, err)
		return nil, err
	}
	return g.Compile(ctx)
}
