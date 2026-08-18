package employee

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/pkgs/utils/notify/email"
	"github.com/ygpkg/yg-go/cache"
	"github.com/ygpkg/yg-go/logs"
	"golang.org/x/crypto/bcrypt"
)

// LoginPasswordRequest .
type LoginPasswordRequest struct {
	// LoginType 登陆方式
	LoginType string `json:"type"`
	Name      string `json:"name"`
	// Password hmac-hmd5 password
	Password string `json:"password"`
}

// SendResetPasswordCodeByEmail 发生邮箱重置密码验证码
func SendResetPasswordCodeByEmail(ctx context.Context, smtpcli *email.EmailAccount, emailAddr string) error {
	var (
		key  = fmt.Sprintf("%s:%s", cacheKeyVerifyCodeRPPrefix, emailAddr)
		code = generateVerifyCode()
	)
	if cache.Std().IsExist(key) {
		logs.WarnContextf(ctx, "[admin] exists key: %s", key)
		return ErrSendVerifyCodeTooBusy
	}

	cache.Std().Set(key, code, time.Second*time.Duration(verifyEffectiveMinutes))

	data := map[string]interface{}{
		"Minutes": verifyEffectiveMinutes,
		"Code":    code,
		"From":    "ROC",
	}
	body, err := generateEmailBodyTemplate(data)
	if err != nil {
		logs.ErrorContextf(ctx, "[admin] generate code body failed, %s", err)
		return err
	}

	if err := smtpcli.SendHTML("密码重置验证码", body, emailAddr); err != nil {
		logs.ErrorContextf(ctx, "[admin] send code failed, %s", err)
		return err
	}

	return nil
}

// ResetPassword 重置密码
func ResetPassword(ctx context.Context, code, email, password string) error {
	var (
		key       = fmt.Sprintf("%s:%s", cacheKeyVerifyCodeRPPrefix, email)
		cacheCode string
	)
	cache.Std().Get(key, &cacheCode)
	if cacheCode != code {
		logs.WarnContextf(ctx, "[admin] not equal, exp: %s, req: %s", cacheCode, code)
		return fmt.Errorf("wrong verify code")
	}

	encPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logs.ErrorContextf(ctx, "[admin] generate password failed: password(%s), %s", password, err)
		return err
	}

	err = dbutil.Account().Table(admintype.TableNameEmployee).
		Where("email = ?", email).
		Update("password", string(encPassword)).
		Error

	if err != nil {
		logs.WarnContextf(ctx, "[admin] update password failed: password(%s), %s", password, err)
		return err
	}

	return nil
}
