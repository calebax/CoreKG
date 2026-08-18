package svcforest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoforest"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/concqueue"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
)

const (
	ResourceTypeForest     = "forest"
	ResourceTypeDir        = "dir"
	ResourceTypeFile       = "file"
	ResourceTypeExcel      = "excel"
	ResourceTypeExcelSheet = "excel_sheet"
	ResourceTypeMysqlDB    = "mysql_db"
	ResourceTypeMysqlTable = "mysql_table"
	ResourceTypeQAPair     = "qa_pair"
)

func GetResourceBaseInfo(ctx *gin.Context, req *dtoforest.GetResourceBaseInfoRequest) (res *dtoforest.GetResourceBaseInfoResponse, err error) {

	res = &dtoforest.GetResourceBaseInfoResponse{}
	authResourceCond := &forest.KeResourceScopeCond{
		BaseCond: forest.BaseCond{
			Uin: runtime.Uin(ctx),
		},
		CompanyID:    runtime.CompanyID(ctx),
		ResourceType: foresttype.ResourceTypeForest,
	}
	authForestIDs, err := forest.NewKeResourceScopeDao().GetAuthResourceIDs(ctx, authResourceCond)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetResourceBaseInfo] fail, err: %v", err)
		return nil, err
	}
	if len(authForestIDs) == 0 {
		logs.WarnContextf(ctx, "[GetResourceBaseInfo] without auth forest")
		return res, nil
	}
	authForestIDs = utils.SliceDuplicate(authForestIDs)

	authForestEntityList, err := forest.NewForestDao().GetListByCond(ctx, &forest.ForestCond{
		IDs:            authForestIDs,
		ForestTypeList: []foresttype.ForestType{foresttype.ForestTypeFile, foresttype.ForestTypeData, foresttype.ForestTypeQA},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[GetResourceBaseInfo] get auth forest fail, err: %v", err)
		return nil, err
	}
	authForestEntityMap := authForestEntityList.ToMap()

	var errs []error
	allNodeList := make([]*dtoforest.ForestResourceTreeNode, 0)
	var mu sync.Mutex
	conc := concqueue.New(6, 6)
	// conc.Submit(func(_ context.Context) error {
	// 	mu.Lock()
	// 	defer mu.Unlock()
	// 	// 查 sheet
	// 	sheetNodeList, err := getExcelSheetNodeList(ctx, authForestIDs, authForestEntityMap)
	// 	if err != nil {
	// 		logs.ErrorContextf(ctx, "[GetResourceBaseInfo] getExcelSheetNodeList fail, err: %v", err)
	// 		errs = append(errs, err)
	// 		return err
	// 	}
	// 	allNodeList = append(allNodeList, sheetNodeList...)
	// 	return nil
	// })
	conc.Submit(func(_ context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		// 查文件
		fileNodeList, err := getFileNodeList(ctx, authForestIDs, authForestEntityMap)
		if err != nil {
			logs.ErrorContextf(ctx, "[GetResourceBaseInfo] getFileNodeList fail, err: %v", err)
			errs = append(errs, err)
			return err
		}
		allNodeList = append(allNodeList, fileNodeList...)
		return nil
	})

	conc.Submit(func(_ context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		// 查目录
		dirNodeList, err := getDirNodeList(ctx, authForestIDs, authForestEntityMap)
		if err != nil {
			logs.ErrorContextf(ctx, "[GetResourceBaseInfo] getDirNodeList fail, err: %v", err)
			errs = append(errs, err)
			return err
		}
		allNodeList = append(allNodeList, dirNodeList...)
		return nil
	})

	conc.Submit(func(_ context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		// 查 table
		mysqlTableNodeList, err := getMysqlTableNodeList(ctx, authForestIDs, authForestEntityMap)
		if err != nil {
			logs.ErrorContextf(ctx, "[GetResourceBaseInfo] getMysqlTableNodeList fail, err: %v", err)
			errs = append(errs, err)
			return err
		}
		allNodeList = append(allNodeList, mysqlTableNodeList...)
		return nil
	})

	conc.Submit(func(_ context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		// 查 table
		mysqlDbNodeList, err := getMysqlDBNodeList(ctx, authForestIDs, authForestEntityMap)
		if err != nil {
			logs.ErrorContextf(ctx, "[GetResourceBaseInfo] getMysqlDBNodeList fail, err: %v", err)
			errs = append(errs, err)
			return err
		}
		allNodeList = append(allNodeList, mysqlDbNodeList...)
		return nil
	})

	conc.Submit(func(_ context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		// 查知识库
		forestNodeList, err := getForestNodeList(ctx, authForestIDs)
		if err != nil {
			logs.ErrorContextf(ctx, "[GetResourceBaseInfo] getForestNodeList fail, err: %v", err)
			errs = append(errs, err)
			return err
		}
		allNodeList = append(allNodeList, forestNodeList...)
		return nil
	})
	if errCnt := conc.StopAndWait(); errCnt > 0 {
		logs.ErrorContextf(ctx, "[GetResourceBaseInfo] conc.StopAndWait fail, errCnt: %d", errCnt)
		return nil, fmt.Errorf("get resource base info fail")
	}

	tree := dtoforest.BuildTree(allNodeList)
	res.Response.Tree = tree
	return res, nil
}

