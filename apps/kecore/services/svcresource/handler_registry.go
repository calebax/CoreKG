package svcresource

import (
	"context"
	"sync"

	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoresource"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
)

// handlerRegistry 资源权限处理器注册中心
type handlerRegistry struct {
	mu       sync.RWMutex
	handlers map[foresttype.ResourceType]ResourceScopeHandler
}

// ResourceScopeHandler 资源权限处理器接口
type ResourceScopeHandler interface {
	// BeforeSetScope 设置权限前的钩子，可以进行业务校验、数据预处理等
	BeforeSetScope(ctx context.Context, req *dtoresource.SetResourceScopeRequest) error

	// AfterSetScope 设置权限后的钩子，可以进行后续业务处理、通知等
	AfterSetScope(ctx context.Context, req *dtoresource.SetResourceScopeRequest, addedScopes foresttype.KeResourceScopeList) error
}

// registry 处理器注册中心实例
var registry *handlerRegistry

func init() {
	registry = &handlerRegistry{
		handlers: make(map[foresttype.ResourceType]ResourceScopeHandler),
	}
	RegisterHandlers(map[foresttype.ResourceType]ResourceScopeHandler{
		foresttype.ResourceTypePlugin: NewPluginScopeHandler(),
	})
}

// RegisterHandlers 批量注册处理器
func RegisterHandlers(handlers map[foresttype.ResourceType]ResourceScopeHandler) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	for resourceType, handler := range handlers {
		registry.handlers[resourceType] = handler
	}
}

// GetHandler 获取资源类型对应的处理器
func GetHandler(resourceType foresttype.ResourceType) (ResourceScopeHandler, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	handler, ok := registry.handlers[resourceType]
	return handler, ok
}
