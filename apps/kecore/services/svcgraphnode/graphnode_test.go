package svcgraphnode

import (
	"context"
	"fmt"
	"testing"

	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

func TestDESCEdgeTag(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"core": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})

	nebulagraph.InitNebulaConf(context.Background())
	cli, err := nebulagraph.NewNebulaCLI(context.Background(), "ke_graph_zXE8VpeNlaG4A0H8WDG7")
	if err != nil {
		t.Fatalf("NewNebulaCLI failed: %v", err)
	}
	defer cli.Release()

	sql := "DESC EDGE `包含11`"
	res, err := cli.ExecuteAndCheck(sql)
	if err != nil {
		t.Fatalf("ExecuteAndCheck failed: %v", err)
	}
	fmt.Println(res)
}
