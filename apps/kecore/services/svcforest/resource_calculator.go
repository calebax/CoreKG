package svcforest

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"github.com/ygpkg/yg-go/types"
)

type ResourceCalcType string

const (
	ResourceCalcTypeFile   ResourceCalcType = "file"
	ResourceCalcTypeMysql  ResourceCalcType = "mysql"
	ResourceCalcTypeQAPair ResourceCalcType = "qa_pair"
)

type ResourceCalculator interface {
	Metrics(ctx context.Context, forestID uint) (*ResourceCalcMetrics, error)
}

type ResourceCalcMetrics struct {
	Count     int64
	SizeBytes int64
}

func NewResourceCalculator(resourceCalcType ResourceCalcType) ResourceCalculator {
	switch resourceCalcType {
	case ResourceCalcTypeFile:
		return &resourceFileCalculator{}
	case ResourceCalcTypeMysql:
		return &resourceMysqlCalculator{}
	case ResourceCalcTypeQAPair:
		return &resourceQAPairCalculator{}
	default:
		return &resourceFileCalculator{}
	}
}

type resourceFileCalculator struct{}

func (f *resourceFileCalculator) Metrics(ctx context.Context, forestID uint) (*ResourceCalcMetrics, error) {
	cond := &forest.ForestFileCond{
		ForestIDs: []uint{forestID},
		IsDir:     types.False,
	}
	fileCount, err := forest.NewForestFileDao().CountByCond(ctx, cond)
	if err != nil {
		logs.ErrorContextf(ctx, "[resourceFileCalculator.Metrics] count file failed, forestID: %d, err: %v", forestID, err)
		return nil, err
	}

	fileSize, err := forest.NewForestFileDao().StatSizeByCond(ctx, cond)
	if err != nil {
		logs.ErrorContextf(ctx, "[resourceFileCalculator.Metrics] stat file size failed, forestID: %d, err: %v", forestID, err)
		return nil, err
	}
	return &ResourceCalcMetrics{
		Count:     fileCount,
		SizeBytes: fileSize,
	}, nil
}

type resourceMysqlCalculator struct{}

func (m *resourceMysqlCalculator) Metrics(ctx context.Context, forestID uint) (*ResourceCalcMetrics, error) {
	tableCount, err := forest.NewForestTableDao().CountByCond(ctx, &forest.ForestTableCond{
		ForestIDs: []uint{forestID},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[resourceMysqlCalculator.Metrics] count table failed, forestID: %d, err: %v", forestID, err)
		return nil, err
	}
	instanceEntity, err := forest.NewForestDBInstanceDao().GetByCond(ctx, &forest.ForestDBInstanceCond{
		ForestID: forestID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[resourceMysqlCalculator.Metrics] get instance failed, forestID: %d, err: %v", forestID, err)
		return nil, err
	}
	if instanceEntity == nil || instanceEntity.ID == 0 {
		logs.ErrorContextf(ctx, "[resourceMysqlCalculator.Metrics] instance not found, forestID: %d", forestID)
		return nil, fmt.Errorf("instance not found, forestID: %d", forestID)
	}
	dbEntityList, err := forest.NewForestDBDao().GetListByCond(ctx, &forest.ForestDBCond{
		ForestID: forestID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[resourceMysqlCalculator.Metrics] get db failed, forestID: %d, err: %v", forestID, err)
		return nil, err
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
		logs.ErrorContextf(ctx, "[resourceMysqlCalculator.Metrics] get storage groups failed, forestID: %d, err: %v", forestID, err)
		return nil, err
	}
	var totalSize int64
	for _, v := range storageGroupsRes.List {
		for _, attr := range v.Attributes {
			switch attr.Key {
			case dbplugins.RecordKeyDataSize:
				dataSize := utils.VToUint64(attr.Value)
				totalSize += int64(dataSize)
			}
		}
	}
	return &ResourceCalcMetrics{
		Count:     tableCount,
		SizeBytes: totalSize,
	}, nil
}

type resourceQAPairCalculator struct{}

func (q *resourceQAPairCalculator) Metrics(ctx context.Context, forestID uint) (*ResourceCalcMetrics, error) {
	forestEntity, err := forest.NewForestDao().GetByID(ctx, forestID)
	if err != nil {
		logs.ErrorContextf(ctx, "[resourceQAPairCalculator.Metrics] get forest failed, forestID: %d, err: %v", forestID, err)
		return nil, err
	}
	if forestEntity == nil || forestEntity.ID == 0 {
		logs.ErrorContextf(ctx, "[resourceQAPairCalculator.Metrics] forest not found, forestID: %d", forestID)
		return nil, fmt.Errorf("forest not found, forestID: %d", forestID)
	}
	esIndex := forestEntity.EsIndex()
	countList, err := forest.GetQuestionCountsByForests(ctx, esIndex, []uint{forestID})
	if err != nil {
		logs.ErrorContextf(ctx, "[resourceQAPairCalculator.Metrics] get question counts failed, forestID: %d, err: %v", forestID, err)
		return nil, err
	}
	var totalCount int64
	for _, v := range countList {
		totalCount += int64(v.Count)
	}
	return &ResourceCalcMetrics{
		Count: totalCount,
	}, nil

}
