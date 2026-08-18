package apis

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/insmtx/corekg/apps/websearch/internal/dto"
	"github.com/insmtx/corekg/apps/websearch/models/cursor"
	"github.com/insmtx/corekg/apps/websearch/models/domain"
	"github.com/insmtx/corekg/apps/websearch/models/searchplan"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// Searcher executes a normalized search request.
type Searcher interface {
	Search(context.Context, domain.SearchRequest) (domain.SearchResponse, error)
}

// HandlerOptions defines the validated dependencies and request policy used by Handler.
type HandlerOptions struct {
	Searcher              Searcher
	Cursor                *cursor.Codec
	Timeout               time.Duration
	MaxTimeout            time.Duration
	CacheBypass           bool
	AllowRequestProviders bool
	EnabledProviders      []string
	ProviderVisibility    string
}

// Handler serves the public WebSearch action without owning a Gin engine.
type Handler struct {
	searcher              Searcher
	cursor                *cursor.Codec
	timeout               time.Duration
	maxTimeout            time.Duration
	cacheBypass           bool
	allowRequestProviders bool
	enabledProviders      []string
	providerVisibility    string
}

// NewHandler validates dependencies and creates a WebSearch handler.
func NewHandler(options HandlerOptions) (*Handler, error) {
	if options.Searcher == nil || options.Cursor == nil {
		return nil, fmt.Errorf("searcher and cursor are required")
	}
	if options.Timeout <= 0 {
		options.Timeout = 20 * time.Second
	}
	if options.MaxTimeout <= 0 {
		options.MaxTimeout = 60 * time.Second
	}
	return &Handler{
		searcher:              options.Searcher,
		cursor:                options.Cursor,
		timeout:               options.Timeout,
		maxTimeout:            options.MaxTimeout,
		cacheBypass:           options.CacheBypass,
		allowRequestProviders: options.AllowRequestProviders,
		enabledProviders:      append([]string(nil), options.EnabledProviders...),
		providerVisibility:    options.ProviderVisibility,
	}, nil
}

