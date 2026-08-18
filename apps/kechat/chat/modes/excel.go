package modes

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/chat/core"

	excelprompt "github.com/insmtx/corekg/apps/kechat/chat/prompt/excel"
	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	ygagent "github.com/insmtx/corekg/pkgs/einotools/agent"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/insmtx/corekg/pkgs/einotools/printer"
	agentservice "github.com/insmtx/corekg/pkgs/einotools/service"
	"github.com/insmtx/corekg/pkgs/einotools/tools"
	"github.com/insmtx/corekg/pkgs/einotools/utils"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
)

type ExcelChatMode struct {
	ctx *gin.Context
}

func NewExcelChatMode(ctx *gin.Context) *ExcelChatMode {
	return &ExcelChatMode{ctx: ctx}
}

func (c *ExcelChatMode) Run(ctx context.Context, ctxData *core.ChatContext) (*core.ChatResult, error) {
	forestFileEntityList, err := forest.NewForestFileDao().GetListByCond(c.ctx, &forest.ForestFileCond{
		IDs:         ctxData.Session.ExcelIDList.Slice(),
		ParseStatus: foresttype.TaskStatusSuccess,
	})
	if err != nil {
		logs.ErrorContextf(c.ctx, "[ReactExcelChatMode] get file list failed, err: %v", err)
		return nil, err
	}

	inputFiles := make([]models.File, 0, len(forestFileEntityList))
	for _, v := range forestFileEntityList {
		file := models.File{
			FileName: v.Name,
			FileSize: v.Size,
		}
		filePath, err := v.GetForestPriviewFilePath()
		if err != nil {
			logs.ErrorContextf(c.ctx, "[ReactExcelChatMode] GetForestPriviewFilePath error %v", err)
			return nil, err
		}
		file.FileURL = fs.Forest.GetPublicURL(*filePath, false)
		inputFiles = append(inputFiles, file)
	}

	sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()),
		sseclient.WithExpiration(time.Minute*5))
	defer sseClient.Close(c.ctx, runtime.RequestID(c.ctx))

	msgs, err := getForestChatHistory(c.ctx, ctxData.Session)
	if err != nil {
		logs.ErrorContextf(c.ctx, "[ReactExcelChatMode] getReactHistory failed, err: %v", err)
		return nil, err
	}

	agentReq := &models.AgentRequest{
		SessionID:  ctxData.Session.ID,
		RequestID:  ctxData.Question.Source.ReqID,
		Query:      ctxData.Question.Source.Question,
		IsStream:   true,
		InputFiles: inputFiles,
		Messages:   msgs,
	}
	msgPrinter := core.GetPrinter(ctxData.Extra)
	if msgPrinter == nil {
		msgPrinter = printer.NewSSEPrinter(agentReq, sseClient, c.ctx.Writer)
		defer msgPrinter.Close(ctx)
	}

	chatModel, err := newToolCallingChatModel(
		c.ctx,
		ctxData.Model,
		chatModelOptions{},
	)
	if err != nil {
		logs.ErrorContextf(c.ctx, "[ReactExcelChatMode] failed to create OpenAiChatModel: %v", err)
		return nil, err
	}

	agentContext := &ygagent.AgentContext{}
	agentContext.SessionID = agentReq.SessionID
	agentContext.RequestID = agentReq.RequestID
	agentContext.Query = agentReq.Query
	agentContext.IsStream = agentReq.IsStream
	agentContext.AgentRequest = agentReq
	agentContext.Printer = msgPrinter
	agentContext.DateInfo = utils.CurrentDateInfoRFC1123()
	agentContext.SystemPrompt = excelprompt.SystemPrompt
	agentContext.ChatModel = chatModel
	agentContext.Tools = []tools.ToolOption{
		tools.ToolOptionCode,
		tools.ToolOptionFile,
		tools.ToolOptionChart,
	}
	agentContext.SaveChartFunc = func(content string) (uint, error) {
		chartEntity := &chattype.ChatChart{
			RequestID:    runtime.RequestID(c.ctx),
			SessionID:    ctxData.Session.ID,
			QuestionID:   ctxData.Question.ID,
			CompanyID:    runtime.CompanyID(c.ctx),
			Uin:          runtime.Uin(c.ctx),
			ChartContent: content,
		}
		if ctxData.Session.SubjectID > 0 {
			chartEntity.SubjectID = ctxData.Session.SubjectID
			chartEntity.SubjectType = chattype.SessionSubjectTypeProject
		}
		if err := chat.NewChatChartDao().Insert(c.ctx, chartEntity); err != nil {
			logs.ErrorContextf(c.ctx, "[ReactExcelChatMode] insert chart error: %v", err)
			return 0, err
		}
		return chartEntity.ID, nil
	}

	reactAgentService := &agentservice.ReactAgentService{}
	result, err := reactAgentService.Handler(c.ctx, agentContext, agentReq, models.WithSummaryMode(true))
	if err != nil {
		logs.ErrorContextf(c.ctx, "[ReactExcelChatMode] Handler error: %v", err)
		return nil, err
	}
	logs.InfoContextf(c.ctx, "result: %s", result)

	chatResult := &core.ChatResult{
		Answer: result,
		Status: chattype.QuestionStatusAnswered,
		Meta: core.ChatResultMeta{
			core.MetaKeyAgentService: reactAgentService,
		},
	}
	// 填充用量与耗时
	chatResult.Usage.PromptTokens = reactAgentService.Stats.TotalUsage.PromptTokens
	chatResult.Usage.CompletionTokens = reactAgentService.Stats.TotalUsage.CompletionTokens
	chatResult.Usage.CacheHitTokens = reactAgentService.Stats.TotalUsage.PromptCacheHitTokens
	chatResult.Usage.CacheMissTokens = reactAgentService.Stats.TotalUsage.PromptCacheMissTokens
	chatResult.Usage.TotalTokens = reactAgentService.Stats.TotalUsage.TotalTokens
	chatResult.Performance.CostSeconds = int(reactAgentService.Stats.DurationMs / 1000)

	return chatResult, nil
}
