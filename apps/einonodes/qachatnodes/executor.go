package qachatnodes

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/insmtx/corekg/apps/einonodes/nodebase"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatclient"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/qachat"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
)

type executor struct {
	*baseHandler
}

func newExecutor() *executor {
	return &executor{
		baseHandler: &baseHandler{},
	}
}

func (e *executor) NewChunkChatNode(ctx context.Context, input string, opts ...any) (output string, err error) {
	err = compose.ProcessState[*State](ctx, func(_ context.Context, state *State) error {
		chatReq := &chattype.ChatRequestBody{
			Stream:     false,
			Model:      chatagent.GetAgentI18nName(ctx, "", global.ChatAgentESChat),
			LLMModelID: state.SessionEntity.ModelID,
			ChatOptions: chattype.ChatOptions{
				Input: []chattype.Input{
					{Name: "input1", Value: state.QuestionEntity.Source.Question}, // 用户问题
					{Name: "input2", Value: input},                                // 知识库检索内容
					{Name: "input3", Value: state.HistoryContext},                 // 对话历史
				},
			},
		}
		wrapper, err := chatclient.NewInternalChat(state.Ctx, state.QuestionEntity.Source.ReqID, "", 2, chatReq)
		if err != nil {
			logs.ErrorContextf(ctx, "[NewChunkChatLambdaNode] failed to create internal chat: %v", err)
			return err
		}
		res, err := wrapper.AgentChatInternal(nil)
		if err != nil {
			logs.ErrorContextf(ctx, "[NewChunkChatLambdaNode] agent chat error: %v", err)
			return err
		}
		output = res.Content
		return nil
	})
	return output, nil
}

func (e *executor) RunMySQLStatementNode(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {
	err = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		if err := e.checkState(state); err != nil {
			logs.ErrorContextf(ctx, "[RunMySQLStatementNode] check state fail, err: %v", err)
			return err
		}
		forestIDRecord := input.Get(RecordKeyForestID)
		forestID := uint(utils.VToUint64(forestIDRecord.Value))

		tableDDLRecord := input.Get(RecordKeyMySQLTableDDLMap)
		var tableDDLMap map[string]string
		if err := json.Unmarshal([]byte(tableDDLRecord.Value), &tableDDLMap); err != nil {
			logs.ErrorContextf(state.Ctx, "[RunMySQLStatementNode] unmarshal table ddl map error: %v", err)
			return err
		}

		wrapper := qachat.NewChatWrapper(state.Ctx, state.QuestionEntity, state.SessionEntity, state.ModelEntity)

		excelDB, err := wrapper.GetExcelDB(forestID)
		if err != nil {
			logs.ErrorContextf(state.Ctx, "[RunMySQLStatementNode] Failed to get mysql db: %v", err)
			return err
		}
		sqlDB, err := excelDB.DB()
		if err != nil {
			logs.ErrorContextf(state.Ctx, "[RunMySQLStatementNode] Failed to get mysql db: %v", err)
			return err
		}
		defer sqlDB.Close()

		// 生成 SQL 并执行
		var sqlResList []map[string]any
		var questionSQL string
		retryErr := lifecycle.Retry(time.Second*5, 3, func() (needRetry bool, err error) {
			logs.InfoContextf(state.Ctx, "[RunMySQLStatementNode] running question: %s", state.QuestionEntity.Source.Question)
			// 执行问题对应的sql
			sql, err := wrapper.GetQuestionMysqlSQL(state.Ctx, excelDB, state.QuestionEntity.Source.Question, tableDDLMap)
			if err != nil {
				logs.ErrorContextf(state.Ctx, "[RunMySQLStatementNode] get question mysql sql error: %v", err)
				return true, err
			}
			questionSQL = sql
			if err := excelDB.WithContext(state.Ctx).Raw(questionSQL).Scan(&sqlResList).Error; err != nil {
				return true, fmt.Errorf("query question sql failed, err: %v", err)
			}
			return false, nil
		})
		if retryErr != nil {
			logs.ErrorContextf(state.Ctx, "[RunMySQLStatementNode] retry error: %v", retryErr)
			return retryErr
		}
		state.QuestionDbDatasetEntity.QueryStatement = questionSQL
		state.QuestionDbDatasetEntity.QueryResult = logs.JSON(sqlResList)

		output.Add(&nodebase.Record{
			Key:   RecordKeyQueryStatement,
			Value: questionSQL,
		})
		output.Add(&nodebase.Record{
			Key:   RecordKeyQueryStatementRes,
			Value: logs.JSON(sqlResList),
		})
		output.Add(&nodebase.Record{
			Key:   RecordKeyMySQLTableDDLMap,
			Value: logs.JSON(tableDDLMap),
		})
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[RunMySQLStatementNode] process state error: %v", err)
		return nil, err
	}
	return output, nil
}

