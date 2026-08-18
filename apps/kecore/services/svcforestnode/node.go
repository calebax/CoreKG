package svcforestnode

import (
	"context"
	"errors"
	"fmt"

	"github.com/insmtx/corekg/apps/kecore/models/coretask"
	"github.com/insmtx/corekg/apps/kecore/models/decoupler"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/apps/kesearch/models/chunk"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

var (
	ErrGetForestFailed       = errors.New("get forest failed")
	ErrNoPermission          = errors.New("no permission")
	ErrGetParentNodeFailed   = errors.New("get parent node failed")
	ErrCheckFileExistsFailed = errors.New("check file exists failed")
	ErrCreateDirFailed       = errors.New("create dir failed")
	ErrGetFileOrDirFailed    = errors.New("get file or dir failed")
	ErrUnknownFileList       = errors.New("unknown file list")
	ErrTaskRunning           = errors.New("task running")
	ErrFileStatusCheckFailed = errors.New("file status check failed")
	ErrDeleteFileOrDirFailed = errors.New("delete file or dir failed")
	ErrGetSourceFileFailed   = errors.New("get source file failed")
	ErrCheckNewNameFailed    = errors.New("check new name failed")
	ErrNewNameExists         = errors.New("new name exists")
	ErrUpdateFileNameFailed  = errors.New("update file name failed")
)

type CreateDirRequest struct {
	Uin       uint
	CompanyID uint
	ForestID  uint
	ParentID  uint
	Name      string
}

type DeletePathRequest struct {
	Uin     uint
	FileIDs []uint
}

type RenamePathRequest struct {
	Uin     uint
	FileID  uint
	NewName string
}

func CreateDir(ctx context.Context, req *CreateDirRequest) (uint, error) {
	var (
		parent *foresttype.KnownowForestFile
		err    error
	)

	forestInfo, err := forest.GetForestByID(ctx, req.ForestID)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrGetForestFailed, err)
	}
	if !perm.HasManageAct(ctx, req.Uin, forestInfo.ID, foresttype.ResourceTypeForest) {
		return 0, ErrNoPermission
	}

	if req.ParentID != 0 {
		parent, err = forest.GetForestFileByID(req.ParentID)
		if err != nil {
			return 0, fmt.Errorf("%w: %v", ErrGetParentNodeFailed, err)
		}
	}

	name := req.Name
	isExist, err := forest.IsExistForestFile(req.ForestID, req.ParentID, name)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrCheckFileExistsFailed, err)
	}
	if isExist {
		logs.DebugContextf(ctx, "dir already exists, create same-name dir with suffix")
		name = decoupler.TruncateName(name)
	}

	file := &foresttype.KnownowForestFile{
		CompanyID: req.CompanyID,
		Uin:       req.Uin,
		ForestID:  req.ForestID,
		IsDir:     1,
		Name:      name,
		ParentID:  req.ParentID,
	}
	if req.ParentID == 0 {
		file.Depth = 1
	} else {
		file.ParentIDs = fmt.Sprintf("%s%d/", parent.ParentIDs, parent.ID)
		file.Depth = parent.Depth + 1
	}

	if err = forest.CreateDir(ctx, file); err != nil {
		return 0, fmt.Errorf("%w: %v", ErrCreateDirFailed, err)
	}
	return file.ID, nil
}

func DeletePath(ctx context.Context, req *DeletePathRequest) error {
	fileList, err := forest.GetDirsFiles(ctx, req.FileIDs)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrGetFileOrDirFailed, err)
	}
	if len(fileList) == 0 {
		return ErrUnknownFileList
	}

	forestInfo, err := forest.GetForestByID(ctx, fileList[0].ForestID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrGetForestFailed, err)
	}
	if !perm.HasManageAct(ctx, req.Uin, forestInfo.ID, foresttype.ResourceTypeForest) {
		return ErrNoPermission
	}

	var fileIDs []uint
	for _, file := range fileList {
		if file.IsDir == 1 {
			continue
		}
		fileIDs = append(fileIDs, file.ID)
	}
	if err := forest.DeleteFilesStatusCheck(ctx, fileIDs); err != nil {
		if errors.Is(err, forest.ErrHasRunningTask) {
			return ErrTaskRunning
		}
		return fmt.Errorf("%w: %v", ErrFileStatusCheckFailed, err)
	}

	if err = dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		graphInfo, err := graph.GetForestGraph(ctx, fileList[0].ForestID)
		if err == nil {
			for _, file := range fileList {
				if err = graph.DeleteFileGraph(ctx, tx, graphInfo, file); err != nil {
					return err
				}
			}
		}
		if len(fileList) > 0 {
			if err = decoupler.DeleteFilesOrDirs(tx, fileList); err != nil {
				return err
			}
		}
		if len(fileIDs) == 0 {
			return nil
		}
		if err = coretask.DeleteTasksByFileIDs(ctx, fileIDs); err != nil {
			return err
		}
		if err = essearch.DeleteFileReferences(ctx, forestInfo.EsIndex(), fileIDs); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("%w: %v", ErrDeleteFileOrDirFailed, err)
	}
	return nil
}

func RenamePath(ctx context.Context, req *RenamePathRequest) error {
	file, err := forest.GetForestFileByID(req.FileID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrGetSourceFileFailed, err)
	}

	forestInfo, err := forest.GetForestByID(ctx, file.ForestID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrGetForestFailed, err)
	}
	if !perm.HasManageAct(ctx, req.Uin, forestInfo.ID, foresttype.ResourceTypeForest) {
		return ErrNoPermission
	}
	if file.Name == req.NewName {
		return nil
	}

	isExist, err := forest.IsExistForestFile(file.ForestID, file.ParentID, req.NewName)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCheckNewNameFailed, err)
	}
	if isExist {
		return ErrNewNameExists
	}

	if err = dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		file.Name = req.NewName
		if err := tx.Save(file).Error; err != nil {
			return err
		}
		return chunk.UpdateChunkFileName(ctx, file.ID, file.Name)
	}); err != nil {
		return fmt.Errorf("%w: %v", ErrUpdateFileNameFailed, err)
	}
	return nil
}
