package employee

import (
	"context"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

var employeeJwtCfg *config.JwtConfig

// GenerateJwtToken 生成 jwt token
func GenerateJwtToken(ctx context.Context, empID uint, loginIP string) string {
	jwtCfg, err := GetJwtSecret()
	if err != nil {
		logs.ErrorContextf(ctx, "[account] generate jwt failed, %s", err)
		return ""
	}
	claims := auth.UserClaims{
		Uin:       empID,
		IssuedAt:  jwt.TimeFunc().Unix(),
		ExpiresAt: jwt.TimeFunc().Add(jwtCfg.Expire).Unix(),
		Issuer:    global.IssuerYYGUAdmin,
		Audience:  global.AudienceAdmin,
	}

	jwtStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(jwtCfg.Secret))
	if err != nil {
		logs.ErrorContext(ctx, "[account] generate jwt failed, %s", err)
		return ""
	}
	return jwtStr
}

// GetJwtSecret 获取 jwt secret
func GetJwtSecret() (*config.JwtConfig, error) {
	if employeeJwtCfg != nil {
		return employeeJwtCfg, nil
	}

	cfg := &config.JwtConfig{}
	err := settings.GetYaml("core", "jwt-yyguadmin", cfg)
	if err != nil {
		return nil, fmt.Errorf("get employee jwt config failed, %s", err.Error())
	}
	employeeJwtCfg = cfg
	return employeeJwtCfg, nil
}

// GetJwtCfgExpiry 获取 jwt 过期时间
func GetJwtCfgExpiry() int64 {
	if employeeJwtCfg != nil {
		sec := employeeJwtCfg.Expire.Seconds()
		if sec < 300 {
			//zsy debug
			return 60 * 60 * 8
		}
		return int64(sec)
	}

	cfg := &config.JwtConfig{}
	err := settings.GetYaml("account", "employee_jwt", cfg)
	if err != nil {
		panic(any(fmt.Errorf("get employee jwt config failed, %s", err)))
	}
	employeeJwtCfg = cfg

	sec := employeeJwtCfg.Expire.Seconds()
	if sec < 300 {
		return 60 * 60 * 8
	}
	return int64(sec)
}
