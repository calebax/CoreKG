package keqa

import (
	"encoding/json"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatclient"
	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/insmtx/corekg/apps/kesearch/models/reranksearch"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"golang.org/x/sync/errgroup"
)

func (w *ForestWrapper) DefaultChat(writeRef bool) (string, error) {
	preSearchResult, _, err := w.wrapper.PreSearchQuestionChunk()
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ForestChat] PreSearchQuestionChunk error: %v", err)
		return "", err
	}

	refList, err := w.wrapper.SupSearchQuestionChunk(preSearchResult)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ForestChat] SupSearchQuestionChunk error: %v", err)
		return "", err
	}
	if writeRef {
		WriteReferenceFile(w.ctx, refList, w.question.Source.ReqID)
	}
	logs.InfoContextf(w.ctx, "[ForestChat] SupSearchQuestionChunk result: %v", len(refList))
	w.refList = refList
	w.question.Source.QueryReferenceList = &refList
	searchStr, err := TransformChatReferenceList(refList)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ForestChat] TransformChatReferenceList error: %v", err)
		return "", err
	}
	w.searchStr = string(searchStr)
	return w.searchStr, err
}

// DescriptionChat 通过摘要回答总结性问题
func (w *ForestWrapper) DescriptionChat(writeRef bool) (string, error) {
	preSearchResult, _, err := w.wrapper.SearchFileDescriptions(w.question.Source.Question + w.question.Source.ImageContent)
	if err != nil {
		return "", err
	}
	refList := w.wrapper.SupQuestionChunk(preSearchResult)
	if writeRef {
		WriteReferenceFile(w.ctx, refList, w.question.Source.ReqID)
	}
	w.refList = refList
	w.question.Source.QueryReferenceList = &refList
	if len(refList) > 150 {
		w.batchSumDesc()
	} else {
		searchStr, err := TransformChatReferenceList(refList)
		if err != nil {
			return "", err
		}
		w.searchStr = string(searchStr)
	}

	return w.searchStr, nil
}

// DefaultRerankChat rerank默认
func (w *ForestWrapper) DefaultRerankChat(writeRef bool) (string, error) {
	res, err := w.wrapper.RerankSearchQuestionChunk(nil)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ForestChat] RerankSearchQuestionChunk error: %v", err)
		return "", err
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
	if writeRef {
		WriteReferenceFile(w.ctx, res, w.question.Source.ReqID)
	}
	w.question.Source.QueryReferenceList = &res
	w.question.Source.ChatReferenceList = &chunkList
	// 序列化成字符串
	jsonString, err := json.Marshal(chunkList.Reference)
	if err != nil {
		logs.ErrorContextf(w.ctx, "json.Marshal err:%v", err)
		return "", err
	}
	w.searchStr = string(jsonString)
	return w.searchStr, err
}

// GraphRerankChat rerank默认
func (w *ForestWrapper) GraphRerankChat(g *errgroup.Group, writeRef bool) (string, error) {
	res, err := w.wrapper.RerankSearchQuestionChunk(reranksearch.GraphSearchConf())
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ForestChat] RerankSearchQuestionChunk error: %v", err)
		return "", err
	}
	chunkList := chattype.ChatReferenceList{}
	chunkIDList := []string{}
	for _, v := range res {
		chunks := map[int]string{}
		for _, c := range v.ChunkList {
			chunks[c.Sequence] = c.Content
			chunkIDList = append(chunkIDList, c.ChunkID)
		}
		chunkList.Reference = append(chunkList.Reference, chattype.ChatReference{
			FileID:   v.FileID,
			Abstract: v.Abstract,
			Chunks:   chunks,
		})
	}
	g.Go(func() error {
		if len(chunkIDList) > 0 && res.Len() > 0 {
			// 判断当前知识库是否已经存在图谱
			graphInfo, err := graph.GetForestGraph(w.ctx, res[0].ForestID)
			if err != nil {
				logs.ErrorContextf(w.ctx, "GraphRerankChat.GetForestGraph err:%v")
				return err
			}
			chunkGraph, err := graph.SearchGraphWithChunkIDs(w.ctx, graphInfo, chunkIDList)
			if err != nil {
				logs.ErrorContextf(w.ctx, "GraphRerankChat.SearchGraphWithChunkIDs err:%v")
				return err
			}
			w.question.Source.GraphReference = chunkGraph
		}
		return nil
	})
	if writeRef {
		WriteReferenceFile(w.ctx, res, w.question.Source.ReqID)
	}
	w.question.Source.QueryReferenceList = &res
	w.question.Source.ChatReferenceList = &chunkList
	// 序列化成字符串
	jsonString, err := json.Marshal(chunkList.Reference)
	if err != nil {
		logs.ErrorContextf(w.ctx, "json.Marshal err:%v", err)
		return "", err
	}
	w.searchStr = string(jsonString)

	return w.searchStr, err
}

