package main

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/ygpkg/yg-go/config"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"

	"github.com/insmtx/corekg/apps/keapp/models/apptype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
)

type natsConfig struct {
	URL    string `yaml:"url"`
	Stream string `yaml:"stream"`
}

type rawConfig struct {
	NATS   natsConfig        `yaml:"nats"`
	Worker map[string]any    `yaml:"worker"`
}

func initDatabase(ctx context.Context, cfg *config.CoreConfig, migrateDB bool) (*gorm.DB, error) {
	err := dbtools.InitMultiDBConn(cfg.MainConf.DatabaseConns)
	if err != nil {
		logs.ErrorContextf(ctx, "[main] connect mysql failed, %s", err)
		return nil, err
	}
	db := dbutil.Core()
	db.Logger = logs.GetGorm("gorm")

	if !migrateDB {
		return db, nil
	}

	err = dbtools.DoInitModels(
		apptype.InitDB,
	)
	if err != nil {
		logs.ErrorContextf(ctx, "[main] init database failed, %s", err)
		return nil, err
	}

	return db, nil
}

func initNATS(ctx context.Context, configPath string) (*nats.Conn, error) {
	var raw rawConfig
	if err := config.LoadYamlLocalFile(configPath, &raw); err != nil {
		logs.WarnContextf(ctx, "[main] load nats config failed: %v", err)
		return nats.Connect(nats.DefaultURL)
	}
	if raw.NATS.URL == "" {
		raw.NATS.URL = nats.DefaultURL
	}
	nc, err := nats.Connect(raw.NATS.URL)
	if err != nil {
		return nil, fmt.Errorf("connect NATS at %s: %w", raw.NATS.URL, err)
	}
	logs.InfoContextf(ctx, "[main] connected to NATS at %s", raw.NATS.URL)
	return nc, nil
}
