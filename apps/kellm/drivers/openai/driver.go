package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/insmtx/corekg/apps/kellm/drivers"
	"github.com/insmtx/corekg/apps/kellm/models/kellmtype"
)

type Driver struct{}

func init() {
	drivers.RegisterDriver(&Driver{})
}

func (d *Driver) Type() kellmtype.ModelType {
	return kellmtype.ModelTypeOpenAI
}

func (d *Driver) ChatCompletions(ctx context.Context, chatCtx *drivers.ChatContext) (*drivers.ProxyResult, error) {
	body, err := json.Marshal(chatCtx.UpstreamRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ChatCompletionsURL(chatCtx.ModelConfig.BaseURL), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+chatCtx.Token)
	for key, value := range chatCtx.ModelConfig.Headers {
		if strings.TrimSpace(key) == "" {
			continue
		}
		req.Header.Set(key, value)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	if chatCtx.UpstreamRequest.Stream {
		return &drivers.ProxyResult{
			StatusCode: resp.StatusCode,
			Header:     cloneHeader(resp.Header),
			BodyReader: resp.Body,
			Stream:     true,
		}, nil
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &drivers.ProxyResult{
		StatusCode: resp.StatusCode,
		Header:     cloneHeader(resp.Header),
		Body:       respBody,
	}, nil
}

func ChatCompletionsURL(baseURL string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return ""
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	if strings.HasSuffix(baseURL, "/chat/completions") {
		return baseURL
	}
	return baseURL + "/chat/completions"
}

func cloneHeader(header http.Header) http.Header {
	dst := make(http.Header, len(header))
	for key, values := range header {
		copied := make([]string, len(values))
		copy(copied, values)
		dst[key] = copied
	}
	return dst
}
