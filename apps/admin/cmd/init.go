package main

import (
	"context"

	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/insmtx/corekg/apps/admin/models/login_setting"
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
		logs.ErrorContext(ctx, "[main] connect redis failed, %s", err)
		// return nil, err
	} else {
		cache.InitCache(redis.NewCache(redispool.Redis()))
	}

	if !migrateDB {
		return db, nil
	}

	err = dbtools.DoInitModels(
		admintype.InitDB,
		wecoms.InitDB,
		settings.InitDB,
		login_setting.InitDB,
		admintype.InitCoreDB,
	)
	if err != nil {
		logs.ErrorContextf(ctx, "[main] init database failed, %s", err)
		return nil, err
	}

	migrateEmployeeUser(ctx, dbutil.Account())

	return db, nil
}
