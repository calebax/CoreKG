package migrator

import (
	"context"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kesearch/models/chunk"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
)

type FileNameMigrator struct {
}

func (m *FileNameMigrator) Migrator(ctx context.Context) error {
	if err := chunk.InitESClient(ctx); err != nil {
		logs.FatalContextf(ctx, "[main] InitESClient failed, %s", err)
	}
	logs.InfoContextf(ctx, "start Migrator FileNameMigrator")
	var files []*foresttype.KnownowForestFile
	err := dbutil.Knownow().Model(&foresttype.KnownowForestFile{}).Find(&files).Error
	if err != nil {
		logs.ErrorContextf(ctx, "MigrateChunkFileName: %v", err)
		return err
	}

	for _, file := range files {
		logs.InfoContextf(ctx, "MigrateChunkFileName: %+v", file)
		err := chunk.UpdateChunkFileName(ctx, file.ID, file.Name)
		if err != nil {
			return err
		}
	}
	logs.InfoContextf(ctx, "success Migrator FileNameMigrator")
	return nil
}
