package forestctl

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/services/svcdbforest"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"gorm.io/gorm"
)

// CreateForestDBInstance 创建数据库实例
// @Tags 数据库知识库
// @Summary 创建数据库实例
// @Description 创建数据库实例
// @Router /forest.CreateForestDBInstance [post]
// @Param user body CreateForestDBInstanceRequest true "入参"
// @Success 200 {object} CreateForestDBInstanceResponse "返回值"
func CreateForestDBInstance(ctx *gin.Context, req *CreateForestDBInstanceRequest, resp *CreateForestDBInstanceResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	// 检查知识库
	forestEntity, err := forest.NewForestDao().GetByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.CreateForestDBInstance] Failed to get forest, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed" // 获取知识库失败
		return
	}
	if forestEntity == nil || forestEntity.ID == 0 {
		logs.WarnContextf(ctx, "[forestctl.CreateForestDBInstance] Forest id: %d not found", req.Request.ForestID)
		resp.Code = errcode.ErrCode_NotFound
		resp.Message = "kecore_forest_not_found" // 知识库不存在
		return
	}
	if forestEntity.DataSourceType != foresttype.ForestDataSourceDB {
		logs.WarnContextf(ctx, "[forestctl.CreateForestDBInstance] Forest id: %d, type: %s, not a database forest", forestEntity.ID, forestEntity.DataSourceType)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_type" // 知识库类型错误
		return
	}

	// 检查是否已经有实例
	instanceCount, err := forest.NewForestDBInstanceDao().CountByCond(ctx, &forest.ForestDBInstanceCond{
		ForestID: req.Request.ForestID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.CreateForestDBInstance] Failed to count forest db instance, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_data_source_failed" // 创建数据源失败
		return
	}
	if instanceCount > 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_data_source_exists" // 数据源已存在
		return
	}

	var pluginDatabaseType dbplugins.DatabaseType
	switch forestEntity.DataSourceSubType {
	case foresttype.ForestDataSourceSubtypeMySQL:
		pluginDatabaseType = dbplugins.DatabaseTypeMySQL
	default:
		logs.WarnContextf(ctx, "[forestctl.CreateForestDBInstance] Forest id: %d, subtype: %s, not support", forestEntity.ID, forestEntity.DataSourceSubType)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_type" // 知识库类型错误
		return
	}

	// 检查实例是否有效
	dbPluginConfig := &dbplugins.PluginConfig{
		Credentials: &dbplugins.Credentials{
			Hostname: req.Request.Host,
			Port:     req.Request.Port,
			Username: req.Request.Username,
			Password: req.Request.Password,
			Database: req.Request.Database,
		},
	}
	isAvailable, err := dbutil.GetDBPluginEngine().ChoosePlugin(pluginDatabaseType).IsAvailable(ctx, dbPluginConfig)
	if !isAvailable {
		logs.ErrorContextf(ctx, "[forestctl.CreateForestDBInstance] Forest id: %d, instance: %s, not available", forestEntity.ID, req.Request.Host)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = err.Error()
		return
	}

	companyID, uin := runtime.CompanyID(ctx), runtime.Uin(ctx)

	// 创建示例、数据库、数据表
	instanceEntity := &foresttype.ForestDBInstance{
		ForestID:         req.Request.ForestID,
		CompanyID:        companyID,
		Uin:              uin,
		OwnershipType:    foresttype.DBInstanceOwnershipTypeExternal,
		InstanceType:     pluginDatabaseType,
		ConnectMode:      foresttype.DBInstanceConnectModeStandard,
		ConnectName:      forestEntity.Name,
		ConnectionStatus: foresttype.DBInstanceConnectionStatusValid,
		Host:             req.Request.Host,
		Port:             req.Request.Port,
		Username:         req.Request.Username,
		Password:         settings.EncryptSecret(req.Request.Password),
		Database:         req.Request.Database,
	}
	buildEntityReq := &svcdbforest.BuildEntityReq{
		ForestEntity:       forestEntity,
		CompanyID:          companyID,
		Uin:                uin,
		PluginDatabaseType: pluginDatabaseType,
		PluginConfig:       dbPluginConfig,
	}
	dbEntity, tableEntityList, err := svcdbforest.BuildEntity(ctx, buildEntityReq)
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.CreateForestDBInstance] Failed to build entity, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_data_source_failed" // 创建数据源失败
		return
	}

	txErr := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建实例
		if err := forest.NewForestDBInstanceDao().WithTx(tx).Insert(ctx, instanceEntity); err != nil {
			logs.ErrorContextf(ctx, "[forestctl.CreateForestDBInstance] Failed to create forest db instance, err: %s", err)
			return err
		}
		// 创建数据库
		dbEntity.DBInstanceID = instanceEntity.ID
		if err := forest.NewForestDBDao().WithTx(tx).Insert(ctx, dbEntity); err != nil {
			logs.ErrorContextf(ctx, "[forestctl.CreateForestDBInstance] Failed to create forest db, err: %s", err)
			return err
		}
		// 创建数据表
		for i := range tableEntityList {
			tableEntityList[i].DBInstanceID = instanceEntity.ID
			tableEntityList[i].DBID = dbEntity.ID
		}
		if len(tableEntityList) > 0 {
			if err := forest.NewForestTableDao().WithTx(tx).BatchInsert(ctx, tableEntityList); err != nil {
				logs.ErrorContextf(ctx, "[forestctl.CreateForestDBInstance] Failed to create forest table, err: %s", err)
				return err
			}
		}
		// 更新知识库状态，如果有一个文件解析成功，则知识库状态更新为解析成功
		forestUpdateMap := map[string]interface{}{
			"knowledge_status": foresttype.TaskStatusSuccess,
		}
		if err := forest.NewForestDao().WithTx(tx).UpdateMap(ctx, req.Request.ForestID, forestUpdateMap); err != nil {
			logs.ErrorContextf(ctx, "[forestctl.CreateForestDBInstance] update forest success error: %v", err)
			return err
		}
		return nil
	})
	if txErr != nil {
		logs.ErrorContextf(ctx, "[forestctl.CreateForestDBInstance] execute transaction failed, err: %s", txErr)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_data_source_failed" // 创建数据源失败
		return
	}
	resp.Response.ForestDBInstanceID = instanceEntity.ID
}

