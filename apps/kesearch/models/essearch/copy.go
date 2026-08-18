package essearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/logs"
)

func escapeNebulaString(s string) string {
	return strings.NewReplacer(
		// `%`, `%%`,
		`\`, `\\`,
	).Replace(s)
}

func CopyForestData(ctx context.Context, srcID uint, forest_info *foresttype.KnownowForest, fileidmap map[uint]uint) error {
	es_index := forest_info.EsIndex()
	esdate, err := SearchForestData(ctx, srcID, es_index)
	if err != nil {
		logs.ErrorContextf(ctx, "CopyForestData SearchForestData err: %v", err)
		return err
	}
	esdate = ChangeSrcData(esdate, forest_info, fileidmap)
	client, err := InitESClient(ctx)
	if err != nil {
		return err
	}

	batchSize := 1000
	totalDocs := len(esdate.Hits.Hits)

	// 分批处理
	for i := 0; i < totalDocs; i += batchSize {
		end := i + batchSize
		if end > totalDocs {
			end = totalDocs
		}

		batch := esdate.Hits.Hits[i:end]

		var buf bytes.Buffer
		for _, doc := range batch {
			meta := []byte(fmt.Sprintf(`{ "index" : { "_index" : "%s", "_id" : "%s" } }%s`, es_index, escapeNebulaString(doc.ID), "\n"))
			data, err := json.Marshal(doc.Source)
			if err != nil {
				logs.ErrorContextf(ctx, "json.Marshal error: %v", err)
				return err
			}
			data = append(data, byte('\n'))
			buf.Grow(len(meta) + len(data))
			buf.Write(meta)
			buf.Write(data)
		}
		resp, err := client.Bulk(
			bytes.NewReader(buf.Bytes()),
			client.Bulk.WithContext(ctx),
			client.Bulk.WithRefresh("true"),
		)
		if err != nil {
			logs.ErrorContextf(ctx, "client.Bulk error: %v", err)
			return err
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return fmt.Errorf("es Bulk failed: %s, error: %s", resp.Status(), string(body))
		}

		resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		println("", string(body))

		logs.InfoContextf(ctx, "Batch %d-%d inserted successfully", i+1, end)
	}

	return nil
}

func ChangeSrcData(data *SearchResult, forest_info *foresttype.KnownowForest, fileidmap map[uint]uint) *SearchResult {
	for i, doc := range data.Hits.Hits {
		doc.ID = fmt.Sprintf("%v_%s", forest_info.ID, doc.ID)
		doc.Source.ForestID = forest_info.ID
		doc.Source.CompanyID = forest_info.CompanyID
		doc.Source.Uin = forest_info.Uin
		doc.Source.FileID = fileidmap[doc.Source.FileID]
		if doc.Source.QAMainID != "" {
			doc.Source.QAMainID = fmt.Sprintf("%v_%s", forest_info.ID, doc.Source.QAMainID)
		}
		if doc.Source.QAAnswerID != "" {
			doc.Source.QAAnswerID = fmt.Sprintf("%v_%s", forest_info.ID, doc.Source.QAAnswerID)
		}
		for j, title := range doc.Source.TitleLevelIDs {
			title = fmt.Sprintf("%v_%s", forest_info.ID, title)
			doc.Source.TitleLevelIDs[j] = title
		}
		for j, reference := range doc.Source.References {
			reference.ChunkID = fmt.Sprintf("%v_%s", forest_info.ID, reference.ChunkID)
			reference.RelationshipID = fmt.Sprintf("%v_%s", forest_info.ID, reference.RelationshipID)
			reference.FileID = fileidmap[reference.FileID]
			doc.Source.References[j] = reference
		}
		now := time.Now()
		doc.Source.CreatedAt = now
		doc.Source.UpdatedAt = now
		data.Hits.Hits[i] = doc
	}
	return data
}

// SearchForestData 搜索知识库全部chunk
func SearchForestData(ctx context.Context, forestID uint, es_index string) (*SearchResult, error) {
	client, err := InitESClient(ctx)
	if err != nil {
		return nil, err
	}
	// 初始查询
	var fullResult SearchResult
	batchSize := 1000
	sid := ""

	for {
		var resp *esapi.Response
		if sid == "" {
			var buf bytes.Buffer
			query := map[string]interface{}{
				"size": batchSize,
				"query": map[string]interface{}{
					"term": map[string]interface{}{
						"forest_id": forestID,
					},
				},
			}
			if err := json.NewEncoder(&buf).Encode(query); err != nil {
				logs.ErrorContextf(ctx, "error encoding query: %s")
				return nil, err
			}
			resp, err = client.Search(
				client.Search.WithContext(ctx),
				client.Search.WithIndex(es_index),
				client.Search.WithBody(&buf),
				client.Search.WithScroll(time.Minute*10),
			)
			if err != nil {
				logs.ErrorContextf(ctx, "Error getting response: %s", err)
				return nil, err
			}
		} else {
			resp, err = client.Scroll(
				client.Scroll.WithScrollID(sid),
				client.Scroll.WithScroll(time.Minute*10))
			if err != nil {
				logs.ErrorContextf(ctx, "Error getting response: %s", err)
				return nil, err
			}
		}
		defer resp.Body.Close()
		// 读取完整响应体
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logs.ErrorContextf(ctx, "error reading body: %v", err)
			return nil, err
		}
		// 解析JSON响应
		var searchRes SearchResult
		if err := json.Unmarshal(body, &searchRes); err != nil {
			logs.ErrorContextf(ctx, "unmarshal ChatResponse error: %v", err)
			return nil, err
		}
		sid = searchRes.ScrollID
		fullResult.Hits.Hits = append(fullResult.Hits.Hits, searchRes.Hits.Hits...)
		if len(searchRes.Hits.Hits) < batchSize {
			break
		}
	}

	return &fullResult, nil
}
