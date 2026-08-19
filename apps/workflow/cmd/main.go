package main

import (
	"context"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"

	"github.com/insmtx/corekg/apps/workflow/conf"
	"github.com/insmtx/corekg/apps/workflow/startup"
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

		return startup.Run(ctx, conf.GetAppConfig())
	}
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
