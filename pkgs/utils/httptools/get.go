package httptools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// maxErrorResponseBytes limits diagnostic response content to 1KB to avoid loading an unbounded error body.
const maxErrorResponseBytes = int64(1024)

// Get sends an HTTP GET request and returns a successful response for the caller to process and close.
func Get(ctx context.Context, rawURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build GET request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send GET request: %w", err)
	}
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return resp, nil
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxErrorResponseBytes))
	if readErr != nil {
		return nil, fmt.Errorf("GET request returned %s and reading error response failed: %w", resp.Status, readErr)
	}
	message := strings.TrimSpace(string(body))
	if message == "" {
		return nil, fmt.Errorf("GET request returned %s", resp.Status)
	}
	return nil, fmt.Errorf("GET request returned %s: %s", resp.Status, message)
}
