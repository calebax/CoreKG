package globalsearchctl

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kesearch/models/globalsearch"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// ForestSearch 知识库搜索
// @Tags 知识库搜索
// @Summary 知识库搜索
// @Description 知识库搜索
// @Router /kesearch.ForestSearch [post]
// @Param user body ForestSearchRequest true "入参"
// @Success 200 {object} ForestSearchResponse "返回值"
func ForestSearch(ctx *gin.Context, req *ForestSearchRequest, resp *ForestSearchResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	company_id := runtime.CompanyID(ctx)
	forest_info, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestByID error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_forest_failed" // 获取知识森林失败
		return
	}
	wrapper, err := globalsearch.NewForestWrapper(&globalsearch.GlobalSearchWrapper{
		Ctx:          ctx,
		Text:         req.Request.Text,
		Uin:          uin,
		ForestIDs:    []uint{forest_info.ID},
		CompanyID:    company_id,
		IsSemantics:  req.Request.IsSemantics,
		EsIndex:      forest_info.EsIndex(),
		ImageUrl:     req.Request.ImageUrl,
		SubjectCount: 4,
		ItemCount:    1,
	})
	if req.Request.IsSemantics {
		wrapper.ItemCount = 3
	}
	if err != nil {
		logs.ErrorContextf(ctx, "HandelGlobalSearch error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_searcher_failed" // 获取搜索器失败
		return
	}
	resdoc, err := wrapper.SearchFileAggs("chunk")
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_query_doc_failed" // 查询文档失败
		return
	}
	resimg, err := wrapper.SearchFileAggs("image")
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_query_image_failed" // 查询图片失败
		return
	}
	resvideo, err := wrapper.SearchFileAggs("video")
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_query_video_failed" // 查询视频失败
		return
	}
	resp.Response.VideoSearchResult = resvideo
	resp.Response.ImageSearchResult = resimg
	resp.Response.DocSearchResult = resdoc
}

// ForestSearchDoc 知识库搜索文档类型
// @Tags 知识库搜索
// @Summary 知识库搜索文档类型
// @Description 知识库搜索文档类型
// @Router /kesearch.ForestSearchDoc [post]
// @Param user body ForestSearchRequest true "入参"
// @Success 200 {object} ForestSearchResponse "返回值"
func ForestSearchDoc(ctx *gin.Context, req *ForestSearchRequest, resp *ForestSearchResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	company_id := runtime.CompanyID(ctx)
	forest_info, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestByID error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_forest_failed" // 获取知识森林失败
		return
	}
	wrapper, err := globalsearch.NewForestWrapper(&globalsearch.GlobalSearchWrapper{
		Ctx:          ctx,
		Text:         req.Request.Text,
		Uin:          uin,
		ForestIDs:    []uint{forest_info.ID},
		CompanyID:    company_id,
		IsSemantics:  req.Request.IsSemantics,
		EsIndex:      forest_info.EsIndex(),
		ImageUrl:     req.Request.ImageUrl,
		SubjectCount: 50,
		ItemCount:    30,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "HandelGlobalSearch error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_searcher_failed" // 获取搜索器失败
		return
	}
	resdoc, err := wrapper.SearchFileAggs("chunk")
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_query_doc_failed" // 查询文档失败
		return
	}
	resp.Response.DocSearchResult = resdoc
}

// ForestSearchImage 知识库搜索图片类型
// @Tags 知识库搜索
// @Summary 知识库搜索图片类型
// @Description 知识库搜索图片类型
// @Router /kesearch.ForestSearchImage [post]
// @Param user body ForestSearchRequest true "入参"
// @Success 200 {object} ForestSearchResponse "返回值"
func ForestSearchImage(ctx *gin.Context, req *ForestSearchRequest, resp *ForestSearchResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	company_id := runtime.CompanyID(ctx)
	forest_info, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestByID error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_forest_failed" // 获取知识森林失败
		return
	}
	wrapper, err := globalsearch.NewForestWrapper(&globalsearch.GlobalSearchWrapper{
		Ctx:          ctx,
		Text:         req.Request.Text,
		Uin:          uin,
		ForestIDs:    []uint{forest_info.ID},
		CompanyID:    company_id,
		IsSemantics:  req.Request.IsSemantics,
		EsIndex:      forest_info.EsIndex(),
		ImageUrl:     req.Request.ImageUrl,
		SubjectCount: 50,
		ItemCount:    30,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "HandelGlobalSearch error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_searcher_failed" // 获取搜索器失败
		return
	}
	resimg, err := wrapper.SearchFileAggs("image")
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_query_image_failed" // 查询图片失败
		return
	}
	resp.Response.ImageSearchResult = resimg
}

// ForestSearchVideo 知识库搜索视频类型
// @Tags 知识库搜索
// @Summary 知识库搜索视频类型
// @Description 知识库搜索视频类型
// @Router /kesearch.ForestSearchVideo [post]
// @Param user body ForestSearchRequest true "入参"
// @Success 200 {object} ForestSearchResponse "返回值"
func ForestSearchVideo(ctx *gin.Context, req *ForestSearchRequest, resp *ForestSearchResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	company_id := runtime.CompanyID(ctx)
	forest_info, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestByID error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_forest_failed" // 获取知识森林失败
		return
	}

	wrapper, err := globalsearch.NewForestWrapper(&globalsearch.GlobalSearchWrapper{
		Ctx:          ctx,
		Text:         req.Request.Text,
		Uin:          uin,
		ForestIDs:    []uint{forest_info.ID},
		CompanyID:    company_id,
		IsSemantics:  req.Request.IsSemantics,
		EsIndex:      forest_info.EsIndex(),
		ImageUrl:     req.Request.ImageUrl,
		SubjectCount: 50,
		ItemCount:    30,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "HandelGlobalSearch error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_searcher_failed" // 获取搜索器失败
		return
	}
	resvideo, err := wrapper.SearchFileAggs("video")
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_query_video_failed" // 查询视频失败
		return
	}
	resp.Response.VideoSearchResult = resvideo
}
