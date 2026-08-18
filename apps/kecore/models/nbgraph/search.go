package nbgraph

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
	nebula_go "github.com/vesoft-inc/nebula-go/v3"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

var (
	ErrCliIsNil       = errors.New("client is nil")
	ErrNewCliNil      = errors.New("new client err")
	ErrCountConfIsNil = errors.New("count config is nil")
	ErrEmptyResult    = errors.New("empty result")
	ErrValidParam     = errors.New("invalid param error")
	ErrResultSetIsNil = errors.New("result set is nil")
)

type WrapperNebulaSearch struct {
	//conf area
	Cli       *NebulaCli
	CountConf *NebulaCountConf

	//param area
	ctx      context.Context
	Space    string
	ForestID uint

	//cache map
	VID2FIDs  map[string][]uint
	FID2FName map[uint]string

	//result area
	NodeIDs   []string
	Wc        []WordsCloud
	ResultSet *nebula_go.ResultSet
	G         *Graph
}

func (w *WrapperNebulaSearch) Close() error {
	if w.Cli != nil {
		w.Cli.Release()
	}
	return nil
}

func NewWrapper(ctx context.Context, forestID uint) (*WrapperNebulaSearch, error) {
	if ctx == nil || forestID <= 0 {
		logs.ErrorContextf(ctx, "NewWrapper with invalid param [ctx:%v],[forestID:%v]", ctx, forestID)
		return nil, ErrValidParam
	}

	conf := &NebulaCountConf{}
	err := settings.GetYaml("knowledge", "nebulacount", conf)
	if err != nil {
		logs.ErrorContextf(ctx, "NewWrapper.Get NebulaCountConf Yaml err: %v", err)
		conf = &NebulaCountConf{
			GraphCount:     3,
			WordCloudCount: 50,
		}
	}

	f, err := forest.GetForestByID(ctx, forestID)
	if err != nil {
		logs.ErrorContextf(ctx, "NewWrapper.Get forestID:%v, err: %v", forestID, err)
		return nil, err
	}

	cli, err := NewNebulaCLI()
	if err != nil {
		logs.ErrorContextf(ctx, "NewWrapper.NewNebulaCLI err: %v", err)
		return nil, ErrNewCliNil
	}
	return &WrapperNebulaSearch{
		Cli:       cli,
		CountConf: conf,
		Space:     f.EsIndex(),
		ctx:       ctx,
		ForestID:  forestID,
		VID2FIDs:  make(map[string][]uint),
		FID2FName: make(map[uint]string),
		G: &Graph{
			Nodes: make([]*Node, 0),
			Edges: make([]Edge, 0),
		},
	}, nil
}

func (w *WrapperNebulaSearch) DoWordCloudNql() error {
	if w.Cli == nil {
		logs.WarnContextf(w.ctx, "WrapperNebulaSearch.GetWordCloud cli is nil")
		return ErrCliIsNil
	}

	var (
		resp *nebula_go.ResultSet
		nql  string
	)

	if w.ForestID == 109 {
		nql = fmt.Sprintf("USE know_19_129; MATCH (v)-[e]-() RETURN id(v) AS vertex_id, COUNT(e) AS degree ORDER BY degree DESC;")
	} else {
		nql = fmt.Sprintf("USE %s; MATCH (v)-[e]-() "+
			"WHERE v.entities.forest_id==%v AND e.forest_id == %v AND NOT v.entities.type IN %v "+
			"RETURN v.entities.node_id AS vertex_id, COUNT(e) AS degree,id(v) AS id ORDER BY degree DESC LIMIT %v;", w.Space, w.ForestID, w.ForestID, "[\"chunk\",\"table\",\"image\",\"title\",\"formula\"]", w.CountConf.WordCloudCount)
	}

	resp, err := w.Cli.ExecuteAndCheck(nql)
	if err != nil {
		logs.ErrorContextf(w.ctx, "WrapperNebulaSearch.GetWordCloud Execute nql err: %v with nql: %v", err, nql)
		return err
	}

	if resp.IsEmpty() {
		logs.DebugContextf(w.ctx, "WrapperNebulaSearch.GetWordCloud: empty response wtih nql: %s", nql)
		return ErrEmptyResult
	}
	w.ResultSet = resp
	return nil
}

