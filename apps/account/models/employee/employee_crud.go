package employee

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
)

// CreateEmployee 创建员工
func CreateEmployee(ctx context.Context, tx *gorm.DB, employee *accounttype.Employee) error {
	return tx.WithContext(ctx).Create(employee).Error
}

// UpdateEmployee 更新用户信息
func UpdateEmployee(ctx context.Context, item UpdateEmployeeItem) error {
	db := dbutil.Account()
	old, err := GetEmployeeByID(ctx, item.EmployeeID)
	if err != nil {
		logs.ErrorContextf(ctx, "[account] get Employee by id(%d) failed, %s", item.EmployeeID, err)
		return err
	}

	if len(item.PositionIDs) > 0 {
		isExist, err := IsExistPositionIDs(db, item.PositionIDs)
		if err != nil {
			return err
		}
		if !isExist {
			return fmt.Errorf("invalid positions")
		}
	}
	userinfo, err := user.GetUserByID(old.UserID)
	if err != nil {
		logs.ErrorContextf(ctx, "[account] get user by id(%d) failed, %s", old.UserID, err)
		return err
	}
	// 更新用户信息
	if item.Username != "" {
		userinfo.Name = item.Username
	}
	if item.Email != "" {
		userinfo.Email = types.String(item.Email)
	}
	if item.Mobile != "" {
		userinfo.Phone = types.String(item.Mobile)
	}
	// individual, err := user.GetIndividual(old.Uin)
	// if err != nil {
	// 	logs.Errorf("[account] get individual by uin(%d) failed, %s", old.Uin, err)
	// 	return err
	// }
	// if item.RealName != "" {
	// 	individual.RealName = item.RealName
	// }

	tx := db.Begin()

	if err := tx.Save(old).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Save(userinfo).Error; err != nil {
		tx.Rollback()
		return err
	}
	// if err := tx.Save(individual).Error; err != nil {
	// 	tx.Rollback()
	// 	return err
	// }
	if err := tx.Unscoped().
		Delete(&accounttype.RelEmployeePosition{}, "employee_id = ?", item.EmployeeID).
		Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, positionID := range item.PositionIDs {
		rel := &accounttype.RelEmployeePosition{
			EmployeeID: item.EmployeeID,
			PositionID: positionID,
		}
		if err := tx.Create(rel).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		logs.ErrorContextf(ctx, "[account] commit transaction failed, %s", err)
		return err
	}

	return nil
}

