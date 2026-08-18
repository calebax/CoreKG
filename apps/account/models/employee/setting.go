package employee

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"gopkg.in/yaml.v3"
)

// QuerySettingListResponse 员工列表
type QuerySettingListResponse struct {
	apiobj.QueryResponse

	Data []*settings.SettingItem
}

// QuerySettingList 查询运营端配置列表
func QuerySettingList(ctx context.Context, opt apiobj.PageQuery, resp *QuerySettingListResponse) (err error) {
	query := dbutil.Core().Table(settings.TableNameSettings).
		WithContext(ctx).
		Where("deleted_at is null")

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "name":
			query = query.Where("name LIKE ?", fmt.Sprintf("%%%s%%", filter.Value[0]))
		case "group":
			query = query.Where("group = ?", "%"+filter.Value[0]+"%")
		case "key":
			query = query.Where("key = ?", "%"+filter.Value[0]+"%")
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

type SettingDetail struct {
	*settings.SettingItem
	ValueParse interface{} `json:"value_parse"`
}

// GetSettingDetail 获取配置详情
func GetSettingDetail(id uint, detail *SettingDetail) error {
	setting, err := GetSettingByID(id)
	if err != nil {
		return err
	}
	detail.SettingItem = setting

	if setting.ValueType == settings.ValueYaml {
		value := make(map[string]interface{})
		err = yaml.Unmarshal([]byte(setting.Value), value)
		if err != nil {
			return err
		}
		detail.ValueParse = value
	}

	return nil
}

func GetSettingByID(id uint) (*settings.SettingItem, error) {
	out := &settings.SettingItem{}
	err := dbutil.Core().Table(settings.TableNameSettings).First(out, id).Error
	if err != nil {
		return nil, err
	}

	return out, nil
}

func CreateSetting(req *settings.SettingItem) error {
	return dbutil.Core().Table(settings.TableNameSettings).Create(req).Error
}

func UpdateSetting(id uint, value string) error {
	s, err := GetSettingByID(id)
	if err != nil {
		return err
	}
	s.Value = value
	err = settings.Updates(s)
	if err != nil {
		return err
	}
	return nil
}
