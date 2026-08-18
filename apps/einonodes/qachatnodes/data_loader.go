package qachatnodes

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/cloudwego/eino/compose"
	"github.com/insmtx/corekg/apps/einonodes/nodebase"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/keqa"
	"github.com/insmtx/corekg/apps/kechat/models/qachat"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/connectors/tokenmgr"
	"github.com/insmtx/corekg/pkgs/einotools/confluence"
	"github.com/insmtx/corekg/pkgs/einotools/gmail"
	"github.com/insmtx/corekg/pkgs/einotools/googledrive"
	"github.com/insmtx/corekg/pkgs/einotools/slack"
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/ygpkg/yg-go/logs"
)

// dataLoader 数据加载器(如问答上下文)
type dataLoader struct {
	*baseHandler
}

func newDataLoader() *dataLoader {
	return &dataLoader{
		baseHandler: &baseHandler{},
	}
}

// QAPairNode 问答对加载节点
func (d *dataLoader) QAPairNode(ctx context.Context, input string, opts ...any) (output string, err error) {
	err = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		defer func() {
			output = state.Goto
		}()

		history, err := d.getForestChatHistory(ctx, state.SessionEntity)
		if err != nil {
			logs.ErrorContextf(ctx, "[QAPairNode] getForestChatHistory error: %v", err)
			return err
		}
		state.HistoryContext = history

		forestIDs := state.SessionEntity.ForestIDList.Slice()
		fileIDs := state.SessionEntity.FileIDList.Slice()
		esIndex := state.SessionEntity.EsIndex

		wrapper, err := keqa.HandelSearchReference(ctx, forestIDs, fileIDs, esIndex, state.QuestionEntity.Source.Question)
		if err != nil {
			logs.ErrorContextf(ctx, "[QAPairNode] Failed to HandelSearchReference: %v", err)
			return err
		}
		// 查找问答对
		qaPair, err := wrapper.SearchWrapper.FindFQAByQuestion()
		if err != nil {
			logs.ErrorContextf(ctx, "[QAPairNode] Failed to FindFQAByQuestion: %v", err)
			return err
		}
		if len(qaPair.Hits.Hits) != 0 {
			logs.InfoContextf(ctx, "[QAPairNode] FindFQAByQuestion result: %v", len(qaPair.Hits.Hits))
			state.QuestionEntity.Source.Answer = qaPair.Hits.Hits[0].Source.QAAnswer
			state.QuestionEntity.Source.Status = chattype.QuestionStatusAnswered
			return nil

		}
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[QAPairNode] ProcessState error: %v", err)
		return "", err
	}
	return output, nil
}

// QueryChunkReferenceNode 查询参考文本节点
func (d *dataLoader) QueryChunkReferenceNode(ctx context.Context, input string, opts ...any) (output string, err error) {
	err = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		forestIDs := state.SessionEntity.ForestIDList.Slice()
		fileIDs := state.SessionEntity.FileIDList.Slice()
		esIndex := state.SessionEntity.EsIndex
		wrapper, err := keqa.HandelSearchReference(ctx, forestIDs, fileIDs, esIndex, state.QuestionEntity.Source.Question)
		if err != nil {
			logs.ErrorContextf(ctx, "[QueryChunkReferenceNode] Failed to HandelSearchReference: %v", err)
			return err
		}
		intention := input
		var searchStr string
		switch intention {
		case "C", "c":
			preSearchResult, _, err := wrapper.SearchFileDescriptions(state.QuestionEntity.Source.Question + state.QuestionEntity.Source.ImageContent)
			if err != nil {
				logs.ErrorContextf(ctx, "[QueryChunkReferenceNode] Failed to SearchFileDescriptions: %v", err)
				return err
			}
			refList := wrapper.SupQuestionChunk(preSearchResult)
			state.QueryReferenceList = &refList
			state.QuestionEntity.Source.QueryReferenceList = &refList

			if len(refList) > 150 {
				searchStr = keqa.BatchSummaryReference(state.Ctx, state.QuestionEntity, refList)
			} else {
				res, err := keqa.TransformChatReferenceList(refList)
				if err != nil {
					logs.ErrorContextf(ctx, "[QueryChunkReferenceNode] Failed to TransformChatReferenceList: %v", err)
					return err
				}
				searchStr = string(res)
			}
		default:
			res, err := wrapper.RerankSearchQuestionChunk(nil)
			if err != nil {
				logs.ErrorContextf(ctx, "NewRerankSearchWrapper error: %v", err)
				return err
			}
			chunkList := chattype.ChatReferenceList{}
			for _, v := range res {
				chunks := map[int]string{}
				for _, c := range v.ChunkList {
					chunks[c.Sequence] = c.Content
				}
				chunkList.Reference = append(chunkList.Reference, chattype.ChatReference{
					FileID:   v.FileID,
					Abstract: v.Abstract,
					Chunks:   chunks,
				})
			}
			state.QuestionEntity.Source.QueryReferenceList = &res
			state.QuestionEntity.Source.ChatReferenceList = &chunkList
			// 序列化成字符串
			jsonString, err := json.Marshal(chunkList.Reference)
			if err != nil {
				logs.ErrorContextf(state.Ctx, "json.Marshal err:%v", err)
				return err
			}
			searchStr = string(jsonString)
		}
		output = searchStr
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[QueryChunkReferenceNode] ProcessState error: %v", err)
		return "", err
	}
	return output, nil
}

