package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/insmtx/corekg/pkgs/connectors/tokenmgr"
	oauthUtils "github.com/insmtx/corekg/pkgs/utils/oauth"
	"github.com/slack-go/slack"
	"github.com/ygpkg/yg-go/logs"
)

const defaultMaxResultNum = 20
const maxResultNum = 500

type Config struct {
	ToolName   string `json:"tool_name"`   // default: slack_search
	ToolDesc   string `json:"tool_desc"`   // default: "Search messages and files in Slack by keyword, user, channel, or date range"
	MaxResults int    `json:"max_results"` // default: 20
}

type slackSearch struct {
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
		toolName = "slack_search"
	}
	if toolDesc == "" {
		toolDesc = "Search messages and files in Slack by keyword, user, channel, or date range"
	}

	ss := &slackSearch{
		conf: conf,
	}

	tl, err := utils.InferTool(toolName, toolDesc, ss.search, utils.WithMarshalOutput(ss.marshalOutput))
	if err != nil {
		return nil, err
	}
	return tl, nil
}

func (ss *slackSearch) search(ctx context.Context, req *SearchRequest) (*SearchResult, error) {
	client, err := ss.getClient(ctx, req.Uin)
	if err != nil {
		logs.WarnContextf(ctx, "slack search get client failed, uin: %d, err: %v", req.Uin, err)
		return &SearchResult{}, nil
	}

	maxResults := ss.conf.MaxResults
	if req.MaxResults > 0 {
		maxResults = req.MaxResults
	}
	if maxResults > maxResultNum {
		maxResults = maxResultNum
	}

	// Build search query
	query := buildSlackQuery(req)

	// Use search.all to search both messages and files
	searchParams := slack.NewSearchParameters()
	searchParams.Count = maxResults

	// Search using search.all API
	searchMessageResults, searchFileResults, err := client.SearchContext(ctx, query, searchParams)
	if err != nil {
		logs.WarnContextf(ctx, "slack search search context failed, uin: %d, err: %v", req.Uin, err)
		return nil, fmt.Errorf("搜索Slack失败: %v", err)
	}
	searchResults := &SearchResult{
		Total:    searchMessageResults.Total + searchFileResults.Total,
		Messages: searchMessageResults,
		Files:    searchFileResults,
	}

	return searchResults, nil
}

func (ss *slackSearch) getClient(ctx context.Context, uin uint) (*slack.Client, error) {
	tokenInfo, ok := tokenmgr.GetToken(ctx, uin, tokenmgr.PlatformSlack)
	if !ok {
		logs.WarnContextf(ctx, "slack search get token failed, uin: %d", uin)
		return nil, fmt.Errorf("未找到Slack授权信息，请先绑定Slack账号")
	}

	client := slack.New(tokenInfo.AccessToken, slack.OptionHTTPClient(oauthUtils.CreateHttpClientWithProxy()))
	return client, nil
}

func (ss *slackSearch) marshalOutput(ctx context.Context, output any) (string, error) {
	searchResults, ok := output.(*SearchResult)
	if !ok {
		logs.WarnContextf(ctx, "slack search marshal output failed, expect %T but given %T", searchResults, output)
		return "", fmt.Errorf("unexpected slack search response, expect %T but given %T", searchResults, output)
	}

	var messages []*Message
	for _, msg := range searchResults.Messages.Matches {
		m := parseSlackMessage(&msg)
		messages = append(messages, &m)
	}

	var files []*File
	for _, file := range searchResults.Files.Matches {
		f := parseSlackFile(&file)
		files = append(files, &f)
	}

	searchResultOut := &SearchResultOut{
		Total: searchResults.Total,
		Messages: &Messages{
			Total:   searchResults.Messages.Total,
			Matches: messages,
		},
		Files: &Files{
			Total:   searchResults.Files.Total,
			Matches: files,
		},
	}

	jsonData, err := json.Marshal(searchResultOut)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func parseSlackMessage(slackMsg *slack.SearchMessage) Message {
	msg := Message{}
	msg.ID = slackMsg.Channel.ID + ":" + slackMsg.Timestamp
	msg.ChannelId = slackMsg.Channel.ID
	msg.ChannelName = slackMsg.Channel.Name
	msg.Type = "message"
	msg.UserId = slackMsg.User
	msg.Username = slackMsg.Username
	msg.Text = slackMsg.Text
	// Extract timestamp
	msg.InternalDate = slackMsg.Timestamp
	timestamp, err := slackTimestampToString(slackMsg.Timestamp)
	if err == nil {
		msg.Date = timestamp
	}

	// Extract file information if present
	// var attachments []Attachment
	if len(slackMsg.Attachments) > 0 {
		// for _, attachment := range slackMsg.Attachments {
		// attachments = append(attachments, Attachment{

		// })
		// }
	}
	return msg
}

func parseSlackFile(slackFile *slack.File) File {
	file := File{}
	file.ID = slackFile.ID
	file.Type = "file"
	file.Name = slackFile.Name
	file.Title = slackFile.Title

	file.Filetype = slackFile.Filetype
	file.Mimetype = slackFile.Mimetype
	file.PrettyType = slackFile.PrettyType
	file.Size = slackFile.Size

	file.UserId = slackFile.User

	// file.Channels = slackFile.Channels
	file.InternalDate = strconv.FormatInt(int64(slackFile.Timestamp), 10)
	file.Date = slackFile.Timestamp.Time().Format("2006-01-02 15:04:05")

	file.IsPublic = slackFile.IsPublic
	file.WebLinkPublic = slackFile.PermalinkPublic
	file.WebLink = slackFile.Permalink
	file.Preview = slackFile.Preview
	file.PreviewHighlight = slackFile.PreviewHighlight
	file.Thumb360 = slackFile.Thumb360
	file.Thumb360Gif = slackFile.Thumb360Gif
	file.Thumb360W = slackFile.Thumb360W
	file.Thumb360H = slackFile.Thumb360H
	return file
}

func slackTimestampToString(ts string) (string, error) {
	if ts == "" {
		return "", nil
	}

	// 拆分秒和小数部分
	parts := strings.Split(ts, ".")
	sec, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return "", err
	}

	nsec := int64(0)
	if len(parts) > 1 {
		frac, err := strconv.ParseFloat("0."+parts[1], 64)
		if err != nil {
			return "", err
		}
		nsec = int64(frac * 1e9) // 转成纳秒
	}

	t := time.Unix(sec, nsec)
	return t.Format("2006-01-02 15:04:05"), nil
}

