package modes

import (
	"context"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/chat/core"
	"github.com/insmtx/corekg/apps/kechat/chat/prompt"
	forestprompt "github.com/insmtx/corekg/apps/kechat/chat/prompt/forest"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/services/devkeywords"
	"github.com/insmtx/corekg/apps/kesearch/pkg/ai/tools"
	ygagent "github.com/insmtx/corekg/pkgs/einotools/agent"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/insmtx/corekg/pkgs/einotools/printer"
	ygagentservice "github.com/insmtx/corekg/pkgs/einotools/service"
	ygagenttools "github.com/insmtx/corekg/pkgs/einotools/tools"
	"github.com/insmtx/corekg/pkgs/einotools/utils"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

type ForestChatMode struct {
	ctx *gin.Context
}

func NewForestChatMode(
	ctx *gin.Context,
) *ForestChatMode {
	return &ForestChatMode{ctx}
}

func (c *ForestChatMode) Run(ctx context.Context, ctxData *core.ChatContext) (*core.ChatResult, error) {
	logs.InfoContextf(c.ctx, "[DEBUG][chunk-empty] ForestChatMode.Run start: session_id=%d, es_index=%s, forest_ids=%v, file_ids=%v, model=%s",
		ctxData.Session.ID, ctxData.Session.EsIndex, ctxData.Session.ForestIDList.Slice(), ctxData.Session.FileIDList.Slice(), ctxData.Model.Model)

	// 查询历史记录
	history, err := getForestChatHistory(c.ctx, ctxData.Session)
	if err != nil {
		logs.ErrorContextf(c.ctx, "[forestChatMode] ForestChat GetForestChatHistory error: %v", err)
		return nil, err
	}

	// 知识库检索Tool
	sharedRefs := &tools.SharedReferences{
		Refs: make([]*chattype.QueryReference, 0),
	}
	forestSearchTool, err := tools.NewForestSearchTool(ctx, &tools.ForestSearchToolConfig{
		Ctx:              c.ctx,
		EsIndex:          ctxData.Session.EsIndex,
		ForestIDs:        ctxData.Session.ForestIDList.Slice(),
		FileIDs:          ctxData.Session.FileIDList.Slice(),
		OriginalQuestion: ctxData.Question.Source.Question,
		ReferencesResult: sharedRefs,
	})
	if err != nil {
		logs.ErrorContextf(c.ctx, "[forestChatMode] ForestChat NewForestSearchTool error: %v", err)
		return nil, err
	}

	agentReq := &models.AgentRequest{
		SessionID: ctxData.Session.ID,
		RequestID: ctxData.Question.Source.ReqID,
		// TODO 修改传入的Question 值
		Query:    ctxData.Question.Source.Question,
		IsStream: true,
		Messages: history,
	}

	msgPrinter := core.GetPrinter(ctxData.Extra)
	if msgPrinter == nil {
		sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()),
			sseclient.WithExpiration(time.Minute*5))
		msgPrinter = printer.NewSSEPrinter(agentReq, sseClient, c.ctx.Writer)
		defer msgPrinter.Close(ctx)
	}

	// DEBUG
	// msgPrinter := printer.NewLogPrinter(agentReq)

	chatModel, err := newToolCallingChatModel(
		c.ctx,
		ctxData.Model,
		ctxData.ModelOptions,
	)
	if err != nil {
		logs.ErrorContextf(ctx, "[forestChatMode] failed to create OpenAiChatModel: %v", err)
		return nil, err
	}

	// 改写问题
	rewrite := agentReq.Query
	rewrite = devkeywords.ReplaceSynonymKeywords(ctx, chatModel, ctxData.Session.CompanyID, rewrite)
	rewrite = devkeywords.ReplaceMajorKeywords(ctx, chatModel, ctxData.Session.CompanyID, rewrite)
	rewriteResult, err := sendQuestionRewriteMessage(ctx, msgPrinter, rewrite)
	if err != nil {
		logs.ErrorContextf(c.ctx, "[forestChatMode] ForestChat sendQuestionRewriteMessage error: %v", err)
		return nil, err
	}
	agentReq.Query = rewrite

	///==============
	agentContext := &ygagent.AgentContext{}
	agentContext.SessionID = agentReq.SessionID
	agentContext.RequestID = agentReq.RequestID
	agentContext.Query = agentReq.Query
	agentContext.IsStream = agentReq.IsStream
	agentContext.MaxStep = 4
	agentContext.AgentRequest = agentReq
	agentContext.Printer = msgPrinter
	agentContext.DateInfo = utils.CurrentDateInfoRFC1123()

	roleName, err := settings.GetText(global.SettingGroupCoreKG, global.SettingKeyLlmRoleName)
	if roleName == "" || err != nil {
		roleName = global.DefaultLlmRoleName
	}
	prompts := prompt.GetKeQAPrompts(roleName)
	temp, ok := prompts[ctxData.Session.PromptMode]
	if !ok {
		temp = prompts["normal"]
	}
	agentContext.SystemPrompt = strings.Replace(forestprompt.ForestKnowledgeSystemPrompt, "[DYNAMIC_STYLE]", temp, 1)
	agentContext.NextStepPrompt = forestprompt.ForestKnowledgeNextStepPrompt
	agentContext.SummarySystemPrompt = strings.Replace(forestprompt.ForestKnowledgeSummarySystemPrompt, "[DYNAMIC_STYLE]", temp, 1)

	agentContext.ChatModel = chatModel
	agentContext.AvailableTools = []tool.BaseTool{
		ygagenttools.WrapToolWithErrorHandling(ctx, forestSearchTool),
	}

	// 写入searching
	msgPrinter.Send(ctx, "", models.MsgTypeExecFlag, models.FlagSearching, true)

	reactAgentService := &ygagentservice.ReactAgentService{}
	result, err := reactAgentService.Handler(ctx, agentContext, agentReq, models.WithSummaryMode(true))
	if err != nil {
		logs.ErrorContextf(ctx, "[forestChatMode] Agent Handler error: %v", err)
		return nil, err
	}
	reactAgentService.Rresult = append(reactAgentService.Rresult, rewriteResult)
	logs.InfoContextf(ctx, "agentContext: %v", len(reactAgentService.Memory.Messages))
	logs.InfoContextf(ctx, "result: %s", result)

	chatResult := &core.ChatResult{
		Answer: result,
		Status: chattype.QuestionStatusAnswered,
	}

	refs := sharedRefs.GetAggregated()
	logs.InfoContextf(c.ctx, "[DEBUG][chunk-empty] ForestChatMode.Run after agent: refs_len=%d, answer_len=%d", len(refs), len(result))

	totalChunks := 0
	for _, r := range refs {
		totalChunks += len(r.ChunkList)
	}
	logs.InfoContextf(c.ctx, "[DEBUG][chunk-empty] ForestChatMode.Run aggregated: file_count=%d, total_chunks=%d", len(refs), totalChunks)

	chunkList := chattype.ChatReferenceList{}
	for _, v := range refs {
		chunks := map[int]string{}
		for _, c := range v.ChunkList {
			chunks[c.Sequence] = c.Content
		}
		chunkList.Reference = append(chunkList.Reference, chattype.ChatReference{
			FileID:   v.FileID,
			Abstract: v.Abstract,
			Chunks:   chunks,
		})
	}

	queryRefList := chattype.QueryReferenceList(refs)
	chatResult.Meta = core.ChatResultMeta{
		core.MetaKeyAgentService:    reactAgentService,
		core.MetaKeyQueryReferences: &queryRefList,
		core.MetaKeyChatReferences:  &chunkList,
	}

	// 设置Token使用情况
	chatResult.Usage.PromptTokens = reactAgentService.Stats.TotalUsage.PromptTokens
	chatResult.Usage.CompletionTokens = reactAgentService.Stats.TotalUsage.CompletionTokens
	chatResult.Usage.CacheHitTokens = reactAgentService.Stats.TotalUsage.PromptCacheHitTokens
	chatResult.Usage.CacheMissTokens = reactAgentService.Stats.TotalUsage.PromptCacheMissTokens
	chatResult.Usage.TotalTokens = reactAgentService.Stats.TotalUsage.TotalTokens

	// 设置时间统计
	// TODO
	// chatResult.Performance.ReasoningSeconds = res.ReasoningTime
	chatResult.Performance.CostSeconds = int(reactAgentService.Stats.DurationMs / 1000)

	return chatResult, nil
}
