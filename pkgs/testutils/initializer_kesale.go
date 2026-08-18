package testutils

import (
	"fmt"
)

// kechatInitializer 为 search 应用提供特定的初始化逻辑
type kesaleInitializer struct {
	*baseInitializer
}

// newKecoreInitializer 创建 search 应用的初始化器
func newkesaleInitializer() (Initializer, error) {
	base, err := newBaseInitializer(AppNameKesale)
	if err != nil {
		return nil, fmt.Errorf("create base initializer: %w", err)
	}

	return &kechatInitializer{
		baseInitializer: base,
	}, nil
}

// Initialize 实现 search 应用特定的初始化逻辑
func (k *kesaleInitializer) Initialize() error {
	// 首先执行基础初始化
	if err := k.baseInitializer.initialize(); err != nil {
		return err
	}
	return nil
}

// Close 实现 kechat 应用特定的关闭逻辑
func (k *kesaleInitializer) Close() error {
	// 在这里实现必要的关闭逻辑
	return k.baseInitializer.Close()
}
