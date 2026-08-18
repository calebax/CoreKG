package graph

import (
	"encoding/json"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/task"
)

// GraphAlgoReq 算法请求体
type GraphAlgoReq struct {
	task.CommonPayload `json:",inline"`
	GraphID            uint                 `json:"graph_id"`
	IsIgnoreStatus     bool                 `json:"is_ignore_status"`
	FileID             uint                 `json:"file_id"`
	EsIndex            string               `json:"es_index"`
	Mode               foresttype.ParseMode `json:"mode"`
	Tags               []Tag                `json:"tags"`
	Edges              []Edge               `json:"edges"`
}

func (g *GraphAlgoReq) String() string {
	jsonData, err := json.Marshal(g)
	if err != nil {
		// logs.Errorf("PDFToString json.Marshal payload error: %v, payload :%+v", err, g)
		return ""
	}
	return string(jsonData)
}

type Tag struct {
	TagName    string                `json:"tag_name"`
	Properties foresttype.Properties `json:"properties"`
	Comment    string                `json:"comment"`
}

type Edge struct {
	EdgeName   string `json:"edge_name"`
	SrcTagName string `json:"src_tag_name"`
	DstTagName string `json:"dst_tag_name"`
	// Comment    string `json:"comment"`
}

// GraphAlgoResp 算法返回体
type GraphAlgoResp struct {
	GraphID uint        `json:"graph_id"`
	FileID  uint        `json:"file_id"`
	Nodes   []Node      `json:"nodes"`
	Edges   []EdgeValue `json:"edges"`
}

func (g *GraphAlgoResp) ReplaceStr() {
	// 修改 Nodes
	for i := range g.Nodes {
		g.Nodes[i].TagName = ReplaceString(g.Nodes[i].TagName)
		g.Nodes[i].Node = ReplaceString(g.Nodes[i].Node)
		for j := range g.Nodes[i].PropertiesValuse {
			g.Nodes[i].PropertiesValuse[j].Name = ReplaceString(g.Nodes[i].PropertiesValuse[j].Name)
			if str, ok := g.Nodes[i].PropertiesValuse[j].Value.(string); ok {
				g.Nodes[i].PropertiesValuse[j].Value = ReplaceValue(str)
			}
		}
	}

	// 修改 Edges
	for i := range g.Edges {
		g.Edges[i].EdgeName = ReplaceString(g.Edges[i].EdgeName)
		g.Edges[i].SrcNode = ReplaceString(g.Edges[i].SrcNode)
		g.Edges[i].SrcTag = ReplaceString(g.Edges[i].SrcTag)
		g.Edges[i].DstNode = ReplaceString(g.Edges[i].DstNode)
		g.Edges[i].DstTag = ReplaceString(g.Edges[i].DstTag)
	}
}

type Node struct {
	TagName          string                      `json:"tag_name"`
	Node             string                      `json:"node"`
	ChunkIDs         []string                    `json:"chunk_ids"`
	PropertiesValuse foresttype.PropertiesValues `json:"properties_values"`
}

type EdgeValue struct {
	EdgeName string `json:"edge_name"`
	SrcNode  string `json:"src_node_name"`
	SrcTag   string `json:"src_tag_name"`
	DstNode  string `json:"dst_node_name"`
	DstTag   string `json:"dst_tag_name"`
	// ChunkID     string `json:"chunk_id"`
	ChunkIDs    []string `json:"chunk_ids"`
	Description string   `json:"description"`

	SrcTagID uint `json:"src_tag_id"`
	DstTagID uint `json:"dst_tag_id"`
}
