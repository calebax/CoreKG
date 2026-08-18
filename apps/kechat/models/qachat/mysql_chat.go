package qachat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatclient"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins/mysqlplugin"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/concqueue"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"gorm.io/gorm"
)

type TableDDL struct {
	Table       string `gorm:"column:Table"`
	CreateTable string `gorm:"column:Create Table"`
	View        string `gorm:"column:View"`
	CreateView  string `gorm:"column:Create View"`
}

type ChartData struct {
	ChartID      uint   `json:"chart_id"`
	SessionID    uint   `json:"session_id"`
	QuestionID   string `json:"question_id"`
	ChartContent string `json:"chart_content"`
}

func (w *ChatWapper) MysqlChat() error {
	questionDatasetEntity := &chattype.ChatQuestionDbDataset{
		SessionID:    w.question.Source.SessionID,
		QuestionID:   w.question.ID,
		DatabaseType: string(dbplugins.DatabaseTypeMySQL),
		RequestID:    runtime.RequestID(w.ctx),
	}
	defer func() {
		if err := chatquestion.NewChatQuestionDbDatasetDao().Insert(w.ctx, questionDatasetEntity); err != nil {
			logs.ErrorContextf(w.ctx, "[MysqlChat] Failed to insert question dataset: %v", err)
		}
	}()
	dbPluginConfig, databaseType, err := w.getMysqlPluginConfig(w.ctx, w.session)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[MysqlChat] Failed to get mysql plugin config: %v", err)
		return err
	}
	tables, err := w.getMysqlTableList(w.ctx, w.session, dbPluginConfig, databaseType)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[MysqlChat] Failed to get mysql table list: %v", err)
		return err
	}
	mysqlPlugin := &mysqlplugin.MySQLPlugin{}
	db, err := mysqlPlugin.DB(w.ctx, dbPluginConfig)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[MysqlChat] Failed to get mysql db: %v", err)
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		logs.ErrorContextf(w.ctx, "[MysqlChat] Failed to get mysql sql db: %v", err)
		return err
	}
	defer sqlDB.Close()

	tableDDLMap, err := w.getMysqlTableDDL(w.ctx, db, tables)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[MysqlChat] Failed to get mysql table ddl: %v", err)
		return err
	}

	choiceTableList, err := w.ChoiceTableList(w.question.Source.Question, tableDDLMap)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[MysqlChat] Failed to choice table list: %v", err)
		return err
	}
	choiceDDLMap := make(map[string]string, len(choiceTableList))
	for _, v := range choiceTableList {
		choiceDDLMap[v] = tableDDLMap[v]
	}
	questionDatasetEntity.TableList = logs.JSON(choiceTableList)

	var sqlResList []map[string]interface{}
	var questionSQL, answer string
	retryErr := lifecycle.Retry(time.Second*5, 3, func() (needRetry bool, err error) {
		logs.InfoContextf(w.ctx, "[MySQLChat] running question: %s", w.question.Source.Question)
		// 执行问题对应的sql
		questionSQL, err = w.GetQuestionMysqlSQL(w.ctx, db, w.question.Source.Question, choiceDDLMap)
		if err != nil {
			return true, err
		}
		if w.isWriteOperationSQL(questionSQL) {
			logs.WarnContextf(w.ctx, "[MySQLChat] sql is write sql, not support, sql: %s, question: %s", questionSQL, w.question.Source.Question)
			answer = i18n.T(runtime.GetLanguage(w.ctx), "kechat_mysql_chat_forbidden_write")
			return false, nil
		}
		if err := db.WithContext(w.ctx).Raw(questionSQL).Scan(&sqlResList).Error; err != nil {
			return true, fmt.Errorf("query question sql failed, err: %v", err)
		}
		return false, nil
	})

	if retryErr != nil {
		return retryErr
	}
	logs.InfoContextf(w.ctx, "[MySQLChat] question sql: %s", questionSQL)
	questionDatasetEntity.QueryStatement = questionSQL
	questionDatasetEntity.QueryResult = logs.JSON(sqlResList)

	if len(sqlResList) == 0 {
		logs.WarnContextf(w.ctx, "no rows in sql, sql: %s", questionSQL)
		if answer == "" {
			answer = i18n.T(runtime.GetLanguage(w.ctx), "kechat_mysql_chat_no_rows")
		}
		llmchat.WriteContent(w.ctx, w.question.Source.ReqID, answer)
		w.question.Source.Answer = answer
		w.question.Source.Status = chattype.QuestionStatusAnswered
		return nil
	}

	// 准备 DDL 列表
	ddlList := make([]string, 0, len(choiceDDLMap))
	for _, tableDDL := range choiceDDLMap {
		ddlList = append(ddlList, tableDDL)
	}

	var wg sync.WaitGroup
	var sqlAnswerRes *llmchat.QaRes
	var transferSqlAnswerErr error
	var chartMap map[chattype.ChartType]string

	wg.Add(1)
	go func() {
		defer wg.Done()
		// copyCtx := w.ctx.Copy()
		sqlAnswerRes, transferSqlAnswerErr = w.TransferMysqlResultToAnswer(w.ctx, w.question.Source.Question, questionSQL, ddlList, sqlResList)
		if transferSqlAnswerErr != nil {
			logs.ErrorContextf(w.ctx, "[MysqlChat] Failed to search answer: %v", transferSqlAnswerErr)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// copyCtx := w.ctx.Copy()
		chartMap = w.BatchGenerateEcharts(w.ctx, w.question.Source.Question, questionSQL, sqlResList)
	}()

	// 等待两个 goroutine 都完成
	wg.Wait()

	// 返回错误（如果有的话）
	if transferSqlAnswerErr != nil {
		return fmt.Errorf("failed to get answer, err: %v", transferSqlAnswerErr)
	}

	sqlTitleFlag := i18n.T(runtime.GetLanguage(w.ctx), "kechat_mysql_chat_sql_title")

	sqlTitle := fmt.Sprintf("\n\n%s:", sqlTitleFlag)
	llmchat.WriteStreamsResult(w.ctx, w.question.Source.ReqID, llmchat.WriteResult{
		Content: sqlTitle,
		Flag:    llmchat.StreamsFlagSQL,
	})
	sqlContent := "```sql\n" + questionSQL + "\n```"
	llmchat.WriteStreamsResult(w.ctx, w.question.Source.ReqID, llmchat.WriteResult{
		Content: sqlContent,
		Flag:    llmchat.StreamsFlagSQL,
	})

	// 处理 echarts 结果
	for _, chartContent := range chartMap {
		llmchat.WriteStreamsResult(w.ctx, w.question.Source.ReqID, llmchat.WriteResult{
			Content: chartContent,
			Flag:    llmchat.StreamsFlagECharts,
		})
	}

	// 处理转换结果
	if sqlAnswerRes != nil {
		answer = sqlAnswerRes.Content
		answer = answer + sqlTitle + "\n" + sqlContent
		for _, chartContent := range chartMap {
			answer = answer + "\n" + chartContent + "\n"
		}
		w.question.Source.Answer = answer
		w.question.Source.Reasoning = sqlAnswerRes.Reasoning
		w.question.Source.ReasoningSeconds = sqlAnswerRes.ReasoningTime
		w.question.Source.CostSeconds = sqlAnswerRes.CostSeconds
		w.question.Source.OutToken = sqlAnswerRes.Usage.CompletionTokens
		w.question.Source.CacheHitToken = sqlAnswerRes.Usage.PromptCacheHitTokens
		w.question.Source.CacheMissToken = sqlAnswerRes.Usage.PromptCacheMissTokens
		w.question.Source.TotalTokens = sqlAnswerRes.Usage.TotalTokens
		w.question.Source.Status = chattype.QuestionStatusAnswered
	}

	return nil
}

