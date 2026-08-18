package testutils

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/ygpkg/yg-go/logs"
)

// adminInitializer 为 search 应用提供特定的初始化逻辑
type adminInitializer struct {
	*baseInitializer
}

// newAdminInitializer 创建应用的初始化器
func newAdminInitializer() (Initializer, error) {
	base, err := newBaseInitializer(AppNameAdmin)
	if err != nil {
		return nil, fmt.Errorf("create base initializer: %w", err)
	}

	return &adminInitializer{
		baseInitializer: base,
	}, nil
}

// Initialize 实现应用特定的初始化逻辑
func (k *adminInitializer) Initialize() error {
	ctx := context.TODO()
	// 首先执行基础初始化
	if err := k.baseInitializer.initialize(); err != nil {
		return err
	}

	if err := essearch.InitEbConfig(ctx); err != nil {
		logs.FatalContextf(ctx, "[main] InitEbConfig failed, %s", err)
	}

	if err := chatquestion.InitHistoryESClient(ctx); err != nil {
		logs.FatalContextf(ctx, "[main] InitHistoryESClient failed, %s", err)
	}

	logs.InfoContextf(ctx, "admin unit testing environment initialized successfully")
	return nil
}

// Close 实现应用特定的关闭逻辑
func (k *adminInitializer) Close() error {
	// 在这里实现必要的关闭逻辑
	return k.baseInitializer.Close()
}
