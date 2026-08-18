package nbgraph

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/logs"
)

// GetForestNodes 获取森林所有节点
func GetForestNodes(ctx context.Context, cli *NebulaCli, forest_id uint, space_name string) ([]*Node, error) {
	resp, err := cli.ExecuteAndCheck(
		fmt.Sprintf("USE %s;MATCH (v)  WHERE v.entities.forest_id == %v RETURN v;", space_name, forest_id))
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}

	var nodes []*Node

	for _, item := range resp.GetRows() {
		var node = &Node{}
		node.ID = string(item.Values[0].VVal.Vid.SVal)
		for _, tag := range item.Values[0].VVal.Tags {
			var t = &Tag{} // tagname
			//get type
			tp := tag.Props["type"]
			t.Type = string(tp.SVal)
			chunk_id := tag.Props["chunk_id"]
			t.ChunkID = string(chunk_id.SVal)
			t.TagName = string(tag.Name)
			company_id := tag.Props["company_id"]
			t.CompanyID = uint(*company_id.IVal)
			forest_id := tag.Props["forest_id"]
			t.ForestID = uint(*forest_id.IVal)
			uin := tag.Props["uin"]
			t.Uin = uint(*uin.IVal)

			node_id := tag.Props["node_id"]
			t.NodeID = string(node_id.SVal)
			//get file_id
			fID := tag.Props["file_id"]
			fIDs := string(fID.SVal)
			fileIDs := strings.Split(fIDs, "&&&")
			var uIDs []uint
			for _, id := range fileIDs {
				idi, err := strconv.Atoi(id)
				if err != nil {
					return nil, err
				}
				uIDs = append(uIDs, uint(idi))
			}
			t.FileIDs = uIDs

			node.Tag = append(node.Tag, t)
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// GetForestEdges 获取森林所有边
func GetForestEdges(ctx context.Context, cli *NebulaCli, forest_id uint, space_name string) ([]*Edge, error) {
	resp, err := cli.ExecuteAndCheck(
		fmt.Sprintf("USE %s;MATCH ()-[e]-() WHERE e.forest_id == %v RETURN e;", space_name, forest_id))
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return nil, err
	}
	var edges []*Edge
	for _, item := range resp.GetRows() {
		edges = append(edges, &Edge{
			Source:    strings.TrimPrefix(string(item.Values[0].EVal.Src.SVal), fmt.Sprintf("%v_", forest_id)),
			Target:    strings.TrimPrefix(string(item.Values[0].EVal.Dst.SVal), fmt.Sprintf("%v_", forest_id)),
			SourceID:  string(item.Values[0].EVal.Src.SVal),
			TargetID:  string(item.Values[0].EVal.Dst.SVal),
			CompanyID: uint(*item.Values[0].EVal.Props["company_id"].IVal),
			ForestID:  uint(*item.Values[0].EVal.Props["forest_id"].IVal),
			Uin:       uint(*item.Values[0].EVal.Props["uin"].IVal),
		})
	}
	return edges, nil
}

// CopyForestGraph 复制森林图
func CopyForestGraph(ctx context.Context, srcID uint, forest_info *foresttype.KnownowForest, fileidmap map[uint]uint) error {
	// uinmap 中为所有file的id
	space_name := forest_info.EsIndex()
	cli, err := NewNebulaCLI()
	if err != nil {
		logs.ErrorContextf(ctx, "ImportGraph NewNebulaCLI ", err)
		return err
	}
	defer cli.Release()

	nodesString, err := BuildForestNode(ctx, cli, srcID, forest_info, fileidmap, space_name)
	if err != nil {
		return err
	}

	for _, str := range nodesString {
		if err := cli.InsertNode(ctx, KnowNodeTagString, str); err != nil {
			logs.ErrorContextf(ctx, "InsertNode err:%v", err)
			return err
		}
	}

	edgeString, err := BuildForestEdge(ctx, cli, srcID, forest_info, space_name)
	if err != nil {
		logs.ErrorContextf(ctx, "ImportGraph BuildGraphEdge ", err)
		return err
	}

	for _, str := range edgeString {
		if err := cli.InsertEdge(ctx, KnowEdgeString, str); err != nil {
			logs.ErrorContextf(ctx, "InsertEdge err:%v", err)
			return err
		}
	}

	return nil
}

func BuildForestNode(ctx context.Context, cli *NebulaCli, srcID uint, forest_info *foresttype.KnownowForest, fileidmap map[uint]uint, space_name string) ([]string, error) {
	var node_builder strings.Builder
	nodes, err := GetForestNodes(ctx, cli, srcID, space_name)
	if err != nil {
		return nil, err
	}
	nodesString := make([]string, 0, (len(nodes)/500)+1)

	for nodeindex, node := range nodes {
		var nodeids []string
		for _, tag := range node.Tag {
			for _, file_id := range node.Tag[0].FileIDs {
				nodeids = append(nodeids, fmt.Sprintf("%v", fileidmap[file_id]))
			}
			filestring := strings.Join(nodeids, "&&&")
			node_builder.WriteString(fmt.Sprintf("\"%v\":(\"%v\",%v,\"%v\",%v,\"%v\",\"%v\",%v),",
				fmt.Sprintf("%v_%s", forest_info.ID, tag.NodeID), tag.ChunkID, forest_info.CompanyID, filestring, forest_info.ID, tag.NodeID, tag.Type, forest_info.Uin))
		}
		if (nodeindex+1)%500 == 0 {
			nodesString = append(nodesString, node_builder.String()[:node_builder.Len()-1])
			node_builder.Reset()
		}
	}
	if node_builder.Len() > 0 {
		nodesString = append(nodesString, node_builder.String()[:node_builder.Len()-1])
	}
	return nodesString, nil
}

// BuildForestEdge .
func BuildForestEdge(ctx context.Context, cli *NebulaCli, srcID uint, forest_info *foresttype.KnownowForest, space_name string) ([]string, error) {
	var edge_builder strings.Builder
	edges, err := GetForestEdges(ctx, cli, srcID, space_name)
	if err != nil {
		return nil, err
	}
	edgesString := make([]string, 0, (len(edges)/500)+1)

	for i, edge := range edges {
		edge_builder.WriteString(fmt.Sprintf(`"%v_%v"->"%v_%v":(%v,%v,%v),`,
			forest_info.ID, edge.Source, forest_info.ID, edge.Target,
			forest_info.CompanyID, forest_info.ID, forest_info.Uin))
		if (i+1)%500 == 0 {
			edgesString = append(edgesString, edge_builder.String()[:edge_builder.Len()-1])
			edge_builder.Reset()
		}
	}
	if edge_builder.Len() > 0 {
		edgesString = append(edgesString, edge_builder.String()[:edge_builder.Len()-1])
	}

	return edgesString, nil
}
