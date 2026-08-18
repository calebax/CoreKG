package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/insmtx/corekg/apps/kecore/models/coretask"
	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/nats-io/nats.go"
	"github.com/ygpkg/yg-go/logs"
)

var pipelineNext = map[string]string{
	coretask.CopyTask:    coretask.PraseTask,
	coretask.Doc2PDFTask: coretask.PraseTask,
	coretask.PraseTask:   coretask.KnowledgeTask,
}

func StartForestResultConsumers(ctx context.Context, nc *nats.Conn) error {
	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("create jetstream context: %w", err)
	}

	if err := task.EnsureResultStream(js); err != nil {
		return fmt.Errorf("ensure result stream: %w", err)
	}

	subscriptions := []struct {
		subject     string
		durable     string
		handler     nats.MsgHandler
		description string
	}{
		{
			subject:     task.ResultCopy,
			durable:     "forest-copy-result",
			handler:     makeForestResultHandler(ctx, "parse_status", true),
			description: "copy/doc2pdf 结果 → 更新 parse_status + 推进下一步",
		},
		{
			subject:     task.ResultPDFExtract,
			durable:     "forest-pdf-extract-result",
			handler:     makeForestResultHandler(ctx, "parse_status", true),
			description: "pdf_extract 结果 → 更新 parse_status + 推进下一步",
		},
		{
			subject:     task.ResultSplitChunk,
			durable:     "forest-knowledge-result",
			handler:     makeForestResultHandler(ctx, "knowledge_status", false),
			description: "split_text_chunk 结果 → 更新 knowledge_status + 汇聚检查",
		},
		{
			subject:     task.ResultDesc,
			durable:     "forest-desc-result",
			handler:     makeForestResultHandler(ctx, "desc_status", false),
			description: "desc 结果 → 更新 desc_status + 汇聚检查",
		},
	}

	for _, sub := range subscriptions {
		if _, err := js.Subscribe(sub.subject, sub.handler, nats.Durable(sub.durable)); err != nil {
			return fmt.Errorf("subscribe %s (%s): %w", sub.subject, sub.description, err)
		}
		logs.InfoContextf(ctx, "[result_consumer] subscribed: %s — %s", sub.subject, sub.description)
	}

	return nil
}

func StartGraphResultConsumers(ctx context.Context, nc *nats.Conn) error {
	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("create jetstream context: %w", err)
	}

	handler := func(msg *nats.Msg) {
		ctx := context.Background()
		var rm task.ResultMessage
		if err := json.Unmarshal(msg.Data, &rm); err != nil {
			logs.ErrorContextf(ctx, "[graph_result] unmarshal: %v", err)
			msg.Ack()
			return
		}

		tsk, err := task.GetTaskByID(rm.TaskID)
		if err != nil {
			logs.ErrorContextf(ctx, "[graph_result] get task %d: %v", rm.TaskID, err)
			msg.Nak()
			return
		}

		tsk.TaskStatus = rm.Status
		tsk.Result = rm.Result
		if rm.ErrorMessage != "" {
			tsk.ErrMsg = rm.ErrorMessage
		}

		if cbErr := coretask.GraphTaskCallBack(ctx, tsk); cbErr != nil {
			logs.ErrorContextf(ctx, "[graph_result] callback task %d: %v", rm.TaskID, cbErr)
		}

		if saveErr := task.SaveTask(tsk); saveErr != nil {
			logs.ErrorContextf(ctx, "[graph_result] save task %d: %v", rm.TaskID, saveErr)
			msg.Nak()
			return
		}

		if rm.Status == task.TaskStatusFail && tsk.Redo <= tsk.TaskConfigRedo {
			task.PushTaskQueue(ctx, tsk.TaskType)
		}

		msg.Ack()
	}

	if _, err := js.Subscribe(task.ResultAnalysis, handler, nats.Durable("graph-analysis-result")); err != nil {
		return fmt.Errorf("subscribe graph result: %w", err)
	}
	logs.InfoContextf(ctx, "[result_consumer] subscribed: %s — 图谱任务回调", task.ResultAnalysis)

	return nil
}

