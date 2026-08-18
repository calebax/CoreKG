package algofilehandle

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"

	"github.com/ygpkg/yg-go/logs"
)

// Graphml 文件解析原始值
type Graphml struct {
	XMLName        xml.Name `xml:"graphml"`
	Text           string   `xml:",chardata"`
	Xmlns          string   `xml:"xmlns,attr"`
	Xsi            string   `xml:"xsi,attr"`
	SchemaLocation string   `xml:"schemaLocation,attr"`
	Key            []struct {
		Text     string `xml:",chardata"`
		ID       string `xml:"id,attr"`
		For      string `xml:"for,attr"`
		AttrName string `xml:"attr.name,attr"`
		AttrType string `xml:"attr.type,attr"`
	} `xml:"key"`
	Graph struct {
		Text        string `xml:",chardata"`
		Edgedefault string `xml:"edgedefault,attr"`
		Node        []struct {
			ID   string `xml:"id,attr"`
			Data []struct {
				Text string `xml:",chardata"`
				Key  string `xml:"key,attr"`
			} `xml:"data"`
		} `xml:"node"`
		Edge []struct {
			ID     string `xml:"id,attr"`
			Source string `xml:"source,attr"`
			Target string `xml:"target,attr"`
			Data   []struct {
				Text string `xml:",chardata"`
				Key  string `xml:"key,attr"`
			} `xml:"data"`
		} `xml:"edge"`
	} `xml:"graph"`
}

// GraphmlSerialization 解析graphml文件
func GraphmlSerialization(graphmlBytes []byte) (*Graphml, error) {
	var graphML Graphml
	err := xml.Unmarshal(graphmlBytes, &graphML)
	if err != nil {
		fmt.Println("Error parsing graphml file:", err)
		return nil, err
	}

	return &graphML, err
}

type Node struct {
	ID          string   `json:"id"`          // id
	Uid         []string `json:"uid"`         // d3
	FilePath    string   `json:"file_path"`   // d2
	SourceID    string   `json:"source_id"`   // d1
	Description []string `json:"description"` // d0
}

type Edge struct {
	ID          string `json:"id"`          // id
	Source      string `json:"source"`      // 起始node
	Target      string `json:"target"`      // 终止node
	FilePath    string `json:"file_path"`   // d6
	SourceID    string `json:"source_id"`   // d5
	Description string `json:"description"` // d4
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// GenerateGraph 制作图对象
func GenerateGraph(graphml *Graphml) (*Graph, error) {
	graph := &Graph{}
	ctx := context.TODO()
	// 处理node
	for _, mlnode := range graphml.Graph.Node {
		node := Node{}
		for _, item := range mlnode.Data {
			switch item.Key {
			case "d1":
				node.SourceID = item.Text
			case "d2":
				node.FilePath = item.Text
			case "d3":
				arr, err := ToStringArray(item.Text)
				if err != nil {
					logs.ErrorContextf(ctx, "d3 ToStringArray Unmarshal failed: %v", err)
					return nil, err
				}
				node.Uid = arr
			case "d0":
				arr, err := ToStringArray(item.Text)
				if err != nil {
					logs.ErrorContextf(ctx, "d0 ToStringArray Unmarshal failed: %v", err)
					return nil, err
				}
				node.Description = arr
			}
		}
		node.ID = mlnode.ID
		graph.Nodes = append(graph.Nodes, node)
	}
	// 处理node
	for _, mledge := range graphml.Graph.Edge {
		edge := Edge{}
		for _, item := range mledge.Data {
			switch item.Key {
			case "d6":
				edge.FilePath = item.Text
			case "d5":
				edge.SourceID = item.Text
			case "d4":
				edge.Description = item.Text
			}
		}
		edge.ID = mledge.ID
		graph.Edges = append(graph.Edges, edge)
	}

	return graph, nil
}

// ToStringArray 将字符串转换为字符串数组
func ToStringArray(str string) ([]string, error) {
	var arr []string
	err := json.Unmarshal([]byte(str), &arr)
	if err != nil {
		logs.ErrorContextf(context.TODO(), "ToStringArray Unmarshal failed: %v , string: %s", err, str)
		return nil, err
	}
	return arr, nil
}
