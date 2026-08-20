package license

import (
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	corekglicense "github.com/insmtx/corekg/apps/corekg/models/license"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/platform/admintype"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
)

const (
	MillSecondsTemplate = "2006-01-02 15:04:05.000"
)

// Checker 负责执行完整的校验和日志记录流程
type Checker struct {
	db      *gorm.DB
	env     corekglicense.Environment
	hmacKey [32]byte
	preHash string
}

func NewChecker(db *gorm.DB, env corekglicense.Environment) *Checker {
	return &Checker{db: db, env: env}
}

// PerformCheck 公开方法，协调整个校验和日志记录
func (c *Checker) PerformCheck(ctx context.Context) {
	var status corekglicense.ValidationStatus
	var message string

	// 运行核心检查逻辑，它会返回最终状态和消息
	status, message = c.runCoreChecks(ctx)
	if status != corekglicense.StatusValid {
		logs.ErrorContextf(ctx, "runCoreChecks failed, status: %d,message: %v", status, message)
		return
	}

	// 仅保留成功的记录到数据库
	if err := c._logResult(ctx, status, message); err != nil {
		logs.ErrorContextf(ctx, "CRITICAL: Failed to log license check result: %v", err)
	}
}

// PureCheck 公开方法,仅校验license是否合法,不对license的日志信息进行校验
func (c *Checker) PureCheck(ctx context.Context) (corekglicense.ValidationStatus, string) {
	// 1. 从环境中获取所有必要信息
	uid, err := c.env.GetUID(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "CRITICAL: Failed to get UID: %v", err)
		return corekglicense.StatusEnvError, fmt.Sprintf("Failed to get environment UID: %v", err)
	}
	rawLicense, err := c.env.GetRawLicense(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "CRITICAL: Failed to get raw license data: %v", err)
		return corekglicense.StatusEnvError, fmt.Sprintf("Failed to get raw license: %v", err)
	}

	// 2. 解析元数据
	meta, signature, jsonData, err := c._parse(ctx, rawLicense)
	if err != nil {
		return corekglicense.StatusInvalidSignature, err.Error()
	}

	if err = c._verifySignature(ctx, signature, jsonData, meta.Issuer); err != nil {
		logs.ErrorContextf(ctx, "verifySignature invalid signature: %v", err)
		return corekglicense.StatusInternalError, err.Error()
	}

	logs.DebugContextf(ctx, "verifySignature valid signature passed")
	// 3. 校验签名
	if err := c._verifyMetadata(ctx, meta, uid); err != nil {
		status := corekglicense.StatusInternalError
		if errors.Is(err, corekglicense.ErrLicenseUIDNotMatch) {
			status = corekglicense.StatusUIDMismatch
		} else if errors.Is(err, corekglicense.ErrLicenseExpired) {
			status = corekglicense.StatusExpired
		}
		return status, err.Error()
	}
	return corekglicense.StatusValid, ""
}

// Meta 公开方法,返回当前license元信息与状态
func (c *Checker) Meta(ctx context.Context) (*admintype.Meta, corekglicense.ValidationStatus) {
	var meta *admintype.Meta
	// 1. 从环境中获取所有必要信息
	uid, err := c.env.GetUID(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "CRITICAL: Failed to get UID: %v", err)
		return nil, corekglicense.StatusEnvError
	}

	rawLicense, err := c.env.GetRawLicense(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "CRITICAL: Failed to get raw license data: %v", err)
		return nil, corekglicense.StatusEnvError
	}

	// 2. 解析元数据
	meta, signature, jsonData, err := c._parse(ctx, rawLicense)
	if err != nil {
		return meta, corekglicense.StatusInvalidSignature
	}
	logs.DebugContextf(ctx, "meta: %v", meta)

	if err = c._verifySignature(ctx, signature, jsonData, meta.Issuer); err != nil {
		logs.ErrorContextf(ctx, "verifySignature invalid signature: %v", err)
		return nil, corekglicense.StatusInternalError
	}

	logs.DebugContextf(ctx, "verifySignature valid signature passed")
	// 3. 校验签名
	if err := c._verifyMetadata(ctx, meta, uid); err != nil {
		status := corekglicense.StatusInternalError
		if errors.Is(err, corekglicense.ErrLicenseUIDNotMatch) {
			status = corekglicense.StatusUIDMismatch
		} else if errors.Is(err, corekglicense.ErrLicenseExpired) {
			status = corekglicense.StatusExpired
		}
		return meta, status
	}
	return meta, corekglicense.StatusValid
}

