package svcdbforest

import (
	"context"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
)

func BuildEntity(ctx context.Context, req *BuildEntityReq) (*foresttype.ForestDB, foresttype.ForestTableList, error) {

	database := req.PluginConfig.Credentials.Database

	storageGroupsRes, err := dbutil.GetDBPluginEngine().ChoosePlugin(req.PluginDatabaseType).GetStorageGroups(ctx, req.PluginConfig, &dbplugins.QueryOption{
		Filters: []dbplugins.Filter{
			{
				Key:    dbplugins.FilterKeySchemas,
				Values: []string{database},
			},
		},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[svcdbforest.BuildEntity] Failed to get storage groups, err: %s", err)
		return nil, nil, err
	}
	var storageGroup dbplugins.StorageGroup
	if len(storageGroupsRes.List) > 0 {
		storageGroup = storageGroupsRes.List[0]
	}
	storageUnitsRes, err := dbutil.GetDBPluginEngine().ChoosePlugin(req.PluginDatabaseType).GetStorageUnits(ctx, req.PluginConfig, database, &dbplugins.QueryOption{})
	if err != nil {
		logs.ErrorContextf(ctx, "[svcdbforest.BuildEntity] Failed to get storage units, err: %s", err)
		return nil, nil, err
	}
	storageUnitMap := make(map[string]dbplugins.StorageUnit)
	for _, v := range storageUnitsRes.List {
		storageUnitMap[v.Name] = v
	}
	var tableEntityList foresttype.ForestTableList
	for _, v := range storageUnitsRes.List {
		item := foresttype.ForestTable{
			Tablename:   v.Name,
			ForestID:    req.ForestEntity.ID,
			ColumnCount: uint(len(storageUnitMap[v.Name].ColumnAttributes)),
			Uin:         req.Uin,
			SyncedAt:    time.Now(),
		}
		for _, attr := range storageUnitMap[v.Name].TableAttributes {
			switch attr.Key {
			case dbplugins.RecordKeyTotalSize:
				item.Size = uint(utils.VToUint64(attr.Value))
			case dbplugins.RecordKeyRowCount:
				item.RowCount = uint(utils.VToUint64(attr.Value))
			}
		}

		tableEntityList = append(tableEntityList, item)
	}

	dbMetadata := foresttype.ForestDBMeta{}
	switch req.PluginDatabaseType {
	case dbplugins.DatabaseTypeMySQL:
		mysqlMetadata := foresttype.ForestDBMysqlMeta{}
		for _, v := range storageGroup.Attributes {
			switch v.Key {
			case dbplugins.RecordKeyCharset:
				mysqlMetadata.Charset = v.Value
			case dbplugins.RecordKeyCollation:
				mysqlMetadata.Collate = v.Value
			}
		}
		dbMetadata.Mysql = mysqlMetadata
	}

	dbEntity := &foresttype.ForestDB{
		ForestID:  req.ForestEntity.ID,
		CompanyID: req.CompanyID,
		Uin:       req.Uin,
		DBName:    req.PluginConfig.Credentials.Database,
		DBMeta:    dbMetadata,
		SyncedAt:  time.Now(),
	}
	for _, v := range storageGroup.Attributes {
		switch v.Key {
		case dbplugins.RecordKeyTotalSize:
			dbEntity.Size = uint(utils.VToUint64(v.Value))
		case dbplugins.RecordKeyRowCount:
			dbEntity.RowCount = uint(utils.VToUint64(v.Value))
		}
	}

	return dbEntity, tableEntityList, nil
}
