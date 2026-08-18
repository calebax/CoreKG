package migrator

import (
	"context"

	"github.com/ygpkg/yg-go/logs"
)

type exampleMigrator struct {
}

func (m *exampleMigrator) Migrator(ctx context.Context) error {
	logs.InfoContextf(ctx, "example migrator start")
	logs.InfoContextf(ctx, "example migrator end")
	return nil
}