func (w *ChatWapper) getMysqlPluginConfig(ctx *gin.Context, sessionEntity *chattype.ChatSession) (*dbplugins.PluginConfig, dbplugins.DatabaseType, error) {
	var forestID uint
	if len(sessionEntity.ForestIDList.Slice()) > 0 {
		forestID = sessionEntity.ForestIDList.Slice()[0]
	}
	instanceEntity, err := forest.NewForestDBInstanceDao().GetByCond(ctx, &forest.ForestDBInstanceCond{
		ForestID: forestID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[MysqlChat.getMysqlPluginConfig] Failed to get db instance: %v", err)
		return nil, "", err
	}
	dbPluginConfig := &dbplugins.PluginConfig{
		Credentials: &dbplugins.Credentials{
			ConnectionID: fmt.Sprintf("%d", forestID),
			Hostname:     instanceEntity.Host,
			Username:     instanceEntity.Username,
			Password:     settings.DecryptSecret(instanceEntity.Password),
			Port:         instanceEntity.Port,
			Database:     instanceEntity.Database,
		},
	}
	return dbPluginConfig, instanceEntity.InstanceType, nil
}

func (w *ChatWapper) getMysqlTableList(ctx *gin.Context, sessionEntity *chattype.ChatSession, dbPluginConfig *dbplugins.PluginConfig, databaseType dbplugins.DatabaseType) ([]string, error) {
	// sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()), sseclient.WithBlockMaxRetry(30))
	// if len(sessionEntity.ForestIDList.Slice()) == 0 {
	// 	logs.ErrorContextf(ctx, "[MysqlChat.getMysqlTableList] forest id list is empty")
	// 	if stopped, err := sseClient.WriteMessage(ctx, ctx.Writer, streamsKey, writeResult.String()); err != nil {
	// 		if strings.Contains(err.Error(), "broken pipe") {
	// 			continue
	// 		}
	// 		logs.ErrorContextf(ctx, "[transferSQLResultToAnswer] Failed to write Answering response to KEQA: %v", err)
	// 		return err
	// 	} else if stopped {
	// 		logs.InfoContextf(ctx, "[transferSQLResultToAnswer] stream has been stopped")
	// 		return nil
	// 	}
	// }
	sessionTables := sessionEntity.DBTableList.Slice()

	isAvailable, err := dbutil.GetDBPluginEngine().ChoosePlugin(dbplugins.DatabaseType(databaseType)).IsAvailable(ctx, dbPluginConfig)
	if err != nil {
		logs.ErrorContextf(ctx, "[MysqlChat.getMysqlTableList] Failed to check db instance: %v", err)
		return nil, err
	}
	if !isAvailable {
		logs.ErrorContextf(ctx, "[MysqlChat.getMysqlTableList] db instance is not available")
		return nil, fmt.Errorf("db instance is not available")
	}
	storageUnitsQueryOption := &dbplugins.QueryOption{
		Filters: []dbplugins.Filter{
			{
				Key:    dbplugins.FilterKeyTables,
				Values: sessionTables,
			},
		},
	}
	storageUnitsRes, err := dbutil.GetDBPluginEngine().ChoosePlugin(dbplugins.DatabaseType(databaseType)).GetStorageUnits(ctx, dbPluginConfig, dbPluginConfig.Credentials.Database, storageUnitsQueryOption)
	if err != nil {
		logs.ErrorContextf(ctx, "[MysqlChat.getMysqlTableList] Failed to get storage units: %v", err)
		return nil, err
	}
	var tables []string
	for _, v := range storageUnitsRes.List {
		tables = append(tables, v.Name)
	}
	return tables, nil
}

func (w *ChatWapper) ChoiceTableList(question string, tableDDLMap map[string]string) ([]string, error) {
	ddlList := make([]string, 0, len(tableDDLMap))
	for _, ddl := range tableDDLMap {
		ddlList = append(ddlList, ddl)
	}

	// 每组20个DDL
	const batchSize = 20
	batches := make([][]string, 0)

	// 将DDL分组
	for i := 0; i < len(ddlList); i += batchSize {
		end := i + batchSize
		if end > len(ddlList) {
			end = len(ddlList)
		}
		batches = append(batches, ddlList[i:end])
	}

	// 如果没有批次，直接返回
	if len(batches) == 0 {
		return []string{}, nil
	}

	// 创建队列，使用5个worker，队列大小为10
	q := concqueue.New(10, 20)
	batchCount := len(batches)
	res := make([][]string, batchCount)
	errs := make([]error, batchCount)

	// 提交任务
	for i := 0; i < batchCount; i++ {
		batchIndex := i // 传递索引值给任务，避免并发时数据竞争
		batch := batches[i]
		q.Submit(func(ctx context.Context) error {
			tables, err := w.batchChoiceTableList(question, batch)
			if err != nil {
				errs[batchIndex] = err
			} else {
				res[batchIndex] = tables
			}
			return err
		})
	}

	// 等待任务完成并关闭队列
	errCnt := q.StopAndWait()

	// 如果有错误，记录日志并返回第一个错误
	if errCnt > 0 {
		return nil, fmt.Errorf("failed to process all batches: %d, errors: %s", errCnt, logs.JSON(errs))
	}

	// 合并所有批次的结果
	var allTables []string
	for _, batchTables := range res {
		if batchTables != nil {
			allTables = append(allTables, batchTables...)
		}
	}

	// 去重
	uniqueTables := utils.SliceDuplicate(allTables)

	logs.InfoContextf(w.ctx, "[MySQLChat] Processed %d batches, found %d unique tables", batchCount, len(uniqueTables))

	// 如果超出20，再缩小范围
	if len(uniqueTables) > 20 {
		newDDLMap := make(map[string]string, len(uniqueTables))
		for _, tbl := range uniqueTables {
			if ddl, ok := tableDDLMap[tbl]; ok {
				newDDLMap[tbl] = ddl
			}
		}
		return w.ChoiceTableList(question, newDDLMap)
	}

	return uniqueTables, nil
}

func (w *ChatWapper) batchChoiceTableList(question string, ddlList []string) ([]string, error) {
	// ddlList := make([]string, 0, len(tableDDLMap))
	// for _, ddl := range tableDDLMap {
	// 	ddlList = append(ddlList, ddl)
	// }
	agentReq := &chattype.ChatRequestBody{
		Stream: false,
		Model:  chatagent.GetAgentI18nName(w.ctx, runtime.GetLanguage(w.ctx), global.ChantAgentMysqlChoiceTableByQuestionKey),
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{
					Name:  "input1",
					Value: question,
				},
				{
					Name:  "input2",
					Value: strings.Join(ddlList, "\n"),
				},
			},
		},
	}

	chatClient, err := chatclient.NewInternalChat(w.ctx, runtime.RequestID(w.ctx), "", 1, agentReq)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[MySQLChat] failed to create internal chat: %v", err)
		return nil, err
	}
	agentRes, err := chatClient.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[MySQLChat] Failed to chat: %v", err)
		return nil, err
	}
	agentResStr := chatclient.ExtractCode("json", agentRes.Content)
	if agentResStr == "" {
		logs.WarnContextf(w.ctx, "[MySQLChat] Failed to extract code from agent res: %s", agentRes.Content)
		return nil, nil
	}
	type ChoiceTableRes struct {
		Tables []string `json:"tables"`
	}
	var choiceTableRes ChoiceTableRes
	if err := json.Unmarshal([]byte(agentResStr), &choiceTableRes); err != nil {
		logs.ErrorContextf(w.ctx, "[MySQLChat] Failed to unmarshal agent res: %v", err)
		return nil, err
	}
	return choiceTableRes.Tables, nil

}

