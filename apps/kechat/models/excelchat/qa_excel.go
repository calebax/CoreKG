package excelchat

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatclient"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

func NewExcelChatWrapper() *excelQA {
	return &excelQA{}
}

type excelQA struct {
}

func (qa *excelQA) GetSheetList(ctx *gin.Context, sessionEntity *chattype.ChatSession) (foresttype.ForestExcelSheetList, error) {
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

func (qa *excelQA) GetTableDDL(ctx *gin.Context, db *gorm.DB, sheetEntityList foresttype.ForestExcelSheetList) (map[string]string, error) {
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

func (qa *excelQA) GetQuestionSQL(ctx *gin.Context, db *gorm.DB, question string, tableDDLMap map[string]string) (string, error) {
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

	w, err := chatclient.NewInternalChat(ctx, runtime.RequestID(ctx), "", 1, req)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to create internal chat: %v", err)
		return "", err
	}
	res, err := w.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(ctx, "agent chat error: %v", err)
		return "", err
	}
	sql := chatclient.ExtractCode("sql", res.Content)
	return sql, nil
}

func (qa *excelQA) TransferSQLResultToAnswer(ctx *gin.Context, question, sql string, ddlList []string, sqlResList []map[string]interface{}) (*llmchat.QaRes, error) {
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
	wrapper, err := chatclient.NewInternalChat(ctx, runtime.RequestID(ctx), "", 2, req)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to create internal chat: %v", err)
		return nil, err
	}
	res, err := wrapper.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(ctx, "agent chat error: %v", err)
		return nil, err
	}

	return res, err
}
