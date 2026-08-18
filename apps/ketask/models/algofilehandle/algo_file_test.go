package algofilehandle

import (
	"fmt"
	"os"
	"testing"
)

func TestGraphml(t *testing.T) {
	data, _ := os.ReadFile("./graph_chunk_entity_relation.graphml")
	gr, _ := GraphmlSerialization(data)
	ph, _ := GenerateGraph(gr)
	fmt.Println(ph)
}

func TestVDB(t *testing.T) {
	// data, _ := os.ReadFile("./vdb_entities.json")
	// vdb := &VdbEntities{}
	data, _ := os.ReadFile("./vdb_chunks.json")
	vdb := &VdbChunks{}
	// data, _ := os.ReadFile("./vdb_relationships.json")
	// vdb := &VdbRelationships{}

	GenerateVdbObject(data, vdb)
	fmt.Println(vdb.Data)
}
