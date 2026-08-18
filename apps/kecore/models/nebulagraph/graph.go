package nebulagraph

import (
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	nebula_go "github.com/vesoft-inc/nebula-go/v3"
	"github.com/ygpkg/yg-go/logs"
)

// GetKnowledgeGraph 获取知识图谱
func (cli *NebulaCli) GetKnowledgeGraph(req KnowledgeGraphReq) (*Graph, error) {
	srcTag := "(v)"
	if req.SrcTag != "" {
		srcTag = fmt.Sprintf("(v:`%s`)", req.SrcTag)
	}
	dstTag := "(a)"
	if req.DstTag != "" {
		dstTag = fmt.Sprintf("(a:`%s`)", req.DstTag)
	}
	if !req.IsTwoWay {
		dstTag = ">" + dstTag
	}
	// 处理 where 条件
	var whereClauses []string
	if req.SrcName != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("id(v)==\"%s\"", req.SrcName))
	}
	if req.DstName != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("id(a)==\"%s\"", req.DstName))
	}

	where := ""
	if len(whereClauses) > 0 {
		where = "WHERE " + strings.Join(whereClauses, " AND ")
	}
	ngql := fmt.Sprintf("MATCH %s-[e]-%s %s RETURN v,e,a Limit %d", srcTag, dstTag, where, req.Limit)
	logs.InfoContextf(cli.ctx, "GetKnowledgeGraph: %s", ngql)
	res, err := cli.ExecuteAndCheck(ngql)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "GetKnowledgeGraph error: %v", err)
		return nil, err
	}
	nodeMap := map[string]*NodeInfo{}
	edges := []EdgeInfo{}
	for i := 0; i < res.GetRowSize(); i++ {
		record, err := res.GetRowValuesByIndex(i)
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetKnowledgeGraph GetRowValuesByIndex error: %v", err)
			return nil, err
		}
		edge, err := cli.getRowEdge(record, "e")
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetKnowledgeGraph getRowEdge error: %v", err)
			return nil, err
		}
		edges = append(edges, *edge)

		nodev, err := cli.getRowNode(record, "v")
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetKnowledgeGraph getRowNode error: %v", err)
			return nil, err
		}
		nodeMap[nodev.Name] = nodev

		nodea, err := cli.getRowNode(record, "a")
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetKnowledgeGraph getRowNode error: %v", err)
			return nil, err
		}
		nodeMap[nodea.Name] = nodea
	}
	nodes := []NodeInfo{}
	for _, n := range nodeMap {
		nodes = append(nodes, *n)
	}
	if _, ok := nodeMap[req.SrcName]; !ok {
		// nodes = append(nodes, *nodeMap[req.SrcName])
		n, err := cli.GetTagNodeInfo(req.SrcName, req.SrcTag)
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetKnowledgeGraph GetTagNodeInfo error: %v", err)
			return nil, err
		}
		if n != nil {
			nodes = append(nodes, *n)
		}
	}
	return &Graph{
		Edges: edges,
		Nodes: nodes,
	}, nil
}

func (cli *NebulaCli) getRowNode(record *nebula_go.Record, colName string) (*NodeInfo, error) {
	nodeData, err := record.GetValueByColName(colName)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "getRowNode GetValueByColName col:%s error: %v", colName, err)
		return nil, err
	}
	node, err := nodeData.AsNode()
	if err != nil {
		logs.ErrorContextf(cli.ctx, "getRowNode AsNode error: %v", err)
		return nil, err
	}
	name, err := node.GetID().AsString()
	if err != nil {
		logs.ErrorContextf(cli.ctx, "getRowNode node.GetID().AsString() error: %v", err)
		return nil, err
	}
	nodeInfo := &NodeInfo{
		Name: name,
	}
	for _, t := range node.GetTags() {
		pros, err := node.Properties(t)
		if err != nil {
			logs.ErrorContextf(cli.ctx, "getRowNode Properties error: %v", err)
			return nil, err
		}
		tag := Tag{
			Name: t,
		}
		for name, v := range pros {
			value := v.String()
			if value == "" || value == `""` || value == "__NULL__" {
				continue
			}
			tag.PropertiesValues = append(tag.PropertiesValues, &foresttype.PropertyValue{
				Name:  name,
				Value: value,
			})
		}
		nodeInfo.Tags = append(nodeInfo.Tags, tag)
	}
	return nodeInfo, nil
}

func (cli *NebulaCli) getRowEdge(record *nebula_go.Record, colName string) (*EdgeInfo, error) {
	edge, err := record.GetValueByColName(colName)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "getRowNode GetValueByColName col:%s error: %v", colName, err)
		return nil, err
	}
	relationship, err := edge.AsRelationship()
	if err != nil {
		logs.ErrorContextf(cli.ctx, "GetKnowledgeGraph AsRelationship error: %v", err)
		return nil, err
	}
	src, err := relationship.GetSrcVertexID().AsString()
	if err != nil {
		logs.ErrorContextf(cli.ctx, "GetKnowledgeGraph src GetSrcVertexID AsString error: %v", err)
		return nil, err
	}
	dst, err := relationship.GetDstVertexID().AsString()
	if err != nil {
		logs.ErrorContextf(cli.ctx, "GetKnowledgeGraph dst GetSrcVertexID AsString error: %v", err)
		return nil, err
	}
	return &EdgeInfo{
		Src:  src,
		Dst:  dst,
		Name: relationship.GetEdgeName(),
	}, nil
}

type KnowledgeGraphReq struct {
	SrcTag   string `json:"src_tag"`
	DstTag   string `json:"dst_tag"`
	SrcName  string `json:"src_name"`
	DstName  string `json:"dst_name"`
	Limit    uint   `json:"limit"`
	IsTwoWay bool   `json:"is_two_way"`
}

