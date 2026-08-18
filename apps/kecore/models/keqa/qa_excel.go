package keqa

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	chatagent2 "github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatclient"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/agentclient"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type excelQA struct{}

// Chat excel 问答
func (qa *excelQA) Chat(ctx *gin.Context, qs *foresttype.KnownowForestQA, session *foresttype.KnownowQASession) (*foresttype.KnownowForestQA, error) {
	logs.InfoContextf(ctx, "[excelQA] Chat: %v,session: %v", qs.Question, session)
	if session.Name == foresttype.DefaultSessionName {
		runes := []rune(qs.Question)
		session.Name = qs.Question
		if len(runes) > 10 {
			session.Name = string(runes[:10])
		}
		logs.InfoContextf(ctx, "[excelQA] session.Name: %v", session.Name)
		err := ModifySession(ctx, session)
		if err != nil {
			logs.ErrorContextf(ctx, "[excelQA] Failed to ModifySession: %v", err)
		}
	}
	// 获取sheet list
	sheetEntityList, err := qa.getSheetList(ctx, session)
	if err != nil {
		return nil, err
	}

	// 构建问答 db
	var forestID uint
	if len(sheetEntityList) > 0 {
		forestID = sheetEntityList[0].ForestID
	}
	dbInstanceEntity, err := forest.NewForestDBInstanceDao().GetByCond(ctx, &forest.ForestDBInstanceCond{
		ForestID: forestID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[Chat] Failed to get db instance: %v", err)
		return nil, err
	}
	dbEntity, err := forest.NewForestDBDao().GetByCond(ctx, &forest.ForestDBCond{
		ForestID:           forestID,
		ForestDBInstanceID: dbInstanceEntity.ID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[Chat] Failed to get db instance: %v", err)
		return nil, err
	}

	excelDB, err := dbutil.GetDB(dbEntity.DBName, dbInstanceEntity.BuildMysqlDNS(dbEntity.DBName, dbEntity.DBMeta.Mysql.Charset))
	if err != nil {
		logs.ErrorContextf(ctx, "[AnalyzeXlsx] get db error: %v", err)
		return nil, err
	}

	// 获取 table ddl
	tableDDLMap, err := qa.getTableDDL(ctx, excelDB, sheetEntityList)
	if err != nil {
		return nil, err
	}
	var sqlResList []map[string]interface{}
	var questionSQL string
	retryErr := lifecycle.Retry(time.Second*5, 3, func() (needRetry bool, err error) {
		logs.InfoContextf(ctx, "[Chat] running question: %s", qs.Question)
		// 执行问题对应的sql
		questionSQL, err = qa.getQuestionSQL(ctx, excelDB, qs.Question, tableDDLMap)
		if err != nil {
			return true, err
		}
		if err := excelDB.WithContext(ctx).Raw(questionSQL).Scan(&sqlResList).Error; err != nil {
			return true, fmt.Errorf("query question sql failed, err: %v", err)
		}
		return false, nil
	})

	if retryErr != nil {
		return nil, retryErr
	}
	var answer string
	if len(sqlResList) == 0 {
		logs.WarnContextf(ctx, "no rows in sql, sql: %s", questionSQL)
		answer = "没有找到结果"
	}

	// 将 sql 结果转为 answer
	ddlList := make([]string, 0, len(tableDDLMap))
	for _, tableDDL := range tableDDLMap {
		ddlList = append(ddlList, tableDDL)
	}
	answer, err = qa.transferSQLResultToAnswer(ctx, qs.ID, qs.Question, questionSQL, ddlList, sqlResList)
	if err != nil {
		logs.ErrorContextf(ctx, "[ForestChat] Failed to search answer: %v", err)
		qs.Status = foresttype.QAStatusFailed
	}
	// 更新 QA 相关表
	qs.Answer = answer
	qs.Status = foresttype.QAStatusAnswered

	if err := dbutil.Knownow().WithContext(ctx).Save(qs).Error; err != nil {
		logs.ErrorContextf(ctx, "[ForestChat] Failed to save QA: %v", err)
		return nil, err
	}
	return qs, nil
}

func (qa *excelQA) getSheetList(ctx *gin.Context, sessionEntity *foresttype.KnownowQASession) (foresttype.ForestExcelSheetList, error) {
	sheetCond := &forest.ForestExcelSheetCond{}
	if len(sessionEntity.ForestIDList.Slice()) > 0 {
		sheetCond.ForestIDs = sessionEntity.ForestIDList.Slice()
	}
	if len(sessionEntity.ExcelIDList.Slice()) > 0 {
		sheetCond.ForestFileIDs = sessionEntity.ExcelIDList.Slice()
	}
	if len(sessionEntity.ExcelSheetIDList.Slice()) > 0 {
		sheetCond.IDs = sessionEntity.ExcelSheetIDList.Slice()
	}
	sheetEntityList, err := forest.NewForestExcelSheetDao().GetListByCond(ctx, sheetCond)
	if err != nil {
		return nil, fmt.Errorf("getSheetList failed, err: %v", err)
	}

	return sheetEntityList, nil
}

func (qa *excelQA) getTableDDL(ctx *gin.Context, db *gorm.DB, sheetEntityList foresttype.ForestExcelSheetList) (map[string]string, error) {
	var tableIDs []uint
	for _, v := range sheetEntityList {
		tableIDs = append(tableIDs, v.ForestTableID)
	}
	tableEntityList, err := forest.NewForestTableDao().GetListByCond(ctx, &forest.ForestTableCond{
		IDs: tableIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("get table by ids failed, err: %v, tableIDs:%s", err, logs.JSON(tableIDs))
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

func (qa *excelQA) getQuestionSQL(ctx *gin.Context, db *gorm.DB, question string, tableDDLMap map[string]string) (string, error) {
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
	agentInputList := []chattype.Input{
		{
			Name:  "input1",
			Value: strings.Join(dataAndDDLList, "\n"),
		},
		{
			Name:  "input2",
			Value: question,
		},
	}
	agentCfg, err := agentclient.GetLLMConfig(ctx, global.SettingGroupKnowledge, global.SettingKeyAgentExcelQuestionToSQL)
	if err != nil {
		return "", fmt.Errorf("get llm config failed: %w", err)
	}
	client := agentclient.NewChatClientWithConfig(nil, agentCfg)
	req := &agentclient.ChatRequestBody{
		Model: chatagent2.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentExcelQuestionToSql),
		ChatOptions: agentclient.ChatOptions{
			Input: agentInputList,
		},
		Stream: false, // 明确设置为非流式
	}
	resp, err := client.SendChat(ctx, req)
	if err != nil {
		return "", fmt.Errorf("request chat err: %v", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no sql")
	}

	sqlContent := resp.Choices[0].Message.Content
	sql := chatclient.ExtractCode("sql", sqlContent)
	return sql, nil
}

func (qa *excelQA) transferSQLResultToAnswer(ctx *gin.Context, questionID uint, question, sql string, ddlList []string, sqlResList []map[string]interface{}) (string, error) {
	resBytes, _ := json.Marshal(sqlResList)
	agentInputList := []chattype.Input{
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
	}
	agentCfg, err := agentclient.GetLLMConfig(ctx, global.SettingGroupKnowledge, global.SettingKeyAgentExcelSQLResultAnalysis)
	if err != nil {
		return "", fmt.Errorf("get llm config failed: %w", err)
	}
	agentClient := agentclient.NewChatClientWithConfig(nil, agentCfg)
	req := &agentclient.ChatRequestBody{
		Model: chatagent2.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentExcelSqlResultAnalysis),
		ChatOptions: agentclient.ChatOptions{
			Input: agentInputList,
		},
	}
	streamsKey := fmt.Sprintf("%d", questionID)
	sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()), sseclient.WithBlockMaxRetry(30))
	defer sseClient.Close(ctx, streamsKey)

	var contentBuilder strings.Builder
	err = agentClient.SendChatStreamWithCallback(ctx, req, func(chunk *agentclient.ChatStreamResponseBody) error {
		for _, v := range chunk.Choices {
			select {
			case <-ctx.Request.Context().Done():
				logs.InfoContextf(ctx, "[transferSQLResultToAnswer] context done")
				return ctx.Request.Context().Err()
			default:
			}
			contentBuilder.WriteString(v.Delta.Content)
			writeResult := WriteResult{
				Content: v.Delta.Content,
			}
			if stopped, err := sseClient.WriteMessage(ctx, ctx.Writer, streamsKey, writeResult.String()); err != nil {
				if strings.Contains(err.Error(), "broken pipe") {
					continue
				}
				logs.ErrorContextf(ctx, "[transferSQLResultToAnswer] Failed to write Answering response to KEQA: %v", err)
				return err
			} else if stopped {
				logs.InfoContextf(ctx, "[transferSQLResultToAnswer] stream has been stopped")
				return nil
			}
		}
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[transferSQLResultToAnswer] SendChatStreamWithCallback failed, err: %v, questionId: %d", err, questionID)
		return contentBuilder.String(), err
	}

	return contentBuilder.String(), err
}
