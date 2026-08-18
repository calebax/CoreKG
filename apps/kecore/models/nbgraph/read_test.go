package nbgraph

import (
	"context"
	"os"
	"testing"
)

func TestImport(t *testing.T) {
	// SearchAll()
	data, err := os.ReadFile("/home/zoe/Downloads/graph_chunk_entity_relation (2).graphml")
	if err != nil {
		panic(err)
	}
	// // data, _ := os.ReadFile("C:\\Users\\13065\\Downloads\\graph_chunk_entity_relation.graphml")
	if err := ImportGraph(context.Background(), data, 19, 129, 1251); err != nil {
		t.Fatal(err)
	}
	// fmt.Println(err)
	// TaskCallBack(6, 102, 829)
}
