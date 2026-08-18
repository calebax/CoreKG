package golbalsearchctl

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/kesearch/models/globalsearch"
	"github.com/insmtx/corekg/version"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type GlobalSearchRequest struct {
	apiobj.BaseRequest
	Request struct {
		Text        string `json:"text"`
		IsSemantics bool   `json:"is_semantics"` // 是否语义搜索
		ImageUrl    string `json:"image_url"`
	}
}

func (req *GlobalSearchRequest) Validity(ctx *gin.Context, resp *GlobalSearchResponse) {
	if req.Request.Text == "" && req.Request.ImageUrl == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_enter_content" // 请输入搜索内容
		return
	}
	if req.Request.ImageUrl != "" && version.DeployMode() != "" {
		req.Request.ImageUrl = fs.SpliceUrl(ctx, req.Request.ImageUrl, ctx.GetHeader("Referer"))
	}
}

type GlobalSearchResponse struct {
	apiobj.BaseResponse
	Response struct {
		DocSearchResult    []*globalsearch.SearchFileType   `json:"doc_search_result,omitempty"`    // 文档搜索结果
		ImageSearchResult  []*globalsearch.SearchFileType   `json:"image_search_result,omitempty"`  // 图片搜索结果
		VideoSearchResult  []*globalsearch.SearchFileType   `json:"video_search_result,omitempty"`  // 视频搜索结果
		AgentSearchResult  []*globalsearch.SearchAgentType  `json:"agent_search_result,omitempty"`  // 智能体搜索结果
		ForestSearchResult []*globalsearch.SearchForestType `json:"forest_search_result,omitempty"` // 知识森林搜索结果
	}
}