func getForestNodeList(ctx *gin.Context, authforestIDs []uint) ([]*dtoforest.ForestResourceTreeNode, error) {
	forestEntityList, err := forest.NewForestDao().GetListByCond(ctx, &forest.ForestCond{
		IDs: authforestIDs,
	})
	if err != nil {
		return nil, err
	}
	nodeList := make([]*dtoforest.ForestResourceTreeNode, 0, len(forestEntityList))
	for _, v := range forestEntityList {
		dataSourceSubtype := v.DataSourceSubType
		if dataSourceSubtype == "" {
			dataSourceSubtype = foresttype.ForestDataSourceSubtypeStandard
		}
		nodeItem := &dtoforest.ForestResourceTreeNode{
			ForestID:                v.ID,
			ID:                      v.ID,
			Name:                    v.Name,
			NodeType:                ResourceTypeForest,
			Key:                     fmt.Sprintf("%s_%s_%d", dataSourceSubtype, ResourceTypeForest, v.ID),
			ForestType:              v.ForestType,
			ForestDataSourceType:    v.DataSourceType,
			ForestDataSourceSubtype: dataSourceSubtype,
		}
		nodeList = append(nodeList, nodeItem)
	}
	return nodeList, err
}

func getDirNodeList(ctx *gin.Context, authForestIDs []uint, forestEntityMap map[uint]foresttype.KnownowForest) ([]*dtoforest.ForestResourceTreeNode, error) {
	dirEntityList, err := forest.NewForestFileDao().GetListByCond(ctx, &forest.ForestFileCond{
		IsDir:     types.True,
		ForestIDs: authForestIDs,
		Enable:    types.True,
	})
	if err != nil {
		return nil, err
	}
	nodeList := make([]*dtoforest.ForestResourceTreeNode, 0, len(dirEntityList))
	for _, v := range dirEntityList {

		forestEntity := forestEntityMap[v.ForestID]
		dataSourceSubType := forestEntity.DataSourceSubType
		if dataSourceSubType == "" {
			dataSourceSubType = foresttype.ForestDataSourceSubtypeStandard
		}
		nodeItem := &dtoforest.ForestResourceTreeNode{
			ForestID:                v.ForestID,
			ForestType:              forestEntity.ForestType,
			ForestDataSourceType:    forestEntity.DataSourceType,
			ForestDataSourceSubtype: dataSourceSubType,
			ID:                      v.ID,
			Name:                    v.Name,
			NodeType:                ResourceTypeDir,
			Key:                     fmt.Sprintf("%s_%s_%d", dataSourceSubType, ResourceTypeDir, v.ID),
		}
		if v.ParentID != 0 {
			nodeItem.ParentKey = fmt.Sprintf("%s_%s_%d", dataSourceSubType, ResourceTypeDir, v.ParentID)
		} else {
			nodeItem.ParentKey = fmt.Sprintf("%s_%s_%d", dataSourceSubType, ResourceTypeForest, v.ForestID)
		}
		nodeList = append(nodeList, nodeItem)
	}
	return nodeList, err
}

