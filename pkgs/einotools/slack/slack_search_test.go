package slack

import (
	"encoding/json"
	"testing"

	"github.com/insmtx/corekg/pkgs/connectors"
	"github.com/stretchr/testify/assert"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

func TestSlackSearch(t *testing.T) {
	initTestDB()
	ctx := t.Context()
	connectors.InitProviders(ctx, "account", "pkl_connect_providers")
	slackReq := &SearchRequest{
		Uin:        330, // Test UIN
		Queries:    []string{"看看"},
		MaxResults: 3,
		InChannels: []string{"#channel-test", "#所有-correlated"},
		FromUsers:  []string{},   // 包括 in:@罗炀涛 映射到 FromUsers
		After:      "2021-12-31", // “2021”解析为年底
		Before:     "2025-09-25", // “昨天”解析为具体日期
	}

	t.Run("Query Building", func(t *testing.T) {
		query := buildSlackQuery(slackReq)
		expectedQuery := "看看 in:#channel-test in:#所有-correlated after:2021-12-31 before:2025-09-25"
		assert.Equal(t, expectedQuery, query)
	})

	t.Run("Search", func(t *testing.T) {
		conf := &Config{
			MaxResults: 5,
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

		slackReqBody, err := marshalString(slackReq)
		assert.NoError(t, err)

		toolOut, err := st.InvokableRun(ctx, slackReqBody)
		// Note: This might fail if no Slack token is available, which is expected in test environment
		if err != nil {
			t.Logf("Expected error in test environment (no Slack token): %v", err)
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
