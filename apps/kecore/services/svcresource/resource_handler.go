package svcresource

import (
	"context"

	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoresource"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/logs"
)

type PluginScopeHandler struct {
}

func NewPluginScopeHandler() *PluginScopeHandler {
	return &PluginScopeHandler{}
}

func (h *PluginScopeHandler) BeforeSetScope(ctx context.Context, req *dtoresource.SetResourceScopeRequest) error {
	logs.InfoContextf(ctx, "[PluginScopeHandler] BeforeSetScope called for ResourceID: %v", req.Request.ResourceID)
	return nil
}

func (h *PluginScopeHandler) AfterSetScope(ctx context.Context, req *dtoresource.SetResourceScopeRequest, addedScopes foresttype.KeResourceScopeList) error {
	logs.InfoContextf(ctx, "[PluginScopeHandler] AfterSetScope called for ResourceID: %v", req.Request.ResourceID)
	return nil
}
