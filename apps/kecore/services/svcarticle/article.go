package svcarticle

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoarticle"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/apps/kecore/services/corearticle"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

func DuplicateArticle(ctx *gin.Context, req *dtoarticle.DuplicateArticleRequest) (res *dtoarticle.DuplicateArticleResponse, err error) {
	res = &dtoarticle.DuplicateArticleResponse{}
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
	duplicateEntity := &foresttype.KeArticle{
		CompanyID:   companyID,
		Uin:         uin,
		Type:        articleEntity.Type,
		Title:       articleEntity.Title + "_副本",
		Description: articleEntity.Description,
		Content:     articleEntity.Content,
		SourceType:  foresttype.SourceTypeArticle,
		SourceID:    articleEntity.ID,
		ForestIDs:   articleEntity.ForestIDs,
		PublicScope: articleEntity.PublicScope,
		AvatarUrl:   articleEntity.AvatarUrl,
	}
	if err := forest.NewArticleDao().Insert(ctx, duplicateEntity); err != nil {
		return nil, err
	}
	res.Response.ArticleID = duplicateEntity.ID
	return res, nil
}

func ExecuteAIWriteCmd(ctx *gin.Context, req *dtoarticle.ExecuteAIWriteCmdRequest) (res *dtoarticle.ExecuteAIWriteCmdResponse, err error) {
	res = &dtoarticle.ExecuteAIWriteCmdResponse{}

	chatModelEntity, err := chat.NewChatModelDao().GetByCond(ctx, &chat.ChatModelCond{
		BaseCond: chat.BaseCond{
			OrderBy: []string{"priority desc"},
		},
		PublicType: chattype.PublecTypeSystem,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "ExecuteAIWriteCmd GetByCond err: %v", err)
		return nil, err
	}

	executor := corearticle.NewAIWriteExecutor(corearticle.AIWriteExecutorParams{
		Ctx:         ctx,
		GinCtx:      ctx,
		ArticleID:   req.Request.ArticleID,
		Cmd:         req.Request.Cmd,
		Content:     req.Request.Content,
		ForestIDs:   req.Request.ForestIDs,
		RequestID:   runtime.RequestID(ctx),
		CompanyID:   runtime.CompanyID(ctx),
		Uin:         runtime.Uin(ctx),
		ChatModelID: chatModelEntity.ID,
	})

	if err = executor.Execute(ctx.Writer); err != nil {
		logs.ErrorContextf(ctx, "ExecuteAIWriteCmd failed: %v", err)
		return nil, err
	}

	return res, nil
}

func ListArticle(ctx *gin.Context, req *dtoarticle.ListArticleRequest) (res *dtoarticle.ListArticleResponse, err error) {
	res = &dtoarticle.ListArticleResponse{}
	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)

	articleCond := &forest.ArticleCond{
		BaseCond: forest.BaseCond{
			OrderBy: req.Request.OrderBy,
			Offset:  req.Request.Offset,
			Limit:   req.Request.Limit,
		},
		Filters: req.Request.Filters,
	}
	tableName := foresttype.TableNameKeArticle
	if len(req.Request.ArticleTypes) > 0 {
		hasTemplateSystem := false
		otherTypes := make([]foresttype.ArticleType, 0, len(req.Request.ArticleTypes))
		for _, t := range req.Request.ArticleTypes {
			if t == foresttype.ArticleTypeTemplateSystem {
				hasTemplateSystem = true
			} else {
				otherTypes = append(otherTypes, t)
			}
		}

		if hasTemplateSystem {
			articleCond.BaseCond.OrCondition = forest.OrCondition{
				Conditions: []string{
					fmt.Sprintf("(%s.type IN ? AND %s.company_id = ? AND %s.uin = ?)", tableName, tableName, tableName),
					fmt.Sprintf("%s.type = ?", tableName),
				},
				Args: []any{
					otherTypes,
					companyID,
					uin,
					foresttype.ArticleTypeTemplateSystem,
				},
			}
		} else {
			articleCond.BaseCond.CompanyID = companyID
			articleCond.BaseCond.Uin = uin
			articleCond.ArticleTypes = req.Request.ArticleTypes
		}
	} else {
		articleCond.BaseCond.CompanyID = companyID
		articleCond.BaseCond.Uin = uin
		articleCond.ArticleTypes = []foresttype.ArticleType{foresttype.ArticleTypeArticle}
	}

	articleList, total, err := forest.NewArticleDao().GetPageListByCond(ctx, articleCond)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListArticle] GetPageListByCond failed, err: %v", err)
		return nil, err
	}

	res.Response.Total = total
	res.Response.Offset = req.Request.Offset
	res.Response.Limit = req.Request.Limit

	manageIDs := perm.GetManageList(ctx, uin, foresttype.ResourceTypeArticle)
	manageMap := make(map[uint]bool, len(manageIDs))
	for _, id := range manageIDs {
		manageMap[id] = true
	}

	data := make([]dtoarticle.ArticleInfoItem, 0, len(articleList))
	for _, art := range articleList {
		data = append(data, dtoarticle.ArticleInfoItem{
			KeArticle: &art,
			IsAdmin:   manageMap[art.ID],
		})
	}
	res.Response.Data = data
	return res, nil
}