func (c *Checker) runCoreChecks(ctx context.Context) (corekglicense.ValidationStatus, string) {
	// 1. 从环境中获取所有必要信息
	uid, err := c.env.GetUID(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "CRITICAL: Failed to get UID: %v", err)
		return corekglicense.StatusEnvError, fmt.Sprintf("Failed to get environment UID: %v", err)
	}
	rawLicense, err := c.env.GetRawLicense(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "CRITICAL: Failed to get raw license data: %v", err)
		return corekglicense.StatusEnvError, fmt.Sprintf("Failed to get raw license: %v", err)
	}

	// 2. 解析元数据
	meta, signature, jsonData, err := c._parse(ctx, rawLicense)
	if err != nil {
		return corekglicense.StatusInvalidSignature, err.Error()
	}

	if err = c._verifySignature(ctx, signature, jsonData, meta.Issuer); err != nil {
		logs.ErrorContextf(ctx, "verifySignature invalid signature: %v", err)
		return corekglicense.StatusInternalError, err.Error()
	}

	logs.DebugContextf(ctx, "verifySignature valid signature passed")
	// 3. 校验签名
	if err := c._verifyMetadata(ctx, meta, uid); err != nil {
		status := corekglicense.StatusInternalError
		if errors.Is(err, corekglicense.ErrLicenseUIDNotMatch) {
			status = corekglicense.StatusUIDMismatch
		} else if errors.Is(err, corekglicense.ErrLicenseExpired) {
			status = corekglicense.StatusExpired
		}
		return status, err.Error()
	}

	// 4. 校验哈希链
	c.hmacKey = sha256.Sum256([]byte(meta.UID + meta.Seed))
	if err := c._verifyHashChain(ctx); err != nil {
		return corekglicense.StatusTampered, err.Error()
	}

	// 所有检查通过
	return corekglicense.StatusValid, corekglicense.StatusValid.String()
}

func (c *Checker) _verifySignature(ctx context.Context, signature, jsonData []byte, issuer string) error {
	pub := global.IssuerPublicKeyMap[issuer]
	logs.DebugContextf(ctx, "Get issuer[%v] publickey[%v]", issuer, pub)
	publickey, err := corekglicense.ParsePublicKey(pub)
	if err != nil {
		logs.ErrorContextf(ctx, "CRITICAL: Failed to parse public key: %v", err)
		return err
	}
	logs.DebugContextf(ctx, "Get publickey[%v]", publickey)

	logs.DebugContextf(ctx, "Get jsondata[%v]", string(jsonData))
	hashed := sha256.Sum256(jsonData)

	if err := rsa.VerifyPKCS1v15(publickey, crypto.SHA256, hashed[:], signature); err != nil {
		logs.ErrorContextf(ctx, "CRITICAL: Failed to verify signature: %v", err)
		return err
	}
	return nil
}
func (c *Checker) _parse(ctx context.Context, rawLicense string) (*admintype.Meta, []byte, []byte, error) {
	parts := strings.Split(rawLicense, ".")
	if len(parts) != 2 {
		logs.ErrorContextf(ctx, "Invalid license format: %s", rawLicense)
		return nil, nil, nil, fmt.Errorf("invalid license format with %d parts", len(parts))
	}
	signature, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		logs.ErrorContextf(ctx, "invalid base64 for signature: %v", err)
		return nil, nil, nil, fmt.Errorf("invalid base64 for signature: %w", err)
	}
	jsonData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		logs.ErrorContextf(ctx, "invalid base64 for data: %v", err)
		return nil, nil, nil, fmt.Errorf("invalid base64 for data: %w", err)
	}

	var meta admintype.Meta
	if err := json.Unmarshal(jsonData, &meta); err != nil {
		logs.ErrorContextf(ctx, "failed to unmarshal license metadata: %v", err)
		return nil, nil, nil, fmt.Errorf("failed to unmarshal license meta: %w", err)
	}
	return &meta, signature, jsonData, nil
}

