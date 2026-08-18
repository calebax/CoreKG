package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/insmtx/corekg/apps/kesearch/models/chunk"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// DeleteFileGraph 删除文件图谱
func DeleteFileGraph(ctx context.Context, tx *gorm.DB, graph *foresttype.ForestGraphInfo, fileInfo *foresttype.KnownowForestFile) error {
	// 1. 获取 chunk 列表（ID 为 string）
	chunkList, err := chunk.ListChunksByFileID(ctx, fileInfo.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteFileGraph ListChunksByFileID error: %v", err)
		return err
	}
	if len(chunkList) == 0 {
		return nil
	}

	// 转 string slice
	chunkIDs := make([]string, 0, len(chunkList))
	for _, ch := range chunkList {
		chunkIDs = append(chunkIDs, ch.ID) // 注意这里 ch.ID 是 string
	}

	dao := NewKeGraphResourceChunkDao().WithTx(tx)

	// 2. 一次性查所有资源
	resources, err := dao.GetListByChunkIDs(ctx, graph.ID, graph.VersionID, chunkIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteFileGraph GetListByChunkIDs error: %v", err)
		return err
	}
	if len(resources) == 0 {
		return nil
	}

	// 3. 资源分组
	nodeRes := map[uint][]*foresttype.KeGraphResourceChunk{}
	edgeRes := map[uint][]*foresttype.KeGraphResourceChunk{}

	for _, r := range resources {
		if r.ResourceType == foresttype.KeGraphResourceChunkTypeNode {
			nodeRes[r.ResourceID] = append(nodeRes[r.ResourceID], &r)
		} else {
			edgeRes[r.ResourceID] = append(edgeRes[r.ResourceID], &r)
		}
	}

	// 删除收集器
	delResIDs := make([]uint, 0)
	delNodeIDs := make([]uint, 0)
	delEdgeIDs := make([]uint, 0)

	nebulaNodes := make([]string, 0)
	nebulaEdges := []struct {
		Type string
		Src  string
		Dst  string
	}{}

	// =============== 处理 Node ===============
	for nodeID, rlist := range nodeRes {
		allList, err := dao.GetListByCond(ctx, &KeGraphResourceChunkCond{
			Filters: []apiobj.Filter{
				{
					Field: "resource_type",
					Value: []string{string(foresttype.KeGraphResourceChunkTypeNode)},
				},
				{
					Field: "resource_id",
					Value: []string{fmt.Sprintf("%d", nodeID)},
				},
			},
		})
		if err != nil {
			return err
		}

		if len(allList) == len(rlist) {
			// 独占引用 → 删除 node
			delNodeIDs = append(delNodeIDs, nodeID)

			node, err := GetTNodeByID(ctx, nodeID)
			if err == nil {
				nebulaNodes = append(nebulaNodes, node.Name)
			}

			for _, r := range rlist {
				delResIDs = append(delResIDs, r.ID)
			}
		} else {
			// 多引用 → 仅删引用
			for _, r := range rlist {
				delResIDs = append(delResIDs, r.ID)
			}
		}
	}

	// =============== 处理 Edge ===============
	for edgeID, rlist := range edgeRes {
		allList, err := dao.GetListByCond(ctx, &KeGraphResourceChunkCond{
			Filters: []apiobj.Filter{
				{
					Field: "resource_type",
					Value: []string{string(foresttype.KeGraphResourceChunkTypeEdge)},
				},
				{
					Field: "resource_id",
					Value: []string{fmt.Sprintf("%d", edgeID)},
				},
			},
		})
		if err != nil {
			return err
		}

		if len(allList) == len(rlist) {
			// 只有该文件引用 → 删除 edge
			edgeInfo, err := GetEdgeInfo(ctx, edgeID)
			if err == nil {
				if edgeInfo == nil {
					continue
				}
				nebulaEdges = append(nebulaEdges, struct {
					Type string
					Src  string
					Dst  string
				}{
					Type: edgeInfo.EdgeTypeName,
					Src:  edgeInfo.SrcNodeName,
					Dst:  edgeInfo.DstNodeName,
				})
			}

			delEdgeIDs = append(delEdgeIDs, edgeID)

			for _, r := range rlist {
				delResIDs = append(delResIDs, r.ID)
			}
		} else {
			// 仅删引用
			for _, r := range rlist {
				delResIDs = append(delResIDs, r.ID)
			}
		}
	}

	// =============== 批量删除资源 ===============
	if len(delResIDs) > 0 {
		if err := dao.BatchDelete(ctx, delResIDs); err != nil {
			return err
		}
	}

	// =============== 批量删除节点 ===============
	err = DeleteTNodeByIDs(ctx, tx, delNodeIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteFileGraph DeleteTNodeByIDs error: %v", err)
		return err
	}
	err = DeleteEdgeWithNodeIDs(ctx, tx, delNodeIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteFileGraph DeleteEdgeWithNodeIDs error: %v", err)
		return err
	}

	// =============== 批量删除 edges ===============
	err = DeleteEdgeByIDs(ctx, tx, delEdgeIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteFileGraph DeleteEdgeByIDs error: %v", err)
		return err
	}

	// =============== Nebula 批量删除 ===============
	cli, err := nebulagraph.NewNebulaCLI(ctx, graph.SpaceName)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteFileGraph NewNebulaCLI error: %v", err)
		return err
	}
	defer cli.Release()
	err = cli.DeleteNodes(nebulaNodes)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteFileGraph DeleteNodes error: %v", err)
		return err
	}

	for _, e := range nebulaEdges {
		_ = cli.DeleteEdge(e.Type, e.Src, e.Dst)
	}

	return nil
}

