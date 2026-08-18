package employee

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/pkgs/utils/validate"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// GetUserByUsername 获取用户信息
func GetUserByUsername(ctx context.Context, username string) (*accounttype.User, error) {
	user := &accounttype.User{}
	err := dbutil.Account().Table((&accounttype.User{}).TableName()).
		Where("name = ?", username).
		Find(user).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[account][%s] found failed, %s", username, err)
		return nil, err
	}
	if user.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return user, nil
}

// GetEmployeeByID 获取员工信息
func GetEmployeeByID(ctx context.Context, id uint) (*accounttype.Employee, error) {
	user := &accounttype.Employee{}
	err := dbutil.Account().Table((&accounttype.Employee{}).TableName()).
		WithContext(ctx).
		Where("id = ?", id).
		Find(user).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[account][%s] found failed, %s", id, err)
		return nil, err
	}
	if user.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return user, nil
}

// GetEmployeeByUIN 获取员工信息
func GetEmployeeByUIN(ctx context.Context, uin uint) (*accounttype.Employee, error) {
	user := &accounttype.Employee{}
	err := dbutil.Account().Table((&accounttype.Employee{}).TableName()).
		WithContext(ctx).
		Where("uin = ?", uin).
		Find(user).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[account][%v] found failed, %s", uin, err)
		return nil, err
	}
	if user.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return user, nil
}

// GetEmployeeByUsername 获取员工信息
func GetEmployeeByUsername(ctx context.Context, username string) (*accounttype.Employee, error) {
	user := &accounttype.Employee{}
	err := dbutil.Account().Table(user.TableName()).
		Where("username = ?", username).
		Find(user).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[account][%s] found failed, %s", username, err)
		return nil, err
	}
	if user.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return user, nil
}

// GetUserByNameOrEmail 根据用户名或邮箱获取用户信息
func GetUserByNameOrEmail(email string) (*accounttype.Employee, error) {
	sql := dbutil.Account().Table((&accounttype.Employee{}).TableName())
	if err := validate.IsEmail(email); err != nil {
		sql = sql.Where("username = ?", email)
	} else {
		sql = sql.Where("email = ?", email)
	}

	user := &accounttype.Employee{}
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
func CheckExistsEmployeeNameOrEmail(ctx context.Context, empID uint, username, email string) (bool, error) {
	var cnt int64
	err := dbutil.Account().Table(accounttype.TableNameEmployee).
		WithContext(ctx).
		Where("username = ? OR email = ?", username, email).
		Where("id != ?", empID).
		Count(&cnt).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CheckExistsEmployeeNameOrEmail faild err: %v", err)
		return false, err
	}
	if cnt > 0 {
		return true, nil
	}
	return false, nil
}

// ListEmployees 员工列表
func ListEmployees(opt apiobj.PageQuery) ([]*accounttype.Employee, error) {
	sql := dbutil.Account().Table((&accounttype.Employee{}).TableName()).
		Where("company_id = ?", opt.CompanyID).
		Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		sql = sql.Limit(opt.Limit)
	}

	users := []*accounttype.Employee{}
	err := sql.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// QueryEmployeeList 查询用户列表