// Search validates one request and writes the stable WebSearch response contract.
func (handler *Handler) Search(ctx *gin.Context) {
	var payload dto.SearchRequest
	if err := dto.Decode(ctx.Request.Body, &payload); err != nil {
		writeProblem(ctx, http.StatusBadRequest, "invalid_request", "Invalid request", "The request body is invalid.", false, "")
		return
	}
	payload.Query = strings.Join(strings.Fields(payload.Query), " ")
	if payload.Query == "" || len([]rune(payload.Query)) > 256 {
		writeProblem(ctx, http.StatusBadRequest, "invalid_request", "Invalid query", "query must contain between 1 and 256 characters.", false, "query")
		return
	}
	if payload.Limit == 0 {
		payload.Limit = 10
	}
	if payload.Limit < 1 || payload.Limit > 20 {
		writeProblem(ctx, http.StatusBadRequest, "invalid_request", "Invalid limit", "limit must be between 1 and 20.", false, "limit")
		return
	}
	if len(payload.Cursor) > 4096 {
		writeProblem(ctx, http.StatusBadRequest, "invalid_request", "Invalid cursor", "cursor must not exceed 4096 characters.", false, "cursor")
		return
	}
	requestTimeout, err := dto.ParseTimeout(payload.Timeout, handler.timeout, handler.maxTimeout)
	if err != nil {
		writeProblem(ctx, http.StatusBadRequest, "invalid_request", "Invalid timeout", err.Error(), false, "timeout")
		return
	}
	providers, err := validateProviders(payload.Routing.Providers, handler.enabledProviders)
	if err != nil {
		writeProblem(ctx, http.StatusBadRequest, "invalid_request", "Invalid providers", err.Error(), false, "routing.providers")
		return
	}
	if len(payload.Routing.Providers) > 0 && !handler.allowRequestProviders {
		writeProblem(ctx, http.StatusForbidden, "provider_selection_forbidden", "Provider selection forbidden", "Request provider selection is disabled.", false, "routing.providers")
		return
	}
	region, err := domain.NormalizeRegion(payload.Filters.Region)
	if err != nil {
		writeProblem(ctx, http.StatusBadRequest, "invalid_request", "Invalid region", err.Error(), false, "filters.region")
		return
	}
	request, err := searchplan.Normalize(domain.SearchRequest{
		Query: payload.Query, Providers: providers, Region: region, Limit: payload.Limit,
		Filters: domain.SearchFilters{
			IncludeDomains: payload.Filters.IncludeDomains,
			ExcludeDomains: payload.Filters.ExcludeDomains,
		},
		QueryOptions: domain.SearchQueryOptions{
			ExactPhrases: payload.QueryOptions.ExactPhrases,
			AnyTerms:     payload.QueryOptions.AnyTerms,
			ExcludeTerms: payload.QueryOptions.ExcludeTerms,
			TitleTerms:   payload.QueryOptions.TitleTerms,
			FileTypes:    payload.QueryOptions.FileTypes,
		},
	})
	if err != nil {
		writeProblem(ctx, http.StatusBadRequest, "invalid_request", "Invalid advanced search options", err.Error(), false, "filters")
		return
	}
	request.Providers = searchplan.CompatibleProviders(request, request.Providers)
	if len(request.Providers) == 0 {
		writeProblem(ctx, http.StatusBadRequest, "unsupported_search_options", "Unsupported search options", "No enabled provider supports all requested query options.", false, "query_options")
		return
	}
	providers = request.Providers
	provider, page := domain.ProviderName(""), 1
	if payload.Cursor != "" {
		state, decodeErr := handler.cursor.Decode(payload.Cursor)
		if decodeErr != nil {
			code := "invalid_cursor"
			if errors.Is(decodeErr, cursor.ErrExpired) {
				code = "cursor_expired"
			}
			writeProblem(ctx, http.StatusBadRequest, code, "Invalid cursor", "The pagination cursor is invalid or expired.", false, "cursor")
			return
		}
		cursorProviders := providerNames(state.Providers)
		if len(payload.Routing.Providers) == 0 {
			providers = cursorProviders
		}
		request.Providers = providers
		request.Provider = domain.ProviderName(state.Provider)
		fingerprintMatches := state.RequestFingerprint == searchplan.Fingerprint(request)
		if state.Version == 1 {
			fingerprintMatches = state.QueryHash == cursor.QueryHash(payload.Query) &&
				region == "" && len(request.Filters.IncludeDomains) == 0 &&
				len(request.Filters.ExcludeDomains) == 0 &&
				len(request.QueryOptions.ExactPhrases)+len(request.QueryOptions.AnyTerms)+
					len(request.QueryOptions.ExcludeTerms)+len(request.QueryOptions.TitleTerms)+
					len(request.QueryOptions.FileTypes) == 0
		}
		if !fingerprintMatches || state.Limit != payload.Limit || !equalProviders(providers, cursorProviders) {
			writeProblem(ctx, http.StatusBadRequest, "cursor_mismatch", "Cursor mismatch", "query, limit, and routing providers must match the first page.", false, "cursor")
			return
		}
		provider, page = domain.ProviderName(state.Provider), state.ProviderPage
		request.ProviderPageToken = state.ProviderPageToken
	}
	if _, ok := stringSet(handler.enabledProviders)[string(provider)]; provider != "" && !ok {
		writeProblem(ctx, http.StatusServiceUnavailable, "provider_unavailable", "Provider unavailable", "The provider pinned by this cursor is not enabled.", true, "cursor")
		return
	}
	requestID := runtime.RequestID(ctx)
	request.Provider, request.Page, request.Refresh, request.RequestID = provider, page, handler.cacheBypass, requestID
	requestContext, cancel := context.WithTimeout(ctx.Request.Context(), requestTimeout)
	defer cancel()
	response, err := handler.searcher.Search(requestContext, request)
	if err != nil {
		logs.WarnContextw(ctx, "search failed", "request_id", requestID, "query_hash", cursor.QueryHash(payload.Query), "provider", provider, "error", err)
		if payload.Cursor != "" {
			writeProblem(ctx, http.StatusServiceUnavailable, "pagination_source_unavailable", "Pagination source unavailable", "The provider selected for this cursor is temporarily unavailable.", true, "cursor")
			return
		}
		writeSearchProblem(ctx, err)
		return
	}
	showProvider := handler.providerVisibility == "public"
	results := make([]dto.SearchResult, 0, len(response.Results))
	for _, item := range response.Results {
		result := dto.SearchResult{ID: resultID(item.URL), URL: item.URL, CanonicalURL: item.CanonicalURL, Domain: item.Domain, Title: item.Title, Snippet: item.Snippet, Rank: item.Rank}
		if showProvider {
			result.Provider = item.Provider
		}
		results = append(results, result)
	}
	pageResponse := dto.PageInfo{}
	hasMore := len(results) == payload.Limit && page < 10
	if response.PaginationKnown {
		hasMore = response.NextPageToken != "" && page < 10
	}
	if hasMore {
		request.Provider = response.Provider
		token, encodeErr := handler.cursor.Encode(cursor.State{
			RequestFingerprint: searchplan.Fingerprint(request),
			Provider:           string(response.Provider),
			Providers:          providerStrings(providers),
			ProviderPage:       page + 1,
			ProviderPageToken:  response.NextPageToken,
			Limit:              payload.Limit,
		})
		if encodeErr != nil {
			writeProblem(ctx, http.StatusInternalServerError, "cursor_encoding_failed", "Pagination unavailable", "The next page could not be created.", true, "")
			return
		}
		pageResponse.NextCursor, pageResponse.HasMore = token, true
	}
	meta := dto.SearchMeta{Cached: response.Meta.Cached, CacheAgeSeconds: response.Meta.CacheAgeSeconds, TookMS: response.Meta.TookMS}
	if showProvider {
		meta.Provider = response.Provider
	}
	logs.InfoContextw(ctx, "search completed", "request_id", requestID, "query_hash", cursor.QueryHash(payload.Query), "query_length", len([]rune(payload.Query)), "provider", response.Provider, "result_count", len(results), "cached", response.Meta.Cached, "took_ms", response.Meta.TookMS)
	warnings := response.Warnings
	if !showProvider {
		warnings = hideProviderWarnings(warnings)
	}
	ctx.JSON(http.StatusOK, dto.SearchResponse{RequestID: requestID, Query: payload.Query, Results: results, Page: pageResponse, Meta: meta, Warnings: warnings, Usage: dto.Usage{Units: 1}})
}

