package gmail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/jaytaylor/html2text"
	"github.com/insmtx/corekg/pkgs/connectors/tokenmgr"
	oauthUtils "github.com/insmtx/corekg/pkgs/utils/oauth"
	"github.com/ygpkg/yg-go/logs"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

const defaultMaxResultNum = 20
const maxResultNum = 500

type Config struct {
	ToolName   string `json:"tool_name"`   // default: gmail_search
	ToolDesc   string `json:"tool_desc"`   // default: "Search emails in Gmail by keyword, sender, recipient, or date range"
	MaxResults int64  `json:"max_results"` // default: 20
}

type gmailSearch struct {
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
		toolName = "gmail_search"
	}
	if toolDesc == "" {
		toolDesc = "Search emails in Gmail by keyword, sender, recipient, or date range"
	}

	gs := &gmailSearch{
		conf: conf,
	}

	tl, err := utils.InferTool(toolName, toolDesc, gs.search, utils.WithMarshalOutput(gs.marshalOutput))
	if err != nil {
		return nil, err
	}
	return tl, nil
}

func (gs *gmailSearch) search(ctx context.Context, req *SearchRequest) (*gmail.ListMessagesResponse, error) {
	service, err := gs.getService(ctx, req.Uin)
	if err != nil {
		logs.WarnContextf(ctx, "gmail search get service failed, uin: %d, err: %v", req.Uin, err)
		return &gmail.ListMessagesResponse{
			Messages:           make([]*gmail.Message, 0),
			ResultSizeEstimate: 0,
		}, nil
	}

	maxResults := gs.conf.MaxResults
	if req.MaxResults > 0 {
		maxResults = req.MaxResults
	}
	if maxResults > maxResultNum {
		maxResults = maxResultNum
	}

	messages, err := service.Users.Messages.List("me").Q(buildGmailQuery(req)).MaxResults(maxResults).Do()
	if err != nil {
		logs.WarnContextf(ctx, "gmail search list messages failed, uin: %d, err: %v", req.Uin, err)
		return nil, fmt.Errorf("获取邮件列表失败: %v", err)
	}

	for i, msg := range messages.Messages {
		fullMsg, err := service.Users.Messages.Get("me", msg.Id).Context(ctx).Do()
		if err != nil {
			// 可以根据需求选择是返回错误，还是跳过某条邮件
			continue
		}

		// 替换原来的 Message 对象
		messages.Messages[i] = fullMsg
	}

	return messages, nil
}

func (gs *gmailSearch) getService(ctx context.Context, uin uint) (*gmail.Service, error) {
	externalToken, ok := tokenmgr.GetToken(ctx, uin, tokenmgr.PlatformGmail)
	if !ok {
		logs.WarnContextf(ctx, "gmail search get token failed, uin: %d", uin)
		return nil, fmt.Errorf("获取token失败")
	}

	client := oauthUtils.CreateOAuth2ClientWithProxy(externalToken.AccessToken)

	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		logs.WarnContextf(ctx, "gmail search create service failed, uin: %d, err: %v", uin, err)
		return nil, fmt.Errorf("创建Gmail服务失败: %v", err)
	}

	return service, nil
}

func (gs *gmailSearch) marshalOutput(ctx context.Context, output any) (string, error) {
	gmr, ok := output.(*gmail.ListMessagesResponse)
	if !ok {
		logs.WarnContextf(ctx, "gmail search marshal output failed, expect %T but given %T", gmr, output)
		return "", fmt.Errorf("unexpected gmail list messages response, expect %T but given %T", gmr, output)
	}

	var messages []*Message
	for _, msg := range gmr.Messages {
		m := parseGmailMessage(msg)
		messages = append(messages, &m)
	}

	searchResult := &SearchResult{
		ResultSizeEstimate: gmr.ResultSizeEstimate,
		Messages:           messages,
	}

	jsonData, err := json.Marshal(searchResult)
	if err != nil {
		logs.WarnContextf(ctx, "gmail search marshal output failed, err: %v", err)
		return "", err
	}
	return string(jsonData), nil
}

func parseGmailMessage(msg *gmail.Message) Message {
	message := Message{
		ID:       msg.Id,
		ThreadID: msg.ThreadId,
		Snippet:  msg.Snippet,
		LabelIDs: msg.LabelIds,
		WebLink:  fmt.Sprintf("https://mail.google.com/mail/u/0/#inbox/%s", msg.Id),
	}

	if msg.Payload != nil {
		// 提取邮件头信息
		for _, header := range msg.Payload.Headers {
			switch header.Name {
			case "Subject":
				message.Subject = header.Value
			case "From":
				message.From = header.Value
			case "Date":
				message.Date = header.Value
			case "To":
				message.To = header.Value
			}
		}
		// 提取邮件正文（body）和附件ID
		body, attachments := extractBodyAndAttachments(msg.Payload)

		message.Data = body
		message.Attachments = attachments
	}

	return message
}

