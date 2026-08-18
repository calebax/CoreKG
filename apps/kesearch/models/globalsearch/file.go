package globalsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/insmtx/corekg/version"
	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/logs"
)

// SearchFileType 知识库高亮
type SearchFileType struct {
	ID                  uint      `json:"id"`
	ForestID            uint      `json:"forest_id"`
	FileName            string    `json:"file_name"`
	UserName            string    `json:"user_name"`
	AvatarURL           string    `json:"avatar_url"`
	CreatedAt           time.Time `json:"created_at"`
	Score               float64   `json:"_score,omitempty"`
	HighlightedFileName string    `json:"highlighted_file_name,omitempty"`
	// 高亮字段
	Highlights []HighlightChunk `json:"highlights,omitempty"`
	IsLike     bool             `json:"is_like,omitempty"`
	IsCollect  bool             `json:"is_collect,omitempty"`
}
type HighlightChunk struct {
	// 高亮字段
	Score                  float64 `json:"_score,omitempty"`
	Description            string  `json:"description"`
	HighlightedDescription string  `json:"highlighted_description,omitempty"`
	ImageURL               string  `json:"image_url,omitempty"`
	Location               [5]int  `json:"location,omitempty"` // 位置
}

// SearchFileAggs 知识库高亮
func (wrapper GlobalSearchWrapper) SearchFileAggs(fileType string) ([]*SearchFileType, error) {
	result, err := wrapper.FindFileChunkAggs(fileType)
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "FindFileChunk error: %v", err)
		return nil, err
	}
	fileIDs := []uint{}
	for _, file := range result.Aggregations.ByFile.Buckets {
		fileIDs = append(fileIDs, file.Key)
	}

	files, err := forest.GetForestFileByIDs(fileIDs)
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "ListForestFile error: %v", err)
		return nil, err
	}

	likes, err := forest.NewUinLikesDao().GetListByCond(wrapper.Ctx, &forest.UinLikesCond{
		BaseCond: forest.BaseCond{
			Uin:       wrapper.Uin,
			CompanyID: wrapper.CompanyID,
		},
		ResourceIDs:  fileIDs,
		ResourceType: foresttype.ResourceTypeForestFile,
	})
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "NewUinLikesDao().GetByCond error: %v", err)
		return nil, err
	}
	collects, err := forest.NewUinCollectionsDao().GetListByCond(wrapper.Ctx, &forest.UinCollectionsCond{
		BaseCond: forest.BaseCond{
			Uin:       wrapper.Uin,
			CompanyID: wrapper.CompanyID,
		},
		ResourceIDs:  fileIDs,
		ResourceType: foresttype.ResourceTypeForestFile,
	})
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "NewUinCollectionsDao().GetByCond error: %v", err)
		return nil, err
	}

	likeMap := utils.ToMap(likes, func(like foresttype.UinLikes) uint {
		return like.ResourceID
	})
	collectMap := utils.ToMap(collects, func(collect foresttype.UinCollections) uint {
		return collect.ResourceID
	})

	filesMap := map[uint]*SearchFileType{}
	// userMap := map[uint]*accounttype.User{}
	for _, file := range files {
		userEntity, exists := wrapper.userMap[file.Uin]
		if !exists {
			userEntity, err = user.GetUserByUin(wrapper.Ctx, file.Uin)
			if err != nil {
				logs.ErrorContextf(wrapper.Ctx, "GetUserByUin error: %v", err)
				continue
			}
			wrapper.userMap[file.Uin] = userEntity
		}
		filesMap[file.ID] = &SearchFileType{
			ID:        file.ID,
			ForestID:  file.ForestID,
			FileName:  file.Name,
			CreatedAt: file.CreatedAt,
			UserName:  userEntity.Name,
			AvatarURL: userEntity.AvatarURL,
		}
	}
	for _, hit := range result.Aggregations.ByFile.Buckets {
		file, exists := filesMap[hit.Key]
		if !exists || file == nil {
			continue
		}
		file.Score = hit.TopDocsPerFile.Hits.MaxScore
		for _, hit := range hit.TopDocsPerFile.Hits.Hits {
			if len(hit.Highlight.Decription) == 0 {
				hit.Highlight.Decription = []string{hit.Source.Description}
			}
			if version.DeployMode() != "" {
				hit.Source.ImageURL = fs.SplitHost(wrapper.Ctx, hit.Source.ImageURL)
			}
			file.Highlights = append(file.Highlights, HighlightChunk{
				Description:            hit.Source.Description,
				ImageURL:               hit.Source.ImageURL,
				HighlightedDescription: hit.Highlight.Decription[0],
				Score:                  hit.Score,
				Location:               hit.Source.Location,
			})
		}
		_, file.IsLike = likeMap[file.ID]
		_, file.IsCollect = collectMap[file.ID]
	}
	// 返回文件
	value := make([]*SearchFileType, 0, len(filesMap))
	for _, v := range filesMap {
		value = append(value, v)
	}
	// 高亮文件名
	ProcessFileHighlight(value, wrapper.keywords)
	// 排序
	for _, v := range value {
		sort.Slice(v.Highlights, func(i, j int) bool {
			return v.Highlights[i].Score > v.Highlights[j].Score
		})
	}
	sort.Slice(value, func(i, j int) bool {
		return value[i].Score > value[j].Score
	})

	return value, nil
}

