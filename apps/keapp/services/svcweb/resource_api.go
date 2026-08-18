package svcweb

import (
	"context"
	"errors"

	"github.com/insmtx/corekg/apps/keapp/models/web"
)

var (
	ErrResourceNotFound    = errors.New("web resource not found")
	ErrDeleteResourceFailed = errors.New("delete web resource failed")
)

func ListWebResources(ctx context.Context, appID uint, limit, offset int) ([]*web.KeWebResource, int64, error) {
	dao := web.NewWebResourceDao()
	return dao.ListByAppID(ctx, appID, limit, offset)
}

func GetWebResource(ctx context.Context, id uint) (*web.KeWebResource, error) {
	dao := web.NewWebResourceDao()
	entity, err := dao.GetByID(ctx, id)
	if err != nil {
		return nil, ErrResourceNotFound
	}
	if entity == nil {
		return nil, ErrResourceNotFound
	}
	return entity, nil
}

func DeleteWebResource(ctx context.Context, id uint) error {
	dao := web.NewWebResourceDao()
	entity, err := dao.GetByID(ctx, id)
	if err != nil || entity == nil {
		return ErrResourceNotFound
	}
	if err := dao.SoftDelete(ctx, id); err != nil {
		return ErrDeleteResourceFailed
	}
	return nil
}
