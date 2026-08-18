package svcdbforest

import (
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/pkgs/utils/ginctx"
	"github.com/ygpkg/yg-go/logs"
)

func SyncMysqlTable() (string, error) {
	ctx := ginctx.InitJobCtx("SyncMysqlTable", "")
	logs.InfoContext(ctx, "[SyncMysqlTable] start sync mysql table")
	// 查询所有数据库
	dbEntityList, err := forest.NewForestDBDao().GetListByCond(ctx, &forest.ForestDBCond{})
	if err != nil {
		logs.ErrorContextf(ctx, "[SyncMysqlTable] get db list fail, err: %v", err)
		return "", err
	}
	for _, dbEntity := range dbEntityList {
		if err := syncMysqlDatabase(ctx, &dbEntity); err != nil {
			logs.WarnContextf(ctx, "[SyncMysqlTable] sync db fail, db id: %d, err: %v", dbEntity.ID, err)
			continue
		}
	}
	return "", nil
}