func buildSlackQuery(req *SearchRequest) string {
	var parts []string

	// 关键词，用 OR 拼接并加括号
	if len(req.Queries) > 0 {
		// TODO 批量速率限制严重，先处理第一个关键词
		parts = append(parts, req.Queries[0])
	}

	// 搜索频道
	for _, ch := range req.InChannels {
		if ch != "" {
			parts = append(parts, "in:"+ch)
		}
	}

	// 搜索用户
	for _, u := range req.FromUsers {
		if u != "" {
			parts = append(parts, "from:"+u)
		}
	}

	// 时间范围
	if req.After != "" {
		parts = append(parts, "after:"+req.After)
	}
	if req.Before != "" {
		parts = append(parts, "before:"+req.Before)
	}
	if req.On != "" {
		parts = append(parts, "on:"+req.On)
	}
	if req.During != "" {
		parts = append(parts, "during:"+req.During)
	}

	// 最终拼接
	return strings.Join(parts, " ")
}

type SearchRequest struct {
	// 系统内部唯一用户标识，用于查找 Slack 授权信息
	Uin uint `json:"uin" jsonschema:"required,description=Unique internal user ID (UIN) for Slack authorization lookup"`

	// 搜索关键词列表，用于查找消息和文件
	Queries []string `json:"queries,omitempty" jsonschema:"description=List of keywords to search messages and files"`

	// 搜索频道列表，例如 #general, #team-marketing
	InChannels []string `json:"in_channels,omitempty" jsonschema:"description=Channels to search in"`

	// 搜索指定用户
	FromUsers []string `json:"from_users,omitempty" jsonschema:"description=Users to search from, e.g., @alice"`

	// 时间范围
	After  string `json:"after,omitempty" jsonschema:"description=After date, e.g., 2025-09-01"`
	Before string `json:"before,omitempty" jsonschema:"description=Before date, e.g., 2025-09-30"`
	On     string `json:"on,omitempty" jsonschema:"description=Exact date, e.g., 2025-09-26"`
	During string `json:"during,omitempty" jsonschema:"description=Month or year, e.g., august"`

	// 返回的最大消息数，默认 20
	MaxResults int `json:"max_results,omitempty" jsonschema:"description=Maximum number of results to return (default 20)"`
}

func (r *SearchRequest) String() string {
	jsonData, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(jsonData)
}

type SearchResult struct {
	Total    int                   `json:"total,omitempty"`
	Messages *slack.SearchMessages `json:"messages,omitempty"`
	Files    *slack.SearchFiles    `json:"files,omitempty"`
}

type SearchResultOut struct {
	Total    int       `json:"total,omitempty"`
	Messages *Messages `json:"messages,omitempty"`
	Files    *Files    `json:"files,omitempty"`
}

type BaseModel struct {
	Type          string `json:"type"`
	ID            string `json:"id"`
	Date          string `json:"date"`
	InternalDate  string `json:"internalDate"`
	WebLink       string `json:"webLink,omitempty"`
	WebLinkPublic string `json:"webLinkPublic,omitempty"`
	UserId        string `json:"userId,omitempty"`
	Username      string `json:"Username,omitempty"`
}

type Messages struct {
	Matches []*Message `json:"messages,omitempty"`
	Total   int        `json:"total,omitempty"`
}

type Message struct {
	BaseModel

	ChannelId   string `json:"channelId"`
	ChannelName string `json:"channelName,omitempty"`
	Text        string `json:"text"`
}

type Files struct {
	Matches []*File `json:"files,omitempty"`
	Total   int     `json:"total,omitempty"`
}
type File struct {
	BaseModel

	Name       string `json:"name"`
	Title      string `json:"title,omitempty"`
	Filetype   string `json:"filetype,omitempty"`
	Mimetype   string `json:"mimetype,omitempty"`
	PrettyType string `json:"pretty_type,omitempty"`
	Size       int    `json:"size"`

	Preview          string `json:"preview"`
	PreviewHighlight string `json:"preview_highlight"`

	IsPublic bool `json:"is_public,omitempty"`

	Thumb360    string `json:"thumb_360"`
	Thumb360Gif string `json:"thumb_360_gif"`
	Thumb360W   int    `json:"thumb_360_w"`
	Thumb360H   int    `json:"thumb_360_h"`
}
