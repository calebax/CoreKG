package chunk

import (
	"context"
	"testing"

	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
)

func TestListChunksByFileID(t *testing.T) {
	ctx := t.Context()
	if err := InitESClient(ctx); err != nil {
		logs.FatalContextf(ctx, "[main] InitHistoryESClient failed, %s", err)
	}
	chunks, err := ListChunksByFileID(context.Background(), 2220)
	if err != nil {
		logs.ErrorContextf(ctx, "err: %v", err)
		return
	}
	for _, v := range chunks {
		logs.InfoContextf(ctx, "chunk: %s, source: %+v", v.ID, v.Source)
	}
}

func TestGetChunkByID(t *testing.T) {
	ctx := t.Context()
	if err := InitESClient(ctx); err != nil {
		logs.FatalContextf(ctx, "[main] InitHistoryESClient failed, %s", err)
	}
	chunks, err := GetChunkByID(context.Background(), "ef330255-ea07-489f-9d6e-aa4dfa661efd")
	if err != nil {
		logs.ErrorContextf(ctx, "err: %v", err)
		return
	}

	logs.InfoContextf(ctx, "chunk: %s, source: %+v", chunks.ID, chunks.Source)
}

func TestDisableFileChunk(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	ctx := t.Context()
	if err := InitESClient(ctx); err != nil {
		logs.FatalContextf(ctx, "[main] InitHistoryESClient failed, %s", err)
	}
	err := DisableFileChunk(context.Background(), 2220, false)
	if err != nil {
		logs.ErrorContextf(ctx, "err: %v", err)
		return
	}
}
