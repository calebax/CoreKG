package webctl

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/keapp/models/web"
	"github.com/insmtx/corekg/apps/keapp/services/svcweb"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
)

func TriggerCrawl(ctx *gin.Context, req *TriggerCrawlRequest, resp *TriggerCrawlResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	task := &web.KeCrawlTask{
		AppID:    req.Request.AppID,
		TaskType: req.Request.TaskType,
		CreatedBy: runtime.Uin(ctx),
	}
	taskID, err := svcweb.TriggerCrawl(ctx, task)
	if err != nil {
		if errors.Is(err, svcweb.ErrTriggerCrawlFailed) {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "keapp_trigger_crawl_failed"
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_trigger_crawl_failed"
		return
	}
	resp.Response.TaskID = taskID
}

func GetCrawlTask(ctx *gin.Context, req *GetCrawlTaskRequest, resp *GetCrawlTaskResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	task, err := svcweb.GetCrawlTask(ctx, req.Request.TaskID)
	if err != nil {
		if errors.Is(err, svcweb.ErrCrawlTaskNotFound) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "keapp_task_not_found"
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_get_task_failed"
		return
	}
	resp.Response.Task = task
}

func ListCrawlTasks(ctx *gin.Context, req *ListCrawlTasksRequest, resp *ListCrawlTasksResponse) {
	items, err := svcweb.ListCrawlTasks(ctx, req.Request.AppID, req.Request.Limit, req.Request.Offset)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_list_tasks_failed"
		return
	}
	resp.Response.Items = items
}

func CancelCrawlTask(ctx *gin.Context, req *CancelCrawlTaskRequest, resp *CancelCrawlTaskResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	if err := svcweb.CancelCrawlTask(ctx, req.Request.TaskID); err != nil {
		if errors.Is(err, svcweb.ErrCrawlTaskNotFound) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "keapp_task_not_found"
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_cancel_task_failed"
		return
	}
}
