package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/insmtx/corekg/pkgs/einotools/agent"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

type ReactAgentService struct {
	State   agent.AgentState
	Memory  models.Memory
	Stats   models.AgentStats
	Rresult []*models.WriteResult
}

func (impl *ReactAgentService) Handler(ctx context.Context,
	agentContext *agent.AgentContext, agentRequest *models.AgentRequest,
	opts ...models.HandlerOption) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	agentContext.Printer.SetCancelFunc(cancel)
	logs.InfoContextf(ctx, "ReactAgentService Handler start: sessionID=%d requestID=%s isStream=%v inputFiles=%d historyMessages=%d maxStep=%d tools=%d",
		agentContext.SessionID, agentContext.RequestID, agentContext.IsStream, len(agentContext.ProductFiles), countRequestMessages(agentRequest), agentContext.MaxStep, len(agentContext.AvailableTools))

	roleName, err := settings.GetText(global.SettingGroupCoreKG, global.SettingKeyLlmRoleName)
	if roleName == "" || err != nil {
		logs.WarnContextf(ctx, "ReactAgentService load role name failed or empty, use default: err=%v", err)
		roleName = global.DefaultLlmRoleName
	}
	agentContext.ModelRoleName = roleName
	logs.InfoContextf(ctx, "ReactAgentService role name resolved: requestID=%s roleName=%s", agentContext.RequestID, roleName)

	options := models.HandlerOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	logs.InfoContextf(ctx, "ReactAgentService options resolved: requestID=%s debug=%v summaryMode=%v", agentContext.RequestID, options.Debug, options.SummaryMode)
	originalIsStream := agentContext.IsStream
	defer func() {
		agentContext.IsStream = originalIsStream
	}()

	logs.InfoContextf(ctx, "ReactAgentService init ReactAgent start: requestID=%s", agentContext.RequestID)
	reactAgent, err := agent.NewReactAgent(ctx, agentContext)
	if err != nil {
		logs.ErrorContextf(ctx, "ReactAgentService init ReactAgent failed: requestID=%s err=%v", agentContext.RequestID, err)
		impl.State = agent.ERROR
		return "", fmt.Errorf("failed to init ReactAgent: %w", err)
	}
	logs.InfoContextf(ctx, "ReactAgentService init ReactAgent success: requestID=%s", agentContext.RequestID)

	logs.InfoContextf(ctx, "ReactAgentService init SummaryAgent start: requestID=%s", agentContext.RequestID)
	summaryAgent, err := agent.NewSummaryAgent(ctx, agentContext)
	if err != nil {
		logs.ErrorContextf(ctx, "ReactAgentService init SummaryAgent failed: requestID=%s err=%v", agentContext.RequestID, err)
		impl.State = agent.ERROR
		return "", fmt.Errorf("failed to init SummaryAgent: %w", err)
	}
	logs.InfoContextf(ctx, "ReactAgentService init SummaryAgent success: requestID=%s", agentContext.RequestID)

	if options.ReactStream != nil {
		agentContext.IsStream = *options.ReactStream
	}
	logs.InfoContextf(ctx, "ReactAgentService run ReactAgent start: requestID=%s isStream=%v", agentContext.RequestID, agentContext.IsStream)
	_, err = reactAgent.Run(ctx, agentContext.Query, options)
	logs.InfoContextf(ctx, "ReactAgentService run ReactAgent finished: requestID=%s err=%v state=%v messages=%d",
		agentContext.RequestID, err, reactAgent.State, len(reactAgent.Memory.Messages))

	enableSummary := options.SummaryMode

	if err != nil {
		if isContextCanceled(ctx, err) {
			logs.InfoContextf(ctx, "ReactAgentService ReactAgent canceled: requestID=%s err=%v", agentContext.RequestID, err)
			impl.updateStats(ctx, []*agent.BaseAgent{&reactAgent.BaseAgent})
			return "", nil
		}

		if errors.Is(err, compose.ErrExceedMaxSteps) {
			// 当前友好处理，强制根据已有内容，执行总结进行内容输出
			logs.WarnContextf(ctx, "ReactAgentService ReactAgent exceed max steps, force enable summary: requestID=%s maxStep=%d messages=%d",
				agentContext.RequestID, agentContext.MaxStep, len(reactAgent.Memory.Messages))
			enableSummary = true
		} else {

			memoryStr, _ := json.Marshal(reactAgent.Memory)
			logs.ErrorContextf(ctx, "ReAct Agent memoryStr: %s", memoryStr)
			logs.ErrorContextf(ctx, "ReAct agent execution failed: %v", err)

			impl.State = agent.ERROR
			return "", fmt.Errorf("react agent execution failed: %w", err)
		}
	}

	if !enableSummary {
		// 不开启总结报告
		logs.InfoContextf(ctx, "ReactAgentService summary disabled, return last message: requestID=%s messages=%d",
			agentContext.RequestID, len(reactAgent.Memory.Messages))
		impl.updateStats(ctx, []*agent.BaseAgent{
			&reactAgent.BaseAgent,
		})
		content := getContent(reactAgent.Memory.GetLastMessage())
		logs.InfoContextf(ctx, "ReactAgentService Handler success without summary: requestID=%s contentLength=%d", agentContext.RequestID, len(content))
		return content, nil
	}

	historyMsgs := reactAgent.Memory.GetLlmMessages()
	if options.SummaryStream != nil {
		agentContext.IsStream = *options.SummaryStream
	}
	logs.InfoContextf(ctx, "ReactAgentService run SummaryAgent start: requestID=%s isStream=%v historyMessages=%d", agentContext.RequestID, agentContext.IsStream, len(historyMsgs))
	summary, err := summaryAgent.RunSummarizeResult(ctx, historyMsgs)
	logs.InfoContextf(ctx, "ReactAgentService run SummaryAgent finished: requestID=%s err=%v state=%v summaryLength=%d messages=%d",
		agentContext.RequestID, err, summaryAgent.State, len(summary), len(summaryAgent.Memory.Messages))

	impl.updateStats(ctx, []*agent.BaseAgent{
		&reactAgent.BaseAgent,
		&summaryAgent.BaseAgent,
	})

	if err != nil {
		if isContextCanceled(ctx, err) {
			logs.InfoContextf(ctx, "ReactAgentService SummaryAgent canceled: requestID=%s err=%v", agentContext.RequestID, err)
			message := summaryAgent.Memory.GetLastMessage()
			if message != nil && message.Payload != nil && message.Payload.Role == schema.Assistant {
				content := getContent(message)
				logs.InfoContextf(ctx, "ReactAgentService return partial summary after cancel: requestID=%s contentLength=%d",
					agentContext.RequestID, len(content))
				return content, nil
			}
			return "", nil
		}
		logs.ErrorContextf(ctx, "ReactAgentService SummaryAgent failed: requestID=%s err=%v", agentContext.RequestID, err)
		impl.State = agent.ERROR
		return "", fmt.Errorf("summary agent execution failed: %w", err)
	}

	logs.InfoContextf(ctx, "ReactAgentService Handler success with summary: requestID=%s summaryLength=%d", agentContext.RequestID, len(summary))
	return summary, nil
}

