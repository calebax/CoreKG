package jobs

import (
	"context"

	"github.com/insmtx/corekg/apps/kecore/services/svcdbforest"
	"github.com/insmtx/corekg/apps/kecore/services/svchotwords"
	"github.com/insmtx/corekg/apps/kecore/services/svcmembership"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/job"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

func CoreCornJob(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			logs.ErrorContextf(context.Background(), "[KeCoreCornJob] panic: %v", err)
		}
	}()
	mysqlCornExpression := "0 0 0 * * *"
	notifyExpiringQuotasCornExpression := "0 0 9 * * *"
	if config.Conf().MainConf.Env == "test" || config.Conf().MainConf.Env == "dev" {
		testMysqlCormExpress, err := settings.GetValue("knowledge", "mysql_corn_expression")
		if err != nil {
			logs.ErrorContextf(context.Background(), "[KeCoreCornJob] get test mysql corn expression fail, err: %v", err)
		}
		if err == nil && testMysqlCormExpress != "" {
			mysqlCornExpression = testMysqlCormExpress
		}
		//
		testNotifyExpiringQuotasCornExpression, err := settings.GetValue("knowledge", "notify_expiring_quotas_corn_expression")
		if err != nil {
			logs.ErrorContextf(context.Background(), "[KeCoreCornJob] get test mysql corn expression fail, err: %v", err)
		}
		if err == nil && testNotifyExpiringQuotasCornExpression != "" {
			notifyExpiringQuotasCornExpression = testNotifyExpiringQuotasCornExpression
		}
	}
	logs.InfoContextf(context.Background(), "[KeCoreCornJob] start register")
	// 每天0点执行一次
	job.RegistryCronFunc(dbutil.Core(), mysqlCornExpression, "SyncMysqlTable", svcdbforest.SyncMysqlTable)
	job.RegistryCronFunc(dbutil.Core(), mysqlCornExpression, "GenerateUsersHotWords", svchotwords.GenerateUsersHotWords)
	job.RegistryCronFunc(dbutil.Core(), "0 0 * * * *", "Sync1", Sync1)
	job.RegistryCronFunc(dbutil.Core(), notifyExpiringQuotasCornExpression, "PackageQuotaExpireNotify", svcmembership.NotifyExpiringQuotas)
	logs.InfoContextf(context.Background(), "[KeCoreCornJob] register complete")
}

func Sync1() (string, error) {
	logs.InfoContextf(context.Background(), "[Sync1] running")
	return "", nil
}
