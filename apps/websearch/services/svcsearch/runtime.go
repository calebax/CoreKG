package svcsearch

import (
	"fmt"
	"net/url"
	"time"

	"github.com/insmtx/corekg/apps/websearch/conf"
	"github.com/insmtx/corekg/apps/websearch/models/cache"
	"github.com/insmtx/corekg/apps/websearch/models/cursor"
	"github.com/insmtx/corekg/apps/websearch/models/debugartifact"
	"github.com/insmtx/corekg/apps/websearch/models/headerprofile"
)

// Runtime owns the long-lived resources used by one WebSearch process.
type Runtime struct {
	// Service executes search orchestration for the process.
	Service *SearchService
	// Cursor encodes and decodes opaque pagination state.
	Cursor *cursor.Codec
	close  func()
}

// NewRuntime builds the provider pools, cache, cursor codec, and search service.
func NewRuntime(configValue conf.Config) (*Runtime, error) {
	headers, err := headerprofile.NewChromiumDesktopPool(configValue.UserAgent)
	if err != nil {
		return nil, fmt.Errorf("create header profiles: %w", err)
	}
	artifacts, err := debugartifact.New(configValue.DebugDir, configValue.DebugPreviewBytes)
	if err != nil {
		return nil, err
	}
	runtimeValue, err := newSearchRuntime(configValue, artifacts, headers)
	if err != nil {
		return nil, fmt.Errorf("create search runtime: %w", err)
	}
	memoryCache, err := cache.NewMemory(configValue.CacheMaxItems, configValue.FreshTTL, configValue.StaleTTL, time.Now)
	if err != nil {
		runtimeValue.Close()
		return nil, err
	}
	service := NewSearchService(runtimeValue.registry, memoryCache, time.Now)
	service.SetTracer(runtimeValue.tracer)
	service.ConfigureLiveGuard(configValue.GlobalInflightMax, configValue.AutoQueueMax)
	cursorCodec, err := cursor.New(configValue.CursorSecret, configValue.CursorTTL, time.Now)
	if err != nil {
		runtimeValue.Close()
		return nil, err
	}
	return &Runtime{Service: service, Cursor: cursorCodec, close: runtimeValue.Close}, nil
}

// Close releases browser, provider-pool, and HTTP transport resources.
func (runtimeValue *Runtime) Close() {
	if runtimeValue != nil && runtimeValue.close != nil {
		runtimeValue.close()
	}
}

func origin(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host + "/"
}
