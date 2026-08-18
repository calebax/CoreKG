package modes

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/chat/core"
	chatprompt "github.com/insmtx/corekg/apps/kechat/chat/prompt/directchat"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/einotools/filecontent"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/insmtx/corekg/pkgs/einotools/printer"
	"github.com/insmtx/corekg/pkgs/einotools/tools"
	"github.com/insmtx/corekg/pkgs/einotools/utils"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

// Constants for default values
const (
	defaultMaxIterations = 20
	defaultTemperature   = 0.4
	sseExpiration        = 5 * time.Minute
	// inlineFileMaxBytes 表示每个附件内联到模型上下文的最大字节数，取 80KB 是为了控制 prompt 体积并保留常见 Markdown 正文的前段信息。
	inlineFileMaxBytes = int64(80_000)
)

type DirectModelChatMode struct {
	ctx *gin.Context
}

func NewDirectModelChatMode(ctx *gin.Context) *DirectModelChatMode {
	return &DirectModelChatMode{ctx: ctx}
}

// Run executes the direct model chat flow
func (c *DirectModelChatMode) Run(ctx context.Context, ctxData *core.ChatContext) (*core.ChatResult, error) {
	// 1. Get chat history
	history, err := getForestChatHistory(ctx, ctxData.Session)
	if err != nil {
		logs.ErrorContextf(c.ctx, "[DirectModelChatMode] Failed to get history: %v", err)
		return nil, err
	}

	// 3. Create Agent Request
	agentReq := c.createAgentRequest(ctx, ctxData, history)

	// 4. Initialize SSE and Printer
	sseClient, msgPrinter := c.initSSEPrinter(agentReq)
	defer sseClient.Close(c.ctx, runtime.RequestID(c.ctx))

	// 5. Initialize Base Agent (Memory & Stats)
	ctx, cancel := context.WithCancel(ctx)
	msgPrinter.SetCancelFunc(cancel)
	ckgAgent := c.initBaseAgent(msgPrinter)

	// 6. Create Chat Model
	chatModel, err := c.createChatModel(ctx, ctxData.Model)
	if err != nil {
		logs.ErrorContextf(c.ctx, "[DirectModelChatMode] Failed to create ChatModel: %v", err)
		return nil, err
	}

	hasInlineContent := false
	hasExternalFiles := false
	for _, inputFile := range agentReq.InputFiles {
		if strings.TrimSpace(inputFile.Content) != "" {
			hasInlineContent = true
			continue
		}
		if strings.TrimSpace(inputFile.FileURL) != "" {
			hasExternalFiles = true
		}
	}

	if hasInlineContent && !hasExternalFiles && !agentReq.Options.EnableWebSearch {
		ckgAgent.FinalSignal.MarkFinal()
		agent, err := c.createADKAgent(ctx, chatModel, ckgAgent.FinalSignal, agentReq, nil)
		if err != nil {
			logs.ErrorContextf(c.ctx, "[DirectModelChatMode]InlineFiles: Failed to create ADK Agent: %v", err)
			return nil, err
		}
		if err = c.runAgentLoop(ctx, agent, agentReq, ckgAgent); err != nil {
			logs.ErrorContextf(c.ctx, "[DirectModelChatMode]InlineFiles: Failed to run agent loop: %v", err)
			return nil, err
		}
		return c.buildChatResult(ckgAgent), nil
	}

	// 7. Create ADK Agent
	includeAnalysisFile := !hasInlineContent || hasExternalFiles
	agentTools, err := c.buildAgentTools(ctx, ckgAgent.FinalSignal, chatModel, agentReq.Options, includeAnalysisFile)
	if err != nil {
		logs.ErrorContextf(c.ctx, "[DirectModelChatMode] Failed to build agent tools: %v", err)
		return nil, err
	}
	agent, err := c.createADKAgent(ctx, chatModel, ckgAgent.FinalSignal, agentReq, agentTools)
	if err != nil {
		logs.ErrorContextf(c.ctx, "[DirectModelChatMode] Failed to create ADK Agent: %v", err)
		return nil, err
	}

	// 8. Run Agent Loop
	reachedMaxStep := false
	err = c.runAgentLoop(ctx, agent, agentReq, ckgAgent)
	if err != nil {
		if isContextCanceled(ctx, err) {
			return c.buildChatResult(ckgAgent), nil
		}

		if errors.Is(err, compose.ErrExceedMaxSteps) {
			// 超过最大迭代,根据当前内容，直接输出
			reachedMaxStep = true
		} else {
			logs.ErrorContextf(c.ctx, "[DirectModelChatMode] Agent execution failed: %v", err)
			return nil, err
		}
	}

	if !ckgAgent.FinalSignal.IsFinal() || reachedMaxStep {
		// 未最终有效输出，根据当前搜集内容，生成答案（异常兜底）
		ckgAgent.FinalSignal.MarkFinal()

		agent, err = c.createADKAgent(ctx, chatModel, ckgAgent.FinalSignal, agentReq, nil)
		if err != nil {
			logs.ErrorContextf(c.ctx, "[DirectModelChatMode]IsFinal: Failed to create ADK Agent: %v", err)
			return nil, err
		}

		err = c.runAgentLoop(ctx, agent, agentReq, ckgAgent)
		if err != nil {
			logs.ErrorContextf(c.ctx, "[DirectModelChatMode]IsFinal: Failed to run agent loop: %v", err)
			return nil, err
		}
	}

	// 9. Build Final Result
	return c.buildChatResult(ckgAgent), nil
}

