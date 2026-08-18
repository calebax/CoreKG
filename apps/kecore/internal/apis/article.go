package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoarticle"
	"github.com/insmtx/corekg/apps/kecore/services/svcarticle"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// DuplicateArticle 复制文章
// @Tags 写作空间
// @Summary 复制文章
// @Description 复制文章
// @Router /forest.DuplicateArticle [post]
// @Param request body dtoarticle.DuplicateArticleRequest true "request"
// @Success 200 {object} dtoarticle.DuplicateArticleResponse "response"
func DuplicateArticle(ctx *gin.Context, req *dtoarticle.DuplicateArticleRequest, resp *dtoarticle.DuplicateArticleResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[DuplicateArticle] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcarticle.DuplicateArticle(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[DuplicateArticle] svcarticle.DuplicateArticle failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_duplicate_article_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ExecuteAIWriteCmd AI写作指令执行接口
// @Tags 写作空间
// @Summary AI写作指令执行接口
// @Description AI写作指令执行接口
// @Router /forest.ExecuteAIWriteCmd [post]
// @Param request body dtoarticle.ExecuteAIWriteCmdRequest true "request"
// @Success 200 {object} dtoarticle.ExecuteAIWriteCmdResponse "response"
func ExecuteAIWriteCmd(ctx *gin.Context, req *dtoarticle.ExecuteAIWriteCmdRequest, resp *dtoarticle.ExecuteAIWriteCmdResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ExecuteAIWriteCmd] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	res, err := svcarticle.ExecuteAIWriteCmd(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ExecuteAIWriteCmd] svcarticle.ExecuteAIWriteCmd failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_execute_ai_write_cmd_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ListArticle 获取文章列表
// @Tags 写作空间
// @Summary 获取文章列表
// @Description 获取文章列表
// @Router /forest.ListArticle [post]
// @Param request body dtoarticle.ListArticleRequest true "request"
// @Success 200 {object} dtoarticle.ListArticleResponse "response"
func ListArticle(ctx *gin.Context, req *dtoarticle.ListArticleRequest, resp *dtoarticle.ListArticleResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListArticle] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	res, err := svcarticle.ListArticle(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListArticle] svcarticle.ListArticle failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_article_get_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// CreateArticle 创建文章
// @Tags 写作空间
// @Summary 创建文章
// @Description 创建文章
// @Router /forest.CreateArticle [post]
// @Param request body dtoarticle.CreateArticleRequest true "request"
// @Success 200 {object} dtoarticle.CreateArticleResponse "response"
func CreateArticle(ctx *gin.Context, req *dtoarticle.CreateArticleRequest, resp *dtoarticle.CreateArticleResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreateArticle] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	res, err := svcarticle.CreateArticle(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateArticle] svcarticle.CreateArticle failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_article_create_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// EditArticle 编辑文章
// @Tags 写作空间
// @Summary 编辑文章
// @Description 编辑文章
// @Router /forest.EditArticle [post]
// @Param request body dtoarticle.EditArticleRequest true "request"
// @Success 200 {object} dtoarticle.EditArticleResponse "response"
func EditArticle(ctx *gin.Context, req *dtoarticle.EditArticleRequest, resp *dtoarticle.EditArticleResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[EditArticle] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	res, err := svcarticle.EditArticle(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[EditArticle] svcarticle.EditArticle failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_article_update_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// DeleteArticle 删除文章
// @Tags 写作空间
// @Summary 删除文章
// @Description 删除文章
// @Router /forest.DeleteArticle [post]
// @Param request body dtoarticle.DeleteArticleRequest true "request"
// @Success 200 {object} dtoarticle.DeleteArticleResponse "response"
func DeleteArticle(ctx *gin.Context, req *dtoarticle.DeleteArticleRequest, resp *dtoarticle.DeleteArticleResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[DeleteArticle] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	res, err := svcarticle.DeleteArticle(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[DeleteArticle] svcarticle.DeleteArticle failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_article_delete_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetArticle 获取文章详情
// @Tags 写作空间
// @Summary 获取文章详情
// @Description 获取文章详情
// @Router /forest.GetArticle [post]
// @Param request body dtoarticle.GetArticleRequest true "request"
// @Success 200 {object} dtoarticle.GetArticleResponse "response"
func GetArticle(ctx *gin.Context, req *dtoarticle.GetArticleRequest, resp *dtoarticle.GetArticleResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetArticle] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	res, err := svcarticle.GetArticle(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetArticle] svcarticle.GetArticle failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_article_get_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// SaveArticleContent 保存文章内容
// @Tags 写作空间
// @Summary 保存文章内容
// @Description 保存文章内容
// @Router /forest.SaveArticleContent [post]
// @Param request body dtoarticle.SaveArticleContentRequest true "request"
// @Success 200 {object} dtoarticle.SaveArticleContentResponse "response"
func SaveArticleContent(ctx *gin.Context, req *dtoarticle.SaveArticleContentRequest, resp *dtoarticle.SaveArticleContentResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[SaveArticleContent] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	res, err := svcarticle.SaveArticleContent(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[SaveArticleContent] svcarticle.SaveArticleContent failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_article_save_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// SaveAsArticleTemplate 保存为写作模板
// @Tags 写作空间
// @Summary 保存为写作模板
// @Description 保存为写作模板（基于文章内容创建 type=template_user 的记录）
// @Router /forest.SaveAsArticleTemplate [post]
// @Param request body dtoarticle.SaveAsTemplateRequest true "request"
// @Success 200 {object} dtoarticle.SaveAsTemplateResponse "response"
func SaveAsArticleTemplate(ctx *gin.Context, req *dtoarticle.SaveAsTemplateRequest, resp *dtoarticle.SaveAsTemplateResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[SaveAsArticleTemplate] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	res, err := svcarticle.SaveAsTemplate(ctx, &req.Request)
	if err != nil {
		logs.ErrorContextf(ctx, "[SaveAsArticleTemplate] svcarticle.SaveAsTemplate failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_save_as_template_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
