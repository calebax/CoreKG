package main

import (
	"context"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/admin/models/login_setting"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/pkgs/wecoms"
	"github.com/ygpkg/yg-go/config"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"github.com/ygpkg/yg-go/storage"
	"gorm.io/gorm"
)

func initDatabase(ctx context.Context, cfg *config.CoreConfig) (*gorm.DB, error) {
	err := dbtools.InitMultiDBConn(cfg.MainConf.DatabaseConns)
	if err != nil {
		logs.ErrorContextf(ctx, "[main] connect mysql failed, %s", err)
		return nil, err
	}
	db := dbutil.Core()
	db.Logger = logs.GetGorm("gorm")

	// 删除已有索引
	sql := `SET @idx_name := (
    SELECT COUNT(*)
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'chat_agent'
      AND index_name = 'name'
);

SET @sql := IF(@idx_name > 0, 'ALTER TABLE chat_agent DROP INDEX name;', 'SELECT 1;');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;`
	if err := dbutil.Core().Exec(sql).Error; err != nil {
		logs.ErrorContextf(ctx, "[main] drop index failed, %s", err)
	}

	// err = redispool.InitRedis("knowledge", "redis")
	// if err != nil {
	// 	logs.Errorf("[main] connect redis failed, %s", err)
	// 	// return nil, err
	// } else {
	// 	cache.InitCache(redis.NewCache(redispool.Redis()))
	// }

	logs.InfoContextf(ctx, "[main] connect mysql success")
	err = dbtools.DoInitModels(
		foresttype.InitDB,
		settings.InitDB,
		chattype.InitDB,
		accounttype.InitDB,
		wecoms.InitDB,
		login_setting.InitDB,
	)
	if err != nil {
		logs.FatalContextf(ctx, "[main] init database failed, %s", err)
		return nil, err
	}
	err = task.InitDB()
	if err != nil {
		logs.FatalContextf(ctx, "init task db failed, %s", err)
		return nil, err
	}
	err = storage.InitDB(db)
	if err != nil {
		logs.FatalContextf(ctx, "init storage db failed, %s", err)
		return nil, err
	}

	return db, nil
}
