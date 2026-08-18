package coze

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/hertz/pkg/app"
	hertzconsts "github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"

	"github.com/insmtx/corekg/apps/workflow/api/model/app/developer_api"
	"github.com/insmtx/corekg/apps/workflow/api/model/workflow"
	"github.com/insmtx/corekg/apps/workflow/application/upload"
	appworkflow "github.com/insmtx/corekg/apps/workflow/application/workflow"
	bizConf "github.com/insmtx/corekg/apps/workflow/bizpkg/config"
	"github.com/insmtx/corekg/apps/workflow/bizpkg/llm/modelbuilder"
	"github.com/ygpkg/yg-go/logs"
	appconsts "github.com/insmtx/corekg/apps/workflow/types/consts"
)

type botConfigCreateRequest struct {
	Scene   int32  `json:"scene" form:"scene" query:"scene"`
	SpaceID int64  `json:"space_id,string" form:"space_id" query:"space_id"`
	Query   string `json:"query" form:"query" query:"query"`
}

type botConfigCreateResponse struct {
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	IconURL            string              `json:"icon_url"`
	IconURI            string              `json:"icon_uri"`
	Prompt             string              `json:"prompt"`
	SuggestedQuestions []string            `json:"suggested_questions"`
	Prologue           string              `json:"prologue"`
	Workflow           []botConfigWorkflow `json:"workflow"`
}

type botConfigModelOutput struct {
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	Prompt             string              `json:"prompt"`
	SuggestedQuestions []string            `json:"suggested_questions"`
	Prologue           string              `json:"prologue"`
	Workflow           []botConfigWorkflow `json:"workflow"`
}

type botConfigWorkflow struct {
	WorkflowID   string `json:"workflow_id"`
	Desc         string `json:"desc,omitempty"`
	PluginID     string `json:"plugin_id"`
	FlowMode     int32  `json:"flow_mode"`
	WorkflowName string `json:"workflow_name"`
}

// CreateBotConfig generates a bot configuration using the builtin default model.
// @router /api/playground_api/bot_config/create [POST]
func CreateBotConfig(ctx context.Context, c *app.RequestContext) {
	var req botConfigCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}

	if req.SpaceID <= 0 {
		invalidParamRequestResponse(c, "space_id is required")
		return
	}

	query := strings.TrimSpace(req.Query)
	if query == "" {
		invalidParamRequestResponse(c, "query is required")
		return
	}

	data, err := generateBotConfig(ctx, req)
	if err != nil {
		logs.WarnContextf(ctx, "CreateBotConfig: model generation failed, fallback to template: %v", err)
		data = fallbackBotConfig(req)
	}

	iconURL, iconURI := getDefaultBotIcon(ctx)
	resp := normalizeBotConfig(req, data)
	resp.IconURL = iconURL
	resp.IconURI = iconURI

	c.JSON(hertzconsts.StatusOK, resp)
}

