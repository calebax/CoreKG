package testutils

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/kecore/models/nbgraph"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins/mysqlplugin"
	"github.com/ygpkg/yg-go/logs"
)

// kecoreInitializer 为 search 应用提供特定的初始化逻辑
type kecoreInitializer struct {
	*baseInitializer
}

// newKecoreInitializer 创建 search 应用的初始化器
func newKecoreInitializer() (Initializer, error) {
	base, err := newBaseInitializer(AppNameKecore)
	if err != nil {
		return nil, fmt.Errorf("create base initializer: %w", err)
	}

	return &kecoreInitializer{
		baseInitializer: base,
	}, nil
}

// Initialize 实现 search 应用特定的初始化逻辑
func (k *kecoreInitializer) Initialize() error {
	ctx := context.TODO()
	// 首先执行基础初始化
	if err := k.baseInitializer.initialize(); err != nil {
		return err
	}

	dbEngine := &dbplugins.Engine{}
	dbEngine.RegistryPlugin(mysqlplugin.NewMySQLPlugin())

	if err := fs.InitForestStorage(); err != nil {
		return fmt.Errorf("init forest storage failed: %w", err)
	}

	if err := nbgraph.InitNebulaConf(ctx); err != nil {
		return fmt.Errorf("init nebula failed: %w", err)
	}

	if err := essearch.InitEbConfig(ctx); err != nil {
		logs.FatalContextf(ctx, "[main] InitEbConfig failed, %s", err)
	}

	if err := chatquestion.InitHistoryESClient(ctx); err != nil {
		logs.FatalContextf(ctx, "[main] InitHistoryESClient failed, %s", err)
	}

	logs.InfoContextf(ctx, "kecore unit testing environment initialized successfully")
	return nil
}

// Close 实现 kecore 应用特定的关闭逻辑
func (k *kecoreInitializer) Close() error {
	// 在这里实现必要的关闭逻辑
	return k.baseInitializer.Close()
}