// TestForestDBInstanceConnection 测试数据库连接
// @Tags 数据库知识库
// @Summary 测试数据库连接
// @Description 测试数据库连接
// @Router /forest.TestForestDBInstanceConnection [post]
// @Param user body TestForestDBInstanceConnectionRequest true "入参"
// @Success 200 {object} TestForestDBInstanceConnectionResponse "返回值"
func TestForestDBInstanceConnection(ctx *gin.Context, req *TestForestDBInstanceConnectionRequest, resp *TestForestDBInstanceConnectionResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	// 检查知识库
	forestEntity, err := forest.NewForestDao().GetByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.TestForestDBInstanceConnection] Failed to get forest, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed" // 获取知识库失败
		return
	}
	if forestEntity == nil || forestEntity.ID == 0 {
		logs.WarnContextf(ctx, "[forestctl.TestForestDBInstanceConnection] Forest id: %d not found", req.Request.ForestID)
		resp.Code = errcode.ErrCode_NotFound
		resp.Message = "kecore_forest_not_found" // 知识库不存在
		return
	}

	dbPluginConfig := &dbplugins.PluginConfig{
		Credentials: &dbplugins.Credentials{
			Hostname: req.Request.Host,
			Port:     req.Request.Port,
			Username: req.Request.Username,
			Password: req.Request.Password,
			Database: req.Request.Database,
		},
	}
	var databaseType dbplugins.DatabaseType
	switch forestEntity.DataSourceSubType {
	case foresttype.ForestDataSourceSubtypeMySQL:
		databaseType = dbplugins.DatabaseTypeMySQL
	default:
		logs.ErrorContextf(ctx, "[forestctl.TestForestDBInstanceConnection] Unsupported database type: %s", forestEntity.DataSourceSubType)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_type" // 知识库类型错误
		return
	}
	isAvailable, err := dbutil.GetDBPluginEngine().ChoosePlugin(dbplugins.DatabaseType(databaseType)).IsAvailable(ctx, dbPluginConfig)
	if !isAvailable {
		logs.WarnContextf(ctx, "[forestctl.TestForestDBInstanceConnection] Forest db instance id: %d is not available", req.Request.ForestID)
		resp.Response.ConnectionStatus = foresttype.DBInstanceConnectionStatusInvalid
		resp.Response.FailureReason = "kecore_connection_failed" // 连接失败
		return
	}
	resp.Response.ConnectionStatus = foresttype.DBInstanceConnectionStatusValid
}

