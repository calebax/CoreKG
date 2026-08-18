package reranksearch

import (
	"math"
	"sort"
	"strings"

	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/ygpkg/yg-go/logs"
)

const (
	minKeywordFallbackTokenRunes = 2
	minKeywordFallbackScore      = 0.45
	maxKeywordFallbackTokens     = 12
	keywordFallbackWeight        = 0.7
	esFallbackWeight             = 0.3
)

type keywordFallbackCandidate struct {
	chunk        *SearchType
	hitKeywords  []string
	rareKeywords []string
	finalScore   float64
	keywordScore float64
	esScore      float64
}

type keywordFallbackChunk struct {
	chunk   *SearchType
	content string
}

func (w *RerankSearchWrapper) selectKeywordFallbackChunks(chunks []*SearchType) []*SearchType {
	keywords, err := w.extractFallbackKeywords()
	if err != nil {
		logs.WarnContextf(w.ctx, "[DEBUG][chunk-empty] SortRerankChunk keyword fallback analyze error: %v", err)
		return nil
	}
	if len(keywords) == 0 {
		logs.InfoContextf(w.ctx, "[DEBUG][chunk-empty] SortRerankChunk keyword fallback: no keywords extracted")
		return nil
	}
	logs.InfoContextf(w.ctx, "[DEBUG][chunk-empty] SortRerankChunk keyword fallback keywords=%v", keywords)

	selected, candidates := selectKeywordFallbackChunksByKeywords(chunks, keywords, w.conf.Topk)
	for _, candidate := range candidates {
		outcome := "dropped_by_keyword_fallback"
		for _, selectedChunk := range selected {
			if selectedChunk == candidate.chunk {
				outcome = "kept_by_keyword_fallback"
				break
			}
		}
		logs.InfoContextf(w.ctx,
			"[DEBUG][chunk-empty] SortRerankChunk keyword fallback candidate: outcome=%s file_id=%d sequence=%d chunk_id=%s final_score=%.6f keyword_score=%.6f es_score=%.6f rerank_score=%.6f hit_keywords=%v rare_keywords=%v",
			outcome, candidate.chunk.FileID, candidate.chunk.Sequence, candidate.chunk.ChunkID, candidate.finalScore, candidate.keywordScore, candidate.esScore, candidate.chunk.RerankScore, candidate.hitKeywords, candidate.rareKeywords)
	}
	return selected
}

func (w *RerankSearchWrapper) extractFallbackKeywords() ([]string, error) {
	keywords, err := essearch.Analyze(w.ctx, w.rerankQuestion)
	if err != nil {
		return nil, err
	}
	tokens := make([]string, 0, len(keywords.Tokens))
	for _, token := range keywords.Tokens {
		tokens = append(tokens, token.Token)
	}
	return filterKeywordFallbackTokens(tokens), nil
}

func selectKeywordFallbackChunksByKeywords(chunks []*SearchType, keywords []string, topk int) ([]*SearchType, []keywordFallbackCandidate) {
	keywords = filterKeywordFallbackTokens(keywords)
	if len(chunks) == 0 || len(keywords) == 0 {
		return nil, nil
	}

	fallbackChunks := buildKeywordFallbackChunks(chunks)
	stats := buildKeywordFallbackStats(fallbackChunks, keywords)
	keywords = stats.activeKeywords
	if len(keywords) == 0 || len(stats.rareKeywords) == 0 {
		return nil, nil
	}

	maxESScore := maxKeywordFallbackESScore(fallbackChunks)

	candidates := make([]keywordFallbackCandidate, 0, len(chunks))
	for _, fallbackChunk := range fallbackChunks {
		candidate := buildKeywordFallbackCandidate(fallbackChunk, keywords, stats, maxESScore)
		if !candidate.hasRareKeywordHit() {
			continue
		}
		if candidate.keywordScore < minKeywordFallbackScore {
			continue
		}
		candidates = append(candidates, candidate)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].finalScore != candidates[j].finalScore {
			return candidates[i].finalScore > candidates[j].finalScore
		}
		if candidates[i].keywordScore != candidates[j].keywordScore {
			return candidates[i].keywordScore > candidates[j].keywordScore
		}
		if len(candidates[i].hitKeywords) != len(candidates[j].hitKeywords) {
			return len(candidates[i].hitKeywords) > len(candidates[j].hitKeywords)
		}
		return candidates[i].chunk.RerankScore > candidates[j].chunk.RerankScore
	})

	if topk <= 0 || topk > len(candidates) {
		topk = len(candidates)
	}
	selected := make([]*SearchType, 0, topk)
	for i := 0; i < topk; i++ {
		selected = append(selected, candidates[i].chunk)
	}
	return selected, candidates
}

func buildKeywordFallbackChunks(chunks []*SearchType) []keywordFallbackChunk {
	fallbackChunks := make([]keywordFallbackChunk, 0, len(chunks))
	for _, chunk := range chunks {
		fallbackChunks = append(fallbackChunks, keywordFallbackChunk{
			chunk:   chunk,
			content: normalizeKeywordFallbackText(keywordFallbackContent(chunk)),
		})
	}
	return fallbackChunks
}

