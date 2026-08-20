package user

import (
	"context"

	"github.com/insmtx/corekg/pkgs/types"
	"github.com/ygpkg/yg-go/logs"
	"golang.org/x/crypto/bcrypt"
)

const (
	_passwordBcryptCost = bcrypt.DefaultCost
)

// EncryptPassword 加密用户密码
func EncryptPassword(ctx context.Context, password string) *string {
	if password == "" {
		return nil
	}
	encPassword, err := bcrypt.GenerateFromPassword([]byte(password), _passwordBcryptCost)
	if err != nil {
		logs.WarnContextf(ctx, "[EncryptPassword] generate password(%s) failed, %s", password, err)
		return nil
	}
	return types.String(string(encPassword))
}

// VerifyPassword 验证密码
func VerifyPassword(ctx context.Context, input, encrypted string) bool {
	if input == "" || encrypted == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(encrypted), []byte(input))
	if err != nil {
		logs.WarnContextf(ctx, "[VerifyPassword] compare password failed:input %s encrypted %s ,err %v", input, encrypted, err)
		return false
	}
	return true
}
