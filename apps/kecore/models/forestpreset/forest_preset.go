package forestpreset

import (
	"context"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/kecore/models/nbgraph"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// PresetForest 预置知识森林
func PresetForests(ctx context.Context, compid, uin uint) error {
	// 没有创建过知识森林会创建几个默认的
	var count int64
	if err := dbutil.Knownow().Table(foresttype.TableNameKnownowForest).
		Where("uin = ?", uin).
		Unscoped().Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	forest_list := []string{"英文文献", "产品说明书", "法律条款"}
	var preset_uin uint = 159
	for _, name := range forest_list {
		frst, files, err := GetPresetForestInfo(ctx, preset_uin, name)
		if err != nil {
			return err
		}
		if frst == nil {
			return nil
		}
		db := dbutil.Knownow()
		// 创建forest
		newForest := &foresttype.KnownowForest{
			CompanyID:       compid,
			Uin:             uin,
			Name:            frst.Name,
			KnowledgeStatus: frst.KnowledgeStatus,
		}
		if err := db.Create(newForest).Error; err != nil {
			return err
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := CopyForest(ctx, frst, newForest); err != nil {
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
				}
				if err := tx.Create(newFile).Error; err != nil {
					return err
				}
				err = CopyFile(ctx, file, newFile)
				if err != nil {
					return err
				}
				nbgraph.TaskCallBack(ctx, newFile)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// PresetForest 预置知识森林
func PresetForest(ctx context.Context, compid, uin uint) error {
	// 没有创建过知识森林会创建一个默认的
	// var count int64
	// if err := dbutil.Knownow().Table(foresttype.TableNameKnownowForest).
	// 	Where("uin = ?", uin).
	// 	Unscoped().Count(&count).Error; err != nil {
	// 	return err
	// }
	// if count > 0 {
	// 	return nil
	// }

	// forestName := "技术协议"
	// forestName := "技术方案"
	// forestName := "测试大纲"
	// forestName := "SCA资料"
	forestName := "新华三Test"

	var preset_uin uint = 159

	frst, files, err := GetPresetForestInfo(ctx, preset_uin, forestName)
	if err != nil {
		return err
	}
	if frst == nil {
		return nil
	}
	db := dbutil.Knownow()
	// 创建forest
	newForest := &foresttype.KnownowForest{
		CompanyID:       compid,
		Uin:             uin,
		Name:            frst.Name,
		KnowledgeStatus: frst.KnowledgeStatus,
	}
	if err := db.Create(newForest).Error; err != nil {
		return err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := CopyForest(ctx, frst, newForest); err != nil {
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
			}
			if err := tx.Create(newFile).Error; err != nil {
				return err
			}
			err = CopyFile(ctx, file, newFile)
			if err != nil {
				return err
			}
			nbgraph.TaskCallBack(ctx, newFile)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// CopyFile 复制文件
func CopyFile(ctx context.Context, srcFile, dstFile *foresttype.KnownowForestFile) error {
	srcPath, err := srcFile.GetForestFilePath()
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestFilePath error %v", err)
		return nil
	}

	dstPath, err := dstFile.GetForestFilePath()
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestFilePath error %v", err)
		return nil
	}
	// 复制文件
	err = fs.Forest.CopyDir(*srcPath, *dstPath)
	if err != nil {
		logs.ErrorContextf(ctx, "[forestpreset] CopyPath failed ,err = %v", err)
		return err
	}
	// 复制算法文件
	srcAlgoDir := fs.FileFileAlgoPath(srcFile.GetAlgoFilePath(), srcFile.ID)
	dstAlgoDir := fs.FileFileAlgoPath(dstFile.GetAlgoFilePath(), dstFile.ID)

	logs.InfoContextf(ctx, "copy file algo: src:%v dst:%v", srcAlgoDir, dstAlgoDir)
	err = fs.Forest.CopyDir(srcAlgoDir, dstAlgoDir)
	if err != nil {
		logs.ErrorContextf(ctx, "[forestpreset] CopyPath failed ,err = %v", err)
		return err
	}

	return nil
}

// CopyForest 复制
func CopyForest(ctx context.Context, srcForest, dstForest *foresttype.KnownowForest) error {
	// 复制知识森林算法文件
	// 复制算法文件
	srcAlgoDir := fs.ForestKnowledgeDirPath(srcForest.Uin, srcForest.ID)
	dstAlgoDir := fs.ForestKnowledgeDirPath(dstForest.Uin, dstForest.ID)

	err := fs.Forest.CopyDir(srcAlgoDir, dstAlgoDir)
	if err != nil {
		logs.ErrorContextf(ctx, "[forestpreset] CopyPath failed ,err = %v", err)
		return err
	}

	return nil
}

// GetPresetForestInfo 查询预置的知识森林信息
func GetPresetForestInfo(ctx context.Context, uin uint, forestName string) (*foresttype.KnownowForest, []*foresttype.KnownowForestFile, error) {
	frst, err := forest.GetForestByName(ctx, uin, forestName)
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
