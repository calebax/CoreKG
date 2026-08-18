package migrator

import (
	"context"

	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/ygpkg/yg-go/logs"
)

type TagNodeMigrator struct {
}

func (m *TagNodeMigrator) Migrator(ctx context.Context) error {
	err := graph.MigrateTNodeToNode(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "Migrator TagNodeMigrator error: %v", err)
		return err
	}
	return nil
}