type NodeInfo struct {
	Name string `json:"name"`
	Tags []Tag  `json:"tags"`
}

type Tag struct {
	Name             string                      `json:"name"`
	PropertiesValues foresttype.PropertiesValues `json:"properties_values"`
}

type EdgeInfo struct {
	Src  string `json:"src"`
	Dst  string `json:"dst"`
	Name string `json:"name"`
}

type Graph struct {
	Nodes []NodeInfo `json:"nodes"`
	Edges []EdgeInfo `json:"edges"`
}

// FindNodePath 查询节点间路径
func (cli *NebulaCli) FindNodePath(src, dst string, maxStep uint) error {
	ngql := fmt.Sprintf("FIND ALL PATH FROM \"%s\" TO \"%s\" OVER * UPTO %d STEPS YIELD path as p;", src, dst, maxStep)
	logs.InfoContextf(cli.ctx, "FindNodePath: %s", ngql)
	res, err := cli.ExecuteAndCheck(ngql)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "FindNodePath error: %v", err)
		return err
	}
	for i := 0; i < res.GetRowSize(); i++ {
		record, err := res.GetRowValuesByIndex(i)
		if err != nil {
			logs.ErrorContextf(cli.ctx, "FindNodePath GetRowValuesByIndex error: %v", err)
			return err
		}
		path, err := record.GetValueByColName("p")
		if err != nil {
			logs.ErrorContextf(cli.ctx, "getRowNode GetValueByColName col:%s error: %v", "p", err)
			return err
		}
		pathInfo, err := path.AsPath()
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetKnowledgeGraph AsPath error: %v", err)
			return err
		}
		println(pathInfo.String(), pathInfo.GetPathLength())
		// println(pathInfo.GetNodes())
		for _, v := range pathInfo.GetNodes() {
			println(v.String())
		}
	}
	return nil
}

// GetNodesGraph 获取节点图
func (cli *NebulaCli) GetNodesGraph(nodeIDs []string) (*Graph, error) {
	if len(nodeIDs) == 0 {
		return nil, nil
	}
	nodeIDs = removeDuplicates(nodeIDs)
	formattedNodeIDs := strings.Join(nodeIDs, `","`)
	searchStr := "GO 1 STEPS FROM \"" + formattedNodeIDs + "\" OVER * BIDIRECT YIELD src(edge) AS src_id, dst(edge) AS dst_id, edge AS e;"
	logs.InfoContextf(cli.ctx, "GetNodesGraph: %s", searchStr)
	res, err := cli.ExecuteAndCheck(searchStr)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "GetNodesGraph error: %v", err)
		return nil, err
	}
	edgeList := []EdgeInfo{}
	for i := 0; i < res.GetRowSize(); i++ {
		record, err := res.GetRowValuesByIndex(i)
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetNodesInfo GetRowValuesByIndex error: %v", err)
			return nil, err
		}
		edge, err := cli.getRowEdge(record, "e")
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetNodesInfo getRowNode error: %v", err)
			return nil, err
		}
		edgeList = append(edgeList, *edge)
		nodeIDs = append(nodeIDs, edge.Src, edge.Dst)
	}
	nodeIDs = removeDuplicates(nodeIDs)
	nodeList, err := cli.GetNodesInfo(nodeIDs)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "GetNodesInfo GetNodesInfo error: %v", err)
		return nil, err
	}

	return &Graph{
		Edges: edgeList,
		Nodes: nodeList,
	}, nil
}

// GetNodesGraph 获取节点图
func (cli *NebulaCli) GetNodesGraphWithStep(nodeIDs []string, step int) (*Graph, error) {
	if len(nodeIDs) == 0 {
		return nil, nil
	}
	if step <= 0 {
		step = 1
	}
	nodeIDs = removeDuplicates(nodeIDs)
	formattedNodeIDs := strings.Join(nodeIDs, `","`)
	searchStr := "GO " + fmt.Sprintf("%d", step) + " STEPS FROM \"" + formattedNodeIDs + "\" OVER * BIDIRECT YIELD src(edge) AS src_id, dst(edge) AS dst_id, edge AS e;"
	logs.InfoContextf(cli.ctx, "GetNodesGraph: %s", searchStr)
	res, err := cli.ExecuteAndCheck(searchStr)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "GetNodesGraph error: %v", err)
		return nil, err
	}
	edgeList := []EdgeInfo{}
	for i := 0; i < res.GetRowSize(); i++ {
		record, err := res.GetRowValuesByIndex(i)
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetNodesInfo GetRowValuesByIndex error: %v", err)
			return nil, err
		}
		edge, err := cli.getRowEdge(record, "e")
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetNodesInfo getRowNode error: %v", err)
			return nil, err
		}
		edgeList = append(edgeList, *edge)
		nodeIDs = append(nodeIDs, edge.Src, edge.Dst)
	}
	nodeIDs = removeDuplicates(nodeIDs)
	nodeList, err := cli.GetNodesInfo(nodeIDs)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "GetNodesInfo GetNodesInfo error: %v", err)
		return nil, err
	}

	return &Graph{
		Edges: edgeList,
		Nodes: nodeList,
	}, nil
}

// removeDuplicates 去除字符串切片中的重复元素，保持原有顺序
func removeDuplicates(slice []string) []string {
	// 使用 map 来记录已经出现过的元素
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice)) // 预分配容量以提高效率
	for _, item := range slice {
		// 如果元素未在 map 中出现过
		if !seen[item] {
			seen[item] = true             // 标记为已出现
			result = append(result, item) // 添加到结果切片
		}
	}
	return result
}
