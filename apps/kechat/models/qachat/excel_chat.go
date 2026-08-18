package qachat

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins/mysqlplugin"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
)

// Chat excel 问答
func (w *ChatWapper) ExcelChat() error {
	questionDatasetEntity := &chattype.ChatQuestionDbDataset{
		SessionID:    w.question.Source.SessionID,
		QuestionID:   w.question.ID,
		DatabaseType: string(dbplugins.DatabaseTypeMySQL),
		RequestID:    runtime.RequestID(w.ctx),
	}
	defer func() {
		if err := chatquestion.NewChatQuestionDbDatasetDao().Insert(w.ctx, questionDatasetEntity); err != nil {
			logs.ErrorContextf(w.ctx, "[ExcelChat] Failed to insert question dataset: %v", err)
		}
	}()
	// 获取sheet list
	sheetEntityList, err := w.GetExcelSheetList(w.ctx, w.session)
	if err != nil {
		return err
	}
	// 构建问答 db
	var forestID uint
	if len(sheetEntityList) > 0 {
		forestID = sheetEntityList[0].ForestID
	}

	excelDB, err := w.GetExcelDB(forestID)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ExcelChat] Failed to get mysql db: %v", err)
		return err
	}

	sqlDB, err := excelDB.DB()
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ExcelChat] Failed to get mysql db: %v", err)
		return err
	}
	defer sqlDB.Close()

	// 获取 table ddl
	var forestTableIDs []uint
	for _, v := range sheetEntityList {
		forestTableIDs = append(forestTableIDs, v.ForestTableID)
	}
	tableDDLMap, err := w.GetExcelMysqlTableDDL(w.ctx, excelDB, forestTableIDs)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ExcelChat] Failed to get table ddl: %v", err)
		return err
	}
	choiceTableList, err := w.ChoiceTableList(w.question.Source.Question, tableDDLMap)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ExcelChat] Failed to choice table list: %v", err)
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
		logs.InfoContextf(w.ctx, "[ExcelChat] running question: %s", w.question.Source.Question)
		// 执行问题对应的sql
		questionSQL, err = w.GetQuestionMysqlSQL(w.ctx, excelDB, w.question.Source.Question, choiceDDLMap)
		if err != nil {
			return true, err
		}
		if err := excelDB.WithContext(w.ctx).Raw(questionSQL).Scan(&sqlResList).Error; err != nil {
			return true, fmt.Errorf("query question sql failed, err: %v", err)
		}
		return false, nil
	})

	if retryErr != nil {
		return retryErr
	}

	logs.InfoContextf(w.ctx, "[ExcelChat] question sql: %s", questionSQL)
	questionDatasetEntity.QueryStatement = questionSQL
	questionDatasetEntity.QueryResult = logs.JSON(sqlResList)

	if len(sqlResList) == 0 {
		logs.WarnContextf(w.ctx, "no rows in sql, sql: %s", questionSQL)
		if answer == "" {
			answer = "未找到相关数据，请调整问题后重新提问。"
		}
		llmchat.WriteContent(w.ctx, w.question.Source.ReqID, answer)
		w.question.Source.Answer = answer
		w.question.Source.Status = chattype.QuestionStatusAnswered
		return nil
	}

	// 将 sql 结果转为 answer
	ddlList := make([]string, 0, len(choiceDDLMap))
	for _, tableDDL := range choiceDDLMap {
		ddlList = append(ddlList, tableDDL)
	}
	var wg sync.WaitGroup
	var sqlAnswerRes *llmchat.QaRes
	var transferSqlAnswerErr error
	var echarts string
	var echartsErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		sqlAnswerRes, transferSqlAnswerErr = w.TransferMysqlResultToAnswer(w.ctx, w.question.Source.Question, questionSQL, ddlList, sqlResList)
		if transferSqlAnswerErr != nil {
			logs.ErrorContextf(w.ctx, "[ExcelChat] Failed to search answer: %v", transferSqlAnswerErr)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		echarts, echartsErr = w.GetEcharts(w.ctx, w.question.Source.Question, sqlResList)
		if echartsErr != nil {
			logs.ErrorContextf(w.ctx, "[ExcelChat] Failed to get echarts: %v", echartsErr)
		}
	}()

	// 等待两个 goroutine 都完成
	wg.Wait()

	// 返回错误（如果有的话）
	if transferSqlAnswerErr != nil {
		return fmt.Errorf("failed to get answer, err: %v, echarts err: %v", transferSqlAnswerErr, echartsErr)
	}

	// 处理 echarts 结果
	if echartsErr == nil && echarts != "" {
		questionDatasetEntity.EchartsDataset = echarts
		llmchat.WriteStreamsResult(w.ctx, w.question.Source.ReqID, llmchat.WriteResult{
			Content: echarts,
			Flag:    llmchat.StreamsFlagECharts,
		})
	}

	if sqlAnswerRes != nil {
		answer = sqlAnswerRes.Content
		answer = answer + "\n" + echarts
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
	return err
}

