package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	hzconfig "github.com/cloudwego/hertz/pkg/common/config"
	"github.com/hertz-contrib/cors"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"

	"github.com/insmtx/corekg/apps/workflow/api/middleware"
	"github.com/insmtx/corekg/apps/workflow/api/router"
	"github.com/insmtx/corekg/apps/workflow/application"
	"github.com/insmtx/corekg/apps/workflow/conf"
	"github.com/insmtx/corekg/apps/workflow/utils/yygudb"
	ygconfig "github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
)

var (
	configFile string

	rootCmd = &cobra.Command{
		Use: "workflow",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			setCrashOutput()

			appCfg := &conf.AppConfig{}
			if configFile != "" {
				err := ygconfig.LoadYamlLocalFile(configFile, appCfg)
				if err != nil {
					logs.Errorf("load config file failed: %v", err)
					return
				}
			} else {
				loaded, err := ygconfig.LoadCoreConfigFromEnv()
				if err != nil {
					logs.Warnf("load remote config failed: %v, using defaults", err)
				}
				if loaded != nil {
					appCfg.CoreConfig = *loaded
				}
			}
			conf.SetAppConfig(appCfg)

			setLogLevel(appCfg.Workflow.LogLevel)
		},
	}
)

func main() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path")
	rootCmd.RunE = mainRun()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func mainRun() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		defer time.Sleep(time.Second)

		if err := yygudb.InitYyguDB(); err != nil {
			return fmt.Errorf("yygudb.InitYyguDB failed: %w", err)
		}

		if err := application.Init(ctx); err != nil {
			return fmt.Errorf("application.Init failed: %w", err)
		}

		startHttpServer()
		return nil
	}
}

func startHttpServer() {
	appCfg := conf.GetAppConfig()

	addr := appCfg.MainConf.HttpAddr
	if addr == "" {
		addr = ":8888"
	}

	maxSize := appCfg.Workflow.MaxRequestBodySize
	if maxSize == 0 {
		maxSize = 1024 * 1024 * 200
	}

	opts := []hzconfig.Option{
		server.WithHostPorts(addr),
		server.WithMaxRequestBodySize(int(maxSize)),
	}

	if appCfg.Workflow.SSL.Enabled {
		cert, err := tls.LoadX509KeyPair(
			appCfg.Workflow.SSL.CertFile,
			appCfg.Workflow.SSL.KeyFile,
		)
		if err != nil {
			logs.Errorf("load ssl cert failed: %v", err)
		}
		cfg := &tls.Config{}
		cfg.Certificates = append(cfg.Certificates, cert)
		opts = append(opts, server.WithTLS(cfg))
		logs.Infof("Use SSL")
	}

	s := server.Default(opts...)

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"*"}
	corsHandler := cors.New(corsConfig)

	s.Use(middleware.ContextCacheMW())
	s.Use(middleware.RequestInspectorMW())
	s.Use(middleware.SetHostMW())
	s.Use(middleware.SetLogIDMW())
	s.Use(corsHandler)
	s.Use(middleware.AccessLogMW())
	s.Use(middleware.OpenapiAuthMW())
	s.Use(middleware.SessionAuthMW())
	s.Use(middleware.I18nMW())

	router.GeneratedRegister(s)

	logs.Infof("start http server at %s", addr)
	s.Spin()
}

func setLogLevel(level string) {
	level = strings.ToLower(level)
	var lvl zapcore.Level
	switch level {
	case "trace", "debug":
		lvl = zapcore.DebugLevel
	case "info", "notice", "":
		lvl = zapcore.InfoLevel
	case "warn":
		lvl = zapcore.WarnLevel
	case "error":
		lvl = zapcore.ErrorLevel
	case "fatal":
		lvl = zapcore.FatalLevel
	default:
		lvl = zapcore.InfoLevel
	}
	logs.SetLevel(lvl)
	logs.Infof("log level: %s", level)
}

func setCrashOutput() {
	crashFile, _ := os.Create("crash.log")
	debug.SetCrashOutput(crashFile, debug.CrashOptions{})
}