// QueryChunkReferenceNode 查询参考文本节点
func (d *dataLoader) ExcelDDLNode(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {
	err = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		if err := d.checkState(state); err != nil {
			logs.ErrorContextf(ctx, "[ExcelDDLNode] check state fail, err: %v", err)
			return err
		}
		wrapper := qachat.NewChatWrapper(state.Ctx, state.QuestionEntity, state.SessionEntity, state.ModelEntity)
		sheetEntityList, err := wrapper.GetExcelSheetList(state.Ctx, state.SessionEntity)
		if err != nil {
			logs.ErrorContextf(ctx, "[ExcelDDLNode] Failed to GetExcelSheetList: %v", err)
			return err
		}
		if len(sheetEntityList) == 0 {
			logs.WarnContextf(ctx, "[ExcelDDLNode] sheetEntityList is empty, questionID: %s", state.QuestionEntity.ID)
		}

		var forestID uint
		var forestTableIDs []uint
		for _, v := range sheetEntityList {
			if forestID == 0 {
				forestID = v.ForestID
			}
			forestTableIDs = append(forestTableIDs, v.ForestTableID)
		}

		excelDB, err := wrapper.GetExcelDB(forestID)
		if err != nil {
			logs.ErrorContextf(ctx, "[ExcelDDLNode] Failed to GetExcelDB: %v", err)
			return err
		}
		sqlDB, err := excelDB.DB()
		if err != nil {
			logs.ErrorContextf(ctx, "[ExcelDDLNode] Failed to GetExcelDB: %v", err)
			return err
		}
		defer sqlDB.Close()

		tableDDLMap, err := wrapper.GetExcelMysqlTableDDL(state.Ctx, excelDB, forestTableIDs)
		if err != nil {
			logs.ErrorContextf(ctx, "[ExcelDDLNode] Failed to GetExcelMysqlTableDDL: %v", err)
			return err
		}
		choiceTableList, err := wrapper.ChoiceTableList(state.QuestionEntity.Source.Question, tableDDLMap)
		if err != nil {
			logs.ErrorContextf(ctx, "[ExcelDDLNode] Failed to ChoiceTableList: %v", err)
			return err
		}
		state.QuestionDbDatasetEntity.TableList = logs.JSON(choiceTableList)

		choiceDDLMap := make(map[string]string, len(choiceTableList))
		for _, v := range choiceTableList {
			choiceDDLMap[v] = tableDDLMap[v]
		}

		// 维护上下文
		output = nodebase.RecordList{}
		output.Add(&nodebase.Record{
			Key:   RecordKeyForestID,
			Value: strconv.Itoa(int(forestID)),
		})
		output.Add(&nodebase.Record{
			Key:   RecordKeyMySQLTableDDLMap,
			Value: logs.JSON(choiceDDLMap),
		})
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[ExcelDDLNode] ProcessState error: %v", err)
		return nil, err
	}
	return output, nil
}

func (d *dataLoader) getForestChatHistory(ctx context.Context, session *chattype.ChatSession) (string, error) {
	quesrions, err := chatquestion.ListSessionQuestionsByUin(ctx, session.Uin, session.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestChatHistory ListSessionQuestions error: %v", err)
		return "", err
	}
	type ForestChatHistory struct {
		Question string `json:"question"`
		Answer   string `json:"answer"`
	}
	chats := []*ForestChatHistory{}
	for _, qa := range quesrions {
		if qa.Source.Status != chattype.QuestionStatusAnswered {
			continue
		}
		chats = append(chats, &ForestChatHistory{
			Question: qa.Source.Question,
			Answer:   qa.Source.Answer,
		})
	}
	chatsJSON, err := json.Marshal(chats)
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestChatHistory json.Marshal error: %v", err)
		return "", err
	}
	return string(chatsJSON), nil
}

