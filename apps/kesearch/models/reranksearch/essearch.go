package reranksearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/logs"
)

// SearchQuestionChunk 根据问题搜索对应的chunk，返回检索结果
func (w *RerankSearchWrapper) SearchQuestionChunk() ([]*SearchType, error) {
	mustQuery := []esquery.Map{
		esquery.BuildMap("bool", esquery.BuildMap("should", []esquery.Map{
			esquery.BuildMap("term", esquery.BuildMap("is_disable", false)),
			esquery.BuildMap("bool", esquery.BuildMap("must_not", esquery.BuildMap("exists", esquery.BuildMap("field", "is_disable")))),
		})),
		esquery.BuildMap("exists", esquery.BuildMap("field", "embedding")),
		esquery.BuildMap("terms", esquery.BuildMap("type", []string{"chunk", "image", "table", "video", "formula"})),
	}
	if len(w.forestIds) != 0 {
		mustQuery = append(mustQuery, esquery.BuildMap("terms", esquery.BuildMap("forest_id", w.forestIds)))
	}
	if len(w.fileIds) != 0 {
		mustQuery = append(mustQuery, esquery.BuildMap("terms", esquery.BuildMap("file_id", w.fileIds)))
	}

	boolQuery := esquery.BuildMap("filter", mustQuery)

	boolQuery["should"] = []esquery.Map{
		esquery.BuildMap("multi_match",
			esquery.BuildMap(
				"query", w.userQuery,
				"fields", []string{fmt.Sprintf("description^%.1f", w.conf.DescriptionWeight)},
				"type", "best_fields",
			)),
	}
	source := `double textScore = _score;
	 double vectorScore = 0.0;
	 if (doc['embedding'].size() > 0) {
        vectorScore = (cosineSimilarity(params.query_vector, 'embedding') + 1.0) / 2.0;
     }
	 return textScore + vectorScore * params.embedding_weight;`
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("script_score", esquery.BuildMap("query",
			esquery.BuildMap("bool", boolQuery),
			"script",
			esquery.BuildMap(
				"source", source,
				"params", esquery.BuildMap(
					"query_vector", w.embedding,
					"embedding_weight", w.conf.EmbeddingWeight,
				)),
		),
		)).
		SetSize(w.conf.Topn*w.conf.FetchFactor).
		Set("_source", esquery.BuildMap("excludes", []string{"embedding", "references"}))
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(w.ctx, "query build error: %v", err)
		return nil, err
	}
	logs.InfoContextf(w.ctx, "querybyte: %v", string(querybyte))
	resp, err := w.cli.Search(
		w.cli.Search.WithIndex(w.indexName),
		w.cli.Search.WithBody(bytes.NewBuffer(querybyte)),
		w.cli.Search.WithContext(w.ctx),
	)
	if err != nil {
		logs.ErrorContextf(w.ctx, "es query failed: %v", err)
		return nil, err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
	// 转换结果
	// 读取完整响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(w.ctx, "error reading body: %v", err)
		return nil, err
	}
	// 解析JSON响应
	var searchRes essearch.SearchResult
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(w.ctx, "unmarshal ChatResponse error: %v", err)
		return nil, err
	}
	logs.InfoContextf(w.ctx, "[DEBUG][chunk-empty] SearchQuestionChunk ES result: total_hits=%d, returned_hits=%d", searchRes.Hits.Total.Value, len(searchRes.Hits.Hits))
	return getSearchType(w.ctx, searchRes), nil
}

