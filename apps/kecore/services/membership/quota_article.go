package membership

import (
	"context"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
)

type articleUsageProvider struct{}

func newArticleUsageProvider() UsageProvider {
	return &articleUsageProvider{}
}

func (p *articleUsageProvider) GetResourceType() QuotaResourceType {
	return QuotaResourceTypeArticle
}

func (p *articleUsageProvider) CalculateUsage(ctx context.Context, req *UsageQueryReq) (*ResourceUsageStatItem, error) {
	total, err := forest.NewArticleDao().CountByCond(ctx, &forest.ArticleCond{
		BaseCond: forest.BaseCond{
			CompanyID: req.CompanyID,
		},
		ArticleTypes: []foresttype.ArticleType{foresttype.ArticleTypeArticle},
	})
	if err != nil {
		return nil, err
	}
	return &ResourceUsageStatItem{
		ResourceType: QuotaResourceTypeArticle,
		Usage:        total,
	}, nil
}
