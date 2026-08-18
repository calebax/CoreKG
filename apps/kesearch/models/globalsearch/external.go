package globalsearch

import (
	"encoding/json"

	"github.com/insmtx/corekg/pkgs/connectors/tokenmgr"
	"github.com/insmtx/corekg/pkgs/einotools/confluence"
	gmailtool "github.com/insmtx/corekg/pkgs/einotools/gmail"
	"github.com/insmtx/corekg/pkgs/einotools/googledrive"
	"github.com/insmtx/corekg/pkgs/einotools/slack"
	"github.com/ygpkg/yg-go/logs"
	"golang.org/x/sync/errgroup"
)

// SearchExternalDataType 知识库高亮
type SearchExternalDataType struct {
	GmailSearch      *gmailtool.SearchResult   `json:"gmail_search,omitempty"`
	GmailDriveSearch *googledrive.SearchResult `json:"gmail_drive_search,omitempty"`
	SlackSearch      *slack.SearchResultOut    `json:"slack_search,omitempty"`
	ConfluenceSearch *confluence.SearchResults `json:"confluence_search,omitempty"`
}

// SearchExternalData 检索外部数据源数据
func (wrapper GlobalSearchWrapper) SearchExternalData() (*SearchExternalDataType, error) {
	res := &SearchExternalDataType{}
	var g errgroup.Group
	// Gmail
	g.Go(func() error {
		gmailres, err := wrapper.SearchGmail()
		if err != nil {
			logs.ErrorContextf(wrapper.Ctx, "[SearchExternalData] SearchGmail error: %v", err)
			return err
		}
		res.GmailSearch = gmailres
		return nil
	})

	// Gmail Drive
	g.Go(func() error {
		gmaildriveres, err := wrapper.SearchGmailDrive()
		if err != nil {
			logs.ErrorContextf(wrapper.Ctx, "[SearchExternalData] SearchGmailDrive error: %v", err)
			return err
		}
		res.GmailDriveSearch = gmaildriveres
		return nil
	})

	// Slack
	g.Go(func() error {
		slackres, err := wrapper.SearchSlack()
		if err != nil {
			logs.ErrorContextf(wrapper.Ctx, "[SearchExternalData] SearchSlack error: %v", err)
			return err
		}
		res.SlackSearch = slackres
		return nil
	})

	// Confluence
	g.Go(func() error {
		confluenceres, err := wrapper.SearchConfluence()
		if err != nil {
			logs.ErrorContextf(wrapper.Ctx, "[SearchExternalData] SearchConfluence error: %v", err)
			return err
		}
		res.ConfluenceSearch = confluenceres
		return nil
	})

	if err := g.Wait(); err != nil {
		logs.ErrorContextf(wrapper.Ctx, "[SearchExternalData] errgroup.Wait error: %v", err)
		return res, err
	}

	return res, nil
}

// SearchGmail 搜索gmail
func (wrapper GlobalSearchWrapper) SearchGmail() (*gmailtool.SearchResult, error) {
	_, err := tokenmgr.GetTokenByUin(wrapper.Ctx, wrapper.Uin, tokenmgr.PlatformGmail)
	if err != nil {
		logs.WarnContextf(wrapper.Ctx, "GetTokenByUin err: %v", err)
		return nil, nil
	}
	conf := &gmailtool.Config{
		MaxResults: int64(wrapper.SubjectCount),
	}
	gmailTool, err := gmailtool.NewTool(wrapper.Ctx, conf)
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "[SearchGmail] new gmail tool error: %v", err)
		return nil, err
	}
	keyWords := []string{}
	for _, v := range wrapper.keywords.Tokens {
		keyWords = append(keyWords, v.Token)
	}
	gsr := &gmailtool.SearchRequest{
		Uin:        wrapper.Uin,
		Queries:    keyWords,
		MaxResults: int64(wrapper.SubjectCount),
	}
	toolOut, err := gmailTool.InvokableRun(wrapper.Ctx, gsr.String())
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "[SearchGmail] gmail tool InvokableRun error: %v", err)
		return nil, err
	}
	var response *gmailtool.SearchResult
	err = json.Unmarshal([]byte(toolOut), &response)
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "[SearchGmail] json.Unmarshal error: %v", err)
		return nil, err
	}
	return response, nil
}

