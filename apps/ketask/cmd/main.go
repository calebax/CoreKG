package main

import (
	"net"
	"time"

	"github.com/insmtx/corekg/apps/ketask"
	"github.com/insmtx/corekg/apps/ketask/internal/jobs"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/resource/locales"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/apis/runtime/server"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/metrics"
	"go.uber.org/zap/zapcore"
)

var (
	configFile string

	workerID        = lifecycle.OwnerID()
	taskType        string
	baseURL         string
	apiKey          string
	routineSize     int
	workerServerURL string
	debug           = false

	rootCmd = &cobra.Command{
		Use: "ketask",
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
			if debug {
				logs.SetLevel(zapcore.DebugLevel)
			}
			logs.DebugContextf(cmd.Context(), "[main] config: %+v\n", cfg)
		},
	}
)

func main() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "configurate file path.")
	rootCmd.PersistentFlags().StringVarP(&taskType, "task_type", "t", "", "Task type to process")
	rootCmd.PersistentFlags().StringVarP(&baseURL, "base_url", "b", "https://tapi.example.com/", "Base URL for the API")
	rootCmd.PersistentFlags().StringVarP(&apiKey, "api_key", "k", "", "API key for authentication")
	rootCmd.PersistentFlags().IntVarP(&routineSize, "worker_routine_size", "r", 1, "Number of concurrent routines to run")
	rootCmd.PersistentFlags().StringVarP(&workerServerURL, "worker_server_url", "a", "http://localhost:5000/local.Run", "Address for the worker server")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "Enable debug mode")

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

		if natErr := initNATS(cmd.Context()); natErr != nil {
			logs.ErrorContextf(cmd.Context(), "[main] init NATS failed, %s", natErr)
			return
		}

		metrics.Init(global.MetricNamespaceCoreKG, global.MetricSubsystemKeParser)
		initTask(cmd.Context())

		jobs.InitResultConsumer(natsConn)

		i18n.Init(locales.I18nConfig, locales.TranslationFs)
		l, err := net.Listen("tcp", cfg.MainConf.HttpAddr)
		if err != nil {
			logs.FatalContextf(cmd.Context(), "[main] listen at %s failed, %s", cfg.MainConf.HttpAddr, err)
			return
		}
		svr := server.NewRouter("/v3/")
		ketask.Routers(svr)

		logs.InfoContextf(cmd.Context(), "[main] start http server at %s", cfg.MainConf.HttpAddr)
		go func() {
			if err := svr.Run(l); err != nil {
				logs.ErrorContextf(cmd.Context(), "[main] run server failed, %s", err)
			}
		}()
		jobs.RunRoutines(lifecycle.Std().Context())
		lifecycle.Std().WaitExit()
	}
}