func (w *RerankSearchWrapper) SortRerankChunk(chunks []*SearchType) ([]*SearchType, error) {
	rchunks := []*SearchType{}
	if len(chunks) == 0 {
		logs.InfoContextf(w.ctx, "[DEBUG][chunk-empty] SortRerankChunk: input_chunks is empty, returning empty")
		return rchunks, nil
	}
	if w.conf.EnableRerank {
		str := []string{}
		for _, v := range chunks {
			str = append(str, v.Description)
		}
		rerank, err := GetRerank(w.ctx, w.rerankQuestion, str)
		if err != nil {
			logs.ErrorContextf(w.ctx, "GetRerank error: %v", err)
			return nil, err
		}
		logs.InfoContextf(w.ctx, "[DEBUG][chunk-empty] SortRerankChunk rerank result: data_count=%d, input_count=%d", len(rerank.Data), len(str))
		for _, v := range rerank.Data {
			if v.Index >= 0 && v.Index < len(chunks) {
				chunks[v.Index].RerankScore = v.Score
			}
		}
		w.logChunkOrder("rerank_scored_order", chunks, true, 0)
		for _, c := range chunks {
			if c.RerankScore >= w.conf.RerankThreshold {
				rchunks = append(rchunks, c)
			}
		}
		w.logChunkSelection("rerank_threshold_selection", chunks, rchunks, true, 0, "kept_by_threshold", "dropped_below_threshold")
		logs.InfoContextf(w.ctx, "[DEBUG][chunk-empty] SortRerankChunk after threshold filter(%.2f): before=%d, after=%d",
			w.conf.RerankThreshold, len(chunks), len(rchunks))
		sort.Slice(rchunks, func(i, j int) bool {
			return rchunks[i].RerankScore > rchunks[j].RerankScore
		})

		thresholdChunks := rchunks
		if len(rchunks) > w.conf.Topn {
			rchunks = rchunks[:w.conf.Topn]
		}
		w.logChunkSelection("rerank_topn_selection", thresholdChunks, rchunks, true, 0, "kept_by_topn", "dropped_by_topn")
		// 兜底
		if len(rchunks) == 0 {
			keywordChunks := w.selectKeywordFallbackChunks(chunks)
			if len(keywordChunks) > 0 {
				rchunks = keywordChunks
				logs.WarnContextf(w.ctx, "[DEBUG][chunk-empty] SortRerankChunk: all chunks filtered by threshold, keyword fallback selected=%d", len(rchunks))
				w.logChunkSelection("rerank_keyword_fallback_selection", chunks, rchunks, false, 0, "kept_by_keyword_fallback", "dropped_by_keyword_fallback")
			}
		}
		if len(rchunks) == 0 && w.conf.FallBackToTopK {
			logs.WarnContextf(w.ctx, "[DEBUG][chunk-empty] SortRerankChunk: all chunks filtered by threshold, fallback to topk=%d", w.conf.Topk)
			sort.Slice(chunks, func(i, j int) bool {
				return chunks[i].RerankScore > chunks[j].RerankScore
			})
			if len(chunks) > w.conf.Topk {
				rchunks = chunks[:w.conf.Topk]
			} else {
				rchunks = chunks
			}
			w.logChunkSelection("rerank_fallback_topk_selection", chunks, rchunks, true, 0, "kept_by_fallback_topk", "dropped_by_fallback_topk")
		} else if len(rchunks) == 0 && !w.conf.FallBackToTopK {
			logs.WarnContextf(w.ctx, "[DEBUG][chunk-empty] SortRerankChunk: all chunks filtered and fallback disabled, returning empty")
		}
	} else {
		if len(chunks) > w.conf.Topn {
			rchunks = chunks[:w.conf.Topn]
		} else {
			rchunks = chunks
		}
	}
	return rchunks, nil
}

