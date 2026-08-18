package jobs

import (
	"bytes"
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

var stdCJMgr *cronJobManager

func InitJobServer(db *gorm.DB, lf *lifecycle.LifeCycle) error {
	if stdCJMgr != nil {
		return nil
	}
	db.Migrator().AutoMigrate(
		&CronJob{},
		&CronJobRecord{},
	)
	stdCJMgr = &cronJobManager{
		db:       db,
		lf:       lf,
		c:        cron.New(),
		cronJobs: map[string]*cronJobWrapper{},
	}
	go stdCJMgr.Start()
	return nil
}

type cronJobManager struct {
	db *gorm.DB
	lf *lifecycle.LifeCycle
	c  *cron.Cron

	cronJobs map[string]*cronJobWrapper
}

func (m *cronJobManager) Start() error {
	m.c.Start()
	return nil
}

func (m *cronJobManager) Register(cj *CronJob, jobFunc JobFunc) error {
	if cw, ok := m.cronJobs[cj.Name]; ok {
		m.c.Remove(cw.entryID)
	}
	nextRunTime, err := nextRunTime(cj.CronStr)
	if err != nil {
		logs.ErrorContextf(m.lf.Context(), "[cronjob][%s] get next run time failed, %s", cj.Name, err)
		return err
	}
	cj.NextRunTime = nextRunTime

	if err := insertOrUpdateCronJob(m.db, cj); err != nil {
		logs.ErrorContextf(m.lf.Context(), "[cronjob][%s] update cronjob failed, %s", cj.Name, err)
		return err
	}

	cw := &cronJobWrapper{
		c:       m.c,
		cj:      cj,
		jobFunc: jobFunc,
		db:      m.db,
	}
	eID, err := m.c.AddFunc(cj.CronStr, cw.doJobFunc)
	if err != nil {
		logs.ErrorContextf(m.lf.Context(), "[cronjob][%s] add cronjob failed, %s", cj.Name, err)
		return err
	}
	cw.entryID = eID

	return nil
}

type cronJobWrapper struct {
	c       *cron.Cron
	cj      *CronJob
	entryID cron.EntryID
	jobFunc JobFunc
	db      *gorm.DB
}

func (cw *cronJobWrapper) doJobFunc() {
	ctx := context.TODO()
	dbCj, err := getCronJob(cw.db, cw.cj.Name)
	if err != nil {
		logs.ErrorContextf(ctx, "[cronjob][%s] get cronjob failed, %s", cw.cj.Name, err)
		return
	}
	if !dbCj.IsEnabled.Value() {
		logs.InfoContextf(ctx, "[cronjob][%s] cronjob is disabled", cw.cj.Name)
		return
	}
	if time.Now().Before(dbCj.NextRunTime) {
		logs.ErrorContextf(ctx, "[cronjob][%s] next run time is not ready, next run time: %s", cw.cj.Name, dbCj.NextRunTime)
		return
	}

	record := &CronJobRecord{
		CronJobID:  dbCj.ID,
		RunTime:    dbCj.NextRunTime,
		Status:     CronJobStatusRunning,
		RunnerName: ownRunnerName,
		StartAt:    time.Now(),
	}
	entry := cw.c.Entry(cw.entryID)
	err = cw.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			logs.WarnContextf(ctx, "[cronjob][%s] create cronjob record failed, %s", dbCj.Name, err)
			return err
		}
		err := tx.Table(TableNameCronJobs).
			Where("id = ?", dbCj.ID).
			Updates(map[string]interface{}{
				"next_run_time": entry.Next,
				"last_run_time": record.RunTime,
				"run_count":     gorm.Expr("run_count + 1"),
			}).Error
		if err != nil {
			logs.WarnContextf(ctx, "[cronjob][%s] update cronjob next run time failed, %s", dbCj.Name, err)
			return err
		}
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[cronjob][%s] create cronjob record failed, skip", dbCj.Name)
		return
	}

	out := new(bytes.Buffer)
	record.Status = CronJobStatusSuccess
	if err := cw.jobFunc(ctx, out); err != nil {
		record.Error = err.Error()
		record.Status = CronJobStatusFailed
		logs.ErrorContextf(ctx, "[cronjob][%s] do job func failed, %s", dbCj.Name, err)
	}
	record.Output = out.String()
	record.CostSeconds = time.Now().Sub(record.StartAt).Seconds()
	if err := cw.db.Save(record).Error; err != nil {
		logs.ErrorContextf(ctx, "[cronjob][%s] update cronjob record failed, %s", dbCj.Name, err)
	}
	err = cw.db.Table(TableNameCronJobs).
		Where("id = ?", dbCj.ID).
		Updates(map[string]interface{}{
			"last_run_status":   record.Status,
			"last_run_error":    record.Error,
			"last_run_output":   record.Output,
			"last_cost_seconds": record.CostSeconds,
			"last_runner_name":  record.RunnerName,
		}).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[cronjob][%s] update cronjob next run time failed, %s", dbCj.Name, err)
	}

	logs.InfoContextf(ctx, "[cronjob][%s] do job func finished, status: %s, next run time: %s",
		dbCj.Name, record.Status, entry.Next)
}

// nextRunTime returns the next time this job is scheduled to run.
func nextRunTime(cronStr string) (time.Time, error) {
	p := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	cs, err := p.Parse(cronStr)
	if err != nil {
		return time.Now(), err
	}
	return cs.Next(time.Now()), nil
}
