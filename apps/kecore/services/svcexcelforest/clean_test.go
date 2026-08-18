package svcexcelforest

import (
	"testing"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/stretchr/testify/assert"
	"github.com/ygpkg/yg-go/logs"
)

func TestCleanSheet(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))
	var forestID, forestFileID, fileID uint = 831, 6200, 36242
	var sheetName, fileUrl string = "金融", "https://example.com:58081/test-knownow/forest/384/831/6200.xlsx"
	req := CleanSheetReq{
		ForestID:     forestID,
		ForestFileID: forestFileID,
		FileID:       fileID,
		FileUrl:      fileUrl,
		SheetName:    sheetName,
	}
	res, err := NewDataClean().CleanSheet(ctx, &req)
	assert.Nil(t, err)
	for _, v := range res.SubSheetList {
		t.Logf("summary: %s", v.Summary)
		t.Logf("remark: %s", v.Remark)
	}
}

func TestBatchInsertToDB(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))
	var forestID, forestFileID, fileID uint = 831, 6200, 36242
	var sheetName, fileUrl string = "金融", "https://example.com:58081/test-knownow/forest/384/831/6200.xlsx"
	cleanReq := CleanSheetReq{
		ForestID:     forestID,
		ForestFileID: forestFileID,
		FileID:       fileID,
		FileUrl:      fileUrl,
		SheetName:    sheetName,
	}
	forestFileEntity, err := forest.NewForestFileDao().GetByID(ctx, forestFileID)
	assert.Nil(t, err)

	// 获取 db
	dbInstanceEntity, err := forest.NewForestDBInstanceDao().GetByCond(ctx, &forest.ForestDBInstanceCond{
		ForestID: forestFileEntity.ForestID,
	})
	assert.Nil(t, err)
	dbEntity, err := forest.NewForestDBDao().GetByCond(ctx, &forest.ForestDBCond{
		ForestID:           forestFileEntity.ForestID,
		ForestDBInstanceID: dbInstanceEntity.ID,
	})
	assert.Nil(t, err)
	excelDB, err := dbutil.GetDB(dbEntity.DBName, dbInstanceEntity.BuildMysqlDNS(dbEntity.DBName, dbEntity.DBMeta.Mysql.Charset))
	assert.Nil(t, err)
	cleanRes, err := NewDataClean().CleanSheet(ctx, &cleanReq)
	assert.Nil(t, err)
	t.Log(logs.JSON(cleanRes))
	for _, subSheet := range cleanRes.SubSheetList {
		insertReq := BatchInsertToDBReq{
			CleanSubSheet: subSheet,
		}

		res, err := NewDataClean().BatchInsertToDB(ctx, excelDB, &insertReq)
		assert.Nil(t, err)
		t.Log(logs.JSON(res))
	}
}

func TestGetHeaderSummary(t *testing.T) {
	data := []string{
		"2025\t2025\t2025\t2025",
		"重点产业\t重点产业\t重点产业\t重点产业",
		"指标\t2024年（亿元）\t同比±%\t占GDP比重±%",
		"*数字经济核心产业增加值\t6305\t7.1\t28.8",
		"*文化产业增加值\t3448\t6.5\t15.8",
	}
	summary, rowNum := NewDataClean().getHeaderSummary(data)
	t.Log(summary)
	t.Log(rowNum)
}
