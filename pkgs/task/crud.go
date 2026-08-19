package task

import (
	"context"
	"errors"
	"log"
	"os"
	"sort"
	"time"

	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

var customLogger = logger.New(
	log.New(os.Stdout, "\r\n", log.LstdFlags),
	logger.Config{
		LogLevel: logger.Silent,
	},
)

func GetOnePendingTask(task_type, worker_id string) (*Task, error) {
	var (
		tsk Task
		ctx = context.TODO()
	)
	db := dbutil.Core().Session(&gorm.Session{Logger: customLogger})
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	err := tx.
		WithContext(ctx).
		Where("task_type = ?", task_type).
		Where("task_status IN (?)", []TaskStatus{TaskStatusPending, TaskStatusFail}).
		Where("redo <= task_config_redo").
		Order("priority DESC, updated_at ASC").
		Clauses(clause.Locking{Strength: "UPDATE", Options: clause.LockingOptionsSkipLocked}).
		First(&tsk).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return nil, nil
		}
		logs.ErrorContextf(ctx, "Failed to find pending task: %v", err)
		tx.Rollback()
		return nil, err
	}
	now := time.Now()
	tsk.TaskStatus = TaskStatusRunning
	tsk.StartAt = &now
	tsk.WorkerID = worker_id
	err = tx.Save(&tsk).Error
	if err != nil {
		logs.ErrorContextf(ctx, "Failed to update task status to running: %v", err)
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return &tsk, nil
}

// PeekOnePendingTask 读取一个可执行任务但不改变其状态。
// 用于任务分发（NATS dispatch）取数，避免把任务预置为 running 而挡住 HTTP worker 消费。
func PeekOnePendingTask(task_type string) (*Task, error) {
	var tsk Task
	err := dbutil.Core().
		Where("task_type = ?", task_type).
		Where("task_status IN (?)", []TaskStatus{TaskStatusPending, TaskStatusFail}).
		Where("redo <= task_config_redo").
		Order("priority DESC, updated_at ASC").
		First(&tsk).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tsk, nil
}

func GetTaskByID(id uint) (*Task, error) {
	var tsk *Task
	err := dbutil.Core().Where("id = ?", id).First(&tsk).Error
	if err != nil {
		return nil, err
	}
	return tsk, nil
}

func GetTaskByIDAndWorkerID(id uint, worker_id string) (*Task, error) {
	var tsk *Task
	err := dbutil.Core().Where("id = ?", id).
		Where("worker_id = ?", worker_id).
		First(&tsk).Error
	if err != nil {
		return nil, err
	}
	return tsk, nil
}

func SaveTask(tsk *Task) error {
	if tsk.StartAt != nil && tsk.EndAt != nil {
		if tsk.StartAt.Before(*tsk.EndAt) {
			tsk.Cost = int64(tsk.EndAt.Sub(*tsk.StartAt).Seconds())
		}
	}
	if tsk.TaskStatus == TaskStatusFail {
		tsk.Redo++
	}
	err := dbutil.Core().Save(tsk).Error
	if err != nil {
		return err
	}
	// 失败时若未超过最大重试次数，重新入队交由 worker 重试。
	if tsk.TaskStatus == TaskStatusFail && tsk.Redo <= tsk.TaskConfigRedo {
		err = PushTaskQueue(context.Background(), tsk.TaskType)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetNextStepTask 获取同一 subject 下第一个未全部完成 step 中可执行的任务。
// 用于 HTTP 回调成功后将流程推进到下一阶段。
func GetNextStepTask(tsk *Task) ([]*Task, error) {
	var allTasks []*Task
	err := dbutil.Core().Where("subject_id = ? AND app_group = ?", tsk.SubjectID, tsk.AppGroup).
		Order("step ASC").
		Find(&allTasks).Error
	if err != nil {
		return nil, err
	}

	stepTaskMap := make(map[int][]*Task)
	stepSet := map[int]struct{}{}
	for _, task := range allTasks {
		stepTaskMap[task.Step] = append(stepTaskMap[task.Step], task)
		stepSet[task.Step] = struct{}{}
	}

	var steps []int
	for step := range stepSet {
		steps = append(steps, step)
	}
	sort.Ints(steps)

	for _, step := range steps {
		tasks := stepTaskMap[step]
		allCompleted := true
		for _, t := range tasks {
			if t.TaskStatus != TaskStatusSuccess {
				allCompleted = false
				break
			}
		}
		if !allCompleted {
			var result []*Task
			for _, t := range tasks {
				if t.TaskStatus == TaskStatusPending || t.TaskStatus == TaskStatusFail || t.TaskStatus == TaskStatusRunning {
					result = append(result, t)
				}
			}
			return result, nil
		}
	}
	return nil, nil
}

func CreateTask(tsk *Task) error {
	if tsk.AppGroup == "" {
		return errors.New("app_group cannot be empty")
	}
	if tsk.SubjectID == 0 {
		return errors.New("subject_id cannot be zero")
	}
	if tsk.TaskType == "" {
		return errors.New("task_type cannot be empty")
	}
	if tsk.Payload == "" {
		return errors.New("payload cannot be empty")
	}
	if tsk.TaskConfigTimeout <= 0 {
		return errors.New("task_config_timeout must be greater than zero")
	}
	if tsk.TaskConfigRedo < 0 {
		return errors.New("task_config_redo cannot be negative")
	}
	err := dbutil.Core().Create(tsk).Error
	if err != nil {
		return err
	}
	return PushTaskQueue(context.Background(), tsk.TaskType)
}
