package essearch

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/logs"
)

func NewEsSearchWrapper(ctx context.Context, indexName string, question string, forestIds []uint, fileIds []uint) (*EsSearchWrapper, error) {
	eb, err := GetEmbedding(question)
	if err != nil {
		logs.ErrorContextf(ctx, "GetEmbedding error: %v", err)
		return nil, err
	}
	escli, err := InitESClient(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "NewWrapper InitESClient error: %v", err)
		return nil, err
	}
	return &EsSearchWrapper{
		ctx:       ctx,
		cli:       escli,
		indexName: indexName,
		question:  question,
		forestIds: forestIds,
		fileIds:   fileIds,
		embedding: eb,
	}, nil
}

type EsSearchWrapper struct {
	ctx context.Context

	cli       *elasticsearch.Client
	indexName string
	question  string
	forestIds []uint
	fileIds   []uint
	embedding ragtypes.Embedding

	pageQuery   *apiobj.PageQuery
	query       esquery.Map
	questionIDs []string
	Total       uint
}

func NewWrapper(ctx context.Context, indexName, question string, questionIDs []string, forestIds []uint, fileIds []uint, pq *apiobj.PageQuery) (*EsSearchWrapper, error) {
	escli, err := InitESClient(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "NewWrapper InitESClient error: %v", err)
		return nil, err
	}

	var eb ragtypes.Embedding
	if len(question) > 0 {
		eb, err = GetEmbedding(question)
		if err != nil {
			logs.ErrorContextf(ctx, "GetEmbedding error: %v", err)
			return nil, err
		}
	}
	query := esquery.Map{}

	if pq == nil {
		pq = &apiobj.PageQuery{
			Offset:  0,
			Limit:   0,
			ListAll: true,
		}
	}

	return &EsSearchWrapper{
		ctx:       ctx,
		cli:       escli,
		indexName: indexName,
		embedding: eb,
		forestIds: forestIds,
		fileIds:   fileIds,

		query:       query,
		pageQuery:   pq,
		questionIDs: questionIDs,
	}, nil
}

// NewPureWrapper will return a pure wrapper for common search
// * it only just does some sample action
func NewPureWrapper(ctx context.Context, indexName string, forestIds, fileIds []uint, escli *elasticsearch.Client) *EsSearchWrapper {
	var err error
	if escli == nil {
		escli, err = InitESClient(ctx)
		if err != nil {
			logs.ErrorContextf(ctx, "NewWrapper InitESClient error: %v", err)
			return nil
		}
	}

	if len(indexName) == 0 {
		logs.ErrorContextf(ctx, "NewWrapper indexName is empty")
		return nil
	}

	return &EsSearchWrapper{
		ctx:       ctx,
		cli:       escli,
		indexName: indexName,
		forestIds: forestIds,
		fileIds:   fileIds,
	}
}
