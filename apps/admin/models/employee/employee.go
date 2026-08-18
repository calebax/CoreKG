package employee

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/pkgs/utils/validate"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// GetEmployeeByID 获取员工信息
func GetEmployeeByID(ctx context.Context, id uint) (*admintype.Employee, error) {
	user := &admintype.Employee{}
	err := dbutil.Account().WithContext(ctx).Table((&admintype.Employee{}).TableName()).
		Where("id = ?", id).
		Find(user).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[admin][%s] found failed, %s", id, err)
		return nil, err
	}
	if user.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return user, nil
}

// GetEmployeeByUin 获取员工信息
func GetEmployeeByUin(ctx context.Context, uin uint) (*admintype.Employee, error) {
	user := &admintype.Employee{}
	err := dbutil.Account().WithContext(ctx).Table((&admintype.Employee{}).TableName()).
		Where("uin = ?", uin).
		Find(user).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[admin][uin_%s] found failed, %s", uin, err)
		return nil, err
	}
	if user.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return user, nil
}

// GetEmployeeByUsername 获取员工信息
func GetEmployeeByUsername(ctx context.Context, username string) (*admintype.Employee, error) {
	user := &admintype.Employee{}
	err := dbutil.Account().Table(user.TableName()).
		Where("username = ?", username).
		Find(user).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[admin][%s] found failed, %s", username, err)
		return nil, err
	}
	if user.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return user, nil
}

// GetUserByNameOrEmail 根据用户名或邮箱获取用户信息
func GetUserByNameOrEmail(email string) (*admintype.Employee, error) {
	sql := dbutil.Account().Table((&admintype.Employee{}).TableName())
	if err := validate.IsEmail(email); err != nil {
		sql = sql.Where("username = ?", email)
	} else {
		sql = sql.Where("email = ?", email)
	}

	user := &admintype.Employee{}
	result := sql.Find(user)
	if err := result.Error; err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return user, nil
}

// CheckExistsEmployeeNameOrEmail .
func CheckExistsEmployeeNameOrEmail(ctx context.Context, empid uint, username, email string) (bool, error) {
	var cnt int64
	err := dbutil.Account().WithContext(ctx).Table(admintype.TableNameEmployee).
		Where("username = ? OR email = ?", username, email).
		Where("id != ?", empid).
		Count(&cnt).Error
	if err != nil {
		return false, err
	}
	if cnt > 0 {
		return true, nil
	}
	return false, nil
}

// ListEmployees 用户列表
func ListEmployees(opt apiobj.PageQuery) ([]*admintype.Employee, error) {
	sql := dbutil.Account().Table((&admintype.Employee{}).TableName()).
		Where("company_id = ?", opt.CompanyID).
		Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		sql = sql.Limit(opt.Limit)
	}

	users := []*admintype.Employee{}
	err := sql.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// ListAllEmployees 获取所有员工
func ListAllEmployees() ([]*admintype.Employee, error) {
	users := []*admintype.Employee{}
	err := dbutil.Account().Table((&admintype.Employee{}).TableName()).
		Where("status = ?", admintype.UserStatusNormal).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// QueryEmployeeList 查询用户列表
func QueryEmployeeList(ctx context.Context, opt apiobj.PageQuery, userlist *EmployeeInfoItemList) error {
	sql := dbutil.Account().WithContext(ctx).Table(admintype.TableNameEmployee).
		Where("deleted_at IS NULL")
	// Where("company_id = ?", opt.CompanyID)
	// Joins("INNER JOIN rbac_rel_prole_binding ON rbac_rel_prole_binding.user_id = core_user.id")

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "auto":
			sql = sql.Where("search_filter LIKE ?", fmt.Sprintf("%%%s%%", filter.Value[0]))
		case "username":
			sql = sql.Where(admintype.TableNameEmployee+".username LIKE ?", "%"+filter.Value[0]+"%")
		case "real_name":
			sql = sql.Where(admintype.TableNameEmployee+".real_name LIKE ?", "%"+filter.Value[0]+"%")
		case "email":
			sql = sql.Where(admintype.TableNameEmployee+".email LIKE ?", "%"+filter.Value[0]+"%")
		case "mobile":
			sql = sql.Where(admintype.TableNameEmployee+".mobile LIKE ?", "%"+filter.Value[0]+"%")
		case "status":
			sql = sql.Where(admintype.TableNameEmployee+".status = ?", filter.Value[0])
		default:
			logs.ErrorContextf(ctx, "[admin][QueryUserList] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	if err := sql.Count(&userlist.Total).Error; err != nil {
		return err
	}
	if userlist.Total == 0 {
		return nil
	}

	if len(opt.OrderBy) > 0 {
		sql = sql.Order(strings.Join(opt.OrderBy, ","))
	}

	sql = sql.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		sql = sql.Limit(opt.Limit)
	}

	err := sql.Find(&userlist.Data).Error
	if err != nil {
		return err
	}
	return nil
}

