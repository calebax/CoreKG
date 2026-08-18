package nebulagraph

import (
	"context"
	"fmt"
	"testing"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

func TestNebula(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	InitNebulaConf(context.Background())
	cli, _ := NewNebulaCLI(context.Background(), "yFkYlhg3KU7jTd5ARatW")
	tag := &foresttype.GraphTag{
		TagName: "驱动",
		TagType: "EDGE",
	}
	res, err := cli.GetTagDesc(tag)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)
}

func TestNebulaResault(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	InitNebulaConf(context.Background())
	cli, _ := NewNebulaCLI(context.Background(), "ke_graph_vsiiSt6lnZF4aF3mndlv")
	a, err := cli.GetKnowledgeGraph(KnowledgeGraphReq{SrcName: "储罐", Limit: 20})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(a.Edges)
	// logs.Infof("%+v", a.Edges)
	// fmt.Println(a.Nodes)
	// for _, v := range a.Nodes {
	// 	for _, t := range v.Tags {
	// 		for _, p := range t.PropertiesValues {
	// 			fmt.Println(p)
	// 		}
	// 	}
	// }
}

func TestNebulaPathResault(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	InitNebulaConf(context.Background())
	cli, _ := NewNebulaCLI(context.Background(), "ke_graph_vsiiSt6lnZF4aF3mndlv")
	err := cli.FindNodePath("提醒指示灯", "提手", 5)
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println(a.Edges)
	// logs.Infof("%+v", a.Edges)
}

func TestGetNodeInfo(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	InitNebulaConf(context.Background())
	cli, _ := NewNebulaCLI(context.Background(), "ke_graph_0WIgXwSvTjbgCZyrTRCE")
	defer cli.Release()
	_, err := cli.GetNodesInfo([]string{"数据知识库对话", "用户"})
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println(a.Edges)
	// logs.Infof("%+v", a)
}

func TestGetNodesGraph(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	InitNebulaConf(context.Background())
	cli, _ := NewNebulaCLI(context.Background(), "ke_graph_0WIgXwSvTjbgCZyrTRCE")
	a, err := cli.GetNodesGraph([]string{"数据知识库对话", "用户"})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(a.Edges)
	// logs.Infof("%+v", a.Edges)
	fmt.Println(a.Nodes)
	for _, v := range a.Nodes {
		for _, t := range v.Tags {
			for _, p := range t.PropertiesValues {
				fmt.Println(p)
			}
		}
	}
}