func (e *executor) TransferSQLNode(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {
	err = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		if err := e.checkState(state); err != nil {
			logs.ErrorContextf(ctx, "[TransferSQLNode] check state fail, err: %v", err)
			return err
		}
		question := state.QuestionEntity.Source.Question
		sql := input.Get(RecordKeyQueryStatement).Value
		sqlResValue := input.Get(RecordKeyQueryStatementRes).Value
		var sqlResList []map[string]any
		if err := json.Unmarshal([]byte(sqlResValue), &sqlResList); err != nil {
			logs.ErrorContextf(state.Ctx, "[TransferSQLNode] unmarshal sql res error: %v", err)
			return err
		}
		ddlMapValue := input.Get(RecordKeyMySQLTableDDLMap).Value
		var ddlMap map[string]string
		if err := json.Unmarshal([]byte(ddlMapValue), &ddlMap); err != nil {
			logs.ErrorContextf(state.Ctx, "[TransferSQLNode] unmarshal ddl map error: %v", err)
			return err
		}
		var ddlList []string
		for _, v := range ddlMap {
			ddlList = append(ddlList, v)
		}
		wrapper := qachat.NewChatWrapper(state.Ctx, state.QuestionEntity, state.SessionEntity, state.ModelEntity)
		sqlAnswerRes, err := wrapper.TransferMysqlResultToAnswer(state.Ctx, question, sql, ddlList, sqlResList)
		if err != nil {
			logs.ErrorContextf(state.Ctx, "[TransferSQLNode] transfer mysql result to answer error: %v", err)
		}
		if sqlAnswerRes != nil {
			output.Add(&nodebase.Record{
				Key:   RecordKeyTextContent,
				Value: sqlAnswerRes.Content,
			})
			state.QuestionEntity.Source.Answer = sqlAnswerRes.Content
			state.QuestionEntity.Source.Reasoning = sqlAnswerRes.Reasoning
			state.QuestionEntity.Source.ReasoningSeconds = sqlAnswerRes.ReasoningTime
			state.QuestionEntity.Source.CostSeconds = sqlAnswerRes.CostSeconds
			state.QuestionEntity.Source.OutToken = sqlAnswerRes.Usage.CompletionTokens
			state.QuestionEntity.Source.CacheHitToken = sqlAnswerRes.Usage.PromptCacheHitTokens
			state.QuestionEntity.Source.CacheMissToken = sqlAnswerRes.Usage.PromptCacheMissTokens
			state.QuestionEntity.Source.TotalTokens = sqlAnswerRes.Usage.TotalTokens
			state.QuestionEntity.Source.Status = chattype.QuestionStatusAnswered
		}
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[TransferSQLNode] process state error: %v", err)
		return nil, err
	}
	return output, nil
}

func (e *executor) GenerateMysqlChart(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {

	err = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		question := state.QuestionEntity.Source.Question
		sqlResValue := input.Get(RecordKeyQueryStatementRes).Value
		sql := input.Get(RecordKeyQueryStatement).Value
		var sqlResList []map[string]any
		if err := json.Unmarshal([]byte(sqlResValue), &sqlResList); err != nil {
			logs.ErrorContextf(state.Ctx, "[GenerateMysqlChart] unmarshal sql res error: %v", err)
			return err
		}
		wrapper := qachat.NewChatWrapper(state.Ctx, state.QuestionEntity, state.SessionEntity, state.ModelEntity)
		chartMap := wrapper.BatchGenerateEcharts(state.Ctx, question, sql, sqlResList)
		if err != nil {
			logs.ErrorContextf(ctx, "[GenerateMysqlChart] IsMeaningfulECharts error: %v", err)
			return err
		}
		for chartType, chartContent := range chartMap {
			output.Add(&nodebase.Record{
				Key:   string(chartType),
				Value: chartContent,
			})
		}

		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[GenerateMysqlChart] ProcessState error: %v", err)
		return nil, err
	}
	return output, nil
}

func (e *executor) BlackNode(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {
	output = input
	return output, nil
}

func (e *executor) BlackNodeCompose(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {
	output = input
	return output, nil
}
func (e *executor) ExternalChatNode(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {
	err = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		if err := e.checkState(state); err != nil {
			logs.ErrorContextf(ctx, "[TransferSQLNode] check state fail, err: %v", err)
			return err
		}
		// 获取历史记录
		history, err := qachat.GetForestChatHistory(ctx, state.SessionEntity)
		if err != nil {
			logs.ErrorContextf(state.Ctx, "[ExternalChatNode] get forest chat history error: %v", err)
			return err
		}

		chatReq := &chattype.ChatRequestBody{
			Stream:     true,
			Model:      chatagent.GetAgentI18nName(state.Ctx, runtime.GetLanguage(state.Ctx), global.ChatAgentQuestionAnswer),
			LLMModelID: state.SessionEntity.ModelID,
			ChatOptions: chattype.ChatOptions{
				Input: []chattype.Input{
					{Name: "input1", Value: state.QuestionEntity.Source.Question}, // 用户问题
					{Name: "input2", Value: e.getExternalData(ctx, input)},        // 知识库检索内容
					{Name: "input3", Value: history},                              // 对话历史
				},
			},
		}
		wrapper, err := chatclient.NewInternalChat(state.Ctx, state.QuestionEntity.Source.ReqID, "", 2, chatReq)
		if err != nil {
			logs.ErrorContextf(ctx, "[ExternalChatNode] failed to create internal chat: %v", err)
			return err
		}
		_, err = wrapper.AgentChatInternal(nil)
		if err != nil {
			logs.ErrorContextf(ctx, "[ExternalChatNode] agent chat error: %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[ExternalChatNode] process state error: %v", err)
		return nil, err
	}
	return output, nil
}

// getExternalData 获取外部数据
func (e *executor) getExternalData(ctx context.Context, input nodebase.RecordList) string {
	externalData := ""
	gmailData := input.Get(RecordKeyGmailData).Value
	slackData := input.Get(RecordKeySlackData).Value
	googleData := input.Get(RecordKeyGoogleDriveData).Value
	if gmailData != "" {
		externalData += "Gmail search Data:\n" + gmailData + "\n"
	}
	if slackData != "" {
		externalData += "slack search Data:\n" + slackData + "\n"
	}
	if googleData != "" {
		externalData += "google drive search Data:\n" + googleData + "\n"
	}
	logs.InfoContextf(ctx, "[ExternalChatNode] get external data: %s", externalData)
	return externalData
}
