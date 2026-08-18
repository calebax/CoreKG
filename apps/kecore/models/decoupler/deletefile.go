package decoupler

import (
	"fmt"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"gorm.io/gorm"
)

// DeleteFileOrDir 删除文件或目录
func DeleteFileOrDir(file *foresttype.KnownowForestFile) error {
	db := dbutil.Knownow()
	// 删除子文件和子目录
	err := db.Where("parent_ids LIKE ?", fmt.Sprintf("%s%d/", file.ParentIDs, file.ID)+"/%").
		Delete(&foresttype.KnownowForestFile{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete children files or directories: %v", err)
	}
	// 删除当前文件或目录
	err = db.Delete(file).Error
	if err != nil {
		return fmt.Errorf("failed to delete file or directory: %v", err)
	}

	return nil
}

// DeleteFilesOrDirs 删除文件或目录
func DeleteFilesOrDirs(tx *gorm.DB, file_list []*foresttype.KnownowForestFile) error {
	sql := tx.Table(foresttype.TableNameKnownowForestFile)
	for _, f := range file_list {
		sql = sql.Where("parent_ids LIKE ?", fmt.Sprintf("%s%d", f.ParentIDs, f.ID)+"/%")
	}
	// 删除子文件和子目录
	err := sql.Delete(&foresttype.KnownowForestFile{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete children files or directories: %v", err)
	}
	// 删除当前文件或目录
	err = tx.Delete(file_list).Error
	if err != nil {
		return fmt.Errorf("failed to delete file or directory: %v", err)
	}

	return nil
}
