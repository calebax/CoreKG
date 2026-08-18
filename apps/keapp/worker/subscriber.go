package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/insmtx/corekg/apps/keapp/models/web"
	"github.com/insmtx/corekg/apps/keapp/services/svcweb"
	"github.com/ygpkg/yg-go/logs"
)

func Start(ctx context.Context, nc *nats.Conn, cfg Config) error {
	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("create JetStream context: %w", err)
	}

	if err := ensureCrawlStream(js, cfg); err != nil {
		return fmt.Errorf("ensure stream: %w", err)
	}

	if err := recoverPendingTasks(ctx, nc, cfg); err != nil {
		logs.ErrorContextf(ctx, "recover pending tasks error: %v", err)
	}

	sub, err := js.PullSubscribe(cfg.Subject, cfg.ConsumerName,
		nats.Durable(cfg.ConsumerName),
		nats.AckWait(cfg.AckWait),
		nats.MaxDeliver(cfg.MaxDeliver),
	)
	if err != nil {
		return fmt.Errorf("create pull subscribe: %w", err)
	}

	ruleDao := web.NewCrawlRuleDao()

	var wg sync.WaitGroup
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					logs.InfoContextf(ctx, "crawl worker %d stopped", workerID)
					return
				default:
				}

				msgs, err := sub.Fetch(1, nats.MaxWait(1*time.Second))
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					if err == nats.ErrTimeout {
						continue
					}
					logs.ErrorContextf(ctx, "crawl worker %d fetch error: %v", workerID, err)
					continue
				}

				if len(msgs) == 0 {
					continue
				}

				msg := msgs[0]
				if processErr := processMessage(ctx, msg, cfg, ruleDao); processErr != nil {
					logs.ErrorContextf(ctx, "crawl worker %d process error: %v", workerID, processErr)
					if nakErr := msg.Nak(); nakErr != nil {
						logs.ErrorContextf(ctx, "crawl worker %d nak error: %v", workerID, nakErr)
					}
					continue
				}

				if ackErr := msg.Ack(); ackErr != nil {
					logs.ErrorContextf(ctx, "crawl worker %d ack error: %v", workerID, ackErr)
				}
			}
		}(i)
	}

	go func() {
		<-ctx.Done()
		if unsubErr := sub.Unsubscribe(); unsubErr != nil {
			logs.ErrorContextf(ctx, "unsubscribe error: %v", unsubErr)
		}
		wg.Wait()
		logs.InfoContextf(ctx, "all crawl workers stopped")
	}()

	logs.InfoContextf(ctx, "crawl worker started with %d concurrency on subject %s", cfg.Concurrency, cfg.Subject)
	return nil
}

func processMessage(ctx context.Context, msg *nats.Msg, cfg Config, ruleDao *web.CrawlRuleDao) error {
	var triggerMsg svcweb.CrawlTriggerMsg
	if err := json.Unmarshal(msg.Data, &triggerMsg); err != nil {
		return fmt.Errorf("unmarshal trigger msg: %w", err)
	}

	taskDao := web.NewCrawlTaskDao()
	task, err := taskDao.GetByID(ctx, triggerMsg.TaskID)
	if err != nil {
		return fmt.Errorf("get task %d: %w", triggerMsg.TaskID, err)
	}
	if task == nil {
		logs.WarnContextf(ctx, "crawl task %d not found, skipping", triggerMsg.TaskID)
		return nil
	}

	if task.Status == web.CrawlTaskCancelled {
		logs.InfoContextf(ctx, "crawl task %d already cancelled, skipping", task.ID)
		return nil
	}

	now := time.Now()
	task.Status = web.CrawlTaskRunning
	task.StartedAt = &now
	if err := taskDao.UpdateStatus(ctx, task.ID, web.CrawlTaskRunning, ""); err != nil {
		return fmt.Errorf("update task %d to running: %w", task.ID, err)
	}

	rules, err := ruleDao.ListByAppID(ctx, task.AppID)
	if err != nil {
		logs.ErrorContextf(ctx, "list rules for app %d: %v", task.AppID, err)
	}

	result := Crawl(ctx, task, rules, cfg.CancelCheckInterval)

	finishedAt := time.Now()
	if result.Err != nil {
		if updateErr := taskDao.UpdateStatus(ctx, task.ID, web.CrawlTaskFailed, result.Err.Error()); updateErr != nil {
			logs.ErrorContextf(ctx, "update task %d failed status error: %v", task.ID, updateErr)
		}
		if task.StartedAt != nil {
			task.FinishedAt = &finishedAt
		}
		return result.Err
	}

	if updateErr := taskDao.UpdateProgress(ctx, task.ID, result.CrawledCount, result.CrawledCount, result.NewCount, result.UpdatedCount, result.SkippedCount); updateErr != nil {
		logs.ErrorContextf(ctx, "update task %d progress error: %v", task.ID, updateErr)
	}

	finalStatus := web.CrawlTaskSuccess
	pendingTask, checkErr := taskDao.GetByID(ctx, task.ID)
	if checkErr == nil && pendingTask != nil && pendingTask.Status == web.CrawlTaskCancelled {
		finalStatus = web.CrawlTaskCancelled
	}

	if statusErr := taskDao.UpdateStatus(ctx, task.ID, finalStatus, ""); statusErr != nil {
		logs.ErrorContextf(ctx, "update task %d final status error: %v", task.ID, statusErr)
	}

	logs.InfoContextf(ctx, "crawl task %d completed: crawled=%d new=%d updated=%d skipped=%d",
		task.ID, result.CrawledCount, result.NewCount, result.UpdatedCount, result.SkippedCount)
	return nil
}

func ensureCrawlStream(js nats.JetStreamContext, cfg Config) error {
	_, err := js.StreamInfo(cfg.StreamName)
	if err == nil {
		return nil
	}
	if err != nats.ErrStreamNotFound {
		return fmt.Errorf("check stream %s: %w", cfg.StreamName, err)
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     cfg.StreamName,
		Subjects: []string{cfg.Subject},
		Storage:  nats.FileStorage,
		MaxAge:   24 * time.Hour,
		MaxMsgs:  100000,
		MaxBytes: 256 * 1024 * 1024,
	})
	if err != nil {
		return fmt.Errorf("create stream %s: %w", cfg.StreamName, err)
	}

	logs.Infof("created NATS JetStream stream: %s", cfg.StreamName)
	return nil
}

func recoverPendingTasks(ctx context.Context, nc *nats.Conn, cfg Config) error {
	taskDao := web.NewCrawlTaskDao()
	tasks, err := taskDao.GetPendingAndRunning(ctx)
	if err != nil {
		return fmt.Errorf("get pending/running tasks: %w", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	logs.Infof("recovering %d pending/running crawl tasks", len(tasks))

	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("create JetStream for recovery: %w", err)
	}

	for _, task := range tasks {
		triggerMsg := svcweb.CrawlTriggerMsg{TaskID: task.ID}
		data, marshalErr := json.Marshal(triggerMsg)
		if marshalErr != nil {
			logs.ErrorContextf(ctx, "marshal recovery msg for task %d: %v", task.ID, marshalErr)
			continue
		}
		if _, pubErr := js.Publish(cfg.Subject, data); pubErr != nil {
			logs.ErrorContextf(ctx, "publish recovery msg for task %d: %v", task.ID, pubErr)
			continue
		}
		logs.Infof("recovered crawl task %d (status=%s)", task.ID, task.Status)
	}

	return nil
}
