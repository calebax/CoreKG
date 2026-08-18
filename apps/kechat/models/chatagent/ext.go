package chatagent

import (
	"context"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/logs"
)

func GenerateExternalAgentToken(ctx context.Context, agentID uint) (string, error) {
	jwtOpt, err := auth.GetJwtSetting(global.IssuerYYGU)
	if err != nil {
		logs.ErrorContext(ctx, "[kecore] get jwt secret failed, %s", err)
		return "", fmt.Errorf("[kecore] get jwt secret failed")
	}

	claims := ExternalClaims{
		ExtID:     agentID,
		ExtType:   ExternalAgent,
		Issuer:    global.IssuerYYGU,
		IssuedAt:  jwt.TimeFunc().Unix(),
		ExpiresAt: jwt.TimeFunc().Add(jwtOpt.Expire).Unix(),
		Audience:  global.AudienceExternal,
	}

	jwtStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(jwtOpt.Secret))
	if err != nil {
		logs.ErrorContext(ctx, "[kecore] generate jwt failed, %s", err)
		return "", fmt.Errorf("[kecore] generate jwt failed, %s", err)
	}
	return jwtStr, nil
}

type ExternalType string

var (
	ExternalAgent ExternalType = "agent"
)

// ExternalClaims 用户信息
type ExternalClaims struct {
	// ID，主体唯一ID
	ExtID uint `json:"ei,omitempty"`
	// ExtType，主体类型
	ExtType ExternalType `json:"et,omitempty"`
	// IssuedAt 创建时间
	IssuedAt int64 `json:"t,omitempty"`
	// ExpiresAt 过期时间
	ExpiresAt int64 `json:"e,omitempty"`
	// Issuer 签发者 区分不同签发者
	Issuer string `json:"i,omitempty"`
	// Audience 接收者
	Audience string `json:"a,omitempty"`
}

// Valid time based claims "exp, iat, nbf".
// There is no accounting for clock skew.
// As well, if any of the above claims are not in the token, it will still
// be considered a valid claim.
func (c ExternalClaims) Valid() error {
	vErr := new(jwt.ValidationError)
	now := jwt.TimeFunc().Unix()

	if c.IssuedAt > now {
		vErr.Inner = fmt.Errorf("token used before issued")
		vErr.Errors |= jwt.ValidationErrorIssuedAt
	}
	if c.ExpiresAt < now {
		vErr.Inner = fmt.Errorf("token is expired")
		vErr.Errors |= jwt.ValidationErrorExpired
	}

	if vErr.Errors == 0 {
		return nil
	}

	return vErr
}