func getFileNodeList(ctx *gin.Context, authForestIDs []uint, forestEntityMap map[uint]foresttype.KnownowForest) ([]*dtoforest.ForestResourceTreeNode, error) {
	// var conditions []string
	// var args []any
	// // 处理名称模糊查询
	// if strings.TrimSpace(name) != "" {
	// 	conditions = append(conditions, "name LIKE ?")
	// 	args = append(args, "%"+strings.TrimSpace(name)+"%")
	// }

	// // 处理指定目录ID查询
	// if len(fileIDs) > 0 {
	// 	validFileIDs := make([]uint, 0, len(fileIDs))
	// 	for _, v := range fileIDs {
	// 		if v > 0 {
	// 			validFileIDs = append(validFileIDs, v)
	// 		}
	// 	}
	// 	if len(validFileIDs) > 0 {
	// 		conditions = append(conditions, "id IN ?")
	// 		args = append(args, fileIDs)
	// 	}
	// }
	// if len(authForestIDs) > 0 {
	// 	conditions = append(conditions, "forest_id IN ?")
	// 	validForestIDs := make([]uint, 0, len(authForestIDs))
	// 	for _, v := range authForestIDs {
	// 		if v > 0 {
	// 			validForestIDs = append(validForestIDs, v)
	// 		}
	// 	}
	// 	args = append(args, validForestIDs)
	// }

	// ========================= filter ban list ==========================

	ap, err := forest.NewAccessProvider(ctx, &forest.ContextModel{
		ResourceType: foresttype.ResourceTypeForestFile,
		ScopeType:    foresttype.ScopeTypeUser,
		ScopeID:      runtime.Uin(ctx),
		Action:       foresttype.ActionBan,
	}).Action(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "filter ban list failed: %v", err)
		return nil, err
	}

	// ========================== filter ban list ==========================

	fileEntityList, err := forest.NewForestFileDao().GetListByCond(ctx, &forest.ForestFileCond{
		ForestIDs:   authForestIDs,
		IsDir:       types.False,
		ParseStatus: foresttype.TaskStatusSuccess,
		Enable:      types.True,
		NotInIDs:    ap.BanList,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[getFileNodeList] fail, err: %v", err)
		return nil, err
	}

	nodeList := make([]*dtoforest.ForestResourceTreeNode, 0, len(fileEntityList))
	for _, v := range fileEntityList {
		forestEntity := forestEntityMap[v.ForestID]
		dataSourceSubType := forestEntity.DataSourceSubType
		if dataSourceSubType == "" {
			dataSourceSubType = foresttype.ForestDataSourceSubtypeStandard
		}
		nodeItem := &dtoforest.ForestResourceTreeNode{
			ForestID:                v.ForestID,
			ForestType:              forestEntity.ForestType,
			ForestDataSourceType:    forestEntity.DataSourceType,
			ForestDataSourceSubtype: dataSourceSubType,
			ID:                      v.ID,
			Name:                    v.Name,
			NodeType:                ResourceTypeFile,
			Key:                     fmt.Sprintf("%s_%d", ResourceTypeFile, v.ID),
		}
		if forestEntity.DataSourceSubType == foresttype.ForestDataSourceSubtypeExcel {
			nodeItem.NodeType = ResourceTypeExcel
			nodeItem.Key = fmt.Sprintf("%s_%d", ResourceTypeExcel, v.ID)
		}
		if v.ParentID != 0 {
			nodeItem.ParentKey = fmt.Sprintf("%s_%s_%d", dataSourceSubType, ResourceTypeDir, v.ParentID)
		} else {
			nodeItem.ParentKey = fmt.Sprintf("%s_%s_%d", dataSourceSubType, ResourceTypeForest, v.ForestID)
		}
		nodeList = append(nodeList, nodeItem)
	}
	return nodeList, nil
}

