package svcarticle

// Deprecated: 文章模板服务已合并到 svcarticle/article.go，通过 type 参数区分

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoarticle"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

func SaveAsArticleTemplate(ctx *gin.Context, req *dtoarticle.SaveAsArticleTemplateRequest) (res *dtoarticle.SaveAsArticleTemplateResponse, err error) {
	res = &dtoarticle.SaveAsArticleTemplateResponse{}
	articleEntity, err := forest.NewArticleDao().GetByID(ctx, req.Request.ArticleID)
	if err != nil {
		return nil, err
	}
	if articleEntity == nil || articleEntity.ID == 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_not_exist"
		return res, nil
	}
	companyID, uin := runtime.CompanyID(ctx), runtime.Uin(ctx)
	templateInsertEntity := &foresttype.KeArticleTemplate{
		CompanyID:    companyID,
		Uin:          uin,
		Name:         articleEntity.Title,
		Description:  articleEntity.Title,
		Content:      articleEntity.Content,
		TemplateType: foresttype.TemplateTypeUser,
		SourceType:   foresttype.SourceTypeArticle,
		SourceID:     req.Request.ArticleID,
	}
	if err := forest.NewArticleTemplateDao().Insert(ctx, templateInsertEntity); err != nil {
		return nil, err
	}
	res.Response.ArticleTemplateID = templateInsertEntity.ID
	return res, nil
}

func DeleteArticleTemplate(ctx *gin.Context, req *dtoarticle.DeleteArticleTemplateRequest) (res *dtoarticle.DeleteArticleTemplateResponse, err error) {
	res = &dtoarticle.DeleteArticleTemplateResponse{}
	templateEntity, err := forest.NewArticleTemplateDao().GetByID(ctx, req.Request.ArticleTemplateID)
	if err != nil {
		return nil, err
	}
	if templateEntity == nil || templateEntity.ID == 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_template_not_exist"
		return res, nil
	}
	if templateEntity.TemplateType != foresttype.TemplateTypeUser || templateEntity.Uin != runtime.Uin(ctx) {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_template_only_user_type_can_delete"
		return res, nil
	}
	if err := forest.NewArticleTemplateDao().Delete(ctx, req.Request.ArticleTemplateID); err != nil {
		return nil, err
	}
	return res, nil
}

func ModifyArticleTemplate(ctx *gin.Context, req *dtoarticle.ModifyArticleTemplateRequest) (res *dtoarticle.ModifyArticleTemplateResponse, err error) {
	res = &dtoarticle.ModifyArticleTemplateResponse{}
	templateEntity, err := forest.NewArticleTemplateDao().GetByID(ctx, req.Request.ArticleTemplateID)
	if err != nil {
		return nil, err
	}
	if templateEntity == nil || templateEntity.ID == 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_template_not_exist"
		return res, nil
	}
	if templateEntity.TemplateType != foresttype.TemplateTypeUser || templateEntity.Uin != runtime.Uin(ctx) {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_template_only_user_type_can_modify"
		return res, nil
	}
	updateMap := map[string]any{
		"name":        req.Request.Name,
		"description": req.Request.Description,
		"content":     req.Request.Content,
	}
	if err := forest.NewArticleTemplateDao().UpdateMap(ctx, req.Request.ArticleTemplateID, updateMap); err != nil {
		return nil, err
	}
	res.Response.ArticleTemplateID = req.Request.ArticleTemplateID
	return res, nil
}

func CreateArticleTemplate(ctx *gin.Context, req *dtoarticle.CreateArticleTemplateRequest) (res *dtoarticle.CreateArticleTemplateResponse, err error) {
	res = &dtoarticle.CreateArticleTemplateResponse{}
	companyID, uin := runtime.CompanyID(ctx), runtime.Uin(ctx)
	templateInsertEntity := &foresttype.KeArticleTemplate{
		CompanyID:    companyID,
		Uin:          uin,
		Name:         req.Request.Name,
		Description:  req.Request.Description,
		Content:      req.Request.Content,
		TemplateType: foresttype.TemplateTypeUser,
		SourceType:   foresttype.SourceTypeManual,
	}
	if err := forest.NewArticleTemplateDao().Insert(ctx, templateInsertEntity); err != nil {
		return nil, err
	}
	res.Response.ArticleTemplateID = templateInsertEntity.ID
	return res, nil
}

func ListArticleTemplate(ctx *gin.Context, req *dtoarticle.ListArticleTemplateRequest) (res *dtoarticle.ListArticleTemplateResponse, err error) {
	res = &dtoarticle.ListArticleTemplateResponse{}

	companyID, uin := runtime.CompanyID(ctx), runtime.Uin(ctx)

	cond := &forest.ArticleTemplateCond{
		BaseCond: forest.BaseCond{
			Offset:  req.Request.Offset,
			Limit:   req.Request.Limit,
			OrderBy: req.Request.OrderBy,
		},
		TemplateType: req.Request.TemplateType,
		SourceType:   req.Request.SourceType,
	}
	switch req.Request.TemplateType {
	case "":
		tableName := foresttype.TableNameKeArticleTemplate
		cond.BaseCond.OrCondition = forest.OrCondition{
			Conditions: []string{
				fmt.Sprintf("%s.template_type = ?", tableName),
				fmt.Sprintf("(%s.uin = ? AND %s.template_type = ?)", tableName, tableName),
			},
			Args: []any{
				foresttype.TemplateTypeSystem,
				uin,
				foresttype.TemplateTypeUser,
			},
		}
	case foresttype.TemplateTypeUser:
		cond.CompanyID = companyID
		cond.Uin = uin
	}
	templateList, total, err := forest.NewArticleTemplateDao().GetPageListByCond(ctx, cond)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListArticleTemplate] GetPageListByCond failed, err: %v", err)
		return nil, err
	}

	data := make([]*foresttype.KeArticleTemplate, 0, len(templateList))
	for i := range templateList {
		data = append(data, &templateList[i])
	}
	res.Response.Data = data
	res.Response.Total = total
	res.Response.Offset = req.Request.Offset
	res.Response.Limit = req.Request.Limit
	return res, nil
}

func GetArticleTemplateDetail(ctx *gin.Context, req *dtoarticle.GetArticleTemplateDetailRequest) (res *dtoarticle.GetArticleTemplateDetailResponse, err error) {
	res = &dtoarticle.GetArticleTemplateDetailResponse{}

	templateEntity, err := forest.NewArticleTemplateDao().GetByID(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetArticleTemplateDetail] GetByID(%v) failed, err: %v", req.Request.ID, err)
		return nil, err
	}
	if templateEntity == nil || templateEntity.ID == 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_template_not_exist"
		return res, nil
	}

	res.Response.KeArticleTemplate = templateEntity
	return res, nil
}
