package apis

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/websearch/models/cache"
	"github.com/insmtx/corekg/apps/websearch/models/cursor"
	"github.com/insmtx/corekg/apps/websearch/models/domain"
	"github.com/insmtx/corekg/apps/websearch/models/provider"
	"github.com/insmtx/corekg/apps/websearch/services/svcsearch"
	"github.com/ygpkg/yg-go/apis/constants"
	"github.com/ygpkg/yg-go/logs"
	"go.uber.org/zap"
)

type benchmarkSearcher struct{}

func (benchmarkSearcher) Name() domain.ProviderName { return domain.ProviderNameBing }

func (benchmarkSearcher) Search(_ context.Context, request domain.SearchRequest) (domain.SearchResponse, error) {
	return domain.SearchResponse{
		Query: request.Query, Provider: domain.ProviderNameBing,
		Results: []domain.SearchResult{{Title: "Go", URL: "https://go.dev", Snippet: "The Go programming language", Rank: 1, Provider: domain.ProviderNameBing}},
		Meta:    domain.Meta{TookMS: 1}, Warnings: []domain.Warning{},
	}, nil
}

func BenchmarkSearchHandlerLatency(b *testing.B) {
	registry := provider.NewRegistry()
	if err := registry.Register(benchmarkSearcher{}); err != nil {
		b.Fatal(err)
	}
	memoryCache, err := cache.NewMemory(1000, time.Minute, time.Hour, time.Now)
	if err != nil {
		b.Fatal(err)
	}
	service := svcsearch.NewSearchService(registry, memoryCache, time.Now)
	codec, err := cursor.New("benchmark-secret", time.Hour, time.Now)
	if err != nil {
		b.Fatal(err)
	}
	handler, err := NewHandler(HandlerOptions{Searcher: service, Cursor: codec, CacheBypass: true, AllowRequestProviders: true, EnabledProviders: []string{"bing"}, ProviderVisibility: "public"})
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
	router.POST("/v3/websearch.Search", handler.Search)
	body := []byte(`{"query":"golang context cancellation","limit":10,"routing":{"providers":["bing"]}}`)
	for _, concurrency := range []int{1, 8, 32} {
		b.Run(fmt.Sprintf("concurrency-%d", concurrency), func(b *testing.B) {
			reportLatency(b, concurrency, func() bool {
				request := httptest.NewRequest(http.MethodPost, "/v3/websearch.Search", bytes.NewReader(body))
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)
				return response.Code == http.StatusOK
			})
		})
	}
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
