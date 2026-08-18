package confluence

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	confulence "github.com/ctreminiom/go-atlassian/confluence"
	models "github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/insmtx/corekg/pkgs/connectors/tokenmgr"
	oauthUtils "github.com/insmtx/corekg/pkgs/utils/oauth"
	"github.com/ygpkg/yg-go/logs"
)

const defaultMaxResultNum = 20
const maxResultNum = 500

const site = "https://api.atlassian.com/ex/confluence/%s"

type Config struct {
	ToolName   string `json:"tool_name"`   // default: confluence_search
	ToolDesc   string `json:"tool_desc"`   // default: "Search pages and content in Confluence by keyword, space, author, or date range"
	MaxResults int    `json:"max_results"` // default: 20
}

type confluenceSearch struct {
	conf *Config
}

func NewTool(ctx context.Context, conf *Config) (tool.InvokableTool, error) {
	toolName := conf.ToolName
	toolDesc := conf.ToolDesc
	maxResults := conf.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResultNum
	}
	if toolName == "" {
		toolName = "confluence_search"
	}
	if toolDesc == "" {
		toolDesc = "Search pages and content in Confluence by keyword, space, author, or date range"
	}

	cs := &confluenceSearch{
		conf: conf,
	}

	tl, err := utils.InferTool(toolName, toolDesc, cs.search, utils.WithMarshalOutput(cs.marshalOutput))
	if err != nil {
		return nil, err
	}
	return tl, nil
}

func (cs *confluenceSearch) search(ctx context.Context, req *SearchRequest) (*models.SearchPageScheme, error) {
	client, err := cs.getClient(ctx, req.Uin)
	if err != nil {
		logs.WarnContextf(ctx, "confluence search get client failed, uin: %d, err: %v", req.Uin, err)
		return &models.SearchPageScheme{}, nil
	}

	maxResults := cs.conf.MaxResults
	if req.MaxResults > 0 {
		maxResults = req.MaxResults
	}
	if maxResults > maxResultNum {
		maxResults = maxResultNum
	}

	// Build search query using CQL (Confluence Query Language)
	cqlQuery := buildCQL(req)

	// Execute search using Confluence search API
	contents, response, err := client.Search.Content(ctx, cqlQuery, nil)

	if err != nil {
		jsonResponse, jerr := json.Marshal(struct {
			Code     int    `json:"code"`
			Endpoint string `json:"endpoint"`
			Method   string `json:"method"`
			Body     string `json:"body"`
		}{
			Code:     response.Code,
			Endpoint: response.Endpoint,
			Method:   response.Method,
			Body:     response.Bytes.String(),
		})
		var responseStr string
		if jerr != nil {
			responseStr = "<failed to marshal response>"
		} else {
			responseStr = string(jsonResponse)
		}

		logs.WarnContextf(ctx,
			"confluence search failed, uin: %d, err: %v, response: %s",
			req.Uin,
			err,
			responseStr,
		)
		return nil, fmt.Errorf("搜索Confluence失败: %v", err)
	}

	return contents, nil
}

func (cs *confluenceSearch) getClient(ctx context.Context, uin uint) (*confulence.Client, error) {
	tokenInfo, ok := tokenmgr.GetToken(ctx, uin, tokenmgr.PlatformConfluence)
	if !ok {
		return nil, fmt.Errorf("未找到Confluence授权信息，请先绑定Confluence账号")
	}

	// Create HTTP client with OAuth token
	client := oauthUtils.CreateOAuth2ClientWithProxy(tokenInfo.AccessToken)

	// Create Confluence client
	siteurl, err := buildConfluenceBaseURL(client)
	if err != nil {
		logs.WarnContextf(ctx, "构建 Confluence 基础 URL 失败, err=%v", err)
		return nil, fmt.Errorf("构建 Confluence 基础 URL 失败: %v", err)
	}
	confluenceClient, err := confulence.New(client, siteurl)
	if err != nil {
		logs.WarnContextf(ctx, "创建 Confluence 客户端失败，siteURL=%s, err=%v", siteurl, err)
		return nil, fmt.Errorf("创建Confluence客户端失败: %v", err)
	}

	return confluenceClient, nil
}

func buildConfluenceBaseURL(client *http.Client) (string, error) {
	const url = "https://api.atlassian.com/oauth/token/accessible-resources"

	// 发起 GET 请求
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 读取响应 body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body failed: %w", err)
	}

	// 解析 JSON
	var resources []AccessibleResource
	if err := json.Unmarshal(bodyBytes, &resources); err != nil {
		return "", fmt.Errorf("json unmarshal failed: %w", err)
	}

	if len(resources) == 0 {
		return "", fmt.Errorf("no accessible resources found")
	}
	return fmt.Sprintf(site, resources[0].ID), nil
}

type AccessibleResource struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	AvatarURL string   `json:"avatarUrl"`
}

