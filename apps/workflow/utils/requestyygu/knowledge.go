package requestyygu

import (
	"context"
	"strings"
	"time"

	"github.com/ygpkg/yg-go/logs"
)

const (
	getForestPath         = "/v3/forest.GetForest"
	rerankSearchChunkPath = "/v3/kesearch.RerankSearchChunk"
	listForestPath        = "/v3/forest.ListForest"
	listFilePath          = "/v3/forest.ListFile"
)

type CoreKGKnowledgeSearchRequest struct {
	IDs          []int64
	MinScore     float64
	TopK         int64
	Query        string
	EnableRerank bool
}

func GetCoreKGKnowledgeSearchChunk(ctx context.Context, req *CoreKGKnowledgeSearchRequest) (*CoreKGKnowledgeSearchResponse, error) {
	if req == nil {
		req = &CoreKGKnowledgeSearchRequest{}
	}
	payload := map[string]interface{}{
		"config": map[string]interface{}{
			"description_weight": 1 - req.MinScore,
			"embedding_weight":   req.MinScore,
			"enabel_abstract":    true,
			"enable_rerank":      req.EnableRerank,
			"fall_back_to_topk":  true,
			"fetch_factor":       2,
			"neighbor_size":      1,
			"rerank_threshold":   0.5,
			"topk":               req.TopK,
			"topm":               req.TopK,
			"topn":               30,
		},
		"forest_ids": req.IDs,
		"question":   req.Query,
	}
	resp := &CoreKGKnowledgeSearchResponse{}
	if err := YyguRequest(ctx, rerankSearchChunkPath, payload, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func ListCoreKGForest(ctx context.Context, req *CoreKGListForestRequest) (*CoreKGListForestResponse, error) {
	if req == nil {
		req = &CoreKGListForestRequest{}
	}
	payload := map[string]interface{}{
		"offset":   req.Offset,
		"limit":    req.Limit,
		"list_all": req.ListAll,
	}
	if len(req.OrderBy) > 0 {
		payload["orderBy"] = req.OrderBy
	}
	filters := append([]CoreKGFilter{}, req.Filters...)
	if len(filters) == 0 {
		// 强制过滤 forest_type 为 file 和 qa
		filters = append(filters, CoreKGFilter{
			Field: "forest_type",
			Value: []string{"file", "qa"},
		})
	}
	if len(filters) > 0 {
		payload["filters"] = filters
	}
	resp := &CoreKGListForestResponse{}
	if err := YyguRequest(ctx, listForestPath, payload, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func ListCoreKGFile(ctx context.Context, req *CoreKGListFileRequest) (*CoreKGListFileResponse, error) {
	if req == nil {
		req = &CoreKGListFileRequest{}
	}
	parentID := req.ParentID
	if parentID == "" {
		parentID = "0"
	}
	limit := req.Limit
	if limit == 0 && !req.ListAll {
		limit = 10
	}
	payload := map[string]interface{}{
		"forest_id": req.ForestID,
		"offset":    req.Offset,
		"limit":     limit,
		"list_all":  req.ListAll,
	}
	filters := req.Filters
	if len(filters) == 0 {
		filters = []CoreKGFilter{{
			Field: "parent_id",
			Value: []string{parentID},
		}}
	}
	if len(filters) > 0 {
		payload["filters"] = filters
	}
	if len(req.OrderBy) > 0 {
		payload["orderBy"] = req.OrderBy
	}

	resp := &CoreKGListFileResponse{}
	if err := YyguRequest(ctx, listFilePath, payload, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func GetCorekgKnowledgeCountAndSize(ctx context.Context, corekgKnowledgeID uint) (count, size uint, err error) {
	resp := &corekgKnowledgeCountAndSizeResponse{}
	if err = YyguRequest(ctx, getForestPath, map[string]interface{}{
		"id": corekgKnowledgeID,
	}, resp); err != nil {
		if strings.Contains(err.Error(), "error code: 400") {
			logs.WarnContextf(ctx, "corekg knowledge id is deleted, corekg knowledge id is %d", corekgKnowledgeID)
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return resp.Data.FileCount, resp.Data.TotalSize, nil
}

type corekgKnowledgeCountAndSizeResponse struct {
	Data struct {
		TotalSize uint `json:"total_size"`
		FileCount uint `json:"file_count"`
	} `json:"data"`
}

type CoreKGKnowledgeSearchResponse struct {
	SearchResult []struct {
		ChunkList []struct {
			Content string  `json:"content"`
			Score   float64 `json:"score"`
		} `json:"chunk_list"`
		FileName string `json:"file_name"`
	} `json:"search_result"`
}

type CoreKGListForestRequest struct {
	Offset  int
	Limit   int
	ListAll bool
	OrderBy []string
	Filters []CoreKGFilter
}

type CoreKGFilter struct {
	Field      string   `json:"field"`
	Value      []string `json:"value"`
	ExactMatch bool     `json:"exact_match,omitempty"`
}

type CoreKGListForestResponse struct {
	Total  int64                 `json:"total"`
	Offset int                   `json:"offset"`
	Limit  int                   `json:"limit"`
	Data   []CoreKGForestSummary `json:"data"`
}

type CoreKGForestSummary struct {
	ID                uint      `json:"ID"`
	Name              string    `json:"name"`
	AvatarURL         string    `json:"avatar_url"`
	Description       string    `json:"description"`
	PublicScope       string    `json:"public_scope"`
	ForestType        string    `json:"forest_type"`
	DataSourceType    string    `json:"data_source_type"`
	DataSourceSubtype string    `json:"data_source_subtype"`
	ManagerIDs        []uint    `json:"manager_ids"`
	ScopeIDs          []uint    `json:"scope_ids"`
	FileCount         int64     `json:"file_count"`
	TotalSize         int64     `json:"total_size"`
	DiskStorage       string    `json:"disk_storage"`
	AppCount          uint      `json:"app_count"`
	IsAdmin           bool      `json:"is_admin"`
	IsSynced          bool      `json:"is_synced"`
	UpdatedAt         time.Time `json:"UpdatedAt"`
	CreatedAt         time.Time `json:"CreatedAt"`
	CompanyID         int64     `json:"company_id"`
	Uin               int64     `json:"uin"`
}

type CoreKGListFileRequest struct {
	ForestID int64
	ParentID string
	Limit    int
	Offset   int
	OrderBy  []string
	Filters  []CoreKGFilter
	ListAll  bool
}

type CoreKGListFileResponse struct {
	Total  int64            `json:"total"`
	Offset int              `json:"offset"`
	Limit  int              `json:"limit"`
	Data   []CoreKGFileNode `json:"data"`
}

type CoreKGFileNode struct {
	ID                int64      `json:"ID"`
	FileID            int64      `json:"file_id"`
	ForestID          int64      `json:"forest_id"`
	Name              string     `json:"name"`
	Ext               string     `json:"ext"`
	Size              int64      `json:"size"`
	IsDir             bool       `json:"is_dir"`
	ParentID          int64      `json:"parent_id"`
	ParentIDs         string     `json:"parent_ids"`
	Depth             int64      `json:"depth"`
	Status            string     `json:"status"`
	FileStatus        string     `json:"file_status"`
	FileProgress      string     `json:"file_progress"`
	ParseStatus       string     `json:"parse_status"`
	KnowledgeStatus   string     `json:"knowledge_status"`
	PreviewAble       string     `json:"preview_able"`
	PreviewTOSURL     string     `json:"preview_tos_url"`
	Uin               int64      `json:"uin"`
	CreatedAt         *time.Time `json:"CreatedAt"`
	UpdatedAt         *time.Time `json:"UpdatedAt"`
	DataSourceType    string     `json:"data_source_type"`
	DataSourceSubtype string     `json:"data_source_subtype"`
}
