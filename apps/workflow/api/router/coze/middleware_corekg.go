package coze

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/insmtx/corekg/apps/workflow/api/model/app/intelligence"
	intelligenceCommon "github.com/insmtx/corekg/apps/workflow/api/model/app/intelligence/common"
	resource "github.com/insmtx/corekg/apps/workflow/api/model/resource"
	resourceCommon "github.com/insmtx/corekg/apps/workflow/api/model/resource/common"
	workflow "github.com/insmtx/corekg/apps/workflow/api/model/workflow"
	"github.com/insmtx/corekg/apps/workflow/application/base/ctxutil"
	"github.com/insmtx/corekg/apps/workflow/domain/permission"
	"github.com/ygpkg/yg-go/logs"
	"github.com/insmtx/corekg/apps/workflow/utils/requestyygu"
)

func corekgCreateResourcePermissionMw() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		reqBody := ctx.Request.Body()
		ctx.Next(c)

		respBody := ctx.Response.Body()
		if len(respBody) == 0 {
			logs.InfoContextf(c, "corekg permission: missing response body")
			return
		}

		path := string(ctx.Path())
		resourceID, resourceType, ok := extractCoreKGResource(c, path, reqBody, respBody)
		if !ok {
			logs.InfoContextf(c, "corekg permission: unable to parse resource info (path=%s)", path)
			return
		}

		logs.InfoContextf(c, "corekg permission: resource_id=%s resource_type=%d", resourceID, resourceType)

		userID := ctxutil.GetUIDFromCtx(c)
		if userID == nil || *userID == 0 {
			logs.ErrorContextf(c, "corekg permission: missing user id in context")
			return
		}

		resourceIDInt, err := strconv.ParseInt(resourceID, 10, 64)
		if err != nil {
			logs.ErrorContextf(c, "corekg permission: invalid resource id: %v", err)
			return
		}

		err = requestyygu.SetCoreKGResourceScope(c, &requestyygu.CoreKGSetResourceScopeRequest{
			ResourceID:     resourceIDInt,
			ResourceType:   int64(resourceType),
			ManageScopeIDs: []int64{*userID},
			ViewScopeIDs:   []int64{*userID},
			ViewScopeType:  "user",
		})
		if err != nil {
			logs.ErrorContextf(c, "corekg permission: set resource scope failed: %v", err)
			return
		}
		logs.InfoContextf(c, "corekg permission: set resource scope success")
	}
}

func corekgLibraryResourcePermissionFilterMw() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		ctx.Next(c)

		respBody := ctx.Response.Body()
		if len(respBody) == 0 {
			return
		}

		path := string(ctx.Path())
		switch path {
		case "/api/plugin_api/library_resource_list":
			filterLibraryResourceListByPermission(c, ctx, respBody)
		case "/api/intelligence_api/search/get_draft_intelligence_list":
			filterDraftIntelligenceListByPermission(c, ctx, respBody)
		case "/api/workflow_api/workflow_list":
			filterWorkflowListByPermission(c, ctx, respBody)
		}
	}
}

