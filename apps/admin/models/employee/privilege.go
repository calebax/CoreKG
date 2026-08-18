package employee

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
)

type QueryPrivilegeListResponse struct {
	apiobj.QueryResponse
	Data []*admintype.Privilege
}

// QueryPrivilegeList 查询权限列表
func QueryPrivilegeList(ctx context.Context, opt apiobj.PageQuery, privilegeList *QueryPrivilegeListResponse) error {
	query := dbutil.Account().WithContext(ctx).Table(admintype.TableNamePrivilege).Where("deleted_at is null")

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "action":
			query = query.Where(admintype.TableNamePrivilege+".action LIKE ?", "%"+filter.Value[0]+"%")
		case "action_path":
			query = query.Where(admintype.TableNamePrivilege+".action_path LIKE ?", "%"+filter.Value[0]+"%")
		case "api":
			query = query.Where(admintype.TableNamePrivilege+".api LIKE ?", "%"+filter.Value[0]+"%")
		case "type":
			query = query.Where(admintype.TableNamePrivilege+".type = ?", filter.Value[0])
		default:
			logs.WarnContextf(ctx, "[opaccount][QueryPrivilegeList] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	if err := query.Count(&privilegeList.Total).Error; err != nil {
		return err
	}
	if privilegeList.Total == 0 {
		return nil
	}

	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}

	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}

	err := query.Find(&privilegeList.Data).Error
	if err != nil {
		return err
	}
	return nil
}

// CreatePrivilege 添加权限及权限
func CreatePrivilege(opt *admintype.Privilege) (*admintype.Privilege, error) {
	db := dbutil.Account()
	opt.ID = 0
	if err := db.Create(opt).Error; err != nil {
		return nil, err
	}

	return opt, nil
}

// GetPrivilegeByID 根据ID获取权限信息
func GetPrivilegeByID(id uint) (*admintype.Privilege, error) {
	position := &admintype.Privilege{}
	err := dbutil.Account().Table(admintype.TableNamePrivilege).Where("id = ?", id).First(position).Error
	if err != nil {
		return nil, err
	}
	return position, nil
}

// ModifyPrivilege 修改权限信息
func ModifyPrivilege(id uint, p *admintype.Privilege) (*admintype.Privilege, error) {
	privilege, err := GetPrivilegeByID(id)
	if err != nil {
		return nil, err
	}

	p.Model = privilege.Model
	if err := dbutil.Account().Save(p).Error; err != nil {
		return nil, err
	}
	return privilege, nil
}

// DeletePrivilege 删除权限及权限权限信息
func DeletePrivilege(id uint) error {
	position, err := GetPrivilegeByID(id)
	if err != nil {
		return err
	}

	if err := dbutil.Account().Delete(position).Error; err != nil {
		return err
	}

	return nil
}