// SearchChunkSequence 搜索chunk上下文
func (w *RerankSearchWrapper) SearchChunkSequence(chunks []*SearchType) ([]*SearchType, error) {
	if len(chunks) == 0 {
		logs.InfoContextf(w.ctx, "[DEBUG][chunk-empty] SearchChunkSequence: input_chunks is empty, returning empty")
		return chunks, nil
	}
	if w.conf.NeighborSize <= 0 {
		return chunks, nil
	}
	var searchCount = len(chunks) * (2*w.conf.NeighborSize + 1)

	shoudMap := []esquery.Map{}
	for _, hit := range chunks {
		sequence := buildChunkWindowSequences(hit.Sequence, w.conf.NeighborSize)
		shoud_item := esquery.BuildMap("bool",
			esquery.BuildMap("must", []esquery.Map{
				esquery.BuildMap("bool", esquery.BuildMap("should", []esquery.Map{
					esquery.BuildMap("term", esquery.BuildMap("is_disable", false)),
					esquery.BuildMap("bool", esquery.BuildMap("must_not", esquery.BuildMap("exists", esquery.BuildMap("field", "is_disable")))),
				})),
				esquery.BuildMap("term", esquery.BuildMap("file_id", hit.FileID)),
				esquery.BuildMap("terms", esquery.BuildMap("sequence", sequence))}))
		shoudMap = append(shoudMap, shoud_item)
	}
	mustQuery := []esquery.Map{
		esquery.BuildMap("terms", esquery.BuildMap("type", []string{"chunk", "image", "table", "video"})),
		esquery.BuildMap("bool",
			esquery.BuildMap(
				"should", shoudMap,
				"minimum_should_match", 1)),
	}
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool",
			esquery.BuildMap("must", mustQuery))).
		SetSize(searchCount).
		Set("_source", esquery.BuildMap("excludes", []string{"embedding"}))
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(w.ctx, "SearchChunkSequence esquery.BuildMap error: %v", err)
		return nil, err
	}
	logs.InfoContextf(w.ctx, "querybyte: %v", string(querybyte))
	resp, err := w.cli.Search(
		w.cli.Search.WithIndex(w.indexName),
		w.cli.Search.WithBody(bytes.NewBuffer(querybyte)),
		w.cli.Search.WithContext(w.ctx),
	)
	if err != nil {
		logs.ErrorContextf(w.ctx, "SearchChunkSequence client.Search error: %v", err)
		return nil, err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(w.ctx, "SearchChunkSequence resp.StatusCode error: %s, body: %s", resp.Status(), string(body))
		return nil, fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
	// 转换结果
	// 读取完整响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(w.ctx, "error reading body: %v", err)
		return nil, err
	}
	// 解析JSON响应
	var searchRes essearch.SearchResult
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(w.ctx, "unmarshal ChatResponse error: %v", err)
		return nil, err
	}

	return getSearchType(w.ctx, searchRes), nil
}

func buildChunkWindowSequences(sequence, neighborSize int) []int {
	sequences := []int{sequence}
	for i := 1; i <= neighborSize; i++ {
		sequences = append(sequences, sequence+i)
		sequences = append(sequences, sequence-i)
	}
	return sequences
}

// JoinNeighborChunks 拼接邻居 chunk
func (w *RerankSearchWrapper) JoinNeighborChunks(chunks []*SearchType, nbchunks []*SearchType) []*SearchType {
	nbchunksMap := map[string]*SearchType{}
	for _, v := range nbchunks {
		key := fmt.Sprintf("%v:%v", v.FileID, v.Sequence)
		nbchunksMap[key] = v
	}
	expandedChunks := []*SearchType{}
	for _, c := range chunks {
		descs := []string{}
		// 当前 chunk 优先，避免 rerank 被上文无关内容带偏。
		descs = append(descs, c.Description)
		// 上文
		for offset := w.conf.NeighborSize; offset > 0; offset-- {
			key := fmt.Sprintf("%v:%v", c.FileID, c.Sequence-offset)
			if left, ok := nbchunksMap[key]; ok {
				descs = append(descs, left.Description)
			}
		}
		// 下文
		for offset := 1; offset <= w.conf.NeighborSize; offset++ {
			key := fmt.Sprintf("%v:%v", c.FileID, c.Sequence+offset)
			if right, ok := nbchunksMap[key]; ok {
				descs = append(descs, right.Description)
			}
		}
		c.Description = strings.Join(descs, " ")
		expandedChunks = append(expandedChunks, c)
	}
	return expandedChunks
}

// GroupByFileID 根据fileid聚合
func (w *RerankSearchWrapper) GroupByFileID(chunks []*SearchType) map[uint][]*SearchType {
	fileChunks := map[uint][]*SearchType{}
	for _, v := range chunks {
		fileChunks[v.FileID] = append(fileChunks[v.FileID], v)
	}
	return fileChunks
}

