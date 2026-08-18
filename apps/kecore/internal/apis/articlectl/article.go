package articlectl

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoarticle"
	"github.com/insmtx/corekg/apps/kecore/models/article"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
)

func ListArticle(ctx *gin.Context, req *ListArticleRequest, resp *ListArticleResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "ListArticle failed: code=%d, msg=%s", resp.Code, resp.Message)
		return
	}
	req.Request.Uin = runtime.Uin(ctx)
	req.Request.CompanyID = runtime.CompanyID(ctx)
	if err := article.ListArticles(ctx, req.Request, &resp.Response); err != nil {
		logs.WarnContextf(ctx, "ListArticle failed: err=%+v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_article_get_failed"))
		return
	}
}

func CreateArticle(ctx *gin.Context, req *CreateArticleRequest, resp *CreateArticleResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "CreateArticle.Validity failed: %s", resp.Message)
		return
	}
	companyID := runtime.CompanyID(ctx)
	uin := runtime.Uin(ctx)

	t, err := article.GetTemplateByID(ctx, req.Request.TemplateID)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateArticle.GetTemplateByID(%v) failed: err=%+v", req.Request.TemplateID, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_article_template_get_failed"))
		return
	}

	art := &foresttype.KeArticle{
		Uin:         uin,
		CompanyID:   companyID,
		AvatarUrl:   req.Request.AvatarUrl,
		Title:       req.Request.Title,
		ForestIDs:   req.Request.ForestIDs,
		PublicScope: req.Request.PublicScope,
		Content:     t.Content,
	}

	if err := dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&art).Error; err != nil {
			logs.ErrorContextf(ctx, "CreateArticle failed: err=%+v", err)
			return err
		}

		//do not append manager action list with admins in company
		if req.Request.PublicScope == foresttype.PublicScopeCompany {
			req.Request.ScopeIDs = types.NewUintArray([]uint{})
		}

		// 去重
		req.Request.ManagerIDs.RemoveDuplicates()
		req.Request.ScopeIDs.RemoveDuplicates()

		uins := types.NewUintArray(append(req.Request.ManagerIDs.Slice(), req.Request.ScopeIDs.Slice()...))
		uins.RemoveDuplicates()

		us := uins.Slice()
		if !employee.CheckUinsValid(ctx, us, companyID) {
			logs.ErrorContextf(ctx, "CheckUinsValid: exist no-local company[%v] uin in uins[%v]", companyID, us)
			runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_invalid_employee_id"))
			return err
		}

		return perm.UpdateResourceScope(ctx, tx, art.ID, foresttype.ResourceTypeArticle, req.Request.ScopeIDs.Slice(), req.Request.ManagerIDs.Slice(), req.Request.PublicScope, companyID)
	}); err != nil {
		logs.ErrorContextf(ctx, "CreateArticle failed: %v", err)
		return
	}
	resp.Response.ID = art.ID
}

func EditArticle(ctx *gin.Context, req *EditArticleRequest, resp *apiobj.BaseResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "EditArticle.Validity failed: %s", resp.Message)
		return
	}

	art, err := article.GetArticleByID(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetArticleByID(%v) err: %v", req.Request.ID, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_article_get_failed"))
		return
	}
	if req.Request.Title != "" {
		art.Title = req.Request.Title
	}
	if len(req.Request.ForestIDs) > 0 {
		art.ForestIDs = req.Request.ForestIDs
	}

	companyID := runtime.CompanyID(ctx)
	uin := runtime.Uin(ctx)

	if !perm.HasManageAct(ctx, uin, req.Request.ID, foresttype.ResourceTypeArticle) {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission_update_resource"))
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeArticle, req.Request.ID)
		return
	}

	if err := dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&art).Error; err != nil {
			logs.ErrorContextf(ctx, "SaveArticle failed: err=%+v", err)
			return err
		}

		//do not append manager action list with admins in company
		if req.Request.PublicScope == foresttype.PublicScopeCompany {
			req.Request.ScopeIDs = types.NewUintArray([]uint{})
		}

		// 去重
		req.Request.ManagerIDs.RemoveDuplicates()
		req.Request.ScopeIDs.RemoveDuplicates()

		uins := types.NewUintArray(append(req.Request.ManagerIDs.Slice(), req.Request.ScopeIDs.Slice()...))
		uins.RemoveDuplicates()

		us := uins.Slice()
		if !employee.CheckUinsValid(ctx, us, companyID) {
			logs.ErrorContextf(ctx, "CheckUinsValid: exist no-local company[%v] uin in uins[%v]", companyID, us)
			runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_invalid_employee_id"))
			return err
		}

		return perm.UpdateResourceScope(ctx, tx, art.ID, foresttype.ResourceTypeArticle, req.Request.ScopeIDs.Slice(), req.Request.ManagerIDs.Slice(), req.Request.PublicScope, companyID)
	}); err != nil {
		logs.ErrorContextf(ctx, "EditArticle failed: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_article_update_failed"))
		return
	}
}

