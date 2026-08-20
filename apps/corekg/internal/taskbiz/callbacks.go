package taskbiz

import (
	"context"
	"encoding/json"

	"github.com/insmtx/corekg/apps/kecore/models/coretask"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
)

// pipelineNext 任务成功后的下一阶段推进映射。
var pipelineNext = map[string]string{
	coretask.CopyTask:    coretask.PraseTask,
	coretask.Doc2PDFTask: coretask.PraseTask,
	coretask.PraseTask:   coretask.KnowledgeTask,
}

// statusFieldByType 各任务类型成功/失败时更新的文件状态字段。
var statusFieldByType = map[string]string{
	coretask.CopyTask:        "parse_status",
	coretask.Doc2PDFTask:     "parse_status",
	coretask.PraseTask:       "parse_status",
	coretask.KnowledgeTask:   "knowledge_status",
	coretask.DescriptionTask: "desc_status",
}

// RegisterForestCallbacks 注册文档摄入链路的 HTTP 回调，使 pipeline 通过
// knowledge.TaskCallBack 回报后，corekg 能推进到下一阶段并更新文件状态。
func RegisterForestCallbacks(ctx context.Context) {
	for taskType := range pipelineNext {
		task.RegisterCallBack(taskType, forestCallBack(ctx, taskType))
	}
	task.RegisterCallBack(coretask.KnowledgeTask, forestCallBack(ctx, coretask.KnowledgeTask))
	logs.InfoContextf(ctx, "[taskbiz] registered forest task callbacks")
}

func forestCallBack(rootCtx context.Context, taskType string) func(ctx context.Context, tsk *task.Task) error {
	return func(ctx context.Context, tsk *task.Task) error {
		logs.InfoContextf(ctx, "[taskbiz] callback task_id=%d type=%s status=%s", tsk.ID, tsk.TaskType, tsk.TaskStatus)

		field, _ := statusFieldByType[taskType]

		if tsk.TaskStatus == task.TaskStatusSuccess {
			if err := coretask.UpdateFileStatus(ctx, tsk.SubjectID, field, task.TaskStatusSuccess); err != nil {
				logs.ErrorContextf(ctx, "[taskbiz] update file status success fail: %v", err)
				return err
			}
			if nextType, ok := pipelineNext[tsk.TaskType]; ok {
				return createAndPushNextTask(ctx, tsk, nextType)
			}
			return checkAllTasksDone(ctx, tsk.SubjectID)
		}

		if err := coretask.UpdateFileStatus(ctx, tsk.SubjectID, field, task.TaskStatusFail); err != nil {
			logs.ErrorContextf(ctx, "[taskbiz] update file status fail: %v", err)
			return err
		}
		// 重试由 task.SaveTask 的 redo 重入队逻辑处理。
		return nil
	}
}

func createAndPushNextTask(ctx context.Context, prevTask *task.Task, nextTaskType string) error {
	var payload ragtask.TaskPayload
	if err := json.Unmarshal([]byte(prevTask.Payload), &payload); err != nil {
		return err
	}
	payload.TaskType = nextTaskType

	// prase_pdf_task 的 payload 由 buildTaskPayload 构造，未含 es_index；推进到
	// knowledge_task 时若为空则按森林补上，否则 pipeline chunker 无法写 ES。
	if nextTaskType == coretask.KnowledgeTask && payload.ESIndex == "" && payload.ForestID > 0 {
		if forestInfo, err := forest.GetForestByID(ctx, payload.ForestID); err == nil {
			payload.ESIndex = forestInfo.EsIndex()
		} else {
			logs.ErrorContextf(ctx, "[taskbiz] createNextTask fill es_index failed: %v", err)
		}
	}

	nextTask := &task.Task{
		Uin:               prevTask.Uin,
		CompanyID:         prevTask.CompanyID,
		TaskType:          nextTaskType,
		TaskStatus:        task.TaskStatusPending,
		SubjectID:         prevTask.SubjectID,
		AppGroup:          prevTask.AppGroup,
		Payload:           payload.String(),
		TaskConfigTimeout: prevTask.TaskConfigTimeout,
		TaskConfigRedo:    prevTask.TaskConfigRedo,
		Priority:          prevTask.Priority,
	}
	if err := task.CreateTask(nextTask); err != nil {
		return err
	}
	logs.InfoContextf(ctx, "[taskbiz] created next task: type=%s id=%d prev=%d", nextTaskType, nextTask.ID, prevTask.ID)
	return nil
}

func checkAllTasksDone(ctx context.Context, subjectID uint) error {
	var pending int64
	dbutil.Core().Model(&task.Task{}).
		Where("subject_id = ? AND app_group = ?", subjectID, coretask.AppGroup).
		Where("task_status NOT IN ?", []task.TaskStatus{task.TaskStatusSuccess, task.TaskStatusCancel}).
		Count(&pending)
	if pending > 0 {
		return nil
	}

	var tsk task.Task
	if err := dbutil.Core().
		Where("subject_id = ? AND app_group = ?", subjectID, coretask.AppGroup).
		First(&tsk).Error; err != nil {
		return err
	}
	var payload ragtask.TaskPayload
	if err := json.Unmarshal([]byte(tsk.Payload), &payload); err != nil {
		return err
	}
	logs.InfoContextf(ctx, "[taskbiz] all tasks done for subject %d, calling SuccessFile", subjectID)
	return coretask.SuccessFile(ctx, &payload)
}
