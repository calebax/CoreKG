package apikey

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// CreatAPIKey 生成API密钥
func CreatAPIKey(ctx context.Context, uin, company_id uint, name, purpose string, expiredAt *time.Time) (*accounttype.APIKey, error) {
	db := dbutil.Account().WithContext(ctx)
	key := &accounttype.APIKey{
		Uin:       uin,
		CompanyID: company_id,
		Name:      name,
		APIKey:    GenerateSecretKey(),
		Purpose:   purpose,
		Status:    accounttype.AccessKeyStatusNormal,
		ExpiredAt: expiredAt,
	}

	if err := db.Create(key).Error; err != nil {
		logs.ErrorContextf(ctx, "[acount][CreatAPIKey] create api key failed: %v", err)
		return nil, err
	}
	return key, nil
}

// GetApiKey 获取API密钥
func GetApiKey(ctx context.Context, key string) (*accounttype.APIKey, error) {
	db := dbutil.Account().WithContext(ctx)
	keyObj := &accounttype.APIKey{}
	if err := db.Where("api_key = ?", key).First(keyObj).Error; err != nil {
		logs.ErrorContextf(ctx, "[acount][GetApiKey] get api key failed: %v", err)
		return nil, err
	}
	return keyObj, nil
}

// GetApiKeyByID 获取API密钥
func GetApiKeyByID(ctx context.Context, id uint) (*accounttype.APIKey, error) {
	db := dbutil.Account().WithContext(ctx)
	keyObj := &accounttype.APIKey{}
	if err := db.Where("id = ?", id).First(keyObj).Error; err != nil {
		logs.ErrorContextf(ctx, "[acount][GetApiKeyByID] get api key failed: %v", err)
		return nil, err
	}
	return keyObj, nil
}

// GetAPIKeyInfo 获取APIKey信息
func GetAPIKeyInfo(ctx context.Context, key string) (*accounttype.APIKey, error) {
	db := dbutil.Account().WithContext(ctx)
	keyObj := &accounttype.APIKey{}
	if err := db.Where("api_key = ?", key).First(keyObj).Error; err != nil {
		logs.ErrorContextf(ctx, "[acount][GetAPIKeyInfo] get api key failed: %v", err)
		return nil, err
	}
	return keyObj, nil
}

// QueryListApiKey 查询API密钥
func QueryListApiKey(ctx context.Context, opt apiobj.PageQuery, keyList *QueryListAPIKeyResponse) error {
	// 初始化查询
	query := dbutil.Account().WithContext(ctx).
		Table(accounttype.TableNameAPIKey).
		Where("uin = ?", opt.Uin).
		Where("deleted_at is null")

	// 添加过滤器
	for _, filter := range opt.Filters {
		switch filter.Field {
		case "name":
			query = query.Where("account_api_key.name = ?", filter.Value[0])
		default:
			logs.WarnContextf(ctx, "[acount][QueryListApiKey] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	// 查询总记录数
	if err := query.Count(&keyList.Total).Error; err != nil {
		return err
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
	err := query.Find(&keyList.Data).Error
	if err != nil {
		return err
	}
	return nil
}

// DeleteApiKey 删除用户密钥
func DeleteApiKey(ctx context.Context, key string) error {
	db := dbutil.Account().WithContext(ctx)
	if err := db.Where("api_key = ?", key).Delete(&accounttype.APIKey{}).Error; err != nil {
		logs.ErrorContextf(ctx, "[acount][DeleteApiKey] delete api key failed: %v", err)
		return err
	}
	return nil
}

// DeleteAPIKeyByID 删除API密钥及其关联权限
func DeleteAPIKeyByID(ctx context.Context, id uint) error {
	db := dbutil.Account().WithContext(ctx)

	err := db.Transaction(func(tx *gorm.DB) error {
		// 删除关联权限
		if err := tx.Where("api_key_id = ?", id).
			Delete(&accounttype.APIKeyPrivilege{}).Error; err != nil {
			logs.ErrorContextf(ctx, "[acount][DeleteAPIKeyByID] delete api key privileges failed: %v", err)
			return fmt.Errorf("failed to delete api key privileges: %w", err)
		}

		// 删除API密钥
		if err := tx.Where("id = ?", id).
			Delete(&accounttype.APIKey{}).Error; err != nil {
			logs.ErrorContextf(ctx, "[acount][DeleteAPIKeyByID] delete api key failed: %v", err)
			return fmt.Errorf("failed to delete api key: %w", err)
		}

		return nil
	})

	if err != nil {
		logs.ErrorContextf(ctx, "[acount][DeleteAPIKeyByID] transaction failed: %v", err)
		return fmt.Errorf("failed to delete api key and its privileges: %w", err)
	}

	return nil
}

func GetAgentApikey(agentID, apkID uint) (*accounttype.APIKey, error) {
	var k *accounttype.APIKey
	if err := dbutil.Account().Table(accounttype.TableNameAPIKey).
		Where("resource_type = ?", chattype.ResourceTypeAgent).
		Where("resource_id = ?", agentID).
		Where("id = ?", apkID).
		First(&k).Error; err != nil {
		return nil, err
	}
	return k, nil

}

type ListAgentApikeyResponse struct {
	apiobj.QueryResponse
	Data []*AgentApiKeyItem
}

type AgentApiKeyItem struct {
	accounttype.APIKey
}

// QueryAgentApiKeyList 查询AgentAPI密钥
func QueryAgentApiKeyList(ctx context.Context, opt apiobj.PageQuery, resp *ListAgentApikeyResponse) error {
	// 初始化查询
	query := dbutil.Account().WithContext(ctx).
		Table(accounttype.TableNameAPIKey).
		Where("resource_type = ?", chattype.ResourceTypeAgent).
		Where("company_id = ?", opt.CompanyID).
		Where("deleted_at is null")

	// 添加过滤器
	for _, filter := range opt.Filters {
		switch filter.Field {
		case "agent_id":
			query = query.Where("resource_id = ?", filter.Value[0])
		case "name":
			query = query.Where("account_api_key.name like ?", fmt.Sprintf("%%%s%%", filter.Value[0]))
		default:
			logs.WarnContextf(ctx, "[acount][QueryListApiKey] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	if err := query.Count(&resp.Total).Error; err != nil {
		return err
	}
	if resp.Total == 0 {
		return nil
	}
	resp.Limit = opt.Limit
	resp.Offset = opt.Offset

	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}

	// 分页逻辑
	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}
	err := query.Find(&resp.Data).Error
	if err != nil {
		return err
	}
	return nil
}

// ListAPIKeyAgentID 获取轻应用APIKey
func ListAPIKeyAgentID(ctx context.Context, companyID, agentID, uin uint) (key string, err error) {
	var data []*accounttype.APIKey
	query := dbutil.Account().WithContext(ctx).
		Table(accounttype.TableNameAPIKey).
		Where("resource_type = ?", chattype.ResourceTypeAgent).
		Where("company_id = ?", companyID).
		Where("deleted_at is null").
		Where("resource_id = ?", agentID).
		Where("status = ?", accounttype.UinStatusNormal)
	if companyID == 0 {
		query.Where("uin = ?", uin)
	}
	err = query.Find(&data).
		Error
	if err != nil {
		logs.ErrorContextf(ctx, "dbutil.AgentApiKeyList error, %s", err.Error())
		return key, err
	}
	if len(data) == 0 {
		return key, nil
	}
	key = data[0].APIKey
	return key, nil
}
