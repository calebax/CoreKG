package reranksearch

import (
	"context"
	"fmt"
	"testing"

	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
)

func TestSearch(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"account": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	ctx := t.Context()
	if err := essearch.InitEbConfig(ctx); err != nil {
		logs.FatalContextf(ctx, "[main] InitEbConfig failed, %s", err)
	}
	chatquestion.InitHistoryESClient(ctx)

	wrapper, err := NewRerankSearchWrapper(context.Background(), "ke_0", "请详细介绍小贝无线产品WAP662H的尺寸和重量？", []uint{41}, nil, GetDefaultConfig(), nil)
	if err != nil {
		logs.ErrorContextf(ctx, "NewRerankSearchWrapper error: %v", err)
		return
	}
	res, err := wrapper.RerankSearchChunk()
	if err != nil {
		logs.ErrorContextf(ctx, "SearchQuestionChunk RerankSearchChunk error: %v", err)
		return
	}

	fmt.Println(len(res))
	for _, v := range res {
		// fmt.Println(v)
		fmt.Println(v.FileName)
	}
}

func TestAAA(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"account": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	ctx := t.Context()
	if err := essearch.InitEbConfig(ctx); err != nil {
		logs.FatalContextf(ctx, "[main] InitEbConfig failed, %s", err)
	}
	chatquestion.InitHistoryESClient(ctx)

	wrapper, err := NewRerankSearchWrapper(context.Background(), "ke_0", "请详细介绍小贝无线产品WAP662H的尺寸和重量？", []uint{41}, nil, GetDefaultConfig(), nil)
	if err != nil {
		logs.ErrorContextf(ctx, "NewRerankSearchWrapper error: %v", err)
		return
	}
	res, err := wrapper.SearchQuestionChunk()
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "RerankSearchChunk SearchQuestionChunk error: %v", err)
		return
	}
	for _, v := range res {
		fmt.Println("-----------------------------------------")
		logs.InfoContextf(ctx, "%+v", v)
	}
	// r1, _ := wrapper.SortRerankChunk(res)
	// r1a, _ := wrapper.SearchChunkSequence(r1)
	// j1 := wrapper.JoinNeighborChunks(r1, r1a)
	// fmt.Println(len(j1))
	// for _, v := range j1 {
	// 	fmt.Println("-----------------------------------------")
	// 	logs.Infof("%+v", v)
	// }
}

func TestRerank(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"account": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	ctx := t.Context()
	if err := essearch.InitEbConfig(ctx); err != nil {
		logs.FatalContextf(ctx, "[main] InitEbConfig failed, %s", err)
	}
	chatquestion.InitHistoryESClient(ctx)

	date, _ := GetRerank(context.Background(), "小贝无线产品WAP662H的尺寸和重量", []string{"新华三"})
	fmt.Println(date.Data[0])
}

func TestParseYgPosString(t *testing.T) {
	fmt.Println(parseYgPosString(t.Context(), "<!--yg_pos1,193,1078,658,1119,1119yg_pos-->"))
}