func getExcelSheetNodeList(ctx *gin.Context, authForestIDs []uint, forestEntityMap map[uint]foresttype.KnownowForest) ([]*dtoforest.ForestResourceTreeNode, error) {
	sheetEntityList, err := forest.NewForestExcelSheetDao().GetListByCond(ctx, &forest.ForestExcelSheetCond{
		ForestIDs: authForestIDs,
		SheetType: foresttype.ExcelSheetTypeNormal,
		Enable:    types.True,
	})
	if err != nil {
		return nil, err
	}
	nodeList := make([]*dtoforest.ForestResourceTreeNode, 0, len(sheetEntityList))
	for _, v := range sheetEntityList {
		forestEntity := forestEntityMap[v.ForestID]
		nodeItem := &dtoforest.ForestResourceTreeNode{
			ForestID:                v.ForestID,
			ForestType:              forestEntity.ForestType,
			ForestDataSourceType:    forestEntity.DataSourceType,
			ForestDataSourceSubtype: forestEntity.DataSourceSubType,
			ID:                      v.ID,
			ParentID:                v.ForestFileID,
			Name:                    v.SheetName,
			NodeType:                ResourceTypeExcelSheet,
			Key:                     fmt.Sprintf("%s_%d", ResourceTypeExcelSheet, v.ID),
			ParentKey:               fmt.Sprintf("%s_%d", ResourceTypeExcel, v.ForestFileID),
		}
		nodeList = append(nodeList, nodeItem)
	}
	return nodeList, err
}

func getMysqlDBNodeList(ctx *gin.Context, authForestIDs []uint, forestEntityMap map[uint]foresttype.KnownowForest) ([]*dtoforest.ForestResourceTreeNode, error) {
	dbEntityList, err := forest.NewForestDBDao().GetListByCond(ctx, &forest.ForestDBCond{
		ForestIDs: authForestIDs,
		Enable:    types.True,
	})
	if err != nil {
		return nil, err
	}
	nodeList := make([]*dtoforest.ForestResourceTreeNode, 0, len(dbEntityList))
	for _, v := range dbEntityList {
		forestEntity := forestEntityMap[v.ForestID]
		if forestEntity.DataSourceSubType != foresttype.ForestDataSourceSubtypeMySQL {
			continue
		}
		nodeItem := &dtoforest.ForestResourceTreeNode{
			ForestID:                v.ForestID,
			ForestType:              forestEntity.ForestType,
			ForestDataSourceType:    forestEntity.DataSourceType,
			ForestDataSourceSubtype: forestEntity.DataSourceSubType,
			ID:                      v.ID,
			ParentID:                v.ForestID,
			Name:                    v.DBName,
			NodeType:                ResourceTypeMysqlDB,
			Key:                     fmt.Sprintf("%s_%d", ResourceTypeMysqlDB, v.ID),
			ParentKey:               fmt.Sprintf("%s_%s_%d", forestEntity.DataSourceSubType, ResourceTypeForest, v.ForestID),
		}

		nodeList = append(nodeList, nodeItem)
	}
	return nodeList, nil
}