// ModifyForestDBInstance 修改数据库实例
// @Tags 数据库知识库
// @Summary 修改数据库实例
// @Description 修改数据库实例
// @Router /forest.ModifyForestDBInstance [post]
// @Param user body ModifyForestDBInstanceRequest true "入参"
// @Success 200 {object} ModifyForestDBInstanceResponse "返回值"
func ModifyForestDBInstance(ctx *gin.Context, req *ModifyForestDBInstanceRequest, resp *ModifyForestDBInstanceResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	forestEntity, err := forest.NewForestDao().GetByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.ModifyForestDBInstance] Failed to get forest, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed" // 获取知识库失败
		return
	}
	if forestEntity == nil || forestEntity.ID == 0 {
		logs.WarnContextf(ctx, "[forestctl.ModifyForestDBInstance] Forest id: %d not found", req.Request.ForestID)
		resp.Code = errcode.ErrCode_NotFound
		resp.Message = "kecore_forest_not_found" // 知识库不存在
		return
	}

	instanceEntity, err := forest.NewForestDBInstanceDao().GetByCond(ctx, &forest.ForestDBInstanceCond{
		ForestID: req.Request.ForestID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.ModifyForestDBInstance] Failed to get forest db instance, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_db_instance_failed" // 获取数据库实例失败
		return
	}
	if instanceEntity == nil || instanceEntity.ID == 0 {
		logs.WarnContextf(ctx, "[forestctl.ModifyForestDBInstance] Forest id: %d not found", req.Request.ForestID)
		resp.Code = errcode.ErrCode_NotFound
		resp.Message = "kecore_db_instance_not_found" // 数据库实例不存在
		return
	}

	var pluginDatabaseType dbplugins.DatabaseType
	switch forestEntity.DataSourceSubType {
	case foresttype.ForestDataSourceSubtypeMySQL:
		pluginDatabaseType = dbplugins.DatabaseTypeMySQL
	default:
		logs.WarnContextf(ctx, "[forestctl.ModifyForestDBInstance] Forest id: %d, subtype: %s, not support", forestEntity.ID, forestEntity.DataSourceSubType)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_type" // 知识库类型错误
		return
	}

	dbPluginConfig := &dbplugins.PluginConfig{
		Credentials: &dbplugins.Credentials{
			Hostname: req.Request.Host,
			Port:     req.Request.Port,
			Username: req.Request.Username,
			Password: req.Request.Password,
			Database: req.Request.Database,
		},
	}
	isAvailable, err := dbutil.GetDBPluginEngine().ChoosePlugin(dbplugins.DatabaseType(instanceEntity.InstanceType)).IsAvailable(ctx, dbPluginConfig)
	if !isAvailable {
		logs.WarnContextf(ctx, "[forestctl.ModifyForestDBInstance] Forest id: %d is not available", req.Request.ForestID)
		resp.Code = errcode.ErrCode_NotFound
		resp.Message = "kecore_connection_failed" // 连接失败
		return
	}

	buildEntityReq := &svcdbforest.BuildEntityReq{
		ForestEntity:       forestEntity,
		CompanyID:          forestEntity.CompanyID,
		Uin:                forestEntity.Uin,
		PluginConfig:       dbPluginConfig,
		PluginDatabaseType: pluginDatabaseType,
	}

	dbEntity, tableEntityList, err := svcdbforest.BuildEntity(ctx, buildEntityReq)
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.ModifyForestDBInstance] Failed to build forest db, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_build_db_failed" // 构建数据库失败
		return
	}

	instanceUpdateMap := map[string]interface{}{
		"host":     req.Request.Host,
		"port":     req.Request.Port,
		"username": req.Request.Username,
		"password": settings.EncryptSecret(req.Request.Password),
		"database": req.Request.Database,
	}

	txErr := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := forest.NewForestDBInstanceDao().UpdateMap(ctx, instanceEntity.ID, instanceUpdateMap); err != nil {
			logs.ErrorContextf(ctx, "[forestctl.ModifyForestDBInstance] Failed to update forest db instance, err: %s", err)
			return err
		}
		// 删除旧的数据库和数据表
		if err := forest.NewForestDBDao().DeleteByForestID(ctx, forestEntity.ID); err != nil {
			logs.ErrorContextf(ctx, "[forestctl.ModifyForestDBInstance] Failed to delete forest db, err: %s", err)
			return err
		}
		if err := forest.NewForestTableDao().DeleteByForestID(ctx, forestEntity.ID); err != nil {
			logs.ErrorContextf(ctx, "[forestctl.ModifyForestDBInstance] Failed to delete forest table, err: %s", err)
			return err
		}

		// 创建新的数据库和数据表
		dbEntity.DBInstanceID = instanceEntity.ID
		if err := forest.NewForestDBDao().Insert(ctx, dbEntity); err != nil {
			logs.ErrorContextf(ctx, "[forestctl.ModifyForestDBInstance] Failed to create forest db, err: %s", err)
			return err
		}
		for i := range tableEntityList {
			tableEntityList[i].DBID = dbEntity.ID
			tableEntityList[i].DBInstanceID = instanceEntity.ID
		}
		if len(tableEntityList) > 0 {
			if err := forest.NewForestTableDao().BatchInsert(ctx, tableEntityList); err != nil {
				logs.ErrorContextf(ctx, "[forestctl.ModifyForestDBInstance] Failed to create forest table, err: %s", err)
				return err
			}
		}

		return nil
	})
	if txErr != nil {
		logs.ErrorContextf(ctx, "[forestctl.ModifyForestDBInstance] execute transaction failed, err: %s", txErr)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_modify_data_source_failed" // 修改数据源失败
		return
	}
	resp.Response.ForestDBInstanceID = instanceEntity.ID
}

