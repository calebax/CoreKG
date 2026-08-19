// Package workflow 提供可复用启动入口，供 workflow 独立运行及 corekg 聚合进程
// 拉启（同进程双 server 模式）共用。
//
// workflow 底层使用 Hertz 框架，路由无法注册到 yg-go 的 *server.Router，
// 因此本包不提供其它聚合应用那样的 Routers(*server.Router)。corekg 通过
// startup.Start(ctx, cfg) 显式拉启 workflow 服务，而不是走 Routers 契约。
package workflow

import "gorm.io/gorm"

// Migrates 无独立 schema 迁移；workflow 的表结构在启动初始化中处理。
func Migrates(db *gorm.DB) error {
	return nil
}

// RunJob 无独立定时任务入口；workflow 的异步任务由自身初始化托管。
func RunJob() error {
	return nil
}
