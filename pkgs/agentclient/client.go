package agentclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

// ChatClient 聊天客户端
type ChatClient struct {
	BaseURL    string
	HTTPClient *http.Client
	APIKey     string
}

// NewChatClient 创建新客户端
func NewChatClient(client *http.Client, baseURL string, apiKey string) *ChatClient {
	if client == nil {
		client = &http.Client{}
	}
	return &ChatClient{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		HTTPClient: client,
	}
}

func NewChatClientWithConfig(client *http.Client, cfg config.LLMModelConfig) *ChatClient {
	if client == nil {
		client = &http.Client{}
	}
	return &ChatClient{
		BaseURL:    cfg.BaseURL,
		APIKey:     cfg.APIKEY,
		HTTPClient: client,
	}
}

// SendChat 发送非流式请求，支持 context，用于取消请求或设置超时
func (c *ChatClient) SendChat(ctx context.Context, reqBody *ChatRequestBody) (*ChatResponseBody, error) {
	if reqBody == nil {
		return nil, errors.New("request body is nil")
	}

	reqBody.Stream = false
	reqBody = reqBody.Pretreat()

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var finalResp ChatResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&finalResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &finalResp, nil
}

type ChatStreamReader struct {
	reader      *bufio.Reader
	currentLine string
	body        io.ReadCloser
}

func (r *ChatStreamReader) ReadNext() (*ChatStreamResponseBody, error) {
	for {
		line, err := r.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		r.currentLine = line

		if line == "data: [DONE]" {
			return nil, io.EOF
		}
		line = strings.TrimPrefix(line, "data: ")

		var resp ChatStreamResponseBody
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			return nil, fmt.Errorf("unmarshal failed: %w, line: %q", err, line)
		}
		return &resp, nil
	}
}

func (r *ChatStreamReader) CurrentLine() string {
	return r.currentLine
}

func (r *ChatStreamReader) Close() error {
	if r.body != nil {
		return r.body.Close()
	}
	return nil
}

// SendChatStreamWithCallback 启动流式请求并通过回调处理每一段响应
func (c *ChatClient) SendChatStreamWithCallback(
	ctx context.Context,
	reqBody *ChatRequestBody,
	onMessage func(*ChatStreamResponseBody) error,
) error {
	if reqBody == nil {
		return errors.New("request body is nil")
	}

	reqBody.Stream = true
	reqBody = reqBody.Pretreat()

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return fmt.Errorf("read error: %w", err)
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if line == "data: [DONE]" {
				break
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			line = strings.TrimPrefix(line, "data: ")

			var respChunk ChatStreamResponseBody
			if err := json.Unmarshal([]byte(line), &respChunk); err != nil {
				return fmt.Errorf("failed to unmarshal line: %w, line: %q", err, line)
			}

			if onMessage != nil {
				if err := onMessage(&respChunk); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type DefaultLLMConfig struct {
	BaseURL   string
	ModelName string
}

func GetLLMConfig(ctx context.Context, group, key string) (config.LLMModelConfig, error) {
	cfg := &config.LLMModelConfig{}

	err := settings.GetYaml(group, key, cfg)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to get llm config(%s/%s): %v", group, key, err)
		return config.LLMModelConfig{}, fmt.Errorf("failed to get llm config: %w", err)
	}

	return *cfg, nil
}