func buildCQL(req *SearchRequest) string {
	var clauses []string

	// 关键词查询
	if len(req.Queries) > 0 {
		var q []string
		for _, keyword := range req.Queries {
			keyword = strings.TrimSpace(keyword)
			if keyword != "" {
				// 匹配标题或正文
				q = append(q, fmt.Sprintf(`title ~ "%s" OR text ~ "%s"`, keyword, keyword))
			}
		}
		if len(q) > 0 {
			clauses = append(clauses, "("+strings.Join(q, " OR ")+")")
		}
	}

	// 作者过滤
	if req.Author != "" {
		clauses = append(clauses, fmt.Sprintf(`creator = "%s"`, req.Author))
	}

	// 内容类型过滤
	if req.ContentType != "" {
		clauses = append(clauses, fmt.Sprintf(`type = "%s"`, req.ContentType))
	}

	// 时间范围过滤
	if req.StartDate != nil {
		clauses = append(clauses, fmt.Sprintf(`created >= "%s"`, req.StartDate.Format(time.RFC3339)))
	}
	if req.EndDate != nil {
		clauses = append(clauses, fmt.Sprintf(`created <= "%s"`, req.EndDate.Format(time.RFC3339)))
	}

	// 拼接所有条件
	cql := strings.Join(clauses, " AND ")
	return cql
}

func (cs *confluenceSearch) marshalOutput(_ context.Context, output any) (string, error) {
	contents, ok := output.(*models.SearchPageScheme)
	if !ok {
		return "", fmt.Errorf("unexpected confluence search response, expect %T but given %T", contents, output)
	}

	results := SearchResults{}

	if contents == nil || len(contents.Results) == 0 {
		return results.String(), nil
	}

	for _, r := range contents.Results {
		result := parseSearchResult(r)
		results.Results = append(results.Results, result)
	}

	results.Total = contents.TotalSize
	return results.String(), nil
}

func parseSearchResult(r *models.SearchResultScheme) *SearchResult {
	if r == nil {
		return nil
	}

	var (
		id, title, typ, status, webLink                string
		parentTitle, parentURL, globalTitle, globalURL string
	)

	// Content 判空
	if r.Content != nil {
		id = r.Content.ID
		title = r.Content.Title
		typ = r.Content.Type
		status = r.Content.Status

		if r.Content.Links != nil {
			webLink = r.Content.Links.Self
		}
	}

	if r.ResultParentContainer != nil {
		parentTitle = r.ResultParentContainer.Title
		parentURL = r.ResultParentContainer.DisplayURL
	}

	if r.ResultGlobalContainer != nil {
		globalTitle = r.ResultGlobalContainer.Title
		globalURL = r.ResultGlobalContainer.DisplayURL
	}

	return &SearchResult{
		ID:                        id,
		Title:                     title,
		Type:                      typ,
		Status:                    status,
		EntityType:                r.EntityType,
		WebLink:                   webLink,
		Excerpt:                   r.Excerpt,
		LastModified:              r.LastModified,
		FriendlyLastModified:      r.FriendlyLastModified,
		ResultParentContainerName: parentTitle,
		ResultParentContainerURL:  parentURL,
		ResultGlobalContainerName: globalTitle,
		ResultGlobalContainerURL:  globalURL,
		Score:                     r.Score,
	}
}

type SearchRequest struct {
	Uin uint `json:"uin" jsonschema:"required,description=System internal unique user identifier (UIN) used to lookup Confluence authorization info"`

	Queries []string `json:"queries,omitempty" jsonschema:"description=List of query keywords to search pages and content"`

	Author string `json:"author,omitempty" jsonschema:"description=Filter content created by this author username"`

	ContentType string `json:"content_type,omitempty" jsonschema:"description=Filter by content type: page, blogpost, etc."`

	StartDate *time.Time `json:"start_date,omitempty" jsonschema:"description=Start date of content creation to search, in RFC3339 format"`

	EndDate *time.Time `json:"end_date,omitempty" jsonschema:"description=End date of content creation to search, in RFC3339 format"`

	MaxResults int `json:"max_results,omitempty" jsonschema:"description=Maximum number of results to return, default 20"`
}

func (r *SearchRequest) String() string {
	jsonData, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(jsonData)
}

func (rs *SearchResults) String() string {
	jsonData, err := json.Marshal(rs)
	if err != nil {
		return ""
	}
	return string(jsonData)
}

type SearchResults struct {
	Results []*SearchResult `json:"results"`
	Total   int             `json:"total,omitempty"`
}

type SearchResult struct {
	ID         string `json:"id"`                // 页面ID
	Title      string `json:"title"`             // 页面标题
	Type       string `json:"type"`              // 内容类型，如 "page"
	Status     string `json:"status"`            // 页面状态，如 "current"
	EntityType string `json:"entityType"`        // 搜索结果类型
	WebLink    string `json:"webLink,omitempty"` // 浏览器访问链接
	Excerpt    string `json:"excerpt,omitempty"` // 匹配摘要
	// Space                     *Space    `json:"space,omitempty"`                   // 页面所属空间
	LastModified              string  `json:"lastModified,omitempty"`              // 最后修改时间（ISO 8601）
	FriendlyLastModified      string  `json:"friendlyLastModified,omitempty"`      // 人性化修改时间
	ResultParentContainerName string  `json:"resultParentContainerName,omitempty"` // 父容器名称
	ResultParentContainerURL  string  `json:"resultParentContainerURL,omitempty"`  // 父容器URL
	ResultGlobalContainerName string  `json:"resultGlobalContainerName,omitempty"` // 全局容器名称（空间名称）
	ResultGlobalContainerURL  string  `json:"resultGlobalContainerURL,omitempty"`  // 全局容器URL
	Score                     float64 `json:"score,omitempty"`                     // 搜索匹配评分
	// Breadcrumbs               []Breadcrumb `json:"breadcrumbs,omitempty"`           // 面包屑信息
}

type Space struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type User struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}
