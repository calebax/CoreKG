package svcweb

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/nats-io/nats.go"

	"github.com/insmtx/corekg/apps/keapp/models/web"
)

var (
	ErrCrawlTaskNotFound   = errors.New("crawl task not found")
	ErrTriggerCrawlFailed  = errors.New("trigger crawl failed")
	ErrCancelCrawlFailed   = errors.New("cancel crawl failed")
	ErrNATSNotConnected    = errors.New("nats connection not set")
)

var natsConn *nats.Conn

const natsSubjectCrawlTrigger = "keapp.crawl.trigger"

type CrawlTriggerMsg struct {
	TaskID uint `json:"task_id"`
}

func SetNATSConn(nc *nats.Conn) {
	natsConn = nc
}

func TriggerCrawl(ctx context.Context, task *web.KeCrawlTask) (uint, error) {
	dao := web.NewCrawlTaskDao()
	task.Status = web.CrawlTaskPending
	if err := dao.Insert(ctx, task); err != nil {
		return 0, ErrTriggerCrawlFailed
	}

	if natsConn == nil {
		return task.ID, ErrNATSNotConnected
	}

	msg := CrawlTriggerMsg{TaskID: task.ID}
	data, err := json.Marshal(msg)
	if err != nil {
		return task.ID, ErrTriggerCrawlFailed
	}
	if err := natsConn.Publish(natsSubjectCrawlTrigger, data); err != nil {
		return task.ID, ErrTriggerCrawlFailed
	}
	return task.ID, nil
}

func GetCrawlTask(ctx context.Context, id uint) (*web.KeCrawlTask, error) {
	dao := web.NewCrawlTaskDao()
	entity, err := dao.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCrawlTaskNotFound
	}
	if entity == nil {
		return nil, ErrCrawlTaskNotFound
	}
	return entity, nil
}

func ListCrawlTasks(ctx context.Context, appID uint, limit, offset int) ([]*web.KeCrawlTask, error) {
	dao := web.NewCrawlTaskDao()
	return dao.ListByAppID(ctx, appID, limit, offset)
}

func CancelCrawlTask(ctx context.Context, id uint) error {
	dao := web.NewCrawlTaskDao()
	entity, err := dao.GetByID(ctx, id)
	if err != nil || entity == nil {
		return ErrCrawlTaskNotFound
	}
	if err := dao.CancelTask(ctx, id); err != nil {
		return ErrCancelCrawlFailed
	}
	return nil
}

func RecrawlResource(ctx context.Context, resourceID uint, createdBy uint) (uint, error) {
	resDao := web.NewWebResourceDao()
	resource, err := resDao.GetByID(ctx, resourceID)
	if err != nil || resource == nil {
		return 0, ErrResourceNotFound
	}

	task := &web.KeCrawlTask{
		AppID:      resource.AppID,
		ResourceID: &resourceID,
		SourceURL:  resource.SourceURL,
		TaskType:   web.CrawlTaskSingle,
		CreatedBy:  createdBy,
	}
	return TriggerCrawl(ctx, task)
}
