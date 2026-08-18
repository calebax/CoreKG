package main

import (
	"net"
	"time"

	"github.com/insmtx/corekg/apps/keapi"
	"github.com/insmtx/corekg/apps/keapi/conf"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/kesearch/models/chunk"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/apps/kesearch/models/globalsearch"
	"github.com/insmtx/corekg/resource/locales"
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
		Use: "keapi",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadCoreConfig(configFile)
			if err != nil {
				logs.WarnContextf(cmd.Context(), "[main] load config failed, %s", err)
			}
			if cfg == nil {
				logs.WarnContextf(cmd.Context(), "[main] use default config.")
				cfg = &config.DefaultConfig
			}

			conf.InitMCPCfg(configFile)
			logs.ReloadConfig(cfg.MainConf.App, cfg.LogsConf)
			logs.DebugContextf(cmd.Context(), "[main] config: %+v\n", cfg)
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

		_, err := initDatabase(cmd.Context(), cfg, migrateDB)
		if err != nil {
			logs.ErrorContextf(cmd.Context(), "[main] init database failed, %s", err)
			return
		}

		metrics.Init(global.MetricNamespaceCoreKG, global.MetricSubsystemKeAPI)

		if err := fs.InitForestStorage(); err != nil {
			logs.FatalContextf(cmd.Context(), "[main] InitForestStorage failed, %s", err)
			return
		}

		if err := chunk.InitESClient(cmd.Context()); err != nil {
			logs.FatalContextf(cmd.Context(), "[main] InitESClient failed, %s", err)
			return
		}

		if err := essearch.InitEbConfig(cmd.Context()); err != nil {
			logs.FatalContextf(cmd.Context(), "[main] InitEbConfig failed, %s", err)
			return
		}

		if err := chatquestion.InitHistoryESClient(cmd.Context()); err != nil {
			logs.FatalContextf(cmd.Context(), "[main] InitHistoryESClient failed, %s", err)
			return
		}

		if err := globalsearch.InitHighLightConfig(cmd.Context()); err != nil {
			logs.FatalContextf(cmd.Context(), "[main] InitHighLightConfig failed, %s", err)
			return
		}

		i18n.Init(locales.I18nConfig, locales.TranslationFs)
		l, err := net.Listen("tcp", cfg.MainConf.HttpAddr)
		if err != nil {
			logs.FatalContextf(cmd.Context(), "[main] listen at %s failed, %s", cfg.MainConf.HttpAddr, err)
			return
		}

		svr := server.NewRouter(global.PrefixAPIV3)
		keapi.Routers(svr)

		logs.InfoContextf(cmd.Context(), "[main] start http server at %s", cfg.MainConf.HttpAddr)
		if err := svr.Run(l); err != nil {
			logs.ErrorContextf(cmd.Context(), "[main] run server failed, %s", err)
			return
		}

		lifecycle.Std().WaitExit()
	}
}