// GetForestDBInstance 数据库实例详情
// @Tags 数据库知识库
// @Summary 数据库实例详情
// @Description 数据库实例详情
// @Router /forest.GetForestDBInstance [post]
// @Param user body GetForestDBInstanceRequest true "入参"
// @Success 200 {object} GetForestDBInstanceResponse "返回值"
func GetForestDBInstance(ctx *gin.Context, req *GetForestDBInstanceRequest, resp *GetForestDBInstanceResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	instanceEntity, err := forest.NewForestDBInstanceDao().GetByCond(ctx, &forest.ForestDBInstanceCond{
		ForestID: req.Request.ForestID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.GetForestDBInstance] Failed to get forest db instance, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_db_instance_failed" // 获取数据库实例失败
		return
	}
	if instanceEntity == nil || instanceEntity.ID == 0 {
		logs.WarnContextf(ctx, "[forestctl.GetForestDBInstance] Forest id: %d not found", req.Request.ForestID)
		// resp.Code = errcode.ErrCode_NotFound
		// resp.Message = "数据库实例不存在"
		return
	}
	baseInfo := ForestDBInstanceBaseInfo{
		ForestID: instanceEntity.ForestID,
		Host:     instanceEntity.Host,
		Port:     instanceEntity.Port,
		Username: instanceEntity.Username,
		Database: instanceEntity.Database,
	}
	resp.Response.ForestDBInstanceID = instanceEntity.ID
	resp.Response.ForestDBInstanceBaseInfo = baseInfo
}