func CreateArticle(ctx *gin.Context, req *dtoarticle.CreateArticleRequest) (res *dtoarticle.CreateArticleResponse, err error) {
	res = &dtoarticle.CreateArticleResponse{}
	companyID := runtime.CompanyID(ctx)
	uin := runtime.Uin(ctx)

	articleType := req.Request.ArticleType
	if articleType == "" {
		articleType = foresttype.ArticleTypeArticle
	}

	content := ""
	if req.Request.SourceType == foresttype.SourceTypeTemplate && req.Request.SourceID > 0 {
		templateArticle, err := forest.NewArticleDao().GetByID(ctx, req.Request.SourceID)
		if err != nil {
			logs.ErrorContextf(ctx, "[CreateArticle] GetByID(%v) failed, err: %v", req.Request.SourceID, err)
			return nil, err
		}
		if templateArticle == nil || templateArticle.ID == 0 {
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_from_article_not_exist"
			return res, nil
		}
		if templateArticle.Type != foresttype.ArticleTypeTemplateSystem && templateArticle.Type != foresttype.ArticleTypeTemplateUser {
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_from_article_not_template"
			return res, nil
		}
		content = templateArticle.Content
	}

	art := &foresttype.KeArticle{
		Type:        articleType,
		Uin:         uin,
		CompanyID:   companyID,
		AvatarUrl:   req.Request.AvatarUrl,
		Title:       req.Request.Title,
		Description: req.Request.Description,
		SourceType:  req.Request.SourceType,
		SourceID:    req.Request.SourceID,
		ForestIDs:   req.Request.ForestIDs,
		PublicScope: req.Request.PublicScope,
		Content:     content,
	}

	if err := forest.NewArticleDao().Insert(ctx, art); err != nil {
		logs.ErrorContextf(ctx, "[CreateArticle] Insert failed, err: %v", err)
		return nil, err
	}

	res.Response.ID = art.ID
	return res, nil
}

func EditArticle(ctx *gin.Context, req *dtoarticle.EditArticleRequest) (res *dtoarticle.EditArticleResponse, err error) {
	res = &dtoarticle.EditArticleResponse{}

	companyID := runtime.CompanyID(ctx)
	uin := runtime.Uin(ctx)

	articleEntity, err := forest.NewArticleDao().GetByID(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "[EditArticle] GetByID(%v) failed, err: %v", req.Request.ID, err)
		return nil, err
	}

	if articleEntity == nil || articleEntity.ID == 0 || articleEntity.CompanyID != companyID {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_not_exist"
		return res, nil
	}

	if articleEntity.Type == foresttype.ArticleTypeTemplateUser && articleEntity.Uin != uin {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_template_only_owner_can_edit"
		return res, nil
	}
	if articleEntity.Type == foresttype.ArticleTypeTemplateSystem {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_template_system_cannot_edit"
		return res, nil
	}

	updateMap := map[string]any{
		"title":       req.Request.Title,
		"description": req.Request.Description,
		"content":     req.Request.Content,
		"forest_ids":  string(req.Request.ForestIDs),
	}

	if len(updateMap) > 0 {
		if err := forest.NewArticleDao().UpdateMap(ctx, req.Request.ID, updateMap); err != nil {
			logs.ErrorContextf(ctx, "[EditArticle] UpdateMap failed, err: %v", err)
			return nil, err
		}
	}

	return res, nil
}

