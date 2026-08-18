package coze

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/insmtx/corekg/apps/workflow/api/model/app/bot_common"
	application "github.com/insmtx/corekg/apps/workflow/application/singleagent"
)

type produceCreateBotRequest struct {
	SpaceID            int64           `json:"space_id,string" form:"space_id" query:"space_id"`
	Name               string          `json:"name" form:"name" query:"name"`
	Description        string          `json:"description" form:"description" query:"description"`
	IconURL            string          `json:"icon_url" form:"icon_url" query:"icon_url"`
	IconURI            string          `json:"icon_uri" form:"icon_uri" query:"icon_uri"`
	Prompt             string          `json:"prompt" form:"prompt" query:"prompt"`
	PluginAPIs         string          `json:"plugin_apis" form:"plugin_apis" query:"plugin_apis"`
	Prologue           string          `json:"prologue" form:"prologue" query:"prologue"`
	SuggestedQuestions []string        `json:"suggested_questions" form:"suggested_questions" query:"suggested_questions"`
	BotSource          int32           `json:"bot_source" form:"bot_source" query:"bot_source"`
	Knowledge          json.RawMessage `json:"knowledge" form:"knowledge" query:"knowledge"`
	Workflow           json.RawMessage `json:"workflow" form:"workflow" query:"workflow"`
}

type produceCreateBotResponse struct {
	Code int32                 `json:"code"`
	Msg  string                `json:"msg"`
	Data *produceCreateBotData `json:"data,omitempty"`
}

