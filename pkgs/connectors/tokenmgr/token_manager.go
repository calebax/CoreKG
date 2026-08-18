package tokenmgr

import (
	"context"
	"fmt"
	"time"

	"github.com/markbates/goth"
	"github.com/ygpkg/yg-go/logs"
)

func GetToken(ctx context.Context, uin uint, provider Platform) (*TokenInfo, bool) {
	externalToken, err := getToken(ctx, uin, provider)
	if err != nil {
		logs.WarnContextf(ctx, "token manager get token failed, uin: %d, provider: %s, err: %v", uin, provider, err)
		return nil, false
	}
	if externalToken == nil {
		return nil, false
	}

	var expiry time.Time
	if externalToken.ExpiresAt != nil {
		expiry = *externalToken.ExpiresAt
	}

	if !expiry.IsZero() && time.Now().After(expiry) {
		tokenInfo, ok := refreshToken(ctx, externalToken.Provider, externalToken.RefreshToken)
		if !ok {
			logs.WarnContextf(ctx, "token manager refresh token failed, uin: %d, provider: %s", uin, provider)
			return nil, false
		}
		externalToken.ExpiresAt = &tokenInfo.Expiry
		externalToken.AccessToken = tokenInfo.AccessToken
		if err := saveToken(ctx, externalToken); err != nil {
			logs.WarnContextf(ctx, "token manager save token failed, uin: %d, provider: %s, err: %v", uin, provider, err)
			return nil, false
		}
		tokenInfo.ExternalID = externalToken.ExternalID
		return tokenInfo, true
	}

	return &TokenInfo{
		ExternalID:   externalToken.ExternalID,
		AccessToken:  externalToken.AccessToken,
		RefreshToken: externalToken.RefreshToken,
		Expiry:       expiry,
		TokenType:    "Bearer",
		Provider:     externalToken.Provider,
	}, true
}

func SaveToken(ctx context.Context, externalToken *ExternalToken) error {
	// 加密token
	encryptedAccessToken, err := encryptToken(externalToken.AccessToken)
	if err != nil {
		logs.WarnContextf(ctx, "token manager save token failed, failed to encrypt access token: %w", err)
		return fmt.Errorf("failed to encrypt access token: %w", err)
	}
	encryptedRefreshToken, err := encryptToken(externalToken.RefreshToken)
	if err != nil {
		logs.WarnContextf(ctx, "token manager save token failed, failed to encrypt refresh token: %w", err)
		return fmt.Errorf("failed to encrypt refresh token: %w", err)
	}
	externalToken.AccessToken = encryptedAccessToken
	externalToken.RefreshToken = encryptedRefreshToken

	// 逻辑定义永久有效
	if externalToken.ExpiresAt == nil || externalToken.ExpiresAt.IsZero() {
		externalToken.ExpiresAt = nil
	}
	return saveToken(ctx, externalToken)
}

func refreshToken(ctx context.Context, provider string, refreshToken string) (*TokenInfo, bool) {
	authProvider, err := goth.GetProvider(provider)
	if err != nil {
		logs.WarnContextf(ctx, "token manager refresh token failed, failed to get provider: %w", err)
		return nil, false
	}

	if !authProvider.RefreshTokenAvailable() {
		logs.WarnContextf(ctx, "token manager refresh token failed, refresh token not available")
		return nil, false
	}

	token, err := authProvider.RefreshToken(refreshToken)
	if err != nil {
		logs.WarnContextf(ctx, "token manager refresh token failed, failed to refresh token: %w", err)
		return nil, false
	}
	return &TokenInfo{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		TokenType:    token.TokenType,
		Provider:     provider,
	}, true
}
