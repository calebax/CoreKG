package services

import (
	"context"

	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type CtxDBKey struct{}

func CtxWithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, CtxDBKey{}, db)
}

func ManagerDBFromCtx(ctx context.Context) *gorm.DB {
	v := ctx.Value(CtxDBKey{})
	db, ok := v.(*gorm.DB)
	if ok && db != nil {
		n := *db
		return &n
	}
	logs.WarnContextf(ctx, "kesaleManager: db not found in ctx")
	return dbutil.Knownow()
}
