package baidu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"github.com/ygpkg/yg-go/logs"
)

const (
	defaultToolName = "baidu_web_search"
	defaultToolDesc = "Search for information on the internet using Baidu. Useful for finding up-to-date information, news, and more."
	defaultBaseURL  = "https://qianfan.baidubce.com/v2/ai_search/web_search"
)

type Config struct {
	ToolName   string `json:"tool_name"`
	ToolDesc   string `json:"tool_desc"`
	MaxResults int    `json:"max_results"`
	ApiKey     string `json:"api_key"`
	// Default: 30 seconds
	Timeout    time.Duration `json:"timeout"`
	HTTPClient *http.Client  `json:"http_client"`
}

type baiduSearchTool struct {
	apiKey     string
	client     *http.Client
	maxResults int
}

func NewBaiduWebSearch(ctx context.Context, config *Config) (tool.InvokableTool, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if config.ApiKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	toolName := config.ToolName
	if toolName == "" {
		toolName = defaultToolName
	}
	toolDesc := config.ToolDesc
	if toolDesc == "" {
		toolDesc = defaultToolDesc
	}

	t, err := buildClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create baidu client: %w", err)
	}

	return utils.NewTool(
		&schema.ToolInfo{
			Name:        toolName,
			Desc:        toolDesc,
			ParamsOneOf: schema.NewParamsOneOfByJSONSchema(getSchema()),
		},
		t.Search,
	), nil
}

func getSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:     string(schema.Object),
		Required: []string{"query"},
		Properties: orderedmap.New[string, *jsonschema.Schema](
			orderedmap.WithInitialData[string, *jsonschema.Schema](
				orderedmap.Pair[string, *jsonschema.Schema]{
					Key: "query",
					Value: &jsonschema.Schema{
						Type:        string(schema.String),
						Description: "The search query keywords",
					},
				},
				orderedmap.Pair[string, *jsonschema.Schema]{
					Key: "time_range",
					Value: &jsonschema.Schema{
						Type:        string(schema.String),
						Description: "根据网页发布时间进行筛选（可选）",
						Default:     "",
						OneOf: []*jsonschema.Schema{
							{
								Type:        string(schema.String),
								Enum:        []any{"week"},
								Description: "最近7天",
							},
							{
								Type:        string(schema.String),
								Enum:        []any{"month"},
								Description: "最近30天",
							},
							{
								Type:        string(schema.String),
								Enum:        []any{"semiyear"},
								Description: "最近180天",
							},
							{
								Type:        string(schema.String),
								Enum:        []any{"year"},
								Description: "最近365天",
							},
						},
					},
				},
			),
		),
	}
}

type SearchRequest struct {
	Query     string `json:"query"`
	TimeRange string `json:"time_range,omitempty"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

type SearchResult struct {
	Title       string  `json:"title"`
	Url         string  `json:"url"`
	Content     string  `json:"content"`
	RerankScore float64 `json:"rerank_score"`
}

func buildClient(_ context.Context, config *Config) (baiduSearchTool, error) {
	maxResults := config.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 50 {
		maxResults = 50
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	httpCli := config.HTTPClient
	if httpCli == nil {
		httpCli = &http.Client{
			Timeout: timeout,
		}
	}

	return baiduSearchTool{
		apiKey:     config.ApiKey,
		client:     httpCli,
		maxResults: maxResults,
	}, nil
}

func (t *baiduSearchTool) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	apiReq := apiRequest{
		Messages: []apiMessage{
			{
				Role:    "user",
				Content: req.Query,
			},
		},
		Edition: "lite",
		ResourceTypeFilter: []apiResource{
			{
				Type: "web",
				TopK: t.maxResults,
			},
		},
	}
	if req.TimeRange != "" {
		apiReq.SearchRecencyFilter = req.TimeRange
	}

	logs.InfoContextf(ctx, "baidu search request: %s", req.Query)

	reqBody, err := json.Marshal(apiReq)
	if err != nil {
		logs.ErrorContextf(ctx, "baidu search failed to marshal request: %s", err.Error())
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, defaultBaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		logs.ErrorContextf(ctx, "baidu search failed to create request: %s", err.Error())
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+t.apiKey)

	resp, err := t.client.Do(httpReq)
	if err != nil {
		logs.ErrorContextf(ctx, "baidu search failed to send request: %s", err.Error())
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "baidu search failed to send request: status=%d body=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("baidu api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Read body to handle potential different structures
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "baidu search failed to read response body: %s", err.Error())
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		logs.ErrorContextf(ctx, "baidu search failed to unmarshal response: %s", err.Error())
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if apiResp.Code != "" {
		logs.ErrorContextf(ctx, "baidu search requestId: %s api error: code=%s message=%s", apiResp.RequestId, apiResp.Code, apiResp.Message)
		return nil, fmt.Errorf("search err: message=%s", apiResp.Message)
	}

	results := apiResp.References

	return &SearchResponse{
		Results: results,
	}, nil
}

// API Request/Response structures
type apiRequest struct {
	Messages            []apiMessage  `json:"messages"`
	Edition             string        `json:"edition"`
	ResourceTypeFilter  []apiResource `json:"resource_type_filter"`
	SearchRecencyFilter string        `json:"search_recency_filter,omitempty"`
}

type apiMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type apiResource struct {
	Type string `json:"type"`
	TopK int    `json:"top_k"`
}

type apiResponse struct {
	RequestId  string         `json:"request_id"`
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	References []SearchResult `json:"references"`
}
