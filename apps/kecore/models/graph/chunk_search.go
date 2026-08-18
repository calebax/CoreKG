package graph

import (
	"context"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/ygpkg/yg-go/logs"
)

// SearchGraphWithChunkIDs 根据chunkIDs查询图谱
func SearchGraphWithChunkIDs(ctx context.Context, graphInfo *foresttype.ForestGraphInfo, chunkIDs []string) (*nebulagraph.Graph, error) {
	rcDao := NewKeGraphResourceChunkDao()

	// 找到所以chunk连接的节点
	refChunks, err := rcDao.GetListByChunkIDs(ctx, graphInfo.ID, graphInfo.VersionID, chunkIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "SearchGraphWithChunkIDs GetListByChunkIDs err: %v")
		return nil, err
	}
	nodeids := []uint{}
	// 找到所有节点
	for _, v := range refChunks {
		if v.ResourceType == foresttype.KeGraphResourceChunkTypeEdge {
			continue
		}
		nodeids = append(nodeids, v.ResourceID)
	}
	nodeNames, err := GetNodeNameByNodeIDs(ctx, nodeids)
	if err != nil {
		logs.ErrorContextf(ctx, "SearchGraphWithChunkIDs GetNodeNameByNodeIDs err: %v")
		return nil, err
	}
	cli, err := nebulagraph.NewNebulaCLI(ctx, graphInfo.SpaceName)
	if err != nil {
		logs.ErrorContextf(ctx, "SearchGraphWithChunkIDs NewNebulaCLI error: %v", err)
		return nil, err
	}
	defer cli.Release()

	chunkGraph, err := cli.GetNodesGraph(nodeNames)
	if err != nil {
		logs.ErrorContextf(ctx, "SearchGraphWithChunkIDs GetNodesGraph error: %v", err)
		return nil, err
	}

	return chunkGraph, nil
}