// QueryEmployee 查询用户详细信息
func QueryEmployee(ctx context.Context, id int64, user *EmployeeDetail) error {
	err := dbutil.Account().WithContext(ctx).Table(admintype.TableNameEmployee).
		Where("id = ?", id).
		Find(&user).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[admin][%d] found failed, %s", id, err)
		return err
	}
	if user == nil {
		err = fmt.Errorf("user with id %d not found", id)
		logs.ErrorContextf(ctx, "[admin][%d] find failed, %s", id, err)
		return err
	}
	return nil
}

// GetEmployeePositionsById 获取员工职位
func GetEmployeePositionsById(id int64, positions []*admintype.Position) error {
	joinstr := fmt.Sprintf("INNER JOIN %s ON %s.id = %s.position_id AND %s.deleted_at IS NULL",
		admintype.TableNameRelEmployeePosition,
		admintype.TableNamePosition,
		admintype.TableNameRelEmployeePosition,
		admintype.TableNameRelEmployeePosition,
	)
	err := dbutil.Account().Table(admintype.TableNamePosition).
		Joins(joinstr).
		Where(admintype.TableNameRelEmployeePosition+".employee_id = ?", id).
		Select(admintype.TableNamePosition + ".*").
		Find(&positions).Error
	if err != nil {
		return err
	}
	return nil
}

// UpdateEmployeeWechatInfo 更新员工的微信信息
func UpdateEmployeeWechatInfo(ctx context.Context, empID uint, wxInfo *EmployeeWechatInfo) (*admintype.Employee, error) {
	emp, err := GetEmployeeByID(ctx, empID)
	if err != nil {
		return nil, err
	}
	isExist, err := IsExistWechatUnionID(ctx, empID, wxInfo.UnionID)
	if err != nil {
		return nil, err
	}
	if isExist {
		return nil, fmt.Errorf("该微信号已经绑定了用户")
	}
	if err := dbutil.Account().Model(emp).Updates(wxInfo).Error; err != nil {
		return nil, err
	}
	return emp, nil
}

// GetEmployeeDetailByID 获取员工详情信息
func GetEmployeeDetailByID(ctx context.Context, id uint, empDetail *EmployeeDetail) error {
	emp, err := GetEmployeeByID(ctx, id)
	if err != nil {
		return err
	}
	positions, err := GetEmployeePositions(ctx, id)
	if err != nil {
		return err
	}
	empDetail.Employee = *emp
	empDetail.Positions = positions

	return nil
}

// ModifyEmployeeSimple 修改员工基本信息
func ModifyEmployeeSimple(ctx context.Context, id uint, info *EmployeeSimpleInfo) error {
	isExist, err := CheckExistsEmployeeNameOrEmail(ctx, id, info.Email, info.Phone)
	if err != nil {
		return err
	}
	if isExist {
		return fmt.Errorf("用户信息已经存在")
	}
	if err := dbutil.Account().Table(admintype.TableNameEmployee).
		Where("id = ?", id).Updates(info).Error; err != nil {
		return err
	}
	return nil
}

// IsExistWechatUnionID 是否存在微信用户ID
func IsExistWechatUnionID(ctx context.Context, empid uint, unionID string) (bool, error) {
	var cnt int64
	err := dbutil.Account().WithContext(ctx).Table(admintype.TableNameEmployee).
		Where("id != ? AND union_id = ? ", empid, unionID).
		Count(&cnt).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[admin] check wechat unionid %s failed, %s", unionID, err)
		return false, err
	}
	return cnt > 0, nil
}
