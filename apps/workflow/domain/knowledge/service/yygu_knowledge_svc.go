package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	knowledgeModel "github.com/insmtx/corekg/apps/workflow/crossdomain/knowledge/model"
	"github.com/insmtx/corekg/apps/workflow/domain/knowledge/entity"
	"github.com/insmtx/corekg/apps/workflow/infra/document/parser"
	"github.com/insmtx/corekg/apps/workflow/pkg/errorx"
	"github.com/insmtx/corekg/apps/workflow/types/errno"
	"github.com/insmtx/corekg/apps/workflow/utils/requestyygu"
)

// ErrYyguUnsupported marks the operations that are not provided by the YYGU knowledge service.
var ErrYyguUnsupported = errors.New("yygu knowledge service: operation not supported")

type yyguKnowledgeSVC struct{}

func NewYyguKnowledgeSVC() Knowledge {
	return &yyguKnowledgeSVC{}
}

func (s *yyguKnowledgeSVC) CreateKnowledge(ctx context.Context, request *CreateKnowledgeRequest) (*CreateKnowledgeResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) UpdateKnowledge(ctx context.Context, request *UpdateKnowledgeRequest) error {
	return ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) DeleteKnowledge(ctx context.Context, request *DeleteKnowledgeRequest) error {
	return ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) CopyKnowledge(ctx context.Context, request *CopyKnowledgeRequest) (*CopyKnowledgeResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) MoveKnowledgeToLibrary(ctx context.Context, request *MoveKnowledgeToLibraryRequest) error {
	return ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) ListKnowledge(ctx context.Context, request *ListKnowledgeRequest) (*ListKnowledgeResponse, error) {
	offset := 0
	limit := 0
	var filters []requestyygu.CoreKGFilter
	if request != nil {
		if request.Page != nil && request.PageSize != nil && *request.Page > 0 && *request.PageSize > 0 {
			offset = (*request.Page - 1) * (*request.PageSize)
			limit = *request.PageSize
		} else if request.PageSize != nil && *request.PageSize > 0 {
			limit = *request.PageSize
		}
		if request.FormatType != nil {
			filters = append(filters, requestyygu.CoreKGFilter{
				Field: "forest_type",
				Value: []string{coreKGForestTypeFromFormatType(*request.FormatType)},
			})
		}
	}

	resp, err := requestyygu.ListCoreKGForest(ctx, &requestyygu.CoreKGListForestRequest{
		Offset: offset,
		Limit:  limit,
		ListAll: func() bool {
			return limit == 0
		}(),
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}

	out := &ListKnowledgeResponse{
		KnowledgeList: make([]*knowledgeModel.Knowledge, 0, len(resp.Data)),
		Total:         resp.Total,
	}
	for _, item := range resp.Data {
		out.KnowledgeList = append(out.KnowledgeList, &knowledgeModel.Knowledge{
			Info: knowledgeModel.Info{
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
			Status: knowledgeModel.KnowledgeStatusEnable,
		})
	}

	return out, nil
}

func (s *yyguKnowledgeSVC) GetKnowledgeByID(ctx context.Context, request *GetKnowledgeByIDRequest) (*GetKnowledgeByIDResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) MGetKnowledgeByID(ctx context.Context, request *MGetKnowledgeByIDRequest) (*MGetKnowledgeByIDResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) CreateDocument(ctx context.Context, request *CreateDocumentRequest) (*CreateDocumentResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) UpdateDocument(ctx context.Context, request *UpdateDocumentRequest) error {
	return ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) DeleteDocument(ctx context.Context, request *DeleteDocumentRequest) error {
	return ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) ExtractPhotoCaption(ctx context.Context, request *ExtractPhotoCaptionRequest) (*ExtractPhotoCaptionResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) ListDocument(ctx context.Context, request *ListDocumentRequest) (*ListDocumentResponse, error) {
	if request == nil {
		return nil, errorx.New(errno.ErrKnowledgeInvalidParamCode, errorx.KV("msg", "request is empty"))
	}
	if request.KnowledgeID == 0 {
		return nil, errorx.New(errno.ErrKnowledgeInvalidParamCode, errorx.KV("msg", "knowledge id is empty"))
	}

	limit := 10
	offset := 0
	if request.Limit != nil {
		limit = *request.Limit
	}
	if request.Offset != nil {
		offset = *request.Offset
	}
	listAll := request.SelectAll || len(request.DocumentIDs) > 0
	if listAll {
		limit = 0
		offset = 0
	}

	listResp, err := requestyygu.ListCoreKGFile(ctx, &requestyygu.CoreKGListFileRequest{
		ForestID: request.KnowledgeID,
		Offset:   offset,
		Limit:    limit,
		ListAll:  listAll,
	})
	if err != nil {
		return nil, err
	}

	documents := make([]*entity.Document, 0, len(listResp.Data))
	for _, item := range listResp.Data {
		if item.IsDir {
			continue
		}
		if !matchDocumentID(request.DocumentIDs, item.ID, item.FileID) {
			continue
		}
		if request.Keyword != nil && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(*request.Keyword)) {
			continue
		}
		documents = append(documents, mapYyguFileNodeToDocument(item, request.KnowledgeID))
	}

	total := listResp.Total
	if listAll || len(request.DocumentIDs) > 0 || request.Keyword != nil {
		total = int64(len(documents))
	}

	resp := &ListDocumentResponse{
		Documents: documents,
		Total:     total,
	}

	effectiveLimit := limit
	if listResp.Limit > 0 {
		effectiveLimit = listResp.Limit
	}
	if effectiveLimit > 0 && int64(offset+effectiveLimit) < listResp.Total {
		resp.HasMore = true
		if len(documents) > 0 {
			cursor := strconv.FormatInt(documents[len(documents)-1].ID, 10)
			resp.NextCursor = &cursor
		}
	}

	return resp, nil
}

func (s *yyguKnowledgeSVC) MGetDocumentProgress(ctx context.Context, request *MGetDocumentProgressRequest) (*MGetDocumentProgressResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) ResegmentDocument(ctx context.Context, request *ResegmentDocumentRequest) (*ResegmentDocumentResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) GetAlterTableSchema(ctx context.Context, request *AlterTableSchemaRequest) (*TableSchemaResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) ValidateTableSchema(ctx context.Context, request *ValidateTableSchemaRequest) (*ValidateTableSchemaResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) GetDocumentTableInfo(ctx context.Context, request *GetDocumentTableInfoRequest) (*GetDocumentTableInfoResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) GetImportDataTableSchema(ctx context.Context, request *ImportDataTableSchemaRequest) (*TableSchemaResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) CreateSlice(ctx context.Context, request *CreateSliceRequest) (*CreateSliceResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) UpdateSlice(ctx context.Context, request *UpdateSliceRequest) error {
	return ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) DeleteSlice(ctx context.Context, request *DeleteSliceRequest) error {
	return ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) ListSlice(ctx context.Context, request *ListSliceRequest) (*ListSliceResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) ListPhotoSlice(ctx context.Context, request *ListPhotoSliceRequest) (*ListPhotoSliceResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) GetSlice(ctx context.Context, request *GetSliceRequest) (*GetSliceResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) MGetSlice(ctx context.Context, request *MGetSliceRequest) (*MGetSliceResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) MGetDocument(ctx context.Context, request *MGetDocumentRequest) (*MGetDocumentResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) Retrieve(ctx context.Context, request *RetrieveRequest) (*RetrieveResponse, error) {
	if request == nil {
		return nil, errors.New("retrieve request is nil")
	}
	if len(request.KnowledgeIDs) == 0 {
		return nil, errors.New("knowledge ids is empty")
	}

	topK := int64(3)
	minScore := 0.5
	enableRerank := false
	if request.Strategy != nil {
		if request.Strategy.TopK != nil {
			topK = *request.Strategy.TopK
		}
		if request.Strategy.MinScore != nil {
			minScore = *request.Strategy.MinScore
		}
		enableRerank = request.Strategy.EnableRerank
	}

	resp, err := requestyygu.GetCoreKGKnowledgeSearchChunk(ctx, &requestyygu.CoreKGKnowledgeSearchRequest{
		IDs:          request.KnowledgeIDs,
		MinScore:     minScore,
		TopK:         topK,
		Query:        request.Query,
		EnableRerank: enableRerank,
	})
	if err != nil {
		return nil, err
	}

	out := &RetrieveResponse{RetrieveSlices: []*knowledgeModel.RetrieveSlice{}}
	for _, r := range resp.SearchResult {
		for _, c := range r.ChunkList {
			text := c.Content
			out.RetrieveSlices = append(out.RetrieveSlices, &knowledgeModel.RetrieveSlice{
				Slice: &knowledgeModel.Slice{
					DocumentName: r.FileName,
					RawContent: []*knowledgeModel.SliceContent{
						{Type: knowledgeModel.SliceContentTypeText, Text: &text},
					},
				},
				Score: c.Score,
			})
		}
	}

	return out, nil
}

func (s *yyguKnowledgeSVC) CreateDocumentReview(ctx context.Context, request *CreateDocumentReviewRequest) (*CreateDocumentReviewResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) MGetDocumentReview(ctx context.Context, request *MGetDocumentReviewRequest) (*MGetDocumentReviewResponse, error) {
	return nil, ErrYyguUnsupported
}

func (s *yyguKnowledgeSVC) SaveDocumentReview(ctx context.Context, request *SaveDocumentReviewRequest) error {
	return ErrYyguUnsupported
}

func coreKGForestTypeFromFormatType(formatType knowledgeModel.DocumentType) string {
	switch formatType {
	case knowledgeModel.DocumentTypeQAPair:
		return "qa"
	case knowledgeModel.DocumentTypeMultimodal:
		return "file"
	default:
		return ""
	}
}

func mapYyguKnowledgeType(item requestyygu.CoreKGForestSummary) knowledgeModel.DocumentType {
	candidates := []string{item.ForestType, item.DataSourceType, item.DataSourceSubtype}
	for _, v := range candidates {
		lv := strings.ToLower(strings.TrimSpace(v))
		switch {
		case lv == "":
			continue
		case lv == "qa":
			return knowledgeModel.DocumentTypeQAPair
		case lv == "data":
			continue
		case lv == "cad":
			return knowledgeModel.DocumentTypeMultimodal
		case strings.Contains(lv, "table"):
			return knowledgeModel.DocumentTypeTable
		case strings.Contains(lv, "excel"):
			return knowledgeModel.DocumentTypeExcel
		case strings.Contains(lv, "database"), strings.Contains(lv, "db"):
			return knowledgeModel.DocumentTypeDatabase
		case strings.Contains(lv, "qa"), strings.Contains(lv, "q&a"), strings.Contains(lv, "question"):
			return knowledgeModel.DocumentTypeQAPair
		case strings.Contains(lv, "multi"):
			return knowledgeModel.DocumentTypeMultimodal
		case strings.Contains(lv, "image"), strings.Contains(lv, "pic"):
			return knowledgeModel.DocumentTypeImage
		}
	}
	return knowledgeModel.DocumentTypeMultimodal
}

func mapYyguFileNodeToDocument(node requestyygu.CoreKGFileNode, knowledgeID int64) *entity.Document {
	ext := strings.TrimPrefix(strings.ToLower(node.Ext), ".")
	docType := mapYyguDocumentType(ext)
	status := mapYyguFileStatus(node)
	statusMsg := firstNonEmpty(node.FileProgress, node.ParseStatus, node.FileStatus, node.Status)
	uri := node.PreviewTOSURL

	return &entity.Document{
		Info: knowledgeModel.Info{
			ID:          node.ID,
			Name:        node.Name,
			CreatorID:   node.Uin,
			CreatedAtMs: toMilli(node.CreatedAt),
			UpdatedAtMs: toMilli(node.UpdatedAt),
		},
		KnowledgeID:   knowledgeID,
		Type:          docType,
		URI:           uri,
		URL:           uri,
		Size:          node.Size,
		FileExtension: parser.FileExtension(ext),
		Status:        status,
		StatusMsg:     statusMsg,
		Source:        entity.DocumentSourceLocal,
	}
}

func mapYyguDocumentType(ext string) knowledgeModel.DocumentType {
	switch ext {
	case "csv", "xlsx", "json", "json_maps":
		return knowledgeModel.DocumentTypeTable
	case "jpg", "jpeg", "png":
		return knowledgeModel.DocumentTypeImage
	default:
		return knowledgeModel.DocumentTypeText
	}
}

func mapYyguFileStatus(node requestyygu.CoreKGFileNode) entity.DocumentStatus {
	status := strings.ToLower(firstNonEmpty(node.FileStatus, node.Status, node.ParseStatus))
	switch {
	case status == "":
		return entity.DocumentStatusEnable
	case strings.Contains(status, "fail"), strings.Contains(status, "error"):
		return entity.DocumentStatusFailed
	case strings.Contains(status, "delete"):
		return entity.DocumentStatusDeleted
	case strings.Contains(status, "finish"), strings.Contains(status, "success"), strings.Contains(status, "done"), strings.Contains(status, "complete"):
		return entity.DocumentStatusEnable
	default:
		return entity.DocumentStatusChunking
	}
}

func matchDocumentID(targetIDs []int64, ids ...int64) bool {
	if len(targetIDs) == 0 {
		return true
	}
	for _, target := range targetIDs {
		for _, id := range ids {
			if target == id {
				return true
			}
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func toMilli(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.UnixMilli()
}
