package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"

	"github.com/insmtx/corekg/apps/websearch/models/domain"
)

var timeoutPattern = regexp.MustCompile(`^([1-9][0-9]*)(ms|s)$`)

// SearchRequest is the public WebSearch request body.
type SearchRequest struct {
	Query   string `json:"query"`
	Limit   int    `json:"limit,omitempty"`
	Cursor  string `json:"cursor,omitempty"`
	Timeout string `json:"timeout,omitempty"`
	Routing struct {
		Providers []domain.ProviderName `json:"providers,omitempty"`
	} `json:"routing,omitempty"`
	Filters struct {
		Region         string   `json:"region,omitempty"`
		IncludeDomains []string `json:"include_domains,omitempty"`
		ExcludeDomains []string `json:"exclude_domains,omitempty"`
	} `json:"filters,omitempty"`
	QueryOptions struct {
		ExactPhrases []string `json:"exact_phrases,omitempty"`
		AnyTerms     []string `json:"any_terms,omitempty"`
		ExcludeTerms []string `json:"exclude_terms,omitempty"`
		TitleTerms   []string `json:"title_terms,omitempty"`
		FileTypes    []string `json:"file_types,omitempty"`
	} `json:"query_options,omitempty"`
}

// SearchResponse is the stable successful WebSearch response body.
type SearchResponse struct {
	RequestID string           `json:"request_id"`
	Query     string           `json:"query"`
	Results   []SearchResult   `json:"results"`
	Page      PageInfo         `json:"page"`
	Meta      SearchMeta       `json:"meta"`
	Warnings  []domain.Warning `json:"warnings"`
	Usage     Usage            `json:"usage"`
}

// SearchResult is one normalized public search result.
type SearchResult struct {
	ID           string              `json:"id"`
	URL          string              `json:"url"`
	CanonicalURL string              `json:"canonical_url"`
	Domain       string              `json:"domain"`
	Title        string              `json:"title"`
	Snippet      string              `json:"snippet"`
	Rank         int                 `json:"rank"`
	Provider     domain.ProviderName `json:"provider,omitempty"`
}

// PageInfo describes opaque cursor pagination state.
type PageInfo struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// SearchMeta contains non-result execution metadata.
type SearchMeta struct {
	Cached          bool                `json:"cached"`
	CacheAgeSeconds int64               `json:"cache_age_seconds,omitempty"`
	TookMS          int64               `json:"took_ms"`
	Provider        domain.ProviderName `json:"provider,omitempty"`
}

// Usage contains API Market billing units for one request.
type Usage struct {
	Units int `json:"units"`
}

// Problem is the stable error response returned by WebSearch.
type Problem struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Code      string `json:"code"`
	Detail    string `json:"detail"`
	RequestID string `json:"request_id"`
	Retryable bool   `json:"retryable"`
	Parameter string `json:"parameter,omitempty"`
}

// Decode reads exactly one strict JSON request value.
func Decode(reader io.Reader, target *SearchRequest) error {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("request must contain one JSON value")
	}
	return nil
}

// ParseTimeout validates the public integer millisecond or second syntax.
func ParseTimeout(value string, defaultValue, maximum time.Duration) (time.Duration, error) {
	if value == "" {
		return defaultValue, nil
	}
	matches := timeoutPattern.FindStringSubmatch(value)
	if matches == nil {
		return 0, fmt.Errorf("timeout must be an integer followed by ms or s")
	}
	number, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("timeout is too large")
	}
	unit := time.Millisecond
	if matches[2] == "s" {
		unit = time.Second
	}
	if number > int64(maximum/unit)+1 {
		return 0, fmt.Errorf("timeout exceeds maximum %s", maximum)
	}
	duration := time.Duration(number) * unit
	if duration < 100*time.Millisecond {
		return 0, fmt.Errorf("timeout must be at least 100ms")
	}
	if duration > maximum {
		return 0, fmt.Errorf("timeout must not exceed %s", maximum)
	}
	return duration, nil
}
