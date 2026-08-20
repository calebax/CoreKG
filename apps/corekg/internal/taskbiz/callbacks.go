package taskbiz

import (
	"context"
	"encoding/json"

	"github.com/insmtx/corekg/apps/kecore/models/coretask"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
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
	// DescriptionTask 是并行叶子任务：成功回调更新 desc_status，并触发文件是否全部完成的汇聚检查。
	task.RegisterCallBack(coretask.DescriptionTask, forestCallBack(ctx, coretask.DescriptionTask))
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

	// prase 任务的 FileURL 指向原始预览文件；推进到 knowledge_task 时必须改为
	// 解析产物目录（content.md），否则 chunker 会把原始二进制当正文切块入库，
	// 导致 ES 的 description 是一堆乱码/二进制（见「总结文档」检索到二进制乱码的问题）。
	if nextTaskType == coretask.KnowledgeTask && payload.FileID > 0 {
		if fileInfo, err := forest.GetForestFileByID(payload.FileID); err == nil {
			payload.FileURL = fs.Forest.GetPublicURL(fs.FileContentPath(fileInfo.GetAlgoFilePath(), payload.FileID), false)
			payload.UploadPath = fs.FileKnowledgeDirPath(fileInfo.GetAlgoFilePath(), payload.FileID)
			logs.InfoContextf(ctx, "[taskbiz] knowledge_task redirect file_url to parsed content: file_id=%d url=%s", payload.FileID, payload.FileURL)
		} else {
			logs.ErrorContextf(ctx, "[taskbiz] createNextTask resolve parsed content path failed: %v", err)
		}
	}

	// doc2pdf 成功推进到 prase_pdf_task 时，上游 doc2pdf 的 payload.FileURL 指向源文件、
	// UploadPath 指向预览 PDF 路径，不能直接复用；应像 GenerateFileTask 那样改指向 doc2pdf
	// 生成的预览 PDF 与算法产物目录，否则伪 .docx/.doc(OLE2) 源文件会被直接喂给 MinerU 而 400。
	if nextTaskType == coretask.PraseTask && payload.FileID > 0 {
		if fileInfo, err := forest.GetForestFileByID(payload.FileID); err == nil {
			if previewPath, pErr := fileInfo.GetForestPriviewFilePath(); pErr == nil && previewPath != nil {
				payload.FileURL = fs.Forest.GetPublicURL(*previewPath, false)
				payload.UploadPath = fs.FileFileAlgoPath(fileInfo.GetAlgoFilePath(), payload.FileID)
				logs.InfoContextf(ctx, "[taskbiz] prase_task redirect file_url to preview file: file_id=%d url=%s", payload.FileID, payload.FileURL)
			} else if pErr != nil {
				logs.WarnContextf(ctx, "[taskbiz] prase task resolve preview file path failed, keep inherited payload: %v", pErr)
			}
		} else {
			logs.ErrorContextf(ctx, "[taskbiz] createNextTask resolve forest file failed: %v", err)
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

	// 解析完成后，除了建库分块（KnowledgeTask），还要并行派发 DescriptionTask，
	// 用于生成 file_description 类文档，支撑知识库层面的“总结/摘要”检索。
	if nextTaskType == coretask.KnowledgeTask {
		if err := createAndPushDescriptionTask(ctx, prevTask, payload); err != nil {
			logs.ErrorContextf(ctx, "[taskbiz] createAndPushDescriptionTask failed: %v", err)
			return err
		}
	}
	return nil
}

// createAndPushDescriptionTask 创建并派发文件描述生成任务（ke.description_task）。
// 描述结果会写入 ES 的 file_description 文档，供 knowledge_summary 检索使用。
func createAndPushDescriptionTask(ctx context.Context, prevTask *task.Task, payload ragtask.TaskPayload) error {
	if payload.ForestID == 0 || payload.FileID == 0 || payload.FileURL == "" {
		logs.WarnContextf(ctx, "[taskbiz] desc task skipped: missing forest/file/file_url")
		return nil
	}
	descPayload := payload
	descPayload.TaskType = coretask.DescriptionTask
	desc := &task.Task{
		Uin:               prevTask.Uin,
		CompanyID:         prevTask.CompanyID,
		TaskType:          coretask.DescriptionTask,
		TaskStatus:        task.TaskStatusPending,
		SubjectID:         prevTask.SubjectID,
		AppGroup:          prevTask.AppGroup,
		Payload:           descPayload.String(),
		TaskConfigTimeout: prevTask.TaskConfigTimeout,
		TaskConfigRedo:    prevTask.TaskConfigRedo,
		Priority:          prevTask.Priority,
	}
	if err := task.CreateTask(desc); err != nil {
		return err
	}
	logs.InfoContextf(ctx, "[taskbiz] created desc task: id=%d prev=%d", desc.ID, prevTask.ID)
	return task.PushTaskQueue(ctx, coretask.DescriptionTask)
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
