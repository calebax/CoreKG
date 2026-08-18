package main

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account"
	"github.com/insmtx/corekg/resource/locales"
	"github.com/insmtx/corekg/pkgs/apis/wecom"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/version"
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
		Use: "account",
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
	rootCmd.AddCommand(version.VersionCmd("account"))
	rootCmd.AddCommand(testForwardAuthCmd())
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

		metrics.Init(global.MetricNamespaceCoreKG, global.MetricSubsystemAccount)

		err = InitProviders(cmd.Context(), cfg)
		if err != nil {
			logs.ErrorContextf(cmd.Context(), "[main] init providers failed, %s", err)
			return
		}

		i18n.Init(locales.I18nConfig, locales.TranslationFs)
		l, err := net.Listen("tcp", cfg.MainConf.HttpAddr)
		if err != nil {
			logs.FatalContextf(cmd.Context(), "[main] listen at %s failed, %s", cfg.MainConf.HttpAddr, err)
			return
		}

		svr := server.NewRouter(global.PrefixAPIV2)
		account.Routers(svr)
		wecom.RegistryRouter(svr.GinEngine().Group(svr.Prefix + "account"))
		// svr.RegistryIntoDB(db)

		logs.InfoContextf(cmd.Context(), "[main] start http server at %s", cfg.MainConf.HttpAddr)
		if err := svr.Run(l); err != nil {
			logs.ErrorContextf(cmd.Context(), "[main] run server failed, %s", err)
			return
		}
		lifecycle.Std().WaitExit()
	}
}

func testForwardAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test-forward-auth",
		Short: "test forward auth",
		Run: func(cmd *cobra.Command, args []string) {
			eng := gin.New()
			eng.Any("auth", func(ctx *gin.Context) {
				act := ExtractRequestAction(ctx.Request)
				logs.InfoContextf(ctx, "action: %s", act)

				// logs.Infof("url: %s", ctx.Request.Header.Get("X-Forwarded-Uri"))
				// logs.Infof("method: %s", ctx.Request.Method)
				logs.InfoContextf(ctx, "header: %s", ctx.Request.Header)
				// logs.Infof("body: %s", ctx.Request.Body)
				// logs.Infof("query: %s", ctx.Request.URL.Query())
				// logs.Infof("form: %s", ctx.Request.Form)
				// logs.Infof("remote addr: %s", ctx.Request.RemoteAddr)
				// // logs.Infof("local addr: %s", ctx.Request.LocalAddr)
				// logs.Infof("host: %s", ctx.Request.Host)
				// logs.Infof("proto: %s", ctx.Request.Proto)

				ctx.JSON(200, gin.H{
					"code": 0,
					"msg":  "ok",
				})
			})
			eng.Run(":30010")
		}}
	return cmd
}

type RequestAction struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Proto      string `json:"proto"`
	Method     string `json:"method"`
	Uri        string `json:"uri"`
	RemoteAddr string `json:"remote_addr"`
}

func (a RequestAction) String() string {
	return fmt.Sprintf("%s %s://%s:%d%s", a.Method, a.Proto, a.Host, a.Port, a.Uri)
}

// ExtractRequestAction 提取请求中的转发信息
func ExtractRequestAction(req *http.Request) *RequestAction {
	action := &RequestAction{
		Host:       req.Header.Get("X-Forwarded-Host"),
		Proto:      req.Header.Get("X-Forwarded-Proto"),
		Method:     req.Header.Get("X-Forwarded-Method"),
		Uri:        req.Header.Get("X-Forwarded-Uri"),
		RemoteAddr: req.Header.Get("X-Forwarded-For"),
	}

	if port := req.Header.Get("X-Forwarded-Port"); port != "" {
		action.Port, _ = strconv.Atoi(port)
	}

	return action
}
