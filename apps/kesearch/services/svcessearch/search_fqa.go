package svcessearch

import (
	"context"

	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/ygpkg/yg-go/logs"
)

// FindFQAByQuestion 搜索问答对
func FindFQAByQuestion(ctx context.Context,
	esIndex string,
	question string,
	forestIDs []uint,
	fileIDs []uint) (*essearch.SearchResult, error) {
	searchWrapper, err := essearch.NewEsSearchWrapper(ctx, esIndex, question, forestIDs, fileIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "NewEsSearchWrapper error: %v", err)
		return nil, err
	}
	return searchWrapper.FindFQAByQuestion()
}
