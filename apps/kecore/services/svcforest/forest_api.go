package svcforest

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/coze"
	"github.com/insmtx/corekg/apps/kecore/models/coretask"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/forestpreset"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"gorm.io/gorm"
)

var (
	ErrForestNameExists      = errors.New("forest name exists")
	ErrGetForestFailed       = errors.New("get forest failed")
	ErrModifyForestFailed    = errors.New("modify forest failed")
	ErrNoPermission          = errors.New("no permission")
	ErrForestInUse           = errors.New("forest in use")
	ErrStatusCheckFailed     = errors.New("status check failed")
	ErrGraphInfoFailed       = errors.New("graph info failed")
	ErrTaskRunning           = errors.New("task running")
	ErrDeleteForestFailed    = errors.New("delete forest failed")
	ErrCozeMappingFailed     = errors.New("coze mapping failed")
	ErrDeleteMappingFailed   = errors.New("delete mapping failed")
	ErrQueryForestListFailed = errors.New("query forest list failed")
)

type CreateForestRequest struct {
	Uin               uint
	CompanyID         uint
	Name              string
	AvatarURL         string
	Description       string
	PublicScope       foresttype.PublicScope
	ForestType        foresttype.ForestType
	DataSourceType    foresttype.ForestDataSourceType
	DataSourceSubtype foresttype.ForestDataSourceSubtype
}

type ListForestRequest struct {
	Uin               uint
	CompanyID         uint
	Query             apiobj.PageQuery
	PresetWhenListing bool
}

type DeleteForestRequest struct {
	Uin      uint
	ForestID uint
	Token    string
}

type UpdateForestRequest struct {
	ForestID    uint
	Name        *string
	AvatarURL   *string
	Description *string
}

func CreateForest(ctx context.Context, req *CreateForestRequest) (uint, error) {
	if forest.CheckForestNameExists(ctx, 0, req.Name, req.CompanyID) {
		return 0, ErrForestNameExists
	}

	forestEntity := &foresttype.KnownowForest{
		CompanyID:         req.CompanyID,
		Uin:               req.Uin,
		Name:              req.Name,
		AvatarUrl:         req.AvatarURL,
		Description:       req.Description,
		PublicScope:       req.PublicScope,
		ForestType:        req.ForestType,
		DataSourceType:    req.DataSourceType,
		DataSourceSubType: req.DataSourceSubtype,
	}

	resourceScopeEntity := foresttype.KeResourceScope{
		ResourceType: foresttype.ResourceTypeForest,
		ScopeType:    foresttype.ScopeTypeUser,
		ScopeID:      req.Uin,
		Action:       foresttype.ActionManage,
	}

	txErr := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := forest.NewForestDao().WithTx(tx).Insert(ctx, forestEntity); err != nil {
			logs.ErrorContextf(ctx, "[CreateForest] failed insert forest entity, err: %v", err)
			return err
		}
		resourceScopeEntity.ResourceID = forestEntity.ID
		if err := forest.NewKeResourceScopeDao().WithTx(tx).Insert(ctx, &resourceScopeEntity); err != nil {
			logs.ErrorContextf(ctx, "[CreateForest] failed insert resource scope entity, err: %v", err)
			return err
		}
		switch req.DataSourceType {
		case foresttype.ForestDataSourceExcel:
			mysqlConfig := &config.MysqlConfig{}
			if err := settings.GetYaml("knowledge", global.DBInstanceSystemSettingKey, mysqlConfig); err != nil {
				logs.ErrorContextf(ctx, "[CreateForest] failed get mysql config, err = %v", err)
				return fmt.Errorf("get mysql config failed, err = %v", err)
			}
			dbName := fmt.Sprintf("ke_excel_%d", req.CompanyID)
			dbInstanceEntity := &foresttype.ForestDBInstance{
				CompanyID:     req.CompanyID,
				ForestID:      forestEntity.ID,
				Uin:           req.Uin,
				OwnershipType: foresttype.DBInstanceOwnershipTypeSystem,
				InstanceType:  dbplugins.DatabaseTypeMySQL,
				Host:          mysqlConfig.Host,
				Port:          uint(mysqlConfig.Port),
				Username:      mysqlConfig.Username,
				Password:      settings.EncryptSecret(mysqlConfig.Password),
				Database:      dbName,
				ConnectMode:   foresttype.DBInstanceConnectModeStandard,
				ConnectName:   mysqlConfig.Host,
			}
			if err := forest.NewForestDBInstanceDao().WithTx(tx).Insert(ctx, dbInstanceEntity); err != nil {
				logs.ErrorContextf(ctx, "[CreateForest] failed insert db instance entity, err: %v", err)
				return err
			}

			companyDBEntity := &foresttype.KeCompanyDB{
				CompanyID:    req.CompanyID,
				DBInstanceID: dbInstanceEntity.ID,
				DBName:       dbName,
			}
			if err := forest.NewKeCompanyDBDao().WithTx(tx).Upsert(ctx, companyDBEntity); err != nil {
				logs.ErrorContextf(ctx, "[CreateForest] failed upsert company db entity, err: %v", err)
				return err
			}
			dbEntity := &foresttype.ForestDB{
				CompanyID:    req.CompanyID,
				ForestID:     forestEntity.ID,
				DBInstanceID: dbInstanceEntity.ID,
				DBName:       dbName,
				DBMeta: foresttype.ForestDBMeta{
					Mysql: foresttype.ForestDBMysqlMeta{
						Charset: foresttype.MysqlDefaultCharset,
						Collate: foresttype.MysqlDefaultCollate,
					},
				},
				SyncedAt: time.Now(),
			}
			if err := forest.NewForestDBDao().WithTx(tx).Insert(ctx, dbEntity); err != nil {
				logs.ErrorContextf(ctx, "[CreateForest] failed insert forest db entity, err: %v", err)
				return err
			}

			forestDB, err := dbutil.GetDB(foresttype.MysqlDefaultInstanceAlias, mysqlConfig.BuildDNS())
			if err != nil {
				logs.ErrorContextf(ctx, "[CreateForest] failed get db client, err: %v", err)
				return err
			}
			createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET %s COLLATE %s;", dbName, foresttype.MysqlDefaultCharset, foresttype.MysqlDefaultCollate)
			if err := forestDB.WithContext(ctx).Exec(createDBSQL).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		logs.ErrorContextf(ctx, "[CreateForest] failed running transaction : %v", txErr)
		return 0, txErr
	}

	return forestEntity.ID, nil
}