func maxKeywordFallbackESScore(chunks []keywordFallbackChunk) float64 {
	maxESScore := 0.0
	for _, chunk := range chunks {
		if chunk.chunk.Score > maxESScore {
			maxESScore = chunk.chunk.Score
		}
	}
	return maxESScore
}

func buildKeywordFallbackCandidate(chunk keywordFallbackChunk, keywords []string, stats keywordFallbackStats, maxESScore float64) keywordFallbackCandidate {
	totalWeight := 0.0
	hitWeight := 0.0
	hitKeywords := []string{}
	rareKeywords := []string{}
	for _, keyword := range keywords {
		weight := keywordFallbackWeightFor(stats, keyword)
		totalWeight += weight
		if strings.Contains(chunk.content, keyword) {
			hitWeight += weight
			hitKeywords = append(hitKeywords, keyword)
			if stats.rareKeywordSet[keyword] {
				rareKeywords = append(rareKeywords, keyword)
			}
		}
	}

	keywordScore := 0.0
	if totalWeight > 0 {
		keywordScore = float64(hitWeight) / float64(totalWeight)
	}
	esScore := 0.0
	if maxESScore > 0 {
		esScore = chunk.chunk.Score / maxESScore
	}
	return keywordFallbackCandidate{
		chunk:        chunk.chunk,
		hitKeywords:  hitKeywords,
		rareKeywords: rareKeywords,
		keywordScore: keywordScore,
		esScore:      esScore,
		finalScore:   keywordFallbackWeight*keywordScore + esFallbackWeight*esScore,
	}
}

func (c keywordFallbackCandidate) hasRareKeywordHit() bool {
	return len(c.rareKeywords) > 0
}

type keywordFallbackStats struct {
	activeKeywords []string
	docFreq        map[string]int
	idf            map[string]float64
	rareKeywords   []string
	rareKeywordSet map[string]bool
}

func buildKeywordFallbackStats(chunks []keywordFallbackChunk, keywords []string) keywordFallbackStats {
	stats := keywordFallbackStats{
		docFreq:        map[string]int{},
		idf:            map[string]float64{},
		rareKeywordSet: map[string]bool{},
	}
	for _, keyword := range keywords {
		for _, chunk := range chunks {
			if strings.Contains(chunk.content, keyword) {
				stats.docFreq[keyword]++
			}
		}
		if stats.docFreq[keyword] > 0 {
			stats.activeKeywords = append(stats.activeKeywords, keyword)
		}
	}
	if len(stats.activeKeywords) == 0 {
		return stats
	}

	minDocFreq := len(chunks) + 1
	for _, keyword := range stats.activeKeywords {
		df := stats.docFreq[keyword]
		stats.idf[keyword] = math.Log(float64(len(chunks)+1)/float64(df+1)) + 1
		if df < minDocFreq {
			minDocFreq = df
		}
	}
	for _, keyword := range stats.activeKeywords {
		if stats.docFreq[keyword] == minDocFreq {
			stats.rareKeywords = append(stats.rareKeywords, keyword)
			stats.rareKeywordSet[keyword] = true
		}
	}
	return stats
}

func keywordFallbackWeightFor(stats keywordFallbackStats, keyword string) float64 {
	return float64(len([]rune(keyword))) * stats.idf[keyword] * stats.idf[keyword]
}

func keywordFallbackContent(chunk *SearchType) string {
	if chunk.rawContent != "" {
		return chunk.rawContent
	}
	return chunk.Description
}

func filterKeywordFallbackTokens(tokens []string) []string {
	seen := map[string]bool{}
	keywords := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = normalizeKeywordFallbackText(token)
		if len([]rune(token)) < minKeywordFallbackTokenRunes || seen[token] {
			continue
		}
		seen[token] = true
		keywords = append(keywords, token)
	}

	sort.SliceStable(keywords, func(i, j int) bool {
		return len([]rune(keywords[i])) > len([]rune(keywords[j]))
	})
	filtered := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		contained := false
		for _, selected := range filtered {
			if strings.Contains(selected, keyword) {
				contained = true
				break
			}
		}
		if contained {
			continue
		}
		filtered = append(filtered, keyword)
		if len(filtered) >= maxKeywordFallbackTokens {
			break
		}
	}
	return filtered
}

func normalizeKeywordFallbackText(s string) string {
	replacer := strings.NewReplacer(
		" ", "",
		"\n", "",
		"\r", "",
		"\t", "",
		"?", "",
		"？", "",
		"。", "",
		"，", "",
		",", "",
		"、", "",
		"：", "",
		":", "",
		"；", "",
		";", "",
		"！", "",
		"!", "",
		"“", "",
		"”", "",
		"\"", "",
		"'", "",
		"‘", "",
		"’", "",
	)
	return replacer.Replace(strings.TrimSpace(s))
}
