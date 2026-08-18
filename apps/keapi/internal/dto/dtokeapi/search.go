package dtokeapi

import (
	"time"

	"github.com/insmtx/corekg/apps/kesearch/models/globalsearch"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// ForestSearchRequest 知识库检索请求
type ForestSearchRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestIDs []uint `json:"forest_ids"`
		Query     string `json:"query"`
	} `json:"request"`
}

func (req *ForestSearchRequest) ValidForestSearch(resp *apiobj.BaseResponse) bool {
	if len(req.Request.ForestIDs) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_forest_ids"
		return false
	}
	if req.Request.Query == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_empty_query"
		return false
	}
	return true
}

type SearchHighlight struct {
	Score                  float64 `json:"_score,omitempty"`
	Description            string  `json:"description"`
	HighlightedDescription string  `json:"highlighted_description,omitempty"`
	ImageURL               string  `json:"image_url,omitempty"`
	Location               [5]int  `json:"location,omitempty"`
}

type SearchFile struct {
	ForestFileID uint              `json:"forest_file_id"`
	ForestID     uint              `json:"forest_id"`
	FileName     string            `json:"file_name"`
	CreatedAt    time.Time         `json:"created_at"`
	Score        float64           `json:"_score,omitempty"`
	Highlights   []SearchHighlight `json:"highlights,omitempty"`
}

func NewSearchFile(item *globalsearch.SearchFileType) *SearchFile {
	if item == nil {
		return nil
	}
	out := &SearchFile{
		ForestFileID: item.ID,
		ForestID:     item.ForestID,
		FileName:     item.FileName,
		CreatedAt:    item.CreatedAt,
		Score:        item.Score,
		Highlights:   make([]SearchHighlight, 0, len(item.Highlights)),
	}
	for _, highlight := range item.Highlights {
		out.Highlights = append(out.Highlights, SearchHighlight{
			Score:                  highlight.Score,
			Description:            highlight.Description,
			HighlightedDescription: highlight.HighlightedDescription,
			ImageURL:               highlight.ImageURL,
			Location:               highlight.Location,
		})
	}
	return out
}

// ForestSearchResponse 知识库检索响应
type ForestSearchResponse struct {
	apiobj.BaseResponse
	Response struct {
		DocSearchResult   []*SearchFile `json:"doc_search_result,omitempty"`
		ImageSearchResult []*SearchFile `json:"image_search_result,omitempty"`
		VideoSearchResult []*SearchFile `json:"video_search_result,omitempty"`
	} `json:"response"`
}