func (w *ChatWapper) getMysqlTableDDL(ctx *gin.Context, db *gorm.DB, tables []string) (map[string]string, error) {
	tableDDLMap := make(map[string]string)
	for _, tableName := range tables {
		var ddl TableDDL
		if err := db.WithContext(ctx).Raw("SHOW CREATE TABLE " + tableName).Scan(&ddl).Error; err != nil {
			return nil, fmt.Errorf("get table ddl failed, err: %v", err)
		}
		logs.InfoContextf(ctx, "table‘s ddl，table: %s， ddl: %s", tableName, logs.JSON(ddl))
		tableDDLMap[tableName] = ddl.CreateTable
	}
	return tableDDLMap, nil
}

func (w *ChatWapper) GetQuestionMysqlSQL(ctx *gin.Context, db *gorm.DB, question string, tableDDLMap map[string]string) (string, error) {
	// 每个 table 获取 10行数据
	var dataAndDDLList []string
	for tableName, ddl := range tableDDLMap {
		var dataList []map[string]interface{}
		if err := db.WithContext(ctx).Raw("select * from " + tableName + " limit 10").Scan(&dataList).Error; err != nil {
			return "", fmt.Errorf("getTenLinesData err: %v", err)
		}
		dataBytes, _ := json.Marshal(dataList)

		dataAndDDLList = append(dataAndDDLList, fmt.Sprintf("表结构：\n%s\n数据：\n%s\n", ddl, string(dataBytes)))
	}

	req := &chattype.ChatRequestBody{
		Stream: false,
		Model:  chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentExcelQuestionToSql),
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{
					Name:  "input1",
					Value: strings.Join(dataAndDDLList, "\n"),
				},
				{
					Name:  "input2",
					Value: question,
				},
			},
		},
	}

	chatClient, err := chatclient.NewInternalChat(ctx, runtime.RequestID(ctx), "", 1, req)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to create internal chat: %v", err)
		return "", err
	}
	res, err := chatClient.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(ctx, "agent chat error: %v", err)
		return "", err
	}
	sql := chatclient.ExtractCode("sql", res.Content)
	return sql, nil
}

