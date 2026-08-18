package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/ygpkg/yg-go/logs"
)

// GetTitleList nebula 搜标题
func GetTitleList(ctx context.Context, cli *nebulagraph.NebulaCli, searchType string) ([]string, error) {
	res, err := cli.ExecuteAndCheck(fmt.Sprintf(`GO 0 TO 5 STEPS FROM "%s" OVER 包含 BIDIRECT YIELD DISTINCT dst(edge)`, searchType))
	if err != nil {
		logs.ErrorContextf(ctx, "GetTitleList search error: %v", err)
		return nil, err
	}
	values, err := res.GetValuesByColName("dst(EDGE)")
	if err != nil {
		logs.ErrorContextf(ctx, "GetTitleList GetValuesByColName error: %v", err)
		return nil, err
	}
	var titles []string
	for _, v := range values {
		str, err := v.AsString()
		if err != nil {
			logs.ErrorContextf(ctx, "GetTitleList AsString error: %v", err)
			return nil, err
		}
		titles = append(titles, str)
	}
	return titles, nil
}

type DirectoryInfo struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Page   int    `json:"page"`
	FileID uint   `json:"file_id"`
}

// GetNodesInfo 获取节点信息
func GetNodesInfo(ctx context.Context, cli *nebulagraph.NebulaCli, titles []string) ([]DirectoryInfo, error) {
	tagname := "目录"
	formattedNodeIDs := fmt.Sprintf(`["%s"]`, strings.Join(titles, `","`))
	searchStr := "MATCH (v:`" + tagname + "`) WHERE id(v) IN " + formattedNodeIDs + " RETURN v;"
	logs.InfoContextf(ctx, "GetNodeInfo: %s", searchStr)
	res, err := cli.ExecuteAndCheck(searchStr)
	if err != nil {
		logs.ErrorContextf(ctx, "fail to insert edge:%v, res:%s", err, logs.JSON(res))
		return nil, err
	}
	if res.IsEmpty() {
		return nil, fmt.Errorf("empty result")
	}
	values, err := res.GetValuesByColName("v")
	nodeList := []DirectoryInfo{}
	for _, v := range values {
		// println(i)
		node, err := v.AsNode()
		if err != nil {
			logs.ErrorContextf(ctx, "GetNodesInfo AsNode error: %v", err)
			return nil, err
		}
		name, err := node.GetID().AsString()
		if err != nil {
			logs.ErrorContextf(ctx, "GetNodesInfo node.GetID().AsString() error: %v", err)
			return nil, err
		}
		pros, err := node.Properties(tagname)
		if err != nil {
			logs.ErrorContextf(ctx, "GetNodesInfo node.Properties(tagname) error: %v", err)
			return nil, err
		}
		page, err := pros["page"].AsInt()
		if err != nil {
			logs.ErrorContextf(ctx, "GetNodesInfo pros[\"page\"].AsInt() error: %v", err)
			return nil, err
		}
		fileID, err := pros["file_id"].AsInt()
		if err != nil {
			logs.ErrorContextf(ctx, "GetNodesInfo pros[\"file_id\"].AsString() error: %v", err)
			return nil, err
		}
		nodeList = append(nodeList, DirectoryInfo{
			ID:     name,
			Title:  name,
			Page:   int(page),
			FileID: uint(fileID),
		})
	}

	return nodeList, nil
}
