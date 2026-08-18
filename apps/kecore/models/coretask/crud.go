package coretask

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// SuccessFile 成功回调
func SuccessFile(ctx context.Context, payload *ragtask.TaskPayload) error {
	if payload.FileID == 0 || payload.ForestID == 0 {
		return fmt.Errorf("subject_id or forest_id is 0")
	}
	err := dbutil.Knownow().Model(&foresttype.KnownowForestFile{}).
		Where("id = ?", payload.FileID).
		Updates(map[string]interface{}{
			"parse_status":    foresttype.TaskStatusSuccess,
			"desc_status":     foresttype.TaskStatusSuccess,
			"analysis_status": foresttype.TaskStatusSuccess,
			"mindmap_status":  foresttype.TaskStatusSuccess,
			// "graph_status":     foresttype.TaskStatusSuccess,
			"knowledge_status": foresttype.TaskStatusSuccess,
		}).Error
	if err != nil {
		logs.ErrorContextf(ctx, "update file status error: %v", err)
		return err
	}
	err = dbutil.Knownow().Model(&foresttype.KnownowForest{}).
		Where("id = ?", payload.ForestID).
		Updates(map[string]interface{}{
			"knowledge_status": foresttype.TaskStatusSuccess,
		}).Error
	if err != nil {
		logs.ErrorContextf(ctx, "update KnownowForest status error: %v", err)
		return err
	}
	return nil
}

func UpdateFileStatus(ctx context.Context, fileID uint, filed string, status task.TaskStatus) error {
	vm := map[string]interface{}{
		filed: status,
	}

	err := dbutil.Knownow().WithContext(ctx).Model(&foresttype.KnownowForestFile{}).
		Where("id = ?", fileID).
		Updates(vm).Error
	if err != nil {
		logs.ErrorContextf(ctx, "update file status error: %v", err)
		return err
	}
	return nil
}
func DeleteTasksByForestID(ctx context.Context, forestID uint) error {
	var fileIDs []uint
	if err := dbutil.Knownow().
		WithContext(ctx).
		Table(foresttype.TableNameKnownowForestFile).
		Where("forest_id = ?", forestID).
		Pluck("id", &fileIDs).
		Error; err != nil {
		return err
	}

	if len(fileIDs) == 0 {
		return nil
	}

	if err := dbutil.Core().Table("core_task ct").
		Where("subject_id in ?", fileIDs).
		Delete(&task.Task{}).
		Error; err != nil {
		return err
	}

	return nil
}

func DeleteTasksByFileIDs(ctx context.Context, fileIDs []uint) error {
	if err := dbutil.Core().WithContext(ctx).Table("core_task ct").
		Where("ct.deleted_at IS NULL").
		Where("app_group = ?", AppGroup).
		Where("ct.subject_id in (?)", fileIDs).
		Delete(&task.Task{}).
		Error; err != nil {
		return err
	}
	return nil
}

// DeleteChunkTask 删除对应文件的chunk任务
func DeleteChunkTask(ctx context.Context, tx *gorm.DB, fileID uint) error {
	err := tx.WithContext(ctx).Table(task.TableNameCoreTask).
		WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("subject_id = ?", fileID).
		Where("app_group = ?", AppGroup).
		Where("task_type in ?", []string{KnowledgeTask, GraphTask}).
		Delete(&task.Task{}).Error
	if err != nil {
		logs.ErrorContextf(ctx, "delete chunk task error: %v", err)
		return err
	}
	return nil
}

// FindPendingTaskBySubjectID 根据subjectID获取任务
func FindPendingTaskBySubjectID(ctx context.Context, subjectID uint, appGroup string) ([]*task.Task, error) {
	var tasks []*task.Task
	err := dbutil.Core().
		WithContext(ctx).
		Where("subject_id = ?", subjectID).
		Where("app_group = ?", appGroup).
		Where("task_status NOT IN (?)", []task.TaskStatus{task.TaskStatusCancel, task.TaskStatusSuccess}).
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// FindCanDoTaskBySubjectID 获取当前subjectID可执行的任务
func FindCanDoTaskBySubjectID(ctx context.Context, subjectID uint, appGroup string) ([]*task.Task, error) {
	var tasks []*task.Task
	err := dbutil.Core().
		WithContext(ctx).
		Where("subject_id = ?", subjectID).
		Where("app_group = ?", appGroup).
		Where("task_status IN (?)", []task.TaskStatus{task.TaskStatusPending, task.TaskStatusFail, task.TaskStatusRunning}).
		Where("redo <= task_config_redo").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
