package main

import (
	"net"

	appconf "github.com/insmtx/corekg/apps/webfetch/conf"
	"github.com/insmtx/corekg/apps/webfetch/internal/apis"
	"github.com/insmtx/corekg/apps/webfetch/services/svcfetch"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/apis/runtime/server"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
)

var (
	configFile string
	apiKey     string
)

var rootCmd = &cobra.Command{
	Use: "webfetch",
	RunE: func(cmd *cobra.Command, _ []string) error {
		coreConfig, err := config.LoadCoreConfig(configFile)
		if err != nil {
			return err
		}
		if err := logs.ReloadConfig(coreConfig.MainConf.App, coreConfig.LogsConf); err != nil {
			return err
		}
		fetchConfig, err := appconf.Load(configFile)
		if err != nil {
			return err
		}
		effectiveAPIKey, err := appconf.ResolveAPIKey(fetchConfig.APIKey, apiKey)
		if err != nil {
			return err
		}
		fetchRuntime, err := svcfetch.NewRuntime(fetchConfig)
		if err != nil {
			return err
		}
		defer fetchRuntime.Close()
		handler, err := apis.NewHandler(apis.HandlerOptions{
			Reader: fetchRuntime.Service, Timeout: fetchConfig.RequestTimeout,
			MaxTimeout: fetchConfig.MaxRequestTimeout, CacheBypass: fetchConfig.CacheBypass,
			LogURLQuery: fetchConfig.LogStoreURLQuery,
		})
		if err != nil {
			return err
		}

		listener, err := net.Listen("tcp", coreConfig.MainConf.HttpAddr)
		if err != nil {
			return err
		}
		router := server.NewRouter("/v3/")
		apis.RegistryRouter(router, handler, effectiveAPIKey)
		if err := router.Run(listener); err != nil {
			return err
		}
		logs.InfoContextf(cmd.Context(), "[main] webfetch listening at %s", coreConfig.MainConf.HttpAddr)
		lifecycle.Std().WaitExit()
		return nil
	},
}

func main() {
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "./config/config.yaml", "configuration file path")
	rootCmd.Flags().StringVar(&apiKey, "api-key", "", "override auth.api_key from the configuration file")
	if err := rootCmd.Execute(); err != nil {
		logs.Fatalf("webfetch failed: %v", err)
	}
}
