package connectors

import (
	"context"
	"time"

	"github.com/insmtx/corekg/pkgs/connectors/tokenmgr"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
)

const externalTokenStatusValid = 1
const externalTokenStatusInvalid = 0

func ListBindings(ctx context.Context, uin uint) ([]*tokenmgr.ExternalToken, error) {
	var bindings []*tokenmgr.ExternalToken

	err := dbutil.Account().WithContext(ctx).
		Table(tokenmgr.ExternalToken{}.TableName()+" AS t").
		Where("t.deleted_at IS NULL").
		Where("t.uin = ?", uin).
		Where("t.status = ?", externalTokenStatusValid).
		Find(&bindings).Error
	if err != nil {
		logs.ErrorContextf(ctx, "ListBindings query failed, uin=%d, err=%v", uin, err)
		return nil, err
	}

	return bindings, nil
}

func DeleteBindingByID(ctx context.Context, id uint) error {
	err := dbutil.Account().WithContext(ctx).
		Model(&tokenmgr.ExternalToken{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     externalTokenStatusInvalid,
			"deleted_at": time.Now(),
		}).Error
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteBindingByID failed, id=%d, err=%v", id, err)
		return err
	}
	return nil
}
