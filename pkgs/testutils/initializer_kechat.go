package testutils

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/ygpkg/yg-go/logs"
)

// kechatInitializer 为 search 应用提供特定的初始化逻辑
type kechatInitializer struct {
	*baseInitializer
}

// newKecoreInitializer 创建 search 应用的初始化器
func newKechatInitializer() (Initializer, error) {
	base, err := newBaseInitializer(AppNameKechat)
	if err != nil {
		return nil, fmt.Errorf("create base initializer: %w", err)
	}

	return &kechatInitializer{
		baseInitializer: base,
	}, nil
}

// Initialize 实现 search 应用特定的初始化逻辑
func (k *kechatInitializer) Initialize() error {
	// 首先执行基础初始化
	ctx := context.TODO()
	if err := k.baseInitializer.initialize(); err != nil {
		return err
	}

	if err := essearch.InitEbConfig(ctx); err != nil {
		logs.FatalContextf(ctx, "[main] InitEbConfig failed, %s", err)
	}
	if err := chatquestion.InitHistoryESClient(ctx); err != nil {
		logs.FatalContextf(ctx, "[main] InitEbConfig failed, %s", err)
	}

	logs.InfoContextf(ctx, "kechat unit testing environment initialized successfully")
	return nil
}

// Close 实现 kechat 应用特定的关闭逻辑
func (k *kechatInitializer) Close() error {
	// 在这里实现必要的关闭逻辑
	return k.baseInitializer.Close()
}
