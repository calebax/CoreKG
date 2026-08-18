package main

import (
	"net"
	"time"

	"github.com/insmtx/corekg/apps/keinit/models/helper"
	"github.com/insmtx/corekg/apps/keinit/models/minio"
	"github.com/insmtx/corekg/resource/locales"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/apis/runtime/server"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
)

var (
	configFile  string
	sqlDir      string
	cozeSqlDir  string
	envFile     string
	settingFile string
	esDSLDir    string
	envs        map[string]string

	rootCmd = &cobra.Command{
		Use: "keinit",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadCoreConfig(configFile)
			if err != nil {
				logs.WarnContextf(cmd.Context(), "[main] load config failed, %s", err)
			}
			if cfg == nil {
				logs.WarnContextf(cmd.Context(), "[main] use default config.")
				cfg = &config.DefaultConfig
			}

			logs.ReloadConfig(cfg.MainConf.App, cfg.LogsConf)
			logs.DebugContextf(cmd.Context(), "[main] config: %+v\n", cfg)

			if envFile != "" {
				envs, err = helper.ReadENV(cmd.Context(), envFile)
				if err != nil {
					logs.FatalContextf(cmd.Context(), "[main] load env file failed, %s", err)
					return
				}
				logs.DebugContextf(cmd.Context(), "[main] envs: %+v\n", envs)
			} else {
				envs, err = helper.ReadENV(cmd.Context())
				if err != nil {
					logs.FatalContextf(cmd.Context(), "[main] load env file failed, %s", err)
					return
				}
				logs.DebugContextf(cmd.Context(), "[main] envs: %+v\n", envs)
			}
		},
	}
)

func main() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "./config/config.yaml", "configurate file path.")
	rootCmd.PersistentFlags().StringVarP(&envFile, "env-file", "e", "", "env file path.")
	rootCmd.PersistentFlags().StringVar(&settingFile, "setting-file", "./config/core_setting.yaml", "core setting file path.")
	rootCmd.PersistentFlags().StringVar(&esDSLDir, "es-dsl-file", "./scripts/es", "es dsl dir path.")

	rootCmd.Flags().StringVarP(&sqlDir, "sql-dir", "s", "./scripts/mysql", "sql dir path.")
	rootCmd.Run = mainRun()
	rootCmd.AddCommand(initMysqlEnvCmd())
	rootCmd.AddCommand(updateSettingCmd())
	rootCmd.AddCommand(initESCmd())
	rootCmd.AddCommand(updateChatModelCmd())
	rootCmd.AddCommand(updateApiKeyCmd())
	rootCmd.AddCommand(runMigratorCmd())
	rootCmd.Execute()
}

func mainRun() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := logs.WithContextFields(cmd.Context(), "cmd", "main")

		var (
			err error
			// esRedo    = 10
			minioRedo = 10
		)

		if err := initMysqlEnv(ctx); err != nil {
			logs.ErrorContextf(ctx, "[main] init database failed, %s", err)
			return
		}

		if err := updateChatModel(ctx); err != nil {
			logs.FatalContextf(ctx, "[main] init database failed, %s", err)
			return
		}

		if err := updateApiKey(ctx); err != nil {
			logs.FatalContextf(ctx, "[main] init database failed, %s", err)
			return
		}

		if err := runMigrator(ctx); err != nil {
			logs.ErrorContextf(ctx, "[main] run migrator failed, err: %s", err)
			return
		}

		err = initES(ctx)
		if err != nil {
			logs.ErrorContextf(ctx, "[main] InitESMapping failed, %s", err)
			return
		}

		for i := 0; i < minioRedo; i++ {
			err = minio.CreateBucket(ctx)
			if err != nil {
				logs.ErrorContextf(ctx, "[main] InitESMapping CreateBucket failed, %s", err)
				time.Sleep(time.Second * 15)
				continue
			}
			break
		}
		for i := 0; i < minioRedo; i++ {
			err = minio.CreateCozeBucket(ctx)
			if err != nil {
				logs.ErrorContextf(ctx, "[main] InitESMapping CreateCozeBucket failed, %s", err)
				time.Sleep(time.Second * 15)
				continue
			}
			break
		}
		logs.InfoContextf(ctx, "Seccess init DB and create Bucket")

		cfg := config.Conf()
		i18n.Init(locales.I18nConfig, locales.TranslationFs)
		l, err := net.Listen("tcp", cfg.MainConf.HttpAddr)
		if err != nil {
			logs.FatalContextf(ctx, "[main] listen at %s failed, %s", cfg.MainConf.HttpAddr, err)
			return
		}
		{
			svr := server.NewRouter("/v3/")
			svr.G("/status.GetClusterID", GetClusterID)
			svr.G("/status.Ping", Ping)
			logs.InfoContextf(ctx, "[main] start http server at %s", cfg.MainConf.HttpAddr)
			if err := svr.Run(l); err != nil {
				logs.ErrorContextf(ctx, "[main] run server failed, %s", err)
				return
			}
		}
		lifecycle.Std().WaitExit()
	}
}
