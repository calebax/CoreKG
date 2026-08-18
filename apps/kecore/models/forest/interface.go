package forest

import (
	"context"
	"errors"

	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoperm"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/logs"
)

// ContextModel 参数封装
type ContextModel struct {
	ResourceID   uint
	ResourceType foresttype.ResourceType
	Opt          *dtoperm.PermOption
	ScopeType    foresttype.ScopeType
	ScopeID      uint
	Action       foresttype.ActionType
}

type AccessResult struct {
	ManagerList []uint
	ViewerList  []uint
	BanList     []uint
	ScopeType   foresttype.PublicScope
}

// AccessProvider define action about different resource
type AccessProvider interface {
	Apply(ctx context.Context) error
	Get(ctx context.Context) (*AccessResult, error)
	Action(ctx context.Context) (*AccessResult, error)
}

func NewAccessProvider(ctx context.Context, model *ContextModel) AccessProvider {
	switch model.ResourceType {
	case foresttype.ResourceTypeForestFile:
		return &FileAccessProvider{Model: model}
	default:
		logs.ErrorContextf(ctx, "NewAccessProvider: unsupported resource type: %v", model.ResourceType)
		return &NilAccessProvider{Model: model}
	}
}

// NilAccessProvider access provider impl for unsupported resource type
type NilAccessProvider struct {
	Model *ContextModel
}

func (n *NilAccessProvider) Apply(ctx context.Context) error {
	logs.ErrorContextf(ctx, "NilAccessProvider: unsupported resource type: %v", n.Model.ResourceType)
	return errors.New("unsupported resource type")
}

func (n *NilAccessProvider) Get(ctx context.Context) (*AccessResult, error) {
	logs.ErrorContextf(ctx, "NilAccessProvider: unsupported resource type: %v", n.Model.ResourceType)
	return nil, errors.New("unsupported resource type")
}

func (n *NilAccessProvider) Action(ctx context.Context) (*AccessResult, error) {
	logs.ErrorContextf(ctx, "NilAccessProvider: unsupported resource type: %v", n.Model.ResourceType)
	return nil, errors.New("unsupported resource type")
}
