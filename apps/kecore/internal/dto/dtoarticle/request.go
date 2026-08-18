package dtoarticle

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/types"
)

type SaveAsArticleTemplateRequest struct {
	apiobj.BaseRequest
	Request SaveAsArticleTemplateEmbedRequest `json:"request"`
}

type SaveAsArticleTemplateEmbedRequest struct {
	// ArticleID 文章ID
	ArticleID uint `json:"article_id" validate:"required"`
}

func (opt *SaveAsArticleTemplateRequest) Validity(resp *SaveAsArticleTemplateResponse) {
	if opt.Request.ArticleID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_id_required"
		return
	}
}

type DeleteArticleTemplateRequest struct {
	apiobj.BaseRequest
	Request DeleteArticleTemplateEmbedRequest `json:"request"`
}

type DeleteArticleTemplateEmbedRequest struct {
	// ArticleTemplateID 模板ID
	ArticleTemplateID uint `json:"article_template_id" validate:"required"`
}

func (opt *DeleteArticleTemplateRequest) Validity(resp *DeleteArticleTemplateResponse) {
	if opt.Request.ArticleTemplateID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_template_id_required"
		return
	}
}

type ModifyArticleTemplateRequest struct {
	apiobj.BaseRequest
	Request ModifyArticleTemplateEmbedRequest `json:"request"`
}

type ModifyArticleTemplateEmbedRequest struct {
	// ArticleTemplateID 模板ID
	ArticleTemplateID uint `json:"article_template_id" validate:"required"`
	// Name 模板名称
	Name string `json:"name"`
	// Description 模板描述
	Description string `json:"description"`
	// Content 模板内容
	Content string `json:"content"`
}

func (opt *ModifyArticleTemplateRequest) Validity(resp *ModifyArticleTemplateResponse) {
	if opt.Request.ArticleTemplateID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_template_id_required"
		return
	}
}

type CreateArticleTemplateRequest struct {
	apiobj.BaseRequest
	Request CreateArticleTemplateEmbedRequest `json:"request"`
}

type CreateArticleTemplateEmbedRequest struct {
	// Name 模板名称
	Name string `json:"name" validate:"required"`
	// Description 模板描述
	Description string `json:"description"`
	// Content 模板内容
	Content string `json:"content"`
}

func (opt *CreateArticleTemplateRequest) Validity(resp *CreateArticleTemplateResponse) {
	if opt.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_template_name_required"
		return
	}
}

type ListArticleTemplateRequest struct {
	apiobj.BaseRequest
	Request ListArticleTemplateEmbedRequest `json:"request"`
}

type ListArticleTemplateEmbedRequest struct {
	apiobj.PageQuery
	// TemplateType 模板类型
	TemplateType foresttype.TemplateType `json:"template_type"`
	// SourceType 来源类型
	SourceType foresttype.SourceType `json:"source_type"`
}

func (opt *ListArticleTemplateRequest) Validity(resp *ListArticleTemplateResponse) {
}

type DuplicateArticleRequest struct {
	apiobj.BaseRequest
	Request DuplicateArticleEmbedRequest `json:"request"`
}

type DuplicateArticleEmbedRequest struct {
	// ArticleID 文章ID
	ArticleID uint `json:"article_id" validate:"required"`
}

func (opt *DuplicateArticleRequest) Validity(resp *DuplicateArticleResponse) {
	if opt.Request.ArticleID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_id_required"
		return
	}
}

type ExecuteAIWriteCmdRequest struct {
	apiobj.BaseRequest
	Request ExecuteAIWriteCmdEmbedRequest `json:"request"`
}

type ExecuteAIWriteCmdEmbedRequest struct {
	// ArticleID 文章ID
	ArticleID uint `json:"article_id" validate:"required"`
	// Cmd 写作指令类型，如缩写abbreviation、扩写expansion、润色embellishment、校阅proofreading、续写continuation
	Cmd foresttype.CmdString `json:"cmd" validate:"required"`
	// Content 需要处理的原始内容
	Content string `json:"content" validate:"required"`
	// ForestIDs 关联的知识库ID列表，用于获取参考素材
	ForestIDs types.UintArray `json:"forest_ids"`
}

func (opt *ExecuteAIWriteCmdRequest) Validity(resp *ExecuteAIWriteCmdResponse) {
	if opt.Request.ArticleID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_invalid_id"
		return
	}
	if len(opt.Request.Cmd) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_AIWrite_cmd_empty"
		return
	}
	if len(opt.Request.Content) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_AIWrite_content_empty"
		return
	}
}

type ListArticleRequest struct {
	apiobj.BaseRequest
	Request ListArticleEmbedRequest `json:"request"`
}

type ListArticleEmbedRequest struct {
	apiobj.PageQuery
	ArticleTypes []foresttype.ArticleType `json:"article_types" form:"article_types"`
}

func (opt *ListArticleRequest) Validity(resp *ListArticleResponse) {
	if opt.Request.Offset < 0 || opt.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_offset_limit_invalid"
		return
	}
	for _, v := range opt.Request.OrderBy {
		switch v {
		case "created_at", "updated_at", "created_at desc", "updated_at desc", "title", "title desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_order_by_empty"
			return
		}
	}
	for _, t := range opt.Request.ArticleTypes {
		switch t {
		case foresttype.ArticleTypeArticle, foresttype.ArticleTypeTemplateSystem, foresttype.ArticleTypeTemplateUser:
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_invalid_article_type"
			return
		}
	}
	for _, v := range opt.Request.Filters {
		switch v.Field {
		case "title":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_field_single_value"
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_field_empty_value"
				return
			}
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_invalid_filter_field_data"
			return
		}
	}
}