// sumDescAgent 根据es查到的desc总结
func SumDescAgent(ctx *gin.Context, question *chattype.ChatQuestion, searchStr string) (string, error) {
	req := &chattype.ChatRequestBody{
		Stream: false,
		Model:  chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentESChat),
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{Name: "input1", Value: question.Source.Question}, // 用户问题
				{Name: "input2", Value: searchStr},                // 知识库检索内容
				{Name: "input3", Value: ""},                       // 对话历史
			},
		},
	}
	wrapper, err := chatclient.NewInternalChat(ctx, question.Source.ReqID, "", 2, req)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to create internal chat: %v", err)
		return "", err
	}
	res, err := wrapper.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(ctx, "agent chat error: %v", err)
		return "", err
	}
	return res.Content, nil
}

// batchSumDesc 分批跑
func (w *ForestWrapper) batchSumDesc() {
	batchSize := 150
	batchCount := (len(w.refList) + batchSize - 1) / batchSize

	type batchResult struct {
		index  int
		answer string
		err    error
	}

	var wg sync.WaitGroup
	resultChan := make(chan batchResult, batchCount)
	// 并发执行每一批 ESChat
	for i := 0; i < len(w.refList); i += batchSize {
		start := i
		end := i + batchSize
		if end > len(w.refList) {
			end = len(w.refList)
		}
		subList := w.refList[start:end]
		batchIndex := i / batchSize
		wg.Add(1)
		go func(idx int, chunk chattype.QueryReferenceList) {
			logs.InfoContextf(w.ctx, "Batch %d: %v", idx+1, len(chunk))
			defer wg.Done()
			searchStr, err := TransformChatReferenceList(chunk)
			if err != nil {
				resultChan <- batchResult{index: idx, answer: "", err: err}
				return
			}
			answer, err := SumDescAgent(w.ctx, w.question, string(searchStr))
			logs.InfoContextf(w.ctx, "Batch %d: %s", idx+1, answer)
			resultChan <- batchResult{index: idx, answer: answer, err: err}
		}(batchIndex, subList)
	}

	// 等待所有 goroutine 结束并关闭通道
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果并排序（因为 goroutine 是并发的）
	batchAnswers := make([]string, batchCount)
	for res := range resultChan {
		if res.err != nil {
			logs.ErrorContextf(w.ctx, "ESChat batch %d error: %v", res.index+1, res.err)
			continue
		}
		batchAnswers[res.index] = res.answer
	}

	// 合并中间结果后进行最终对话
	w.searchStr = strings.Join(batchAnswers, "\n")

}

func BatchSummaryReference(ctx *gin.Context, chatQuestion *chattype.ChatQuestion, refList chattype.QueryReferenceList) string {
	batchSize := 150
	batchCount := (len(refList) + batchSize - 1) / batchSize

	type batchResult struct {
		index  int
		answer string
		err    error
	}

	var wg sync.WaitGroup
	resultChan := make(chan batchResult, batchCount)
	// 并发执行每一批 ESChat
	for i := 0; i < len(refList); i += batchSize {
		start := i
		end := i + batchSize
		if end > len(refList) {
			end = len(refList)
		}
		subList := refList[start:end]
		batchIndex := i / batchSize
		wg.Add(1)
		go func(idx int, chunk chattype.QueryReferenceList) {
			logs.InfoContextf(ctx, "Batch %d: %v", idx+1, len(chunk))
			defer wg.Done()
			searchStr, err := TransformChatReferenceList(chunk)
			if err != nil {
				resultChan <- batchResult{index: idx, answer: "", err: err}
				return
			}
			answer, err := SumDescAgent(ctx, chatQuestion, string(searchStr))
			logs.InfoContextf(ctx, "Batch %d: %s", idx+1, answer)
			resultChan <- batchResult{index: idx, answer: answer, err: err}
		}(batchIndex, subList)
	}

	// 等待所有 goroutine 结束并关闭通道
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果并排序（因为 goroutine 是并发的）
	batchAnswers := make([]string, batchCount)
	for res := range resultChan {
		if res.err != nil {
			logs.ErrorContextf(ctx, "ESChat batch %d error: %v", res.index+1, res.err)
			continue
		}
		batchAnswers[res.index] = res.answer
	}

	// 合并中间结果后进行最终对话
	return strings.Join(batchAnswers, "\n")

}

