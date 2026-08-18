package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"

	"github.com/insmtx/corekg/apps/webfetch/models/domain"
)

var timeoutPattern = regexp.MustCompile(`^([1-9][0-9]*)(ms|s)$`)

// FetchRequest is the public WebFetch request body.
type FetchRequest struct {
	URL     string `json:"url" binding:"required,max=2048"`
	Timeout string `json:"timeout,omitempty"`
	Output  struct {
		Format   domain.OutputFormat `json:"format"`
		MaxChars int                 `json:"max_chars"`
	} `json:"output"`
}

// FetchResponse is the stable successful WebFetch response body.
type FetchResponse struct {
	RequestID string               `json:"request_id"`
	Document  Document             `json:"document"`
	Meta      Meta                 `json:"meta"`
	Warnings  []domain.ReadWarning `json:"warnings"`
	Usage     Usage                `json:"usage"`
}

// Document contains the normalized fetched content.
type Document struct {
	URL         string              `json:"url"`
	FinalURL    string              `json:"final_url"`
	Title       string              `json:"title,omitempty"`
	Author      string              `json:"author,omitempty"`
	PublishedAt string              `json:"published_at,omitempty"`
	Language    string              `json:"language,omitempty"`
	SourceType  domain.SourceType   `json:"source_type"`
	ContentType string              `json:"content_type"`
	StatusCode  int                 `json:"status_code"`
	Content     string              `json:"content"`
	Format      domain.OutputFormat `json:"format"`
	RetrievedAt time.Time           `json:"retrieved_at"`
}

// Meta contains non-document execution metadata.
type Meta struct {
	Cached          bool                 `json:"cached"`
	Transport       domain.ReadTransport `json:"transport"`
	Truncated       bool                 `json:"truncated"`
	ContentLength   int                  `json:"content_length"`
	CacheAgeSeconds int64                `json:"cache_age_seconds,omitempty"`
	TookMS          int64                `json:"took_ms"`
}

// Usage contains API Market billing units for one request.
type Usage struct {
	Units int `json:"units"`
}

// Problem is the stable error response returned by WebFetch.
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
func Decode(reader io.Reader, target *FetchRequest) error {
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