func generateBotConfig(ctx context.Context, req botConfigCreateRequest) (*botConfigModelOutput, error) {
	modelList, err := bizConf.ModelConf().GetOnlineModelListWithLimit(ctx, 1)
	if err != nil {
		return nil, err
	}
	if len(modelList) == 0 {
		return nil, errors.New("no available model")
	}

	chatModel, err := modelbuilder.BuildModelWithConf(ctx, modelList[0])
	if err != nil {
		return nil, err
	}

	workflowCandidates, err := listAvailableWorkflows(ctx, req.SpaceID)
	if err != nil {
		logs.WarnContextf(ctx, "CreateBotConfig: list workflows failed: %v", err)
	}
	lookup := buildWorkflowLookup(workflowCandidates)

	systemPrompt := `你是【单智能体配置生成器】。
你的任务是：根据用户需求生成一个完整的智能体配置。
只能返回一个 JSON 对象，不得输出任何解释性文本。

【输出字段（必须全部包含）】
- name
- description
- prompt
- suggested_questions
- prologue
- workflow

【字段要求】
1. name：简短明确，长度不超过 50 字。
2. description：描述智能体的用途与能力，长度不超过 2000 字。
3. workflow：数组类型：
   - 仅当用户提供的 workflow 名称与描述与该智能体任务**明确相关且必要**时，才可以选择对应 workflow。
   - 不得臆测或泛化选择 workflow，不确定相关时必须返回空数组。
   - 如需使用工作流，格式为：{"workflow":[{"workflow_id":"xxx"}]}。
   - 如无需使用工作流，返回：{"workflow":[]}。
   - workflow 数组元素只能包含 workflow_id 字段，不得包含其他字段。
4. prompt：使用 Markdown 格式，且必须包含三个部分：
- 使用 Markdown 格式书写智能体提示词。 
- 内容应明确该智能体的身份定位、能力范围与行为边界。 
- 用于指导智能体如何理解用户需求、如何回应、以及应避免的行为。
- ⚠️【关于 workflow 的嵌入要求】  
  - 当 workflow 字段不为空时，必须在 prompt 中以自然语言方式说明工作流的使用时机与作用。  
  - 描述中必须显式引用 workflow_id，且引用格式必须为：  
    
   {{workflow_id}}
    
  - 语义上应体现：  
    - 在什么条件或场景下使用 {{workflow_id}} 
    - 使用 {{workflow_id}} 来完成什么能力或步骤  
    - 该工作流的结果如何影响最终回复  
  - {{workflow_id}} 必须与 workflow 字段中选择的 workflow_id 完全一致。  
  - 该说明应自然融入整体提示词语义中，而不是作为固定模板段或占位区块。
5. suggested_questions：数组形式，必须包含 3 条精炼示例问题。
6. prologue：简短自然的问候或引导语。


【语言要求】
- 输出语言必须与用户需求使用相同的语言。

【严格禁止】
- 不得输出 JSON 之外的任何文本。
- 不得输出 Markdown 外层包裹。
- 不得缺少字段或新增字段。
- 不得输出多个 JSON 对象。`
	userPrompt := strings.Join([]string{
		fmt.Sprintf("场景：%d", req.Scene),
		fmt.Sprintf("需求：%s", strings.TrimSpace(req.Query)),
		formatWorkflowCandidates(workflowCandidates),
	}, "\n")

	resp, err := chatModel.Generate(ctx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || strings.TrimSpace(resp.Content) == "" {
		return nil, errors.New("model returned empty response")
	}

	data, err := parseBotConfigOutput(resp.Content)
	if err != nil {
		return nil, err
	}
	data.Workflow = enrichWorkflowList(data.Workflow, lookup)
	data.Prompt = normalizeWorkflowPrompt(data.Prompt, data.Workflow)
	return data, nil
}

var workflowPromptRef = regexp.MustCompile(`\{\{\s*([^\s{}]+)\s*\}\}`)

func normalizeWorkflowPrompt(prompt string, workflows []botConfigWorkflow) string {
	if strings.TrimSpace(prompt) == "" {
		return prompt
	}

	workflowLookup := make(map[string]botConfigWorkflow, len(workflows))
	workflowNameLookup := make(map[string]botConfigWorkflow, len(workflows))
	for _, workflow := range workflows {
		workflowID := strings.TrimSpace(workflow.WorkflowID)
		if workflowID == "" {
			continue
		}
		workflowLookup[workflowID] = workflow
		workflowName := workflowNameKey(workflow.WorkflowName)
		if workflowName != "" {
			workflowNameLookup[workflowName] = workflow
		}
	}

	return workflowPromptRef.ReplaceAllStringFunc(prompt, func(match string) string {
		parts := workflowPromptRef.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		workflowID := strings.TrimSpace(parts[1])
		if workflowID == "" {
			return ""
		}
		workflow, ok := workflowLookup[workflowID]
		if !ok {
			if candidate, matchByName := workflowNameLookup[workflowNameKey(workflowID)]; matchByName {
				workflow = candidate
			} else {
				return ""
			}
		}
		workflowName := strings.TrimSpace(workflow.WorkflowName)
		if workflowName == "" {
			workflowName = workflowID
		}
		blockID := strings.TrimSpace(workflow.WorkflowID)
		if blockID == "" {
			blockID = workflowID
		}
		return fmt.Sprintf("{#LibraryBlock id=%q uuid=%q type=%q#}%s{#/LibraryBlock#}", blockID, uuid.NewString(), "workflow", workflowName)
	})
}

func parseBotConfigOutput(raw string) (*botConfigModelOutput, error) {
	candidate := strings.TrimSpace(raw)
	if candidate == "" {
		return nil, errors.New("empty model output")
	}

	data, err := parseBotConfigCandidate(candidate)
	if err == nil {
		return data, nil
	}

	obj := extractJSONObject(candidate)
	if obj != "" && obj != candidate {
		if data, objErr := parseBotConfigCandidate(obj); objErr == nil {
			return data, nil
		}
	}

	return nil, err
}

func parseBotConfigCandidate(candidate string) (*botConfigModelOutput, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(candidate), &raw); err != nil {
		return nil, err
	}

	output := &botConfigModelOutput{
		Name:        readString(raw, "name"),
		Description: readString(raw, "description"),
		Prompt:      readString(raw, "prompt"),
		Prologue:    readString(raw, "prologue"),
	}
	output.SuggestedQuestions = readStringSlice(raw, "suggested_questions", "suggestedQuestions")
	output.Workflow = readWorkflowList(raw, "workflow", "workflows")

	return output, nil
}

