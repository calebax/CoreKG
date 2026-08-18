package main

import (
	"net"

	appconf "github.com/insmtx/corekg/apps/websearch/conf"
	"github.com/insmtx/corekg/apps/websearch/internal/apis"
	"github.com/insmtx/corekg/apps/websearch/services/svcsearch"
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
	Use: "websearch",
	RunE: func(cmd *cobra.Command, _ []string) error {
		coreConfig, err := config.LoadCoreConfig(configFile)
		if err != nil {
			return err
		}
		if err := logs.ReloadConfig(coreConfig.MainConf.App, coreConfig.LogsConf); err != nil {
			return err
		}
		searchConfig, err := appconf.Load(configFile)
		if err != nil {
			return err
		}
		effectiveAPIKey, err := appconf.ResolveAPIKey(searchConfig.APIKey, apiKey)
		if err != nil {
			return err
		}
		searchRuntime, err := svcsearch.NewRuntime(searchConfig)
		if err != nil {
			return err
		}
		defer searchRuntime.Close()
		handler, err := apis.NewHandler(apis.HandlerOptions{
			Searcher: searchRuntime.Service, Cursor: searchRuntime.Cursor,
			Timeout: searchConfig.TotalTimeout, MaxTimeout: searchConfig.MaxRequestTimeout,
			CacheBypass: searchConfig.CacheBypass, AllowRequestProviders: searchConfig.AllowRequestProviders,
			EnabledProviders: searchConfig.EnabledProviders, ProviderVisibility: searchConfig.ProviderVisibility,
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
		logs.InfoContextf(cmd.Context(), "[main] websearch listening at %s", coreConfig.MainConf.HttpAddr)
		lifecycle.Std().WaitExit()
		return nil
	},
}

func main() {
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "./config/config.yaml", "configuration file path")
	rootCmd.Flags().StringVar(&apiKey, "api-key", "", "override auth.api_key from the configuration file")
	if err := rootCmd.Execute(); err != nil {
		logs.Fatalf("websearch failed: %v", err)
	}
}
