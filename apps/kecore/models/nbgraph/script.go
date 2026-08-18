package nbgraph

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/ygpkg/yg-go/logs"
)

// GraphmlSerialization serialization graphml file
func GraphmlSerialization(graphmlBytes []byte) (*Graphml, error) {

	var graphML Graphml
	err := xml.Unmarshal(graphmlBytes, &graphML)
	if err != nil {
		fmt.Println("Error parsing graphml file:", err)
		return nil, err
	}

	return &graphML, err
}

func replaceString(str string) string {
	// 创建一个 Replacer，替换四组字符为空
	replacer := strings.NewReplacer(
		"|", "",
		">", "",
		"<", "",
		"\"", "",
		"'", "",
		"&gt;", "",
		"\n", "",
	)
	return replacer.Replace(str)
}

// BuildGraphNode build a string slice for nodes that stored in nebula graph
func BuildGraphNode(graphml *Graphml) ([]string, error) {

	res := make([]string, 0, (len(graphml.Graph.Node)/500)+1)
	var builder strings.Builder

	//trim all invalid character
	// reg := regexp.MustCompile(`[。，‘“’”：！| \r\n\t]|"<SEP>"`)
	// quoteReg := regexp.MustCompile(`"`)

	var tp, dsc, sid, cluster string
	for nodeIndex, node := range graphml.Graph.Node {
		for _, kv := range node.Data {
			kv.Text = replaceString(kv.Text)
			switch kv.Key {
			case "d1":
				tp = kv.Text
			case "d2":
				dsc = kv.Text
			case "d3":
				sid = kv.Text
			case "d0":
				cluster = kv.Text
			}
		}
		// node.ID = `"` + strings.ReplaceAll(node.ID[1:len(node.ID)-1], `"`, `'`) + `"`
		node.ID = replaceString(node.ID)
		builder.WriteString(fmt.Sprintf("\"%v\":(\"%v\",\"%v\",\"%v\",\"%v\"),", node.ID, tp, dsc, sid, cluster))
		if (nodeIndex+1)%500 == 0 {
			res = append(res, builder.String()[:builder.Len()-1])
			builder.Reset()
		}
	}
	if builder.Len() > 0 {

		res = append(res, builder.String()[:builder.Len()-1])
	}

	return res, nil
}

// BuildGraphEdge build a string slice for edges stored in nebula graph
func BuildGraphEdge(graphml *Graphml) ([]string, error) {
	res := make([]string, 0, (len(graphml.Graph.Edge)/500)+1)

	var builder strings.Builder
	var dsc, sid, wg string

	//trim all invalid character
	// reg := regexp.MustCompile(`[。，‘“’”：！\r\n\t]|"<SEP>"`)

	for edgeIndex, edge := range graphml.Graph.Edge {
		for _, kv := range edge.Data {
			kv.Text = replaceString(kv.Text)
			switch kv.Key {
			case "d5":
				wg = kv.Text
			case "d6":
				dsc = kv.Text
			case "d8":
				sid = kv.Text
			}
		}

		// edge.Source = `"` + strings.ReplaceAll(edge.Source[1:len(edge.Source)-1], `"`, `'`) + `"`
		// edge.Target = `"` + strings.ReplaceAll(edge.Target[1:len(edge.Target)-1], `"`, `'`) + `"`
		edge.Source = replaceString(edge.Source)
		edge.Target = replaceString(edge.Target)
		builder.WriteString(fmt.Sprintf(`"%v"->"%v":("%v",%v,"%v"),`, edge.Source, edge.Target, dsc, wg, sid))
		if (edgeIndex+1)%500 == 0 {
			res = append(res, builder.String()[:builder.Len()-1])
			builder.Reset()
		}
	}

	if builder.Len() > 0 {
		res = append(res, builder.String()[:builder.Len()-1])
	}

	return res, nil
}

// ImportGraph import graph from graphml into nebulaGraph
func ImportGraph(ctx context.Context, graphmlBytes []byte, uin, forestID, fileID uint) error {
	logs.InfoContextf(ctx, "[nebula] start ImportGraph :%v", fileID)
	//serialize graphml into struct
	graph, err := GraphmlSerialization(graphmlBytes)
	if err != nil {
		logs.ErrorContextf(ctx, "ImportGraph GraphmlSerialization ", err)
		return err
	}

	cli, err := NewNebulaCLI()
	if err != nil {
		logs.ErrorContextf(ctx, "ImportGraph NewNebulaCLI ", err)
		return err
	}
	defer cli.Release()

	if err = cli.InitSpaceSchema(ctx, uin, forestID, fileID); err != nil {
		logs.ErrorContextf(ctx, "ImportGraph InitSpaceSchema ", err)
		return err
	}

	nodesString, err := BuildGraphNode(graph)
	if err != nil {
		logs.ErrorContextf(ctx, "ImportGraph BuildGraphNode ", err)
		return err
	}

	//loop to check until detecting tag exists
	for {
		if err := cli.InsertNode(ctx, fmt.Sprintf("doc_%v", fileID)+KnowNodeTagString, `"know_detect_node":("","","","")`); err != nil {
			logs.DebugContextf(ctx, "detecting tag %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		if _, err := cli.ExecuteAndCheck(DELETE + VERTEX + `"know_detect_node"`); err != nil {
			logs.ErrorContextf(ctx, "ImportGraph ExecuteAndCheck DELETE VERTEX know_detect_node", err)
			return err
		}
		break
	}

	for _, str := range nodesString {
		// fmt.Println(str)
		if err := cli.InsertNode(ctx, fmt.Sprintf("doc_%v", fileID)+KnowNodeTagString, str); err != nil {
			logs.ErrorContextf(ctx, "InsertNode err:%v", err)
			return err
		}
	}

	edgeString, err := BuildGraphEdge(graph)
	if err != nil {
		logs.ErrorContextf(ctx, "ImportGraph BuildGraphEdge ", err)
		return err
	}

	for _, str := range edgeString {
		// fmt.Println(str)
		if err := cli.InsertEdge(ctx, fmt.Sprintf("doc_%v", fileID)+KnowEdgeString, str); err != nil {
			logs.ErrorContextf(ctx, "InsertEdge err:%v", err)
			return err
		}
	}
	return nil
}

// TaskCallBack forest's lib generated task call back
func TaskCallBack(ctx context.Context, finfo *foresttype.KnownowForestFile) error {

	// content, err := fs.GetForestGraphmlContent(forestID)
	// 改为导入单文档
	content, err := fs.GetFileGraphmlContent(finfo.GetAlgoFilePath(), finfo.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "TaskCallBack GetFileGraphmlContent %v", err)
		return err
	}

	//import into graph db
	if err = ImportGraph(ctx, content, finfo.Uin, finfo.ForestID, finfo.ID); err != nil {
		logs.ErrorContextf(ctx, "TaskCallBack ImportGraph %v", err)
		return err
	}

	return nil
}
