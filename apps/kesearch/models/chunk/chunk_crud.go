package chunk

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/logs"
)

var escli *elasticsearch.Client

// ListChunksByFileID 获取文件chunk
func ListChunksByFileID(ctx context.Context, fileID uint) ([]*Chunk, error) {
	mustQuery := []esquery.Map{
		esquery.BuildMap("term", esquery.BuildMap("file_id", fileID)),
		// esquery.BuildMap("term", esquery.BuildMap("uin", uin)),
		esquery.BuildMap("terms", esquery.BuildMap("type", []string{"chunk", "image", "table", "video", "formula"})),
	}
	sort := []esquery.Map{
		esquery.BuildMap("sequence", esquery.BuildMap("order", "asc")),
	}
	boolQuery := esquery.BuildMap("must", mustQuery)
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", boolQuery)).
		SetSort(sort).
		SetSize(10000).
		SetSource([]string{"type", "tokens", "file_id", "sequence", "is_disable",
			"location", "description", "image_url", "formula",
			"title_level_ids", "title_level", "forest_id", "company_id", "uin", "chunk_size", "yg_location", "table"})
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(ctx, "query build error: %v", err)
		return nil, err
	}
	logs.InfoContextf(ctx, "querybyte: %v", string(querybyte))
	resp, err := escli.Search(
		escli.Search.WithIndex(DefaultIndex),
		escli.Search.WithBody(bytes.NewBuffer(querybyte)),
		escli.Search.WithContext(ctx),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es query failed: %v", err)
		return nil, err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "es query failed: %s, error: %s", resp.Status(), string(body))
		return nil, fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "error reading body: %v", err)
		return nil, err
	}
	// 解析JSON响应
	var searchRes ChunkSearchResult
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(ctx, "unmarshal ChatResponse error: %v", err)
		return nil, err
	}
	for i, v := range searchRes.Hits.Hits {
		if v.Source.YGLocation != "" {
			v.Source.Location = parseYgPosString(v.Source.YGLocation)
		}
		searchRes.Hits.Hits[i] = v
	}
	return searchRes.Hits.Hits, nil
}

// ListChunksByFileIDAndSequences 根据文件 ID 和 chunk 序号列表获取 chunks。
func ListChunksByFileIDAndSequences(ctx context.Context, index string, fileID uint, sequences []int) ([]*Chunk, error) {
	if index == "" {
		index = DefaultIndex
	}
	if len(sequences) == 0 {
		return []*Chunk{}, nil
	}
	mustQuery := []esquery.Map{
		esquery.BuildMap("term", esquery.BuildMap("file_id", fileID)),
		esquery.BuildMap("terms", esquery.BuildMap("sequence", sequences)),
		esquery.BuildMap("terms", esquery.BuildMap("type", []string{"chunk", "image", "table", "video", "formula"})),
	}
	sort := []esquery.Map{
		esquery.BuildMap("sequence", esquery.BuildMap("order", "asc")),
	}
	boolQuery := esquery.BuildMap("must", mustQuery)
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", boolQuery)).
		SetSort(sort).
		SetSize(len(sequences)).
		SetSource([]string{"type", "tokens", "file_id", "sequence", "is_disable",
			"location", "description", "image_url", "formula",
			"title_level_ids", "title_level", "forest_id", "company_id", "uin", "chunk_size", "yg_location", "table"})
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(ctx, "query build error: %v", err)
		return nil, err
	}
	logs.InfoContextf(ctx, "querybyte: %v", string(querybyte))
	resp, err := escli.Search(
		escli.Search.WithIndex(index),
		escli.Search.WithBody(bytes.NewBuffer(querybyte)),
		escli.Search.WithContext(ctx),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es query failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "es query failed: %s, error: %s", resp.Status(), string(body))
		return nil, fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "error reading body: %v", err)
		return nil, err
	}
	var searchRes ChunkSearchResult
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(ctx, "unmarshal ChatResponse error: %v", err)
		return nil, err
	}
	for i, v := range searchRes.Hits.Hits {
		if v.Source != nil && v.Source.YGLocation != "" {
			v.Source.Location = parseYgPosString(v.Source.YGLocation)
		}
		searchRes.Hits.Hits[i] = v
	}
	return searchRes.Hits.Hits, nil
}

