package svcmigrate

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtomigrate"
	"github.com/ygpkg/yg-go/logs"
)

func NewMigrator(ctx context.Context, businessType BusinessType) (Migrator, error) {
	switch businessType {
	case BusinessTypeMigratePackageQuota:
		return &MigratePackageQuotaMigrator{}, nil
	default:
		logs.ErrorContextf(ctx, "unknown business type: %s", businessType)
		return nil, fmt.Errorf("unknown business type: %s", businessType)
	}
}

func MigrateInterface(ctx *gin.Context, req *dtomigrate.MigrateInterfaceRequest) (res *dtomigrate.MigrateInterfaceResponse, err error) {
	res = &dtomigrate.MigrateInterfaceResponse{}
	migrator, err := NewMigrator(ctx, BusinessType(req.Request.BusinessType))
	if err != nil {
		logs.ErrorContextf(ctx, "new migrator(%s) failed, err: %v", req.Request.BusinessType, err)
		return res, fmt.Errorf("new migrator(%s) failed, err: %v", req.Request.BusinessType, err)
	}
	if err = migrator.Run(ctx); err != nil {
		logs.ErrorContextf(ctx, "migrate interface(%s) run failed, err: %v", req.Request.BusinessType, err)
		return res, fmt.Errorf("migrate interface(%s) run failed, err: %v", req.Request.BusinessType, err)
	}
	return res, nil
}
