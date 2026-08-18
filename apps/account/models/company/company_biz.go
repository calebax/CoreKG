package company

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/article"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
)

type CompanyInfo struct {
	Name        string `json:"name"`
	Alias       string `json:"alias"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Address     string `json:"address"`
	Tel         string `json:"tel"`
	Email       string `json:"email"`
	Website     string `json:"website"`
	UserID      uint   `json:"user_id"`
}

type QuotaTriple struct {
	Used     interface{} `json:"used"`
	Quota    interface{} `json:"quota"`
	UseRatio interface{} `json:"use_ratio"`
}

type Quota struct {
	Version  accounttype.CompanyVersion `json:"version"`
	Disk     QuotaTriple                `json:"disk"`
	QA       QuotaTriple                `json:"qa"`
	Agent    QuotaTriple                `json:"agent"`
	Employee QuotaTriple                `json:"employee"`
	Graph    QuotaTriple                `json:"graph"`
	Article  QuotaTriple                `json:"article"`
}

func GetCompanyQuota(ctx context.Context, companyID uint) (*Quota, error) {
	cmp, err := GetCompany(companyID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCompanyQuota: get company failed, %v", err)
		return nil, fmt.Errorf("GetCompany(%d): %w", companyID, err)
	}

	diskSize, err := forest.GetFilesSizeByCompanyID(ctx, companyID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCompanyQuota: GetFilesSizeByCompanyID failed, %v", err)
		return nil, err
	}
	emps, err := employee.GetEmployeeByCompanyID(companyID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCompanyQuota: GetEmployeeByCompanyID failed, %v", err)
		return nil, err
	}

	qas, err := chatquestion.GetUnscopedQAByCompanyID(ctx, companyID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCompanyQuota: GetUnscopedQAByCompanyID failed, %v", err)
		return nil, err
	}
	ags, err := chatagent.GetALLAgentsByCompanyID(ctx, companyID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCompanyQuota: GetAllAgentsByCompanyID failed, %v", err)
		return nil, err
	}
	gs, err := graph.GetGraphByCompanyID(ctx, cmp.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCompanyQuota: GetGraphByCompanyID failed, %v", err)
		return nil, err
	}
	arts, err := article.GetArticleByCompanyID(ctx, companyID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCompanyQuota: GetArticleByCompanyID failed, %v", err)
		return nil, err
	}

	return &Quota{
		Version: cmp.Version,
		Disk: QuotaTriple{
			forest.FormatFileSize(diskSize),
			forest.FormatFileSize(cmp.Quota.DiskQuota),
			fmt.Sprintf("%.3f", float64(diskSize)/float64(cmp.Quota.DiskQuota))},
		QA: QuotaTriple{
			qas,
			cmp.Quota.QAQuota,
			fmt.Sprintf("%.3f", float64(qas)/float64(cmp.Quota.QAQuota))},
		Agent: QuotaTriple{
			len(ags),
			cmp.Quota.AgentQuota,
			fmt.Sprintf("%.3f", float64(len(ags))/float64(cmp.Quota.AgentQuota))},
		Employee: QuotaTriple{
			len(emps),
			cmp.Quota.EmployeeQuota,
			fmt.Sprintf("%.3f", float64(len(emps))/float64(cmp.Quota.EmployeeQuota))},
		Graph: QuotaTriple{
			len(gs),
			cmp.Quota.GraphQuota,
			fmt.Sprintf("%.3f", float64(len(gs))/float64(cmp.Quota.GraphQuota))},
		Article: QuotaTriple{
			len(arts),
			cmp.Quota.ArticleQuota,
			fmt.Sprintf("%.3f", float64(len(arts))/float64(cmp.Quota.ArticleQuota))},
	}, nil
}

type Message struct {
	MsgType string      `json:"msgtype"`
	Text    TextMessage `json:"text"`
}

type TextMessage struct {
	Content string `json:"content"`
}

func WechatBotWebhook(ctx context.Context, userData *accounttype.CompanyUpgradeApply, clientIP string) (err error) {
	var webhookURL string
	switch userData.Type {
	case accounttype.FormTypeDotpenContact:
		webhookURL, err = settings.GetValue("corekg", "official_dotpen_website_wechat_webhook_url")
		if err != nil {
			logs.ErrorContextf(ctx, "WechatBotWebhook: get official_dotpen_website_wechat_webhook_url failed, %v", err)
			return err
		}
	default:
		webhookURL, err = settings.GetValue("corekg", "official_website_wechat_webhook_url")
		if err != nil {
			logs.ErrorContextf(ctx, "WechatBotWebhook: get official_website_wechat_webhook_url failed, %v", err)
			return err
		}
	}
	if webhookURL == "" {
		logs.ErrorContext(ctx, "WechatBotWebhook: webhook URL is empty")
		return fmt.Errorf("webhook URL is empty")
	}

	method := "POST"
	str := `名字：` + userData.Name + `
手机号：` + userData.Phone + `
公司名称：` + userData.CompanyName + `
公司规模：` + userData.Scale + `
所属行业：` + userData.Industry + `
	客户诉求：` + userData.Claim + `
客户端IP：` + clientIP + `
时间：` + time.Now().Format("2006-01-02 15:04:05")
	msg := Message{
		MsgType: "text",
		Text: TextMessage{
			Content: str,
		},
	}
	jsonData, err := json.Marshal(msg)
	if err != nil {
		logs.ErrorContext(ctx, err)
		return
	}
	payload := bytes.NewReader(jsonData)
	client := &http.Client{}
	req, err := http.NewRequest(method, webhookURL, payload)
	if err != nil {
		logs.ErrorContext(ctx, err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		logs.ErrorContext(ctx, err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("http.Status: %s", res.Status)
		return
	}

	return
}