// SearchGmailDrive 搜索gmail-drive
func (wrapper GlobalSearchWrapper) SearchGmailDrive() (*googledrive.SearchResult, error) {
	_, err := tokenmgr.GetTokenByUin(wrapper.Ctx, wrapper.Uin, tokenmgr.PlatformGoogleDrive)
	if err != nil {
		logs.WarnContextf(wrapper.Ctx, "GetTokenByUin err: %v", err)
		return nil, nil
	}
	conf := &googledrive.Config{
		MaxResults: int64(wrapper.SubjectCount),
	}
	gmailDriveTool, err := googledrive.NewTool(wrapper.Ctx, conf)
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "[SearchGmailDrive] new tool error: %v", err)
		return nil, err
	}
	keyWords := []string{}
	for _, v := range wrapper.keywords.Tokens {
		keyWords = append(keyWords, v.Token)
	}
	gsr := &googledrive.SearchRequest{
		Uin:        wrapper.Uin,
		Queries:    keyWords,
		MaxResults: int64(wrapper.SubjectCount),
	}
	toolOut, err := gmailDriveTool.InvokableRun(wrapper.Ctx, gsr.String())
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "[SearchGmailDrive] tool InvokableRun error: %v", err)
		return nil, err
	}
	var response *googledrive.SearchResult
	err = json.Unmarshal([]byte(toolOut), &response)
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "[SearchGmailDrive] json.Unmarshal error: %v", err)
		return nil, err
	}
	return response, nil
}

// SearchSlack 搜索搜索slack
func (wrapper GlobalSearchWrapper) SearchSlack() (*slack.SearchResultOut, error) {
	_, err := tokenmgr.GetTokenByUin(wrapper.Ctx, wrapper.Uin, tokenmgr.PlatformSlack)
	if err != nil {
		logs.WarnContextf(wrapper.Ctx, "GetTokenByUin err: %v", err)
		return nil, nil
	}
	conf := &slack.Config{
		MaxResults: wrapper.SubjectCount,
	}
	slackTool, err := slack.NewTool(wrapper.Ctx, conf)
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "[SearchSlack] new tool error: %v", err)
		return nil, err
	}
	keyWords := []string{}
	for _, v := range wrapper.keywords.Tokens {
		keyWords = append(keyWords, v.Token)
	}
	gsr := &slack.SearchRequest{
		Uin:        wrapper.Uin,
		Queries:    keyWords,
		MaxResults: wrapper.SubjectCount,
	}
	toolOut, err := slackTool.InvokableRun(wrapper.Ctx, gsr.String())
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "[SearchSlack] tool InvokableRun error: %v", err)
		return nil, err
	}
	var response *slack.SearchResultOut
	err = json.Unmarshal([]byte(toolOut), &response)
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "[SearchSlack] json.Unmarshal error: %v", err)
		return nil, err
	}
	return response, nil
}

// SearchConfluence 搜索搜索Confluence
func (wrapper GlobalSearchWrapper) SearchConfluence() (*confluence.SearchResults, error) {
	_, err := tokenmgr.GetTokenByUin(wrapper.Ctx, wrapper.Uin, tokenmgr.PlatformConfluence)
	if err != nil {
		logs.WarnContextf(wrapper.Ctx, "GetTokenByUin err: %v", err)
		return nil, nil
	}
	conf := &confluence.Config{
		MaxResults: wrapper.SubjectCount,
	}
	confluenceTool, err := confluence.NewTool(wrapper.Ctx, conf)
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "[SearchConfluence] new tool error: %v", err)
		return nil, err
	}
	keyWords := []string{}
	for _, v := range wrapper.keywords.Tokens {
		keyWords = append(keyWords, v.Token)
	}
	gsr := &confluence.SearchRequest{
		Uin:        wrapper.Uin,
		Queries:    keyWords,
		MaxResults: wrapper.SubjectCount,
	}
	toolOut, err := confluenceTool.InvokableRun(wrapper.Ctx, gsr.String())
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "[SearchConfluence] tool InvokableRun error: %v", err)
		return nil, err
	}
	var response *confluence.SearchResults
	err = json.Unmarshal([]byte(toolOut), &response)
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "[SearchConfluence] json.Unmarshal error: %v", err)
		return nil, err
	}
	return response, nil
}