func (w *ChatWapper) TransferMysqlResultToAnswer(ctx *gin.Context, question, sql string, ddlList []string, sqlResList []map[string]interface{}) (*llmchat.QaRes, error) {
	resBytes, _ := json.Marshal(sqlResList)
	req := &chattype.ChatRequestBody{
		Stream: true,
		Model:  chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentExcelSqlResultAnalysis),
		// LLMModelID: modelID, // 若支持选中模型需要传模型id
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{
					Name:  "input1",
					Value: string(resBytes),
				},
				{
					Name:  "input2",
					Value: question,
				},
				{
					Name:  "input3",
					Value: strings.Join(ddlList, "\n"),
				},
				{
					Name:  "input4",
					Value: sql,
				},
			},
		},
	}
	chatClient, err := chatclient.NewInternalChat(ctx, runtime.RequestID(ctx), "", 2, req)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to create internal chat: %v", err)
		return nil, err
	}
	res, err := chatClient.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(ctx, "agent chat error: %v", err)
		return nil, err
	}

	return res, err
}

func (w *ChatWapper) GetEcharts(ctx *gin.Context, question string, sqlResList []map[string]interface{}) (string, error) {
	resBytes, _ := json.Marshal(sqlResList)
	req := &chattype.ChatRequestBody{
		Stream: false,
		Model:  chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChantAgentMysqlGenerateEcharts),
		// LLMModelID: modelID, // 若支持选中模型需要传模型id
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{
					Name:  "input1",
					Value: string(resBytes),
				},
				{
					Name:  "input2",
					Value: question,
				},
			},
		},
	}
	chatClient, err := chatclient.NewInternalChat(ctx, runtime.RequestID(ctx), "", 2, req)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to create internal chat: %v", err)
		return "", err
	}
	agentRes, err := chatClient.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetEchartsByChartType] agent chat error: %v", err)
		return "", err
	}
	echartStr := chatclient.ExtractCode("json", agentRes.Content)
	if agentRes.Content == "{}" {
		return "", nil
	}

	markdownEchartStr := fmt.Sprintf("```echarts\n%s\n```", echartStr)
	return markdownEchartStr, nil
}

