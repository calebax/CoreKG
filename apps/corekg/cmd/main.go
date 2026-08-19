package main

import (
	"context"
	"net"
	"time"

	"github.com/insmtx/corekg/apps/corekg"
	"github.com/insmtx/corekg/apps/corekg/internal/jobs"
	"github.com/insmtx/corekg/apps/corekg/internal/taskbiz"
	"github.com/insmtx/corekg/apps/corekg/mds"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kecore"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/kecore/models/nbgraph"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/insmtx/corekg/apps/kesearch/models/chunk"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/apps/kesearch/models/globalsearch"
	wfstartup "github.com/insmtx/corekg/apps/workflow/startup"
	"github.com/insmtx/corekg/resource/locales"
	"github.com/insmtx/corekg/pkgs/connectors"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/version"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/apis/runtime/server"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
)

var (
	configFile string

	rootCmd = &cobra.Command{
		Use: "corekg",
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
		ctx := cmd.Context()
		_, err := initDatabase(ctx, cfg, migrateDB)
		if err != nil {
			logs.ErrorContextf(ctx, "[main] init database failed, %s", err)
			return
		}

		// 初始化插件
		dbutil.InitializePlugins()

		initTask(ctx)

		// 可选：初始化 NATS 任务桥接，使上传文档后的任务能派发（严格 NATS 派发）。
		// NATS 不可用时仅告警不阻断启动，HTTP 轮询路径不依赖 NATS。
		if nc, natErr := initNATS(ctx); natErr != nil {
			logs.WarnContextf(ctx, "[main] init NATS task bridge failed (continue): %s", natErr)
		} else if nc != nil {
			defer nc.Close()
		}

		// 注册文档摄入链路回调，使 HTTP worker 回报后能推进下一阶段 / 更新文件状态。
		taskbiz.RegisterForestCallbacks(ctx)

		opts := []server.RouterOption{
			server.WithPrefixes([]string{global.PrefixAPIV2}),
			server.WithDeployMode(version.DeployMode()),
		}
		if global.GetEnableLicenseCheckBool() && version.DeployMode() != global.DeployModeTencentFree {
			opts = append(opts, server.WithMiddleware(mds.LicenseCheck(
				//corekg api
				"corekg.CheckLicense",
				"corekg.GetLicenseInfo",
				"corekg.RegisterLicense",
				//corekg docs
				"/v2/corekg.docs",
				"/v3/corekg.docs",
				//keparser api
				"knowledge.GetPendingTask",
				"knowledge.TaskCallBack",
				"knowledge.CheckInstance",
				"knowledge.GetInstanceInfo",
				"knowledge.TaskCheckQueueCount",
				// auth page
				"account.UploadOrganizeLogo",
				"account.UploadWebSiteLogo",
				"account.GetGlobalInfo",
				"account.GetCompanyInfo",
				"forest.GetCommonInfo",
				// login setting
				"account.GetLoginSetting",
				"account.LoginByPasswordPrivate",
				//default passwd
				"account.ChangeDefaultPassword",
				"account.ChangePasswordNotice",
				"account.ChooseUin",
				"account.GetCompanyAdmins",
				"forest.GetMessageCount",
				"account.ListUin",
				"account.DetailPersonalCenter",
			)))
		}

		svr := server.NewRouter(
			global.PrefixAPIV3,
			opts...,
		)

		// 初始化文件存储
		err = fs.InitForestStorage()
		if err != nil {
			logs.FatalContextf(cmd.Context(), "[main] InitForestStorage failed, %s", err)
			return
		}
		if err := essearch.InitEbConfig(ctx); err != nil {
			logs.FatalContextf(cmd.Context(), "[main] InitEbConfig failed, %s", err)
		}

		if err := chatquestion.InitHistoryESClient(ctx); err != nil {
			logs.FatalContextf(cmd.Context(), "[main] InitHistoryESClient failed, %s", err)
		}
		if err := chunk.InitESClient(ctx); err != nil {
			logs.FatalContextf(cmd.Context(), "[main] InitESClient failed, %s", err)
		}

		if version.DeployMode() != global.DeployModeOpenPO && global.GetEnableNebulaGraphBool() {
			err = nbgraph.InitNebulaConf(ctx)
			if err != nil {
				logs.FatalContextf(cmd.Context(), "[main] InitNebulaConf failed, %s", err)
				return
			}
			err = nebulagraph.InitNebulaConf(ctx)
			if err != nil {
				logs.FatalContextf(cmd.Context(), "[main] InitNebulaConf failed, %s", err)
				return
			}
		} else {
			logs.InfoContextf(cmd.Context(), "[main] NebulaGraph is disabled, skip initializing nebula graph config.")
		}

		if err := globalsearch.InitHighLightConfig(ctx); err != nil {
			logs.FatalContextf(cmd.Context(), "[main] InitHighLightConfig failed, %s", err)
		}
		err = connectors.InitProviders(cmd.Context(), "account", "pkl_connect_providers")
		if err != nil {
			logs.FatalContextf(cmd.Context(), "InitProviders error: %v", err)
			return
		}

		i18n.Init(locales.I18nConfig, locales.TranslationFs)

		kecore.RunJob(ctx)

		// 按配置决定是否在 corekg 进程内拉启 workflow Hertz server（双 server 聚合）。
		// 默认 enabled=false 时不启动，行为与改造前一致。
		maybeStartWorkflow(ctx, configFile)

		l, err := net.Listen("tcp", cfg.MainConf.HttpAddr)
		if err != nil {
			logs.FatalContextf(cmd.Context(), "[main] listen at %s failed, %s", cfg.MainConf.HttpAddr, err)
			return
		}
		corekg.Routers(svr)
		logs.InfoContextf(cmd.Context(), "[main] start http server at %s", cfg.MainConf.HttpAddr)
		if err := svr.Run(l); err != nil {
			logs.FatalContextf(cmd.Context(), "[main] run server failed, %s", err)
			return
		}
		jobs.RunRoutines(lifecycle.Std().Context())
		lifecycle.Std().WaitExit()
	}
}

// maybeStartWorkflow 依据配置的 workflow.enabled 决定是否在 corekg 进程内拉启
// workflow Hertz server。默认关闭不启动；开启后按 required 决定失败语义。
func maybeStartWorkflow(ctx context.Context, configFile string) {
	appCfg, enabled, ok := loadWorkflowConfig(configFile)
	if !ok {
		return
	}
	if !enabled {
		logs.InfoContextf(ctx, "[main] workflow is disabled, skip starting workflow server.")
		return
	}

	err := wfstartup.Start(ctx, appCfg)
	if err != nil {
		if appCfg.Workflow.Required {
			logs.FatalContextf(ctx, "[main] start workflow server failed (required): %v", err)
			return
		}
		logs.ErrorContextf(ctx, "[main] start workflow server failed (module unavailable, continue): %v", err)
		return
	}
	logs.InfoContextf(ctx, "[main] workflow server started on %s", appCfg.Workflow.HttpAddr)
}
