package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoarticle"
	"github.com/insmtx/corekg/apps/kecore/services/svcarticle"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// SaveAsArticleTemplateDeprecated 保存为写作模板（旧版）
// Deprecated: 请使用 apis.SaveAsArticleTemplate（统一文章接口，通过 type 参数区分）
// @Tags 写作空间
// @Summary 保存为写作模板
// @Description 保存为写作模板
// @Router /forest.SaveAsArticleTemplate [post]
// @Param request body dtoarticle.SaveAsArticleTemplateRequest true "request"
// @Success 200 {object} dtoarticle.SaveAsArticleTemplateResponse "response"
func SaveAsArticleTemplateDeprecated(ctx *gin.Context, req *dtoarticle.SaveAsArticleTemplateRequest, resp *dtoarticle.SaveAsArticleTemplateResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[SaveAsArticleTemplate] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcarticle.SaveAsArticleTemplate(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[SaveAsArticleTemplate] svcarticle.SaveAsArticleTemplate failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_article_template_save_as_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// DeleteArticleTemplate 删除写作模板
// Deprecated: 文章模板接口已合并到文章统一接口，请使用对应的 Article 接口 + type 参数
// @Tags 写作空间
// @Summary 删除写作模板
// @Description 删除写作模板
// @Router /forest.DeleteArticleTemplate [post]
// @Param request body dtoarticle.DeleteArticleTemplateRequest true "request"
// @Success 200 {object} dtoarticle.DeleteArticleTemplateResponse "response"
func DeleteArticleTemplate(ctx *gin.Context, req *dtoarticle.DeleteArticleTemplateRequest, resp *dtoarticle.DeleteArticleTemplateResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[DeleteArticleTemplate] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	res, err := svcarticle.DeleteArticleTemplate(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[DeleteArticleTemplate] svcarticle.DeleteArticleTemplate failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_article_template_delete_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ModifyArticleTemplate 修改写作模板
// Deprecated: 文章模板接口已合并到文章统一接口，请使用对应的 Article 接口 + type 参数
// @Tags 写作空间
// @Summary 修改写作模板
// @Description 修改写作模板
// @Router /forest.ModifyArticleTemplate [post]
// @Param request body dtoarticle.ModifyArticleTemplateRequest true "request"
// @Success 200 {object} dtoarticle.ModifyArticleTemplateResponse "response"
func ModifyArticleTemplate(ctx *gin.Context, req *dtoarticle.ModifyArticleTemplateRequest, resp *dtoarticle.ModifyArticleTemplateResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ModifyArticleTemplate] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	res, err := svcarticle.ModifyArticleTemplate(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ModifyArticleTemplate] svcarticle.ModifyArticleTemplate failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_article_template_modify_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// CreateArticleTemplate 创建写作模板
// Deprecated: 文章模板接口已合并到文章统一接口，请使用对应的 Article 接口 + type 参数
// @Tags 写作空间
// @Summary 创建写作模板
// @Description 创建写作模板
// @Router /forest.CreateArticleTemplate [post]
// @Param request body dtoarticle.CreateArticleTemplateRequest true "request"
// @Success 200 {object} dtoarticle.CreateArticleTemplateResponse "response"
func CreateArticleTemplate(ctx *gin.Context, req *dtoarticle.CreateArticleTemplateRequest, resp *dtoarticle.CreateArticleTemplateResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreateArticleTemplate] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcarticle.CreateArticleTemplate(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateArticleTemplate] svcarticle.CreateArticleTemplate failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_article_template_create_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ListArticleTemplate 查看模板列表
// Deprecated: 文章模板接口已合并到文章统一接口，请使用对应的 Article 接口 + type 参数
// @Tags 写作空间
// @Summary 查看模板列表
// @Description 查看模板列表
// @Router /forest.ListTemplate [post]
// @Param request body dtoarticle.ListArticleTemplateRequest true "request"
// @Success 200 {object} dtoarticle.ListArticleTemplateResponse "response"
func ListArticleTemplate(ctx *gin.Context, req *dtoarticle.ListArticleTemplateRequest, resp *dtoarticle.ListArticleTemplateResponse) {
	res, err := svcarticle.ListArticleTemplate(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListArticleTemplate] svcarticle.ListArticleTemplate failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_article_template_list_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetArticleTemplate 查看模板内容
// Deprecated: 文章模板接口已合并到文章统一接口，请使用对应的 Article 接口 + type 参数
// @Tags 写作空间
// @Summary 查看模板内容
// @Description 查看模板内容
// @Router /forest.GetArticleTemplate [post]
// @Param request body dtoarticle.GetArticleTemplateDetailRequest true "request"
// @Success 200 {object} dtoarticle.GetArticleTemplateDetailResponse "response"
func GetArticleTemplate(ctx *gin.Context, req *dtoarticle.GetArticleTemplateDetailRequest, resp *dtoarticle.GetArticleTemplateDetailResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetArticleTemplate] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	res, err := svcarticle.GetArticleTemplateDetail(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetArticleTemplate] svcarticle.GetArticleTemplateDetail failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_article_template_get_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