func readString(raw map[string]any, keys ...string) string {
	for _, key := range keys {
		val, ok := raw[key]
		if !ok || val == nil {
			continue
		}
		if s, ok := val.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func readStringSlice(raw map[string]any, keys ...string) []string {
	for _, key := range keys {
		val, ok := raw[key]
		if !ok || val == nil {
			continue
		}
		switch typed := val.(type) {
		case []any:
			items := make([]string, 0, len(typed))
			for _, item := range typed {
				if s, ok := item.(string); ok {
					items = append(items, strings.TrimSpace(s))
				}
			}
			return items
		case []string:
			items := make([]string, 0, len(typed))
			for _, item := range typed {
				items = append(items, strings.TrimSpace(item))
			}
			return items
		case string:
			return splitLines(typed)
		}
	}
	return nil
}

func readWorkflowList(raw map[string]any, keys ...string) []botConfigWorkflow {
	for _, key := range keys {
		val, ok := raw[key]
		if !ok || val == nil {
			continue
		}
		items, ok := val.([]any)
		if !ok {
			continue
		}
		workflows := make([]botConfigWorkflow, 0, len(items))
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if wf, ok := parseWorkflowItem(m); ok {
				workflows = append(workflows, wf)
			}
		}
		return workflows
	}
	return nil
}

func parseWorkflowItem(raw map[string]any) (botConfigWorkflow, bool) {
	workflowID := readString(raw, "workflow_id", "workflowId")
	if workflowID == "" {
		return botConfigWorkflow{}, false
	}
	return botConfigWorkflow{
		WorkflowID:   workflowID,
		PluginID:     "",
		FlowMode:     0,
		WorkflowName: readString(raw, "workflow_name", "workflowName"),
	}, true
}

func splitLines(input string) []string {
	lines := strings.Split(input, "\n")
	items := make([]string, 0, len(lines))
	for _, line := range lines {
		value := strings.TrimSpace(line)
		if value == "" {
			continue
		}
		items = append(items, value)
	}
	return items
}

func extractJSONObject(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```JSON")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
		trimmed = strings.TrimSuffix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
	}

	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		return trimmed[start : end+1]
	}
	return ""
}

func normalizeBotConfig(req botConfigCreateRequest, data *botConfigModelOutput) botConfigCreateResponse {
	if data == nil {
		data = &botConfigModelOutput{}
	}

	name := strings.TrimSpace(data.Name)
	if name == "" {
		name = defaultBotName(req.Query)
	}
	name = trimRunes(name, 50)

	description := strings.TrimSpace(data.Description)
	if description == "" {
		description = defaultBotDescription(req.Query)
	}
	description = trimRunes(description, 2000)

	prompt := strings.TrimSpace(data.Prompt)
	if prompt == "" {
		prompt = defaultBotPrompt(req.Query)
	}

	suggested := normalizeSuggestedQuestions(data.SuggestedQuestions)
	if len(suggested) == 0 {
		suggested = defaultSuggestedQuestions(req.Query)
	}
	if len(suggested) > 3 {
		suggested = suggested[:3]
	}

	prologue := strings.TrimSpace(data.Prologue)
	if prologue == "" {
		prologue = defaultBotPrologue(req.Query)
	}

	return botConfigCreateResponse{
		Name:               name,
		Description:        description,
		Prompt:             prompt,
		SuggestedQuestions: suggested,
		Prologue:           prologue,
		Workflow:           normalizeWorkflowList(data.Workflow),
	}
}

func normalizeSuggestedQuestions(items []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func defaultBotName(query string) string {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return "智能助手"
	}
	return trimRunes(fmt.Sprintf("%s智能助手", trimmed), 50)
}

func defaultBotDescription(query string) string {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return "一个能够帮助你完成需求的智能体。"
	}
	return fmt.Sprintf("一个专注于 %s 的智能体。", trimmed)
}

func defaultBotPrompt(query string) string {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		trimmed = "你的需求"
	}
	return strings.Join([]string{
		"# 角色",
		"你是一个专业的智能助手。",
		"# 技能",
		fmt.Sprintf("你擅长处理 %s。", trimmed),
		"# 限制",
		"回答需简洁、准确，必要时主动提出澄清问题。",
	}, "\n")
}

func defaultBotPrologue(query string) string {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return "你好，我可以帮助你处理需求。"
	}
	return fmt.Sprintf("你好，我可以帮你处理 %s。", trimmed)
}

func defaultSuggestedQuestions(query string) []string {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return []string{
			"你可以帮我完成这个任务吗？",
			"有没有更好的处理思路？",
			"请给我一个示例。",
		}
	}
	return []string{
		fmt.Sprintf("你可以帮我处理 %s 吗？", trimmed),
		fmt.Sprintf("%s 有什么推荐流程？", trimmed),
		fmt.Sprintf("请给我一个关于 %s 的示例。", trimmed),
	}
}

func trimRunes(input string, max int) string {
	if max <= 0 {
		return ""
	}
	if utf8.RuneCountInString(input) <= max {
		return input
	}
	runes := []rune(input)
	if len(runes) <= max {
		return input
	}
	return string(runes[:max])
}

func fallbackBotConfig(req botConfigCreateRequest) *botConfigModelOutput {
	return &botConfigModelOutput{
		Name:               defaultBotName(req.Query),
		Description:        defaultBotDescription(req.Query),
		Prompt:             defaultBotPrompt(req.Query),
		SuggestedQuestions: defaultSuggestedQuestions(req.Query),
		Prologue:           defaultBotPrologue(req.Query),
		Workflow:           []botConfigWorkflow{},
	}
}

func getDefaultBotIcon(ctx context.Context) (string, string) {
	iconURI := appconsts.DefaultAgentIcon
	iconResp, err := upload.SVC.GetIcon(ctx, &developer_api.GetIconRequest{
		IconType: developer_api.IconType_Bot,
	})
	if err != nil {
		logs.WarnContextf(ctx, "CreateBotConfig: get default icon failed: %v", err)
		return "", iconURI
	}
	if iconResp == nil || iconResp.Data == nil || len(iconResp.Data.IconList) == 0 {
		return "", iconURI
	}

	icon := iconResp.Data.IconList[0]
	if icon.URI != "" {
		iconURI = icon.URI
	}

	return icon.URL, iconURI
}

func normalizeWorkflowList(items []botConfigWorkflow) []botConfigWorkflow {
	if len(items) == 0 {
		return []botConfigWorkflow{}
	}
	result := make([]botConfigWorkflow, 0, len(items))
	for _, item := range items {
		workflowID := strings.TrimSpace(item.WorkflowID)
		if workflowID == "" {
			continue
		}
		workflowName := strings.TrimSpace(item.WorkflowName)
		result = append(result, botConfigWorkflow{
			WorkflowID:   workflowID,
			WorkflowName: workflowName,
			PluginID:     workflowID,
			FlowMode:     0,
		})
	}
	if len(result) == 0 {
		return []botConfigWorkflow{}
	}
	return result
}

func listAvailableWorkflows(ctx context.Context, spaceID int64) ([]botConfigWorkflow, error) {
	spaceStr := strconv.FormatInt(spaceID, 10)
	page := int32(1)
	size := int32(50)
	status := workflow.WorkFlowListStatus_HadPublished
	req := &workflow.GetWorkFlowListRequest{
		Page:    &page,
		Size:    &size,
		SpaceID: &spaceStr,
		Status:  &status,
	}

	resp, err := appworkflow.SVC.ListWorkflow(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Data == nil || len(resp.Data.WorkflowList) == 0 {
		return nil, nil
	}

	items := make([]botConfigWorkflow, 0, len(resp.Data.WorkflowList))
	for _, wf := range resp.Data.WorkflowList {
		if wf == nil {
			continue
		}
		items = append(items, botConfigWorkflow{
			WorkflowID:   wf.WorkflowID,
			WorkflowName: wf.Name,
			Desc:         wf.Desc,
			PluginID:     wf.PluginID,
			FlowMode:     int32(wf.FlowMode),
		})
	}

	return items, nil
}

func buildWorkflowLookup(candidates []botConfigWorkflow) map[string]botConfigWorkflow {
	lookup := make(map[string]botConfigWorkflow, len(candidates))
	for _, candidate := range candidates {
		if candidate.WorkflowID == "" {
			continue
		}
		lookup[candidate.WorkflowID] = candidate
	}
	return lookup
}

func workflowNameKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func enrichWorkflowList(items []botConfigWorkflow, lookup map[string]botConfigWorkflow) []botConfigWorkflow {
	if len(items) == 0 {
		return []botConfigWorkflow{}
	}
	nameLookup := make(map[string]botConfigWorkflow, len(lookup))
	for _, candidate := range lookup {
		key := workflowNameKey(candidate.WorkflowName)
		if key == "" {
			continue
		}
		nameLookup[key] = candidate
	}
	result := make([]botConfigWorkflow, 0, len(items))
	for _, item := range items {
		workflowID := strings.TrimSpace(item.WorkflowID)
		workflowName := strings.TrimSpace(item.WorkflowName)
		candidate, ok := lookup[workflowID]
		if !ok {
			nameKey := workflowNameKey(workflowID)
			if nameKey == "" {
				continue
			}
			if matched, matchByName := nameLookup[nameKey]; matchByName {
				candidate = matched
				workflowID = strings.TrimSpace(candidate.WorkflowID)
				if workflowName == "" {
					workflowName = strings.TrimSpace(candidate.WorkflowName)
				}
			} else {
				continue
			}
		}
		if workflowID == "" {
			workflowID = strings.TrimSpace(candidate.WorkflowID)
		}
		if workflowName == "" {
			workflowName = strings.TrimSpace(candidate.WorkflowName)
		}
		desc := strings.TrimSpace(item.Desc)
		if desc == "" {
			desc = strings.TrimSpace(candidate.Desc)
		}
		result = append(result, botConfigWorkflow{
			WorkflowID:   workflowID,
			WorkflowName: workflowName,
			Desc:         desc,
			PluginID:     workflowID,
			FlowMode:     0,
		})
	}
	if len(result) == 0 {
		return []botConfigWorkflow{}
	}
	return result
}

func formatWorkflowCandidates(candidates []botConfigWorkflow) string {
	if len(candidates) == 0 {
		return "可用工作流列表：无"
	}

	const maxCandidates = 20
	limit := len(candidates)
	if limit > maxCandidates {
		limit = maxCandidates
	}

	lines := make([]string, 0, limit+1)
	lines = append(lines, "可用工作流列表（可按需选择，不必一定使用）：")
	for i := 0; i < limit; i++ {
		wf := candidates[i]
		desc := strings.TrimSpace(wf.Desc)
		if desc == "" {
			desc = "无描述"
		}
		lines = append(lines, fmt.Sprintf("%d. {workflow_id:%s, name:%s, desc:%s}", i+1, wf.WorkflowID, wf.WorkflowName, desc))
	}

	return strings.Join(lines, "\n")
}
