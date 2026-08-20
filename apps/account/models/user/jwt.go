package user

import (
	"context"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/platform/login_setting"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/random"
)

// GenerateJwtToken 生成 jwt token
func GenerateJwtToken(ctx context.Context, uinID uint, loginWay auth.LoginWay, loginIP string, issuer string, expires ...time.Duration) string {
	var expire time.Duration
	if len(expires) > 0 {
		expire = expires[0]
	}
	return generateJwtToken(ctx, uinID, loginWay, loginIP, issuer, expire)
}

func generateJwtToken(ctx context.Context, uinID uint, loginWay auth.LoginWay, loginIP string, issuer string, expire time.Duration) string {
	jwtOpt, err := auth.GetJwtSetting(issuer)
	if err != nil {
		logs.ErrorContextf(ctx, "[account] get jwt secret failed, %s", err)
		return ""
	}
	claims := auth.UserClaims{
		Uin:       uinID,
		Issuer:    issuer,
		IssuedAt:  jwt.TimeFunc().Unix(),
		ExpiresAt: jwt.TimeFunc().Add(jwtOpt.Expire).Unix(),
		LoginWay:  loginWay,
		Audience:  global.AudienceUser,
	}
	if expire > 0 {
		claims.ExpiresAt = jwt.TimeFunc().Add(expire).Unix()
	}
	if loginWay == login_setting.LoginWayOpoEmail && expire <= 0 {
		claims.ExpiresAt = jwt.TimeFunc().Add(365 * 24 * time.Hour).Unix()
	}

	jwtStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(jwtOpt.Secret))
	if err != nil {
		logs.ErrorContextf(ctx, "[account] generate jwt failed, %s", err)
		return ""
	}
	return jwtStr
}

const (
	RefreshTokenLen    = 16
	RefreshTokenExpire = 5
)

// GenerateRefreshToken 生成 refresh token存入redis
func GenerateRefreshToken(ctx context.Context, userID uint) (string, error) {
	key := fmt.Sprintf("refresh_token:%d", userID)
	refreshToken := random.String(RefreshTokenLen)
	err := redispool.SetString(key, refreshToken, RefreshTokenExpire*time.Minute)
	if err != nil {
		logs.ErrorContextf(ctx, "[account] generate refresh token failed, %s", err)
		return "", err
	}
	return refreshToken, nil
}

// GetRefreshToken 获取 refresh token
func GetRefreshToken(userID uint) (string, error) {
	key := fmt.Sprintf("refresh_token:%d", userID)
	return redispool.GetString(key)
}
