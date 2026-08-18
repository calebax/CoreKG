package svcreranksearch

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/insmtx/corekg/apps/kesearch/internal/dto/dtoreranksearch"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/apps/kesearch/models/reranksearch"
	"github.com/insmtx/corekg/apps/kesearch/services/svcessearch"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// RerankSearchChunk rerank搜索接口
func RerankSearchChunk(ctx *gin.Context, req *dtoreranksearch.RerankSearchChunkRequest) (res *dtoreranksearch.RerankSearchChunkResponse, err error) {
	res = &dtoreranksearch.RerankSearchChunkResponse{}
	// 后期如果多索引修改ke_0
	searchQaResult, err := svcessearch.FindFQAByQuestion(ctx, "ke_0", req.Request.Question, req.Request.ForestIDs, req.Request.FileIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "RerankSearchChunk FindFQAByQuestion error: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kechat_internal_error" // 内部错误
		return
	}
	if searchResult, ok := buildFQAResult(ctx, searchQaResult); ok {
		res.Response.SearchResult = searchResult
		return res, nil
	}
	wrapper, err := reranksearch.NewRerankSearchWrapper(ctx, "ke_0", req.Request.Question,
		req.Request.ForestIDs, req.Request.FileIDs, req.Request.Config, nil)
	if err != nil {
		logs.ErrorContextf(ctx, "RerankSearchChunk NewRerankSearchWrapper error: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kechat_internal_error" // 内部错误
		return
	}
	searchRes, err := wrapper.RerankSearchChunk()
	if err != nil {
		logs.ErrorContextf(ctx, "RerankSearchChunk RerankSearchChunk error: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kesearch_search_failed" // 搜索失败
		return
	}
	res.Response.SearchResult = searchRes
	return res, nil
}

func buildFQAResult(ctx *gin.Context, searchQaResult *essearch.SearchResult) (chattype.QueryReferenceList, bool) {
	if searchQaResult == nil || len(searchQaResult.Hits.Hits) == 0 {
		return nil, false
	}
	logs.InfoContextf(ctx, "RerankSearchChunk FindFQAByQuestion result: %v", len(searchQaResult.Hits.Hits))
	qaHit := searchQaResult.Hits.Hits[0]
	qaSource := qaHit.Source
	qaChunkID := qaHit.ID
	if qaSource.QAAnswerID != "" {
		qaChunkID = qaSource.QAAnswerID
	}
	result := chattype.QueryReferenceList{
		{
			FileID:         qaSource.FileID,
			FileName:       qaSource.FileName,
			ForestID:       qaSource.ForestID,
			Uin:            qaSource.Uin,
			CreatedAt:      qaSource.CreatedAt,
			DataSourceType: chattype.DataSourceTypeDC,
			ChunkList: chattype.QueryReferenceChunkList{
				{
					Type:     ragtypes.ChunkTypeFQA,
					ChunkID:  qaChunkID,
					Sequence: qaSource.Sequence,
					Content:  qaSource.QAAnswer,
					Score:    qaHit.Score,
					Location: qaSource.Location,
				},
			},
		},
	}
	return result, true
}
