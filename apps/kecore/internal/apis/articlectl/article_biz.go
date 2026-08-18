package articlectl

import (
	"github.com/insmtx/corekg/apps/kecore/models/article"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/types"
)

type ListArticleRequest struct {
	apiobj.BaseRequest
	Request apiobj.PageQuery
}

type ListArticleResponse struct {
	apiobj.BaseResponse
	Response article.ListArticleResponse
}

func (in ListArticleRequest) Validity(out *apiobj.BaseResponse) {
	if in.Request.Offset < 0 || in.Request.Limit < 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "kechat_offset_limit_invalid"
		return
	}
	for _, v := range in.Request.OrderBy {
		switch v {
		case "created_at", "updated_at", "created_at desc", "updated_at desc", "title", "title desc":
		default:
			out.Code = errcode.ErrCode_BadRequest
			out.Message = "kecore_order_by_empty"
			return
		}
	}
	for _, v := range in.Request.Filters {
		switch v.Field {
		case "title":
			if len(v.Value) != 1 {
				out.Code = errcode.ErrCode_BadRequest
				out.Message = "kecore_filter_field_single_value"
				return
			}
			if v.Value[0] == "" {
				out.Code = errcode.ErrCode_BadRequest
				out.Message = "kecore_filter_field_empty_value"
				return
			}

		default:
			out.Code = errcode.ErrCode_BadRequest
			out.Message = "kecore_invalid_filter_field_data" // 查询条件中的字段不存在, {{.field}}
			out.MessageData = map[string]interface{}{
				"field": v.Field,
			}
			return
		}
	}
}

type CreateArticleRequest struct {
	apiobj.BaseRequest
	Request struct {
		Title      string          `json:"title"`
		AvatarUrl  string          `json:"avatar_url"`
		TemplateID uint            `json:"template_id"`
		ForestIDs  types.UintArray `json:"forest_ids"`
		//权限
		PublicScope foresttype.PublicScope `json:"public_scope"`
		ScopeIDs    types.UintArray        `json:"scope_ids"`
		ManagerIDs  types.UintArray        `json:"manager_ids"`
	}
}

type CreateArticleResponse struct {
	apiobj.BaseResponse
	Response struct {
		ID uint `json:"id"`
	}
}

func (in CreateArticleRequest) Validity(out *apiobj.BaseResponse) {
	if in.Request.TemplateID < 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "kecore_article_template_invalid_id"
		return
	}
	if len(in.Request.Title) <= 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "kecore_article_title_empty"
		return
	}
	switch in.Request.PublicScope {
	case foresttype.PublicScopeCompany, foresttype.PublicScopeCustom:
	default:
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "kecore_resource_scope_invalid"
	}
}

type EditArticleRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID         uint            `json:"id"`
		Title      string          `json:"title"`
		TemplateID uint            `json:"template_id"`
		ForestIDs  types.UintArray `json:"forest_ids"`
		//权限
		PublicScope foresttype.PublicScope `json:"public_scope"`
		ScopeIDs    types.UintArray        `json:"scope_ids"`
		ManagerIDs  types.UintArray        `json:"manager_ids"`
	}
}

func (in EditArticleRequest) Validity(out *apiobj.BaseResponse) {
	if in.Request.ID < 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "kecore_article_invalid_id"
		return
	}
	if in.Request.TemplateID < 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "kecore_article_template_invalid_id"
		return
	}
	if len(in.Request.Title) <= 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "kecore_article_title_empty"
		return
	}
	switch in.Request.PublicScope {
	case foresttype.PublicScopeCompany, foresttype.PublicScopeCustom:
	default:
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "kecore_resource_scope_invalid"
	}
}

type DeleteArticleRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID uint `json:"id"`
	}
}

func (in DeleteArticleRequest) Validity(out *apiobj.BaseResponse) {
	if in.Request.ID < 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "kecore_article_invalid_id"
		return
	}
}

type GetArticleRequest apiobj.DetailIdRequest
type GetArticleResponse struct {
	apiobj.BaseResponse
	Response struct {
		*foresttype.KeArticle
	}
}

func (in GetArticleRequest) Validity(out *apiobj.BaseResponse) {
	if in.Request.ID < 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "kecore_article_invalid_id"
		return
	}
}

type SaveArticleContentRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID      uint   `json:"id"`
		Content string `json:"content"`
		// ForestIDs 文章关联的知识库 ID 列表
		ForestIDs types.UintArray `json:"forest_ids"`
	}
}

func (in SaveArticleContentRequest) Validity(_ *apiobj.BaseResponse) {
	return
}

type ListArticleTemplateResponse struct {
	apiobj.BaseResponse
	Response struct {
		Data []*foresttype.KeArticleTemplate `json:"data"`
	}
}

type GetArticleTemplateRequest apiobj.DetailIdRequest

type GetArticleTemplateResponse struct {
	apiobj.BaseResponse
	Response struct {
		*foresttype.KeArticleTemplate
	}
}

func (in GetArticleTemplateRequest) Validity(out *GetArticleTemplateResponse) {
	if in.Request.ID < 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "kecore_article_template_invalid_id"
		return
	}
}

type AIWriteRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID        uint                 `json:"id"`
		Cmd       foresttype.CmdString `json:"cmd"`
		Content   string               `json:"content"`
		ForestIDs types.UintArray      `json:"forest_ids"`
	}
}

func (in AIWriteRequest) Validity(out *apiobj.BaseResponse) {
	if in.Request.ID < 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "kecore_article_invalid_id"
		return
	}
	if len(in.Request.Cmd) <= 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "kecore_article_AIWrite_cmd_empty"
		return
	}

	if len(in.Request.Content) <= 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "kecore_article_AIWrite_content_empty"
		return
	}
}
