package main

import (
	"context"

	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/cache"
	"github.com/ygpkg/yg-go/cache/redis"
	"github.com/ygpkg/yg-go/config"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

func initDatabase(ctx context.Context, cfg *config.CoreConfig, migrateDB bool) (*gorm.DB, error) {
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

	// err = dbtools.DoInitModels(
	// 	agentpdftype.InitDB,
	// )
	if err != nil {
		logs.ErrorContextf(ctx, "[main] init database failed, %s", err)
		return nil, err
	}

	return db, nil
}
