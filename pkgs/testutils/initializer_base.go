package testutils

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"runtime"

	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/cache"
	"github.com/ygpkg/yg-go/cache/redis"
	"github.com/ygpkg/yg-go/config"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
)

// baseInitializer 提供基础的初始化器实现
type baseInitializer struct {
	cfg *config.CoreConfig
}

// newBaseInitializer 创建基础初始化器
func newBaseInitializer(appName AppName) (*baseInitializer, error) {
	if appName == "" {
		return nil, fmt.Errorf("app name is empty")
	}
	cfg, loadErr := loadConfig(string(appName))
	if loadErr != nil {
		return nil, fmt.Errorf("load config: %w", loadErr)
	}
	return &baseInitializer{
		cfg: cfg,
	}, nil
}

// Initialize 实现基础的初始化逻辑
func (b *baseInitializer) initialize() error {
	// 在这里实现基础的初始化逻辑

	ctx := context.TODO()

	// 初始化数据库连接
	if err := dbtools.InitMultiDBConn(b.cfg.MainConf.DatabaseConns); err != nil {
		return fmt.Errorf("connect mysql failed: %w", err)
	}

	// 初始化 Redis
	if err := redispool.InitRedis("knowledge", "loc_redis"); err != nil {
		logs.WarnContextf(ctx, "connect redis failed, %s", err)
	} else {
		cache.InitCache(redis.NewCache(redispool.Redis()))
	}
	dbutil.InitializePlugins()
	return nil
}

// Close 实现基础的清理逻辑
func (*baseInitializer) Close() error {
	// 在这里实现基础的清理逻辑
	return nil
}

// loadConfig 从指定路径加载配置文件
func loadConfig(appName string) (*config.CoreConfig, error) {
	ctx := context.TODO()
	_, callFile, _, _ := runtime.Caller(0)
	rootDir := path.Dir(path.Dir(path.Dir(callFile)))
	configFile := filepath.Join(rootDir, "apps", appName, "conf/test/config.yaml")

	cfg, loadConfigErr := config.LoadCoreConfig(configFile)
	if loadConfigErr != nil {
		return nil, fmt.Errorf("load config: %w", loadConfigErr)
	}
	if err := logs.ReloadConfig(cfg.MainConf.App, cfg.LogsConf); err != nil {
		logs.ErrorContextf(ctx, "reload log config failed, %s", err)
		return nil, err
	}
	return cfg, nil
}
