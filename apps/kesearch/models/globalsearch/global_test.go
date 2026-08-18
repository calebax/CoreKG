package globalsearch

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/stretchr/testify/assert"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/logs"
)

func TestSearchAgent(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	ctx := t.Context()
	essearch.InitEbConfig(t.Context())
	// SearchAgent("人工智能的概念", 1, 2)
	wrapper, err := NewForestWrapper(&GlobalSearchWrapper{
		Ctx:          context.Background(),
		Text:         "1",
		Uin:          434,
		CompanyID:    18,
		IsSemantics:  true,
		EsIndex:      "ke_0",
		ImageUrl:     "",
		SubjectCount: 3,
		ItemCount:    3,
	})
	if err != nil {
		t.Fatalf("HandelGlobalSearch error: %v", err)
	}
	res, err := wrapper.SearchAgent()
	if err != nil {
		t.Fatalf("SearchForest error: %v", err)
	}
	for _, agent := range res {
		logs.InfoContextf(ctx, "agent: %+v", agent)
	}
}

func TestSearchFile(t *testing.T) {
	testutils.Initialize(testutils.AppNameKesearch)
	defer testutils.Close()

	ctx := testutils.NewCtx(testutils.WithUin(384))

	err := InitHighLightConfig(ctx)
	assert.Nil(t, err)

	wrapper, err := NewForestWrapper(&GlobalSearchWrapper{
		Ctx:          ctx,
		Text:         "人工智能的概念",
		Uin:          384,
		CompanyID:    18,
		ForestIDs:    []uint{41},
		IsSemantics:  true,
		EsIndex:      "ke_0",
		ImageUrl:     "",
		SubjectCount: 3,
		ItemCount:    3,
	})
	assert.Nil(t, err)

	res, err := wrapper.FindFileChunkAggs("chunk")
	assert.Nil(t, err)
	for _, v := range res.Aggregations.ByFile.Buckets {
		t.Logf("v:%s", logs.JSON(v.TopDocsPerFile.Hits.Hits))
	}
}

func TestMigrateFileName(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	ctx := t.Context()
	// 查询所有文件
	filename := "新的文件名.pdf"
	fileID := 2220
	// 根据文件id修改文件名
	client, err := essearch.InitESClient(ctx)
	if err != nil {
		return
	}

	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("term", esquery.BuildMap("file_id", fileID))).
		Set("script", esquery.BuildMap("source", "ctx._source.file_name = params.new_name",
			"lang", "painless",
			"params", esquery.BuildMap("new_name", filename)))
	querybyte, err := query.BuildBytes()
	if err != nil {
		return
	}
	logs.InfoContextf(ctx, "update file name querybyte: %v", string(querybyte))

	resp, err := client.UpdateByQuery(
		[]string{"ke_0"},
		client.UpdateByQuery.WithBody(bytes.NewBuffer(querybyte)),
		client.UpdateByQuery.WithContext(ctx),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "update file name error: %v", err)
		return
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "update file name error: %v", string(body))
		return
	}
}
