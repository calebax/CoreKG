package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/encryptor"
	"github.com/ygpkg/yg-go/logs"
)

// CreateUserOption 创建用户选项
type CreateUserOption struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Password     string `json:"password"`
	CompanyQuota uint   `json:"company_quota"`
}

// IsExist 检查用户是否存在
func (opt *CreateUserOption) IsExist(exceptIDs ...uint) (bool, error) {
	var cnt int64
	query := dbutil.Account().Table(accounttype.TableNameUser).
		Where("deleted_at is null").
		Where("(email = ? or phone = ?)", opt.Email, opt.Phone)
	if len(exceptIDs) > 0 {
		query = query.Where("id not in (?)", exceptIDs)
	}
	err := query.Count(&cnt).Error
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// CreateUser 创建用户
func CreateUser(ctx context.Context, opt *CreateUserOption) (*accounttype.User, error) {
	passwd := EncryptPassword(ctx, opt.Password)
	if passwd == nil {
		logs.ErrorContextf(ctx, "[user][CreateUser] encrypt password failed")
		return nil, errors.New("created password failed")
	}
	var e, p *string = nil, nil
	if len(opt.Email) > 0 {
		e = &opt.Email
	}

	if len(opt.Phone) > 0 {
		p = &opt.Phone
	}

	u := &accounttype.User{
		Identify:     encryptor.GenerateUUID(),
		Name:         opt.Name,
		Email:        e,
		Phone:        p,
		Password:     passwd,
		CompanyQuota: opt.CompanyQuota,
	}
	if err := dbutil.Account().WithContext(ctx).Create(u).Error; err != nil {
		logs.ErrorContextf(ctx, "[user][CreateUser] create user failed: %v", err)
		return nil, err
	}
	return u, nil
}

// ModifyUser 修改用户基本信息
func ModifyUser(ctx context.Context, id uint, opt *CreateUserOption) (*accounttype.User, error) {
	u, err := GetUserByID(id)
	if err != nil {
		logs.ErrorContextf(ctx, "[user][ModifyUser] get user %d failed: %v", id, err)
		return nil, err
	}
	u.Name = opt.Name
	u.Email = &opt.Email
	u.Phone = &opt.Phone
	u.CompanyQuota = opt.CompanyQuota

	if err := dbutil.Account().WithContext(ctx).Save(u).Error; err != nil {
		logs.ErrorContextf(ctx, "[user][ModifyUser] modify user failed: %v", err)
		return nil, err
	}
	return u, nil
}

// ModifyUserPassword 修改用户密码
func ModifyUserPassword(ctx context.Context, id uint, password string) (*accounttype.User, error) {
	u, err := GetUserByID(id)
	if err != nil {
		logs.ErrorContextf(ctx, "[user][ModifyUser] get user %d failed: %v", id, err)
		return nil, err
	}

	passwd := EncryptPassword(ctx, password)
	if passwd == nil {
		logs.ErrorContextf(ctx, "[user][CreateUser] encrypt password failed")
		return nil, errors.New("created password failed")
	}
	u.Password = passwd
	if err := dbutil.Account().WithContext(ctx).Save(u).Error; err != nil {
		logs.ErrorContextf(ctx, "[user][CreateUser] create user failed: %v", err)
		return nil, err
	}
	return u, nil
}

// QueryUserListResponse 团队列表响应
type QueryUserListResponse struct {
	apiobj.QueryResponse
	Data []*QueryUserListItem `json:"data"`
}

// QueryUserListItem 团队列表项
type QueryUserListItem struct {
	accounttype.User
	EmployeeCount int64 `json:"employee_count" gorm:"column:employee_count"`
}

// QueryUserList 查询对比列表
func QueryUserList(ctx context.Context, opt *apiobj.PageQuery, ret *QueryUserListResponse) error {
	query := dbutil.Account().Table(accounttype.TableNameUser).
		WithContext(ctx).
		Where("deleted_at is null").
		Joins("LEFT JOIN (select count(*) as c,user_id from account_employee where deleted_at is null group by user_id) as emp " +
			" on emp.user_id=user.id ")

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "name":
			query = query.Where("name like (?)", "%"+filter.Value[0]+"%")
		case "phone":
			query = query.Where("phone like (?)", "%"+filter.Value[0]+"%")
		default:
			logs.WarnContextf(ctx, "[user][QueryUserList] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	if err := query.Count(&ret.Total).Error; err != nil {
		return err
	}
	if ret.Total == 0 {
		return nil
	}

	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	} else {
		query = query.Order("id desc")
	}

	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	} else {
		query = query.Limit(10)
	}

	err := query.Select("user.*,emp.c as employee_count").Find(&ret.Data).Error
	if err != nil {
		return err
	}
	return nil
}

// UserDetail 用户详情
type UserDetail struct {
	accounttype.User
}

// GetUserByID 通过ID获取用户
func GetUserByID(id uint) (*accounttype.User, error) {
	out := &accounttype.User{}
	err := dbutil.Account().First(out, id).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetUserDetail 获取用户详情
func GetUserDetail(id uint) (*UserDetail, error) {
	out := &UserDetail{}
	u, err := GetUserByID(id)
	if err != nil {
		return nil, err
	}
	out.User = *u
	// todo
	return out, nil
}
