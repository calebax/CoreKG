package lkx

import (
	"context"

	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
)

// SaveInfo 保存信息
func SaveInfo(ctx context.Context, data *admintype.LkxCustomerInfo) error {
	err := dbutil.Account().WithContext(ctx).Table(admintype.TableNameLkxCustomerInfo).Save(data).Error
	if err != nil {
		logs.ErrorContextf(ctx, "save info failed, %s", err)
		return err
	}
	return nil
}
