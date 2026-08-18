package main

import (
	"context"

	"github.com/insmtx/corekg/apps/keinit/models/mysql"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/config"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
)

func updateSettingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "update-setting",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := logs.WithContextFields(context.Background(), "cmd", "update-setting")
			logs.InfoContextf(ctx, "start init-mysql")
			if err := updateSetting(ctx); err != nil {
				logs.ErrorContextf(ctx, "update-setting failed, %s", err)
				return
			}
			logs.InfoContextf(ctx, "end update-setting")
		},
	}
	return cmd
}

func updateSetting(ctx context.Context) error {
	cfg := config.Conf()
	if cfg.MainConf.DatabaseConns == nil {
		logs.ErrorContextf(ctx, "mysql config is empty")
		return nil
	}
	err := dbtools.InitMultiDBConn(cfg.MainConf.DatabaseConns)
	if err != nil {
		logs.ErrorContextf(ctx, "connect mysql failed, %s", err)
		return err
	}
	logs.InfoContextf(ctx, "connect mysql success")

	err = mysql.UpsertCoreSettings(ctx, dbtools.Core(), settingFile)
	if err != nil {
		logs.ErrorContextf(ctx, "init mysql data failed: %s", err)
		return err
	}

	return nil
}