func extractBodyAndAttachments(payload *gmail.MessagePart) (string, []Attachment) {
	var bodies []string
	var attachments []Attachment

	var parsePart func(part *gmail.MessagePart)
	parsePart = func(part *gmail.MessagePart) {
		if part == nil {
			return
		}

		// 先递归解析子 Parts
		if len(part.Parts) > 0 {
			for _, p := range part.Parts {
				parsePart(p)
			}
		} else {
			// 正文处理
			if part.MimeType == "text/plain" || part.MimeType == "text/html" {
				if part.Body != nil && part.Body.Data != "" {
					decoded, err := decodeBase64(part.Body.Data)
					if err == nil {
						content := string(decoded)
						if part.MimeType == "text/html" {
							txt, err := html2text.FromString(content)
							if err == nil {
								content = txt
							}
						}
						bodies = append(bodies, content)
					} else {
						bodies = append(bodies, part.Body.Data) // fallback
					}
				}
			}

			// 附件处理
			if part.Filename != "" && part.Body != nil && part.Body.AttachmentId != "" {
				attachments = append(attachments, Attachment{
					ID:       part.Body.AttachmentId,
					Filename: part.Filename,
				})
			}
		}
	}

	parsePart(payload)

	// TODO 邮件会出现同时有 text/plain 和 text/html 部分，当前策略会提取文本内容合并
	finalBody := strings.Join(bodies, "\n\n")

	return finalBody, attachments
}

func decodeBase64(data string) ([]byte, error) {
	if m := len(data) % 4; m != 0 {
		data += strings.Repeat("=", 4-m)
	}
	return base64.URLEncoding.DecodeString(data)
}

func buildGmailQuery(req *SearchRequest) string {
	var parts []string

	// 关键词
	if len(req.Queries) > 0 {
		if len(req.Queries) == 1 {
			parts = append(parts, req.Queries[0])
		} else {
			// Gmail 搜索语法中需要明确写 OR
			parts = append(parts, "("+strings.Join(req.Queries, " OR ")+")")
		}
	}

	// 发件人
	if req.From != "" {
		parts = append(parts, fmt.Sprintf("from:%s", req.From))
	}

	// 收件人
	if req.To != "" {
		parts = append(parts, fmt.Sprintf("to:%s", req.To))
	}

	// 起始日期
	if req.StartDate != nil {
		parts = append(parts, fmt.Sprintf("after:%s", req.StartDate.Format("2006/01/02")))
	}

	// 结束日期
	if req.EndDate != nil {
		parts = append(parts, fmt.Sprintf("before:%s", req.EndDate.Format("2006/01/02")))
	}

	// 拼接成一个字符串
	return strings.Join(parts, " ")
}

type SearchRequest struct {
	Uin uint `json:"uin" jsonschema:"required,description=System internal unique user identifier (UIN) used to lookup Gmail authorization info"`

	Queries []string `json:"omitempty" jsonschema:"required,description=List of query keywords to search emails"`

	From string `json:"from,omitempty" jsonschema:"description=Filter emails from this sender email address"`

	To string `json:"to,omitempty" jsonschema:"description=Filter emails sent to this recipient email address"`

	StartDate *time.Time `json:"start_date,omitempty" jsonschema:"description=Start date of emails to search, in RFC3339 format"`

	EndDate *time.Time `json:"end_date,omitempty" jsonschema:"description=End date of emails to search, in RFC3339 format"`

	MaxResults int64 `json:"max_results,omitempty" jsonschema:"description=Maximum number of emails to return, default 20"`
}

func (r *SearchRequest) String() string {
	jsonData, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(jsonData)
}

type SearchResult struct {
	ResultSizeEstimate int64      `json:"resultSizeEstimate,omitempty"`
	Messages           []*Message `json:"items"`
}

type Message struct {
	ID           string       `json:"id"`
	ThreadID     string       `json:"threadId"`
	LabelIDs     []string     `json:"labelIds"`
	Snippet      string       `json:"snippet"`
	Subject      string       `json:"subject"`
	To           string       `json:"to"`
	From         string       `json:"from"`
	Date         string       `json:"date"`
	InternalDate string       `json:"internalDate"`
	Attachments  []Attachment `json:"attachments"`
	Data         string       `json:"data"`
	WebLink      string       `json:"webLink"`
}

type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
}
