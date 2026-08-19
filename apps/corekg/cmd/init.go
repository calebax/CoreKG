package main

import (
	"context"
	"os"

	wfconf "github.com/insmtx/corekg/apps/workflow/conf"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	apptype "github.com/insmtx/corekg/apps/keapp/models/apptype"
	webmodels "github.com/insmtx/corekg/apps/keapp/models/web"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/pkgs/wecoms"
	"github.com/nats-io/nats.go"
	"github.com/ygpkg/yg-go/cache"
	"github.com/ygpkg/yg-go/cache/redis"
	ygconfig "github.com/ygpkg/yg-go/config"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"gorm.io/gorm"
)

func initDatabase(ctx context.Context, cfg *ygconfig.CoreConfig, migrateDB bool) (*gorm.DB, error) {
	err := dbtools.InitMultiDBConn(cfg.MainConf.DatabaseConns)
	if err != nil {
		logs.ErrorContextf(ctx, "[main] connect mysql failed, %s", err)
		return nil, err
	}
	db := dbutil.Core()
	db.Logger = logs.GetGorm("gorm")

	err = redispool.InitRedis("knowledge", "redis")
	if err != nil {
		logs.ErrorContextf(ctx, "[main] connect redis failed, %s", err)
		// return nil, err
	} else {
		cache.InitCache(redis.NewCache(redispool.Redis()))
	}

	if !migrateDB {
		return db, nil
	}

	err = dbtools.DoInitModels(
		foresttype.InitDB,
		chattype.InitDB,
		accounttype.InitDB,
		apptype.InitDB,
		webmodels.InitDB,
		wecoms.InitDB,
		settings.InitDB,
	)
	if err != nil {
		logs.ErrorContextf(ctx, "[main] init database failed, %s", err)
		return nil, err
	}

	return db, nil
}

func initTask(ctx context.Context) {
	err := task.InitDB()
	if err != nil {
		logs.FatalContextf(ctx, "init task db failed, %s", err)
		return
	}
}

// initNATS 初始化 NATS 任务桥接，使 PushTaskQueue 能向 JetStream dispatch/result
// 流派发任务（corekg 单体内仅有 HTTP worker 消费，NATS 供自有 worker/ketask 消费）。
// 优先读环境变量 NATS_URL，缺省使用本地 docker-compose 映射地址。
func initNATS(ctx context.Context) (*nats.Conn, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4225"
	}
	opts := []nats.Option{
		nats.Name("corekg"),
		nats.MaxReconnects(-1),
	}
	nc, err := nats.Connect(natsURL, opts...)
	if err != nil {
		return nil, err
	}
	bridge := task.NewNATSBridge(nc)
	task.SetNATSBridge(bridge)
	if streamErr := bridge.EnsureStreams(); streamErr != nil {
		return nil, streamErr
	}
	logs.Infof("[main] NATS task bridge connected: %s", natsURL)
	return nc, nil
}

// loadWorkflowConfig 从同一 configFile 独立解析 workflow 配置段。
//
// yg-go 的 CoreConfig loader 可能丢弃未知的 workflow 段，因此这里用
// workflow 自己的 AppConfig 结构对同一文件做二次解析，仅消费 workflow
// 需要的字段，不覆盖 corekg 的 main / database_conns。
// 返回值：(workflow 配置, 是否启用 workflow, 解析是否成功)。
func loadWorkflowConfig(configFile string) (*wfconf.AppConfig, bool, bool) {
	appCfg := &wfconf.AppConfig{}
	if configFile == "" {
		return nil, false, false
	}
	err := ygconfig.LoadYamlLocalFile(configFile, appCfg)
	if err != nil {
		logs.Errorf("[main] load workflow config failed: %v", err)
		return nil, false, false
	}
	return appCfg, appCfg.Workflow.Enabled, true
}
