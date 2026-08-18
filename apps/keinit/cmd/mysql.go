package main

import (
	"context"
	"time"

	"github.com/insmtx/corekg/apps/keinit/models/mysql"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/config"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
)

func initMysqlEnvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "init-mysql",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := logs.WithContextFields(context.Background(), "cmd", "init-mysql")
			logs.InfoContextf(ctx, "start init-mysql")
			if err := initMysqlEnv(ctx); err != nil {
				logs.ErrorContextf(ctx, "init-mysql failed, %s", err)
				return
			}
			logs.InfoContextf(ctx, "end init-mysql")
		},
	}
	cmd.Flags().StringVarP(&sqlDir, "sql-dir", "s", "./scripts/mysql", "sql dir path.")
	cmd.Flags().StringVarP(&cozeSqlDir, "coze-sql-dir", "c", "./scripts/mysqlcoze", "coze sql dir path.")
	return cmd
}

func initMysqlEnv(ctx context.Context) error {
	cfg := config.Conf()
	if cfg.MainConf.DatabaseConns == nil {
		logs.ErrorContextf(ctx, "mysql config is empty")
		return nil
	}

	var err error
	for i := 0; ; i++ {
		err := dbtools.InitMultiDBConn(cfg.MainConf.DatabaseConns)
		if err != nil {
			logs.ErrorContextf(ctx, "connect mysql failed, %s retry %d times", err, i)
			time.Sleep(time.Second * 10)
			continue
		}

		if err := mysql.CheckConnect(dbtools.Core()); err != nil {
			logs.ErrorContextf(ctx, "ping mysql failed: %s", err)
			time.Sleep(time.Second * 10)
			continue
		}
		if err := mysql.CheckConnect(dbutil.Coze()); err != nil {
			logs.ErrorContextf(ctx, "ping coze mysql failed: %s", err)
			time.Sleep(time.Second * 10)
			continue
		}

		logs.InfoContextf(ctx, "connect mysql success")
		break
	}
	if err != nil {
		logs.ErrorContextf(ctx, "connect mysql failed: %s", err)
		return err
	}

	err = mysql.InitMysqlData(ctx, dbtools.Core(), sqlDir, envs)
	if err != nil {
		logs.ErrorContextf(ctx, "init mysql data failed: %s", err)
		return err
	}
	logs.InfoContextf(ctx, "init mysql data success")

	if err := mysql.InitMysqlData(ctx, dbutil.Coze(), cozeSqlDir, envs); err != nil {
		logs.ErrorContextf(ctx, "init coze mysql data failed: %s", err)
		return err
	}
	logs.InfoContextf(ctx, "init coze mysql data success")

	err = mysql.UpsertCoreSettings(ctx, dbtools.Core(), settingFile)
	if err != nil {
		logs.ErrorContextf(ctx, "upsert core settings failed: %s", err)
		return err
	}
	logs.InfoContextf(ctx, "upsert core settings success")

	return nil
}
