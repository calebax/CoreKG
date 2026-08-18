package main

import (
	"context"
	"net"
	"time"

	"github.com/insmtx/corekg/apps/keapp"
	"github.com/insmtx/corekg/apps/keapp/worker"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/apis/runtime/server"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/metrics"
)

var (
	configFile string

	rootCmd = &cobra.Command{
		Use: "keapp",
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
			logs.FatalContextf(cmd.Context(), "[main] init database failed, %s", err)
			return
		}

		nc, err := initNATS(cmd.Context(), configFile)
		if err != nil {
			logs.FatalContextf(cmd.Context(), "[main] init NATS failed, %s", err)
			return
		}
		defer nc.Close()

		var raw rawConfig
		if loadErr := config.LoadYamlLocalFile(configFile, &raw); loadErr != nil {
			logs.WarnContextf(cmd.Context(), "[main] load worker config failed, use defaults: %v", loadErr)
		}
		workerCfg := worker.DefaultConfig()

		metrics.Init(global.MetricNamespaceCoreKG, global.MetricSubsystemKeApp)

		l, err := net.Listen("tcp", cfg.MainConf.HttpAddr)
		if err != nil {
			logs.FatalContextf(cmd.Context(), "[main] listen at %s failed, %s", cfg.MainConf.HttpAddr, err)
			return
		}
		svr := server.NewRouter("/v3/")
		keapp.Routers(svr)
		keapp.RunJob()

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()
		keapp.StartWorker(ctx, nc, workerCfg)

		logs.Infof("[main] start http server at %s", cfg.MainConf.HttpAddr)
		if err := svr.Run(l); err != nil {
			logs.ErrorContextf(cmd.Context(), "[main] run server failed, %s", err)
			return
		}
		lifecycle.Std().WaitExit()
	}
}