func filterLibraryResourceListByPermission(ctx context.Context, reqCtx *app.RequestContext, respBody []byte) {
	var resp resource.LibraryResourceListResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		logs.WarnContextf(ctx, "corekg permission: unmarshal library resource list failed: %v", err)
		return
	}
	if resp.Code != 0 || len(resp.ResourceList) == 0 {
		return
	}

	userID := ctxutil.GetUIDFromCtx(ctx)
	if userID == nil || *userID == 0 {
		logs.WarnContextf(ctx, "corekg permission: missing user id in context")
		return
	}

	groupedIDs := make(map[permission.ResourceType][]int64)
	for _, res := range resp.ResourceList {
		if res == nil || !res.IsSetResID() || !res.IsSetResType() {
			continue
		}
		resType, ok := mapResourceTypeToPermissionResourceType(permissionResourceTypeSourceLibrary, int64(res.GetResType()))
		if !ok {
			continue
		}
		groupedIDs[resType] = append(groupedIDs[resType], res.GetResID())
	}

	allowedIDs := filterResourceIDsByType(ctx, *userID, groupedIDs, string(permission.ActionRead))

	filtered := make([]*resourceCommon.ResourceInfo, 0, len(resp.ResourceList))
	for _, res := range resp.ResourceList {
		if res == nil || !res.IsSetResID() || !res.IsSetResType() {
			continue
		}
		if _, ok := allowedIDs[res.GetResID()]; ok {
			filtered = append(filtered, res)
		}
	}

	resp.ResourceList = filtered
	data, err := json.Marshal(&resp)
	if err != nil {
		logs.WarnContextf(ctx, "corekg permission: marshal library resource list failed: %v", err)
		return
	}
	reqCtx.Response.SetBodyRaw(data)
}

func filterDraftIntelligenceListByPermission(ctx context.Context, reqCtx *app.RequestContext, respBody []byte) {
	var resp intelligence.GetDraftIntelligenceListResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		logs.WarnContextf(ctx, "corekg permission: unmarshal draft intelligence list failed: %v", err)
		return
	}
	if resp.Code != 0 || resp.Data == nil || len(resp.Data.Intelligences) == 0 {
		return
	}

	userID := ctxutil.GetUIDFromCtx(ctx)
	if userID == nil || *userID == 0 {
		logs.WarnContextf(ctx, "corekg permission: missing user id in context")
		return
	}

	groupedIDs := make(map[permission.ResourceType][]int64)
	for _, item := range resp.Data.Intelligences {
		if item == nil || item.BasicInfo == nil {
			continue
		}
		resourceID := item.BasicInfo.GetID()
		if resourceID <= 0 {
			continue
		}
		resourceType, ok := mapResourceTypeToPermissionResourceType(permissionResourceTypeSourceIntelligence, int64(item.GetType()))
		if !ok {
			continue
		}
		groupedIDs[resourceType] = append(groupedIDs[resourceType], resourceID)
	}

	allowedIDs := filterResourceIDsByType(ctx, *userID, groupedIDs, string(permission.ActionRead))

	filtered := make([]*intelligence.IntelligenceData, 0, len(resp.Data.Intelligences))
	for _, item := range resp.Data.Intelligences {
		if item == nil || item.BasicInfo == nil {
			continue
		}
		resourceID := item.BasicInfo.GetID()
		if _, ok := allowedIDs[resourceID]; ok {
			filtered = append(filtered, item)
		}
	}

	resp.Data.Intelligences = filtered
	resp.Data.Total = int32(len(filtered))
	data, err := json.Marshal(&resp)
	if err != nil {
		logs.WarnContextf(ctx, "corekg permission: marshal draft intelligence list failed: %v", err)
		return
	}
	reqCtx.Response.SetBodyRaw(data)
}