// ListForestDB 获取数据库列表
// @Tags 数据库知识库
// @Summary 获取数据库列表
// @Description 获取数据库列表
// @Router /forest.ListForestDB [post]
// @Param user body ListForestDBRequest true "入参"
// @Success 200 {object} ListForestDBResponse "返回值"
func ListForestDB(ctx *gin.Context, req *ListForestDBRequest, resp *ListForestDBResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	// 检查知识库
	forestEntity, err := forest.NewForestDao().GetByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.ListForestDB] Failed to get forest, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed" // 获取知识库失败
		return
	}
	if forestEntity == nil || forestEntity.ID == 0 {
		logs.WarnContextf(ctx, "[forestctl.ListForestDB] Forest id: %d not found", req.Request.ForestID)
		resp.Code = errcode.ErrCode_NotFound
		resp.Message = "kecore_forest_not_found" // 知识库不存在
		return
	}

	instanceEntity, err := forest.NewForestDBInstanceDao().GetByCond(ctx, &forest.ForestDBInstanceCond{
		ForestID: req.Request.ForestID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.ListForestDB] Failed to get forest db instance, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_db_instance_failed" // 获取数据库实例失败
		return
	}
	if instanceEntity == nil || instanceEntity.ID == 0 {
		logs.WarnContextf(ctx, "[forestctl.ListForestDB] Forest id: %d not found", req.Request.ForestID)
		resp.Response.DBList = make([]ListForestDBListItem, 0)
		resp.Response.Offset = req.Request.Offset
		resp.Response.Limit = req.Request.Limit
		return
	}

	// 查询对应的数据库
	dbEntityList, total, err := forest.NewForestDBDao().GetPageListByCond(ctx, &forest.ForestDBCond{
		ForestID: req.Request.ForestID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.ListForestDB] Failed to get forest db, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_db_failed" // 获取数据库失败
		return
	}
	if len(dbEntityList) == 0 {
		resp.Response.DBList = make([]ListForestDBListItem, 0)
		resp.Response.Offset = req.Request.Offset
		resp.Response.Limit = req.Request.Limit
		return
	}
	var schemas []string
	dbEntityMap := make(map[string]foresttype.ForestDB)
	for _, v := range dbEntityList {
		schemas = append(schemas, v.DBName)
		dbEntityMap[v.DBName] = v
	}

	dbPluginConfig := &dbplugins.PluginConfig{
		Credentials: &dbplugins.Credentials{
			ConnectionID: fmt.Sprintf("%d", instanceEntity.ID),
			Hostname:     instanceEntity.Host,
			Port:         instanceEntity.Port,
			Username:     instanceEntity.Username,
			Password:     settings.DecryptSecret(instanceEntity.Password),
			Database:     instanceEntity.Database,
		},
	}
	opt := &dbplugins.QueryOption{
		Filters: []dbplugins.Filter{
			{
				Key:    dbplugins.FilterKeySchemas,
				Values: schemas,
			},
		},
	}
	storageGroupsRes, err := dbutil.GetDBPluginEngine().ChoosePlugin(dbplugins.DatabaseType(instanceEntity.InstanceType)).GetStorageGroups(ctx, dbPluginConfig, opt)
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.ListForestDB] Failed to get storage groups, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_storage_groups_failed" // 获取存储组失败
		return
	}

	// 查询对应的数据库
	list := make([]ListForestDBListItem, 0, len(storageGroupsRes.List))
	for _, v := range storageGroupsRes.List {
		dbEntity := dbEntityMap[v.Name]
		logs.DebugContextf(ctx, "[forestctl.ListForestDB] dbEntity: %s", logs.JSON(dbEntity))
		item := ListForestDBListItem{
			ForestID:           dbEntity.ForestID,
			ForestDBInstanceID: dbEntity.DBInstanceID,
			InstanceType:       instanceEntity.InstanceType,
			ForestDbID:         dbEntity.ID,
			DBName:             v.Name,
			Enable:             dbEntity.Enable,
		}
		for _, attr := range v.Attributes {
			switch attr.Key {
			case dbplugins.RecordKeyDataSize:
				dataSize := utils.VToUint64(attr.Value)
				dataSizeMB, err := utils.ConvertBytes(uint(dataSize), utils.UnitMB)
				if err != nil {
					logs.ErrorContextf(ctx, "[forestctl.ListForestDB] Failed to convert bytes, err: %s", err)
				}
				item.DataSize = dataSizeMB
			case dbplugins.RecordKeyRowCount:
				item.DataRows = uint(utils.VToUint64(attr.Value))
			}
		}
		list = append(list, item)
	}
	resp.Response.Total = total
	resp.Response.DBList = list
	resp.Response.Offset = req.Request.Offset
	resp.Response.Limit = req.Request.Limit
}

