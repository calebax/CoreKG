package searchctl

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/keapi/internal/dto/dtokeapi"
	"github.com/insmtx/corekg/apps/kesearch/models/globalsearch"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
)

// ForestSearch 知识库检索
func ForestSearch(ctx *gin.Context, req *dtokeapi.ForestSearchRequest, resp *dtokeapi.ForestSearchResponse) {
	if !req.ValidForestSearch(&resp.BaseResponse) {
		return
	}

	wrapper, err := globalsearch.NewForestWrapper(&globalsearch.GlobalSearchWrapper{
		Ctx:          ctx,
		Text:         req.Request.Query,
		Uin:          runtime.Uin(ctx),
		CompanyID:    runtime.CompanyID(ctx),
		ForestIDs:    req.Request.ForestIDs,
		EsIndex:      "ke_0",
		SubjectCount: 4,
		ItemCount:    1,
	})
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_searcher_failed"
		return
	}

	docList, err := wrapper.SearchFileAggs("chunk")
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_query_doc_failed"
		return
	}

	imageList, err := wrapper.SearchFileAggs("image")
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_query_image_failed"
		return
	}

	videoList, err := wrapper.SearchFileAggs("video")
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_query_video_failed"
		return
	}

	resp.Response.DocSearchResult = make([]*dtokeapi.SearchFile, 0, len(docList))
	for _, item := range docList {
		resp.Response.DocSearchResult = append(resp.Response.DocSearchResult, dtokeapi.NewSearchFile(item))
	}

	resp.Response.ImageSearchResult = make([]*dtokeapi.SearchFile, 0, len(imageList))
	for _, item := range imageList {
		resp.Response.ImageSearchResult = append(resp.Response.ImageSearchResult, dtokeapi.NewSearchFile(item))
	}

	resp.Response.VideoSearchResult = make([]*dtokeapi.SearchFile, 0, len(videoList))
	for _, item := range videoList {
		resp.Response.VideoSearchResult = append(resp.Response.VideoSearchResult, dtokeapi.NewSearchFile(item))
	}
}
