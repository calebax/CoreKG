package jobs

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/insmtx/corekg/apps/admin/models/admintype"
	license2 "github.com/insmtx/corekg/apps/admin/models/license"
	"github.com/insmtx/corekg/apps/corekg/models/license"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// LicenseCheckRoutine is a routine for license check
func LicenseCheckRoutine(ctx context.Context) {
	var err error
	//do start
	logs.InfoContext(ctx, "LicenseCheckRoutine started")
	if err = CheckLicense(ctx); err != nil {
		logs.ErrorContext(ctx, "LicenseCheckRoutine init check failed: %v", err.Error())
		return
	}

	now := time.Now()

	// 设置今天中午12点的时间
	nextCheckTime := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())

	// 如果今天12点已经过了，将执行时间推迟到明天
	if now.After(nextCheckTime) {
		nextCheckTime = nextCheckTime.Add(24 * time.Hour)
	}

	// 计算首次执行的等待时长
	initialDelay := nextCheckTime.Sub(now)

	// 使用time.After来等待首次执行
	select {
	case <-ctx.Done():
		logs.InfoContext(ctx, "LicenseCheckRoutine stopped gracefully during initial delay.")
		return
	case <-time.After(initialDelay):
		// 首次执行
		if err = CheckLicense(ctx); err != nil {
			logs.ErrorContextf(ctx, "Failed to process license check: %v", err)
			return
		}
	}

	// 首次执行后，创建一个24小时的定时器，确保后续在每天的同一时间点触发
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logs.InfoContext(ctx, "LicenseCheckRoutine stopped gracefully")
			return
		case <-ticker.C:
			if err = CheckLicense(ctx); err != nil {
				logs.ErrorContextf(ctx, "license check err: %v", err)
				return
			}
		}
	}
}

const DeploymentEnv = "DEPLOYMENT_ENV"

func RedisKeyLicenseCheck() string {
	return "Corekg:LicenseCheckRedisKey"
}

var (
	MaxRetry         = 5
	RetryDelay       = 2 * time.Second
	ExpireKeyTimeOut = 3 * time.Minute
)

func CheckLicense(ctx context.Context) (err error) {
	logs.InfoContext(ctx, "Executing license check...")
	db := dbutil.Core()

	// 1. 从环境变量或配置中获取当前环境类型
	// e.g., "kubernetes"
	envTypeStr := os.Getenv(DeploymentEnv)

	// 2. 使用工厂创建对应的 Environment 实例
	env, err := license.NewEnvironment(license.EnvType(envTypeStr))
	if err != nil {
		logs.ErrorContextf(ctx, "Failed to create environment: %v", err)
		return
	}

	// 3. 创建校验器 Checker
	checker := license2.NewChecker(db, env)
	key := RedisKeyLicenseCheck()

	logs.InfoContext(ctx, "Acquire license distribution mutex ...")
	for i := range MaxRetry {
		nx, err := redispool.SetNX(key, "", ExpireKeyTimeOut)
		// 4. 执行检查（包含日志记录）
		if err != nil {
			logs.ErrorContextf(ctx, "CheckLicense: redispool.SetNX err: %v", err)
			return err
		}
		if nx {
			break
		}
		logs.ErrorContextf(ctx, "check is already running, this check will retry after %v", RetryDelay)
		time.Sleep(RetryDelay)
		if i == MaxRetry-1 {
			return fmt.Errorf("retry after %v times and faild, skip this check", MaxRetry)
		}
	}

	checker.PerformCheck(ctx)

	if err := redispool.Del(key); err != nil {
		logs.ErrorContextf(ctx, "CheckLicense: faild to del rediskey(%v) err: %v", key, err)
		return err
	}

	return
}

// Exit will send a terminated signal to process
func Exit(ctx context.Context) {
	//get current pid
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		logs.WarnContextf(ctx, "could not get current pid：%v", err)
	}

	//send signal to main proc
	if err = proc.Signal(syscall.SIGTERM); err != nil {
		logs.WarnContextf(ctx, "send signal faild：%v", err)
	}
}

// logInitialError 在无法创建环境等早期阶段记录错误
func logInitialError(ctx context.Context, db *gorm.DB, msg string) {
	logEntry := admintype.DailyLog{
		PreviousHash: "N/A",
		CurrentHash:  "N/A",
		Valid:        -1,
		Message:      msg,
	}
	if err := db.WithContext(ctx).Create(&logEntry).Error; err != nil {
		logs.ErrorContextf(ctx, "logInitialError: Failed to log initial error: %v", err)
	}
}