func (c *DirectModelChatMode) buildSystemMessages(ctx context.Context, roleName string, finalSignal *tools.FinalAnswerSignal, agentReq *models.AgentRequest) ([]*schema.Message, error) {
	systemPrompt := chatprompt.ChineseSystemPrompt
	if finalSignal.IsFinal() {
		systemPrompt = chatprompt.ChineseSystemFinishPrompt
	}
	sysPromptTemplate := prompt.FromMessages(schema.GoTemplate,
		schema.SystemMessage(systemPrompt),
	)

	files, _ := json.Marshal(agentReq.InputFiles)
	params := map[string]any{
		"roleName":   roleName,
		"date":       utils.CurrentDateInfoRFC1123(),
		"query":      agentReq.Query,
		"inputFiles": string(files),
	}

	return sysPromptTemplate.Format(ctx, params)
}

func (c *DirectModelChatMode) createAgentRequest(ctx context.Context, ctxData *core.ChatContext, history []schema.Message) *models.AgentRequest {
	options := &models.AgentOptions{
		EnableWebSearch: false,
	}

	if src := ctxData.Question.Source; src != nil {
		if src.Extra != nil && src.Extra.Agent != nil {
			options = &models.AgentOptions{
				EnableWebSearch: src.Extra.Agent.EnableWebSearch,
			}
		}
	}

	var inputFiles []models.File

	extra := ctxData.Question.Source.Extra
	if extra != nil && extra.Input != nil {
		attachments := extra.Input.Attachments
		inputFiles = make([]models.File, 0, len(attachments))

		for _, a := range attachments {
			fileURL := a.Url
			if a.MdUrl != "" {
				fileURL = a.MdUrl
			}
			description := "文件没有可用 md_url；如 content 为空，表示后端无法直接读取该附件正文。"
			if a.MdUrl != "" {
				description = "文件已由后端解析为 Markdown 正文，content 字段为可用于回答的文件内容。"
			}
			content, truncated := c.readInputContent(ctx, a.Name, fileURL, a.MdUrl != "")
			inputFiles = append(inputFiles, models.File{
				FileName:    a.Name,
				Type:        a.Type,
				FileURL:     fileURL,
				Content:     content,
				Truncated:   truncated,
				Description: description,
			})
		}
	}

	return &models.AgentRequest{
		SessionID:  ctxData.Session.ID,
		RequestID:  ctxData.Question.Source.ReqID,
		Query:      ctxData.Question.Source.Question,
		InputFiles: inputFiles,
		IsStream:   true,
		Messages:   history,
		Options:    options,
	}
}

