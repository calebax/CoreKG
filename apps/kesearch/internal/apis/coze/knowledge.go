package coze

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kesearch/models/globalsearch"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// KnowledgeSearch coze调用知识库检索
// @Tags 知识库搜索
// @Summary 知识库搜索文档类型
// @Description 知识库搜索文档类型
// @Router /kesearch.KnowledgeSearch [post]
// @Param user body KnowledgeSearchRequest true "入参"
// @Success 200 {object} ForestSearchResponse "返回值"
func KnowledgeSearch(ctx *gin.Context, req *KnowledgeSearchRequest, resp *ForestSearchResponse) {

	uinInfo, err := forest.NewForestDao().GetByID(ctx, req.Request.ForestID)

	uin := uinInfo.Uin
	company_id := uinInfo.CompanyID
	forest_info, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestByID error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "获取知识森林失败"
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
		SubjectCount: 50,
		ItemCount:    30,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "HandelGlobalSearch error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "获取搜索器失败"
		return
	}
	resdoc, err := wrapper.SearchFileAggs("chunk")
	if err != nil {
		logs.ErrorContextf(ctx, "SearchFile error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "查询文档失败"
		return
	}
	resp.Response.DocSearchResult = resdoc
}
