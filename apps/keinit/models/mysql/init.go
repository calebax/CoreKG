package mysql

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/apps/keinit/models/helper"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// InitMysqlData 初始化数据
func InitMysqlData(ctx context.Context, db *gorm.DB, scriptDir string, envs map[string]string) error {
	if db == nil {
		logs.ErrorContextf(ctx, "db is nil, script dir: %s", scriptDir)
		return fmt.Errorf("db is nil, script dir: %s", scriptDir)
	}
	if err := db.AutoMigrate(&InitExecRecourd{}); err != nil {
		logs.ErrorContextf(ctx, "auto migrate init exec record table error: %v", err)
		return err
	}

	{
		tplFiles, err := helper.ListFileWithExt(ctx, scriptDir, ".sqltpl")
		if err != nil {
			logs.ErrorContextf(ctx, "list tpl files error: %v", err)
			return err
		}
		for _, tplFile := range tplFiles {
			err := ExecuteVariableTpl(ctx, db, tplFile, envs)
			if err != nil {
				logs.ErrorContextf(ctx, "execute tpl file (%s) error: %v", tplFile, err)
				return err
			}
			logs.InfoContextf(ctx, "execute tpl file (%s) success", tplFile)
		}
	}

	{
		err := ExecuteScripts(ctx, db, scriptDir)
		if err != nil {
			logs.ErrorContextf(ctx, "execute sql files error: %v", err)
			return err
		}
		logs.InfoContextf(ctx, "execute sql files success")
	}
	return nil
}

// CheckConnect 检查连接是否正常
func CheckConnect(db *gorm.DB) error {
	originDb, err := db.DB()
	if err != nil {
		return err
	}
	err = originDb.Ping()
	if err != nil {
		return err
	}
	return nil
}