func (w *ChatWapper) BatchGenerateEcharts(ctx *gin.Context, question, sql string, sqlResList []map[string]interface{}) map[chattype.ChartType]string {
	echartMap := make(map[chattype.ChartType]string)
	for chartType := range chattype.ValidChartTypeMap {
		echartStr, err := w.getEchartsByChartType(ctx, question, sql, sqlResList, chartType)
		if err != nil {
			logs.ErrorContextf(ctx, "get echart by chart type error: %v", err)
			continue
		}
		if echartStr == "" {
			logs.InfoContextf(ctx, "get echart by chart type is empty, chartType: %s", chartType)
			continue
		}
		logs.InfoContextf(ctx, "get echart by chart type success, chartType: %s, echartStr: %s", chartType, echartStr)
		echartMap[chartType] = echartStr
	}
	return echartMap

}

func (w *ChatWapper) getEchartsByChartType(ctx *gin.Context, question, sql string, sqlResList []map[string]interface{}, chartType chattype.ChartType) (string, error) {
	if _, ok := chattype.ValidChartTypeMap[chartType]; !ok {
		logs.ErrorContextf(ctx, "[GetEchartsByChartType] chartType is not valid, chartType: %s", chartType)
		return "", fmt.Errorf("chartType is not valid, chartType: %s", chartType)
	}
	intentReq := &chattype.ChatRequestBody{
		Stream:     false,
		Model:      chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentMysqlIntentIsMeaningfulECharts),
		LLMModelID: w.model.ID,
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{
					Name:  "input1",
					Value: question,
				},
				{
					Name:  "input2",
					Value: sql,
				},
				{
					Name:  "input3",
					Value: string(chartType),
				},
			},
		},
	}
	intentClient, err := chatclient.NewInternalChat(ctx, runtime.RequestID(ctx), "", 1, intentReq)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetEchartsByChartType] failed to create internal chat: %v", err)
		return "", err
	}
	intentRes, err := intentClient.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetEchartsByChartType] agent chat error: %v", err)
		return "", err
	}
	var isMeaningful bool
	if intentRes.Content == "Y" {
		isMeaningful = true
	}
	if !isMeaningful {
		logs.InfoContextf(ctx, "[GetEchartsByChartType] isMeaningful is false, chartType: %s", chartType)
		return "", nil
	}
	resBytes, _ := json.Marshal(sqlResList)
	req := &chattype.ChatRequestBody{
		Stream:     false,
		Model:      chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentMysqlGenerateEchartsByChartType),
		LLMModelID: w.model.ID,
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{
					Name:  "input1",
					Value: string(resBytes),
				},
				{
					Name:  "input2",
					Value: question,
				},
				{
					Name:  "input3",
					Value: string(chartType),
				},
			},
		},
	}
	chatClient, err := chatclient.NewInternalChat(ctx, runtime.RequestID(ctx), "", 2, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetEchartsByChartType] failed to create internal chat: %v", err)
		return "", err
	}
	agentRes, err := chatClient.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetEchartsByChartType] agent chat error: %v", err)
		return "", err
	}
	echartStr := chatclient.ExtractCode("json", agentRes.Content)
	if agentRes.Content == "{}" {
		return "", nil
	}
	chartEntity := &chattype.ChatChart{
		RequestID:    runtime.RequestID(ctx),
		SessionID:    w.session.ID,
		QuestionID:   w.question.ID,
		CompanyID:    runtime.CompanyID(ctx),
		Uin:          runtime.Uin(ctx),
		ChartType:    chartType,
		ChartContent: echartStr,
	}
	if w.session.SubjectID > 0 {
		chartEntity.SubjectID = w.session.SubjectID
		chartEntity.SubjectType = chattype.SessionSubjectTypeProject
	}
	if err := chat.NewChatChartDao().Insert(ctx, chartEntity); err != nil {
		logs.ErrorContextf(ctx, "[GetEchartsByChartType] insert chart error: %v", err)
		return "", err
	}
	chartData := ChartData{
		ChartID:      chartEntity.ID,
		SessionID:    w.session.ID,
		QuestionID:   w.question.ID,
		ChartContent: echartStr,
	}

	markdownEchartStr := fmt.Sprintf("```echarts\n%s\n```", logs.JSON(chartData))
	return markdownEchartStr, nil
}