type CreateArticleRequest struct {
	apiobj.BaseRequest
	Request CreateArticleEmbedRequest `json:"request"`
}

type CreateArticleEmbedRequest struct {
	// ArticleType 文章类型
	ArticleType foresttype.ArticleType `json:"article_type" form:"article_type"`
	// Title 文章标题
	Title string `json:"title" validate:"required"`
	// Description 文章描述
	Description string `json:"description"`
	// AvatarUrl 文章封面图URL
	AvatarUrl     string                `json:"avatar_url"`
	// SourceType 记录来源类型
	SourceType foresttype.SourceType `json:"source_type" form:"source_type"`
	// SourceID 来源资源ID（当 source_type=template 时为模板ID）
	SourceID uint `json:"source_id" form:"source_id"`
	// ForestIDs 关联知识库ID列表
	ForestIDs types.UintArray `json:"forest_ids"`
	// PublicScope 公开范围
	PublicScope foresttype.PublicScope `json:"public_scope"`
	// ScopeIDs 可见范围ID列表
	ScopeIDs types.UintArray `json:"scope_ids"`
	// ManagerIDs 管理员ID列表
	ManagerIDs types.UintArray `json:"manager_ids"`
}

func (opt *CreateArticleRequest) Validity(resp *CreateArticleResponse) {
	if len(opt.Request.Title) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_title_empty"
		return
	}
	switch opt.Request.PublicScope {
	case foresttype.PublicScopeCompany, foresttype.PublicScopeCustom:
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_resource_scope_invalid"
	}
}

type EditArticleRequest struct {
	apiobj.BaseRequest
	Request EditArticleEmbedRequest `json:"request"`
}

type EditArticleEmbedRequest struct {
	// ID 文章ID
	ID uint `json:"id" validate:"required"`
	// ArticleType 文章类型
	ArticleType foresttype.ArticleType `json:"article_type" form:"article_type"`
	// Title 文章标题
	Title string `json:"title"`
	// Description 文章描述
	Description string                `json:"description"`
	Content     string                `json:"content"`
	SourceType  foresttype.SourceType `json:"source_type" form:"source_type"`
	SourceID    uint                  `json:"source_id" form:"source_id"`
	// ForestIDs 关联知识库ID列表
	ForestIDs types.UintArray `json:"forest_ids"`
	// PublicScope 公开范围
	PublicScope foresttype.PublicScope `json:"public_scope"`
	// ScopeIDs 可见范围ID列表
	ScopeIDs types.UintArray `json:"scope_ids"`
	// ManagerIDs 管理员ID列表
	ManagerIDs types.UintArray `json:"manager_ids"`
}

func (opt *EditArticleRequest) Validity(resp *EditArticleResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_invalid_id"
		return
	}
	if len(opt.Request.Title) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_title_empty"
		return
	}
	switch opt.Request.PublicScope {
	case foresttype.PublicScopeCompany, foresttype.PublicScopeCustom:
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_resource_scope_invalid"
	}
}

type DeleteArticleRequest struct {
	apiobj.BaseRequest
	Request DeleteArticleEmbedRequest `json:"request"`
}

type DeleteArticleEmbedRequest struct {
	// ID 文章ID
	ID uint `json:"id" validate:"required"`
}

func (opt *DeleteArticleRequest) Validity(resp *DeleteArticleResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_invalid_id"
		return
	}
}

type GetArticleRequest struct {
	apiobj.BaseRequest
	Request GetArticleEmbedRequest `json:"request"`
}

type GetArticleEmbedRequest struct {
	// ID 文章ID
	ID uint `json:"id" validate:"required"`
}

func (opt *GetArticleRequest) Validity(resp *GetArticleResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_invalid_id"
		return
	}
}

type SaveArticleContentRequest struct {
	apiobj.BaseRequest
	Request SaveArticleContentEmbedRequest `json:"request"`
}

type SaveArticleContentEmbedRequest struct {
	// ID 文章ID
	ID uint `json:"id" validate:"required"`
	// Content 文章内容
	Content string `json:"content"`
	// ForestIDs 关联知识库ID列表
	ForestIDs types.UintArray `json:"forest_ids"`
}

func (opt *SaveArticleContentRequest) Validity(resp *SaveArticleContentResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_invalid_id"
		return
	}
	if len(opt.Request.Content) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_content_empty"
		return
	}
}

type GetArticleTemplateDetailRequest struct {
	apiobj.BaseRequest
	Request GetArticleTemplateDetailEmbedRequest `json:"request"`
}

type GetArticleTemplateDetailEmbedRequest struct {
	// ID 模板ID
	ID uint `json:"id" validate:"required"`
}

func (opt *GetArticleTemplateDetailRequest) Validity(resp *GetArticleTemplateDetailResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_template_invalid_id"
		return
	}
}

type SaveAsTemplateRequest struct {
	apiobj.BaseRequest
	Request SaveAsTemplateEmbedRequest `json:"request"`
}

type SaveAsTemplateEmbedRequest struct {
	// ArticleID 文章ID
	ArticleID uint `json:"article_id" form:"article_id" binding:"required"`
}

func (req *SaveAsTemplateRequest) Validity(resp *SaveAsTemplateResponse) {
	if req.Request.ArticleID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_article_id_required"
		return
	}
}