func QueryEmployeeList(ctx context.Context, opt apiobj.PageQuery, userlist *EmployeeInfoItemList) error {
	// 基础查询：从 Employee 表开始
	sql := dbutil.Account().Table(accounttype.TableNameEmployee+" e").
		WithContext(ctx).
		Select(`
            e.*,
            u.name AS user_name,
            u.bio AS user_bio,
            u.email AS user_email,
            u.phone AS user_phone,
            ui.uin_status AS uin_status,
            ui.issuer AS issuer,
            i.real_name AS real_name,
            u.avatar_url AS avatar_url,
            i.id_card AS id_card,
            i.real_name_status AS real_name_status
        `).
		Joins("LEFT JOIN user u ON e.user_id = u.id").
		Joins("LEFT JOIN user_identification ui ON e.uin = ui.id").
		Joins("LEFT JOIN individual i ON e.user_id = i.user_id").
		Where("e.company_id = ?", opt.CompanyID).
		Where("e.deleted_at is null")

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "auto":
			sql = sql.Where("search_filter LIKE ?", fmt.Sprintf("%%%s%%", filter.Value[0]))
		case "name":
			sql = sql.Where("u.name = ?", filter.Value[0])
		case "sys_role":
			sql = sql.Where("e.sys_role = ?", filter.Value[0])
		default:
			logs.ErrorContextf(ctx, "[account][QueryUserList] invalid filter field: %s", filter.Field)
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

// QueryEmployeeSimpleList 查询用户的 uin 和用户名
func QueryEmployeeSimpleList(ctx context.Context, opt apiobj.PageQuery, userlist *EmployeeSimpleInfoItemList) error {
	sql := dbutil.Account().Table(accounttype.TableNameEmployee+" e").
		WithContext(ctx).
		Select("e.uin, e.id, u.name AS user_name,us.phone as phone,us.email as email, e.created_at,e.sys_role as sys_role").
		Joins("LEFT JOIN user_identification u ON e.user_id = u.user_id "+
			"AND (u.subject_type = ? AND u.subject_id = ?) AND u.deleted_at IS NULL AND e.uin = u.id", accounttype.SubjectTypeCompany, opt.CompanyID).
		Joins("INNER JOIN user us ON us.id = u.user_id AND us.deleted_at IS NULL").
		Where("e.company_id = ?", opt.CompanyID).
		Where("e.deleted_at IS NULL")

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "auto":
			sql = sql.Where("u.name LIKE ?", fmt.Sprintf("%%%s%%", filter.Value[0]))
		case "user_name":
			sql = sql.Where("u.name = ?", filter.Value[0])
		default:
			logs.ErrorContextf(ctx, "[account][QueryEmployeeSimpleList] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	// 处理 BeginTime 和 EndTime
	if !opt.BeginTime.IsZero() {
		sql = sql.Where("e.created_at >= ?", opt.BeginTime)
	}
	if !opt.EndTime.IsZero() {
		sql = sql.Where("e.created_at <= ?", opt.EndTime)
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

	if err := sql.Find(&userlist.Data).Error; err != nil {
		return err
	}
	return nil
}

// QueryEmployee 查询用户详细信息
func QueryEmployee(ctx context.Context, id uint, user *EmployeeDetail) error {
	err := dbutil.Account().Table(accounttype.TableNameEmployee+" e").
		Select(`
			e.*,
			ui.name AS user_name,
			u.bio AS user_bio,
			u.email AS email,
			u.phone AS phone,
			i.real_name AS real_name
		`).
		Joins("LEFT JOIN user u ON e.user_id = u.id").
		Joins("LEFT JOIN user_identification ui ON e.uin = ui.id").
		Joins("LEFT JOIN individual i ON e.user_id = i.user_id").
		Where("e.id = ?", id).
		First(&user).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[account][%d] found failed, %s", id, err)
		return err
	}
	// if user == nil {
	// 	err = fmt.Errorf("user with id %d not found", id)
	// 	logs.Errorf("[account][%d] find failed, %s", id, err)
	// 	return err
	// }
	return nil
}

// GetEmployeePositionsById 获取员工职位
func GetEmployeePositionsById(ctx context.Context, id int64, positions []*accounttype.Position) error {
	joinStr := fmt.Sprintf("INNER JOIN %s ON %s.id = %s.position_id AND %s.deleted_at IS NULL", accounttype.TableNameRelEmployeePosition, accounttype.TableNamePosition, accounttype.TableNameRelEmployeePosition, accounttype.TableNameRelEmployeePosition)
	err := dbutil.Account().Table(accounttype.TableNamePosition).
		WithContext(ctx).
		Joins(joinStr).
		Where(accounttype.TableNameRelEmployeePosition+".employee_id = ?", id).
		Select(accounttype.TableNamePosition + ".*").
		Find(&positions).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetEmployeePositionsById faild err: %v", err)
		return err
	}
	return nil
}

// UpdateEmployeeWechatInfo 更新员工的微信信息
func UpdateEmployeeWechatInfo(ctx context.Context, empID uint, wxInfo *EmployeeWechatInfo) (*accounttype.Employee, error) {
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
	err := QueryEmployee(ctx, id, empDetail)
	if err != nil {
		return err
	}
	positions, err := GetEmployeePositions(id)
	if err != nil {
		return err
	}
	// empDetail.Employee = *emp
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
		logs.ErrorContext(ctx, "ModifyEmployeeSimple failed, user info already exists")
		return fmt.Errorf("用户信息已经存在")
	}
	if err := dbutil.Account().Table(accounttype.TableNameEmployee).
		Where("id = ?", id).Updates(info).Error; err != nil {
		return err
	}
	return nil
}

// IsExistWechatUnionID 是否存在微信用户ID
func IsExistWechatUnionID(ctx context.Context, empID uint, unionID string) (bool, error) {
	var cnt int64
	err := dbutil.Account().Table(accounttype.TableNameEmployee).
		WithContext(ctx).
		Where("id != ? AND union_id = ? ", empID, unionID).
		Count(&cnt).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[account] check wechat unionid %s failed, %s", unionID, err)
		return false, err
	}
	return cnt > 0, nil
}
