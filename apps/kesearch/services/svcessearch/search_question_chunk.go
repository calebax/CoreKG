package svcessearch

import (
	"context"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/keqa"
	"github.com/insmtx/corekg/apps/kesearch/models/reranksearch"
	"github.com/ygpkg/yg-go/logs"
)

func RerankSearchQuestionChunk(ctx context.Context,
	esIndex string,
	question string,
	forestIDs []uint,
	fileIDs []uint,
	searchConf *reranksearch.SearchConfig,
	originalQuestion string) (chattype.QueryReferenceList, error) {
	searchOpts := &reranksearch.RerankSearchOptions{
		OriginalQuestion: originalQuestion,
	}
	w, err := reranksearch.NewRerankSearchWrapper(ctx, esIndex, question, forestIDs, fileIDs, searchConf, searchOpts)
	if err != nil {
		logs.ErrorContextf(ctx, "[RerankSearchQuestionChunk] NewRerankSearchWrapper error: %v", err)
		return nil, err
	}
	res, err := w.RerankSearchChunk()
	if err != nil {
		logs.ErrorContextf(ctx, "[RerankSearchQuestionChunk] RerankSearchChunk error: %v", err)
		return nil, err
	}

	res, err = keqa.RewriteChatReferenceImageURLs(ctx, res)
	if err != nil {
		logs.ErrorContextf(ctx, "[RerankSearchQuestionChunk] RewriteChatReferenceImageURLs error: %v", err)
		return nil, err
	}

	return res, nil
}

func SearchDescription(ctx *gin.Context,
	esIndex string,
	question string,
	forestIDs []uint,
	fileIDs []uint) (string, chattype.QueryReferenceList, error) {
	searchReferenceWrapper, err := keqa.HandelSearchReference(ctx, forestIDs, fileIDs, esIndex, question)
	if err != nil {
		logs.ErrorContextf(ctx, "SearchDescription HandelSearchReference error: %v", err)
		return "", nil, err
	}

	preSearchResult, _, err := searchReferenceWrapper.SearchFileDescriptions(question)
	if err != nil {
		logs.ErrorContextf(ctx, "SearchDescription SearchFileDescriptions error: %v", err)
		return "", nil, err
	}
	refList := searchReferenceWrapper.SupQuestionChunk(preSearchResult)

	var searchResultStr string
	if len(refList) > 150 {
		searchResultStr, err = batchSumDesc(ctx, question, refList)
		if err != nil {
			logs.ErrorContextf(ctx, "SearchDescription batchSumDesc error: %v", err)
			return "", nil, err
		}
	} else {
		searchStr, err := keqa.TransformChatReferenceList(refList)
		if err != nil {
			logs.ErrorContextf(ctx, "SearchDescription TransformChatReferenceList error: %v", err)
			return "", nil, err
		}
		searchResultStr = string(searchStr)
	}

	return searchResultStr, refList, nil
}

func batchSumDesc(ctx *gin.Context, question string, refList chattype.QueryReferenceList) (string, error) {
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
			searchStr, err := keqa.TransformChatReferenceList(chunk)
			if err != nil {
				resultChan <- batchResult{index: idx, answer: "", err: err}
				return
			}
			question := &chattype.ChatQuestion{
				Source: &chattype.Question{
					Question: question,
				},
			}
			answer, err := keqa.SumDescAgent(ctx, question, string(searchStr))
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
	return strings.Join(batchAnswers, "\n"), nil

}
