package globalsearchctl

import (
	"github.com/insmtx/corekg/apps/kesearch/models/globalsearch"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type ForestSearchRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID    uint   `json:"forest_id"`
		Text        string `json:"text"`
		IsSemantics bool   `json:"is_semantics"` // 是否语义搜索
		ImageUrl    string `json:"image_url"`
	}
}

func (req *ForestSearchRequest) Validity(resp *ForestSearchResponse) {
	if req.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_select_forest" // 请选择知识森林
		return
	}
	if req.Request.Text == "" && req.Request.ImageUrl == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_enter_search_content" // 请输入搜索内容
		return
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
