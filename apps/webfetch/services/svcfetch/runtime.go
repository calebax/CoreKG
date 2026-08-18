package svcfetch

import (
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/webfetch/conf"
	readpipe "github.com/insmtx/corekg/apps/webfetch/models/fetch"
	readcache "github.com/insmtx/corekg/apps/webfetch/models/fetch/cache"
	"github.com/insmtx/corekg/apps/webfetch/models/fetch/converter"
	"github.com/insmtx/corekg/apps/webfetch/models/fetch/detector"
	"github.com/insmtx/corekg/apps/webfetch/models/fetch/extractor"
	"github.com/insmtx/corekg/apps/webfetch/models/fetch/quality"
	readreader "github.com/insmtx/corekg/apps/webfetch/models/fetch/reader"
	"github.com/insmtx/corekg/apps/webfetch/models/fetch/safeurl"
	"github.com/insmtx/corekg/apps/webfetch/models/fetch/site"
	"github.com/insmtx/corekg/apps/webfetch/models/headerprofile"
	"github.com/insmtx/corekg/apps/webfetch/models/transport/chromebrowser"
)

// Runtime owns the long-lived resources used by one WebFetch process.
type Runtime struct {
	// Service executes the content-read pipeline for the process.
	Service *ReadService
	close   func()
}

// NewRuntime builds the readers, extraction pipeline, cache, and read service.
func NewRuntime(configValue conf.Config) (*Runtime, error) {
	policy := safeurl.NewPolicy(nil, safeurl.Config{AllowedHosts: configValue.HostAllowlist})
	httpReader, err := readreader.NewHTTPReader(policy, readreader.Config{Timeout: configValue.HTTPTimeout, MaxBodyBytes: configValue.MaxBodyBytes, MaxRedirects: configValue.MaxRedirects, UserAgent: configValue.UserAgent})
	if err != nil {
		return nil, fmt.Errorf("create HTTP reader: %w", err)
	}
	closeResources := func() {}
	var browserReader readpipe.ResourceReader
	if configValue.BrowserEnabled {
		headers, profileErr := headerprofile.NewChromiumDesktopPool(configValue.UserAgent)
		if profileErr != nil {
			return nil, fmt.Errorf("create header profiles: %w", profileErr)
		}
		client, clientErr := chromebrowser.New(chromebrowser.Config{ProfileDir: configValue.ChromeProfileDir, ExecPath: configValue.ChromePath, Timeout: configValue.BrowserTimeout, PostLoadWait: configValue.BrowserWait, Headless: configValue.ChromeHeadless, DisableSandbox: configValue.ChromeNoSandbox, MaxBodyBytes: int(configValue.MaxBodyBytes), MaxConcurrentTabs: configValue.BrowserSlots, HeaderProfiles: headers})
		if clientErr != nil {
			return nil, fmt.Errorf("create browser runtime: %w", clientErr)
		}
		closeResources = client.Close
		browserReader, err = readreader.NewBrowserReader(policy, client, configValue.MaxBodyBytes)
		if err != nil {
			closeResources()
			return nil, err
		}
	}
	htmlExtractor, err := extractor.NewHTMLExtractor()
	if err != nil {
		closeResources()
		return nil, err
	}
	extractors, err := extractor.NewRegistry(htmlExtractor, extractor.PlainTextExtractor{})
	if err != nil {
		closeResources()
		return nil, err
	}
	converters, err := converter.NewRegistry(converter.MarkdownConverter{}, converter.TextConverter{})
	if err != nil {
		closeResources()
		return nil, err
	}
	cache, err := readcache.NewMemory(configValue.CacheMaxItems, configValue.FreshTTL, configValue.StaleTTL, time.Now)
	if err != nil {
		closeResources()
		return nil, err
	}
	service, err := NewReadService(ReadServiceConfig{Policy: policy, Strategies: site.NewRegistry(nil, site.GenericStrategy{}), HTTPReader: httpReader, BrowserReader: browserReader, Detector: detector.NewMIMETypeDetector(), Extractors: extractors, Evaluator: quality.NewArticleQualityEvaluator(200), Converters: converters, Cache: cache, OperationTimeout: configValue.RequestTimeout, Now: time.Now})
	if err != nil {
		closeResources()
		return nil, err
	}
	return &Runtime{Service: service, close: closeResources}, nil
}

// Close releases browser resources owned by the runtime.
func (runtimeValue *Runtime) Close() {
	if runtimeValue != nil && runtimeValue.close != nil {
		runtimeValue.close()
	}
}
