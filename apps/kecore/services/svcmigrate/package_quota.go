package svcmigrate

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
)

const (
	BusinessTypeMigratePackageQuota BusinessType = "migrate_package_quota"
)

type MigratePackageQuotaMigrator struct {
}

func (m *MigratePackageQuotaMigrator) Run(ctx context.Context) error {
	logs.InfoContextf(ctx, "migrate package quota migrator run")

	var (
		cmpIDs        []uint
		cmpPkgIDs     []uint
		pendingCmpIDs []uint
		cmpQuota      []*foresttype.KeCompanyQuota
		freePackage   = &foresttype.KePackage{}
		timeNow       = time.Now()
		timeForever   = time.Date(2099, 1, 1, 0, 0, 0, 0, time.Now().Location())
	)
	// Get company ids that do not have any package quota
	if err := dbutil.Account().WithContext(ctx).
		Table(accounttype.TableNameCompany).
		Where("deleted_at IS NULL").
		Pluck("id", &cmpIDs).
		Error; err != nil {
		logs.ErrorContextf(ctx, "migrate package quota migrator get company ids failed, err: %v", err)
		return err
	}

	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKeCompanyQuota).
		Where("deleted_at IS NULL").
		Select("company_id").
		Distinct().
		Pluck("company_id", &cmpPkgIDs).
		Error; err != nil {
		logs.ErrorContextf(ctx, "migrate package quota migrator get company package quota ids failed, err: %v", err)
		return err
	}
	cmpPkgIDMap := utils.ToMap(cmpPkgIDs, func(cmpPkgID uint) uint {
		return cmpPkgID
	})

	for _, cmpID := range cmpIDs {
		if _, ok := cmpPkgIDMap[cmpID]; !ok {
			pendingCmpIDs = append(pendingCmpIDs, cmpID)
		}
	}

	// Get free package level
	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKePackage).
		Where("deleted_at IS NULL").
		Where("source_type = ?", foresttype.PackageSourceTypeSystem).
		Where("status = ?", foresttype.PackageStatusOnline).
		Where("period_type = ?", foresttype.PackagePeriodTypeLifetime).
		Where("level = ?", foresttype.PackageLevel1).
		First(&freePackage).
		Error; err != nil {
		logs.ErrorContextf(ctx, "migrate package quota migrator get free package failed, err: %v", err)
		return err
	}

	// Do a iterate to create each company's package quota
	for _, v := range pendingCmpIDs {
		cmpQuota = append(cmpQuota, &foresttype.KeCompanyQuota{
			CompanyID:     v,
			SourceType:    foresttype.CompanyQuotaSourceTypeManual,
			PackageLevel:  freePackage.Level,
			AgentQuota:    freePackage.AgentQuota,
			QaQuota:       freePackage.QaQuota,
			DiskQuota:     freePackage.DiskQuota,
			EmployeeQuota: freePackage.EmployeeQuota,
			EffectiveAt:   &timeNow,
			ExpireAt:      &timeForever,
		})
	}

	// Create in batch to create company's package quota

	if err := dbutil.Knownow().WithContext(ctx).CreateInBatches(cmpQuota, 50).Error; err != nil {
		logs.ErrorContextf(ctx, "migrate package quota migrator create company package quota failed, err: %v", err)
		return fmt.Errorf("migrate package quota migrator create company package quota failed, err: %v", err)
	}

	logs.InfoContextf(ctx, "migrate package quota migrator end")
	return nil
}
