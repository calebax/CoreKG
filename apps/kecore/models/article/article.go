package article

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type InfoItem struct {
	foresttype.KeArticle
	IsAdmin bool `json:"isAdmin"`
}

type ListArticleResponse struct {
	Data []InfoItem `json:"data"`
	apiobj.QueryResponse
}

// ListArticles will list articles
func ListArticles(ctx context.Context, opt apiobj.PageQuery, articleList *ListArticleResponse) error {
	query := dbutil.Knownow().Table(foresttype.TableNameKeArticle).
		Where("deleted_at IS NULL")
	for _, filter := range opt.Filters {
		switch filter.Field {
		case "title":
			query = query.Where("title LIKE ?", "%"+filter.Value[0]+"%")
		default:
			logs.WarnContextf(ctx, "[ListArticles] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	// ======== 核心权限控制逻辑 ========
	// 构建有权限查看的 article_id 子查询
	authIDs := dbutil.Knownow().Table(foresttype.TableNameKeResourceScope).
		Select("resource_id").
		Where("resource_type = ? AND deleted_at IS NULL", foresttype.ResourceTypeArticle).
		Where("("+
			// 1. 公开权限 (action = 'view' 且 scope_type = 'public')
			"(action = ? AND scope_type = ?) OR "+
			// 2. 个人管理权限 (action = 'manage' 且 scope_type = 'user' 且 scope_id = 当前用户)
			"(action = ? AND scope_type = ? AND scope_id = ?) OR "+
			// 3. 个人查看权限 (action = 'view' 且 scope_type = 'user' 且 scope_id = 当前用户)
			"(action = ? AND scope_type = ? AND scope_id = ?) OR "+
			// 4. 公司权限 (action = 'view' 且 scope_type = 'company' 且 scope_id = 当前公司)
			"(action = ? AND scope_type = ? AND scope_id = ?)"+
			")", foresttype.ActionView, foresttype.ScopeTypePublic, // 公开
			foresttype.ActionManage, foresttype.ScopeTypeUser, opt.Uin, // 个人管理
			foresttype.ActionView, foresttype.ScopeTypeUser, opt.Uin, // 个人查看
			foresttype.ActionView, foresttype.ScopeTypeCompany, opt.CompanyID) // 公司权限

	//有权限查看
	query = query.Where("id IN (?)", authIDs)

	// 处理 BeginTime 和 EndTime
	if !opt.BeginTime.IsZero() {
		query = query.Where("created_at >= ?", opt.BeginTime)
	}
	if !opt.EndTime.IsZero() {
		query = query.Where("created_at <= ?", opt.EndTime)
	}

	if err := query.Count(&articleList.Total).Error; err != nil {
		return err
	}
	if articleList.Total == 0 {
		logs.DebugContextf(ctx, "[ListArticles] no article found")
		return nil
	}
	logs.DebugContextf(ctx, "[ListArticles] found %d articles", articleList.Total)

	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}

	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}

	articleList.Offset = opt.Offset
	articleList.Limit = opt.Limit

	if err := query.Find(&articleList.Data).Error; err != nil {
		logs.ErrorContextf(ctx, "[ListArticles] failed to query data: %v", err)
		return err
	}

	// 查询管理权限
	manageForestIDs := make(map[uint]bool)

	manageScopesIDs := perm.GetManageList(ctx, opt.Uin, foresttype.ResourceTypeArticle)
	for _, scope := range manageScopesIDs {
		manageForestIDs[scope] = true
	}

	for _, frs := range articleList.Data {
		frs.IsAdmin = manageForestIDs[frs.ID]
	}

	return nil
}

func GetArticleByID(ctx context.Context, id uint) (*foresttype.KeArticle, error) {
	var article foresttype.KeArticle
	if err := dbutil.Knownow().WithContext(ctx).Where("id = ?", id).First(&article).Error; err != nil {
		logs.ErrorContextf(ctx, "[GetArticleByID] get article(%v) failed: %v", id, err)
		return nil, err
	}
	return &article, nil
}

func DeleteArticleByID(ctx context.Context, id uint, tx *gorm.DB) error {
	return tx.WithContext(ctx).Delete(&foresttype.KeArticle{}, id).Error
}

func ListTemplate(ctx context.Context, _ apiobj.PageQuery) ([]*foresttype.KeArticleTemplate, error) {
	var tmpl []*foresttype.KeArticleTemplate

	if err := dbutil.Knownow().
		WithContext(ctx).
		Where("deleted_at IS NULL").
		Find(&tmpl).
		Error; err != nil {
		logs.ErrorContextf(ctx, "[ListTemplate] list article template failed: %v", err)
		return nil, err
	}
	return tmpl, nil
}

func GetTemplateByID(ctx context.Context, id uint) (*foresttype.KeArticleTemplate, error) {
	var template foresttype.KeArticleTemplate
	if err := dbutil.Knownow().WithContext(ctx).Where("id = ?", id).First(&template).Error; err != nil {
		logs.ErrorContextf(ctx, "[GetTemplateByID] GetTemplateByID(%v) failed: %v", id, err)
		return nil, err
	}
	return &template, nil
}

// GetArticleByCompanyID get company's article
func GetArticleByCompanyID(ctx context.Context, companyID uint) ([]*foresttype.KeArticle, error) {
	var res []*foresttype.KeArticle
	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKeArticle).
		Where("company_id = ?", companyID).
		Find(&res).
		Error; err != nil {
		logs.ErrorContextf(ctx, "[GetArticleByCompanyID] get article failed: %v", err)
		return nil, err
	}
	return res, nil
}