func makeForestResultHandler(rootCtx context.Context, statusField string, hasPipelineNext bool) nats.MsgHandler {
	return func(msg *nats.Msg) {
		var rm task.ResultMessage
		if err := json.Unmarshal(msg.Data, &rm); err != nil {
			logs.ErrorContextf(rootCtx, "[forest_result] unmarshal: %v", err)
			msg.Ack()
			return
		}

		ctx := context.Background()
		logs.InfoContextf(ctx, "[forest_result] task_id=%d type=%s status=%s field=%s", rm.TaskID, rm.TaskType, rm.Status, statusField)

		tsk, err := task.GetTaskByID(rm.TaskID)
		if err != nil {
			logs.ErrorContextf(ctx, "[forest_result] get task %d: %v", rm.TaskID, err)
			msg.Nak()
			return
		}

		tsk.TaskStatus = rm.Status
		tsk.Result = rm.Result
		if rm.ErrorMessage != "" {
			tsk.ErrMsg = rm.ErrorMessage
		}

		if saveErr := task.SaveTask(tsk); saveErr != nil {
			logs.ErrorContextf(ctx, "[forest_result] save task %d: %v", rm.TaskID, saveErr)
			msg.Nak()
			return
		}

		if rm.Status == task.TaskStatusSuccess {
			coretask.UpdateFileStatus(ctx, tsk.SubjectID, statusField, task.TaskStatusSuccess)

			if hasPipelineNext {
				if nextType, ok := pipelineNext[rm.TaskType]; ok {
					if createErr := createAndPushNextTask(ctx, tsk, nextType); createErr != nil {
						logs.ErrorContextf(ctx, "[forest_result] create next task %s for task %d: %v", nextType, rm.TaskID, createErr)
					}
				}
			} else {
				if checkErr := checkAllTasksDone(ctx, tsk.SubjectID); checkErr != nil {
					logs.ErrorContextf(ctx, "[forest_result] check all tasks done for subject %d: %v", tsk.SubjectID, checkErr)
				}
			}
		} else if rm.Status == task.TaskStatusFail {
			coretask.UpdateFileStatus(ctx, tsk.SubjectID, statusField, task.TaskStatusFail)
			if tsk.Redo <= tsk.TaskConfigRedo {
				if pushErr := task.PushTaskQueue(ctx, tsk.TaskType); pushErr != nil {
					logs.ErrorContextf(ctx, "[forest_result] push retry task %d: %v", rm.TaskID, pushErr)
				}
			}
		}

		msg.Ack()
	}
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
		return fmt.Errorf("get any task for subject %d: %w", subjectID, err)
	}

	var payload ragtask.TaskPayload
	if err := json.Unmarshal([]byte(tsk.Payload), &payload); err != nil {
		return fmt.Errorf("unmarshal payload for subject %d: %w", subjectID, err)
	}

	logs.InfoContextf(ctx, "[forest_result] all tasks done for subject %d, calling SuccessFile", subjectID)
	return coretask.SuccessFile(ctx, &payload)
}

func createAndPushNextTask(ctx context.Context, prevTask *task.Task, nextTaskType string) error {
	var payload ragtask.TaskPayload
	if err := json.Unmarshal([]byte(prevTask.Payload), &payload); err != nil {
		return fmt.Errorf("unmarshal prev payload: %w", err)
	}

	payload.CommonPayload.TaskType = nextTaskType
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
		return fmt.Errorf("create task %s: %w", nextTaskType, err)
	}

	logs.InfoContextf(ctx, "[forest_result] created next task: type=%s id=%d prev_task=%d", nextTaskType, nextTask.ID, prevTask.ID)
	return nil
}
