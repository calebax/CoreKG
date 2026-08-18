package mysql

import (
	"context"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/insmtx/corekg/apps/keinit/models/helper"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

var sqlVersionRe = regexp.MustCompile(`^v(\d+)\.(\d+)_(\d+)__(.+)\.sql$`)

type sqlMigration struct {
	path  string
	name  string
	major int
	minor int
	seq   int
	desc  string
	valid bool
}

// ExecuteScripts 执行脚本
func ExecuteScripts(ctx context.Context, db *gorm.DB, scriptDir string) error {
	succExists, err := ListAllSuccessExecRecourdMap(db)
	if err != nil {
		logs.ErrorContextf(ctx, "list all success exec record map error: %v", err)
		return err
	}

	sqlFiles, err := helper.ListFileWithExt(ctx, scriptDir, ".sql")
	if err != nil {
		logs.ErrorContextf(ctx, "list sql files error: %v", err)
		return err
	}

	migrations := make([]sqlMigration, 0, len(sqlFiles))
	for _, f := range sqlFiles {
		base := filepath.Base(f)
		m := sqlMigration{path: f, name: base}
		matches := sqlVersionRe.FindStringSubmatch(base)

		if len(matches) == 5 {
			m.major, _ = strconv.Atoi(matches[1])
			m.minor, _ = strconv.Atoi(matches[2])
			m.seq, _ = strconv.Atoi(matches[3])
			m.desc = matches[4]
			m.valid = true
		}
		migrations = append(migrations, m)
	}

	sort.Slice(migrations, func(i, j int) bool {
		mi, mj := migrations[i], migrations[j]

		if !mi.valid || !mj.valid {
			return mi.name < mj.name
		}

		if mi.major != mj.major {
			return mi.major < mj.major
		}
		if mi.minor != mj.minor {
			return mi.minor < mj.minor
		}
		if mi.seq != mj.seq {
			return mi.seq < mj.seq
		}
		return mi.desc < mj.desc
	})

	for _, m := range migrations {
		if _, ok := succExists[m.name]; ok {
			logs.InfoContextf(ctx, "skip execute sql file (%s), because it has been executed", m.path)
			continue
		}

		err := ExecuteSQLFile(ctx, db, m.path)
		if err != nil {
			logs.ErrorContextf(ctx, "execute sql file (%s) error: %v", m.path, err)
			return err
		}
		logs.InfoContextf(ctx, "execute sql file (%s) success", m.path)
	}
	return nil
}
