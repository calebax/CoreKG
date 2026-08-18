package main

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/config"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

func updateChatModelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "update-chatmodel",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := logs.WithContextFields(context.Background(), "cmd", "update-chatmodel")
			logs.InfoContextf(ctx, "start update-chatmodel")
			if err := updateChatModel(ctx); err != nil {
				logs.FatalContextf(ctx, "update-chatmodel failed, %s", err)
				return
			}
			logs.InfoContextf(ctx, "end update-chatmodel")
		},
	}
	return cmd
}

func updateChatModel(ctx context.Context) error {
	cfg := config.Conf()
	if cfg.MainConf.DatabaseConns == nil {
		logs.ErrorContextf(ctx, "mysql config is empty")
		return fmt.Errorf("mysql config is empty")
	}
	err := dbtools.InitMultiDBConn(cfg.MainConf.DatabaseConns)
	if err != nil {
		logs.ErrorContextf(ctx, "connect mysql failed, %s", err)
		return err
	}
	logs.InfoContextf(ctx, "connect mysql success, %s", logs.JSON(envs))
	requiredEnvs := []string{
		"LLM_MODEL_SHOWNAME",
		"LLM_MODEL_APIKEY",
		"LLM_MODEL",
		"LLM_MODEL_URL",
	}
	// 检查所有必需的环境变量
	for _, key := range requiredEnvs {
		if err := checkEnv(envs, key); err != nil {
			logs.ErrorContextf(ctx, "check env failed, %s", err)
			return err
		}
	}
	logs.InfoContextf(ctx, "check env success")
	err = chatmodel.UpdateModel(ctx, &chattype.ChatModel{
		Model: gorm.Model{
			ID:        1,
			UpdatedAt: time.Now(),
			CreatedAt: time.Now(),
			DeletedAt: gorm.DeletedAt{},
		},
		Uin:                 1,
		CompanyID:           1,
		ShowName:            envs["LLM_MODEL_SHOWNAME"],
		APIKey:              envs["LLM_MODEL_APIKEY"],
		ModelName:           envs["LLM_MODEL"],
		ModelUrl:            envs["LLM_MODEL_URL"],
		PublecType:          "system",
		SupportFunctionCall: chattype.SupportFunctionCallSupported,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "update model failed, %s", err)
		return err
	}
	logs.InfoContextf(ctx, "update model success")
	return nil
}

func checkEnv(envs map[string]string, key string) error {
	if _, ok := envs[key]; !ok {
		return fmt.Errorf("env not found %s", key)
	}
	return nil
}