func writeSearchProblem(ctx *gin.Context, err error) {
	var searchErr *domain.SearchError
	if !errors.As(err, &searchErr) {
		writeProblem(ctx, http.StatusBadGateway, "upstream_failed", "Search failed", "Search providers are currently unavailable.", true, "")
		return
	}
	status, code := http.StatusBadGateway, string(searchErr.Code)
	switch searchErr.Code {
	case domain.ErrInvalidRequest:
		status = http.StatusBadRequest
	case domain.ErrProviderNotFound:
		status = http.StatusServiceUnavailable
	case domain.ErrSearchQueueFull, domain.ErrRateLimited:
		status = http.StatusTooManyRequests
	case domain.ErrUpstreamTimeout:
		status, code = http.StatusGatewayTimeout, "deadline_exceeded"
	}
	writeProblem(ctx, status, code, "Search failed", searchErr.Message, searchErr.Retryable, "")
}

func writeProblem(ctx *gin.Context, status int, code, title, detail string, retryable bool, parameter string) {
	ctx.Header("Content-Type", "application/problem+json")
	ctx.AbortWithStatusJSON(status, dto.Problem{Type: "https://api.example.com/problems/" + strings.ReplaceAll(code, "_", "-"), Title: title, Status: status, Code: code, Detail: detail, RequestID: runtime.RequestID(ctx), Retryable: retryable, Parameter: parameter})
}

func stringSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[strings.TrimSpace(value)] = struct{}{}
	}
	return result
}

func resultID(rawURL string) string { return "res_" + cursor.QueryHash(rawURL)[:16] }

func validateProviders(requested []domain.ProviderName, defaults []string) ([]domain.ProviderName, error) {
	if len(requested) == 0 {
		return providerNames(defaults), nil
	}
	enabled, seen := stringSet(defaults), map[string]struct{}{}
	result := make([]domain.ProviderName, 0, len(requested))
	for _, raw := range requested {
		name := strings.ToLower(strings.TrimSpace(string(raw)))
		if _, ok := enabled[name]; !ok {
			return nil, fmt.Errorf("provider %q is not enabled", name)
		}
		if _, ok := seen[name]; ok {
			return nil, fmt.Errorf("provider %q is duplicated", name)
		}
		seen[name] = struct{}{}
		result = append(result, domain.ProviderName(name))
	}
	return result, nil
}

func providerNames(values []string) []domain.ProviderName {
	result := make([]domain.ProviderName, len(values))
	for index, value := range values {
		result[index] = domain.ProviderName(value)
	}
	return result
}

func providerStrings(values []domain.ProviderName) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = string(value)
	}
	return result
}

func equalProviders(left, right []domain.ProviderName) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func hideProviderWarnings(values []domain.Warning) []domain.Warning {
	result := append([]domain.Warning(nil), values...)
	for index := range result {
		if result[index].Code == domain.WarningCodeProviderFallback {
			result[index].Message = "A search provider failed; a fallback provider was used"
			continue
		}
		replacer := strings.NewReplacer("baidu", "provider", "Baidu", "Provider", "bing", "provider", "Bing", "Provider", "brave", "provider", "Brave", "Provider", "duckduckgo", "provider", "DuckDuckGo", "Provider")
		result[index].Message = replacer.Replace(result[index].Message)
	}
	return result
}
