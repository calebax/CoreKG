package main

import (
	"context"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	apptype "github.com/insmtx/corekg/apps/keapp/models/apptype"
	webmodels "github.com/insmtx/corekg/apps/keapp/models/web"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/task"
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
