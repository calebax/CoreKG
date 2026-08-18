package globalsearchctl

import (
	"github.com/insmtx/corekg/apps/kesearch/models/globalsearch"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type GlobalSearchRequest struct {
	apiobj.BaseRequest
	Request struct {
		// Text 搜索内容
		Text string `json:"text"`
		// IsSemantics 是否语义搜索
		IsSemantics bool `json:"is_semantics"`
		// ImageUrl 图片地址
		ImageUrl string `json:"image_url"`
		// ForestIDs 知识库ID列表
		ForestIDs []uint `json:"forest_ids"`
	}
}

func (req *GlobalSearchRequest) Validity(resp *GlobalSearchResponse) {
	if req.Request.Text == "" && req.Request.ImageUrl == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kesearch_enter_search_content" // 请输入搜索内容
		return
	}
}

type GlobalSearchResponse struct {
	apiobj.BaseResponse
	Response struct {
		DocSearchResult      []*globalsearch.SearchFileType       `json:"doc_search_result,omitempty"`      // 文档搜索结果
		ImageSearchResult    []*globalsearch.SearchFileType       `json:"image_search_result,omitempty"`    // 图片搜索结果
		VideoSearchResult    []*globalsearch.SearchFileType       `json:"video_search_result,omitempty"`    // 视频搜索结果
		AgentSearchResult    []*globalsearch.SearchAgentType      `json:"agent_search_result,omitempty"`    // 智能体搜索结果
		ForestSearchResult   []*globalsearch.SearchForestType     `json:"forest_search_result,omitempty"`   // 知识森林搜索结果
		ExternalSearchResult *globalsearch.SearchExternalDataType `json:"external_search_result,omitempty"` // 外部数据源搜索结果
	}
}