// QueryChunkReferenceNode 查询参考文本节点
func (d *dataLoader) GmailDataNode(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {
	err = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		if err := d.checkState(state); err != nil {
			logs.ErrorContextf(ctx, "[GmailDataNode] check state fail, err: %v", err)
			return err
		}
		useGmail := input.Get(RecordKeyUseGmail)
		if useGmail.Value == "" {
			return nil
		}
		keyWordStr := input.Get(RecordKeyESKeyWords)
		if keyWordStr.Value == "" {
			return nil
		}
		keyWord := types.NewStringArray([]string{})
		err := keyWord.UnmarshalJSON([]byte(keyWordStr.Value))
		if err != nil {
			logs.ErrorContextf(ctx, "[GmailDataNode] keyWord.UnmarshalJSON error: %v", err)
			return err
		}
		// TODO TO CoreSettings
		conf := &gmail.Config{
			MaxResults: 3,
		}
		gmailTool, err := gmail.NewTool(ctx, conf)
		if err != nil {
			logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] new gmail tool error: %v", err)
			return err
		}
		gsr := &gmail.SearchRequest{
			Uin:        state.QuestionEntity.Source.Uin,
			Queries:    keyWord.Slice(),
			MaxResults: 3,
		}
		toolOut, err := gmailTool.InvokableRun(ctx, gsr.String())
		if err != nil {
			logs.ErrorContextf(ctx, "[ExternalDataChatToolsBuilder] gmail tool InvokableRun error: %v", err)
			return err
		}
		logs.InfoContextf(ctx, "GmailDataNode toolOut: %s", toolOut)
		// 更新搜索到的结果
		chunk := &chattype.QueryReference{
			DataSourceType: "gmail",
			ChunkList: chattype.QueryReferenceChunkList{chattype.QueryReferenceChunk{
				Content: toolOut,
			}},
		}
		// 先确保不为nil
		if state.QuestionEntity.Source.QueryReferenceList == nil {
			state.QuestionEntity.Source.QueryReferenceList = &chattype.QueryReferenceList{}
		}
		// 创建新的切片并赋值
		newList := append(*state.QuestionEntity.Source.QueryReferenceList, chunk)
		state.QuestionEntity.Source.QueryReferenceList = &newList
		output.Add(&nodebase.Record{
			Key:   RecordKeyGmailData,
			Value: toolOut,
		})
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[GmailDataNode] ProcessState error: %v", err)
		return nil, err
	}
	return output, nil
}

// QueryChunkReferenceNode 查询参考文本节点
func (d *dataLoader) TestNode(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {
	err = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		useTest := input.Get(RecordKeyUseTest)
		if useTest.Value == "" {
			return nil
		}
		logs.InfoContextf(ctx, "GmailDataNode toolOut: ------------------------test")
		// output = nodebase.RecordList{}
		output.Add(&nodebase.Record{
			Key:   "TestData",
			Value: "testtesttest",
		})
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[TestNode] ProcessState error: %v", err)
		return nil, err
	}
	return output, nil
}

// QueryChunkReferenceNode 查询参考文本节点
func (d *dataLoader) GetSessionToolsNode(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {
	err = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		if err := d.checkState(state); err != nil {
			logs.ErrorContextf(ctx, "[GetSessionToolsNode] check state fail, err: %v", err)
			return err
		}
		// 获取session的tokenlist
		tokens, err := tokenmgr.GetTokensByIDS(ctx, state.SessionEntity.ExternalTokenIDList.Slice())
		if err != nil {
			logs.ErrorContextf(ctx, "[GetSessionToolsNode] GetTokensByIDS error: %v", err)
			return err
		}
		// 继承入参传入的值
		for _, v := range input {
			output.Add(v)
		}
		// 获取都有啥什么工具
		// TODO 后期如果同平台多账号需要特殊处理
		for _, v := range tokens {
			switch v.Platform {
			case tokenmgr.PlatformGmail:
				output.Add(&nodebase.Record{
					Key:   RecordKeyUseGmail,
					Value: RecordKeyUseGmail,
				})
			case tokenmgr.PlatformSlack:
				output.Add(&nodebase.Record{
					Key:   RecordKeyUseSlack,
					Value: RecordKeyUseSlack,
				})
			case tokenmgr.PlatformGoogleDrive:
				output.Add(&nodebase.Record{
					Key:   RecordKeyUseGoolgeDrive,
					Value: RecordKeyUseGoolgeDrive,
				})
			}
		}
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[GetSessionToolsNode] ProcessState error: %v", err)
		return nil, err
	}
	return output, nil
}