func filterWorkflowListByPermission(ctx context.Context, reqCtx *app.RequestContext, respBody []byte) {
	var resp workflow.GetWorkFlowListResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		logs.WarnContextf(ctx, "corekg permission: unmarshal workflow list failed: %v", err)
		return
	}
	if resp.Code != 0 || resp.Data == nil || len(resp.Data.WorkflowList) == 0 {
		return
	}

	userID := ctxutil.GetUIDFromCtx(ctx)
	if userID == nil || *userID == 0 {
		logs.WarnContextf(ctx, "corekg permission: missing user id in context")
		return
	}

	workflowIDMap := make(map[string]int64, len(resp.Data.WorkflowList))
	workflowIDs := make([]int64, 0, len(resp.Data.WorkflowList))
	for _, item := range resp.Data.WorkflowList {
		if item == nil {
			continue
		}
		workflowID := item.GetWorkflowID()
		if workflowID == "" {
			continue
		}
		id, err := strconv.ParseInt(workflowID, 10, 64)
		if err != nil {
			logs.WarnContextf(ctx, "corekg permission: invalid workflow id '%s': %v", workflowID, err)
			continue
		}
		workflowIDMap[workflowID] = id
		workflowIDs = append(workflowIDs, id)
	}

	groupedIDs := make(map[permission.ResourceType][]int64)
	groupedIDs[permission.ResourceTypeWorkflow] = workflowIDs
	allowedIDs := filterResourceIDsByType(ctx, *userID, groupedIDs, string(permission.ActionRead))

	filteredWorkflowList := make([]*workflow.Workflow, 0, len(resp.Data.WorkflowList))
	for _, item := range resp.Data.WorkflowList {
		if item == nil {
			continue
		}
		id, ok := workflowIDMap[item.GetWorkflowID()]
		if !ok {
			continue
		}
		if _, ok = allowedIDs[id]; ok {
			filteredWorkflowList = append(filteredWorkflowList, item)
		}
	}

	filteredAuthList := make([]*workflow.ResourceAuthInfo, 0, len(resp.Data.AuthList))
	for _, item := range resp.Data.AuthList {
		if item == nil {
			continue
		}
		id, err := strconv.ParseInt(item.GetWorkflowID(), 10, 64)
		if err != nil {
			continue
		}
		if _, ok := allowedIDs[id]; ok {
			filteredAuthList = append(filteredAuthList, item)
		}
	}

	resp.Data.WorkflowList = filteredWorkflowList
	resp.Data.AuthList = filteredAuthList
	resp.Data.Total = int64(len(filteredWorkflowList))

	data, err := json.Marshal(&resp)
	if err != nil {
		logs.WarnContextf(ctx, "corekg permission: marshal workflow list failed: %v", err)
		return
	}
	reqCtx.Response.SetBodyRaw(data)
}

func filterResourceIDsByType(ctx context.Context, operatorID int64, groupedIDs map[permission.ResourceType][]int64, action string) map[int64]struct{} {
	allowed := make(map[int64]struct{})
	for resType, ids := range groupedIDs {
		allowedIDs, err := requestyygu.FilterCoreKGResourceIDsByScopePermission(ctx, operatorID, int64(resType), ids, action)
		if err != nil {
			logs.WarnContextf(ctx, "corekg permission: filter resource ids failed: %v (operator_id=%d resource_type=%d action=%s resource_ids=%v)", err, operatorID, int64(resType), action, ids)
			continue
		}
		if len(allowedIDs) == 0 {
			continue
		}
		for id := range allowedIDs {
			allowed[id] = struct{}{}
		}
	}
	return allowed
}

type permissionResourceTypeSource string

const (
	permissionResourceTypeSourceLibrary      permissionResourceTypeSource = "library"
	permissionResourceTypeSourceIntelligence permissionResourceTypeSource = "intelligence"
)

func mapResourceTypeToPermissionResourceType(source permissionResourceTypeSource, resourceType int64) (permission.ResourceType, bool) {
	switch source {
	case permissionResourceTypeSourceLibrary:
		switch resourceCommon.ResType(resourceType) {
		case resourceCommon.ResType_Plugin:
			return permission.ResourceTypePlugin, true
		case resourceCommon.ResType_Workflow:
			return permission.ResourceTypeWorkflow, true
		case resourceCommon.ResType_Knowledge:
			return permission.ResourceTypeKnowledge, true
		case resourceCommon.ResType_Prompt:
			return permission.ResourceTypePrompt, true
		case resourceCommon.ResType_Database:
			return permission.ResourceTypeDatabase, true
		case resourceCommon.ResType_UI:
			return permission.ResourceTypeUI, true
		default:
			return 0, false
		}
	case permissionResourceTypeSourceIntelligence:
		switch intelligenceCommon.IntelligenceType(resourceType) {
		case intelligenceCommon.IntelligenceType_Bot:
			return permission.ResourceTypeAgent, true
		case intelligenceCommon.IntelligenceType_Project:
			return permission.ResourceTypeProject, true
		default:
			return 0, false
		}
	default:
		return 0, false
	}
}

