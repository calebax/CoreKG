// Deprecated : 目前看起来不太需要了

package adapter

import (
	"context"
	"errors"
	"strings"

	"github.com/insmtx/corekg/apps/workflow/crossdomain/knowledge"
	"github.com/insmtx/corekg/apps/workflow/crossdomain/knowledge/model"
	"github.com/insmtx/corekg/apps/workflow/utils/requestyygu"
)

// ErrUnsupported marks operations this adapter intentionally does not implement.
var ErrUnsupported = errors.New("yygu knowledge adapter: operation not supported")

// Ensure compile-time conformance.
var _ knowledge.Knowledge = (*YyguKnowledgeAdapter)(nil)

type YyguKnowledgeAdapter struct{}

func NewYyguKnowledgeAdapter() knowledge.Knowledge {
	return &YyguKnowledgeAdapter{}
}

func (a *YyguKnowledgeAdapter) ListKnowledge(ctx context.Context, req *model.ListKnowledgeRequest) (*model.ListKnowledgeResponse, error) {
	offset := 0
	limit := 0
	if req != nil {
		if req.Page != nil && req.PageSize != nil && *req.Page > 0 && *req.PageSize > 0 {
			offset = (*req.Page - 1) * (*req.PageSize)
			limit = *req.PageSize
		} else if req.PageSize != nil && *req.PageSize > 0 {
			limit = *req.PageSize
		}
	}

	resp, err := requestyygu.ListCoreKGForest(ctx, &requestyygu.CoreKGListForestRequest{
		Offset: offset,
		Limit:  limit,
		ListAll: func() bool {
			return limit == 0
		}(),
	})
	if err != nil {
		return nil, err
	}

	result := &model.ListKnowledgeResponse{
		KnowledgeList: make([]*model.Knowledge, 0, len(resp.Data)),
		Total:         resp.Total,
	}

	for _, item := range resp.Data {
		k := &model.Knowledge{
			Info: model.Info{
				ID:          int64(item.ID),
				Name:        item.Name,
				Description: item.Description,
				IconURI:     item.AvatarURL,
				IconURL:     item.AvatarURL,
				SpaceID:     item.CompanyID,
				CreatorID:   item.Uin,
				CreatedAtMs: item.CreatedAt.UnixMilli(),
				UpdatedAtMs: item.UpdatedAt.UnixMilli(),
			},
			Type:   mapYyguKnowledgeType(item),
			Status: model.KnowledgeStatusEnable,
		}
		result.KnowledgeList = append(result.KnowledgeList, k)
	}

	return result, nil
}

func (a *YyguKnowledgeAdapter) Retrieve(ctx context.Context, req *model.RetrieveRequest) (*model.RetrieveResponse, error) {
	if req == nil {
		return nil, errors.New("retrieve request is nil")
	}
	if len(req.KnowledgeIDs) == 0 {
		return nil, errors.New("knowledge ids is empty")
	}

	topK := int64(3)
	minScore := 0.5
	enableRerank := false
	if req.Strategy != nil {
		if req.Strategy.TopK != nil {
			topK = *req.Strategy.TopK
		}
		if req.Strategy.MinScore != nil {
			minScore = *req.Strategy.MinScore
		}
		enableRerank = req.Strategy.EnableRerank
	}

	resp, err := requestyygu.GetCoreKGKnowledgeSearchChunk(ctx, &requestyygu.CoreKGKnowledgeSearchRequest{
		IDs:          req.KnowledgeIDs,
		MinScore:     minScore,
		TopK:         topK,
		Query:        req.Query,
		EnableRerank: enableRerank,
	})
	if err != nil {
		return nil, err
	}

	out := &model.RetrieveResponse{RetrieveSlices: []*model.RetrieveSlice{}}
	for _, r := range resp.SearchResult {
		for _, c := range r.ChunkList {
			text := c.Content
			out.RetrieveSlices = append(out.RetrieveSlices, &model.RetrieveSlice{
				Slice: &model.Slice{
					DocumentName: r.FileName,
					RawContent: []*model.SliceContent{
						{Type: model.SliceContentTypeText, Text: &text},
					},
				},
				Score: c.Score,
			})
		}
	}

	return out, nil
}

func (a *YyguKnowledgeAdapter) DeleteKnowledge(ctx context.Context, request *model.DeleteKnowledgeRequest) error {
	return ErrUnsupported
}

func (a *YyguKnowledgeAdapter) GetKnowledgeByID(ctx context.Context, request *model.GetKnowledgeByIDRequest) (*model.GetKnowledgeByIDResponse, error) {
	return nil, ErrUnsupported
}

func (a *YyguKnowledgeAdapter) MGetKnowledgeByID(ctx context.Context, request *model.MGetKnowledgeByIDRequest) (*model.MGetKnowledgeByIDResponse, error) {
	return nil, ErrUnsupported
}

func (a *YyguKnowledgeAdapter) Store(ctx context.Context, document *model.CreateDocumentRequest) (*model.CreateDocumentResponse, error) {
	return nil, ErrUnsupported
}

func (a *YyguKnowledgeAdapter) Delete(ctx context.Context, r *model.DeleteDocumentRequest) (*model.DeleteDocumentResponse, error) {
	return nil, ErrUnsupported
}

func (a *YyguKnowledgeAdapter) ListKnowledgeDetail(ctx context.Context, req *model.ListKnowledgeDetailRequest) (*model.ListKnowledgeDetailResponse, error) {
	return nil, ErrUnsupported
}

func (a *YyguKnowledgeAdapter) MGetSlice(ctx context.Context, request *model.MGetSliceRequest) (*model.MGetSliceResponse, error) {
	return nil, ErrUnsupported
}

func (a *YyguKnowledgeAdapter) MGetDocument(ctx context.Context, request *model.MGetDocumentRequest) (*model.MGetDocumentResponse, error) {
	return nil, ErrUnsupported
}

func mapYyguKnowledgeType(item requestyygu.CoreKGForestSummary) model.DocumentType {
	candidates := []string{item.ForestType, item.DataSourceType, item.DataSourceSubtype}
	for _, v := range candidates {
		lv := strings.ToLower(strings.TrimSpace(v))
		switch {
		case lv == "":
			continue
		case lv == "qa":
			return model.DocumentTypeQAPair
		case lv == "data":
			continue
		case lv == "cad":
			return model.DocumentTypeMultimodal
		case strings.Contains(lv, "table"):
			return model.DocumentTypeExcel
		case strings.Contains(lv, "excel"):
			return model.DocumentTypeExcel
		case strings.Contains(lv, "database"), strings.Contains(lv, "db"):
			return model.DocumentTypeDatabase
		case strings.Contains(lv, "qa"), strings.Contains(lv, "q&a"), strings.Contains(lv, "question"):
			return model.DocumentTypeQAPair
		case strings.Contains(lv, "multi"):
			return model.DocumentTypeMultimodal
		}
	}
	return model.DocumentTypeMultimodal
}
