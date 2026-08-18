package main

import (
	"context"
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/config"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
)

func updateApiKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "update-api-key",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := logs.WithContextFields(context.Background(), "cmd", "update-api-key")
			logs.InfoContextf(ctx, "start update-api-key")
			if err := updateApiKey(ctx); err != nil {
				logs.ErrorContextf(ctx, "update-api-key cmd run failed, %s", err)
				return
			}
		},
	}
	return cmd
}

func updateApiKey(ctx context.Context) error {
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
	logs.InfoContextf(ctx, "[updateApiKey] connect mysql success")
	requiredEnvs := []string{
		"AGENT_MODEL_APIKEY",
	}
	// 检查所有必需的环境变量
	for _, key := range requiredEnvs {
		if err := checkEnv(envs, key); err != nil {
			logs.ErrorContextf(ctx, "check env failed, %s", err)
			return err
		}
	}
	logs.InfoContextf(ctx, "check env success")
	updateMap := map[string]any{
		"api_key":    envs["AGENT_MODEL_APIKEY"],
		"updated_at": time.Now(),
	}
	if err := dbtools.Account().WithContext(ctx).Model(&accounttype.APIKey{}).Where("name = ? AND purpose = ?", "prod", "prod").Updates(updateMap).Error; err != nil {
		logs.ErrorContextf(ctx, "update apiKey failed, %s", err)
		return err
	}
	return nil
}