// 定义 ES 响应结构体，方便解析
type esSearchResponse struct {
	Hits struct {
		Hits []struct {
			Source struct {
				FileID uint `json:"file_id"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func GetFileIDsByChunkIDs(ctx context.Context, chunkIDs []string) ([]uint, error) {
	// 1. 边界检查
	if len(chunkIDs) == 0 {
		return []uint{}, nil
	}

	// 2. 初始化客户端
	escli, err := essearch.InitESClient(ctx)
	if err != nil {
		return nil, err
	}

	// 3. 构建查询体 (DSL)
	// 相当于 SQL: SELECT file_id FROM ke_0 WHERE _id IN (chunkIDs)
	queryBody := map[string]interface{}{
		"query": map[string]interface{}{
			"ids": map[string]interface{}{
				"values": chunkIDs, // 匹配 _id 字段
			},
		},
		"_source": []string{"file_id"}, // 性能优化：只返回 file_id 字段
		"size":    len(chunkIDs),       // 确保返回所有匹配的文档，而不是默认的 10 条
	}

	// 序列化请求体
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(queryBody); err != nil {
		logs.ErrorContextf(ctx, "encode query body failed: %s", err)
		return nil, err
	}

	// 4. 执行搜索
	res, err := escli.Search(
		escli.Search.WithContext(ctx),
		escli.Search.WithIndex("ke_0"),
		escli.Search.WithBody(&buf),
		escli.Search.WithTrackTotalHits(false),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es search request failed: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	// 5. 检查响应状态
	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			logs.ErrorContextf(ctx, "parse es error response failed: %s", err)
			return nil, err
		}
		logs.ErrorContextf(ctx, "es search returned error: %v", e)
		return nil, fmt.Errorf("es search error: %s", res.Status())
	}

	// 6. 解析响应数据
	var result esSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		logs.ErrorContextf(ctx, "decode response body failed: %s", err)
		return nil, err
	}

	// 7. 提取 file_id 并去重
	// 使用 map 去重
	uniqueFileIDs := make(map[uint]struct{})
	for _, hit := range result.Hits.Hits {
		uniqueFileIDs[hit.Source.FileID] = struct{}{}
	}

	// 8. 构造返回结果
	files := make([]uint, 0, len(uniqueFileIDs))
	for fileID := range uniqueFileIDs {
		files = append(files, fileID)
	}

	return files, nil
}
