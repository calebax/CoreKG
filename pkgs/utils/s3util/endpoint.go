package s3util

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ResolveS3Endpoint returns an absolute scheme://host endpoint for S3 public access.
// It prefers an absolute Referer and falls back to forwarded/request host metadata.
func ResolveS3Endpoint(referer string, req *http.Request) (string, error) {
	if endpoint := endpointFromURL(referer); endpoint != "" {
		return endpoint, nil
	}
	if req == nil {
		return "", fmt.Errorf("request is nil")
	}

	host := firstHeaderValue(req.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(req.Host)
	}
	if host == "" {
		return "", fmt.Errorf("request host is empty")
	}

	scheme := firstHeaderValue(req.Header.Get("X-Forwarded-Proto"))
	if scheme == "" && req.URL != nil {
		scheme = strings.TrimSpace(req.URL.Scheme)
	}
	if scheme == "" {
		if req.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	return fmt.Sprintf("%s://%s", scheme, host), nil
}

func endpointFromURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return ""
	}

	return fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
}

func firstHeaderValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if idx := strings.Index(value, ","); idx >= 0 {
		value = value[:idx]
	}

	return strings.TrimSpace(value)
}