type draftBotCreateResp struct {
	Code int64               `json:"code"`
	Data *draftBotCreateData `json:"data"`
	Msg  string              `json:"msg"`
}

type draftBotCreateData struct {
	BotID json.RawMessage `json:"bot_id"`
}

type produceBotCreateResp struct {
	Code int64                 `json:"code"`
	Data *produceBotCreateData `json:"data"`
	Msg  string                `json:"msg"`
}

type produceBotCreateData struct {
	BotID json.RawMessage `json:"bot_id"`
}

type registerPluginMetaResp struct {
	Code     int64           `json:"code"`
	PluginID json.RawMessage `json:"plugin_id"`
}

type createWorkflowResp struct {
	Code int64               `json:"code"`
	Data *createWorkflowData `json:"data"`
}

type createWorkflowData struct {
	WorkflowID json.RawMessage `json:"workflow_id"`
}

type upsertPromptResourceResp struct {
	Code int64                     `json:"code"`
	Data *upsertPromptResourceData `json:"data"`
}

type upsertPromptResourceData struct {
	ID json.RawMessage `json:"id"`
}

func extractCoreKGResource(ctx context.Context, path string, reqBody []byte, respBody []byte) (string, permission.ResourceType, bool) {
	switch path {
	case "/api/draftbot/create":
		return extractDraftBotResource(ctx, respBody)
	case "/api/draftbot/duplicate":
		return extractDraftBotResource(ctx, respBody)
	case "/api/playground_api/produce/create_bot":
		return extractProduceBotResource(ctx, respBody)
	case "/api/plugin_api/register_plugin_meta":
		return extractPluginResource(ctx, respBody)
	case "/api/workflow_api/create":
		return extractWorkflowResource(ctx, respBody)
	case "/api/workflow_api/copy":
		return extractWorkflowResource(ctx, respBody)
	case "/api/playground_api/upsert_prompt_resource":
		return extractPromptResource(ctx, reqBody, respBody)
	default:
		return "", 0, false
	}
}

func extractDraftBotResource(ctx context.Context, body []byte) (string, permission.ResourceType, bool) {
	var resp draftBotCreateResp
	if err := json.Unmarshal(body, &resp); err != nil {
		logs.InfoContextf(ctx, "corekg permission: unmarshal response failed: %v", err)
		return "", 0, false
	}

	if resp.Code != 0 {
		logs.InfoContextf(ctx, "corekg permission: draftbot create failed (code=%d)", resp.Code)
		return "", 0, false
	}

	if resp.Data == nil || len(resp.Data.BotID) == 0 {
		return "", 0, false
	}

	resourceID, ok := parseJSONID(resp.Data.BotID)
	if !ok {
		return "", 0, false
	}

	return resourceID, permission.ResourceTypeAgent, true
}

func extractProduceBotResource(ctx context.Context, body []byte) (string, permission.ResourceType, bool) {
	var resp produceBotCreateResp
	if err := json.Unmarshal(body, &resp); err != nil {
		logs.InfoContextf(ctx, "corekg permission: unmarshal produce bot response failed: %v", err)
		return "", 0, false
	}

	if resp.Code != 0 {
		logs.InfoContextf(ctx, "corekg permission: produce bot create failed (code=%d)", resp.Code)
		return "", 0, false
	}

	if resp.Data == nil || len(resp.Data.BotID) == 0 {
		return "", 0, false
	}

	resourceID, ok := parseJSONID(resp.Data.BotID)
	if !ok {
		return "", 0, false
	}

	return resourceID, permission.ResourceTypeAgent, true
}

