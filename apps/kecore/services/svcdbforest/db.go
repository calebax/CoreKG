package svcdbforest

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
)

func syncMysqlDatabase(ctx *gin.Context, dbEntity *foresttype.ForestDB) error {
	instanceEntity, err := forest.NewForestDBInstanceDao().GetByCond(ctx, &forest.ForestDBInstanceCond{
		ForestID: dbEntity.ForestID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[syncMysqlDatabase] get db instance fail, db id: %d, err: %v", dbEntity.ID, err)
		return err
	}
	if instanceEntity == nil || instanceEntity.ID == 0 {
		logs.WarnContextf(ctx, "[syncMysqlDatabase] db instance not found, db id: %d", dbEntity.ID)
		return fmt.Errorf("db instance not found, db id: %d", dbEntity.ID)
	}

	forestEntity, err := forest.NewForestDao().GetByID(ctx, dbEntity.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "[syncMysqlDatabase] get forest fail, db id: %d, err: %v", dbEntity.ID, err)
		return err
	}
	if forestEntity == nil || forestEntity.ID == 0 {
		logs.WarnContextf(ctx, "[syncMysqlDatabase] forest not found, db id: %d", dbEntity.ID)
		return nil
	}

	oldTableEntityList, err := forest.NewForestTableDao().GetListByCond(ctx, &forest.ForestTableCond{
		ForestIDs: []uint{dbEntity.ForestID},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[syncMysqlDatabase] get forest table list fail, db id: %d, err: %v", dbEntity.ID, err)
		return err
	}
	enableStatusMap := make(map[string]types.Bool)
	for _, tableEntity := range oldTableEntityList {
		enableStatusMap[tableEntity.Tablename] = tableEntity.Enable
	}

	var pluginDatabaseType dbplugins.DatabaseType
	switch forestEntity.DataSourceSubType {
	case foresttype.ForestDataSourceSubtypeMySQL:
		pluginDatabaseType = dbplugins.DatabaseTypeMySQL
	default:
		logs.WarnContextf(ctx, "[syncMysqlDatabase] Forest id: %d, subtype: %s, not support", forestEntity.ID, forestEntity.DataSourceSubType)
		return nil
	}

	dbPluginConfig := &dbplugins.PluginConfig{
		Credentials: &dbplugins.Credentials{
			Hostname: instanceEntity.Host,
			Port:     instanceEntity.Port,
			Username: instanceEntity.Username,
			Password: settings.DecryptSecret(instanceEntity.Password),
			Database: instanceEntity.Database,
		},
	}
	isAvailable, err := dbutil.GetDBPluginEngine().ChoosePlugin(dbplugins.DatabaseType(instanceEntity.InstanceType)).IsAvailable(ctx, dbPluginConfig)
	if !isAvailable {
		logs.WarnContextf(ctx, "[syncMysqlDatabase] db instance not available, db id: %d, err: %v", dbEntity.ID, err)
		return fmt.Errorf("db instance not available, db id: %d", dbEntity.ID)
	}

	buildEntityReq := &BuildEntityReq{
		ForestEntity:       forestEntity,
		CompanyID:          forestEntity.CompanyID,
		Uin:                forestEntity.Uin,
		PluginConfig:       dbPluginConfig,
		PluginDatabaseType: pluginDatabaseType,
	}
	newDBEntity, tableEntityList, err := BuildEntity(ctx, buildEntityReq)
	if err != nil {
		logs.ErrorContextf(ctx, "[syncMysqlDatabase] build entity fail, db id: %d, err: %v", dbEntity.ID, err)
		return err
	}
	for i := range tableEntityList {
		tableEntityList[i].Enable = enableStatusMap[tableEntityList[i].Tablename]
	}

	txErr := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除旧的数据库和数据表
		if err := forest.NewForestDBDao().DeleteByForestID(ctx, forestEntity.ID); err != nil {
			logs.ErrorContextf(ctx, "[syncMysqlDatabase] Failed to delete forest db, err: %s", err)
			return err
		}
		if err := forest.NewForestTableDao().DeleteByForestID(ctx, forestEntity.ID); err != nil {
			logs.ErrorContextf(ctx, "[syncMysqlDatabase] Failed to delete forest table, err: %s", err)
			return err
		}

		// 创建新的数据库和数据表
		newDBEntity.DBInstanceID = instanceEntity.ID
		newDBEntity.Enable = dbEntity.Enable
		if err := forest.NewForestDBDao().Insert(ctx, newDBEntity); err != nil {
			logs.ErrorContextf(ctx, "[syncMysqlDatabase] Failed to create forest db, err: %s", err)
			return err
		}
		for i := range tableEntityList {
			tableEntityList[i].DBID = newDBEntity.ID
			tableEntityList[i].DBInstanceID = instanceEntity.ID
		}
		if len(tableEntityList) > 0 {
			if err := forest.NewForestTableDao().BatchInsert(ctx, tableEntityList); err != nil {
				logs.ErrorContextf(ctx, "[syncMysqlDatabase] Failed to create forest table, err: %s", err)
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		logs.ErrorContextf(ctx, "[syncMysqlDatabase] sync mysql database fail, db id: %d, err: %v", dbEntity.ID, txErr)
		return txErr
	}

	return nil
}
