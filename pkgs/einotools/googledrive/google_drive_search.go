package googledrive

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/insmtx/corekg/pkgs/connectors/tokenmgr"
	oauthUtils "github.com/insmtx/corekg/pkgs/utils/oauth"
	"github.com/ygpkg/yg-go/logs"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const defaultMaxResultNum = 20
const maxResultNum = 500

type Config struct {
	ToolName   string `json:"tool_name"`   // default: google_drive_search
	ToolDesc   string `json:"tool_desc"`   // default: "Search files in Google Drive by name, type, content, or date range"
	MaxResults int64  `json:"max_results"` // default: 20
}

type googleDriveSearch struct {
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
		toolName = "google_drive_search"
	}
	if toolDesc == "" {
		toolDesc = "Search files in Google Drive by name, type, content, or date range"
	}

	gds := &googleDriveSearch{
		conf: conf,
	}

	tl, err := utils.InferTool(toolName, toolDesc, gds.search, utils.WithMarshalOutput(gds.marshalOutput))
	if err != nil {
		return nil, err
	}
	return tl, nil
}

func (gds *googleDriveSearch) search(ctx context.Context, req *SearchRequest) (*drive.FileList, error) {
	service, err := gds.getService(ctx, req.Uin)
	if err != nil {
		logs.WarnContextf(ctx, "google drive search get service failed, uin: %d, err: %v", req.Uin, err)
		return &drive.FileList{
			Files: make([]*drive.File, 0),
		}, nil
	}

	maxResults := gds.conf.MaxResults
	if req.MaxResults > 0 {
		maxResults = req.MaxResults
	}
	if maxResults > maxResultNum {
		maxResults = maxResultNum
	}

	// Build search query
	query := buildDriveQuery(req)

	// Execute search
	files, err := service.Files.List().
		Q(query).
		PageSize(maxResults).
		Fields("files(id,name,mimeType,size,createdTime,modifiedTime,parents,webViewLink,webContentLink,owners,shared,trashed,hasThumbnail,thumbnailLink),nextPageToken").
		Do()

	if err != nil {
		logs.WarnContextf(ctx, "google drive search list files failed, uin: %d, err: %v", req.Uin, err)
		return nil, fmt.Errorf("获取文件列表失败: %v", err)
	}

	return files, nil
}

func (gds *googleDriveSearch) getService(ctx context.Context, uin uint) (*drive.Service, error) {
	externalToken, ok := tokenmgr.GetToken(ctx, uin, tokenmgr.PlatformGoogleDrive)
	if !ok {
		logs.WarnContextf(ctx, "google drive search get token failed, uin: %d", uin)
		return nil, fmt.Errorf("获取token失败")
	}

	client := oauthUtils.CreateOAuth2ClientWithProxy(externalToken.AccessToken)

	service, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		logs.WarnContextf(ctx, "google drive search create service failed, uin: %d, err: %v", uin, err)
		return nil, fmt.Errorf("创建Google Drive服务失败: %v", err)
	}

	return service, nil
}

func (gds *googleDriveSearch) marshalOutput(ctx context.Context, output any) (string, error) {
	fileList, ok := output.(*drive.FileList)
	if !ok {
		logs.WarnContextf(ctx, "google drive search marshal output failed, expect %T but given %T", fileList, output)
		return "", fmt.Errorf("unexpected google drive file list response, expect %T but given %T", fileList, output)
	}

	var files []*FileInfo
	for _, file := range fileList.Files {
		f := parseDriveFile(file)
		files = append(files, &f)
	}

	searchResult := &SearchResult{
		Files: files,
	}

	jsonData, err := json.Marshal(searchResult)
	if err != nil {
		logs.WarnContextf(ctx, "google drive search marshal output failed, err: %v", err)
		return "", err
	}
	return string(jsonData), nil
}