// GetESAnalyzeNode 获取es分词结果
func (d *dataLoader) GetESAnalyzeNode(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {
	err = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		if err := d.checkState(state); err != nil {
			logs.ErrorContextf(ctx, "[GetESAnalyzeNode] check state fail, err: %v", err)
			return err
		}
		keywords, err := essearch.Analyze(ctx, state.QuestionEntity.Source.Question)
		if err != nil {
			logs.ErrorContextf(ctx, "[GetESAnalyzeNode] essearch.Analyze error: %v", err)
			return err
		}
		words := types.NewStringArray([]string{})
		for _, v := range keywords.Tokens {
			words.Append(v.Token)
		}
		jsonword, err := words.MarshalJSON()
		if err != nil {
			logs.ErrorContextf(ctx, "[GetESAnalyzeNode] words.MarshalJSON error: %v", err)
			return err
		}
		for _, v := range input {
			output.Add(v)
		}
		output.Add(&nodebase.Record{
			Key:   RecordKeyESKeyWords,
			Value: string(jsonword),
		})
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[GetSessionToolsNode] ProcessState error: %v", err)
		return nil, err
	}
	return output, nil
}

// SlackDataNode slack查询
func (d *dataLoader) SlackDataNode(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {
	err = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		if err := d.checkState(state); err != nil {
			logs.ErrorContextf(ctx, "[SlackDataNode] check state fail, err: %v", err)
			return err
		}
		useSlack := input.Get(RecordKeyUseSlack)
		if useSlack.Value == "" {
			return nil
		}
		keyWordStr := input.Get(RecordKeyESKeyWords)
		if keyWordStr.Value == "" {
			return nil
		}
		keyWord := types.NewStringArray([]string{})
		err := keyWord.UnmarshalJSON([]byte(keyWordStr.Value))
		if err != nil {
			logs.ErrorContextf(ctx, "[SlackDataNode] keyWord.UnmarshalJSON error: %v", err)
			return err
		}
		conf := &slack.Config{
			MaxResults: 5,
		}
		slackTool, err := slack.NewTool(ctx, conf)
		if err != nil {
			logs.ErrorContextf(ctx, "[SlackDataNode] new gmail tool error: %v", err)
			return err
		}
		slackReq := &slack.SearchRequest{
			Uin:        state.QuestionEntity.Source.Uin, // Test UIN
			Queries:    keyWord.Slice(),
			MaxResults: 3,
		}
		toolOut, err := slackTool.InvokableRun(ctx, slackReq.String())
		if err != nil {
			logs.ErrorContextf(ctx, "[SlackDataNode] gmail tool InvokableRun error: %v", err)
			return err
		}
		logs.InfoContextf(ctx, "SlackDataNode toolOut: %s", toolOut)
		// 更新搜索到的结果
		chunk := &chattype.QueryReference{
			DataSourceType: "slack",
			ChunkList: chattype.QueryReferenceChunkList{chattype.QueryReferenceChunk{
				Content: toolOut,
			}},
		}
		// 先确保不为nil
		if state.QuestionEntity.Source.QueryReferenceList == nil {
			state.QuestionEntity.Source.QueryReferenceList = &chattype.QueryReferenceList{}
		}
		// 创建新的切片并赋值
		newList := append(*state.QuestionEntity.Source.QueryReferenceList, chunk)
		state.QuestionEntity.Source.QueryReferenceList = &newList
		output.Add(&nodebase.Record{
			Key:   RecordKeySlackData,
			Value: toolOut,
		})
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[SlackDataNode] ProcessState error: %v", err)
		return nil, err
	}
	return output, nil
}

