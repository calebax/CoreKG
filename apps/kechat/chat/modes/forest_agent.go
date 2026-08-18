package modes

import (
	"context"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/chat/core"
	forestagentprompt "github.com/insmtx/corekg/apps/kechat/chat/prompt/forest_agent"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	kesearchtools "github.com/insmtx/corekg/apps/kesearch/pkg/ai/tools"
	ygagent "github.com/insmtx/corekg/pkgs/einotools/agent"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/insmtx/corekg/pkgs/einotools/printer"
	ygagentservice "github.com/insmtx/corekg/pkgs/einotools/service"
	ygagenttools "github.com/insmtx/corekg/pkgs/einotools/tools"
	"github.com/insmtx/corekg/pkgs/einotools/utils"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
)

type ForestAgentChatMode struct {
	ctx *gin.Context
}

func NewForestAgentChatMode(ctx *gin.Context) *ForestAgentChatMode {
	return &ForestAgentChatMode{ctx: ctx}
}

func (c *ForestAgentChatMode) Run(ctx context.Context, ctxData *core.ChatContext) (*core.ChatResult, error) {
	history, err := getForestChatHistory(c.ctx, ctxData.Session)
	if err != nil {
		logs.ErrorContextf(c.ctx, "[ForestAgentChatMode] getForestChatHistory error: %v", err)
		return nil, err
	}

	sharedRefs := &kesearchtools.SharedReferences{
		Refs: make([]*chattype.QueryReference, 0),
	}
	forestSearchTool, err := kesearchtools.NewForestSearchTool(ctx, &kesearchtools.ForestSearchToolConfig{
		Ctx:              c.ctx,
		EsIndex:          ctxData.Session.EsIndex,
		ForestIDs:        ctxData.Session.ForestIDList.Slice(),
		FileIDs:          ctxData.Session.FileIDList.Slice(),
		OriginalQuestion: ctxData.Question.Source.Question,
		ReferencesResult: sharedRefs,
	})
	if err != nil {
		logs.ErrorContextf(c.ctx, "[ForestAgentChatMode] NewForestSearchTool error: %v", err)
		return nil, err
	}

	inputFiles, err := c.buildInputFiles(ctxData.Session.FileIDList.Slice())
	if err != nil {
		logs.ErrorContextf(c.ctx, "[ForestAgentChatMode] buildInputFiles error: %v", err)
		return nil, err
	}

	agentReq := &models.AgentRequest{
		SessionID:  ctxData.Session.ID,
		RequestID:  ctxData.Question.Source.ReqID,
		Query:      ctxData.Question.Source.Question,
		IsStream:   true,
		InputFiles: inputFiles,
		Messages:   history,
	}

	msgPrinter := core.GetPrinter(ctxData.Extra)
	if msgPrinter == nil {
		sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()),
			sseclient.WithExpiration(time.Minute*5))
		msgPrinter = printer.NewSSEPrinter(agentReq, sseClient, c.ctx.Writer)
		defer msgPrinter.Close(ctx)
	}

	chatModel, err := newToolCallingChatModel(
		c.ctx,
		ctxData.Model,
		ctxData.ModelOptions,
	)
	if err != nil {
		logs.ErrorContextf(ctx, "[ForestAgentChatMode] failed to create OpenAiChatModel: %v", err)
		return nil, err
	}

	agentContext := &ygagent.AgentContext{}
	agentContext.SessionID = agentReq.SessionID
	agentContext.RequestID = agentReq.RequestID
	agentContext.Query = agentReq.Query
	agentContext.IsStream = agentReq.IsStream
	agentContext.MaxStep = 6
	agentContext.AgentRequest = agentReq
	agentContext.Printer = msgPrinter
	agentContext.DateInfo = utils.CurrentDateInfoRFC1123()
	agentContext.SystemPrompt = forestagentprompt.ForestAgentSystemPrompt
	agentContext.NextStepPrompt = forestagentprompt.ForestAgentNextStepPrompt
	agentContext.SummarySystemPrompt = forestagentprompt.BuildForestAgentSummarySystemPrompt(
		forestagentprompt.SummaryPromptOptions{
			EnableReference: core.GetEnableReference(ctxData.Extra),
			ExtraPrompt:     core.GetSummarySystemPrompt(ctxData.Extra),
		},
	)
	agentContext.ChatModel = chatModel
	agentContext.Tools = []ygagenttools.ToolOption{
		ygagenttools.ToolOptionCode,
		ygagenttools.ToolOptionFile,
	}
	agentContext.AvailableTools = []tool.BaseTool{
		ygagenttools.WrapToolWithErrorHandling(ctx, forestSearchTool),
	}

	reactAgentService := &ygagentservice.ReactAgentService{}
	result, err := reactAgentService.Handler(ctx, agentContext, agentReq,
		models.WithSummaryMode(true),
		models.WithAgentStageStreamMode(false, true),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "[ForestAgentChatMode] Agent Handler error: %v", err)
		return nil, err
	}

	chatResult := &core.ChatResult{
		Answer: result,
		Status: chattype.QuestionStatusAnswered,
	}

	refs := sharedRefs.GetAggregated()
	chunkList := chattype.ChatReferenceList{}
	for _, v := range refs {
		chunks := map[int]string{}
		for _, chunk := range v.ChunkList {
			chunks[chunk.Sequence] = chunk.Content
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
	chatResult.Usage.PromptTokens = reactAgentService.Stats.TotalUsage.PromptTokens
	chatResult.Usage.CompletionTokens = reactAgentService.Stats.TotalUsage.CompletionTokens
	chatResult.Usage.CacheHitTokens = reactAgentService.Stats.TotalUsage.PromptCacheHitTokens
	chatResult.Usage.CacheMissTokens = reactAgentService.Stats.TotalUsage.PromptCacheMissTokens
	chatResult.Usage.TotalTokens = reactAgentService.Stats.TotalUsage.TotalTokens
	chatResult.Performance.CostSeconds = int(reactAgentService.Stats.DurationMs / 1000)

	return chatResult, nil
}

func (c *ForestAgentChatMode) buildInputFiles(fileIDs []uint) ([]models.File, error) {
	if len(fileIDs) == 0 {
		return nil, nil
	}

	forestFileEntityList, err := forest.NewForestFileDao().GetListByCond(c.ctx, &forest.ForestFileCond{
		IDs:         fileIDs,
		ParseStatus: foresttype.TaskStatusSuccess,
	})
	if err != nil {
		return nil, err
	}

	inputFiles := make([]models.File, 0, len(forestFileEntityList))
	for _, v := range forestFileEntityList {
		inputFile := models.File{
			FileName: v.Name,
			FileSize: v.Size,
		}
		fileURL, useParsedMarkdown, err := c.resolveInputFileURL(v)
		if err != nil {
			return nil, err
		}
		if useParsedMarkdown {
			inputFile.Type = "text"
			inputFile.Description = "文件解析后的 markdown 内容"
		}
		inputFile.FileURL = fileURL
		inputFiles = append(inputFiles, inputFile)
	}

	return inputFiles, nil
}

func shouldUseParsedMarkdown(file foresttype.KnownowForestFile) bool {
	switch strings.ToLower(file.Ext) {
	case ".xlsx", ".xls", ".csv":
		return false
	default:
		return file.ParseStatus == foresttype.TaskStatusSuccess
	}
}

func (c *ForestAgentChatMode) resolveInputFileURL(file foresttype.KnownowForestFile) (string, bool, error) {
	if shouldUseParsedMarkdown(file) {
		contentPath := fs.FileContentPath(file.GetAlgoFilePath(), file.ID)
		contentFile, err := fs.Forest.ReadFile(contentPath)
		if err == nil {
			contentFile.Close()
			return fs.Forest.GetPublicURL(contentPath, false), true, nil
		}
		logs.WarnContextf(c.ctx, "[ForestAgentChatMode] parsed markdown not found, fallback to preview file, file_id=%d, err=%v", file.ID, err)
	}

	filePath, err := file.GetForestPriviewFilePath()
	if err != nil {
		return "", false, err
	}
	return fs.Forest.GetPublicURL(*filePath, false), false, nil
}
