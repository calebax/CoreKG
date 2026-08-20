package login_setting

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/version"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
)

// CreateLoginSetting 创建LoginSetting
func CreateLoginSetting(info *LoginSetting) error {
	return dbutil.Account().Table(TableNameLoginSetting).Create(info).Error
}

// QueryLoginSettingListResponse 登录配置列表
type QueryLoginSettingListResponse struct {
	apiobj.QueryResponse
	Data []*LoginSetting
}

// QueryLoginSettingList 查询登录配置列表
func QueryLoginSettingList(ctx context.Context, opt apiobj.PageQuery, resp *QueryLoginSettingListResponse) (err error) {
	query := dbutil.Account().Table(TableNameLoginSetting).Where("deleted_at is null")
	for _, filter := range opt.Filters {
		switch filter.Field {
		case "domain_name":
			query = query.Where("domain_name LIKE ?", fmt.Sprintf("%%%s%%", filter.Value[0]))
		case "env":
			query = query.Where("env = ?", "%"+filter.Value[0]+"%")
		case "methods":
			query = query.Where("methods = ?", "%"+filter.Value[0]+"%")
		case "title":
			query = query.Where("title LIKE ?", "%"+filter.Value[0]+"%")
		default:
			logs.WarnContextf(ctx, "[opaccount][QueryEmployeeList] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	if err := query.Count(&resp.Total).Error; err != nil {
		return err
	}
	if resp.Total == 0 {
		return nil
	}

	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}

	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}

	err = query.Find(&resp.Data).Error
	if err != nil {
		return err
	}
	return nil
}

// UpdateLoginSetting 修改配置
func UpdateLoginSetting(info *LoginSetting) error {
	err := dbutil.Account().Table(TableNameLoginSetting).
		Select("IsEnableWeChat", "IsEnableWeChatCom", "IsEnablePhone", "IsEnablePassword", "IsEnableEmail", "AllowRegister").
		Where("id = ?", info.ID).
		Updates(info).Error
	if err != nil {
		return err
	}
	return nil
}

// DeleteLoginSetting 删除配置
func DeleteLoginSetting(info *LoginSetting) error {
	err := dbutil.Account().Table(TableNameLoginSetting).Where("id = ?", info.ID).Delete(info).Error
	if err != nil {
		return err
	}
	return nil
}

// GetLoginSettingByPath 根据环境、域名和路径获取配置信息
func GetLoginSettingByPath(domainName, path string) (*LoginSetting, error) {
	if strings.HasPrefix(domainName, "https://") {
		domainName = strings.TrimPrefix(domainName, "https://")
	} else if strings.HasPrefix(domainName, "http://") {
		domainName = strings.TrimPrefix(domainName, "http://")
	}
	parts := strings.Split(domainName, "/")
	if len(parts) > 0 {
		domainName = parts[0]
	}
	var infos []*LoginSetting
	err := dbutil.Account().Table(TableNameLoginSetting).
		Where("domain_name = ? ", domainName).
		Find(&infos).Error
	if err != nil {
		return nil, err
	}

	if version.DeployMode() != "" {
		var setting *LoginSetting
		err := dbutil.Account().Table(TableNameLoginSetting).
			Where("id = ? ", 1).First(&setting).Error
		if err != nil {
			return nil, err
		}
		return setting, nil
	}

	for {
		for _, v := range infos {
			if v.Path == path {
				return v, nil
			}
		}
		path = reducePath(path)
		if path == "" {
			break
		}
	}

	return nil, fmt.Errorf("login setting not found")
}

// reducePath 减少路径长度
func reducePath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], "/")
	}
	return ""
}

// GetLoginSettingByID 根据ID获取配置信息
func GetLoginSettingByID(id uint) (*LoginSetting, error) {
	info := &LoginSetting{}
	err := dbutil.Account().Table(TableNameLoginSetting).Where("id = ?", id).First(info).Error
	if err != nil {
		return nil, err
	}
	return info, nil
}

// GetLoginSettingByIssuer 根据issues获取配置信息
func GetLoginSettingByIssuer(issuer string) (*LoginSetting, error) {
	info := &LoginSetting{}
	err := dbutil.Account().Table(TableNameLoginSetting).Where("issuer = ?", issuer).First(info).Error
	if err != nil {
		return nil, err
	}
	return info, nil
}
