package sandbox

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ygpkg/yg-go/cache"
	"github.com/ygpkg/yg-go/cache/redis"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
)

func TestSandbox(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"core": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=True&loc=Local",
		"chat": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	ctx := t.Context()
	// 初始化 Redis
	if err := redispool.InitRedis("core", "loc_redis"); err != nil {
		logs.WarnContextf(ctx, "connect redis failed, %s", err)
	} else {
		cache.InitCache(redis.NewCache(redispool.Redis()))
	}
	t.Run("sandboxModeLocalCommand", func(t *testing.T) {
		cfg := &Config{
			Mode: SandboxModeLocalCommand,
		}
		s, err := NewSandbox(cfg)
		if err != nil {
			t.Fatalf("NewSandbox() error = %v", err)
		}
		testCode(t, s)
	})

	t.Run("SandboxModeRemoteHTTP", func(t *testing.T) {
		cfg := &Config{
			Mode:        SandboxModeRemoteHTTP,
			HttpBaseURL: "https://example.com",
			HttpToken:   "8b2a12f5-4c45-4fe2-83a7-04fc09db5087",
			Timeout:     300,
		}
		s, err := NewSandbox(cfg)
		if err != nil {
			t.Fatalf("NewSandbox() error = %v", err)
		}
		testCode(t, s)
		testPandasExcel(t, s)
	})
}

func testPandasExcel(t *testing.T, s Sandbox) {
	code := `
import pandas as pd

url = "https://example.com:58081/test-knownow/forest/20260121/871-7S41sCG1t.xlsx"
df = pd.read_excel(url)
print(len(df))
`
	ctx := t.Context()
	execResult, err := s.Exec(context.Background(), "python", code)
	if err != nil {
		t.Fatalf("testPandasExcel Exec() error = %v", err)
	}
	logs.InfoContext(ctx, "testPandasExcel Exec() result = %v", toJson(execResult))
}

func testCode(t *testing.T, s Sandbox) {
	code := `
for i in range(3):
    print(i)
`
	ctx := t.Context()
	checkRsult, err := s.CheckSyntax(context.Background(), "python", code)
	if err != nil {
		t.Fatalf("CheckSyntax() error = %v", err)
	}
	logs.InfoContextf(ctx, "CheckSyntax() result = %v", toJson(checkRsult))

	execResult, err := s.Exec(context.Background(), "python", code)
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}
	logs.InfoContext(ctx, "Exec() result = %v", toJson(execResult))

}

func toJson(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
