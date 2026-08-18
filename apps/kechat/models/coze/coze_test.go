package coze

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

func TestWorkflowChat(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"core": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=True&loc=Local",
		"chat": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	token := os.Getenv("COREKG_TEST_TOKEN")
	if token == "" {
		t.Fatal("COREKG_TEST_TOKEN 未设置")
	}
	res, err := WorkflowChat(context.Background(), token,
		"7554430302517460992", chattype.InputList{
			{Name: "input", Value: "介绍下人工智能"},
		})
	fmt.Println(err)
	// logs.Infof("%+v", res)
	println(res)
}