func parseDriveFile(file *drive.File) FileInfo {
	// Extract owner information
	owners := make([]string, 0, len(file.Owners))
	for _, owner := range file.Owners {
		if owner.EmailAddress != "" {
			owners = append(owners, owner.EmailAddress)
		} else if owner.DisplayName != "" {
			owners = append(owners, owner.DisplayName)
		}
	}

	return FileInfo{
		ID:             file.Id,
		Name:           file.Name,
		MimeType:       file.MimeType,
		Size:           file.Size,
		CreatedTime:    file.CreatedTime,
		ModifiedTime:   file.ModifiedTime,
		ParentIDs:      file.Parents,
		WebLink:        file.WebViewLink,
		WebContentLink: file.WebContentLink,
		Owners:         owners,
		Shared:         file.Shared,
		Trashed:        file.Trashed,
		HasThumbnail:   file.HasThumbnail,
		ThumbnailLink:  file.ThumbnailLink,
	}
}

func escapeQueryValue(val string) string {
	// 先转义反斜杠，再转义单引号
	val = strings.ReplaceAll(val, `\`, `\\`)
	val = strings.ReplaceAll(val, `'`, `\'`)
	return val
}

func buildDriveQuery(req *SearchRequest) string {
	var parts []string

	// 关键词 (fullText contains)
	if len(req.Queries) > 0 {
		var queryParts []string
		for _, q := range req.Queries {
			escaped := escapeQueryValue(q)
			queryParts = append(queryParts, fmt.Sprintf("fullText contains '%s'", escaped))
		}
		if len(queryParts) == 1 {
			parts = append(parts, queryParts[0])
		} else {
			parts = append(parts, "("+strings.Join(queryParts, " OR ")+")")
		}
	}

	// 创建时间范围
	if req.CreatedAfter != nil {
		parts = append(parts, fmt.Sprintf("createdTime > '%s'", req.CreatedAfter.Format(time.RFC3339)))
	}
	if req.CreatedBefore != nil {
		parts = append(parts, fmt.Sprintf("createdTime < '%s'", req.CreatedBefore.Format(time.RFC3339)))
	}

	// 修改时间范围
	if req.ModifiedAfter != nil {
		parts = append(parts, fmt.Sprintf("modifiedTime > '%s'", req.ModifiedAfter.Format(time.RFC3339)))
	}
	if req.ModifiedBefore != nil {
		parts = append(parts, fmt.Sprintf("modifiedTime < '%s'", req.ModifiedBefore.Format(time.RFC3339)))
	}

	return strings.Join(parts, " and ") + " and trashed = false"
}

// SearchRequest represents a search request
type SearchRequest struct {
	Uin uint `json:"uin" jsonschema:"required,description=System internal unique user identifier (UIN) used to lookup Drive authorization info"`

	Queries []string `json:"queries,omitempty" jsonschema:"required,description=List of query keywords to search files"`

	// 创建时间范围
	CreatedAfter  *time.Time `json:"created_after,omitempty" jsonschema:"description=Filter files created after this time (RFC3339 format)"`
	CreatedBefore *time.Time `json:"created_before,omitempty" jsonschema:"description=Filter files created before this time (RFC3339 format)"`

	// 修改时间范围
	ModifiedAfter  *time.Time `json:"modified_after,omitempty" jsonschema:"description=Filter files modified after this time (RFC3339 format)"`
	ModifiedBefore *time.Time `json:"modified_before,omitempty" jsonschema:"description=Filter files modified before this time (RFC3339 format)"`

	MaxResults int64 `json:"max_results,omitempty" jsonschema:"description=Maximum number of files to return, default 20"`
}

func (r *SearchRequest) String() string {
	jsonData, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(jsonData)
}

// FileInfo represents file information from Google Drive
type FileInfo struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	MimeType       string   `json:"mimeType"`
	Size           int64    `json:"size"`
	CreatedTime    string   `json:"createdTime"`
	ModifiedTime   string   `json:"modifiedTime"`
	ParentIDs      []string `json:"parentTds"`
	WebLink        string   `json:"webLink"`
	WebContentLink string   `json:"webContentLink"`
	Owners         []string `json:"owners"`
	Shared         bool     `json:"shared"`
	Trashed        bool     `json:"trashed"`
	HasThumbnail   bool     `json:"hasThumbnail"`
	ThumbnailLink  string   `json:"thumbnailLink"`
}

// SearchResult contains the search results and pagination information
type SearchResult struct {
	Files []*FileInfo `json:"files"`
}