// FindFileChunkAggs 根据文件类型搜索
func (wrapper GlobalSearchWrapper) FindFileChunkAggs(fileType string) (*HighlightChunkAgg, error) {
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
	default:
		typeList = []string{"chunk", "table"}
	}

	ap, err := forest.NewAccessProvider(wrapper.Ctx, &forest.ContextModel{
		ResourceType: foresttype.ResourceTypeForestFile,
		ScopeType:    foresttype.ScopeTypeUser,
		ScopeID:      wrapper.Uin,
		Action:       foresttype.ActionBan,
	}).Action(wrapper.Ctx)
	if err != nil {
		logs.ErrorContextf(wrapper.Ctx, "filter ban list failed: %v", err)
		return nil, err
	}

	mustMap := []esquery.Map{
		esquery.BuildMap("exists", esquery.BuildMap("field", "embedding")),
		esquery.BuildMap("terms", esquery.BuildMap("type", typeList)),
		esquery.BuildMap("term", esquery.BuildMap("company_id", wrapper.CompanyID)),
		esquery.BuildMap("terms", esquery.BuildMap("forest_id", wrapper.ViewForestIDs)),
	}
	if len(wrapper.ForestIDs) > 0 {
		mustMap = append(mustMap, esquery.BuildMap("terms", esquery.BuildMap("forest_id", wrapper.ForestIDs)))
	}

	boolQuery := esquery.BuildMap("filter", mustMap)

	boolQuery["should"] = []esquery.Map{
		esquery.BuildMap("match",
			esquery.BuildMap("description", wrapper.Text)),
		esquery.BuildMap("match",
			esquery.BuildMap("file_name", esquery.BuildMap("query", wrapper.Text, "boost", 20))),
	}

	if len(ap.BanList) > 0 {
		boolQuery["must_not"] = []esquery.Map{
			esquery.BuildMap("terms", esquery.BuildMap(
				"file_id", ap.BanList,
			)),
		}
	}

	// 使用语义
	if wrapper.IsSemantics {
		// 如果是语义搜索，使用embedding
		// 获取embedding 结构
		// eb, err := essearch.GetEmbedding(wrapper.Text)
		// if err != nil {
		// 	logs.Errorf("GetEmbedding error: %v", err)
		// }
		boolQuery["should"] = append(boolQuery["should"].([]esquery.Map),
			esquery.BuildMap("script_score",
				esquery.BuildMap(
					"query", esquery.BuildMap("match_all", esquery.Map{}),
					"script",
					esquery.BuildMap(
						"source", "double score = cosineSimilarity(params.query_vector, 'embedding') + 1.0; return score > 1.7 ? score : 0;",
						"params", esquery.BuildMap("query_vector", wrapper.embedding)),
				)))
	}
	aggsMap := esquery.BuildMap("by_file",
		esquery.BuildMap("terms",
			esquery.BuildMap("field", "file_id", "size", wrapper.SubjectCount, "order", esquery.BuildMap("top_score", "desc")),
			// 文件数3
			"aggs", esquery.BuildMap("top_score", esquery.BuildMap("max", esquery.BuildMap("script", "_score")),
				"top_docs_per_file",
				esquery.BuildMap("top_hits",
					esquery.BuildMap("size", wrapper.ItemCount,
						"_source", esquery.BuildMap("includes", []string{"description", "file_id", "type", "company_id", "forest_id", "image_url", "file_name", "location"}),
						"sort", []esquery.Map{esquery.BuildMap("_score", "desc")},
						"highlight", esquery.BuildMap(
							"fields", esquery.BuildMap("description",
								esquery.BuildMap("pre_tags", []string{highlightConfig.HighLightPrefix},
									"post_tags", []string{highlightConfig.HighLightSuffix},
									"fragment_size", 1000,
								))))))))

	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", boolQuery)).
		SetSize(0).
		SetAggs(aggsMap).
		Set("min_score", 1.7)
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
	var searchRes HighlightChunkAgg
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(wrapper.Ctx, "unmarshal ChatResponse error: %v", err)
		return nil, err
	}
	return &searchRes, nil
}

// ProcessFileHighlight 处理整个 文件 列表，添加高亮字段
func ProcessFileHighlight(files []*SearchFileType, keywords *essearch.AnalyzeResultList) {
	for _, file := range files {
		file.HighlightedFileName = HighlightKeywords(file.FileName, keywords)
	}
}

type HighlightChunkAgg struct {
	Aggregations struct {
		ByFile struct {
			DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
			SumOtherDocCount        int `json:"sum_other_doc_count"`
			Buckets                 []struct {
				Key            uint `json:"key"`
				DocCount       int  `json:"doc_count"`
				TopDocsPerFile struct {
					Hits struct {
						Total struct {
							Value    int    `json:"value"`
							Relation string `json:"relation"`
						} `json:"total"`
						MaxScore float64 `json:"max_score"`
						Hits     []struct {
							Index  string  `json:"_index"`
							ID     string  `json:"_id"`
							Score  float64 `json:"_score"`
							Source struct {
								ForestID    uint   `json:"forest_id"`
								CompanyID   uint   `json:"company_id"`
								Type        string `json:"type"`
								FileID      uint   `json:"file_id"`
								FileName    string `json:"file_name"`
								Description string `json:"description"`
								ImageURL    string `json:"image_url"`
								Location    [5]int `json:"location"`
							} `json:"_source"`
							Highlight struct {
								Decription []string `json:"description"`
							} `json:"highlight"`
						} `json:"hits"`
					} `json:"hits"`
				} `json:"top_docs_per_file"`
			} `json:"buckets"`
		} `json:"by_file"`
	} `json:"aggregations"`
}