// ESChat 根据es查到的chunk来进行问答
func (w *ForestWrapper) ESChat(modelID uint, desc bool) (*llmchat.QaRes, error) {
	logs.DebugContextf(w.ctx, "用户问题:\n%v\n", w.question.Source.Question)
	logs.DebugContextf(w.ctx, "对话历史\n%v\n", w.history)
	agentName := global.ChatAgentQuestionAnswer
	if desc {
		agentName = global.ChatAgentDescriptionQuestionAnswer
	}

	req := &chattype.ChatRequestBody{
		Stream:     true,
		Model:      chatagent.GetAgentI18nName(w.ctx, runtime.GetLanguage(w.ctx), agentName),
		LLMModelID: modelID,
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{Name: "input1", Value: w.question.Source.Question}, // 用户问题
				{Name: "input2", Value: w.searchStr},                // 知识库检索内容
				{Name: "input3", Value: w.history},                  // 对话历史
			},
		},
	}
	wrapper, err := chatclient.NewInternalChat(w.ctx, w.question.Source.ReqID, "", 2, req)
	if err != nil {
		logs.ErrorContextf(w.ctx, "failed to create internal chat: %v", err)
		return nil, err
	}
	res, err := wrapper.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(w.ctx, "agent chat error: %v", err)
		return nil, err
	}
	return res, nil
}

// ESChat 根据es查到的chunk来进行问答
func (w *ForestWrapper) PromptChat(modelID uint, desc bool, promptKey string) (*llmchat.QaRes, error) {
	if desc {
		return w.ESChat(modelID, desc)
	}

	// 拼装提示词question，发送模型
	if modelID == 0 {
		modelID = 1
	}
	model, err := chatmodel.GetModelByID(w.ctx, modelID)
	if err != nil {
		logs.ErrorContextf(w.ctx, "AgentChat GetModelByID failed ,err %s", err)
		return nil, err
	}
	req := &chattype.ChatRequestBody{
		Stream: true,
		// Model:      chatagent.GetAgentI18nName(w.ctx, runtime.GetLanguage(w.ctx), agentName),
		LLMModelID: modelID,
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{Name: "input1", Value: w.question.Source.Question}, // 用户问题
				{Name: "input2", Value: w.searchStr},                // 知识库检索内容
				{Name: "input3", Value: w.history},                  // 对话历史
			},
		},
	}
	ques := &chattype.ChatQuestion{
		Source: &chattype.Question{
			CompanyID:    w.question.Source.CompanyID,
			Uin:          w.question.Source.Uin,
			AgentStep:    2,
			ReqID:        w.question.Source.ReqID,
			Status:       chattype.QuestionStatusPending,
			ApiKeyID:     0,
			BaseAgentID:  0,
			AgentVersion: 0,
			ModelID:      model.ID,
			UserInput:    req,
			AgentName:    "prompt_chat",
		},
	}
	defer chatquestion.CreateQuestion(w.ctx, ques)
	roleName, err := settings.GetText(global.SettingGroupCoreKG, global.SettingKeyLlmRoleName)
	if roleName == "" || err != nil {
		roleName = global.DefaultLlmRoleName
	}
	prompts := GetKeQAPrompts(roleName)
	messages := []*llmchat.Message{}
	temp, ok := prompts[promptKey]
	if !ok {
		temp = prompts["normal"]
	}
	updatedTemplate := replaceInputPlaceholders(temp, &req.ChatOptions.Input)
	messages = append(messages, &llmchat.Message{
		Role:    llmchat.MessageRoleUser,
		Content: updatedTemplate,
	})
	request := &llmchat.ChatReqBody{
		Messages: messages,
		Stream:   ques.Source.UserInput.Stream,
		// Temperature: &agentInfo.Temperature,
	}
	if ques.Source.UserInput.Stream {
		request.StreamOptions = llmchat.NewStreamOptions()
	}
	wrapper := llmchat.NewLLmChatWrapper(w.ctx, request, model)
	// 直接返回模型原始结果到调用方
	res, err := wrapper.InternalChatResponse(nil)
	if err != nil {
		logs.ErrorContextf(w.ctx, "PromptChat InternalChatResponse failed ,err %s", err)
		// return
		ques.Source.Status = chattype.QuestionStatusError
	}
	if res != nil {
		ques.Source.Answer = res.Content
		ques.Source.Reasoning = res.Reasoning
		ques.Source.ReasoningSeconds = res.ReasoningTime
		ques.Source.CostSeconds = res.CostSeconds
		ques.Source.OutToken = res.Usage.CompletionTokens
		ques.Source.CacheHitToken = res.Usage.PromptCacheHitTokens
		ques.Source.CacheMissToken = res.Usage.PromptCacheMissTokens
		ques.Source.TotalTokens = res.Usage.TotalTokens
		ques.Source.Status = chattype.QuestionStatusAnswered
	}
	return res, nil
}

