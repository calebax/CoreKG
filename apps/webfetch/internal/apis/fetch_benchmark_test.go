package apis

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/webfetch/models/domain"
	readcache "github.com/insmtx/corekg/apps/webfetch/models/fetch/cache"
	"github.com/insmtx/corekg/apps/webfetch/models/fetch/converter"
	"github.com/insmtx/corekg/apps/webfetch/models/fetch/detector"
	"github.com/insmtx/corekg/apps/webfetch/models/fetch/extractor"
	"github.com/insmtx/corekg/apps/webfetch/models/fetch/quality"
	"github.com/insmtx/corekg/apps/webfetch/models/fetch/site"
	"github.com/insmtx/corekg/apps/webfetch/services/svcfetch"
	"github.com/ygpkg/yg-go/apis/constants"
	"github.com/ygpkg/yg-go/logs"
	"go.uber.org/zap"
)

type benchmarkPolicy struct{}

func (benchmarkPolicy) ValidateAndResolve(_ context.Context, rawURL string) (domain.SafeTarget, error) {
	parsed, err := url.Parse(rawURL)
	return domain.SafeTarget{URL: parsed}, err
}

type benchmarkResourceReader struct{ body []byte }

func (benchmarkResourceReader) Name() domain.ImplementationName {
	return domain.ImplementationNameHTTPReader
}

func (reader benchmarkResourceReader) Read(_ context.Context, target domain.SafeTarget) (domain.Resource, error) {
	return domain.Resource{URL: target.URL.String(), FinalURL: target.URL.String(), StatusCode: http.StatusOK, ContentType: "text/html", Body: reader.body, Transport: domain.ReadTransportHTTP}, nil
}

func BenchmarkFetchHandlerLatency(b *testing.B) {
	htmlExtractor, err := extractor.NewHTMLExtractor()
	if err != nil {
		b.Fatal(err)
	}
	extractors, err := extractor.NewRegistry(htmlExtractor, extractor.PlainTextExtractor{})
	if err != nil {
		b.Fatal(err)
	}
	converters, err := converter.NewRegistry(converter.MarkdownConverter{}, converter.TextConverter{})
	if err != nil {
		b.Fatal(err)
	}
	memoryCache, err := readcache.NewMemory(1000, time.Minute, time.Hour, time.Now)
	if err != nil {
		b.Fatal(err)
	}
	service, err := svcfetch.NewReadService(svcfetch.ReadServiceConfig{
		Policy: benchmarkPolicy{}, Strategies: site.NewRegistry(nil, site.GenericStrategy{}),
		HTTPReader: benchmarkResourceReader{body: benchmarkHTML(64 << 10)}, Detector: detector.NewMIMETypeDetector(),
		Extractors: extractors, Evaluator: quality.NewArticleQualityEvaluator(200), Converters: converters,
		Cache: memoryCache, OperationTimeout: 20 * time.Second, Now: time.Now,
	})
	if err != nil {
		b.Fatal(err)
	}
	handler, err := NewHandler(HandlerOptions{Reader: service, CacheBypass: true})
	if err != nil {
		b.Fatal(err)
	}
	router := gin.New()
	nopLogger := zap.NewNop().Sugar()
	router.Use(func(ctx *gin.Context) {
		ctx.Set(constants.CtxKeyRequestID, "req_benchmark")
		logs.SetContextLogger(ctx, nopLogger)
		ctx.Next()
	})
	router.POST("/v3/webfetch.Fetch", handler.Fetch)
	body := []byte(`{"url":"https://go.dev/blog/context","timeout":"20s","output":{"format":"markdown","max_chars":30000}}`)
	for _, concurrency := range []int{1, 8, 32} {
		b.Run(fmt.Sprintf("concurrency-%d", concurrency), func(b *testing.B) {
			reportLatency(b, concurrency, func() bool {
				request := httptest.NewRequest(http.MethodPost, "/v3/webfetch.Fetch", bytes.NewReader(body))
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)
				return response.Code == http.StatusOK
			})
		})
	}
}

func benchmarkHTML(targetBytes int) []byte {
	const paragraph = `<p>Go programs express error handling with explicit values and use context cancellation to coordinate work across request-scoped operations.</p>`
	count := targetBytes/len(paragraph) + 1
	return []byte(`<html><head><title>Go Context</title></head><body><main><article><h1>Go Context</h1>` + strings.Repeat(paragraph, count) + `</article></main></body></html>`)
}

func reportLatency(b *testing.B, concurrency int, operation func() bool) {
	b.Helper()
	durations := make([]int64, b.N)
	jobs := make(chan int)
	var failed atomic.Bool
	var workers sync.WaitGroup
	workers.Add(concurrency)
	for worker := 0; worker < concurrency; worker++ {
		go func() {
			defer workers.Done()
			for index := range jobs {
				startedAt := time.Now()
				if !operation() {
					failed.Store(true)
				}
				durations[index] = time.Since(startedAt).Nanoseconds()
			}
		}()
	}
	b.ReportAllocs()
	b.ResetTimer()
	startedAt := time.Now()
	for index := 0; index < b.N; index++ {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
	elapsed := time.Since(startedAt)
	b.StopTimer()
	if failed.Load() {
		b.Fatal("handler returned a non-200 response")
	}
	sort.Slice(durations, func(left, right int) bool { return durations[left] < durations[right] })
	var total int64
	for _, duration := range durations {
		total += duration
	}
	b.ReportMetric(float64(durations[0]), "min-ns/op")
	b.ReportMetric(float64(total)/float64(len(durations)), "avg-ns/op")
	b.ReportMetric(float64(percentile(durations, 50)), "p50-ns/op")
	b.ReportMetric(float64(percentile(durations, 95)), "p95-ns/op")
	b.ReportMetric(float64(percentile(durations, 99)), "p99-ns/op")
	b.ReportMetric(float64(durations[len(durations)-1]), "max-ns/op")
	b.ReportMetric(float64(b.N)/elapsed.Seconds(), "req/s")
}

func percentile(sorted []int64, value int) int64 {
	index := (len(sorted) - 1) * value / 100
	return sorted[index]
}
