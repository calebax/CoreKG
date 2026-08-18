package connectors

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/pkgs/connectors"
	"github.com/insmtx/corekg/pkgs/connectors/tokenmgr"
)

// CreateExternalBindingReq 创建外部绑定请求
type CreateExternalBindingReq struct {
	Uin          uint              `json:"uin" binding:"required"`
	CompanyID    uint              `json:"company_id" binding:"required"`
	Platform     tokenmgr.Platform `json:"platform" binding:"required"`
	Provider     string            `json:"provider" binding:"required"`
	ExternalID   string            `json:"external_id" binding:"required"`
	Email        string            `json:"email,omitempty"`
	Avatar       string            `json:"avatar,omitempty"`
	AccessToken  string            `json:"access_token" binding:"required"`
	RefreshToken string            `json:"refresh_token,omitempty"`
	ExpiresIn    time.Time         `json:"expires_in,omitempty"`
}

// UpdateExternalBindingReq 更新外部绑定请求
type UpdateExternalBindingReq struct {
	ID           uint   `json:"id" binding:"required"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	Status       *int8  `json:"status,omitempty"`
}

// DeleteExternalBindingReq 删除外部绑定请求
type DeleteExternalBindingReq struct {
	ID uint `json:"id" binding:"required"`
}

// QueryExternalBindingReq 查询外部绑定请求
type QueryExternalBindingReq struct {
	Uin        uint   `json:"uin,omitempty"`
	Platform   string `json:"platform,omitempty"`
	ExternalID string `json:"external_id,omitempty"`
	Status     *int8  `json:"status,omitempty"`
	Page       int    `json:"page,omitempty"`
	PageSize   int    `json:"page_size,omitempty"`
}

type BindingProvider struct {
	Provider string `json:"provider"` // 平台标识，如 "gmail", "slack"
	Logo     string `json:"logo"`     // 图标 URL
}

// CreateExternalBinding 创建外部绑定
func CreateExternalBinding(ctx context.Context, req *CreateExternalBindingReq) (*tokenmgr.ExternalToken, error) {
	binding := &tokenmgr.ExternalToken{
		Uin:          req.Uin,
		CompanyID:    req.CompanyID,
		Platform:     req.Platform,
		Provider:     req.Provider,
		ExternalID:   req.ExternalID,
		Email:        req.Email,
		Avatar:       req.Avatar,
		AccessToken:  req.AccessToken,
		RefreshToken: req.RefreshToken,
		ExpiresAt:    &req.ExpiresIn,
		Status:       1, // 默认绑定状态
	}

	if err := tokenmgr.SaveToken(ctx, binding); err != nil {
		return nil, fmt.Errorf("failed to create external binding: %w", err)
	}

	return binding, nil
}

func QueryBindings(ctx context.Context, req *QueryExternalBindingReq) ([]*tokenmgr.ExternalToken, error) {
	return ListBindings(ctx, req.Uin)
}

func ListSupportedProviders(ctx context.Context) ([]*connectors.ProviderInfo, error) {
	providers, err := connectors.ListSupportedProviders()
	if err != nil {
		return nil, fmt.Errorf("failed to list supported providers: %w", err)
	}
	result := make([]*connectors.ProviderInfo, 0)
	for i := range providers {
		provider := providers[i]
		result = append(result, &provider)
	}
	return result, nil
}

// DeleteExternalBinding 删除外部绑定
func DeleteExternalBinding(ctx context.Context, req *DeleteExternalBindingReq) error {
	return DeleteBindingByID(ctx, req.ID)
}