// isWriteOperationSQL 判断 SQL 语句是否是写操作
func (w *ChatWapper) isWriteOperationSQL(sql string) bool {
	// 解析 SQL 语句
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[MysqlChat.isWriteOperationSQL] Failed to parse sql, sql:%s, err: %v", sql, err)
		// 如果解析失败，可以根据关键字简单判断
		return isWriteByKeyword(sql)
	}

	// 根据语句类型判断是否是写操作
	switch stmt.(type) {
	case *sqlparser.Insert:
		return true
	case *sqlparser.Update:
		return true
	case *sqlparser.Delete:
		return true
	case *sqlparser.DDL:
		return true
	case *sqlparser.Set:
		return true
	default:
		// 对于未知类型或不支持的类型，使用关键字判断
		return isWriteByKeyword(sql)
	}
}

func (w *ChatWapper) IsMeaningfulECharts(ctx *gin.Context, question, sql string, chartType chattype.ChartType) (bool, error) {
	if _, ok := chattype.ValidChartTypeMap[chartType]; !ok {
		logs.ErrorContextf(ctx, "[IsMeaningfulECharts] chartType is not valid, chartType: %s", chartType)
		return false, fmt.Errorf("chartType is not valid, chartType: %s", chartType)
	}
	req := &chattype.ChatRequestBody{
		Stream:     false,
		Model:      chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentMysqlIntentIsMeaningfulECharts),
		LLMModelID: w.model.ID,
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{
					Name:  "input1",
					Value: question,
				},
				{
					Name:  "input2",
					Value: sql,
				},
				{
					Name:  "input3",
					Value: string(chartType),
				},
			},
		},
	}

	chatClient, err := chatclient.NewInternalChat(ctx, runtime.RequestID(ctx), "", 1, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[IsMeaningfulECharts] failed to create internal chat: %v", err)
		return false, err
	}
	agentRes, err := chatClient.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(ctx, "[IsMeaningfulECharts] agent chat error: %v", err)
		return false, err
	}
	switch agentRes.Content {
	case "Y":
		return true, nil
	case "N":
		return false, nil
	default:
		return false, nil
	}
}

// isWriteByKeyword 通过关键字判断是否是写操作（备用方法）
func isWriteByKeyword(sql string) bool {
	sql = strings.TrimSpace(strings.ToUpper(sql))

	writeKeywords := []string{
		// DML 写操作
		"INSERT", "UPDATE", "DELETE", "REPLACE",
		// DDL 操作
		"CREATE", "DROP", "ALTER", "TRUNCATE",
		// 权限操作
		"GRANT", "REVOKE",
		// 配置操作
		"SET", "RENAME",
		// 数据导入
		"LOAD",
		// 事务控制
		"BEGIN", "START", "COMMIT", "ROLLBACK",
		// 锁操作
		"LOCK", "UNLOCK",
		// 维护操作
		"FLUSH", "RESET", "OPTIMIZE", "ANALYZE", "REPAIR", "CHECK",
		// 存储过程
		"CALL",
	}

	for _, keyword := range writeKeywords {
		if strings.HasPrefix(sql, keyword) {
			return true
		}
	}

	// 特殊处理 START TRANSACTION
	if strings.HasPrefix(sql, "START") && strings.Contains(sql, "TRANSACTION") {
		return true
	}

	return false
}
