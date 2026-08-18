package main

import (
	"net"
	"time"

	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kesearch"
	"github.com/insmtx/corekg/apps/kesearch/models/chunk"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/apps/kesearch/models/globalsearch"
	"github.com/insmtx/corekg/resource/locales"
	"github.com/insmtx/corekg/pkgs/connectors"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/apis/runtime/server"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/metrics"
)

var (
	configFile string

	rootCmd = &cobra.Command{
		Use: "kesearch",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			cfg, err := config.LoadCoreConfig(configFile)
			if err != nil {
				logs.WarnContextf(ctx, "[main] load config failed, %s", err)
			}
			if cfg == nil {
				logs.WarnContextf(ctx, "[main] use default config.")
				cfg = &config.DefaultConfig
			}

			logs.ReloadConfig(cfg.MainConf.App, cfg.LogsConf)
			logs.DebugContextf(ctx, "[main] config: %+v\n", cfg)
		},
	}
)

func main() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "configurate file path.")

	rootCmd.Run = mainRun()
	rootCmd.Execute()
}

func mainRun() func(cmd *cobra.Command, args []string) {
	var migrateDB bool
	rootCmd.Flags().BoolVar(&migrateDB, "migrate-db", false, "auto migrate db struct before start.")

	return func(cmd *cobra.Command, args []string) {
		cfg := config.Conf()
		defer time.Sleep(time.Second)
		ctx := cmd.Context()
		_, err := initDatabase(ctx, cfg, migrateDB)
		if err != nil {
			logs.ErrorContextf(ctx, "[main] init database failed, %s", err)
			return
		}

		metrics.Init(global.MetricNamespaceCoreKG, global.MetricSubsystemKeSearch)

		// initTask()

		i18n.Init(locales.I18nConfig, locales.TranslationFs)
		l, err := net.Listen("tcp", cfg.MainConf.HttpAddr)
		if err != nil {
			logs.FatalContextf(ctx, "[main] listen at %s failed, %s", cfg.MainConf.HttpAddr, err)
			return
		}
		svr := server.NewRouter("/v3/")
		kesearch.Routers(svr)

		if err := essearch.InitEbConfig(ctx); err != nil {
			logs.FatalContextf(ctx, "[main] InitEbConfig failed, %s", err)
		}
		if err := chunk.InitESClient(ctx); err != nil {
			logs.FatalContextf(ctx, "[main] InitHistoryESClient failed, %s", err)
		}

		if err := globalsearch.InitHighLightConfig(ctx); err != nil {
			logs.FatalContextf(ctx, "[main] InitHighLightConfig failed, %s", err)
		}

		if err := chatquestion.InitHistoryESClient(ctx); err != nil {
			logs.FatalContextf(ctx, "[main] InitHistoryESClient failed, %s", err)
		}
		err = connectors.InitProviders(ctx, "account", "pkl_connect_providers")
		if err != nil {
			logs.ErrorContextf(ctx, "InitProviders error: %v", err)
			return
		}

		logs.InfoContextf(ctx, "[main] start http server at %s", cfg.MainConf.HttpAddr)
		if err := svr.Run(l); err != nil {
			logs.ErrorContextf(ctx, "[main] run server failed, %s", err)
			return
		}
		lifecycle.Std().WaitExit()
	}
}
