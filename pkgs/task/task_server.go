package task

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// GetPendingTask 获取一个待执行任务（HTTP worker 轮询）。
//
// 说明：pipeline / clients.task_worker 等 HTTP worker 通过此端点拉取任务，任务来源为
// MySQL core_task 表中当前 pending/fail 的可执行任务。任务创建时同样会发布一条 NATS
// dispatch 消息（见 PushTaskQueue），二者以数据库行级锁（GetOnePendingTask）保证不会
// 被重复消费：HTTP worker 直接以该 DB 行作为队列消费。
func GetPendingTask(ctx *gin.Context, req *GetPendingTestRequest, resp *GetPendingTestResponse) {
	if req.Validity(resp); resp.Code != 0 {
		return
	}

	tsk, err := GetOnePendingTask(req.Request.TaskType, req.Request.WorkerID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetPendingTask GetOnePendingTask task_type: %v, worker_id: %v, error: %v", req.Request.TaskType, req.Request.WorkerID, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "task_query_task_failed" // 查询任务失败
		return
	}
	if tsk == nil {
		logs.InfoContextf(ctx, "GetPendingTask no task, task_type: %v, worker_id: %v", req.Request.TaskType, req.Request.WorkerID)
		resp.Code = errcode.ErrCode_NotFound
		resp.Message = "task_no_task" // 暂无任务
		return
	}

	resp.Response.TaskID = tsk.ID
	resp.Response.Payload = tsk.Payload
}

// TaskCallBack 任务回调（HTTP worker 回报执行结果）。
func TaskCallBack(ctx *gin.Context, req *TaskCallBackRequest, resp *TaskCallBackResponse) {
	if req.Validity(resp); resp.Code != 0 {
		return
	}

	logs.InfoContextf(ctx, "task callback task_id: %v, status: %v", req.Request.TaskID, req.Request.Status)
	tsk, err := GetTaskByIDAndWorkerID(req.Request.TaskID, req.Request.WorkerID)
	if err != nil {
		logs.ErrorContextf(ctx, "TaskCallBack GetTaskByIDAndWorkerID task_id: %v, worker_id: %v, error: %v", req.Request.TaskID, req.Request.WorkerID, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "task_get_task_failed_or_timeout" // 获取任务失败，或任务以超时
		return
	}

	tsk.TaskStatus = req.Request.Status
	if req.Request.Status == TaskStatusFail {
		tsk.ErrMsg = req.Request.ErrorMessage
	} else {
		tsk.Result = req.Request.Result
	}
	now := time.Now()
	tsk.EndAt = &now

	// 调用业务回调（如更新文件状态 / 推进下一阶段）。未注册则视为仅落库。
	if tc, err := GetCallBack(tsk.TaskType); err == nil && tc.CallBack != nil {
		if cbErr := tc.CallBack(ctx, tsk); cbErr != nil {
			logs.ErrorContextf(ctx, "TaskCallBack callback task_id: %v, task_type: %v, error: %v", req.Request.TaskID, tsk.TaskType, cbErr)
			tsk.TaskStatus = TaskStatusFail
			tsk.ErrMsg = cbErr.Error()
		}
	}

	err = SaveTask(tsk)
	if err != nil {
		logs.ErrorContextf(ctx, "TaskCallBack SaveTask task_id: %v, worker_id: %v, error: %v", req.Request.TaskID, req.Request.WorkerID, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "task_save_task_failed" // 保存任务失败
		return
	}
}
