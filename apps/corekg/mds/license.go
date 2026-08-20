package mds

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/pkgs/platform/admintype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

func LicenseCheck(whitelist ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestPath := ctx.Request.URL.Path

		// 判断是否匹配白名单路径
		for _, skip := range whitelist {
			if strings.Contains(requestPath, skip) {
				logs.InfoContextf(ctx, "[LicenseCheck] 跳过 license 校验, path: %s 命中白名单: %s", requestPath, skip)
				ctx.Next()
				return
			}
		}

		if err := CheckValidLogEntry(ctx); err != nil {
			logs.ErrorContextf(ctx, "license check failed: %v", err)
			ctx.AbortWithStatusJSON(http.StatusForbidden, apiobj.BaseResponse{Code: 403, Message: "License认证失败"})
			return
		}
		logs.DebugContextf(ctx, "license check succeed")
	}
}

// CheckValidLogEntry Here's the refactored function
func CheckValidLogEntry(ctx context.Context) error {
	lg := &admintype.DailyLog{}

	// Calculate the time 72 hours ago
	since := time.Now().Add(-72 * time.Hour)

	// Find any valid log entry from the last 72 hours.
	// We don't need to get the latest, just check for existence.
	// The query now combines a time filter and the 'Valid' status.
	if err := dbtools.Core().WithContext(ctx).
		Where("date >= ? AND valid = ?", since, 1).
		First(&lg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logs.InfoContextf(ctx, "No valid log entry found in the last 72 hours.")
			return gorm.ErrRecordNotFound // Return an explicit error
		} else {
			logs.ErrorContextf(ctx, "Failed to get valid log entry: %v", err)
			return err
		}
	}

	// If we reached this point, a valid record was found.
	logs.InfoContextf(ctx, "A valid log entry was found in the last 72 hours: %+v", *lg)
	return nil
}