func (w *WrapperNebulaSearch) DoSubLimitNql() error {
	if w.Cli == nil {
		return ErrCliIsNil
	}
	var queryStrs []string
	for _, nodeID := range w.NodeIDs {
		str := " (GO 1 STEPS FROM \"" + nodeID + "\" OVER * BIDIRECT" +
			" WHERE properties($$).forest_id == " + fmt.Sprint(w.ForestID) +
			" AND properties(EDGE).forest_id == " + fmt.Sprint(w.ForestID) +
			" AND NOT properties($$).type IN [\"chunk\",\"table\",\"image\",\"title\",\"formula\"] " +
			fmt.Sprintf(" YIELD src(edge) AS src_id, dst(edge) AS dst_id, edge AS e | LIMIT %v)", w.CountConf.GraphCount)
		queryStrs = append(queryStrs, str)
	}

	nql := "USE " + w.Space + ";" + strings.Join(queryStrs, " UNION ALL ")

	resp, err := w.Cli.Execute(nql)
	if err != nil {
		logs.ErrorContextf(w.ctx, "WrapperNebulaSearch.DoSubLimitNql Execute nql err: %v with nql: %v", err, nql)
		return err
	}

	if resp.IsEmpty() {
		logs.DebugContextf(w.ctx, "WrapperNebulaSearch.DoSubLimitNql: empty response wtih nql: %s", nql)
		return ErrEmptyResult
	}
	w.ResultSet = resp
	return nil
}

func (w *WrapperNebulaSearch) DoGoFromIDNql(nodeID string) error {
	if w.Cli == nil {
		logs.WarnContextf(w.ctx, "WrapperNebulaSearch.GetNodesByID cli is nil")
		return ErrCliIsNil
	}
	var (
		nql  string
		resp *nebula_go.ResultSet
	)

	if w.ForestID == 109 {
		nql = fmt.Sprintf("USE know_19_129" +
			";GO 1 STEPS FROM " + nodeID + " OVER * BIDIRECT YIELD src(edge) AS src_id, dst(edge) AS dst_id, edge AS e;")
	} else {
		nql = fmt.Sprintf("USE " + w.Space +
			";GO 1 STEPS FROM \"" + nodeID + "\" OVER * BIDIRECT " +
			"WHERE properties($$).forest_id == " + fmt.Sprint(w.ForestID) + " " +
			"AND properties(EDGE).forest_id == " + fmt.Sprint(w.ForestID) + " " +
			"AND NOT properties($$).type IN [\"chunk\",\"table\",\"image\",\"title\"] " +
			"YIELD src(edge) AS src_id, dst(edge) AS dst_id, edge AS e | LIMIT 200;")
	}

	resp, err := w.Cli.ExecuteAndCheck(nql)
	if err != nil {
		logs.ErrorContextf(w.ctx, "WrapperNebulaSearch.GetNodesByID Execute nql err: %v with nql: %v", err, nql)
		return err
	}

	if resp.IsEmpty() {
		logs.DebugContextf(w.ctx, "WrapperNebulaSearch.GetNodesByID: empty response wtih nql: %s", nql)
		return ErrEmptyResult
	}
	w.ResultSet = resp
	return nil
}

func (w *WrapperNebulaSearch) BuildWordCloud() error {
	if w.ResultSet == nil {
		return ErrResultSetIsNil
	}
	for _, item := range w.ResultSet.GetRows() {
		w.Wc = append(w.Wc, WordsCloud{
			Word:   string(item.Values[0].SVal),
			Weight: *item.Values[1].IVal,
			ID:     string(item.Values[2].SVal),
		})
		w.NodeIDs = append(w.NodeIDs, escapeVidForNQL(string(item.Values[2].SVal)))
	}
	return nil
}