// DeleteEmployee 删除员工
func DeleteEmployee(id uint) error {
	// 开启事务
	tx := dbutil.Account().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询员工的 uin
	var employee accounttype.Employee
	if err := tx.Table(accounttype.TableNameEmployee).
		Where("id = ?", id).
		First(&employee).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to find employee: %v", err)
	}

	// 删除 运营端员工
	if err := tx.Table(accounttype.TableNameEmployee).
		Where("id = ?", id).
		Delete(&accounttype.Employee{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete employee: %v", err)
	}

	// // 删除 Individual 个人
	// if err := tx.Table(accounttype.TableNameIndividual).
	// 	Where("uin = ?", employee.Uin).
	// 	Delete(&accounttype.Individual{}).Error; err != nil {
	// 	tx.Rollback()
	// 	return fmt.Errorf("failed to delete individual: %v", err)
	// }

	// 删除用户标识 UIN
	if err := tx.Table(accounttype.TableNameUserIdentification).
		Where("id = ?", employee.Uin).
		Delete(&accounttype.UserIdentification{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user identification: %v", err)
	}

	//delete users's resource scope
	if err := dbutil.Knownow().
		Where("scope_type = ? AND scope_id = ?",
			foresttype.ScopeTypeUser, employee.Uin).
		Delete(&foresttype.KeResourceScope{}).
		Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete keResourceScope: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// GetEmployeeByUin 根据 Uin 获取员工
func GetEmployeeByUin(uin uint) (*accounttype.Employee, error) {
	var employee accounttype.Employee
	if err := dbutil.Account().Table(accounttype.TableNameEmployee).Where("uin = ?", uin).Where("deleted_at IS NULL").First(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

// GetEmployeeByCompanyID 根据 CompanyID 获取员工
func GetEmployeeByCompanyID(companyID uint) ([]*accounttype.Employee, error) {
	var employees []*accounttype.Employee
	if err := dbutil.Account().Table(accounttype.TableNameEmployee).Where("company_id = ?", companyID).Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}

// GetCompanyEmployeeInfo 获取公司员工信息
func GetCompanyEmployeeInfo(uins []uint) ([]*CompanyEmployeeInfo, error) {
	var company []*CompanyEmployeeInfo
	err := dbutil.Account().Table(accounttype.TableNameEmployee+" e").
		Select(`
			e.*,
			c.logo as company_logo,
			c.name as company_name,
			c.company_status,
			c.user_id as company_user_id
		`).
		Joins("INNER JOIN company c ON e.company_id = c.id AND c.deleted_at IS NULL").
		Where("e.deleted_at IS NULL").
		Where("e.uin IN (?)", uins).
		Find(&company).Error
	if err != nil {
		return nil, err
	}
	return company, nil
}

// GetEmployeeByUinAndCompanyID 根据Uin 和 CompanyID 获取员工
func GetEmployeeByUinAndCompanyID(uin uint, companyID uint) (*accounttype.Employee, error) {
	var employee accounttype.Employee
	if err := dbutil.Account().Table(accounttype.TableNameEmployee).Where("uin = ?", uin).Where("company_id = ?", companyID).First(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

func CheckUinsValid(ctx context.Context, uins []uint, companyID uint) bool {
	var c int64
	if err := dbutil.Account().Table(accounttype.TableNameEmployee+" e").
		Joins("LEFT JOIN user_identification u ON e.user_id = u.user_id "+
			"AND (u.subject_type = ? AND u.subject_id = ?) AND u.deleted_at IS NULL AND e.uin = u.id", accounttype.SubjectTypeCompany, companyID).
		Joins("INNER JOIN user us ON us.id = u.user_id AND us.deleted_at IS NULL").
		Where("e.company_id = ?", companyID).
		Where("e.deleted_at IS NULL").
		Where("e.uin IN (?)", uins).
		Count(&c).Error; err != nil {
		logs.ErrorContextf(ctx, "CheckUinsValid(uins[%v],companyID[%v]) faild: %v", uins, companyID, err)
		return false
	}
	return c == int64(len(uins))
}

type AdminEmpItem struct {
	UserID     uint   `json:"user_id"`
	Uin        uint   `json:"uin"`
	EmployeeID uint   `json:"employee_id"`
	UserName   string `json:"user_name"`
}

func GetAdminEmployeeByCompanyUD(cmpID uint) (res []*AdminEmpItem, err error) {
	if err = dbutil.Account().Table(accounttype.TableNameEmployee+" e").
		Select("us.id as user_id ,e.uin as uin, e.id as employee_id, u.name AS user_name").
		Joins("LEFT JOIN user_identification u ON e.user_id = u.user_id "+
			"AND (u.subject_type = ? AND u.subject_id = ?) AND u.deleted_at IS NULL AND e.uin = u.id", accounttype.SubjectTypeCompany, cmpID).
		Joins("INNER JOIN user us ON us.id = u.user_id AND us.deleted_at IS NULL").
		Where("e.company_id = ?", cmpID).
		Where("e.sys_role = ?", accounttype.SysRoleSysAdmin).
		Where("e.deleted_at IS NULL").
		Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

// DeleteEmployeeWithUser 删除员工及用户
func DeleteEmployeeWithUser(tx *gorm.DB, id uint) error {
	return tx.Transaction(func(tx *gorm.DB) error {
		var employee accounttype.Employee
		if err := tx.Table(accounttype.TableNameEmployee).
			Where("id = ?", id).
			First(&employee).Error; err != nil {
			return fmt.Errorf("failed to find employee: %v", err)
		}

		if err := tx.Table(accounttype.TableNameEmployee).
			Where("id = ?", id).
			Delete(&accounttype.Employee{}).Error; err != nil {
			return fmt.Errorf("failed to delete employee: %v", err)
		}

		// 删除用户标识 UIN
		if err := tx.Table(accounttype.TableNameUserIdentification).
			Where("id = ?", employee.Uin).
			Delete(&accounttype.UserIdentification{}).Error; err != nil {
			return fmt.Errorf("failed to delete user identification: %v", err)
		}

		var c int64
		if err := tx.Table(accounttype.TableNameUserIdentification).
			Where("deleted_at IS NULL").
			Where("user_id = ?", employee.UserID).
			Count(&c).Error; err != nil {
			return fmt.Errorf("faild to count uin: %v", err)
		}

		if c == 0 {
			if err := tx.Where("id = ?", employee.UserID).Delete(&accounttype.User{}).Error; err != nil {
				return fmt.Errorf("failed to delete user: %v", err)
			}
		}

		//delete users's resource scope
		if err := dbutil.Knownow().
			Where("scope_type = ? AND scope_id = ?",
				foresttype.ScopeTypeUser, employee.Uin).
			Delete(&foresttype.KeResourceScope{}).
			Error; err != nil {
			return fmt.Errorf("failed to delete keResourceScope: %v", err)
		}

		return nil
	})

}
