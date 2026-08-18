package nbgraph

import (
	"context"
	"fmt"
	"testing"

	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

func Test(t *testing.T) {
	// str, _ := GetWordCloud(6, 77)
	str, err := GetWordCloudGraph(context.Background(), 4, "ke_0")
	fmt.Println(str)
	// _, err := GetNode(6, 105, "OPENAI")

	// err := DeleteTag(6, 105, 1)
	fmt.Println(err)
	fmt.Println(len(str.Nodes))
}

func TestGetForestEdges(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	InitNebulaConf(context.Background())
	cli, err := NewNebulaCLI()
	if err != nil {
		// logs.Errorf("ImportGraph NewNebulaCLI ", err)
		return
	}
	defer cli.Release()
	edges, err := GetForestEdges(context.Background(), cli, 41, "ke_0")
	// logs.Infof("edges: %+v,tags: %+v", edges[0], edges[0].Tag[0])
	// logs.Infof("edges: %+v", edges[0])
	fmt.Println(len(edges))
	fmt.Println(err)
}
