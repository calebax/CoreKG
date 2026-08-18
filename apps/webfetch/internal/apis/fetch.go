package apis

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/insmtx/corekg/apps/webfetch/internal/dto"
	"github.com/insmtx/corekg/apps/webfetch/models/domain"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// Reader executes a normalized content-read request.
type Reader interface {
	Read(context.Context, domain.ReadRequest) (domain.ReadResponse, error)
}

// HandlerOptions defines the validated dependencies and request policy used by Handler.
type HandlerOptions struct {
	Reader      Reader
	Timeout     time.Duration
	MaxTimeout  time.Duration
	CacheBypass bool
	LogURLQuery bool
}

// Handler serves the public WebFetch action without owning a Gin engine.
type Handler struct {
	reader      Reader
	timeout     time.Duration
	maxTimeout  time.Duration
	cacheBypass bool
	logURLQuery bool
}

// NewHandler validates dependencies and creates a WebFetch handler.
func NewHandler(options HandlerOptions) (*Handler, error) {
	if options.Reader == nil {
		return nil, errors.New("reader is required")
	}
	if options.Timeout <= 0 {
		options.Timeout = 20 * time.Second
	}
	if options.MaxTimeout <= 0 {
		options.MaxTimeout = 60 * time.Second
	}
	return &Handler{reader: options.Reader, timeout: options.Timeout, maxTimeout: options.MaxTimeout, cacheBypass: options.CacheBypass, logURLQuery: options.LogURLQuery}, nil
}

// Fetch validates one request and writes the stable WebFetch response contract.
func (handler *Handler) Fetch(ctx *gin.Context) {
	var payload dto.FetchRequest
	if err := dto.Decode(ctx.Request.Body, &payload); err != nil {
		writeProblem(ctx, http.StatusBadRequest, "invalid_request", "Invalid request", "The request body is invalid.", false, "")
		return
	}
	requestTimeout, err := dto.ParseTimeout(payload.Timeout, handler.timeout, handler.maxTimeout)
	if err != nil {
		writeProblem(ctx, http.StatusBadRequest, "invalid_request", "Invalid timeout", err.Error(), false, "timeout")
		return
	}
	request := domain.ReadRequest{URL: payload.URL, Format: payload.Output.Format, MaxChars: payload.Output.MaxChars, Refresh: handler.cacheBypass, RequestID: runtime.RequestID(ctx)}
	requestContext, cancel := context.WithTimeout(ctx.Request.Context(), requestTimeout)
	defer cancel()
	response, err := handler.reader.Read(requestContext, request)
	if err != nil {
		logs.WarnContextw(ctx, "read failed", "request_id", request.RequestID, "url", redactURL(payload.URL, handler.logURLQuery), "error", err)
		writeReadProblem(ctx, err)
		return
	}
	logs.InfoContextw(ctx, "read completed", "request_id", request.RequestID, "url", redactURL(payload.URL, handler.logURLQuery), "transport", response.Meta.Transport, "cached", response.Meta.Cached, "took_ms", response.Meta.TookMS)
	ctx.JSON(http.StatusOK, dto.FetchResponse{
		RequestID: request.RequestID,
		Document: dto.Document{
			URL: response.URL, FinalURL: response.FinalURL, Title: response.Title, Author: response.Author,
			PublishedAt: response.PublishedAt, Language: response.Language, SourceType: response.SourceType,
			ContentType: response.ContentType, StatusCode: response.StatusCode, Content: response.Content,
			Format: response.ContentFormat, RetrievedAt: time.Now().UTC(),
		},
		Meta: dto.Meta{
			Cached: response.Meta.Cached, Transport: response.Meta.Transport, Truncated: response.Truncated,
			ContentLength: response.ContentLength, CacheAgeSeconds: response.Meta.CacheAgeSeconds, TookMS: response.Meta.TookMS,
		},
		Warnings: response.Warnings,
		Usage:    dto.Usage{Units: 1},
	})
}

func writeReadProblem(ctx *gin.Context, err error) {
	var readErr *domain.ReadError
	if !errors.As(err, &readErr) {
		writeProblem(ctx, http.StatusBadGateway, "upstream_failed", "WebFetch failed", "The target could not be fetched.", true, "")
		return
	}
	status, code, retryable := http.StatusBadGateway, string(readErr.Code), readErr.Retryable
	switch readErr.Code {
	case domain.ErrInvalidRequest:
		status = http.StatusBadRequest
	case domain.ErrUnsafeURL:
		status = http.StatusForbidden
	case domain.ErrUnsupportedContentType:
		status = http.StatusUnsupportedMediaType
	case domain.ErrContentTooLarge:
		status = http.StatusRequestEntityTooLarge
	case domain.ErrExtractionFailed:
		status = http.StatusUnprocessableEntity
	case domain.ErrFetchTimeout:
		status, code = http.StatusGatewayTimeout, "deadline_exceeded"
	}
	if code == "captcha_required" {
		status, retryable = http.StatusUnprocessableEntity, false
	}
	writeProblem(ctx, status, code, "WebFetch failed", readErr.Message, retryable, "")
}

func writeProblem(ctx *gin.Context, status int, code, title, detail string, retryable bool, parameter string) {
	ctx.Header("Content-Type", "application/problem+json")
	ctx.AbortWithStatusJSON(status, dto.Problem{Type: "https://api.example.com/problems/" + strings.ReplaceAll(code, "_", "-"), Title: title, Status: status, Code: code, Detail: detail, RequestID: runtime.RequestID(ctx), Retryable: retryable, Parameter: parameter})
}

func redactURL(rawURL string, includeQuery bool) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "invalid-url"
	}
	parsed.User = nil
	parsed.Fragment = ""
	if !includeQuery {
		parsed.RawQuery = ""
	}
	return parsed.String()
}
