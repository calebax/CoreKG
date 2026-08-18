package coze

import (
	"github.com/insmtx/corekg/apps/kesearch/models/globalsearch"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type KnowledgeSearchRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID    uint   `json:"forest_id"`
		Text        string `json:"text"`
		IsSemantics bool   `json:"is_semantics"` // 是否语义搜索
	}
}

type ForestSearchResponse struct {
	apiobj.BaseResponse
	Response struct {
		DocSearchResult   []*globalsearch.SearchFileType `json:"doc_search_result,omitempty"`   // 文档搜索结果
		ImageSearchResult []*globalsearch.SearchFileType `json:"image_search_result,omitempty"` // 图片搜索结果
		VideoSearchResult []*globalsearch.SearchFileType `json:"video_search_result,omitempty"` // 视频搜索结果
	}
}
