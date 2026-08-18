package globalsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/logs"
)

// SearchImageType 知识库图片高亮
type SearchImageType struct {
	ID                     uint    `json:"id"`
	ForestID               uint    `json:"forest_id"`
	FileName               string  `json:"file_name"`
	Score                  float64 `json:"_score,omitempty"`
	HighlightedFileName    string  `json:"highlighted_file_name,omitempty"`
	Description            string  `json:"description"`
	HighlightedDescription string  `json:"highlighted_description,omitempty"`
	ImageURL               string  `json:"image_url,omitempty"`
	Location               [5]int  `json:"location,omitempty"` // 位置
}

// FindFileImage 搜索图片类型
func (wrapper GlobalSearchWrapper) FindFileImage(fileType string) (*essearch.SearchResult, error) {
	client, err := essearch.InitESClient(wrapper.Ctx)
	if err != nil {
		return nil, err
	}
	// 找不同的chunk类型
	var typeList []string
	switch fileType {
	case "image":
		typeList = []string{"image"}
	case "video":
		typeList = []string{"video"}
	}
	mustMap := []esquery.Map{
		esquery.BuildMap("terms", esquery.BuildMap("type", typeList)),
		esquery.BuildMap("term", esquery.BuildMap("company_id", wrapper.CompanyID))}
	if len(wrapper.ForestIDs) > 0 {
		mustMap = append(mustMap, esquery.BuildMap("terms", esquery.BuildMap("forest_id", wrapper.ForestIDs)))
	}
	boolQuery := esquery.BuildMap("must", mustMap)

	boolQuery["should"] = []esquery.Map{
		esquery.BuildMap("match",
			esquery.BuildMap("description", wrapper.Text)),
	}
	// 使用语义
	if wrapper.IsSemantics {
		boolQuery["should"] = append(boolQuery["should"].([]esquery.Map),
			esquery.BuildMap("script_score",
				esquery.BuildMap(
					"query", esquery.BuildMap("match_all", esquery.Map{}),
					"script",
					esquery.BuildMap(
						"source", "double score = cosineSimilarity(params.query_vector, 'embedding') + 1.0; return score > 1.0 ? score : 0;",
						"params", esquery.BuildMap("query_vector", wrapper.embedding)),
				)))
	}

	highLight := esquery.BuildMap("fields", esquery.BuildMap("description",
		esquery.BuildMap("pre_tags", []string{highlightConfig.HighLightPrefix},
			"post_tags", []string{highlightConfig.HighLightSuffix})))

	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", boolQuery)).
		SetHighlight(highLight).
		SetSize(wrapper.SubjectCount)
	querybyte, err := query.BuildBytes()
	if err != nil {
		return nil, err
	}
	logs.InfoContextf(wrapper.Ctx, "FindFileChunk querybyte: %v", string(querybyte))
	resp, err := client.Search(
		client.Search.WithIndex(wrapper.EsIndex),
		client.Search.WithBody(bytes.NewBuffer(querybyte)),
		client.Search.WithContext(wrapper.Ctx),
	)
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "FindFileChunkAggs error: %v", err)
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
		logs.ErrorContextf(wrapper.Ctx, "error reading body: %v", err)
		return nil, err
	}
	// 解析JSON响应
	var searchRes essearch.SearchResult
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(wrapper.Ctx, "unmarshal ChatResponse error: %v", err)
		return nil, err
	}
	return &searchRes, nil
}
