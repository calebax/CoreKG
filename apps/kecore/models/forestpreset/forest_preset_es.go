package forestpreset

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/nbgraph"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"gorm.io/gorm"
)

type PresetForestConf struct {
	ForestIDs []uint `yaml:"forest_ids"`
}

// PresetESForests 预置知识森林
func PresetESForests(ctx context.Context, compid, uin uint) error {
	// 组织下没有过知识库
	var count int64
	if err := dbutil.Knownow().Table(foresttype.TableNameKnownowForest).
		Where("company_id = ?", compid).
		Unscoped().Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	conf := &PresetForestConf{}
	err := settings.GetYaml("knowledge", "preset_forest", conf)
	if err != nil {
		logs.ErrorContextf(ctx, "get preset_forest config error: %v", err)
	}
	for _, forest_id := range conf.ForestIDs {
		frst, files, err := GetPresetForestInfoByID(ctx, forest_id)
		if err != nil {
			return err
		}
		if frst == nil {
			return nil
		}
		db := dbutil.Knownow()
		// key:源文件id，value:目标文件id
		fileidmap := make(map[uint]uint)
		fileidmap[0] = 0
		// 创建forest
		newForest := &foresttype.KnownowForest{
			CompanyID:       compid,
			Uin:             uin,
			Name:            frst.Name,
			KnowledgeStatus: frst.KnowledgeStatus,
			AvatarUrl:       frst.AvatarUrl,
			Description:     frst.Description,
			ForestType:      frst.ForestType,
			ConfigID:        frst.ConfigID,
			PublicScope:     foresttype.PublicScopeCompany,
			Count:           frst.Count,
			DiskStorage:     frst.DiskStorage,
		}

		if err := db.Create(newForest).Error; err != nil {
			logs.ErrorContextf(ctx, "Create newForest error: %v", err)
			return err
		}
		err = db.Transaction(func(tx *gorm.DB) error {

			rss := []*foresttype.KeResourceScope{
				{
					ResourceType: foresttype.ResourceTypeForest,
					ResourceID:   newForest.ID,
					ScopeType:    foresttype.ScopeTypeUser,
					ScopeID:      uin,
					Action:       foresttype.ActionManage,
				},
				{
					ResourceType: foresttype.ResourceTypeForest,
					ResourceID:   newForest.ID,
					ScopeType:    foresttype.ScopeTypeCompany,
					ScopeID:      uin,
					Action:       foresttype.ActionView,
				},
			}

			//preset manager
			if err := tx.CreateInBatches(&rss, len(rss)).Error; err != nil {
				logs.ErrorContextf(ctx, "Create newForest resource scope error: %v", err)
				return err
			}

			if err := CopyForest(ctx, frst, newForest); err != nil {
				logs.ErrorContextf(ctx, "CopyForest error: %v", err)
				return err
			}
			for _, file := range files {
				// 只复制一层目录
				if file.IsDir.Value() || file.ParentID != 0 || file.Depth > 1 {
					continue
				}
				newFile := &foresttype.KnownowForestFile{
					CompanyID:       compid,
					Uin:             uin,
					ForestID:        newForest.ID,
					IsDir:           file.IsDir,
					ParentID:        0,
					Name:            file.Name,
					Size:            file.Size,
					Ext:             file.Ext,
					ParentIDs:       "",
					Depth:           1,
					ParseStatus:     file.ParseStatus,
					GraphStatus:     file.GraphStatus,
					AnalysisStatus:  file.AnalysisStatus,
					KnowledgeStatus: file.KnowledgeStatus,
					CoreFileID:      file.CoreFileID,
					PriviewFileID:   file.PriviewFileID,
					PriviewExt:      file.PriviewExt,
					Status:          file.Status,
					DescStatus:      file.DescStatus,
				}
				if err := tx.Create(newFile).Error; err != nil {
					logs.ErrorContextf(ctx, "Create newFile error: %v", err)
					return err
				}
				err = CopyFile(ctx, file, newFile)
				if err != nil {
					logs.ErrorContextf(ctx, "CopyFile error: %v", err)
					return err
				}
				fmt.Println("newFile: ", newFile)
				fileidmap[file.ID] = newFile.ID
			}
			return nil
		})
		if err != nil {
			logs.ErrorContextf(ctx, "CopyForest error: %v", err)
			return err
		}
		logs.InfoContextf(ctx, "start CopyForestData fileidmap: ", fileidmap)
		// 复制es数据
		err = essearch.CopyForestData(ctx, forest_id, newForest, fileidmap)
		if err != nil {
			logs.ErrorContextf(ctx, "essearch.CopyForestData error: %v")
			return err
		}
		logs.InfoContextf(ctx, "start CopyForestGraph")
		err = nbgraph.CopyForestGraph(ctx, forest_id, newForest, fileidmap)
		if err != nil {
			logs.ErrorContextf(ctx, "nbgraph.CopyForestGraph error: %v", err)
			return err
		}
	}

	return nil
}

// GetPresetForestInfo 查询预置的知识森林信息
func GetPresetForestInfoByID(ctx context.Context, forest_id uint) (*foresttype.KnownowForest, []*foresttype.KnownowForestFile, error) {
	frst, err := forest.GetForestByID(ctx, forest_id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, []*foresttype.KnownowForestFile{}, nil
		}
		return nil, nil, err
	}
	files, err := forest.ListAllForestFile(frst.ID)
	if err != nil {
		return nil, nil, err
	}
	return frst, files, nil
}
