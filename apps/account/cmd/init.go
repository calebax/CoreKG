package main

import (
	"context"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/pkgs/connectors"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/pkgs/wecoms"
	"github.com/ygpkg/yg-go/cache"
	"github.com/ygpkg/yg-go/cache/redis"
	"github.com/ygpkg/yg-go/config"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"gorm.io/gorm"
)

func initDatabase(ctx context.Context, cfg *config.CoreConfig, migrateDB bool) (*gorm.DB, error) {
	err := dbtools.InitMultiDBConn(cfg.MainConf.DatabaseConns)
	if err != nil {
		logs.ErrorContextf(ctx, "[main] connect mysql failed, %s", err)
		return nil, err
	}
	db := dbutil.Account()
	db.Logger = logs.GetGorm("gorm")

	redisKey := "redis"
	if cfg.MainConf.Env == "dev" {
		redisKey = "loc_redis"
	}
	err = redispool.InitRedis("core", redisKey)
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
		accounttype.InitDB,
		wecoms.InitDB,
		settings.InitDB,
	)
	if err != nil {
		logs.ErrorContextf(ctx, "[main] init database failed, %s", err)
		return nil, err
	}

	return db, nil
}

func InitProviders(ctx context.Context, cfg *config.CoreConfig) error {
	return connectors.InitProviders(ctx, "account", "pkl_connect_providers")
}