func (w *ChatWapper) GetExcelSheetList(ctx *gin.Context, sessionEntity *chattype.ChatSession) (foresttype.ForestExcelSheetList, error) {
	sheetCond := &forest.ForestExcelSheetCond{
		Enable: types.True,
	}
	if len(sessionEntity.ForestIDList.Slice()) > 0 {
		sheetCond.ForestIDs = sessionEntity.ForestIDList.Slice()
	}
	if len(sessionEntity.ExcelIDList.Slice()) > 0 {
		sheetCond.ForestFileIDs = sessionEntity.ExcelIDList.Slice()
	}
	if len(sessionEntity.ExcelSheetIDList.Slice()) > 0 {
		sheetCond.ParentIDs = sessionEntity.ExcelSheetIDList.Slice()
	}
	sheetEntityList, err := forest.NewForestExcelSheetDao().GetListByCond(ctx, sheetCond)
	if err != nil {
		return nil, fmt.Errorf("getSheetList failed, err: %v", err)
	}

	return sheetEntityList, nil
}

func (w *ChatWapper) GetExcelMysqlTableDDL(ctx *gin.Context, db *gorm.DB, forestTableIDs []uint) (map[string]string, error) {
	tableEntityList, err := forest.NewForestTableDao().GetListByCond(ctx, &forest.ForestTableCond{
		IDs: forestTableIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("get table by ids failed, err: %v, forestTableIDs:%s", err, logs.JSON(forestTableIDs))
	}
	var tableNameList []string
	for _, v := range tableEntityList {
		tableNameList = append(tableNameList, v.Tablename)
	}
	tableDDLMap := make(map[string]string)
	for _, tableName := range tableNameList {
		var ddl TableDDL
		if err := db.WithContext(ctx).Raw("SHOW CREATE TABLE " + tableName).Scan(&ddl).Error; err != nil {
			return nil, fmt.Errorf("get table ddl failed, err: %v", err)
		}
		tableDDLMap[ddl.Table] = ddl.CreateTable
	}
	return tableDDLMap, nil
}

func (w *ChatWapper) GetExcelDB(forestID uint) (*gorm.DB, error) {
	dbInstanceEntity, err := forest.NewForestDBInstanceDao().GetByCond(w.ctx, &forest.ForestDBInstanceCond{
		ForestID: forestID,
	})
	if err != nil {
		logs.ErrorContextf(w.ctx, "[GetExcelDB] Failed to get db instance, forestID: %d, err: %v", forestID, err)
		return nil, err
	}
	dbEntity, err := forest.NewForestDBDao().GetByCond(w.ctx, &forest.ForestDBCond{
		ForestID:           forestID,
		ForestDBInstanceID: dbInstanceEntity.ID,
	})
	if err != nil {
		logs.ErrorContextf(w.ctx, "[GetExcelDB] Failed to get db, forestID: %d, err: %v", forestID, err)
		return nil, err
	}
	mysqlConfig := &config.MysqlConfig{}
	if err := settings.GetYaml("knowledge", global.DBInstanceSystemReadonlySettingKey, mysqlConfig); err != nil {
		logs.ErrorContextf(w.ctx, "[GetExcelDB] failed get mysql config, forestID: %d, err: %v", err)
		return nil, fmt.Errorf("get mysql config failed, err = %v", err)
	}
	mysqlPlugin := &mysqlplugin.MySQLPlugin{}
	dbPluginConfig := &dbplugins.PluginConfig{
		Credentials: &dbplugins.Credentials{
			ConnectionID: fmt.Sprintf("%d", forestID),
			Hostname:     mysqlConfig.Host,
			Username:     mysqlConfig.Username,
			Password:     mysqlConfig.Password,
			Port:         uint(mysqlConfig.Port),
			Database:     dbEntity.DBName,
		},
	}
	excelDB, err := mysqlPlugin.DB(w.ctx, dbPluginConfig)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[GetExcelDB] Failed to get mysql db, err: %v", err)
		return nil, err
	}
	return excelDB, nil
}
