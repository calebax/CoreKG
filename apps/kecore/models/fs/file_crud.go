package fs

import (
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/storage"
)

// CreateFileInfo 创建文件信息
func CreateFileInfo(fi *storage.FileInfo) error {
	err := dbutil.Core().Create(fi).Error
	return err
}
