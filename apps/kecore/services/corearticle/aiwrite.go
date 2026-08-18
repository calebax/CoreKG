package corearticle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/chat/modelhelper"
	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kesearch/pkg/ai/tools"
	"github.com/insmtx/corekg/apps/kesearch/services/svcessearch"
	ygagent "github.com/insmtx/corekg/pkgs/einotools/agent"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/insmtx/corekg/pkgs/einotools/printer"
	ygagentservice "github.com/insmtx/corekg/pkgs/einotools/service"
	"github.com/insmtx/corekg/pkgs/einotools/utils"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
)

type AIWriteExecutor struct {
	ctx         context.Context
	ginCtx      *gin.Context
	articleID   uint
	cmd         foresttype.CmdString
	content     string
	forestIDs   types.UintArray
	requestID   string
	companyID   uint
	uin         uint
	chatModelID uint
}

type AIWriteExecutorParams struct {
	Ctx         context.Context
	GinCtx      *gin.Context
	ArticleID   uint
	Cmd         foresttype.CmdString
	Content     string
	ForestIDs   types.UintArray
	RequestID   string
	CompanyID   uint
	Uin         uint
	ChatModelID uint
}

func NewAIWriteExecutor(params AIWriteExecutorParams) *AIWriteExecutor {
	return &AIWriteExecutor{
		ctx:         params.Ctx,
		ginCtx:      params.GinCtx,
		articleID:   params.ArticleID,
		cmd:         params.Cmd,
		content:     params.Content,
		forestIDs:   params.ForestIDs,
		requestID:   params.RequestID,
		companyID:   params.CompanyID,
		uin:         params.Uin,
		chatModelID: params.ChatModelID,
	}
}

