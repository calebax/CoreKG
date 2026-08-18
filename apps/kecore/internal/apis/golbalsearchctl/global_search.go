package golbalsearchctl

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kesearch/models/globalsearch"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// GlobalSearch 全局搜索
// @Tags 全局搜索
// @Summary 全局搜索
// @Description 全局搜索
// @Router /forest.GlobalSearch [post]
// @Param user body GlobalSearchRequest true "入参"
// @Success 200 {object} GlobalSearchResponse "返回值"
func GlobalSearch(ctx *gin.Context, req *GlobalSearchRequest, resp *GlobalSearchResponse) {
	if req.Validity(ctx, resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	company_id := runtime.CompanyID(ctx)
	wrapper, err := globalsearch.NewForestWrapper(&globalsearch.GlobalSearchWrapper{
		Ctx:         ctx,
		Text:        req.Request.Text,
		Uin:         uin,
		CompanyID:   company_id,
		IsSemantics: req.Request.IsSemantics,
		// TODO: 这里需要查公司所有索引并且进行检索
		EsIndex:      "ke_0",
		ImageUrl:     req.Request.ImageUrl,
		SubjectCount: 4,
		ItemCount:    1,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "HandelGlobalSearch error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_searcher_failed" // 获取搜索器失败
		return
	}
	if req.Request.IsSemantics {
		wrapper.ItemCount = 3
	}
	resdoc, err := wrapper.SearchFileAggs("chunk")
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_doc_failed" // 查询文档失败
		return
	}
	resimg, err := wrapper.SearchFileAggs("image")
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_image_failed" // 查询图片失败
		return
	}
	resvideo, err := wrapper.SearchFileAggs("video")
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_video_failed" // 查询视频失败
		return
	}
	resagent, err := wrapper.SearchAgent()
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_agent_failed" // 查询智能体失败
		return
	}
	resforest, err := wrapper.SearchForest()
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_forest_failed" // 查询知识库失败
		return
	}
	resp.Response.ForestSearchResult = resforest
	resp.Response.AgentSearchResult = resagent
	resp.Response.VideoSearchResult = resvideo
	resp.Response.ImageSearchResult = resimg
	resp.Response.DocSearchResult = resdoc
}