func DeleteArticle(ctx *gin.Context, req *dtoarticle.DeleteArticleRequest) (res *dtoarticle.DeleteArticleResponse, err error) {
	res = &dtoarticle.DeleteArticleResponse{}

	articleEntity, err := forest.NewArticleDao().GetByID(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "[DeleteArticle] GetByID(%v) failed, err: %v", req.Request.ID, err)
		return nil, err
	}
	if articleEntity == nil || articleEntity.ID == 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_not_exist"
		return res, nil
	}

	if articleEntity.Type == foresttype.ArticleTypeTemplateSystem {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_template_system_cannot_delete"
		return res, nil
	}
	if articleEntity.Type == foresttype.ArticleTypeTemplateUser && articleEntity.Uin != runtime.Uin(ctx) {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_template_only_owner_can_delete"
		return res, nil
	}

	if err := forest.NewArticleDao().Delete(ctx, req.Request.ID); err != nil {
		logs.ErrorContextf(ctx, "[DeleteArticle] Delete failed, err: %v", err)
		return nil, err
	}

	return res, nil
}

func GetArticle(ctx *gin.Context, req *dtoarticle.GetArticleRequest) (res *dtoarticle.GetArticleResponse, err error) {
	res = &dtoarticle.GetArticleResponse{}

	articleEntity, err := forest.NewArticleDao().GetByID(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetArticle] GetByID(%v) failed, err: %v", req.Request.ID, err)
		return nil, err
	}
	if articleEntity == nil || articleEntity.ID == 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_not_exist"
		return res, nil
	}

	res.Response.KeArticle = articleEntity

	templateList, err := forest.NewArticleDao().GetListByCond(ctx, &forest.ArticleCond{
		BaseCond: forest.BaseCond{
			CompanyID: articleEntity.CompanyID,
			OrCondition: forest.OrCondition{
				Conditions: []string{
					"ke_article.type = ?",
					"(ke_article.uin = ? AND ke_article.type = ?)",
				},
				Args: []any{
					foresttype.ArticleTypeTemplateSystem,
					articleEntity.Uin,
					foresttype.ArticleTypeTemplateUser,
				},
			},
		},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[GetArticle] GetTemplateList failed, err: %v", err)
		return nil, err
	}

	articleTemplates := make([]dtoarticle.ArticleTemplateItem, 0, len(templateList))
	for _, tmpl := range templateList {
		articleTemplates = append(articleTemplates, dtoarticle.ArticleTemplateItem{
			ArticleTemplateID:   tmpl.ID,
			ArticleTemplateName: tmpl.Title,
		})
	}
	res.Response.ArticleTemplates = articleTemplates

	return res, nil
}

func SaveArticleContent(ctx *gin.Context, req *dtoarticle.SaveArticleContentRequest) (res *dtoarticle.SaveArticleContentResponse, err error) {
	res = &dtoarticle.SaveArticleContentResponse{}

	articleEntity, err := forest.NewArticleDao().GetByID(ctx, req.Request.ID)
	if err != nil {
		return nil, err
	}
	if articleEntity == nil || articleEntity.ID == 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_not_exist"
		return res, nil
	}
	if articleEntity.Type == foresttype.ArticleTypeTemplateSystem {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_template_system_cannot_edit"
		return res, nil
	}

	updateMap := map[string]interface{}{
		"content": req.Request.Content,
	}
	if err := forest.NewArticleDao().UpdateMap(ctx, req.Request.ID, updateMap); err != nil {
		logs.ErrorContextf(ctx, "[SaveArticleContent] UpdateMap failed, err: %v", err)
		return nil, err
	}

	return res, nil
}

func SaveAsTemplate(ctx *gin.Context, req *dtoarticle.SaveAsTemplateEmbedRequest) (res *dtoarticle.SaveAsTemplateResponse, err error) {
	res = &dtoarticle.SaveAsTemplateResponse{}

	companyID, uin := runtime.CompanyID(ctx), runtime.Uin(ctx)

	articleEntity, err := forest.NewArticleDao().GetByID(ctx, req.ArticleID)
	if err != nil {
		return nil, err
	}
	if articleEntity == nil || articleEntity.ID == 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_article_not_exist"
		return res, nil
	}

	template := &foresttype.KeArticle{
		Type:        foresttype.ArticleTypeTemplateUser,
		Title:       articleEntity.Title,
		Description: articleEntity.Description,
		Content:     articleEntity.Content,
		SourceType:  foresttype.SourceTypeArticle,
		SourceID:    articleEntity.ID,
		CompanyID:   companyID,
		Uin:         uin,
		PublicScope: foresttype.PublicScopeCompany,
	}

	if err := forest.NewArticleDao().Insert(ctx, template); err != nil {
		return nil, err
	}

	res.Response.ArticleID = template.ID
	return res, nil
}
