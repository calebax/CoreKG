package svcexcelforest

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/storage"
	"gorm.io/gorm"
)

var TaskStatusFields = []string{
	"parse_status",
	"mindmap_status",
	"analysis_status",
	"knowledge_status",
	"graph_status",
	"desc_status",
}

// AnalyzeXlsx 分析xlsx文件
func AnalyzeXlsx(ctx *gin.Context, req *AnalyzeXlsxReq) (err error) {
	// 添加 recover 机制
	defer func() {
		if r := recover(); r != nil {
			logs.ErrorContextf(ctx, "[AnalyzeXlsx] panic recovered: %v", r)
			// 将 panic 转换为 error 返回
			err = fmt.Errorf("panic occurred: %v", r)
		}
	}()

	// 获取文件
	forestFileEntity, err := forest.NewForestFileDao().GetByID(ctx, req.ForestFileID)
	if err != nil {
		logs.ErrorContextf(ctx, "[AnalyzeXlsx] get forest file error: %v", err)
		return err
	}

	fileEntity, err := storage.GetFileByID(dbutil.Core().WithContext(ctx), forestFileEntity.CoreFileID)
	if err != nil {
		logs.ErrorContextf(ctx, "[AnalyzeXlsx] get file error: %v", err)
		return err
	}

	// 获取 db
	dbInstanceEntity, err := forest.NewForestDBInstanceDao().GetByCond(ctx, &forest.ForestDBInstanceCond{
		ForestID: forestFileEntity.ForestID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[AnalyzeXlsx] get db instance error: %v", err)
		return err
	}
	dbEntity, err := forest.NewForestDBDao().GetByCond(ctx, &forest.ForestDBCond{
		ForestID:           forestFileEntity.ForestID,
		ForestDBInstanceID: dbInstanceEntity.ID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[AnalyzeXlsx] get db error: %v", err)
		return err
	}
	excelDB, err := dbutil.GetDB(dbEntity.DBName, dbInstanceEntity.BuildMysqlDNS(dbEntity.DBName, dbEntity.DBMeta.Mysql.Charset))
	if err != nil {
		logs.ErrorContextf(ctx, "[AnalyzeXlsx] get db error: %v", err)
		return err
	}

	// 更新文件初始状态
	forestFileInitUpdateMap := map[string]interface{}{
		"preview_able": foresttype.PreViewAbleStatusAccept,
	}
	for _, v := range TaskStatusFields {
		forestFileInitUpdateMap[v] = foresttype.TaskStatusRunning
	}

	if err := forest.NewForestFileDao().UpdateMap(ctx, req.ForestFileID, forestFileInitUpdateMap); err != nil {
		logs.ErrorContextf(ctx, "[AnalyzeXlsx] update forest file error: %v", err)
		return err
	}

	sheetList, err := getSheetList(ctx, fileEntity.PublicURL)
	if err != nil {
		logs.ErrorContextf(ctx, "[AnalyzeXlsx] get sheet list error: %v", err)
		return err
	}

	sheetMetadataMap := make(map[string][]SheetMetadata)

	for _, sheetName := range sheetList {
		cleaner := NewDataClean()
		// sheetName, fileEntity.PublicURL, headerRowNum
		cleanRes, err := cleaner.CleanSheet(ctx, &CleanSheetReq{
			ForestID:     forestFileEntity.ForestID,
			ForestFileID: forestFileEntity.ID,
			FileID:       fileEntity.ID,
			FileUrl:      fileEntity.PublicURL,
			SheetName:    sheetName,
		})
		if err != nil {
			logs.ErrorContextf(ctx, "[AnalyzeXlsx] clean sheet error: %v", err)
			return err
		}
		if len(cleanRes.SubSheetList) == 0 {
			logs.InfoContextf(ctx, "sheet %s has no data", sheetName)
			continue
		}
		for _, subSheet := range cleanRes.SubSheetList {
			insertReq := BatchInsertToDBReq{
				CleanSubSheet: subSheet,
			}
			insertRes, err := cleaner.BatchInsertToDB(ctx, excelDB, &insertReq)
			if err != nil {
				logs.ErrorContextf(ctx, "[AnalyzeXlsx] batch insert to db error: %v", err)
				continue
			}
			if !insertRes.SheetMetadata.IsValid {
				logs.InfoContextf(ctx, "sheet %s is not valid", sheetName)
				continue
			}
			sheetMetadataMap[sheetName] = append(sheetMetadataMap[sheetName], insertRes.SheetMetadata)
		}
	}

	var tableInsertEntityList foresttype.ForestTableList
	var parentSheetInsertEntityList foresttype.ForestExcelSheetList
	var subSheetInsertEntityList foresttype.ForestExcelSheetList
	for sheet, metadataList := range sheetMetadataMap {
		parentSheetInsertEntityList = append(parentSheetInsertEntityList, foresttype.ForestExcelSheet{
			ForestID:     forestFileEntity.ForestID,
			ForestFileID: forestFileEntity.ID,
			SheetName:    sheet,
			SheetType:    foresttype.ExcelSheetTypeNormal,
			SheetMeta: foresttype.ForestExcelSheetMeta{
				Summary: sheet,
			},
		})
		for _, metadata := range metadataList {
			tableInsertEntityList = append(tableInsertEntityList, foresttype.ForestTable{
				ForestID:  forestFileEntity.ForestID,
				DBID:      dbEntity.ID,
				Tablename: metadata.TableName,
				SyncedAt:  time.Now(),
			})
			subSheetInsertEntityList = append(subSheetInsertEntityList, foresttype.ForestExcelSheet{
				ForestID:     forestFileEntity.ForestID,
				ForestFileID: forestFileEntity.ID,
				SheetName:    sheet,
				SheetType:    foresttype.ExcelSheetTypeSub,
				HeaderMode:   metadata.HeaderMode,
				HeaderRowNum: metadata.HeaderRowNum,
				SheetMeta: foresttype.ForestExcelSheetMeta{
					ColumnMetadataList: metadata.ColumnMetaDataList,
					Summary:            metadata.Summary,
					Remark:             metadata.Remark,
				},
			})
		}

	}

	txErr := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := forest.NewForestExcelSheetDao().BatchInsert(ctx, parentSheetInsertEntityList); err != nil {
			logs.ErrorContextf(ctx, "[AnalyzeXlsx] batch insert sheet error: %v", err)
			return err
		}
		if err := forest.NewForestTableDao().BatchInsert(ctx, tableInsertEntityList); err != nil {
			logs.ErrorContextf(ctx, "[AnalyzeXlsx] batch insert table error: %v", err)
			return err
		}
		for i := range tableInsertEntityList {
			subSheetInsertEntityList[i].ForestTableID = tableInsertEntityList[i].ID
		}
		parentSheetIDMap := make(map[string]uint)
		for i := range parentSheetInsertEntityList {
			parentSheetIDMap[parentSheetInsertEntityList[i].SheetName] = parentSheetInsertEntityList[i].ID
		}
		for i := range subSheetInsertEntityList {
			subSheetInsertEntityList[i].ParentID = parentSheetIDMap[subSheetInsertEntityList[i].SheetName]
		}
		if err := forest.NewForestExcelSheetDao().BatchInsert(ctx, subSheetInsertEntityList); err != nil {
			logs.ErrorContextf(ctx, "[AnalyzeXlsx] batch insert sheet error: %v", err)
			return err
		}

		// 更新状态为成功
		forestFileSuccessUpdateMap := make(map[string]interface{})
		for _, v := range TaskStatusFields {
			forestFileSuccessUpdateMap[v] = foresttype.TaskStatusSuccess
		}
		if err := forest.NewForestFileDao().UpdateMap(ctx, forestFileEntity.ID, forestFileSuccessUpdateMap); err != nil {
			logs.ErrorContextf(ctx, "[AnalyzeXlsx] update forest file success error: %v", err)
			return err
		}

		// 更新知识库状态，如果有一个文件解析成功，则知识库状态更新为解析成功
		forestUpdateMap := map[string]interface{}{
			"knowledge_status": foresttype.TaskStatusSuccess,
		}
		if err := forest.NewForestDao().UpdateMap(ctx, forestFileEntity.ForestID, forestUpdateMap); err != nil {
			logs.ErrorContextf(ctx, "[AnalyzeXlsx] update forest success error: %v", err)
			return err
		}

		return nil
	})
	if txErr != nil {
		logs.ErrorContextf(ctx, "[AnalyzeXlsx] transaction error: %v", txErr)
		// 更新文件状态为失败
		forestFileFailUpdateMap := make(map[string]interface{})
		for _, v := range TaskStatusFields {
			forestFileFailUpdateMap[v] = foresttype.TaskStatusFail
		}
		if err := forest.NewForestFileDao().UpdateMap(ctx, req.ForestFileID, forestFileFailUpdateMap); err != nil {
			logs.ErrorContextf(ctx, "[AnalyzeXlsx] update forest file fail error: %v", err)
			return err
		}
		return txErr
	}

	return nil
}
