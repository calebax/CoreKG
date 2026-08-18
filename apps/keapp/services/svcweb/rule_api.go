package svcweb

import (
	"context"
	"errors"

	"github.com/insmtx/corekg/apps/keapp/models/web"
)

var (
	ErrRuleNotFound      = errors.New("crawl rule not found")
	ErrAddRuleFailed     = errors.New("add crawl rule failed")
	ErrUpdateRuleFailed  = errors.New("update crawl rule failed")
	ErrDeleteRuleFailed  = errors.New("delete crawl rule failed")
)

func AddCrawlRule(ctx context.Context, entity *web.KeWebCrawlRule) error {
	dao := web.NewCrawlRuleDao()
	if err := dao.Insert(ctx, entity); err != nil {
		return ErrAddRuleFailed
	}
	return nil
}

func ListCrawlRules(ctx context.Context, appID uint) ([]*web.KeWebCrawlRule, error) {
	dao := web.NewCrawlRuleDao()
	return dao.ListByAppID(ctx, appID)
}

func UpdateCrawlRule(ctx context.Context, entity *web.KeWebCrawlRule) error {
	dao := web.NewCrawlRuleDao()
	existing, err := dao.GetByID(ctx, entity.ID)
	if err != nil || existing == nil {
		return ErrRuleNotFound
	}
	if err := dao.Update(ctx, entity); err != nil {
		return ErrUpdateRuleFailed
	}
	return nil
}

func DeleteCrawlRule(ctx context.Context, id uint) error {
	dao := web.NewCrawlRuleDao()
	existing, err := dao.GetByID(ctx, id)
	if err != nil || existing == nil {
		return ErrRuleNotFound
	}
	if err := dao.Delete(ctx, id); err != nil {
		return ErrDeleteRuleFailed
	}
	return nil
}