// GetFileAbstractByFileID 根据文件 ID 获取文件摘要。
func GetFileAbstractByFileID(ctx context.Context, index string, fileID uint) (string, error) {
	if index == "" {
		index = DefaultIndex
	}
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", esquery.BuildMap("filter", []esquery.Map{
			esquery.BuildMap("term", esquery.BuildMap("type", ragtypes.ChunkTypeFileDescription)),
			esquery.BuildMap("term", esquery.BuildMap("file_id", fileID)),
		}))).
		SetSize(1).
		SetSource([]string{"file_id", "abstract"})
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(ctx, "query build error: %v", err)
		return "", err
	}
	resp, err := escli.Search(
		escli.Search.WithIndex(index),
		escli.Search.WithBody(bytes.NewBuffer(querybyte)),
		escli.Search.WithContext(ctx),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es query failed: %v", err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "es query failed: %s, error: %s", resp.Status(), string(body))
		return "", fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "error reading body: %v", err)
		return "", err
	}
	var searchRes struct {
		Hits struct {
			Hits []struct {
				Source struct {
					Abstract string `json:"abstract"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(ctx, "unmarshal file abstract error: %v", err)
		return "", err
	}
	if len(searchRes.Hits.Hits) == 0 {
		return "", nil
	}
	return searchRes.Hits.Hits[0].Source.Abstract, nil
}

func parseYgPosString(input string) [5]int {
	if input == "" {
		return [5]int{}
	}
	// 定义起始和结束标记
	prefix := "<!--yg_pos"
	suffix := "yg_pos-->"

	// 检查字符串是否以指定前缀开头
	if !strings.HasPrefix(input, prefix) {
		logs.Warnf("err str prefix :%s", input)
		return [5]int{}
	}

	// 检查字符串是否以指定后缀结尾
	if !strings.HasSuffix(input, suffix) {
		logs.Warnf("err str suffix :%s", input)
		return [5]int{}
	}

	// 提取标记之间的内容
	contentStartIndex := len(prefix)
	contentEndIndex := len(input) - len(suffix)
	if contentEndIndex <= contentStartIndex {
		// 标记之间没有内容或格式错误
		return [5]int{}
	}
	content := input[contentStartIndex:contentEndIndex]

	// 按逗号分割内容
	parts := strings.Split(content, ",")

	// 创建一个整数切片，容量为分割后的部分数量
	numbers := [5]int{}
	// 遍历每个分割后的字符串部分
	for i, part := range parts {
		// 去除可能存在的前后空白字符
		trimmedPart := strings.TrimSpace(part)
		// 将字符串转换为整数
		num, err := strconv.Atoi(trimmedPart)
		if err != nil {
			// 如果转换失败，返回错误
			logs.Warnf("can not change '%s' int, str:'%s' err: %w", trimmedPart, input, err)
			return [5]int{}
		}
		// 将转换后的整数添加到切片中
		if i > 4 {
			return numbers
		}
		numbers[i] = num
	}

	return numbers
}

// CreateQuestion 创建问题
func CreateChunk(ctx context.Context, chunk *Chunk) error {
	resp, err := escli.Index(
		DefaultIndex,
		strings.NewReader(chunk.Source.String()),
		escli.Index.WithContext(ctx),
		escli.Index.WithRefresh("true"),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateChunk init history es client failed: %s", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "CreateChunk es query failed: %s, error: %s", resp.Status(), string(body))
		return fmt.Errorf("CreateChunk es query failed: %s, error: %s", resp.Status(), string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateChunk error reading body: %v", err)
		return err
	}
	// 解析JSON响应
	var searchRes createRespnse
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(ctx, "unmarshal createRespnse error: %v", err)
		return err
	}
	chunk.ID = searchRes.ID
	return nil
}

// GetQuetionByID 根据id获取问题
func GetChunkByID(ctx context.Context, id string) (*Chunk, error) {
	resp, err := escli.Get(
		DefaultIndex,
		id,
		escli.Get.WithContext(ctx),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es query failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "es query failed: %s, error: %s", resp.Status(), string(body))
		return nil, fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "error reading body: %v", err)
		return nil, err
	}
	// 解析JSON响应
	var question Chunk
	if err := json.Unmarshal(body, &question); err != nil {
		logs.ErrorContextf(ctx, "unmarshal ChatResponse error: %v", err)
		return nil, err
	}
	return &question, nil
}

// GetChunkBySequence 根据序列获取分片
func GetChunkBySequence(ctx context.Context, fileID uint, sequence int) (*Chunk, error) {
	mustQuery := []esquery.Map{
		esquery.BuildMap("term", esquery.BuildMap("file_id", fileID)),
		esquery.BuildMap("term", esquery.BuildMap("sequence", sequence)),
		esquery.BuildMap("terms", esquery.BuildMap("type", []string{"chunk", "image", "table", "video", "formula"})),
	}
	boolQuery := esquery.BuildMap("must", mustQuery)
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", boolQuery)).
		SetSize(1).
		SetSource([]string{"type", "tokens", "file_id", "sequence", "is_disable",
			"location", "description", "image_url", "formula",
			"title_level_ids", "title_level", "forest_id", "company_id", "uin", "chunk_size", "yg_location", "table"})
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(ctx, "query build error: %v", err)
		return nil, err
	}
	logs.InfoContextf(ctx, "querybyte: %v", string(querybyte))
	resp, err := escli.Search(
		escli.Search.WithIndex(DefaultIndex),
		escli.Search.WithBody(bytes.NewBuffer(querybyte)),
		escli.Search.WithContext(ctx),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es query failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "es query failed: %s, error: %s", resp.Status(), string(body))
		return nil, fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "error reading body: %v", err)
		return nil, err
	}
	// 解析JSON响应
	var searchRes ChunkSearchResult
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(ctx, "unmarshal ChatResponse error: %v", err)
		return nil, err
	}
	if len(searchRes.Hits.Hits) == 0 {
		return nil, fmt.Errorf("chunk not found by file_id: %d sequence: %d", fileID, sequence)
	}
	for i, v := range searchRes.Hits.Hits {
		if v.Source != nil && v.Source.YGLocation != "" {
			v.Source.Location = parseYgPosString(v.Source.YGLocation)
		}
		searchRes.Hits.Hits[i] = v
	}
	return searchRes.Hits.Hits[0], nil
}

// UpdateChunk 更新chunk
func UpdateChunk(ctx context.Context, chunk *Chunk) error {
	updateBody := map[string]interface{}{
		"doc": chunk.Source,
	}
	// 将数据转换为 JSON
	body, err := json.Marshal(updateBody)
	if err != nil {
		return fmt.Errorf("failed to marshal update body: %w", err)
	}
	resp, err := escli.Update(
		DefaultIndex,
		chunk.ID,
		strings.NewReader(string(body)),
		escli.Update.WithContext(ctx),
		escli.Update.WithRefresh("true"),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateQuestion init history es client failed: %s", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "UpdateQuestion es query failed: %s, error: %s", resp.Status(), string(body))
		return fmt.Errorf("UpdateQuestion es query failed: %s, error: %s", resp.Status(), string(body))
	}
	return nil
}

// DeleteChunk 删除chunk
func DeleteChunk(ctx context.Context, id string) error {
	resp, err := escli.Delete(
		DefaultIndex,
		id,
		escli.Delete.WithContext(ctx),
		escli.Delete.WithRefresh("true"),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteChunk init history es client failed: %s", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "DeleteChunk es query failed: %s, error: %s", resp.Status(), string(body))
		return fmt.Errorf("DeleteChunk es query failed: %s, error: %s", resp.Status(), string(body))
	}
	return nil
}

// MinusSequence 减少chunk 当前 sequence 后面自减1
func MinusSequence(ctx context.Context, fileID uint, sequence int) error {
	mustQuery := []esquery.Map{
		esquery.BuildMap("term", esquery.BuildMap("file_id", fileID)),
		esquery.BuildMap("terms", esquery.BuildMap("type",
			[]string{"chunk", "image", "table", "video", "formula"})),
		esquery.BuildMap("range", esquery.BuildMap("sequence", esquery.BuildMap("gt", sequence))),
	}
	boolQuery := esquery.BuildMap("must", mustQuery)
	query := esquery.NewBuilder().
		Set("script",
			esquery.BuildMap("source", "ctx._source.sequence -= 1", "lang", "painless")).
		SetQuery(esquery.BuildMap("bool", boolQuery))
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(ctx, "query build error: %v", err)
		return err
	}
	logs.InfoContextf(ctx, "querybyte: %v", string(querybyte))
	resp, err := escli.UpdateByQuery(
		[]string{DefaultIndex},
		escli.UpdateByQuery.WithBody(bytes.NewBuffer(querybyte)),
		escli.UpdateByQuery.WithContext(ctx),
		escli.UpdateByQuery.WithRefresh(true),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es query failed: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "es query failed: %s, error: %s", resp.Status(), string(body))
		return fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
	return nil
}

// MinusSequence 减少chunk 当前 sequence 后面自增1
func PlusSequence(ctx context.Context, fileID uint, sequence int) error {
	mustQuery := []esquery.Map{
		esquery.BuildMap("term", esquery.BuildMap("file_id", fileID)),
		esquery.BuildMap("terms", esquery.BuildMap("type",
			[]string{"chunk", "image", "table", "video", "formula"})),
		esquery.BuildMap("range", esquery.BuildMap("sequence", esquery.BuildMap("gt", sequence))),
	}
	boolQuery := esquery.BuildMap("must", mustQuery)
	query := esquery.NewBuilder().
		Set("script",
			esquery.BuildMap("source", "ctx._source.sequence += 1", "lang", "painless")).
		SetQuery(esquery.BuildMap("bool", boolQuery))
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(ctx, "query build error: %v", err)
		return err
	}
	logs.InfoContextf(ctx, "querybyte: %v", string(querybyte))
	resp, err := escli.UpdateByQuery(
		[]string{DefaultIndex},
		escli.UpdateByQuery.WithBody(bytes.NewBuffer(querybyte)),
		escli.UpdateByQuery.WithContext(ctx),
		escli.UpdateByQuery.WithRefresh(true),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es query failed: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "es query failed: %s, error: %s", resp.Status(), string(body))
		return fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
	return nil
}

// GetSHA256Hash 获取sha256 hash
func GetSHA256Hash(data string) string {
	hasher := sha256.New()
	_, err := hasher.Write([]byte(data))
	if err != nil {
		return ""
	}
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

// UpdateChunkFileName 更新chunk 文件名
func UpdateChunkFileName(ctx context.Context, fileID uint, fileName string) error {
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("term", esquery.BuildMap("file_id", fileID))).
		Set("script", esquery.BuildMap("source", "ctx._source.file_name = params.new_name",
			"lang", "painless",
			"params", esquery.BuildMap("new_name", fileName)))
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(ctx, "query build error: %v", err)
		return err
	}
	logs.InfoContextf(ctx, "UpdateChunkFileName querybyte: %v", string(querybyte))
	resp, err := escli.UpdateByQuery(
		[]string{"ke_0"},
		escli.UpdateByQuery.WithBody(bytes.NewBuffer(querybyte)),
		escli.UpdateByQuery.WithContext(ctx),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateChunkFileName error: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "UpdateChunkFileName error: %v", string(body))
		return err
	}
	return nil
}