// GlobalSearchDoc 全局搜索文档类型
// @Tags 全局搜索
// @Summary 全局搜索文档类型
// @Description 全局搜索文档类型
// @Router /forest.GlobalSearchDoc [post]
// @Param user body GlobalSearchRequest true "入参"
// @Success 200 {object} GlobalSearchResponse "返回值"
func GlobalSearchDoc(ctx *gin.Context, req *GlobalSearchRequest, resp *GlobalSearchResponse) {
	if req.Validity(ctx, resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	company_id := runtime.CompanyID(ctx)
	wrapper, err := globalsearch.NewForestWrapper(&globalsearch.GlobalSearchWrapper{
		Ctx:         ctx,
		Text:        req.Request.Text,
		Uin:         uin,
		CompanyID:   company_id,
		IsSemantics: req.Request.IsSemantics,
		// TODO: 这里需要查公司所有索引并且进行检索
		EsIndex:      "ke_0",
		ImageUrl:     req.Request.ImageUrl,
		SubjectCount: 50,
		ItemCount:    30,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "HandelGlobalSearch error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_searcher_failed" // 获取搜索器失败
		return
	}
	resdoc, err := wrapper.SearchFileAggs("chunk")
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_doc_failed" // 查询文档失败
		return
	}
	resp.Response.DocSearchResult = resdoc
}

// GlobalSearchImage 全局搜索图片类型
// @Tags 全局搜索
// @Summary 全局搜索图片类型
// @Description 全局搜索图片类型
// @Router /forest.GlobalSearchImage [post]
// @Param user body GlobalSearchRequest true "入参"
// @Success 200 {object} GlobalSearchResponse "返回值"
func GlobalSearchImage(ctx *gin.Context, req *GlobalSearchRequest, resp *GlobalSearchResponse) {
	if req.Validity(ctx, resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	company_id := runtime.CompanyID(ctx)
	wrapper, err := globalsearch.NewForestWrapper(&globalsearch.GlobalSearchWrapper{
		Ctx:         ctx,
		Text:        req.Request.Text,
		Uin:         uin,
		CompanyID:   company_id,
		IsSemantics: req.Request.IsSemantics,
		// TODO: 这里需要查公司所有索引并且进行检索
		EsIndex:      "ke_0",
		ImageUrl:     req.Request.ImageUrl,
		SubjectCount: 50,
		ItemCount:    30,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "HandelGlobalSearch error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_searcher_failed" // 获取搜索器失败
		return
	}
	resimg, err := wrapper.SearchFileAggs("image")
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_image_failed" // 查询图片失败
		return
	}
	resp.Response.ImageSearchResult = resimg
}

// GlobalSearchVideo 全局搜索视频类型
// @Tags 全局搜索
// @Summary 全局搜索视频类型
// @Description 全局搜索视频类型
// @Router /forest.GlobalSearchVideo [post]
// @Param user body GlobalSearchRequest true "入参"
// @Success 200 {object} GlobalSearchResponse "返回值"
func GlobalSearchVideo(ctx *gin.Context, req *GlobalSearchRequest, resp *GlobalSearchResponse) {
	if req.Validity(ctx, resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	company_id := runtime.CompanyID(ctx)
	wrapper, err := globalsearch.NewForestWrapper(&globalsearch.GlobalSearchWrapper{
		Ctx:         ctx,
		Text:        req.Request.Text,
		Uin:         uin,
		CompanyID:   company_id,
		IsSemantics: req.Request.IsSemantics,
		// TODO: 这里需要查公司所有索引并且进行检索
		EsIndex:      "ke_0",
		ImageUrl:     req.Request.ImageUrl,
		SubjectCount: 50,
		ItemCount:    30,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "HandelGlobalSearch error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_searcher_failed" // 获取搜索器失败
		return
	}
	resvideo, err := wrapper.SearchFileAggs("video")
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_video_failed" // 查询视频失败
		return
	}
	resp.Response.VideoSearchResult = resvideo
}

// GlobalSearchAgent 全局搜索智能体类型
// @Tags 全局搜索
// @Summary 全局搜索智能体类型
// @Description 全局搜索智能体类型
// @Router /forest.GlobalSearchAgent [post]
// @Param user body GlobalSearchRequest true "入参"
// @Success 200 {object} GlobalSearchResponse "返回值"
func GlobalSearchAgent(ctx *gin.Context, req *GlobalSearchRequest, resp *GlobalSearchResponse) {
	if req.Validity(ctx, resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	company_id := runtime.CompanyID(ctx)
	wrapper, err := globalsearch.NewForestWrapper(&globalsearch.GlobalSearchWrapper{
		Ctx:         ctx,
		Text:        req.Request.Text,
		Uin:         uin,
		CompanyID:   company_id,
		IsSemantics: req.Request.IsSemantics,
		// TODO: 这里需要查公司所有索引并且进行检索
		EsIndex:      "ke_0",
		ImageUrl:     req.Request.ImageUrl,
		SubjectCount: 50,
		ItemCount:    30,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "HandelGlobalSearch error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_searcher_failed" // 获取搜索器失败
		return
	}
	resagent, err := wrapper.SearchAgent()
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_agent_failed" // 查询智能体失败
		return
	}
	resp.Response.AgentSearchResult = resagent
}

// GlobalSearchForest 全局搜索知识库
// @Tags 全局搜索
// @Summary 全局搜索知识库
// @Description 全局搜索知识库
// @Router /forest.GlobalSearchForest [post]
// @Param user body GlobalSearchRequest true "入参"
// @Success 200 {object} GlobalSearchResponse "返回值"
func GlobalSearchForest(ctx *gin.Context, req *GlobalSearchRequest, resp *GlobalSearchResponse) {
	if req.Validity(ctx, resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	company_id := runtime.CompanyID(ctx)
	wrapper, err := globalsearch.NewForestWrapper(&globalsearch.GlobalSearchWrapper{
		Ctx:         ctx,
		Text:        req.Request.Text,
		Uin:         uin,
		CompanyID:   company_id,
		IsSemantics: req.Request.IsSemantics,
		// TODO: 这里需要查公司所有索引并且进行检索
		EsIndex:      "ke_0",
		ImageUrl:     req.Request.ImageUrl,
		SubjectCount: 50,
		ItemCount:    30,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "HandelGlobalSearch error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_searcher_failed" // 获取搜索器失败
		return
	}
	resforest, err := wrapper.SearchForest()
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_forest_failed" // 查询知识库失败
		return
	}
	resp.Response.ForestSearchResult = resforest
}