func getMysqlTableNodeList(ctx *gin.Context, authForestIDs []uint, forestEntityMap map[uint]foresttype.KnownowForest) ([]*dtoforest.ForestResourceTreeNode, error) {

	tableEntityList, err := forest.NewForestTableDao().GetListByCond(ctx, &forest.ForestTableCond{
		ForestIDs: authForestIDs,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[getMysqlTableNodeList] fail, err: %v", err)
		return nil, err
	}
	nodeList := make([]*dtoforest.ForestResourceTreeNode, 0, len(tableEntityList))
	for _, v := range tableEntityList {
		forestEntity := forestEntityMap[v.ForestID]
		if forestEntity.DataSourceSubType != foresttype.ForestDataSourceSubtypeMySQL {
			continue
		}
		nodeItem := &dtoforest.ForestResourceTreeNode{
			ForestID:                v.ForestID,
			ForestType:              forestEntity.ForestType,
			ForestDataSourceType:    forestEntity.DataSourceType,
			ForestDataSourceSubtype: forestEntity.DataSourceSubType,
			ID:                      v.ID,
			ParentID:                v.DBID,
			Name:                    v.Tablename,
			NodeType:                ResourceTypeMysqlTable,
			Key:                     fmt.Sprintf("%s_%d", ResourceTypeMysqlTable, v.ID),
			ParentKey:               fmt.Sprintf("%s_%d", ResourceTypeMysqlDB, v.DBID),
		}
		nodeList = append(nodeList, nodeItem)
	}
	return nodeList, nil
}

func RedisKeyFileEnableStatusKey(id uint) string {
	return fmt.Sprintf("File[%v]EnableStatusKey", id)
}

// SetResourceEnable will set a resource enable or disable
func SetResourceEnable(ctx *gin.Context, req *dtoforest.SetResourceEnableRequest) (res *dtoforest.SetResourceEnableResponse, err error) {
	res = &dtoforest.SetResourceEnableResponse{}
	switch req.Request.ResourceType {
	case ResourceTypeFile:
		fID, err := strconv.ParseUint(req.Request.ResourceIDs[0], 10, 64)
		if err != nil {
			logs.ErrorContextf(ctx, "[SetResourceEnable] strconv.ParseUint(%v) failed, err: %v", req.Request.ResourceIDs[0], err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_invalid_file_id"
			return res, nil
		}

		uFID := uint(fID)
		f, err := forest.GetForestFileByID(uFID)
		if err != nil {
			logs.ErrorContextf(ctx, "[SetResourceEnable] svcforest.SetResourceEnable(file_id:%v) failed, err: %v", uFID, err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_get_forest_file_failed"
			return res, nil
		}
		frs, err := forest.GetForestByID(ctx, f.ForestID)
		if err != nil {
			logs.ErrorContextf(ctx, "[SetResourceEnable] svcforest.SetResourceEnable(forest_id:%v) failed, err: %v", f.ForestID, err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_get_forest_failed"
			return res, nil
		}

		//TODO process ids with for range
		//file is a dir
		var (
			fIDs          = []uint{uFID}
			isDisableFlag bool
		)

		if f.IsDir == 1 {
			//resource is a dir, set all children enable/disable
			fs, err := forest.GetDirFiles(ctx, uFID)
			if err != nil {
				logs.ErrorContextf(ctx, "[SetResourceEnable] svcforest.SetResourceEnable(file_id:%v) failed, err: %v", uFID, err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_get_dir_files_failed"
				return res, nil
			}
			fIDs = []uint{uFID}
			for _, v := range fs {
				fIDs = append(fIDs, v.ID)
			}
		}

		//init DisableFlag
		if req.Request.Enable == 1 {
			isDisableFlag = false
		} else {
			isDisableFlag = true
		}

		nx, err := redispool.SetNX(RedisKeyFileEnableStatusKey(uFID), "", time.Minute*10)
		if err != nil {
			logs.ErrorContextf(ctx, "redispool.SetNX error")
			return res, err
		}
		if !nx {
			logs.WarnContextf(ctx, "could't get fils's redis lock")
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_resource_busy"
			return res, nil
		}

		defer func() {
			if err := redispool.Del(RedisKeyFileEnableStatusKey(uFID)); err != nil {
				logs.ErrorContextf(ctx, "redispool.Del faild")
			}
		}()

		if err = dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
			//update mysql
			if err := forest.NewForestFileDao().WithTx(tx).UpdateIDsMap(ctx, fIDs, map[string]interface{}{
				"enable": req.Request.Enable,
			}); err != nil {
				logs.ErrorContextf(ctx, "[SetResourceEnable] svcforest.SetResourceEnable(file_id:%v) failed, err: %v", uFID, err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_update_dir_files_failed"
				return err
			}

			return nil
		}); err != nil {
			logs.ErrorContextf(ctx, "[SetResourceEnable] fail, err: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_set_forest_failed"
			return res, nil
		}
		go func() {
			//update es
			//TODO process ids with for range
			gctx := ctx.Copy()
			if err := UpdateChunkDisableByFileIDs(gctx, fIDs, isDisableFlag, frs.EsIndex()); err != nil {
				logs.ErrorContextf(gctx, "[SetResourceEnable] svcforest.SetResourceEnable(file_id:%v) failed, err: %v", uFID, err)
			} else {
				logs.InfoContextf(gctx, "[SetResourceEnable] svcforest.SetResourceEnable(file_id:%v) success", uFID)
			}
		}()

	case ResourceTypeQAPair:
		//set es update query index
		frs, err := forest.GetForestByID(ctx, req.Request.ForestID)
		if err != nil {
			logs.ErrorContextf(ctx, "[SetResourceEnable] svcforest.SetResourceEnable(forest_id:%v) failed, err: %v", req.Request.ForestID, err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_get_forest_failed"
			return res, nil
		}
		esCli, err := essearch.InitESClient(ctx)
		if err != nil {
			logs.ErrorContextf(ctx, "[SetResourceEnable] svcforest.SetResourceEnable(forest_id:%v) failed, err: %v", req.Request.ForestID, err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_init_es_client_failed"
			return res, nil
		}
		updateDoc := make(map[string]interface{})
		updateDoc["enable"] = req.Request.Enable

		bulkBody := new(bytes.Buffer)
		e := json.NewEncoder(bulkBody)
		for _, v := range req.Request.ResourceIDs {
			if err := e.Encode(map[string]interface{}{
				"update": map[string]interface{}{
					"_index": frs.EsIndex(),
					"_id":    v,
				},
			}); err != nil {
				logs.ErrorContextf(ctx, "[SetResourceEnable] encode failed, err: %v", err)
				return nil, err
			}
			if err := e.Encode(map[string]interface{}{"doc": updateDoc}); err != nil {
				logs.ErrorContextf(ctx, "[SetResourceEnable] encode doc failed, err: %v", err)
				return nil, err
			}
		}
		resp, err := esCli.Bulk(bulkBody, esCli.Bulk.WithContext(ctx), esCli.Bulk.WithRefresh("true"))
		if err != nil {
			return res, fmt.Errorf("ES bulk request failed: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logs.ErrorContextf(ctx, "[SetResourceEnable] ES bulk response body close failed, err: %v", err)
			}
		}(resp.Body)
		if resp.IsError() {
			body, _ := io.ReadAll(resp.Body)
			logs.ErrorContextf(ctx, "[SetResourceEnable] ES bulk error: %s search: %v", string(body), bulkBody.String())
			return nil, fmt.Errorf("ES bulk error: %s search: %v", string(body), bulkBody.String())
		}
	case ResourceTypeMysqlDB:
		//Parse all string to uint
		dbIDs := make([]uint, 0, len(req.Request.ResourceIDs))
		for _, v := range req.Request.ResourceIDs {
			dbID, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				logs.ErrorContextf(ctx, "[SetResourceEnable] strconv.ParseUint(%v) failed, err: %v", v, err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_invalid_db_id"
				return res, nil
			}
			dbIDs = append(dbIDs, uint(dbID))
		}

		if err := dbutil.Knownow().Transaction(func(db *gorm.DB) error {
			if err := forest.NewForestDBDao().WithTx(db).UpdateIDsMap(ctx, dbIDs, map[string]interface{}{
				"enable": req.Request.Enable,
			}); err != nil {
				logs.ErrorContextf(ctx, "[SetResourceEnable] svcforest.SetResourceEnable(db_ids:%v) failed, err: %v", dbIDs, err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_update_db_ids_failed"
				return err
			}

			if err := forest.NewForestTableDao().WithTx(db).UpdateForestIsDMap(ctx, dbIDs, map[string]interface{}{
				"enable": req.Request.Enable,
			}); err != nil {
				logs.ErrorContextf(ctx, "[SetResourceEnable] svcforest.SetResourceEnable(db_ids:%v) failed, err: %v", dbIDs, err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_update_db_ids_failed"
				return err
			}
			return nil
		}); err != nil {
			logs.ErrorContextf(ctx, "[SetResourceEnable] svcforest.SetResourceEnable(db_ids:%v) failed, err: %v", dbIDs, err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_update_db_ids_failed"
			return res, nil
		}
	case ResourceTypeExcel:
		excelIDs := make([]uint, 0, len(req.Request.ResourceIDs))
		for _, v := range req.Request.ResourceIDs {
			excelID, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				logs.ErrorContextf(ctx, "[SetResourceEnable] strconv.ParseUint(%v) failed, err: %v", v, err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_invalid_db_id"
				return res, nil
			}
			excelIDs = append(excelIDs, uint(excelID))
		}

		if err := dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
			if err := forest.NewForestFileDao().WithTx(tx).UpdateIDsMap(ctx, excelIDs, map[string]interface{}{
				"enable": req.Request.Enable,
			}); err != nil {
				logs.ErrorContextf(ctx, "[SetResourceEnable] svcforest.SetResourceEnable(excel_ids:%v) failed, err: %v", excelIDs, err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_update_excel_sheets_failed"
				return err
			}
			if err := forest.NewForestExcelSheetDao().WithTx(tx).UpdateExcelIDsMap(ctx, excelIDs, map[string]interface{}{
				"enable": req.Request.Enable,
			}); err != nil {
				logs.ErrorContextf(ctx, "[SetResourceEnable] svcforest.SetResourceEnable(excel_ids:%v) failed, err: %v", excelIDs, err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_update_excel_sheets_failed"
				return err
			}
			return nil
		}); err != nil {
			logs.ErrorContextf(ctx, "[SetResourceEnable] svcforest.SetResourceEnable(excel_ids:%v) failed, err: %v", excelIDs, err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_update_dir_files_failed"
			return res, nil
		}
	case ResourceTypeMysqlTable:
		tableIDs := make([]uint, 0, len(req.Request.ResourceIDs))
		for _, v := range req.Request.ResourceIDs {
			tableID, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				logs.ErrorContextf(ctx, "[SetResourceEnable] strconv.ParseUint(%v) failed, err: %v", v, err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_invalid_table_id"
				return res, nil
			}
			tableIDs = append(tableIDs, uint(tableID))
		}
		if err := forest.NewForestTableDao().UpdateMapByIDs(ctx, tableIDs, map[string]interface{}{
			"enable": req.Request.Enable,
		}); err != nil {
			logs.ErrorContextf(ctx, "[SetResourceEnable] svcforest.SetResourceEnable(table_ids:%v) failed, err: %v", tableIDs, err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_update_table_ids_failed"
			return res, nil
		}
	case ResourceTypeExcelSheet:
		sheetIDs := make([]uint, 0, len(req.Request.ResourceIDs))
		for _, v := range req.Request.ResourceIDs {
			sheetID, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				logs.ErrorContextf(ctx, "[SetResourceEnable] strconv.ParseUint(%v) failed, err: %v", v, err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_invalid_db_id"
				return res, nil
			}
			sheetIDs = append(sheetIDs, uint(sheetID))
		}

		if err := forest.NewForestExcelSheetDao().UpdateIDsMap(ctx, sheetIDs, map[string]interface{}{
			"enable": req.Request.Enable,
		}); err != nil {
			logs.ErrorContextf(ctx, "[SetResourceEnable] svcforest.SetResourceEnable(sheet_ids:%v) failed, err: %v", sheetIDs, err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_update_excel_sheets_failed"
			return res, nil
		}
	}
	return res, nil
}

// UpdateChunkDisableByFileIDs will update chunk disable by file ids
func UpdateChunkDisableByFileIDs(ctx context.Context, fileIDs []uint, disable bool, index string) error {
	// 1. 初始化 ES 客户端
	client, err := essearch.InitESClient(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "InitESClient failed: %v", err)
		return err
	}

	// 2. 构建 Query DSL (使用 map[string]interface{})
	updateBody := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{
					// type in (?) => terms 查询
					{
						"terms": map[string]interface{}{
							"type": []ragtypes.ChunkType{"chunk", "image", "table", "video", "formula"},
						},
					},
					// file_id in (?) => terms 查询
					{
						"terms": map[string]interface{}{
							"file_id": fileIDs,
						},
					},
				},
			},
		},
		// 3. 构建 Script DSL
		"script": map[string]interface{}{
			"source": "ctx._source.is_disable = params.new_value", // painless脚本
			"lang":   "painless",
			"params": map[string]interface{}{
				"new_value": disable, // 传入目标布尔值
			},
		},
		// 4. 处理版本冲突：继续处理其他文档
		"conflicts": "proceed",
	}

	// 5. 将 DSL 转换为 []byte
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(updateBody); err != nil {
		logs.ErrorContextf(ctx, "failed to encode update DSL: %v", err)
		return fmt.Errorf("failed to encode update DSL: %w", err)
	}

	logs.DebugContextf(ctx, "ES UpdateByQuery DSL: %s", buf.String())

	// 6. 执行 ES UpdateByQuery 查询
	resp, err := client.UpdateByQuery(
		[]string{index}, // 索引名作为字符串数组传入
		client.UpdateByQuery.WithBody(&buf),
		client.UpdateByQuery.WithContext(ctx),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es UpdateByQuery request failed: %v", err)
		return fmt.Errorf("es UpdateByQuery request failed: %w", err)
	}
	defer resp.Body.Close()

	// 7. 处理响应结果
	if resp.IsError() {
		// 读取并返回ES的错误信息
		var e map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
			logs.ErrorContextf(ctx, "es error response unmarshal failed: %v", err)
			return fmt.Errorf("es error response unmarshal failed: %s", resp.Status())
		}
		logs.ErrorContextf(ctx, "es UpdateByQuery failed [%s]: %v", resp.Status(), e)
		return fmt.Errorf("es UpdateByQuery failed [%s]: %v", resp.Status(), e)
	}
	return nil
}

func UpdateForestDescription(ctx *gin.Context, req *dtoforest.UpdateForestDescriptionRequest) (res *dtoforest.UpdateForestDescriptionResponse, err error) {
	res = &dtoforest.UpdateForestDescriptionResponse{}
	frs, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "[UpdateForestDescription] GetForestByID(%v) failed, err: %v", req.Request.ForestID, err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_get_forest_by_id_failed"
		return res, nil
	}

	frs.Description = req.Request.Description
	if err = dbutil.Knownow().WithContext(ctx).Save(frs).Error; err != nil {
		logs.ErrorContextf(ctx, "[UpdateForestDescription] Save(%v) failed, err: %v", frs, err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_update_forest_description_failed"
		return res, nil
	}

	return res, nil
}

// GetOriginResource 获取资源 url 列表
func GetOriginResource(ctx *gin.Context, req *dtoforest.GetOriginResourceRequest) (res *dtoforest.GetOriginResourceResponse, err error) {
	res = &dtoforest.GetOriginResourceResponse{}
	switch req.Request.ResourceType {
	case ResourceTypeForest:
		// Get all sub-resource of the forest ids
		fWithUrls, err := forest.GetForestFilePublicUrls(ctx, req.Request.ResourceIDs)
		if err != nil {
			logs.ErrorContextf(ctx, "[GetOriginResource] GetForestFilePublicUrls(%v) failed, err: %v", req.Request.ResourceIDs, err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_get_forest_public_urls_failed"
			return res, nil
		}

		for _, f := range fWithUrls {
			res.Response.Data = append(res.Response.Data, &dtoforest.Resource{
				ID: f.ID,
				Meta: map[string]interface{}{
					"forest":    f.KnownowForest,
					"file_list": f.Files,
				},
				ResourceType: ResourceTypeForest,
			})
		}
	}
	return res, nil
}
