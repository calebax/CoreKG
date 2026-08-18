package mysql

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/insmtx/corekg/apps/workflow/conf"
	"github.com/ygpkg/yg-go/logs"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

func New() (*gorm.DB, error) {
	appCfg := conf.GetAppConfig()
	dsn, ok := appCfg.MainConf.DatabaseConns["opencoze"]
	if !ok || dsn == "" {
		return nil, fmt.Errorf("database_conns[opencoze] not found in config")
	}

	return newWithDSN(dsn, 10, 100, 3600, 600)
}

func newWithDSN(dsn string, maxIdle, maxOpen, maxLifetime, maxIdleTime int) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		return nil, fmt.Errorf("mysql open, dsn: %s, err: %w", dsn, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logs.Errorf("InitDB. db.DB() fail. err:%v", err)
		return nil, err
	}

	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(maxIdleTime) * time.Second)

	return db, nil
}

func InitCoreDB() error {
	appCfg := conf.GetAppConfig()
	dsn, ok := appCfg.MainConf.DatabaseConns["core"]
	if !ok || dsn == "" {
		return fmt.Errorf("database_conns[core] not found in config")
	}
	return dbtools.InitMultiDBConn(map[string]string{
		"core": dsn,
	})
}
