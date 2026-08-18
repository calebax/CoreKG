package reranksearch

import (
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/logs"
)

// RerankSearchChunk rerank版本搜索chunk
func (w *RerankSearchWrapper) RerankSearchChunk() (chattype.QueryReferenceList, error) {
	res, err := w.SearchQuestionChunk()
	if err != nil {
		logs.ErrorContextf(w.ctx, "RerankSearchChunk SearchQuestionChunk error: %v", err)
		return nil, err
	}
	logs.InfoContextf(w.ctx, "[DEBUG][chunk-empty] RerankSearchChunk step1_ES_search: result_count=%d", len(res))
	w.logChunkOrder("step1_ES_search_order", res, false, 0)
	seqChunk, err := w.SearchChunkSequence(res)
	if err != nil {
		logs.ErrorContextf(w.ctx, "RerankSearchChunk SearchChunkSequence error: %v", err)
		return nil, err
	}
	logs.InfoContextf(w.ctx, "[DEBUG][chunk-empty] RerankSearchChunk step2_neighbor_search: result_count=%d", len(seqChunk))
	w.logChunkOrder("step2_neighbor_search_order", seqChunk, false, 0)
	nbres := w.JoinNeighborChunks(res, seqChunk)
	logs.InfoContextf(w.ctx, "[DEBUG][chunk-empty] RerankSearchChunk step3_join_neighbor: result_count=%d", len(nbres))
	w.logChunkOrder("step3_join_neighbor_order", nbres, false, 0)
	nbsres, err := w.SortRerankChunk(nbres)
	if err != nil {
		logs.ErrorContextf(w.ctx, "RerankSearchChunk SortRerankChunk error: %v", err)
		return nil, err
	}
	logs.InfoContextf(w.ctx, "[DEBUG][chunk-empty] RerankSearchChunk step4_rerank: result_count=%d", len(nbsres))
	beforeTopm := nbsres
	// 最大返回个数
	if w.conf.Topm < len(nbsres) {
		nbsres = nbsres[:w.conf.Topm]
	}
	w.logChunkSelection("step5_topm_selection", beforeTopm, nbsres, true, 0, "kept_by_topm", "dropped_by_topm")
	logs.InfoContextf(w.ctx, "[DEBUG][chunk-empty] RerankSearchChunk step5_topm_filter: result_count=%d, topm=%d", len(nbsres), w.conf.Topm)
	fileChunkMap := w.GroupByFileID(nbsres)
	w.logChunkOrder("step6_group_by_file_input", nbsres, true, 0)
	logs.InfoContextf(w.ctx, "[DEBUG][chunk-empty] RerankSearchChunk step6_group_by_file: file_count=%d", len(fileChunkMap))
	abstRerankres := map[uint]string{}
	if w.conf.EnabelAbstract {
		abstres, err := w.SearchFilesAbstract(fileChunkMap)
		if err != nil {
			logs.ErrorContextf(w.ctx, "RerankSearchChunk SearchFilesAbstract error: %v", err)
			return nil, err
		}
		abstRerankres, err = w.RerankAbstract(abstres)
		if err != nil {
			logs.ErrorContextf(w.ctx, "RerankSearchChunk RerankAbstract error: %v", err)
			return nil, err
		}
	}
	finalResult := w.Resault(fileChunkMap, abstRerankres)
	logs.InfoContextf(w.ctx, "[DEBUG][chunk-empty] RerankSearchChunk step7_final_result: chunk_count=%d", len(finalResult))
	return finalResult, nil
}
