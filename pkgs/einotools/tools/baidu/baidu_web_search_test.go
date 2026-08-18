package baidu

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/settings"
)

func TestBaiduWebSearch(t *testing.T) {
	// export BAIDU_API_KEY="your_api_key"
	// apiKey := os.Getenv("BAIDU_API_KEY")
	initTestDB()
	apiKey, err := settings.GetText("corekg", "baidu_bce_api_key")
	assert.NoError(t, err)

	if apiKey == "" {
		t.Skip("BAIDU_API_KEY not set, skipping integration test")
	}

	ctx := t.Context()

	t.Run("Search", func(t *testing.T) {
		conf := &Config{
			ApiKey:     apiKey,
			MaxResults: 3,
		}
		st, err := NewBaiduWebSearch(ctx, conf)
		assert.NoError(t, err)

		_, err = st.Info(ctx)
		assert.NoError(t, err)

		// 构造请求参数
		req := &SearchRequest{
			Query: "Baidu Qianfan",
		}
		reqBody, err := marshalString(req)
		assert.NoError(t, err)

		// 调用工具
		toolOut, err := st.InvokableRun(ctx, reqBody)
		assert.NoError(t, err)
		t.Logf("Tool output: %v", toolOut)

		// 简单的结果验证
		var resp SearchResponse
		err = json.Unmarshal([]byte(toolOut), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Results, "Should return search results")
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