// ListForestTable 获取数据表列表
// @Tags 数据库知识库
// @Summary 获取数据表列表
// @Description 获取数据表列表
// @Router /forest.ListForestTable [post]
// @Param user body ListForestTableRequest true "入参"
// @Success 200 {object} ListForestTableResponse "返回值"
func ListForestTable(ctx *gin.Context, req *ListForestTableRequest, resp *ListForestTableResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	// 获取实例
	instanceEntity, err := forest.NewForestDBInstanceDao().GetByCond(ctx, &forest.ForestDBInstanceCond{
		ForestID: req.Request.ForestID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.ListForestTable] Failed to get forest db instance, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_db_instance_failed" // 获取数据库实例失败
		return
	}

	dbEntity, err := forest.NewForestDBDao().GetByCond(ctx, &forest.ForestDBCond{
		ForestID:     req.Request.ForestID,
		ForestDBName: req.Request.ForestDbName,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.ListForestTable] Failed to get forest db, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_db_failed" // 获取数据库失败
		return
	}

	dbPluginConfig := &dbplugins.PluginConfig{
		Credentials: &dbplugins.Credentials{
			ConnectionID: fmt.Sprintf("%d", instanceEntity.ID),
			Hostname:     instanceEntity.Host,
			Port:         instanceEntity.Port,
			Username:     instanceEntity.Username,
			Password:     settings.DecryptSecret(instanceEntity.Password),
			Database:     instanceEntity.Database,
		},
	}
	queryOption := &dbplugins.QueryOption{
		Limit:      req.Request.Limit,
		Offset:     req.Request.Offset,
		OrderRules: req.Request.OrderBy,
	}
	for _, filter := range req.Request.Filters {
		switch filter.Field {
		case "forest_table_name":
			queryOption.Filters = append(queryOption.Filters, dbplugins.Filter{
				Key:    dbplugins.FilterKeyTable,
				Values: filter.Value,
			})
		}
	}
	storageUnitsRes, err := dbutil.GetDBPluginEngine().ChoosePlugin(dbplugins.DatabaseType(instanceEntity.InstanceType)).GetStorageUnits(ctx, dbPluginConfig, req.Request.ForestDbName, queryOption)
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.ListForestTable] Failed to get storage units, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_storage_units_failed" // 获取数据表失败
		return
	}

	list := make([]ListForestTableItem, 0, len(storageUnitsRes.List))
	for _, v := range storageUnitsRes.List {
		item := ListForestTableItem{
			ForestID:           dbEntity.ForestID,
			ForestDbID:         dbEntity.ID,
			ForestDBInstanceID: dbEntity.DBInstanceID,
			InstanceType:       instanceEntity.InstanceType,
			ForestTableName:    v.Name,
			Enable:             dbEntity.Enable,
		}
		for _, attr := range v.TableAttributes {
			switch attr.Key {
			case dbplugins.RecordKeyDataSize:
				dataSize := utils.VToUint64(attr.Value)
				dataSizeMB, err := utils.ConvertBytes(uint(dataSize), utils.UnitMB)
				if err != nil {
					logs.ErrorContextf(ctx, "[forestctl.ListForestTable] Failed to convert bytes, dataSize: %d, err: %s", dataSize, err)
				}
				item.DataSize = dataSizeMB
			case dbplugins.RecordKeyRowCount:
				item.DataRows = uint(utils.VToUint64(attr.Value))
			}
		}
		list = append(list, item)
	}
	resp.Response.TableList = list
	resp.Response.Total = storageUnitsRes.Total
	resp.Response.Offset = req.Request.Offset
	resp.Response.Limit = req.Request.Limit
}

