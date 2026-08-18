package keqa

import (
	"context"
	"strings"
	"sync"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/agentclient"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/logs"
)

// DescriptionChat 通过摘要回答总结性问题
func DescriptionChat(wrapper *searchReferenceWrapper, sseClient *sseclient.SSEClient, qs *foresttype.KnownowForestQA, history string, session *foresttype.KnownowQASession) (string, string, error) {
	preSearchResult, files, err := wrapper.SearchFileDescriptions(qs.Question + qs.ImageContent)
	if err != nil {
		return "", "", err
	}
	WriteReferenceFile(wrapper.ctx, sseClient, files, qs.ID)
	refList := wrapper.SupQuestionChunk(preSearchResult)
	qs.QueryReferenceList = refList
	queryStr := ""
	if len(refList) > 150 {
		queryStr = BatchSumDesc(wrapper.ctx, refList, qs.Question)
	} else {
		searchStr, err := TransformChatReferenceList(wrapper.ctx, refList)
		if err != nil {
			return "", "", err
		}
		queryStr = string(searchStr)
	}

	answer, reason, err := ESChat(wrapper.ctx, sseClient, queryStr, history, qs, session)
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "ESChat error: %v", err)
	}
	return answer, reason, nil
}

// BatchSumDesc 分批跑
func BatchSumDesc(ctx context.Context, refList foresttype.ChatReferenceList, question string) string {
	batchSize := 150
	batchCount := (len(refList) + batchSize - 1) / batchSize

	type batchResult struct {
		index  int
		answer string
		err    error
	}

	var wg sync.WaitGroup
	resultChan := make(chan batchResult, batchCount)
	cfg, err := agentclient.GetLLMConfig(ctx, global.SettingGroupKnowledge, global.SettingKeyAgentEsChat)
	if err != nil {
		logs.ErrorContextf(ctx, "get llm config failed: %v", err)
		return ""
	}
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
		go func(idx int, chunk foresttype.ChatReferenceList) {
			logs.InfoContextf(ctx, "Batch %d: %v", idx+1, len(chunk))
			defer wg.Done()
			searchStr, err := TransformChatReferenceList(ctx, chunk)
			if err != nil {
				resultChan <- batchResult{index: idx, answer: "", err: err}
				return
			}
			answer, err := SumDescAgent(ctx, cfg, string(searchStr), "", question)
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
	finalInput := strings.Join(batchAnswers, "\n")

	return finalInput
}
