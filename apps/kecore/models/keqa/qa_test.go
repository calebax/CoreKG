package keqa

import (
	"context"
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
)

func TestImageSearchFile(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	// fore := &foresttype.KnownowForest{
	// 	Model: gorm.Model{
	// 		ID: 41,
	// 	},
	// 	ConfigID: 0,
	// }
	// ids, err := ImageSearchFile(context.Background(), fore, "https://example.com:53081/test-knownow/algo-lke/434/41/1419/images/c417909fc955a6fb90541d42c86aa5159f40b043ec9036dc8dd30e41276a9184.jpg")
	// fmt.Println(ids, err)

	ids, err := DoImageParseRequest(context.Background(), "https://example.com:53081/test-knownow/algo-lke/434/41/1419/images/c417909fc955a6fb90541d42c86aa5159f40b043ec9036dc8dd30e41276a9184.jpg")
	fmt.Println(ids, err)
}

func TestPreSearchQuestionChunk(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	essearch.InitEbConfig(context.TODO())
	ctx := &gin.Context{}
	wrapper, err := HandelSearchReference(ctx, []uint{96}, nil, "ke_0", "相机无法打开问题")
	if err != nil {
		logs.ErrorContextf(ctx, "[ForestChat] Failed to HandelSearchReference: %v", err)
		return
	}
	// start := time.Now()
	preSearchResult, _, err := wrapper.PreSearchQuestionChunk()
	if err != nil {
		return
	}
	// logs.Infof("数据处理完成，耗时: %v", time.Since(start))
	fmt.Println(len(preSearchResult.Hits.Hits))
}