func (c *Checker) _verifyMetadata(ctx context.Context, meta *admintype.Meta, currentUID string) error {
	if meta.UID != currentUID {
		logs.ErrorContextf(ctx, "metadata UID does not match,license_UID[%v],currentUID[%v]", meta.UID, currentUID)
		return corekglicense.ErrLicenseUIDNotMatch
	}
	if time.Now().After(meta.ExpiredAt) {
		logs.ErrorContextf(ctx, "metadata expired at[%v] now[%v]", meta.ExpiredAt, time.Now())
		return corekglicense.ErrLicenseExpired
	}
	if time.Now().Before(meta.CreatedAt) {
		logs.ErrorContextf(ctx, "metadata created at[%v] now[%v]", meta.CreatedAt, time.Now())
		return corekglicense.ErrLicenseExpired
	}
	return nil
}

func (c *Checker) _verifyHashChain(ctx context.Context) error {
	var dailyLogs []admintype.DailyLog
	if err := c.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("date asc").
		Find(&dailyLogs).
		Error; err != nil {
		logs.ErrorContextf(ctx, "failed to find daily logs: %v", err)
		return fmt.Errorf("failed to query daily logs: %w", err)
	}

	if len(dailyLogs) == 0 {
		logs.InfoContextf(ctx, "daily logs is empty,valid hashchain passed")
		mac := hmac.New(sha256.New, c.hmacKey[:])
		mac.Write(c.hmacKey[:])
		c.preHash = hex.EncodeToString(mac.Sum(nil))
		return nil
	}
	logs.DebugContextf(ctx, "found latest log id[%v]", dailyLogs[len(dailyLogs)-1].ID)

	// 检查时间是否回拨
	lastLog := dailyLogs[len(dailyLogs)-1]
	lastT, err := time.ParseInLocation(MillSecondsTemplate, lastLog.Date, time.Local)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to parse last log date: %v", err)
		return fmt.Errorf("failed to parse last log date: %w", err)
	}

	now := time.Now()
	if now.Before(lastT) {
		logs.ErrorContextf(ctx, "system time appears to have been moved backwards, last log date: %v", lastLog.Date)
		return fmt.Errorf("system time appears to have been moved backwards, last log date: %v", lastLog.Date)
	}

	// 校验哈希链
	c.preHash = hex.EncodeToString(c.hmacKey[:])
	for i, logEntry := range dailyLogs {
		if i == 0 {
			mac := hmac.New(sha256.New, c.hmacKey[:])
			mac.Write(c.hmacKey[:])
			c.preHash = hex.EncodeToString(mac.Sum(nil))
		}
		mac := hmac.New(sha256.New, c.hmacKey[:])
		mac.Write([]byte(logEntry.Date + c.preHash))
		expectedHash := hex.EncodeToString(mac.Sum(nil))

		if expectedHash != logEntry.CurrentHash {
			logs.ErrorContextf(ctx, "try use this to debug expectedhash comes from [%v],preHash[%v],hmackey[%v]", logEntry.Date, c.preHash, c.hmacKey[:])
			logs.ErrorContextf(ctx, "hash chain integrity compromised at date: %s, expected:[%v],current[id:%v][%v]", logEntry.Date, expectedHash, logEntry.ID, logEntry.CurrentHash)
			return corekglicense.ErrLicenseTampered
		}
		c.preHash = logEntry.CurrentHash
	}

	return nil
}

func (c *Checker) _logResult(ctx context.Context, status corekglicense.ValidationStatus, message string) error {
	mac := hmac.New(sha256.New, c.hmacKey[:])
	now := time.Now().Format(MillSecondsTemplate)
	mac.Write([]byte(now + c.preHash))
	currHash := hex.EncodeToString(mac.Sum(nil))

	var validFlag types.Bool = 1
	if status != corekglicense.StatusValid {
		validFlag = -1
	}

	logEntry := admintype.DailyLog{
		Date: now, PreviousHash: c.preHash, CurrentHash: currHash, Valid: validFlag, Message: message,
	}

	return c.db.WithContext(ctx).Create(&logEntry).Error
}