func (impl *ReactAgentService) updateStats(ctx context.Context, agents []*agent.BaseAgent) {
	if len(agents) == 0 {
		logs.WarnContextf(ctx, "ReactAgentService updateStats skipped: empty agents")
		return
	}

	logs.InfoContextf(ctx, "ReactAgentService updateStats start: agents=%d", len(agents))
	for idx, baseAgent := range agents {
		logs.InfoContextf(ctx, "ReactAgentService updateStats merge agent: index=%d state=%v messages=%d totalTokens=%d",
			idx, baseAgent.State, len(baseAgent.Memory.Messages), baseAgent.Stats.TotalUsage.TotalTokens)
		impl.State = baseAgent.State
		for _, msg := range baseAgent.Memory.Messages {
			impl.Memory.PrivateAddMessage(msg)
		}
		impl.Stats.AddTotalUsage(&baseAgent.Stats.TotalUsage)
	}
	impl.Stats.StartTimestamp = agents[0].Stats.StartTimestamp
	impl.Stats.EndTimestamp = agents[len(agents)-1].Stats.EndTimestamp
	impl.Stats.DurationMs = impl.Stats.EndTimestamp - impl.Stats.StartTimestamp
	logs.InfoContextf(ctx, "ReactAgentService updateStats finished: state=%v messages=%d totalTokens=%d durationMs=%d",
		impl.State, len(impl.Memory.Messages), impl.Stats.TotalUsage.TotalTokens, impl.Stats.DurationMs)
}

func isContextCanceled(ctx context.Context, err error) bool {
	return errors.Is(err, context.Canceled) || ctx.Err() != nil
}

func countRequestMessages(agentRequest *models.AgentRequest) int {
	if agentRequest == nil {
		return 0
	}
	return len(agentRequest.Messages)
}

func getContent(msg *models.Message) string {
	if msg == nil || msg.Payload == nil {
		return ""
	}
	return msg.Payload.Content
}
