package nebulagraph

import (
	"fmt"
	"strings"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	nebula_go "github.com/vesoft-inc/nebula-go/v3"
	"github.com/ygpkg/yg-go/logs"
)

// InsertNode 插入节点
func (cli *NebulaCli) InsertNode(node *foresttype.TagNodeInfo) error {
	str := node.InsertStr()
	logs.InfoContextf(cli.ctx, "insert node:%s", str)
	res, err := cli.ExecuteWithRetry(str)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to insert node:%v, res:%s", err, str)
		return err
	}
	logs.InfoContextf(cli.ctx, "insert node success:%s", logs.JSON(res))
	return nil
}

// InsertEdge 插入边
func (cli *NebulaCli) InsertEdge(edge *foresttype.EdgeInfo) error {
	str := edge.InsertStr()
	logs.InfoContextf(cli.ctx, "insert edge:%s", str)
	res, err := cli.ExecuteWithRetry(str)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to insert edge:%v, res:%s", err, logs.JSON(res))
		return err
	}
	logs.InfoContextf(cli.ctx, "insert edge success:%+v", res)
	return nil
}

// InsertEdges 批量插入边
// 按边类型分组，相同类型的边批量插入，不同类型的边分别插入
func (cli *NebulaCli) InsertEdges(edges []*foresttype.EdgeInfo) error {
	if len(edges) == 0 {
		return nil
	}
	if len(edges) == 1 {
		return cli.InsertEdge(edges[0])
	}

	// 按边类型分组
	edgesByType := make(map[string][]*foresttype.EdgeInfo)
	for _, edge := range edges {
		edgeType := edge.EdgeTypeName
		edgesByType[edgeType] = append(edgesByType[edgeType], edge)
	}

	// 对每种类型的边进行批量插入
	for edgeType, typeEdges := range edgesByType {
		str := foresttype.BatchInsertStr(typeEdges)
		logs.InfoContextf(cli.ctx, "batch insert edges (type:%s):%s", edgeType, str)
		res, err := cli.ExecuteWithRetry(str)
		if err != nil {
			logs.ErrorContextf(cli.ctx, "fail to batch insert edges (type:%s):%v, res:%s", edgeType, err, logs.JSON(res))
			return err
		}
		logs.InfoContextf(cli.ctx, "batch insert edges (type:%s) success:%+v", edgeType, res)
	}

	return nil
}

// GetNodeInfo 获取节点详情
func (cli *NebulaCli) GetNodeInfo(nodeID string) (*NodeInfo, error) {
	searchStr := "MATCH (v) WHERE id(v) == \"" + nodeID + "\" RETURN v;"
	logs.InfoContextf(cli.ctx, "GetNodeInfo: %s", searchStr)
	res, err := cli.ExecuteAndCheck(searchStr)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to insert edge:%v, res:%s", err, logs.JSON(res))
		return nil, err
	}
	if res.IsEmpty() {
		return nil, fmt.Errorf("empty result")
	}
	nodev := &NodeInfo{}
	for i := 0; i < res.GetRowSize(); i++ {
		record, err := res.GetRowValuesByIndex(i)
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetNodeInfo GetRowValuesByIndex error: %v", err)
			return nil, err
		}
		nodev, err = cli.getRowNode(record, "v")
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetNodeInfo getRowNode error: %v", err)
			return nil, err
		}
	}
	return nodev, nil
}

// GetTagNodeInfo 获取节点详情 null 不报错
func (cli *NebulaCli) GetTagNodeInfo(nodeID, tag string) (*NodeInfo, error) {
	v := "(v)"
	if tag != "" {
		v = fmt.Sprintf("(v:`%s`)", tag)
	}
	searchStr := "MATCH " + v + " WHERE id(v) == \"" + nodeID + "\" RETURN v;"
	logs.InfoContextf(cli.ctx, "GetNodeInfo: %s", searchStr)
	res, err := cli.ExecuteAndCheck(searchStr)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to insert edge:%v, res:%s", err, logs.JSON(res))
		return nil, err
	}
	if res.IsEmpty() {
		return nil, nil
	}
	nodev := &NodeInfo{}
	for i := 0; i < res.GetRowSize(); i++ {
		record, err := res.GetRowValuesByIndex(i)
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetTagNodeInfo GetRowValuesByIndex error: %v", err)
			return nil, err
		}
		nodev, err = cli.getRowNode(record, "v")
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetTagNodeInfo getRowNode error: %v", err)
			return nil, err
		}
	}
	return nodev, nil
}