func ListForest(ctx context.Context, req *ListForestRequest) (*forest.ForestInfoItemList, error) {
	query := req.Query
	query.Uin = req.Uin
	query.CompanyID = req.CompanyID

	if req.PresetWhenListing {
		if err := forestpreset.PresetESForests(ctx, req.CompanyID, req.Uin); err != nil {
			logs.ErrorContextf(ctx, "ListForest PresetForest failed,err = %v", err)
		}
	}

	out := &forest.ForestInfoItemList{}
	if err := forest.QueryListForest(ctx, query, out); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryForestListFailed, err)
	}
	return out, nil
}

func UpdateForest(ctx context.Context, req *UpdateForestRequest) error {
	forestInfo, err := forest.GetForestByID(ctx, req.ForestID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrGetForestFailed, err)
	}

	nextName := forestInfo.Name
	nextAvatarURL := forestInfo.AvatarUrl
	nextDescription := forestInfo.Description
	if req.Name != nil {
		nextName = *req.Name
	}
	if req.AvatarURL != nil {
		nextAvatarURL = *req.AvatarURL
	}
	if req.Description != nil {
		nextDescription = *req.Description
	}

	if forestInfo.Name == nextName &&
		forestInfo.AvatarUrl == nextAvatarURL &&
		forestInfo.Description == nextDescription {
		return nil
	}

	if forest.CheckForestNameExists(ctx, forestInfo.ID, nextName, forestInfo.CompanyID) {
		return ErrForestNameExists
	}

	forestInfo.Name = nextName
	forestInfo.AvatarUrl = nextAvatarURL
	forestInfo.Description = nextDescription
	if err := forest.ModifyForest(ctx, forestInfo); err != nil {
		return fmt.Errorf("%w: %v", ErrModifyForestFailed, err)
	}
	return nil
}

func DeleteForest(ctx *gin.Context, req *DeleteForestRequest) error {
	forestInfo, err := forest.GetForestByID(ctx, req.ForestID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrGetForestFailed, err)
	}
	if !perm.HasManageAct(ctx, req.Uin, req.ForestID, foresttype.ResourceTypeForest) {
		return ErrNoPermission
	}

	if err = forest.DeleteForestStatusCheck(ctx, forestInfo.ID); err != nil {
		if errors.Is(err, forest.ErrHasRunningTask) {
			return ErrForestInUse
		}
		return fmt.Errorf("%w: %v", ErrStatusCheckFailed, err)
	}

	graphInfo, err := graph.GetForestGraph(ctx, forestInfo.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%w: %v", ErrGraphInfoFailed, err)
	}
	if err == nil && (graphInfo.Status == foresttype.GraphStatusRunning || graphInfo.Status == foresttype.GraphStatusPending) {
		return ErrTaskRunning
	}

	if err = dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		if graphInfo != nil {
			if err = graph.DeleteGraphTX(ctx, tx, graphInfo.ID); err != nil {
				return err
			}
			if err = coretask.DeleteTasksByGraphVersion(ctx, graphInfo.VersionID); err != nil {
				return err
			}
		}
		if err = coretask.DeleteTasksByForestID(ctx, forestInfo.ID); err != nil {
			return err
		}
		if err = forest.DeleteForest(ctx, tx, req.Uin, forestInfo.ID); err != nil {
			return err
		}
		return essearch.DeleteForest(ctx, forestInfo.EsIndex(), forestInfo.ID)
	}); err != nil {
		return fmt.Errorf("%w: %v", ErrDeleteForestFailed, err)
	}

	cozeMapping, err := chattype.GetCozeMappingByCoreKGID(ctx, req.ForestID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCozeMappingFailed, err)
	}
	cozeURL, err := settings.GetText("corekg", "coze_url")
	if err != nil {
		return nil
	}
	for _, item := range cozeMapping {
		if err = coze.DeleteCozeKnowledgeAPI(ctx, item.CozeID, req.Token, cozeURL); err != nil {
			logs.ErrorContextf(ctx, "DeleteCozeKnowledgeAPI(%v) failed, err %s", req.ForestID, err)
		}
	}
	if err = chattype.DeleteCozeMappingByCorekgID(ctx, req.ForestID); err != nil {
		return fmt.Errorf("%w: %v", ErrDeleteMappingFailed, err)
	}
	return nil
}
