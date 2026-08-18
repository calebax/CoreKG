package main

import (
	"context"
	"os"

	"github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/nats-io/nats.go"
	"github.com/ygpkg/yg-go/cache"
	"github.com/ygpkg/yg-go/cache/redis"
	"github.com/ygpkg/yg-go/config"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

var natsConn *nats.Conn

func initDatabase(ctx context.Context, cfg *config.CoreConfig, migrateDB bool) (*gorm.DB, error) {
	err := dbtools.InitMultiDBConn(cfg.MainConf.DatabaseConns)
	if err != nil {
		logs.ErrorContextf(ctx, "[main] connect mysql failed, %s", err)
		return nil, err
	}
	db := dbutil.Core()
	db.Logger = logs.GetGorm("gorm")

	err = redispool.InitRedis("ketask", "redis")
	if err != nil {
		logs.ErrorContextf(ctx, "[main] connect redis failed, %s", err)
		// return nil, err
	} else {
		cache.InitCache(redis.NewCache(redispool.Redis()))
	}

	if !migrateDB {
		return db, nil
	}

	if err != nil {
		logs.ErrorContextf(ctx, "[main] init database failed, %s", err)
		return nil, err
	}

	return db, nil
}

func initNATS(ctx context.Context) error {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	opts := []nats.Option{
		nats.Name("ketask"),
		nats.MaxReconnects(-1),
	}
	nc, err := nats.Connect(natsURL, opts...)
	if err != nil {
		return err
	}
	logs.InfoContextf(ctx, "[main] NATS connected: %s", natsURL)

	bridge := task.NewNATSBridge(nc)
	task.SetNATSBridge(bridge)

	if streamErr := bridge.EnsureStreams(); streamErr != nil {
		return streamErr
	}

	natsConn = nc
	return nil
}

func initTask(ctx context.Context) {
	err := task.InitDB()
	if err != nil {
		logs.FatalContextf(ctx, "init task db failed, %s", err)
		return
	}
}
