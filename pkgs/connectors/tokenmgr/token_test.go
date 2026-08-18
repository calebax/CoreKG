package tokenmgr

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/insmtx/corekg/pkgs/connectors"
	"github.com/stretchr/testify/assert"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

func TestGetToken(t *testing.T) {
	initTestDB()
	connectors.InitProviders(t.Context(), "account", "pkl_connect_providers")
	ctx := context.Background()

	externalToken, ok := GetToken(ctx, 671, "gmail")

	// 打印 ok
	assert.True(t, ok, "expected ok to be true")

	data, err := json.Marshal(externalToken)
	if err != nil {
		t.Errorf("marshal error: %v", err)
	} else {
		t.Logf("externalToken: %s", string(data))
	}

}

func initTestDB() {
	if err := dbtools.InitMultiDBConn(map[string]string{
		"account": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=True&loc=Local",
	}); err != nil {
		panic(err)
	}
}