func DeleteArticle(ctx *gin.Context, req *DeleteArticleRequest, resp *apiobj.BaseResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "DeleteArticle.Validity failed: %s", resp.Message)
		return
	}

	uin := runtime.Uin(ctx)

	if !perm.HasManageAct(ctx, uin, req.Request.ID, foresttype.ResourceTypeArticle) {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission_update_resource"))
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeArticle, req.Request.ID)
		return
	}

	if err := dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		if err := perm.DeleteResourceScope(ctx, req.Request.ID, foresttype.ResourceTypeArticle, tx); err != nil {
			logs.ErrorContextf(ctx, "DeleteResourceScope failed: err=%+v", err)
			return err
		}
		return article.DeleteArticleByID(ctx, req.Request.ID, tx)
	}); err != nil {
		logs.ErrorContextf(ctx, "DeleteArticleByID(%v) err: %v", req.Request.ID, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_article_delete_failed"))
		return
	}
}

func GetArticle(ctx *gin.Context, req *GetArticleRequest, resp *GetArticleResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "GetArticle.Validity failed: %s", resp.Message)
		return
	}

	art, err := article.GetArticleByID(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetArticleByID(%v) err: %v", req.Request.ID, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_article_get_failed"))
		return
	}
	resp.Response = struct{ *foresttype.KeArticle }{KeArticle: art}
}

func SaveArticleContent(ctx *gin.Context, req *SaveArticleContentRequest, resp *apiobj.BaseResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "SaveArticleContent.Validity failed: %s", resp.Message)
		return
	}

	uin := runtime.Uin(ctx)

	if !perm.HasManageAct(ctx, uin, req.Request.ID, foresttype.ResourceTypeArticle) {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission_update_resource"))
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeArticle, req.Request.ID)
		return
	}
	art, err := article.GetArticleByID(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetArticleByID(%v) err: %v", req.Request.ID, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_article_get_failed"))
		return
	}

	art.Content = req.Request.Content
	if err = dbutil.Knownow().WithContext(ctx).Save(art).Error; err != nil {
		logs.ErrorContextf(ctx, "SaveArticleContent.Save() err: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_article_save_failed"))
		return
	}
}

func ListTemplate(ctx *gin.Context, req *dtoarticle.ListArticleTemplateRequest, resp *ListArticleTemplateResponse) {
	tmpls, err := article.ListTemplate(ctx, req.Request.PageQuery)
	if err != nil {
		logs.ErrorContextf(ctx, "ListTemplate() err: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_article_template_list_failed"))
		return
	}
	resp.Response.Data = tmpls
}

func GetArticleTemplate(ctx *gin.Context, req *GetArticleTemplateRequest, resp *GetArticleTemplateResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "GetArticleTemplate.Validity failed: %s", resp.Message)
		return
	}
	tmpl, err := article.GetTemplateByID(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetTemplateByID(%v) err: %v", req.Request.ID, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_article_template_get_failed"))
		return
	}
	resp.Response = struct{ *foresttype.KeArticleTemplate }{KeArticleTemplate: tmpl}
}

// Deprecated: 该接口已废弃
// AIWrite 智能撰写
// @Tags 写作空间
// @Summary 智能撰写
// @Description 智能撰写
// @Router /forest.AIWrite [post]
// @Param user body AIWriteRequest true "入参"l
// @Success 200 {object} apiobj.BaseResponse "返回值"
func AIWrite(ctx *gin.Context, req *AIWriteRequest, resp *apiobj.BaseResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "AIWrite.Validity failed: %s", resp.Message)
		return
	}
	w, err := article.NewAIWriteWrapper(ctx, req.Request.Cmd, req.Request.Content, "", req.Request.ForestIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "AIWrite.NewAIWriteWrapper() err: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_article_AIWrite_init_failed"))
		return
	}
	if err = w.DoCmd(); err != nil {
		logs.ErrorContextf(ctx, "AIWrite.DoCmd() err: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_article_AIWrite_exec_failed"))
		return
	}
}
