package dtoarticle

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

// Deprecated: 文章模板已合并到文章接口，通过 type 参数区分
type SaveAsArticleTemplateResponse struct {
	apiobj.BaseResponse
	Response SaveAsArticleTemplateEmbedResponse `json:"response"`
}

// Deprecated: 文章模板已合并到文章接口，通过 type 参数区分
type SaveAsArticleTemplateEmbedResponse struct {
	// ArticleTemplateID 模板ID
	ArticleTemplateID uint `json:"article_template_id"`
}

// Deprecated: 文章模板已合并到文章接口，通过 type 参数区分
type DeleteArticleTemplateResponse struct {
	apiobj.BaseResponse
	Response DeleteArticleTemplateEmbedResponse `json:"response"`
}

// Deprecated: 文章模板已合并到文章接口，通过 type 参数区分
type DeleteArticleTemplateEmbedResponse struct{}

// Deprecated: 文章模板已合并到文章接口，通过 type 参数区分
type ModifyArticleTemplateResponse struct {
	apiobj.BaseResponse
	Response ModifyArticleTemplateEmbedResponse `json:"response"`
}

// Deprecated: 文章模板已合并到文章接口，通过 type 参数区分
type ModifyArticleTemplateEmbedResponse struct {
	// ArticleTemplateID 模板ID
	ArticleTemplateID uint `json:"article_template_id"`
}

// Deprecated: 文章模板已合并到文章接口，通过 type 参数区分
type CreateArticleTemplateResponse struct {
	apiobj.BaseResponse
	Response CreateArticleTemplateEmbedResponse `json:"response"`
}

// Deprecated: 文章模板已合并到文章接口，通过 type 参数区分
type CreateArticleTemplateEmbedResponse struct {
	// ArticleTemplateID 模板ID
	ArticleTemplateID uint `json:"article_template_id"`
}

type DuplicateArticleResponse struct {
	apiobj.BaseResponse
	Response DuplicateArticleEmbedResponse `json:"response"`
}

type DuplicateArticleEmbedResponse struct {
	// ArticleID 文章ID
	ArticleID uint `json:"article_id"`
}

// Deprecated: 文章模板已合并到文章接口，通过 type 参数区分
type ListArticleTemplateResponse struct {
	apiobj.BaseResponse
	Response ListArticleTemplateEmbedResponse `json:"response"`
}

// Deprecated: 文章模板已合并到文章接口，通过 type 参数区分
type ListArticleTemplateEmbedResponse struct {
	apiobj.QueryResponse
	Data []*foresttype.KeArticleTemplate `json:"data"`
}

type ExecuteAIWriteCmdResponse struct {
	apiobj.BaseResponse
	Response ExecuteAIWriteCmdEmbedResponse `json:"response"`
}

type ExecuteAIWriteCmdEmbedResponse struct {
	// Result 执行结果
	Result string `json:"result"`
}

type ArticleInfoItem struct {
	*foresttype.KeArticle
	// IsAdmin 是否为管理员
	IsAdmin bool `json:"isAdmin"`
}

type ListArticleResponse struct {
	apiobj.BaseResponse
	// Response 文章列表的返回数据
	Response ListArticleEmbedResponse `json:"response"`
}

type ListArticleEmbedResponse struct {
	apiobj.QueryResponse
	// Data 文章列表数据
	Data []ArticleInfoItem `json:"data"`
}

type CreateArticleResponse struct {
	apiobj.BaseResponse
	// Response 创建文章的返回数据
	Response CreateArticleEmbedResponse `json:"response"`
}

type CreateArticleEmbedResponse struct {
	// ID 文章ID
	ID uint `json:"id"`
}

type EditArticleResponse struct {
	apiobj.BaseResponse
	// Response 编辑文章的返回数据
	Response EditArticleEmbedResponse `json:"response"`
}

type EditArticleEmbedResponse struct{}

type DeleteArticleResponse struct {
	apiobj.BaseResponse
	// Response 删除文章的返回数据
	Response DeleteArticleEmbedResponse `json:"response"`
}

type DeleteArticleEmbedResponse struct{}

type GetArticleResponse struct {
	apiobj.BaseResponse
	// Response 获取文章详情的返回数据
	Response GetArticleEmbedResponse `json:"response"`
}

type ArticleTemplateItem struct {
	// ArticleTemplateID 模板ID
	ArticleTemplateID uint `json:"article_template_id"`
	// ArticleTemplateName 模板名称
	ArticleTemplateName string `json:"article_template_name"`
}

type GetArticleEmbedResponse struct {
	*foresttype.KeArticle
	// ArticleTemplates 文章关联的模板列表
	ArticleTemplates []ArticleTemplateItem `json:"article_templates"`
}

type SaveArticleContentResponse struct {
	apiobj.BaseResponse
	// Response 保存文章内容的返回数据
	Response SaveArticleContentEmbedResponse `json:"response"`
}

type SaveArticleContentEmbedResponse struct{}

// Deprecated: 文章模板已合并到文章接口，通过 type 参数区分
type GetArticleTemplateDetailResponse struct {
	apiobj.BaseResponse
	// Response 获取模板详情的返回数据
	Response GetArticleTemplateDetailEmbedResponse `json:"response"`
}

// Deprecated: 文章模板已合并到文章接口，通过 type 参数区分
type GetArticleTemplateDetailEmbedResponse struct {
	*foresttype.KeArticleTemplate
}

type SaveAsTemplateResponse struct {
	apiobj.BaseResponse
	Response SaveAsTemplateEmbedResponse `json:"response"`
}

type SaveAsTemplateEmbedResponse struct {
	// ArticleID 文章ID
	ArticleID uint `json:"article_id"`
}
