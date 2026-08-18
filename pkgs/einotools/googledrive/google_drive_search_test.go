package googledrive

import (
	"encoding/json"
	"testing"

	"github.com/insmtx/corekg/pkgs/connectors"
	"github.com/stretchr/testify/assert"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

func TestGoogleDriveSearch(t *testing.T) {
	initTestDB()
	ctx := t.Context()
	connectors.InitProviders(ctx, "account", "pkl_connect_providers")
	t.Run("Search", func(t *testing.T) {
		conf := &Config{
			MaxResults: 3,
		}
		st, err := NewTool(ctx, conf)
		assert.NoError(t, err)

		tl, err := st.Info(ctx)
		assert.NoError(t, err)

		js, err := tl.ToJSONSchema()
		assert.NoError(t, err)
		body, err := js.MarshalJSON()
		assert.NoError(t, err)
		t.Logf("JSON Schema body: %s", string(body))

		// Test search with various parameters
		driveReq := &SearchRequest{
			Uin:        671, // Test UIN
			Queries:    []string{"天空之城"},
			MaxResults: 3,
		}
		driveReqBody, err := marshalString(driveReq)
		assert.NoError(t, err)

		toolOut, err := st.InvokableRun(ctx, driveReqBody)
		// Note: This might fail if no Google Drive token is available, which is expected in test environment
		if err != nil {
			t.Logf("Expected error in test environment (no Google Drive token): %v", err)
		} else {
			t.Logf("Tool output: %v", toolOut)
		}
	})

}

func initTestDB() {
	if err := dbtools.InitMultiDBConn(map[string]string{
		"account": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=True&loc=Local",
	}); err != nil {
		panic(err)
	}
}

func marshalString(v interface{}) (string, error) {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
