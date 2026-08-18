package tests

import (
	"context"

	"github.com/ygpkg/yg-go/config"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

var testdb *gorm.DB

func GetTestDB() (*gorm.DB, error) {
	if testdb != nil {
		return testdb, nil
	}
	ctx := context.TODO()
	cfg := config.Conf()
	db, err := dbtools.InitDBConn("", cfg.MainConf.DatabaseConns["default"])
	if err != nil {
		logs.ErrorContextf(ctx, "[main] connect mysql failed, %s", err)
		return nil, err
	}
	db.Logger = logs.GetGorm("gorm")
	return db, nil
}