// GetForestTableMetadata 获取数据表元数据
// @Tags 数据库知识库
// @Summary 获取数据表元数据
// @Description 获取数据表元数据
// @Router /forest.GetForestTableMetadata [post]
// @Param user body GetForestTableMetadataRequest true "入参"
// @Success 200 {object} GetForestTableMetadataResponse "返回值"
func GetForestTableMetadata(ctx *gin.Context, req *GetForestTableMetadataRequest, resp *GetForestTableMetadataResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	// 查询实例
	instanceEntity, err := forest.NewForestDBInstanceDao().GetByCond(ctx, &forest.ForestDBInstanceCond{
		ForestID: req.Request.ForestID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.GetForestTableMetadata] Failed to get forest db instance, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_db_instance_failed" // 获取数据库实例失败
		return
	}
	dbPluginConfig := &dbplugins.PluginConfig{
		Credentials: &dbplugins.Credentials{
			ConnectionID: fmt.Sprintf("%d", instanceEntity.ID),
			Hostname:     instanceEntity.Host,
			Port:         instanceEntity.Port,
			Username:     instanceEntity.Username,
			Password:     settings.DecryptSecret(instanceEntity.Password),
			Database:     instanceEntity.Database,
		},
	}
	storageUnitsRes, err := dbutil.GetDBPluginEngine().ChoosePlugin(dbplugins.DatabaseType(instanceEntity.InstanceType)).GetStorageUnits(ctx, dbPluginConfig, instanceEntity.Database, &dbplugins.QueryOption{
		Filters: []dbplugins.Filter{
			{
				Key:    dbplugins.FilterKeyTables,
				Values: []string{req.Request.ForestTableName},
			},
		},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[forestctl.GetForestTableMetadata] Failed to get storage units, err: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_storage_units_failed" // 获取数据表失败
		return
	}
	if len(storageUnitsRes.List) == 0 {
		logs.WarnContextf(ctx, "[forestctl.GetForestTableMetadata] Forest table : %s not found", req.Request.ForestTableName)
		resp.Code = errcode.ErrCode_NotFound
		resp.Message = "kecore_table_not_found" // 数据表不存在
		return
	}
	storageUnit := storageUnitsRes.List[0]
	list := make([]GetForestTableMetadataColumnItem, 0)
	for _, v := range storageUnit.ColumnAttributes {
		item := GetForestTableMetadataColumnItem{
			ColumnName: v.Key,
		}
		for extraKey, extraValue := range v.Extra {
			switch extraKey {
			case dbplugins.RecordKeyColumnComment:
				item.ColumnComment = extraValue
			case dbplugins.RecordKeyColumnType:
				item.ColumnType = extraValue
			}
		}
		list = append(list, item)
	}
	resp.Response.ColumnList = list
}
