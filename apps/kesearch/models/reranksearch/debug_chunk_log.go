package reranksearch

import (
	"sort"
	"strings"

	"github.com/ygpkg/yg-go/logs"
)

const chunkRankLogTag = "[DEBUG][chunk-rank]"

func (w *RerankSearchWrapper) logChunkOrder(stage string, chunks []*SearchType, sortByRerank bool, limit int) {
	ordered := cloneChunkSlice(chunks)
	if sortByRerank {
		sort.SliceStable(ordered, func(i, j int) bool {
			return ordered[i].RerankScore > ordered[j].RerankScore
		})
	}
	if limit <= 0 || limit > len(ordered) {
		limit = len(ordered)
	}
	logs.InfoContextf(w.ctx, "%s stage=%s count=%d sort_by_rerank=%v limit=%d",
		chunkRankLogTag, stage, len(chunks), sortByRerank, limit)
	for i := 0; i < limit; i++ {
		c := ordered[i]
		logs.InfoContextf(w.ctx,
			"%s stage=%s rank=%d file_id=%d sequence=%d chunk_id=%s type=%s es_score=%.6f rerank_score=%.6f desc_preview=%q",
			chunkRankLogTag, stage, i+1, c.FileID, c.Sequence, c.ChunkID, c.Type, c.Score, c.RerankScore, chunkDescPreview(c.Description),
		)
	}
}

func (w *RerankSearchWrapper) logChunkSelection(stage string, chunks []*SearchType, selected []*SearchType, sortByRerank bool, limit int, keptReason, droppedReason string) {
	selectedMap := map[*SearchType]bool{}
	for _, c := range selected {
		selectedMap[c] = true
	}

	ordered := cloneChunkSlice(chunks)
	if sortByRerank {
		sort.SliceStable(ordered, func(i, j int) bool {
			return ordered[i].RerankScore > ordered[j].RerankScore
		})
	}
	if limit <= 0 || limit > len(ordered) {
		limit = len(ordered)
	}
	logs.InfoContextf(w.ctx, "%s stage=%s count=%d selected=%d sort_by_rerank=%v limit=%d",
		chunkRankLogTag, stage, len(chunks), len(selected), sortByRerank, limit)
	for i := 0; i < limit; i++ {
		c := ordered[i]
		outcome := droppedReason
		if selectedMap[c] {
			outcome = keptReason
		}
		logs.InfoContextf(w.ctx,
			"%s stage=%s rank=%d outcome=%s file_id=%d sequence=%d chunk_id=%s type=%s es_score=%.6f rerank_score=%.6f desc_preview=%q",
			chunkRankLogTag, stage, i+1, outcome, c.FileID, c.Sequence, c.ChunkID, c.Type, c.Score, c.RerankScore, chunkDescPreview(c.Description),
		)
	}
}

func cloneChunkSlice(chunks []*SearchType) []*SearchType {
	ordered := make([]*SearchType, len(chunks))
	copy(ordered, chunks)
	return ordered
}

func chunkDescPreview(desc string) string {
	desc = strings.Join(strings.Fields(desc), " ")
	const maxLen = 180
	if len(desc) <= maxLen {
		return desc
	}
	runes := []rune(desc)
	if len(runes) <= maxLen {
		return desc
	}
	return string(runes[:maxLen])
}