// SearchFilesAbstract 搜索摘要
func (w *RerankSearchWrapper) SearchFilesAbstract(filechunks map[uint][]*SearchType) (map[uint]string, error) {
	keys := make([]uint, 0, len(filechunks))
	for k := range filechunks {
		keys = append(keys, k)
	}
	query := map[string]interface{}{
		"size": 1,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{
					{"term": map[string]interface{}{"type": ragtypes.ChunkTypeFileDescription}},
					{"terms": map[string]interface{}{"file_id": keys}},
				},
			},
		},
	}
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}
	resp, err := w.cli.Search(
		w.cli.Search.WithContext(w.ctx),
		w.cli.Search.WithIndex(w.indexName),
		w.cli.Search.WithBody(bytes.NewReader(queryJSON)),
	)
	if err != nil {
		logs.ErrorContextf(w.ctx, "es query failed: %v", err)
		return nil, err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
	// 转换结果
	// 读取完整响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(w.ctx, "error reading body: %v", err)
		return nil, err
	}
	// 解析JSON响应
	var searchRes essearch.FileDescResult
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(w.ctx, "unmarshal esresponse error: %v", err)
		return nil, err
	}
	abstMap := map[uint]string{}
	for _, v := range searchRes.Hits.Hits {
		abstMap[v.Source.FileID] = v.Source.Abstract
	}
	return abstMap, err
}

// RerankAbstract 对摘要进行rerank
func (w *RerankSearchWrapper) RerankAbstract(absts map[uint]string) (map[uint]string, error) {
	// 是否启用 rerank
	if w.conf.EnableRerank && len(absts) > 0 {
		// 提取摘要文本
		absTexts := make([]string, 0, len(absts))
		fidList := make([]uint, 0, len(absts))
		for fid, abs := range absts {
			fidList = append(fidList, fid)
			absTexts = append(absTexts, abs)
		}

		// 调用 rerank
		rerank, err := GetRerank(w.ctx, w.rerankQuestion, absTexts)
		if err != nil {
			logs.ErrorContextf(w.ctx, "GetRerank error: %v", err)
			return nil, err
		}

		for i, fid := range fidList {
			if i < len(rerank.Data) {
				if score := rerank.Data[i]; score.Score >= w.conf.RerankThreshold {
					absts[fid] = absTexts[i]
				}
			}
		}
	}
	return absts, nil
}

func (w *RerankSearchWrapper) Resault(filechunks map[uint][]*SearchType, abstract map[uint]string) chattype.QueryReferenceList {
	chunk := chattype.QueryReferenceList{}
	fileMap := map[uint]*foresttype.KnownowForestFile{}
	userMap := map[uint]*accounttype.User{}
	for fileID, v := range filechunks {
		abst, ok := abstract[fileID]
		if !ok {
			abst = ""
		}
		file, ok := fileMap[fileID]
		if !ok {
			f, err := forest.GetForestFileByID(fileID)
			if err != nil {
				logs.ErrorContextf(w.ctx, "get file err:%v", err)
				continue
			}
			file = f
			fileMap[fileID] = f
		}
		userEntity, exists := userMap[file.Uin]
		if !exists {
			u, err := user.GetUserByUin(w.ctx, file.Uin)
			if err != nil {
				logs.ErrorContextf(w.ctx, "GetUserByUin error: %v", err)
				continue
			}
			userEntity = u
			userMap[file.Uin] = userEntity
		}
		filechunk := &chattype.QueryReference{
			FileID:         fileID,
			Abstract:       abst,
			DataSourceType: "DC",
			FileName:       file.Name,
			Uin:            file.Uin,
			CreatedAt:      file.CreatedAt,
			UserName:       userEntity.Name,
			AvatarURL:      userEntity.AvatarURL,
		}
		for _, c := range v {
			// filechunk.FileName = c.FileName
			filechunk.ForestID = c.ForestID
			ck := chattype.QueryReferenceChunk{
				ChunkID:  c.ChunkID,
				Sequence: c.Sequence,
				Content:  c.Description,
				ImageURL: c.ImageURL,
				Score:    c.RerankScore,
				Location: c.Location,
				Type:     c.Type,
			}
			filechunk.ChunkList = append(filechunk.ChunkList, ck)
		}
		chunk = append(chunk, filechunk)
	}
	return chunk
}