// GoogleDriveDataNode google drive查询
func (d *dataLoader) GoogleDriveDataNode(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {
	err = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		if err := d.checkState(state); err != nil {
			logs.ErrorContextf(ctx, "[GoogleDriveDataNode] check state fail, err: %v", err)
			return err
		}
		useGoogleDrive := input.Get(RecordKeyUseGoolgeDrive)
		if useGoogleDrive.Value == "" {
			return nil
		}
		keyWordStr := input.Get(RecordKeyESKeyWords)
		if keyWordStr.Value == "" {
			return nil
		}
		keyWord := types.NewStringArray([]string{})
		err := keyWord.UnmarshalJSON([]byte(keyWordStr.Value))
		if err != nil {
			logs.ErrorContextf(ctx, "[GoogleDriveDataNode] keyWord.UnmarshalJSON error: %v", err)
			return err
		}
		conf := &googledrive.Config{
			MaxResults: 5,
		}
		googledriveTool, err := googledrive.NewTool(ctx, conf)
		if err != nil {
			logs.ErrorContextf(ctx, "[GoogleDriveDataNode] new gmail tool error: %v", err)
			return err
		}
		driveReq := &googledrive.SearchRequest{
			Uin:        state.QuestionEntity.Source.Uin, // Test UIN
			Queries:    keyWord.Slice(),
			MaxResults: 3,
		}
		toolOut, err := googledriveTool.InvokableRun(ctx, driveReq.String())
		if err != nil {
			logs.ErrorContextf(ctx, "[GoogleDriveDataNode] gmail tool InvokableRun error: %v", err)
			return err
		}
		logs.InfoContextf(ctx, "GoogleDriveDataNode toolOut: %s", toolOut)
		// 更新搜索到的结果
		chunk := &chattype.QueryReference{
			DataSourceType: "googledrive",
			ChunkList: chattype.QueryReferenceChunkList{chattype.QueryReferenceChunk{
				Content: toolOut,
			}},
		}
		// 先确保不为nil
		if state.QuestionEntity.Source.QueryReferenceList == nil {
			state.QuestionEntity.Source.QueryReferenceList = &chattype.QueryReferenceList{}
		}
		// 创建新的切片并赋值
		newList := append(*state.QuestionEntity.Source.QueryReferenceList, chunk)
		state.QuestionEntity.Source.QueryReferenceList = &newList
		output.Add(&nodebase.Record{
			Key:   RecordKeyGoogleDriveData,
			Value: toolOut,
		})
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[GoogleDriveDataNode] ProcessState error: %v", err)
		return nil, err
	}
	return output, nil
}

// ConfluenceDataNode confluence查询
func (d *dataLoader) ConfluenceDataNode(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {
	err = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		if err := d.checkState(state); err != nil {
			logs.ErrorContextf(ctx, "[ConfluenceDataNode] check state fail, err: %v", err)
			return err
		}
		useConfluence := input.Get(RecordKeyUseConfluence)
		if useConfluence.Value == "" {
			return nil
		}
		keyWordStr := input.Get(RecordKeyESKeyWords)
		if keyWordStr.Value == "" {
			return nil
		}
		keyWord := types.NewStringArray([]string{})
		err := keyWord.UnmarshalJSON([]byte(keyWordStr.Value))
		if err != nil {
			logs.ErrorContextf(ctx, "[ConfluenceDataNode] keyWord.UnmarshalJSON error: %v", err)
			return err
		}
		conf := &confluence.Config{
			MaxResults: 5,
		}
		confluenceTool, err := confluence.NewTool(ctx, conf)
		if err != nil {
			logs.ErrorContextf(ctx, "[ConfluenceDataNode] new confluence tool error: %v", err)
			return err
		}
		driveReq := &confluence.SearchRequest{
			Uin:        state.QuestionEntity.Source.Uin,
			Queries:    keyWord.Slice(),
			MaxResults: 4,
		}
		toolOut, err := confluenceTool.InvokableRun(ctx, driveReq.String())
		if err != nil {
			logs.ErrorContextf(ctx, "[ConfluenceDataNode] confluence tool InvokableRun error: %v", err)
			return err
		}
		logs.InfoContextf(ctx, "ConfluenceDataNode toolOut: %s", toolOut)
		// 更新搜索到的结果
		chunk := &chattype.QueryReference{
			DataSourceType: "confluence",
			ChunkList: chattype.QueryReferenceChunkList{chattype.QueryReferenceChunk{
				Content: toolOut,
			}},
		}
		// 先确保不为nil
		if state.QuestionEntity.Source.QueryReferenceList == nil {
			state.QuestionEntity.Source.QueryReferenceList = &chattype.QueryReferenceList{}
		}
		// 创建新的切片并赋值
		newList := append(*state.QuestionEntity.Source.QueryReferenceList, chunk)
		state.QuestionEntity.Source.QueryReferenceList = &newList
		output.Add(&nodebase.Record{
			Key:   RecordKeyConfluenceData,
			Value: toolOut,
		})
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[ConfluenceDataNode] ProcessState error: %v", err)
		return nil, err
	}
	return output, nil
}