func extractPluginResource(ctx context.Context, body []byte) (string, permission.ResourceType, bool) {
	var resp registerPluginMetaResp
	if err := json.Unmarshal(body, &resp); err != nil {
		logs.InfoContextf(ctx, "corekg permission: unmarshal plugin response failed: %v", err)
		return "", 0, false
	}

	if resp.Code != 0 {
		logs.InfoContextf(ctx, "corekg permission: register plugin meta failed (code=%d)", resp.Code)
		return "", 0, false
	}

	if len(resp.PluginID) == 0 {
		return "", 0, false
	}

	resourceID, ok := parseJSONID(resp.PluginID)
	if !ok {
		return "", 0, false
	}

	return resourceID, permission.ResourceTypePlugin, true
}

func extractWorkflowResource(ctx context.Context, body []byte) (string, permission.ResourceType, bool) {
	var resp createWorkflowResp
	if err := json.Unmarshal(body, &resp); err != nil {
		logs.InfoContextf(ctx, "corekg permission: unmarshal workflow response failed: %v", err)
		return "", 0, false
	}

	if resp.Code != 0 {
		logs.InfoContextf(ctx, "corekg permission: create workflow failed (code=%d)", resp.Code)
		return "", 0, false
	}

	if resp.Data == nil || len(resp.Data.WorkflowID) == 0 {
		return "", 0, false
	}

	resourceID, ok := parseJSONID(resp.Data.WorkflowID)
	if !ok {
		return "", 0, false
	}

	return resourceID, permission.ResourceTypeWorkflow, true
}

func extractPromptResource(ctx context.Context, reqBody []byte, respBody []byte) (string, permission.ResourceType, bool) {
	if !shouldCreatePromptPermission(ctx, reqBody) {
		return "", 0, false
	}

	var resp upsertPromptResourceResp
	if err := json.Unmarshal(respBody, &resp); err != nil {
		logs.InfoContextf(ctx, "corekg permission: unmarshal prompt response failed: %v", err)
		return "", 0, false
	}

	if resp.Code != 0 {
		logs.InfoContextf(ctx, "corekg permission: upsert prompt resource failed (code=%d)", resp.Code)
		return "", 0, false
	}

	if resp.Data == nil || len(resp.Data.ID) == 0 {
		return "", 0, false
	}

	resourceID, ok := parseJSONID(resp.Data.ID)
	if !ok {
		return "", 0, false
	}

	return resourceID, permission.ResourceTypePrompt, true
}

func shouldCreatePromptPermission(ctx context.Context, body []byte) bool {
	if len(body) == 0 {
		logs.InfoContextf(ctx, "corekg permission: empty prompt request body")
		return false
	}

	var req struct {
		Prompt json.RawMessage `json:"prompt"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		logs.InfoContextf(ctx, "corekg permission: unmarshal prompt request failed: %v", err)
		return false
	}

	if len(req.Prompt) == 0 {
		return false
	}

	var prompt map[string]json.RawMessage
	if err := json.Unmarshal(req.Prompt, &prompt); err != nil {
		logs.InfoContextf(ctx, "corekg permission: unmarshal prompt body failed: %v", err)
		return false
	}

	if raw, exists := prompt["id"]; exists {
		return isZeroJSONID(raw)
	}

	return true
}

func isZeroJSONID(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return true
	}

	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return true
	}

	var s string
	if err := json.Unmarshal(trimmed, &s); err == nil {
		return s == "" || s == "0"
	}

	var n json.Number
	if err := json.Unmarshal(trimmed, &n); err == nil {
		if n.String() == "" {
			return true
		}
		if v, err := strconv.ParseInt(n.String(), 10, 64); err == nil {
			return v == 0
		}
		if v, err := strconv.ParseFloat(n.String(), 64); err == nil {
			return v == 0
		}
		return false
	}

	return false
}

func parseJSONID(raw json.RawMessage) (string, bool) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if s != "" {
			return s, true
		}
	}

	var n json.Number
	if err := json.Unmarshal(raw, &n); err == nil {
		if n.String() != "" {
			return n.String(), true
		}
	}

	return "", false
}
