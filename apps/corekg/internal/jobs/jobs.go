package jobs

import (
	"context"

	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/version"
	"github.com/ygpkg/yg-go/logs"
)

// RunRoutines starts the necessary routines for the knowledge job processing.
func RunRoutines(ctx context.Context) error {
	{
		if version.DeployMode() != global.DeployModeTencentFree {
			licenseCtx := logs.WithContextFields(ctx, "routine", "license_valid")
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logs.ErrorContextf(licenseCtx, "license_valid recovered from panic: %v", r)
					}
				}()
				LicenseCheckRoutine(licenseCtx)
			}()
		}
	}

	return nil
}
