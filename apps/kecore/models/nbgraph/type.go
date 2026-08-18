package nbgraph

import "encoding/xml"

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
			Text string `xml:",chardata"`
			ID   string `xml:"id,attr"`
			Data []struct {
				Text string `xml:",chardata"`
				Key  string `xml:"key,attr"`
			} `xml:"data"`
		} `xml:"node"`
		Edge []struct {
			Text   string `xml:",chardata"`
			Source string `xml:"source,attr"`
			Target string `xml:"target,attr"`
			Data   []struct {
				Text string `xml:",chardata"`
				Key  string `xml:"key,attr"`
			} `xml:"data"`
		} `xml:"edge"`
	} `xml:"graph"`
}

const (
	KnowNodeTagDFString = "(source_id string,description string,type string,clusters string) " // 前缀文件明
	KnowNodeTagString   = "entities(chunk_id,company_id,file_id,forest_id,node_id,type,uin) "  // 前缀文件明

	KnowEdgeDFString = "_edge(description string,weight double,source_id string) " // 前缀文件明
	KnowEdgeString   = "relationships(company_id,forest_id,uin) "
)

// WordsCloud 词云图数据结构
type WordsCloud struct {
	Word   string `json:"word"`
	Weight int64  `json:"weight"`
	ID     string `json:"id"`
}

type Tag struct {
	TagName   string `json:"tag_name"` // tagname
	CompanyID uint   `json:"company_id"`
	ForestID  uint   `json:"forest_id"`
	Uin       uint   `json:"uin"`
	NodeID    string `json:"node_id"`
	//tmp
	Cluster     string `json:"cluster"`
	Description string `json:"description"`
	SourceID    string `json:"source_id"`
	Type        string `json:"type"`
	FileIDs     []uint `json:"file_ids"`
	ChunkID     string `json:"chunk_id"`
}

type Node struct {
	ID  string `json:"id"`
	Tag []*Tag `json:"tag"`
}
type Edge struct {
	Source    string `json:"source"`
	Target    string `json:"target"`
	SourceID  string `json:"source_id"`
	TargetID  string `json:"target_id"`
	CompanyID uint   `json:"company_id"`
	ForestID  uint   `json:"forest_id"`
	Uin       uint   `json:"uin"`
	//tmp
	Description string `json:"description"`
	Weight      int64  `json:"weight"`
	Name        string `json:"name"`
}
type Graph struct {
	Nodes []*Node `json:"nodes"`
	Edges []Edge  `json:"edges"`
}
