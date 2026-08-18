package migrator

import (
	"context"
	"fmt"

	"github.com/ygpkg/yg-go/logs"
)

type MigrateType string

const (
	ExampleMigrate       MigrateType = "example_migrate"
	DepartmentOrgMigrate MigrateType = "department_org_migrate"
	FileNameOrgMigrate   MigrateType = "filename_org_migrate"
	GraphTagNodeMigrate  MigrateType = "graph_tag_node_migrate"
	CozeModelSyncMigrate MigrateType = "coze_model_sync_migrate"
)

type Migrator interface {
	Migrator(ctx context.Context) error
}

var migratorMap = make(map[MigrateType]Migrator)

func Register(migratorType MigrateType, m Migrator) {
	migratorMap[migratorType] = m
}

func Run(ctx context.Context) error {
	var errs []error
	for migratorType, m := range migratorMap {
		if err := m.Migrator(ctx); err != nil {
			logs.ErrorContextf(ctx, "migrator failed, migratorType: %s, err: %v", migratorType, err)
			errs = append(errs, err)
		}
	}
	var err error
	for _, e := range errs {
		err = fmt.Errorf("%s, %s", err, e)
	}
	return err
}