func (w *WrapperNebulaSearch) BuildGraph() error {
	if w.ResultSet == nil {
		return ErrResultSetIsNil
	}
	for _, item := range w.ResultSet.GetRows() {
		w.G.Edges = append(w.G.Edges, Edge{
			Source:    strings.TrimPrefix(string(item.Values[0].SVal), fmt.Sprintf("%v_", w.ForestID)),
			Target:    strings.TrimPrefix(string(item.Values[1].SVal), fmt.Sprintf("%v_", w.ForestID)),
			SourceID:  string(item.Values[0].SVal),
			TargetID:  string(item.Values[1].SVal),
			CompanyID: uint(*item.Values[2].EVal.Props["company_id"].IVal),
			ForestID:  uint(*item.Values[2].EVal.Props["forest_id"].IVal),
			Uin:       uint(*item.Values[2].EVal.Props["uin"].IVal),
		})
	}
	// 遍历out.edges所有项得到其中所用到节点，其中值不重复
	var nodeSet = make(map[string]struct{})
	for _, item := range w.G.Edges {
		nodeSet[strings.ReplaceAll(item.SourceID, `\`, `\\`)] = struct{}{}
		nodeSet[strings.ReplaceAll(item.TargetID, `\`, `\\`)] = struct{}{}
	}
	w.NodeIDs = make([]string, 0, len(nodeSet))
	for nodeID := range nodeSet {
		w.NodeIDs = append(w.NodeIDs, escapeVidForNQL(nodeID))
	}
	if err := w.GetNodesInfo(); err != nil {
		return err
	}
	return nil
}

func (w *WrapperNebulaSearch) GetNodesInfo() error {
	if w.Cli == nil {
		logs.WarnContextf(w.ctx, "WrapperNebulaSearch.GetNodesInfo cli is nil")
		return ErrCliIsNil
	}
	for i, nodeID := range w.NodeIDs {
		formattedNodeIDs := escapeNebulaString(nodeID)
		w.NodeIDs[i] = formattedNodeIDs
	}
	formattedNodeIDs := fmt.Sprintf(`["%s"]`, strings.Join(w.NodeIDs, `","`))

	if w.ForestID == 109 {
		w.Space = "know_19_129"
	}

	var (
		nql = fmt.Sprintf("USE " + w.Space +
			";MATCH (v) WHERE id(v) IN " + formattedNodeIDs + " RETURN v;")
		resp *nebula_go.ResultSet
	)

	resp, err := w.Cli.ExecuteAndCheck(nql)
	if err != nil {
		logs.ErrorContextf(w.ctx, "WrapperNebulaSearch.GetNodesInfo Execute nql err: %v with nql: %v", err, nql)
		return err
	}
	if resp.IsEmpty() {
		logs.DebugContextf(w.ctx, "WrapperNebulaSearch.GetNodesInfo: empty response wtih nql: %s", nql)
		return ErrEmptyResult
	}
	var (
		isChunk = false
	)
	for _, item := range resp.GetRows() {
		var node = &Node{}
		// node.ID = string(item.Values[0].VVal.Vid.SVal)
		for _, tag := range item.Values[0].VVal.Tags {
			var t = &Tag{}
			//get type
			tp := tag.Props["type"]
			t.Type = string(tp.SVal)
			//if type is chunk then do not display
			if t.Type == "chunk" {
				isChunk = true
				break
			}
			t.TagName = string(tag.Name)
			companyId := tag.Props["company_id"]
			t.CompanyID = uint(*companyId.IVal)
			forestId := tag.Props["forest_id"]
			t.ForestID = uint(*forestId.IVal)
			uin := tag.Props["uin"]
			t.Uin = uint(*uin.IVal)

			node_id := tag.Props["node_id"]
			t.NodeID = string(node_id.SVal)
			// todo 目前返回给前端的都一样
			node.ID = string(node_id.SVal)

			//get file_id
			fID := tag.Props["file_id"]
			fIDs := string(fID.SVal)

			fileIDs := strings.Split(fIDs, "&&&")
			for _, fIDStr := range fileIDs {
				fIDUint, err := strconv.Atoi(fIDStr)
				if err != nil {
					return err
				}
				//build a cluster map
				w.VID2FIDs[node.ID] = append(w.VID2FIDs[node.ID], uint(fIDUint))
				//construct unique fids
				w.FID2FName[uint(fIDUint)] = ""
			}
			node.Tag = append(node.Tag, t)
		}
		if isChunk {
			isChunk = false
			continue
		}
		w.G.Nodes = append(w.G.Nodes, node)
	}

	//parse FileMap
	fileIDs := make([]uint, 0, len(w.FID2FName))
	for k := range w.FID2FName {
		fileIDs = append(fileIDs, k)
	}
	fs, err := forest.GetForestFileByIDs(fileIDs)
	if err != nil {
		logs.WarnContextf(w.ctx, "WrapperNebulaSearch.GetNodesInfo: GetForestFileByIDs err: %v", err)
		return err
	}
	for _, f := range fs {
		w.FID2FName[f.ID] = f.Name
	}

	for _, n := range w.G.Nodes {
		var s []string
		fIDs := w.VID2FIDs[n.ID]
		for _, fID := range fIDs {
			s = append(s, w.FID2FName[fID])
		}
		n.Tag[0].Cluster = strings.Join(s, "\n")
	}
	return nil
}

func (w *WrapperNebulaSearch) GetWordCloudGraph() error {
	if w.Cli == nil {
		return ErrCliIsNil
	}
	if err := w.DoWordCloudNql(); err != nil {
		return err
	}
	if err := w.BuildWordCloud(); err != nil {
		return err
	}
	var wls []string
	for _, word := range w.Wc {
		wls = append(wls, word.Word)
	}

	if w.ForestID == 109 {
		result := strings.Join(wls, `","`)
		result = result[:500]
		old, err := GetNodesByIDOld(w.ctx, 19, 109, result)
		if err != nil {
			return err
		}
		w.G = old
		return nil
	}

	if err := w.DoSubLimitNql(); err != nil {
		return err
	}
	if err := w.BuildGraph(); err != nil {
		return err
	}
	return nil
}

func escapeNebulaString(s string) string {
	return strings.NewReplacer(
		`%`, `%%`,
	).Replace(s)
}

// GetWordCloud 获取词云图所需数据
func GetWordCloud(ctx context.Context, forest_id uint, space_name string) ([]WordsCloud, error) {
	countconf, err := GetNebulaCountConf(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}
	var wordsClouds []WordsCloud
	cli, err := NewNebulaCLI()
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}
	defer cli.Release()
	ngql := fmt.Sprintf("USE %s; MATCH (v)-[e]-() WHERE v.entities.forest_id==%v AND e.forest_id == %v AND NOT v.entities.type IN %v "+
		"RETURN id(v) AS vertex_id, COUNT(e) AS degree ORDER BY degree DESC LIMIT %v;", space_name, forest_id, forest_id, "[\"chunk\",\"table\",\"image\",\"title\"]", countconf.WordCloudCount)
	if forest_id == 109 {
		ngql = fmt.Sprintf("USE know_19_129; MATCH (v)-[e]-() RETURN id(v) AS vertex_id, COUNT(e) AS degree ORDER BY degree DESC;")
	}
	resp, err := cli.Execute(ngql)
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}
	if resp.IsEmpty() {
		fmt.Println(ngql)
		return nil, fmt.Errorf("empty result")
	}
	for _, item := range resp.GetRows() {
		// fmt.Println(string(item.String()))
		wordsClouds = append(wordsClouds, WordsCloud{
			Word:   string(item.Values[0].SVal),
			Weight: *item.Values[1].IVal,
		})
	}
	return wordsClouds, nil
}

// GetNodeByID 根据ID获取相连图节点
func GetNodesByID(ctx context.Context, forestID uint, space_name, nodeID string) (*Graph, error) {
	var out = &Graph{}
	cli, err := NewNebulaCLI()
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}
	defer cli.Release()
	var resp *nebula_go.ResultSet
	if forestID == 109 {
		logs.DebugContextf(ctx, "try 109")
		resp, err = cli.Execute("USE know_19_129" +
			";GO 1 STEPS FROM " + nodeID + " OVER * BIDIRECT YIELD src(edge) AS src_id, dst(edge) AS dst_id, edge AS e;")
		if err != nil {
			logs.ErrorContextf(ctx, err.Error())
			return nil, err
		}
	} else {
		resp, err = cli.Execute("USE " + space_name +
			";GO 1 STEPS FROM \"" + nodeID + "\" OVER * BIDIRECT " +
			"WHERE properties($$).forest_id == " + fmt.Sprint(forestID) + " " +
			"AND properties(EDGE).forest_id == " + fmt.Sprint(forestID) + " " +
			"AND NOT properties($$).type IN [\"chunk\",\"table\",\"image\",\"title\"] " +
			"YIELD src(edge) AS src_id, dst(edge) AS dst_id, edge AS e | LIMIT 200;")
		if err != nil {
			logs.ErrorContextf(ctx, err.Error())
			return nil, err
		}
	}

	if resp.IsEmpty() {
		return nil, fmt.Errorf("empty result")
	}

	// 先查询边
	for _, item := range resp.GetRows() {
		// fmt.Println(item)
		out.Edges = append(out.Edges, Edge{
			Source:    string(item.Values[0].SVal),
			Target:    string(item.Values[1].SVal),
			CompanyID: uint(*item.Values[2].EVal.Props["company_id"].IVal),
			ForestID:  uint(*item.Values[2].EVal.Props["forest_id"].IVal),
			Uin:       uint(*item.Values[2].EVal.Props["uin"].IVal),
		})
	}
	// 遍历out.edges所有项得到其中所用到节点，其中值不重复
	var nodeSet = make(map[string]struct{})
	for _, item := range out.Edges {
		nodeSet[strings.ReplaceAll(item.Source, `\`, `\\`)] = struct{}{}
		nodeSet[strings.ReplaceAll(item.Target, `\`, `\\`)] = struct{}{}
	}
	nodeIDs := make([]string, 0, len(nodeSet))
	for nodeID := range nodeSet {
		nodeIDs = append(nodeIDs, nodeID)
	}
	// for nodeID := range nodeSet {
	// 	node, err := GetNodeInfo(uin, forestID, nodeID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	out.Nodes = append(out.Nodes, *node)
	// }
	nodes, err := GetNodesInfo(ctx, forestID, space_name, nodeIDs)
	if err != nil {
		return nil, err
	}
	out.Nodes = nodes
	return out, nil
}

// GetNode 获取节点属性
func GetNodeInfo(ctx context.Context, space_name, nodeID string) (*Node, error) {
	cli, err := NewNebulaCLI()
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}
	defer cli.Release()
	resp, err := cli.ExecuteAndCheck("USE " + space_name +
		";MATCH (v) WHERE id(v) == \"" + nodeID + "\" RETURN v;")
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}
	if resp.IsEmpty() {
		return nil, fmt.Errorf("empty result")
	}
	var node = &Node{}
	for _, item := range resp.GetRows() {
		// fmt.Println(string(item.Values[0].VVal.Vid.SVal)) // id
		node.ID = string(item.Values[0].VVal.Vid.SVal)
		for _, tag := range item.Values[0].VVal.Tags {
			var t = &Tag{} // tagname
			t.TagName = string(tag.Name)
			company_id := tag.Props["company_id"]
			t.CompanyID = uint(*company_id.IVal)
			forest_id := tag.Props["forest_id"]
			t.ForestID = uint(*forest_id.IVal)
			uin := tag.Props["uin"]
			t.CompanyID = uint(*uin.IVal)
			node.Tag = append(node.Tag, t)
		}
	}
	return node, nil
}

// GetNodes 获取多节点属性
func GetNodesInfo(ctx context.Context, forest_id uint, space_name string, nodeIDs []string) ([]*Node, error) {
	cli, err := NewNebulaCLI()
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}
	defer cli.Release()
	formattedNodeIDs := fmt.Sprintf(`["%s"]`, strings.Join(nodeIDs, `","`))

	if forest_id == 109 {
		space_name = "know_19_129"
	}

	resp, err := cli.ExecuteAndCheck("USE " + space_name +
		";MATCH (v) WHERE id(v) IN " + formattedNodeIDs + " RETURN v;")
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}
	if resp.IsEmpty() {
		return nil, fmt.Errorf("empty result")
	}
	var (
		nodes   []*Node
		isChunk = false
		//store map about fileid:filename
		fileMap = map[string]string{}
	)
	for _, item := range resp.GetRows() {
		var node = &Node{}
		node.ID = string(item.Values[0].VVal.Vid.SVal)
		for _, tag := range item.Values[0].VVal.Tags {
			var t = &Tag{} // tagname
			//get type
			tp := tag.Props["type"]
			t.Type = string(tp.SVal)
			//if type is chunk then do not display
			if t.Type == "chunk" {
				isChunk = true
				break
			}
			t.TagName = string(tag.Name)
			company_id := tag.Props["company_id"]
			t.CompanyID = uint(*company_id.IVal)
			forest_id := tag.Props["forest_id"]
			t.ForestID = uint(*forest_id.IVal)
			uin := tag.Props["uin"]
			t.Uin = uint(*uin.IVal)
			//get file_id
			fID := tag.Props["file_id"]
			fIDs := string(fID.SVal)
			if s, ok := fileMap[fIDs]; ok {
				//find a cache,take it
				t.Cluster = s
				node.Tag = append(node.Tag, t)
				continue
			}
			fileIDs := strings.Split(fIDs, "&&&")
			var uIDs []uint
			for _, id := range fileIDs {
				idi, err := strconv.Atoi(id)
				if err != nil {
					return nil, err
				}
				uIDs = append(uIDs, uint(idi))
			}
			fNames, err := forest.GetForestFileNameByIDs(uIDs)
			if err != nil {
				return nil, err
			}
			fileMap[fIDs] = strings.Join(fNames, "\n")
			t.Cluster = fileMap[fIDs]
			node.Tag = append(node.Tag, t)
		}
		if isChunk {
			isChunk = false
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// GetGraph 获取整个知识图谱
func GetWordCloudGraph(ctx context.Context, forest_id uint, space_name string) (*Graph, error) {
	words, err := GetWordCloud(ctx, forest_id, space_name)
	if err != nil {
		return nil, err
	}
	// 提取所有 Word 字段的值
	var wordList []string
	for _, word := range words {
		wordList = append(wordList, word.Word)
	}

	// 拼接成一个字符串
	result := strings.Join(wordList, `","`)
	// fmt.Println(result)

	if forest_id == 109 {
		fmt.Println("=============GET 109 Graph===============")
		result = result[:500]
		old, err := GetNodesByIDOld(ctx, 19, 109, result)
		if err != nil {
			return nil, err
		}
		return old, nil
	}

	// graphdata, err := GetNodesByID(forest_id, space_name, result)
	// if err != nil {
	// 	return nil, err
	// }
	graphdata, err := GetNodesByIDWithFile(ctx, forest_id, space_name, wordList)
	if err != nil {
		return nil, err
	}
	return graphdata, err
}

// NebulaCountConf
type NebulaCountConf struct {
	GraphCount     int `yaml:"graph_count"`
	WordCloudCount int `yaml:"wordcloud_count"`
}

func GetNebulaCountConf(ctx context.Context) (*NebulaCountConf, error) {
	defult := &NebulaCountConf{
		GraphCount:     3,
		WordCloudCount: 50,
	}
	conf := &NebulaCountConf{}
	err := settings.GetYaml("knowledge", "nebulacount", conf)
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return defult, nil
	}
	return conf, nil
}

// GetNodeByID 根据ID获取相连图节点
func GetNodesByIDWithFile(ctx context.Context, forestID uint, space_name string, nodeIds []string) (*Graph, error) {
	countconf, err := GetNebulaCountConf(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}
	var out = &Graph{}
	cli, err := NewNebulaCLI()
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}
	defer cli.Release()

	var resp *nebula_go.ResultSet

	queryStrs := []string{}
	for _, nodeID := range nodeIds {
		str := " (GO 1 STEPS FROM \"" + nodeID + "\" OVER * BIDIRECT" +
			" WHERE properties($$).forest_id == " + fmt.Sprint(forestID) +
			" AND properties(EDGE).forest_id == " + fmt.Sprint(forestID) +
			" AND NOT properties($$).type IN [\"chunk\",\"table\",\"image\",\"title\"] " +
			fmt.Sprintf(" YIELD src(edge) AS src_id, dst(edge) AS dst_id, edge AS e | LIMIT %v)", countconf.GraphCount)
		queryStrs = append(queryStrs, str)
	}

	querystr := "USE " + space_name + ";" + strings.Join(queryStrs, " UNION ALL ")
	fmt.Println(querystr)
	logs.InfoContextf(ctx, "querystr: %s", querystr)

	resp, err = cli.Execute(querystr)
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}

	if resp.IsEmpty() {
		return nil, fmt.Errorf("empty result")
	}
	logs.InfoContextf(ctx, "GetNodesByIDWithFile resp.GetRows(): %v", len(resp.GetRows()))
	// 先查询边
	for _, item := range resp.GetRows() {
		// fmt.Println(item)
		out.Edges = append(out.Edges, Edge{
			Source:    string(item.Values[0].SVal),
			Target:    string(item.Values[1].SVal),
			CompanyID: uint(*item.Values[2].EVal.Props["company_id"].IVal),
			ForestID:  uint(*item.Values[2].EVal.Props["forest_id"].IVal),
			Uin:       uint(*item.Values[2].EVal.Props["uin"].IVal),
		})
	}
	// 遍历out.edges所有项得到其中所用到节点，其中值不重复
	var nodeSet = make(map[string]struct{})
	for _, item := range out.Edges {
		nodeSet[strings.ReplaceAll(item.Source, `\`, `\\`)] = struct{}{}
		nodeSet[strings.ReplaceAll(item.Target, `\`, `\\`)] = struct{}{}
	}
	nodeIDs := make([]string, 0, len(nodeSet))
	for nodeID := range nodeSet {
		nodeIDs = append(nodeIDs, nodeID)
	}
	// for nodeID := range nodeSet {
	// 	node, err := GetNodeInfo(uin, forestID, nodeID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	out.Nodes = append(out.Nodes, *node)
	// }
	nodes, err := GetNodesInfo(ctx, forestID, space_name, nodeIDs)
	if err != nil {
		return nil, err
	}
	out.Nodes = nodes
	return out, nil
}

func GetNodesByIDOld(ctx context.Context, uin, forestID uint, nodeID string) (*Graph, error) {
	var out = &Graph{}
	cli, err := NewNebulaCLI()
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}
	defer cli.Release()
	resp, err := cli.Execute("USE know_19_129" +
		";GO 1 STEPS FROM \"" + nodeID + "\" OVER * BIDIRECT YIELD src(edge) AS src_id, dst(edge) AS dst_id, edge AS e;")
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}
	if resp.IsEmpty() {
		return nil, fmt.Errorf("empty result")
	}

	// 先查询边
	for _, item := range resp.GetRows() {
		// fmt.Println(item)
		out.Edges = append(out.Edges, Edge{
			Source:      string(item.Values[0].SVal),
			Target:      string(item.Values[1].SVal),
			Description: string(item.Values[2].EVal.Props["description"].SVal),
			Weight:      int64(*resp.GetRows()[0].Values[2].EVal.Props["weight"].FVal),
			Name:        string(resp.GetRows()[0].Values[2].EVal.Name),
		})
	}
	// 遍历out.edges所有项得到其中所用到节点，其中值不重复
	var nodeSet = make(map[string]struct{})
	for _, item := range out.Edges {
		nodeSet[item.Source] = struct{}{}
		nodeSet[item.Target] = struct{}{}
	}
	nodeIDs := make([]string, 0, len(nodeSet))
	for nodeID := range nodeSet {
		nodeIDs = append(nodeIDs, nodeID)
	}
	// for nodeID := range nodeSet {
	// 	node, err := GetNodeInfo(uin, forestID, nodeID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	out.Nodes = append(out.Nodes, *node)
	// }
	nodes, err := GetNodesInfoOld(ctx, uin, forestID, nodeIDs)
	if err != nil {
		return nil, err
	}
	out.Nodes = nodes
	return out, nil
}

// GetNodesInfoOld 获取多节点属性
func GetNodesInfoOld(ctx context.Context, uin, forestID uint, nodeIDs []string) ([]*Node, error) {
	cli, err := NewNebulaCLI()
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}
	defer cli.Release()
	formattedNodeIDs := fmt.Sprintf(`["%s"]`, strings.Join(nodeIDs, `","`))
	nsql := fmt.Sprintf("USE know_19_129" +
		";MATCH (v) WHERE id(v) IN " + formattedNodeIDs + " RETURN v;")
	fmt.Print(nsql)
	resp, err := cli.ExecuteAndCheck(nsql)
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}
	if resp.IsEmpty() {
		return nil, fmt.Errorf("empty result")
	}
	var nodes []*Node
	for _, item := range resp.GetRows() {
		var node = &Node{}
		node.ID = string(item.Values[0].VVal.Vid.SVal)
		for _, tag := range item.Values[0].VVal.Tags {
			var t = &Tag{} // tagname
			t.TagName = string(tag.Name)
			clusters := tag.Props["clusters"]
			t.Cluster = string(clusters.SVal) // cluster
			description := tag.Props["description"]
			t.Description = string(description.SVal)
			source_id := tag.Props["source_id"]
			t.SourceID = string(source_id.SVal)
			typ := tag.Props["type"]
			t.Type = string(typ.SVal)
			node.Tag = append(node.Tag, t)
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}
