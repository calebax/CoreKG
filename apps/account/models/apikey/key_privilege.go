package apikey

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type QueryListAPIKeyPrivilegeResponse struct {
	apiobj.QueryResponse
	Data []*ListAPIKeyPrivilege
}

type ListAPIKeyPrivilege struct {
	accounttype.APIKeyPrivilege
	API         string `json:"api"`
	Description string `json:"description"`
	Action      string `json:"action"`
	ActionPath  string `json:"action_path"`
	Type        string `json:"type"`
}

// QueryListApiKeyPrivilege 查询API密钥权限列表
func QueryListApiKeyPrivilege(ctx context.Context, opt apiobj.PageQuery, apiKeyID uint, keyList *QueryListAPIKeyPrivilegeResponse) error {

	query := dbutil.Account().WithContext(ctx).
		Table(accounttype.TableNameAPIKeyPrivilege).
		Select(`
            account_api_key_privilege.*, 
            account_privilege.api as api,
            account_privilege.description as description,
            account_privilege.action as action,
            account_privilege.action_path as action_path,
            account_privilege.type as type
        `).
		Joins("JOIN account_api_key ON account_api_key_privilege.api_key_id = account_api_key.id").
		Joins("JOIN account_privilege ON account_privilege.id = account_api_key_privilege.api_id").
		Where("account_api_key_privilege.deleted_at IS NULL").
		Where("account_api_key.id = ?", apiKeyID)

	// 添加过滤器
	for _, filter := range opt.Filters {
		switch filter.Field {
		case "name":
			query = query.Where("account_api_key.name = ?", filter.Value[0])
		default:
			logs.WarnContextf(ctx, "[QueryListApiKeyPrivilege] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	// 查询总记录数
	if err := query.Count(&keyList.Total).Error; err != nil {
		return fmt.Errorf("failed to count records: %v", err)
	}
	if keyList.Total == 0 {
		return nil
	}

	// 添加排序
	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}

	// 分页逻辑
	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}

	if err := query.Find(&keyList.Data).Error; err != nil {
		return fmt.Errorf("failed to fetch records: %v", err)
	}
	return nil
}

// AddApiKeyPrivilege 添加API密钥权限
func AddApiKeyPrivilege(keyID uint, privilegeIDs []uint) error {
	db := dbutil.Account()
	if len(privilegeIDs) > 0 {
		isExist, err := employee.IsExistPrivilegeIDs(privilegeIDs)
		if err != nil {
			return err
		}
		if !isExist {
			return errors.New("权限有误")
		}
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if len(privilegeIDs) > 0 {
			rel := make([]*accounttype.APIKeyPrivilege, 0)
			for _, p := range privilegeIDs {
				rel = append(rel, &accounttype.APIKeyPrivilege{
					ApiKeyID: keyID,
					ApiID:    p,
				})
			}
			if err := tx.Create(rel).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

// DeleteApiKeyPrivilege 删除API密钥权限
func DeleteApiKeyPrivilege(keyID uint, privilegeIDs []uint) error {
	db := dbutil.Account()
	if len(privilegeIDs) > 0 {
		isExist, err := employee.IsExistPrivilegeIDs(privilegeIDs)
		if err != nil {
			return err
		}
		if !isExist {
			return errors.New("权限有误")
		}

	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if len(privilegeIDs) > 0 {
			if err := tx.Table(accounttype.TableNameAPIKeyPrivilege).Unscoped().
				Where("api_key_id = ?", keyID).
				Where("api_id in (?)", privilegeIDs).
				Delete(&accounttype.APIKeyPrivilege{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// IsExistApiKeyPrivilege 判断API密钥权限是否已存在
func IsExistApiKeyPrivilege(ctx context.Context, keyID uint, PrivilegeIDs []uint) (bool, error) {
	var count int64
	err := dbutil.Account().Table(accounttype.TableNameAPIKeyPrivilege).WithContext(ctx).
		Where("api_key_id = ?", keyID).
		Where("api_id in (?)", PrivilegeIDs).
		Count(&count).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[IsExistApiKeyPrivilege] failed to count records: %v", err)
		return false, err
	}
	if len(PrivilegeIDs) != int(count) {
		return false, nil
	}
	return true, nil
}

// GetApiKeyPrivilegeByAPIKeyID 通过API密钥ID获取权限列表
func GetApiKeyPrivilegeByAPIKeyID(ctx context.Context, keyID uint) ([]uint, error) {
	var privilegeIDs []uint
	err := dbutil.Account().Table(accounttype.TableNameAPIKeyPrivilege).WithContext(ctx).
		Select("api_id").
		Where("api_key_id = ?", keyID).
		Find(&privilegeIDs).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[GetApiKeyPrivilegeByAPIKeyID] failed to fetch records: %v", err)
		return nil, err
	}

	return privilegeIDs, nil
}

// GetApiKeyPrivilegeByAPIKeyIDAndAPIID 通过API密钥ID和API权限ID获取权限列表
func GetApiKeyPrivilegeByAPIKeyIDAndAPIID(ctx context.Context, keyID uint, apiID uint) (bool, error) {
	var count int64
	err := dbutil.Account().Table(accounttype.TableNameAPIKeyPrivilege).WithContext(ctx).
		Where("api_key_id = ?", keyID).
		Where("api_id = ?", apiID).
		Count(&count).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[GetApiKeyPrivilegeByAPIKeyIDAndAPIID] failed to count records: %v", err)
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

// GetAPIPrivilegeByAPI 根据API获取权限信息
func GetAPIPrivilegeByAPI(ctx context.Context, api string) (*accounttype.APIPrivilege, error) {
	position := &accounttype.APIPrivilege{}
	err := dbutil.Account().Table(accounttype.TableNameAPIPrivilege).WithContext(ctx).
		Where("api = ?", api).First(position).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetAPIPrivilegeByAPI faild err: %v", err)
		return nil, err
	}
	return position, nil
}
