package reranksearch

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/ygpkg/yg-go/logs"
)

type RerankSearchWrapper struct {
	ctx context.Context

	cli       *elasticsearch.Client
	indexName string
	// rerankQuestion keeps the user's original intent for relevance scoring.
	rerankQuestion string
	userQuery      string
	forestIds      []uint
	fileIds        []uint
	embedding      ragtypes.Embedding
	conf           *SearchConfig
}

type RerankSearchOptions struct {
	OriginalQuestion string
}

func resolveRerankQuestion(question string, opts *RerankSearchOptions) string {
	if opts != nil && opts.OriginalQuestion != "" {
		return opts.OriginalQuestion
	}
	return question
}

// NewRerankSearchWrapper 创建一个包装器
func NewRerankSearchWrapper(ctx context.Context, indexName, question string, forestIds []uint, fileIds []uint, conf *SearchConfig, opts *RerankSearchOptions) (*RerankSearchWrapper, error) {
	rerankQuestion := resolveRerankQuestion(question, opts)
	logs.InfoContextf(ctx, "[DEBUG][chunk-empty] NewRerankSearchWrapper start: index=%s, question=%s, rerank_question=%s, forest_ids=%v, file_ids=%v",
		indexName, question, rerankQuestion, forestIds, fileIds)

	escli, err := essearch.InitESClient(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "[DEBUG][chunk-empty] NewRerankSearchWrapper InitESClient error: %v", err)
		return nil, err
	}
	logs.InfoContextf(ctx, "[DEBUG][chunk-empty] NewRerankSearchWrapper InitESClient success")

	quest, err := UserQueryRewrite(ctx, question)
	if err != nil {
		logs.ErrorContextf(ctx, "[DEBUG][chunk-empty] NewRerankSearchWrapper UserQueryRewrite error: %v", err)
		return nil, err
	}
	logs.InfoContextf(ctx, "[DEBUG][chunk-empty] NewRerankSearchWrapper UserQueryRewrite: original=%s, rewritten=%s", question, quest)

	eb, err := essearch.GetEmbedding(quest)
	if err != nil {
		logs.ErrorContextf(ctx, "[DEBUG][chunk-empty] NewRerankSearchWrapper GetEmbedding error: %v", err)
		return nil, err
	}
	logs.InfoContextf(ctx, "[DEBUG][chunk-empty] NewRerankSearchWrapper GetEmbedding success, embedding_len=%d", len(eb))

	if conf == nil {
		conf = GetDefaultConfig()
	}
	logs.InfoContextf(ctx, "[DEBUG][chunk-empty] NewRerankSearchWrapper config: topn=%d, topm=%d, topk=%d, fetch_factor=%d, rerank_threshold=%.2f, enable_rerank=%v, fallback_to_topk=%v, neighbor_size=%d, embed_weight=%.2f, desc_weight=%.2f",
		conf.Topn, conf.Topm, conf.Topk, conf.FetchFactor, conf.RerankThreshold, conf.EnableRerank, conf.FallBackToTopK, conf.NeighborSize, conf.EmbeddingWeight, conf.DescriptionWeight)

	return &RerankSearchWrapper{
		ctx:            ctx,
		cli:            escli,
		indexName:      indexName,
		rerankQuestion: rerankQuestion,
		userQuery:      quest,
		forestIds:      forestIds,
		fileIds:        fileIds,
		embedding:      eb,
		conf:           conf,
	}, nil
}
