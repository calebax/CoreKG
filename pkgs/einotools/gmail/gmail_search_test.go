package gmail

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/insmtx/corekg/pkgs/connectors"
	"github.com/stretchr/testify/assert"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"

	oauthUtils "github.com/insmtx/corekg/pkgs/utils/oauth"
)

func TestSearch(t *testing.T) {
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

		gsr := &SearchRequest{
			Uin:     671,
			Queries: []string{"IN-48311029", "zhangjunhu007"},
		}
		gsrBody, err := marshalString(gsr)
		assert.NoError(t, err)

		toolOut, err := st.InvokableRun(ctx, gsrBody)
		assert.NoError(t, err)
		t.Logf("Tool output: %v", toolOut)
		println(toolOut)
	})

}

func TestSearchWithProxy(t *testing.T) {
	initTestDB()
	ctx := t.Context()
	connectors.InitProviders(ctx, "account", "pkl_connect_providers")

	client := oauthUtils.CreateHttpClientWithProxy()

	resp, err := client.Post("https://www.googleapis.com/gmail/v1/users/me/messages", "application/json", nil)
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	t.Logf("Response body: %s", string(body))

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
