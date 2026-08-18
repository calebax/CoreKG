package kesale

import (
	"context"

	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/job"
	"github.com/ygpkg/yg-go/logs"
)

func runCornJob() {
	defer func() {
		if err := recover(); err != nil {
			logs.ErrorContextf(context.Background(), "[KeSaleCornJob] panic: %v", err)
		}
	}()

	ctx := context.Background()
	spec := "0 * * * * *"
	// Test
	// spec = "*/10 * * * * *"

	logs.InfoContextf(ctx, "[KeSaleCornJob] register verify pending orders cron: %s", spec)

	job.RegistryCronFunc(dbutil.Core(), spec, "VerifyPendingOrders", func() (string, error) {
		Manager().VerifyOrderStatus(context.Background())
		return "", nil
	})

	logs.InfoContextf(ctx, "[KeSaleCornJob] register complete")
}