func (e *AIWriteExecutor) Execute(writer io.Writer) error {
	logs.DebugContextf(e.ctx, "AIWriteExecutor cmd=%v, forestIDs=%v", e.cmd, e.forestIDs)

	var chatModelEntity *chattype.ChatModel
	var err error

	if e.chatModelID > 0 {
		chatModelEntity, err = chat.NewChatModelDao().GetByID(e.ctx, e.chatModelID)
		if err != nil {
			logs.ErrorContextf(e.ctx, "AIWriteExecutor GetByID err: %v", err)
			return fmt.Errorf("get chat model by id failed: %w", err)
		}
	} else {
		chatModelEntity, err = chat.NewChatModelDao().GetByCond(e.ctx, &chat.ChatModelCond{
			BaseCond: chat.BaseCond{
				OrderBy: []string{"priority desc"},
			},
			PublicType: chattype.PublecTypeSystem,
		})
		if err != nil {
			logs.ErrorContextf(e.ctx, "AIWriteExecutor GetByCond err: %v", err)
			return fmt.Errorf("get chat model by condition failed: %w", err)
		}
	}

	temperature := float32(0.5)
	toolCallingChatModel, err := modelhelper.NewToolCallingChatModel(e.ctx, chatModelEntity, modelhelper.ToolCallingChatModelOptions{
		Temperature: &temperature,
	})
	if err != nil {
		logs.ErrorContextf(e.ctx, "AIWriteExecutor NewToolCallingChatModel err: %v", err)
		return fmt.Errorf("create chat model failed: %w", err)
	}

	agentReq := &models.AgentRequest{
		RequestID: e.requestID,
		Query:     e.content,
		IsStream:  true,
	}

	sseClient := sseclient.New(
		sseclient.WithRedisClient(redispool.Redis()),
		sseclient.WithExpiration(time.Minute*5),
	)
	msgPrinter := printer.NewSSEPrinter(agentReq, sseClient, writer)
	defer msgPrinter.Close(e.ctx)

	hasReference := len(e.forestIDs) > 0
	var referenceContent string
	var sharedRefs *tools.SharedReferences

	if hasReference {
		sharedRefs = &tools.SharedReferences{
			Refs: make([]*chattype.QueryReference, 0),
		}
		searchRes, searchErr := svcessearch.RerankSearchQuestionChunk(
			e.ctx, global.EsIndexKGDefault, e.content, e.forestIDs.Slice(), []uint{}, nil, "",
		)
		if searchErr != nil {
			logs.ErrorContextf(e.ctx, "AIWriteExecutor pre-search err: %v", searchErr)
			hasReference = false
		} else {
			sharedRefs.Append(searchRes...)
			referenceContent = formatReferenceForPrompt(searchRes)
		}
	}

	templatePrompt := getPromptByCmd(e.cmd, hasReference)

	promptData := &PromptData{
		Content:     e.content,
		Instruction: string(e.cmd),
		Reference:   referenceContent,
	}
	renderedPrompt, err := renderPrompt(templatePrompt, promptData)
	if err != nil {
		logs.ErrorContextf(e.ctx, "AIWriteExecutor renderPrompt err: %v", err)
		return fmt.Errorf("render prompt failed: %w", err)
	}

	agentContext := &ygagent.AgentContext{}
	agentContext.RequestID = agentReq.RequestID
	agentContext.Query = agentReq.Query
	agentContext.IsStream = agentReq.IsStream
	agentContext.MaxStep = 1
	agentContext.AgentRequest = agentReq
	agentContext.Printer = msgPrinter
	agentContext.DateInfo = utils.CurrentDateInfoRFC1123()
	agentContext.SystemPrompt = renderedPrompt
	agentContext.ChatModel = toolCallingChatModel
	agentContext.AvailableTools = nil

	reactAgentService := &ygagentservice.ReactAgentService{}
	result, err := reactAgentService.Handler(e.ctx, agentContext, agentReq)
	if err != nil {
		logs.ErrorContextf(e.ctx, "AIWriteExecutor ReactAgentService.Handler err: %v", err)
		return fmt.Errorf("execute ai write failed: %w", err)
	}

	logs.DebugContextf(e.ctx, "AIWriteExecutor result=%v", result[:min(100, len(result))])

	resultContent := result
	if sharedRefs != nil && len(sharedRefs.Refs) > 0 {
		refsData, err := json.Marshal(sharedRefs.GetAggregated())
		if err != nil {
			logs.WarnContextf(e.ctx, "AIWriteExecutor marshal refs err: %v", err)
		} else {
			resultContent = result + "\n\n[References]\n" + string(refsData)
		}
	}

	historyEntity := &foresttype.KeArticleHistory{
		ArticleID: e.articleID,
		Cmd:       e.cmd,
		Content:   e.content,
		Result:    resultContent,
		CompanyID: e.companyID,
		Uin:       e.uin,
	}

	if err := forest.NewArticleHistoryDao().Insert(e.ctx, historyEntity); err != nil {
		logs.WarnContextf(e.ctx, "AIWriteExecutor insert article history err: %v", err)
	}

	return nil
}

func formatReferenceForPrompt(refs []*chattype.QueryReference) string {
	type referenceItem struct {
		FileID    uint   `json:"file_id"`
		FileName  string `json:"file_name"`
		ChunkList []struct {
			Sequence int    `json:"sequence"`
			Content  string `json:"content"`
		} `json:"chunk_list"`
	}

	items := make([]referenceItem, 0, len(refs))
	for _, ref := range refs {
		item := referenceItem{
			FileID:   ref.FileID,
			FileName: ref.FileName,
		}
		for _, chunk := range ref.ChunkList {
			item.ChunkList = append(item.ChunkList, struct {
				Sequence int    `json:"sequence"`
				Content  string `json:"content"`
			}{
				Sequence: chunk.Sequence,
				Content:  chunk.Content,
			})
		}
		items = append(items, item)
	}

	data, err := json.Marshal(items)
	if err != nil {
		logs.Warnf("formatReferenceForPrompt marshal err: %v", err)
		return ""
	}
	return string(data)
}
