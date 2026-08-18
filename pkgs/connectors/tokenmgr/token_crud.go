package tokenmgr

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
)

func getToken(ctx context.Context, uin uint, provider Platform) (*ExternalToken, error) {
	var externalToken ExternalToken
	if err := dbutil.Account().Where("uin = ? and provider = ?", uin, provider).First(&externalToken).Error; err != nil {
		logs.ErrorContextf(ctx, "token manager get token failed, uin: %d, provider: %s, err: %v", uin, provider, err)
		return nil, fmt.Errorf("get token failed: %w", err)
	}
	return &externalToken, nil
}

func saveToken(ctx context.Context, externalToken *ExternalToken) error {
	if err := dbutil.Account().Save(externalToken).Error; err != nil {
		logs.ErrorContextf(ctx, "token manager save token failed, uin: %d, provider: %s, err: %v", externalToken.Uin, externalToken.Provider, err)
		return fmt.Errorf("save token failed: %w", err)
	}
	return nil
}

// GetTokensByIDS 获取token
func GetTokensByIDS(ctx context.Context, ids []uint) ([]ExternalToken, error) {
	var externalTokens []ExternalToken
	if err := dbutil.Account().WithContext(ctx).
		Where("id in (?)", ids).Find(&externalTokens).Error; err != nil {
		logs.ErrorContextf(ctx, "GetTokensByIDS error: %v", err)
		return nil, err
	}
	return externalTokens, nil
}

// GetTokenByUin
func GetTokenByUin(ctx context.Context, uin uint, provider Platform) (*ExternalToken, error) {
	var externalToken *ExternalToken
	if err := dbutil.Account().WithContext(ctx).
		Where("uin = ? and provider = ?", uin, provider).
		First(&externalToken).Error; err != nil {
		logs.WarnContextf(ctx, "GetTokenByUin err:%v", err)
		return nil, err
	}
	return externalToken, nil
}