// WriteReferenceFile 写入检索到的文件
func WriteReferenceFile(ctx *gin.Context, ref chattype.QueryReferenceList, reqID string) {
	sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()),
		sseclient.WithExpiration(time.Minute*5))
	sseClient.SetHeaders(ctx.Writer)
	for _, r := range ref {
		if stoped, err := sseClient.WriteMessage(ctx, ctx.Writer, reqID, llmchat.WriteResult{
			Flag:      llmchat.FlagFound,
			Reference: r,
		}.String()); err != nil {
			defer sseClient.Close(ctx, reqID)
			logs.ErrorContextf(ctx, "[forestqa] Failed to write Answering response to KEQA: %v", err)
			continue
		} else if stoped {
			defer sseClient.Close(ctx, reqID)
			logs.ErrorContextf(ctx, "[forestqa] stream Stoped by KEQA")
			return
		}
	}
}

// WriteReferenceFile 写入检索到的文件
func WriteSearchQA(ctx *gin.Context, answer, reqID string) {
	sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()),
		sseclient.WithExpiration(time.Minute*5))
	defer sseClient.Close(ctx, reqID)
	sseClient.SetHeaders(ctx.Writer)
	if stoped, err := sseClient.WriteMessage(ctx, ctx.Writer, reqID, llmchat.WriteResult{
		Content: answer,
	}.String()); err != nil {
		logs.ErrorContextf(ctx, "[WriteSearchQA] Failed to write Thinking response to KEQA: %v", err)
	} else if stoped {
		logs.ErrorContextf(ctx, "[WriteSearchQA] stream Stoped by KEQA")
		return
	}
}

// WriteReferenceFile 写入检索到的文件
func WriteFlag(ctx *gin.Context, reqID string, flag llmchat.FlagAnswer) {
	sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()),
		sseclient.WithExpiration(time.Minute*5))
	sseClient.SetHeaders(ctx.Writer)
	if stoped, err := sseClient.WriteMessage(ctx, ctx.Writer, reqID, llmchat.WriteResult{
		Flag: flag,
	}.String()); err != nil {
		defer sseClient.Close(ctx, reqID)
		logs.ErrorContextf(ctx, "[forestqa] Failed to write Answering response to KEQA: %v", err)
		return
	} else if stoped {
		defer sseClient.Close(ctx, reqID)
		logs.ErrorContextf(ctx, "[forestqa] stream Stoped by KEQA")
		return
	}
}

// ReplaceInputPlaceholders 替换模板中的占位符 inputN 为对应的 input.Value
func replaceInputPlaceholders(template string, inputs *chattype.InputList) string {
	// 将 inputs 转换为 map，方便快速查找
	inputMap := make(map[string]string)
	for _, input := range *inputs {
		inputMap[input.Name] = input.Value
	}

	// 正则表达式匹配所有 {{inputN}}
	re := regexp.MustCompile(`\{\{input\d+\}\}`)

	// 替换所有匹配的占位符
	template = re.ReplaceAllStringFunc(template, func(match string) string {
		// 提取占位符名称（如 input1）
		name := match[2 : len(match)-2]
		if value, ok := inputMap[name]; ok {
			return value
		}
		// 如果未找到对应的 input.Value，保留占位符
		return match
	})

	return template
}