// GetNodesInfo 获取节点详情
func (cli *NebulaCli) GetNodesInfo(nodeIDs []string) ([]NodeInfo, error) {
	formattedNodeIDs := fmt.Sprintf(`["%s"]`, strings.Join(nodeIDs, `","`))
	searchStr := "MATCH (v) WHERE id(v) IN " + formattedNodeIDs + " RETURN v;"
	logs.InfoContextf(cli.ctx, "GetNodeInfo: %s", searchStr)
	res, err := cli.ExecuteAndCheck(searchStr)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to insert edge:%v, res:%s", err, logs.JSON(res))
		return nil, err
	}
	if res.IsEmpty() {
		return nil, fmt.Errorf("empty result")
	}
	nodeList := []NodeInfo{}
	for i := 0; i < res.GetRowSize(); i++ {
		record, err := res.GetRowValuesByIndex(i)
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetNodesInfo GetRowValuesByIndex error: %v", err)
			return nil, err
		}
		nodev, err := cli.getRowNode(record, "v")
		if err != nil {
			logs.ErrorContextf(cli.ctx, "GetNodesInfo getRowNode error: %v", err)
			return nil, err
		}
		nodeList = append(nodeList, *nodev)
	}
	return nodeList, nil
}

func (cli *NebulaCli) DeleteNode(nodeID string) error {
	str := fmt.Sprintf("DELETE VERTEX \"%s\" WITH EDGE;", nodeID)
	_, err := cli.ExecuteAndCheck(str)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to insert node:%v, res:%s", err, str)
		return err
	}
	return nil
}

// DeleteTag 删除节点的指定tag（保留其他tag和vertex本身）
// DELETE TAG tag_name FROM vertex_id
func (cli *NebulaCli) DeleteTag(nodeID, tagName string) error {
	bt := "`"
	str := fmt.Sprintf("DELETE TAG %s%s%s FROM \"%s\";", bt, tagName, bt, nodeID)
	logs.InfoContextf(cli.ctx, "DeleteTag nql: %s", str)
	_, err := cli.ExecuteAndCheck(str)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to delete tag:%v, res:%s", err, str)
		return err
	}
	return nil
}

func (cli *NebulaCli) DeleteNodes(nodeIDs []string) error {
	node := strings.Join(nodeIDs, "\",\"")
	str := fmt.Sprintf("DELETE VERTEX \"%s\" WITH EDGE;", node)
	_, err := cli.ExecuteAndCheck(str)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to insert node:%v, res:%s", err, str)
		return err
	}
	return nil
}

func (cli *NebulaCli) DeleteEdge(edgeType, src, dst string) error {
	// 技巧：把反引号定义为变量，这样就不会跟 Go 的字符串语法打架了
	bt := "`"

	// 目标 SQL: DELETE EDGE `11111` "555555"->"测试节点5";
	// 这样写绝对不会错：
	str := fmt.Sprintf("DELETE EDGE %s%s%s \"%s\"->\"%s\";",
		bt, edgeType, bt, // 对应 `edgeType`
		src, // 对应 "src" (注意前面的 \" 是 Go 的转义，不是 SQL 的)
		dst, // 对应 "dst"
	)

	_, err := cli.Execute(str)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to delete edge: %v, sql: %s", err, str)
		return err
	}
	return nil
}

const (
	// DefaultMaxRetries 最大重试次数
	DefaultMaxRetries = 50

	// DefaultRetryInterval 每次重试的固定间隔
	DefaultRetryInterval = 2 * time.Second
)

// isSchemaSyncError 判断是否为 Schema 同步延迟导致的错误
// 当 Storage 还没收到 Meta 的 Schema 更新时，会报这些错
func isSchemaSyncError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// 覆盖 Edge、Tag 和 Space 的未找到错误
	return strings.Contains(msg, "EdgeNotFound") ||
		strings.Contains(msg, "TagNotFound") ||
		strings.Contains(msg, "SpaceNotFound") ||
		strings.Contains(msg, "not found") ||
		strings.Contains(msg, "Unknown column")
}

// Execute 执行 nGQL 语句，内置 Schema 同步等待机制
// 如果遇到 EdgeNotFound/TagNotFound 等错误，会自动重试
func (cli *NebulaCli) ExecuteWithRetry(stmt string) (*nebula_go.ResultSet, error) {
	var (
		res *nebula_go.ResultSet
		err error
	)

	for i := range DefaultMaxRetries {
		// 调用基础执行方法
		res, err = cli.ExecuteAndCheck(stmt)

		// 1. 执行成功，直接返回
		if err == nil {
			// 如果经历过重试，打印一条 Info 记录一下
			if i > 0 {
				logs.InfoContextf(cli.ctx, "Execute success after %d retries: %s", i, stmt)
			}
			return res, nil
		}

		// 2. 检查是否为 Schema 同步延迟错误
		if isSchemaSyncError(err) {
			// 打印 Debug/Warn 日志，说明正在等待同步
			logs.WarnContextf(cli.ctx, "Schema syncing (Storage not ready), retry %d/%d in %v... Error: %v",
				i+1, DefaultMaxRetries, DefaultRetryInterval, err)

			// 阻塞等待
			time.Sleep(DefaultRetryInterval)
			continue
		}

		// 3. 如果是语法错误(SyntaxError)或其他逻辑错误，直接返回，不重试
		logs.ErrorContextf(cli.ctx, "Execute failed (non-retriable): %v, stmt: %s", err, stmt)
		return nil, err
	}

	// 4. 超过最大重试次数
	return nil, fmt.Errorf("execute failed after %d retries, last error: %v", DefaultMaxRetries, err)
}
