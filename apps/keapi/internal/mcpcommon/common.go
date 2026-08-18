package mcpcommon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/ygpkg/yg-go/logs"
)

type contextKeyType string

const ctxKeyRawAPIKey contextKeyType = "mcp_raw_api_key"

func RawAPIKeyFromContext(ctx context.Context) string {
	v := ctx.Value(ctxKeyRawAPIKey)
	if v == nil {
		return ""
	}
	return v.(string)
}

func ContextWithRawAPIKey(ctx context.Context, apiKey string) context.Context {
	return context.WithValue(ctx, ctxKeyRawAPIKey, apiKey)
}

type InternalClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewInternalClient(baseURL string) *InternalClient {
	baseURL = strings.TrimRight(baseURL, "/") + "/v3/"
	return &InternalClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"response"`
}

func (c *InternalClient) CallAPI(ctx context.Context, apiKey string, action string, reqBody any) (json.RawMessage, error) {
	wrappedReq := struct {
		Request any `json:"request"`
	}{
		Request: reqBody,
	}

	bodyBytes, err := json.Marshal(wrappedReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request body failed: %w", err)
	}

	url := c.BaseURL + action
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call internal API %s failed: %w", action, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("internal API %s returned status %d: %s", action, resp.StatusCode, string(respBody))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w, body: %s", err, string(respBody))
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("internal API %s error code %d: %s", action, apiResp.Code, apiResp.Message)
	}

	logs.InfoContextf(ctx, "[mcp_internal_client] call %s success", action)
	return apiResp.Data, nil
}

func (c *InternalClient) UploadFile(ctx context.Context, apiKey string, forestID string, parentID string, fileName string, fileData []byte) (json.RawMessage, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, fmt.Errorf("create form file failed: %w", err)
	}
	if _, err := part.Write(fileData); err != nil {
		return nil, fmt.Errorf("write file content failed: %w", err)
	}

	if err := writer.WriteField("forest_id", forestID); err != nil {
		return nil, fmt.Errorf("write forest_id field failed: %w", err)
	}
	if parentID != "" {
		if err := writer.WriteField("parent_id", parentID); err != nil {
			return nil, fmt.Errorf("write parent_id field failed: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer failed: %w", err)
	}

	url := c.BaseURL + "keapi.UploadFile"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("create upload request failed: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload file failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read upload response failed: %w", err)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal upload response failed: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("upload file error code %d: %s", apiResp.Code, apiResp.Message)
	}

	return apiResp.Data, nil
}