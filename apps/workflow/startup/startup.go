// Package startup 提供 workflow 应用的可复用启动入口。
//
// workflow 可独立运行（`make run APP=workflow`），也可由 corekg 聚合进程拉启
// （同进程双 server 模式）。两者共用本包的初始化与 HTTP server 构建逻辑，
// 避免核心启动路径在多个入口间重复。
package startup

import (
	"context"
	"crypto/tls"
	"fmt"

	hzconfig "github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"

	"github.com/insmtx/corekg/apps/workflow/api/middleware"
	"github.com/insmtx/corekg/apps/workflow/api/router"
	"github.com/insmtx/corekg/apps/workflow/application"
	"github.com/insmtx/corekg/apps/workflow/conf"
	"github.com/insmtx/corekg/apps/workflow/utils/yygudb"
	"github.com/ygpkg/yg-go/logs"
)

// Start 初始化 workflow 并在本进程内拉起 Hertz HTTP server。
//
// 供 corekg 聚合进程调用：完成配置注入、DB 与应用初始化，随后**异步**启动
// HTTP server（内部 `go srv.Spin()`），返回后调用方（corekg 的主端口阻塞
// server）可继续。若初始化或 server 构建失败则返回错误。
func Start(ctx context.Context, appCfg *conf.AppConfig) error {
	conf.SetAppConfig(appCfg)

	if err := yygudb.InitYyguDB(); err != nil {
		return fmt.Errorf("startup: yygudb.InitYyguDB failed: %w", err)
	}
	if err := application.Init(ctx); err != nil {
		return fmt.Errorf("startup: application.Init failed: %w", err)
	}

	s, addr := buildServer(appCfg)
	router.GeneratedRegister(s)

	// 放入 goroutine 异步服务，避免阻塞调用方（corekg 主进程）后续启动流程。
	go func() {
		logs.Infof("start workflow http server at %s", addr)
		s.Spin()
	}()
	return nil
}

// Run 与 Start 等价，但会阻塞直到 HTTP server 退出。
//
// 供 workflow 独立二进制（`apps/workflow/cmd`）调用，行为与原 `cmd/main.go`
// 保持一致：启动后进程常驻。
func Run(ctx context.Context, appCfg *conf.AppConfig) error {
	conf.SetAppConfig(appCfg)

	if err := yygudb.InitYyguDB(); err != nil {
		return fmt.Errorf("startup: yygudb.InitYyguDB failed: %w", err)
	}
	if err := application.Init(ctx); err != nil {
		return fmt.Errorf("startup: application.Init failed: %w", err)
	}

	s, addr := buildServer(appCfg)
	router.GeneratedRegister(s)

	logs.Infof("start workflow http server at %s", addr)
	s.Spin()
	return nil
}

// buildServer 依据配置构建 Hertz server 并装配中间件，返回 server 与其监听地址。
//
// 构建逻辑自原 `cmd/main.go` 的 `startHttpServer` 迁移而来。
func buildServer(appCfg *conf.AppConfig) (*server.Hertz, string) {
	addr := appCfg.Workflow.HttpAddr
	if addr == "" {
		// 兼容独立运行形态：沿用 main.http_addr
		addr = appCfg.MainConf.HttpAddr
	}
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

	return s, addr
}