type produceCreateBotData struct {
	BotID       string `json:"bot_id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IconURL     string `json:"icon_url,omitempty"`
	Link        string `json:"link,omitempty"`
}

// ProduceCreateBot handles bot creation for produce API.
// @router /api/playground_api/produce/create_bot [POST]
func ProduceCreateBot(ctx context.Context, c *app.RequestContext) {
	var req produceCreateBotRequest
	if err := c.BindAndValidate(&req); err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}

	if req.SpaceID <= 0 {
		invalidParamRequestResponse(c, "space id is not set")
		return
	}

	if req.Name == "" {
		invalidParamRequestResponse(c, "name is nil")
		return
	}

	if req.IconURI == "" {
		invalidParamRequestResponse(c, "icon uri is nil")
		return
	}

	if utf8.RuneCountInString(req.Name) > 50 {
		invalidParamRequestResponse(c, "name is too long")
		return
	}

	if utf8.RuneCountInString(req.Description) > 2000 {
		invalidParamRequestResponse(c, "description is too long")
		return
	}

	knowledge, err := parseKnowledge(req.Knowledge)
	if err != nil {
		invalidParamRequestResponse(c, "knowledge is invalid")
		return
	}

	workflowInfos, err := parseWorkflowInfos(req.Workflow)
	if err != nil {
		invalidParamRequestResponse(c, "workflow is invalid")
		return
	}

	pluginInfos, err := parsePluginAPIs(req.PluginAPIs)
	if err != nil {
		invalidParamRequestResponse(c, "plugin_apis is invalid")
		return
	}

	agentID, err := application.SingleAgentSVC.CreateFullSingleAgent(ctx, &application.FullSingleAgentCreateRequest{
		SpaceID:            req.SpaceID,
		Name:               req.Name,
		Description:        req.Description,
		IconURI:            req.IconURI,
		Prompt:             req.Prompt,
		PluginInfos:        pluginInfos,
		Prologue:           req.Prologue,
		SuggestedQuestions: req.SuggestedQuestions,
		Knowledge:          knowledge,
		Workflow:           workflowInfos,
	})
	if err != nil {
		internalServerErrorResponse(ctx, c, err)
		return
	}

	resp := produceCreateBotResponse{
		Code: 0,
		Msg:  "success",
		Data: &produceCreateBotData{
			BotID:       strconv.FormatInt(agentID, 10),
			Name:        req.Name,
			Description: req.Description,
			IconURL:     req.IconURL,
		},
	}

	c.JSON(consts.StatusOK, resp)
}

func parsePluginAPIs(raw string) ([]*bot_common.PluginInfo, error) {
	payload := strings.TrimSpace(raw)
	if payload == "" {
		return nil, nil
	}

	var data any
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return nil, err
	}

	if value, ok := data.(string); ok {
		return parsePluginAPIs(value)
	}

	items, ok := data.([]any)
	if !ok {
		return nil, errors.New("plugin_apis should be an array")
	}

	infos := make([]*bot_common.PluginInfo, 0, len(items))
	for _, item := range items {
		info, ok := parsePluginAPIItem(item)
		if ok {
			infos = append(infos, info)
		}
	}

	if len(infos) == 0 {
		return nil, nil
	}

	return infos, nil
}

func parseKnowledge(raw json.RawMessage) (*bot_common.Knowledge, error) {
	payload, empty, err := normalizeJSONPayload(raw)
	if err != nil {
		return nil, err
	}
	if empty {
		return nil, nil
	}

	var knowledge bot_common.Knowledge
	if err := json.Unmarshal(payload, &knowledge); err != nil {
		return nil, err
	}

	return &knowledge, nil
}

func parseWorkflowInfos(raw json.RawMessage) ([]*bot_common.WorkflowInfo, error) {
	payload, empty, err := normalizeJSONPayload(raw)
	if err != nil {
		return nil, err
	}
	if empty {
		return nil, nil
	}

	var infos []*bot_common.WorkflowInfo
	if err := json.Unmarshal(payload, &infos); err == nil {
		return infos, nil
	}

	var info bot_common.WorkflowInfo
	if err := json.Unmarshal(payload, &info); err == nil {
		return []*bot_common.WorkflowInfo{&info}, nil
	}

	return nil, errors.New("workflow should be an array")
}

func parsePluginAPIItem(item any) (*bot_common.PluginInfo, bool) {
	m, ok := item.(map[string]any)
	if !ok {
		return nil, false
	}

	pluginID, hasPluginID := parseInt64Field(m, "plugin_id", "pluginId", "pluginID")
	apiID, hasAPIID := parseInt64Field(m, "api_id", "apiId", "apiID")
	if !hasPluginID && !hasAPIID {
		return nil, false
	}

	info := &bot_common.PluginInfo{
		PluginFrom: bot_common.PluginFromPtr(bot_common.PluginFrom_Default),
	}

	if hasPluginID {
		info.PluginId = &pluginID
	}
	if hasAPIID {
		info.ApiId = &apiID
	}

	if name, ok := parseStringField(m, "api_name", "apiName", "name"); ok {
		info.ApiName = &name
	}

	return info, true
}

func parseInt64Field(data map[string]any, keys ...string) (int64, bool) {
	for _, key := range keys {
		value, ok := data[key]
		if !ok {
			continue
		}
		if id, ok := parseInt64Value(value); ok {
			return id, true
		}
	}
	return 0, false
}

func parseStringField(data map[string]any, keys ...string) (string, bool) {
	for _, key := range keys {
		value, ok := data[key]
		if !ok {
			continue
		}
		s, ok := value.(string)
		if ok && s != "" {
			return s, true
		}
	}
	return "", false
}

func parseInt64Value(value any) (int64, bool) {
	switch v := value.(type) {
	case json.Number:
		if id, err := v.Int64(); err == nil {
			return id, true
		}
		if f, err := v.Float64(); err == nil {
			return int64(f), true
		}
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return 0, false
		}
		if id, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			return id, true
		}
		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return int64(f), true
		}
	}
	return 0, false
}

func normalizeJSONPayload(raw json.RawMessage) (json.RawMessage, bool, error) {
	if len(raw) == 0 {
		return nil, true, nil
	}

	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return nil, true, nil
	}

	var asString string
	if err := json.Unmarshal([]byte(trimmed), &asString); err == nil {
		asString = strings.TrimSpace(asString)
		if asString == "" || asString == "null" {
			return nil, true, nil
		}
		return json.RawMessage(asString), false, nil
	}

	return json.RawMessage(trimmed), false, nil
}
