package essearch

import (
	"context"
	"fmt"
	"testing"

	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
)

func TestSearchTital(t *testing.T) {
	// DeleteFileReferences("ke_test", []uint{123})
	ctx := context.Background()
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	InitEbConfig(ctx)
	wrapper, err := NewEsSearchWrapper(context.Background(), "ke_0", "相机无法打开问题", []uint{96}, nil)
	if err != nil {
		logs.ErrorContextf(ctx, "GetEmbedding error: %v", err)
		return
	}

	a, err := wrapper.SearchTitle()
	if err != nil {
		logs.ErrorContextf(ctx, "SearchTitle error: %v", err)
		return
	}
	for _, v := range a.Hits.Hits {
		fmt.Println(v.ID)
	}
}

func TestSummarizeSearch(t *testing.T) {
	ctx := context.Background()
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	InitEbConfig(ctx)
	wrapper, err := NewEsSearchWrapper(context.Background(), "ke_0", "如何处理倒车雷达 ECU 与网关控制模块通讯故障?", []uint{1244}, nil)
	if err != nil {
		logs.ErrorContextf(ctx, "GetEmbedding error: %v", err)
		return
	}

	a, err := wrapper.SummarizeSearch()
	if err != nil {
		logs.ErrorContextf(ctx, "SummarizeSearch error: %v", err)
		return
	}
	for _, v := range a.Hits.Hits {
		fmt.Println(v.ID)
		fmt.Println(v.Source.Description)
	}
}