func (c *DirectModelChatMode) readInputContent(ctx context.Context, fileName, fileURL string, parsed bool) (string, bool) {
	content, _, truncated, err := filecontent.Read(ctx, fileURL, inlineFileMaxBytes)
	if err != nil {
		logs.WarnContextf(ctx, "[DirectModelChatMode] read input content failed, fileName:%s parsed:%t err:%v", fileName, parsed, err)
		return "", false
	}
	if truncated {
		logs.InfoContextf(ctx, "[DirectModelChatMode] inline file content truncated, fileName:%s maxBytes:%d", fileName, inlineFileMaxBytes)
	}
	return content, truncated
}

func (c *DirectModelChatMode) initSSEPrinter(agentReq *models.AgentRequest) (*sseclient.SSEClient, printer.Printer) {
	sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()),
		sseclient.WithExpiration(sseExpiration))

	msgPrinter := printer.NewSSEPrinter(agentReq, sseClient, c.ctx.Writer)
	return sseClient, msgPrinter
}

func (c *DirectModelChatMode) initBaseAgent(msgPrinter printer.Printer) *BaseAgent {
	agent := &BaseAgent{
		Memory:      models.NewMemory(),
		Printer:     msgPrinter,
		MaxStep:     defaultMaxIterations,
		FinalSignal: &tools.FinalAnswerSignal{},
		Stats:       models.AgentStats{},
	}
	agent.Stats.Start()
	return agent
}

func (c *DirectModelChatMode) createChatModel(ctx context.Context, modelCfg *chattype.ChatModel) (model.ToolCallingChatModel, error) {
	return newToolCallingChatModel(
		c.ctx,
		modelCfg,
		chatModelOptions{
			Temperature: ptrFloat32(defaultTemperature),
		},
	)
}

func (c *DirectModelChatMode) buildAgentTools(ctx context.Context, finalSignal *tools.FinalAnswerSignal,
	chatModel model.ToolCallingChatModel, options *models.AgentOptions, includeAnalysisFile bool) ([]tool.BaseTool, error) {

	finalAnswerMarkerTool, err := tools.NewFinalAnswerMarkerTool(ctx, &tools.FinalAnswerMarkerConfig{
		FinalSignal: finalSignal,
	})
	if err != nil {
		return nil, err
	}

	agentTools := []tool.BaseTool{finalAnswerMarkerTool}

	var toolOptions []tools.ToolOption
	if includeAnalysisFile {
		toolOptions = append(toolOptions, tools.ToolOptionAnalysisFile)
	}
	if options.EnableWebSearch {
		toolOptions = append(toolOptions, tools.ToolOptionSearch)
	}
	addTools, err := tools.GetTools(ctx, chatModel, toolOptions, nil)
	if err != nil {
		logs.ErrorContextf(ctx, "[DirectModelChatMode] Failed to get tools: %v", err)
		return nil, err
	}
	agentTools = append(agentTools, addTools...)

	return agentTools, nil
}

func (c *DirectModelChatMode) createADKAgent(
	ctx context.Context,
	chatModel model.ToolCallingChatModel,
	finalSignal *tools.FinalAnswerSignal,
	agentReq *models.AgentRequest,
	agentTools []tool.BaseTool,
) (*adk.ChatModelAgent, error) {
	roleName, err := settings.GetText(global.SettingGroupCoreKG, global.SettingKeyLlmRoleName)
	if roleName == "" || err != nil {
		roleName = global.DefaultLlmRoleName
	}

	customGenModelInput := func(ctx context.Context, instruction string, input *adk.AgentInput) ([]adk.Message, error) {
		// 这里并不会被每轮对话前调用，只是在初始化时调用一次。用来占位确保存在system消息。
		msgs := make([]adk.Message, 0, 1+len(input.Messages))
		msgs = append(msgs, schema.SystemMessage(instruction))
		msgs = append(msgs, input.Messages...)
		return msgs, nil
	}

	modelAgentConfig := &adk.ChatModelAgentConfig{
		Name:          "atlas_agent",
		Description:   "A knowledgeable assistant agent that searches the internet and delivers accurate, well-structured answers.",
		Model:         chatModel,
		GenModelInput: customGenModelInput,
		Middlewares: []adk.AgentMiddleware{
			{
				BeforeChatModel: func(ctx context.Context, state *adk.ChatModelAgentState) error {
					sysMessages, err := c.buildSystemMessages(ctx, roleName, finalSignal, agentReq)
					if err != nil {
						logs.ErrorContextf(c.ctx, "[DirectModelChatMode] Failed to build system messages: %v", err)
						return err
					}
					for i, msg := range state.Messages {
						if msg.Role == schema.System {
							state.Messages[i] = sysMessages[0]
							break
						}
					}
					return nil
				},
			},
		},
		MaxIterations: defaultMaxIterations,
	}

	if agentTools != nil {
		modelAgentConfig.ToolsConfig = adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: agentTools},
		}
	}

	return adk.NewChatModelAgent(ctx, modelAgentConfig)
}

func (c *DirectModelChatMode) runAgentLoop(
	ctx context.Context,
	agent *adk.ChatModelAgent,
	agentReq *models.AgentRequest,
	ckgAgent *BaseAgent,
) error {
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: agentReq.IsStream,
	})

	inputMessages := make([]adk.Message, 0)

	// 历史对话
	historyMessageContent := buildConversationSummary(agentReq.Messages)
	inputMessages = append(inputMessages, schema.UserMessage(historyMessageContent))

	// 本轮短期会话
	if len(ckgAgent.Memory.Messages) == 0 && len(agentReq.Query) != 0 {
		ckgAgent.Memory.AddMessageWithType("user", schema.UserMessage(agentReq.Query))
	}
	inputMessages = append(inputMessages, ckgAgent.Memory.GetLlmMessages()...)

	events := runner.Run(ctx, inputMessages)

	for {
		event, ok := events.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			logs.ErrorContextf(c.ctx, "[DirectModelChatMode] Agent event error: %v", event.Err)
			continue
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			output := event.Output.MessageOutput
			if output.IsStreaming && output.MessageStream != nil {
				ckgAgent.handleStreamMessage(ctx, output)
			} else if output.Message != nil {
				ckgAgent.handleRegularMessage(ctx, output)
			}
		}
	}
	return nil
}

func (c *DirectModelChatMode) buildChatResult(ckgAgent *BaseAgent) *core.ChatResult {
	var answer string
	if ckgAgent.Memory != nil {
		msg := ckgAgent.Memory.GetLastMessage()
		if msg != nil && msg.Payload != nil && msg.Payload.Role == schema.Assistant {
			answer = msg.Payload.Content
		}
	}

	chatResult := &core.ChatResult{
		Answer: answer,
		Status: chattype.QuestionStatusAnswered,
	}

	// TODO 后续agent封装，迁移stop调用
	ckgAgent.Stats.Stop()

	// Usage stats
	chatResult.Usage.PromptTokens = ckgAgent.Stats.TotalUsage.PromptTokens
	chatResult.Usage.CompletionTokens = ckgAgent.Stats.TotalUsage.CompletionTokens
	chatResult.Usage.CacheHitTokens = ckgAgent.Stats.TotalUsage.PromptCacheHitTokens
	chatResult.Usage.CacheMissTokens = ckgAgent.Stats.TotalUsage.PromptCacheMissTokens
	chatResult.Usage.TotalTokens = ckgAgent.Stats.TotalUsage.TotalTokens

	// 设置时间统计
	// TODO
	// chatResult.Performance.ReasoningSeconds = res.ReasoningTime
	chatResult.Performance.CostSeconds = int(ckgAgent.Stats.DurationMs / 1000)

	return chatResult
}
