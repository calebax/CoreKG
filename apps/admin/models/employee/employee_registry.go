package employee

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/insmtx/corekg/apps/admin/models/user"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/pkgs/utils/notify/email"
	"github.com/ygpkg/yg-go/cache/cachetype"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/random"
	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
)

const (
	verifyCodeLen                = 6
	verifyEffectiveMinutes int64 = 5

	cacheKeyVerifyCodePrefix   = "registry_verify_code"
	cacheKeyVerifyCodeRPPrefix = "reset_password_verify_code"
)

var (
	// ErrSendVerifyCodeTooBusy ,
	ErrSendVerifyCodeTooBusy = fmt.Errorf("send verify code too busy")
)

// CreateEmployee create user
func CreateEmployee(ctx context.Context, empItem *CreateEmployeeItem) (*admintype.Employee, error) {
	db := dbutil.Account()
	if err := empItem.Check(); err != nil {
		return nil, err
	}
	if len(empItem.PositionIDs) > 0 {
		isExist, err := IsExistPositionIDs(db, empItem.PositionIDs)
		if err != nil {
			logs.ErrorContextf(ctx, "CreateEmployee: IsExistPositionIDs error: %v", err)
			return nil, err
		}
		if !isExist {
			logs.ErrorContextf(ctx, "CreateEmployee: IsExistPositionIDs is false error: %v", err)
			return nil, fmt.Errorf("invalid positions")
		}
	}

	emp := &admintype.Employee{
		Username: empItem.Username,
		RealName: empItem.RealName,
		Gender:   empItem.Gender,
		Email:    &empItem.Email,
		Mobile:   &empItem.Phone,
		Status:   admintype.UserStatusNormal,
	}
	emp.FillSearchFilter()
	if empItem.Password != "" {
		passwd := user.EncryptPassword(ctx, empItem.Password)
		if passwd == nil {
			logs.ErrorContextf(ctx, "[user][CreateUser] encrypt password failed")
			return nil, errors.New("created password failed")
		}
		emp.Password = types.Password(*passwd)
	}
	//
	uin := &accounttype.UserIdentification{
		Name:        emp.NickName,
		UserID:      0,
		SubjectType: "",
		SubjectID:   0,
		UinStatus:   accounttype.UinStatusNormal,
		Issuer:      global.IssuerYYGUAdmin,
	}

	err := dbutil.Account().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(uin).Error; err != nil {
			logs.ErrorContextf(ctx, "[admin] create uin failed, %s", err)
			return err
		}
		emp.Uin = uin.ID
		if err := tx.Create(emp).Error; err != nil {
			logs.ErrorContextf(ctx, "[admin] create user (%s) failed, %s", emp.RealName, err)
			return err
		}
		for _, positionID := range empItem.PositionIDs {
			rel := &admintype.RelEmployeePosition{
				EmployeeID: emp.ID,
				PositionID: positionID,
			}
			if err := tx.Create(rel).Error; err != nil {
				logs.ErrorContextf(ctx, "[admin] create user position (%d) failed, %s", positionID, err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[admin] create user transaction failed, %s", err)
		return nil, err
	}

	return emp, nil
}

func checkCreateEmployee(tx *gorm.DB, req *admintype.Employee) error {
	sql := tx.Table((*admintype.Employee)(nil).TableName())
	if req.Email != nil && req.Mobile != nil {
		sql = sql.Where("username = ? OR mobile = ? OR email = ?",
			req.Username, *req.Mobile, *req.Email)
	} else if req.Email != nil {
		sql = sql.Where("username = ? OR email = ?", req.Username, *req.Email)
	} else if req.Mobile != nil {
		sql = sql.Where("username = ? OR mobile = ?", req.Username, *req.Mobile)
	} else {
		sql = sql.Where("username = ?", req.Username)
	}

	var count int64
	if err := sql.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("exists same username/mobile/email")
	}

	return nil
}

// generateVerifyCode .
func generateVerifyCode() string {
	return random.Number(verifyCodeLen)
}

// SendRegistryCodeByEmail 发生邮箱注册验证码
func SendRegistryCodeByEmail(ctx context.Context, che cachetype.Cache, smtpcli *email.EmailAccount, username, emailAddr string) error {
	var (
		key  = fmt.Sprintf("%s:%s:%s", cacheKeyVerifyCodePrefix, username, emailAddr)
		code = generateVerifyCode()
	)
	if che.IsExist(key) {
		return ErrSendVerifyCodeTooBusy
	}

	che.Set(key, code, time.Minute*time.Duration(verifyEffectiveMinutes))

	data := map[string]interface{}{
		"Minutes": verifyEffectiveMinutes,
		"Code":    code,
		"From":    "ROC",
	}
	body, err := generateEmailBodyTemplate(data)
	if err != nil {
		return err
	}

	if err := smtpcli.SendHTML("注册验证码", body, emailAddr); err != nil {
		logs.ErrorContextf(ctx, "[admin] send code failed, %s", err)
		return err
	}

	return nil
}

func generateEmailBodyTemplate(data map[string]interface{}) (string, error) {
	t, err := template.New("email_verify_code").Parse(emailBodyTemplate)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// // RegistryByEmail 通过邮箱注册新用户
// func RegistryByEmail(db *gorm.DB, che cache.Cache, code string, user *User) error {
// 	var (
// 		key       = fmt.Sprintf("%s:%s:%s", cacheKeyVerifyCodePrefix, user.Username, *user.Email)
// 		cacheCode = che.Get(key).(string)
// 	)
// 	if cacheCode != code {
// 		logs.Warnf("[admin] not equal, exp: %s, req: %s", cacheCode, code)
// 		return fmt.Errorf("wrong verify code")
// 	}

// 	if err := CreateUser(db, user); err != nil {
// 		logs.Errorf("[admin] insert user (%s) failed, %s", user.Username, err)
// 		return err
// 	}
// 	return nil
// }

// UpdateEmployee 更新用户信息
func UpdateEmployee(ctx context.Context, item UpdateEmployeeItem) error {
	db := dbutil.Account()
	old, err := GetEmployeeByID(ctx, item.EmployeeID)
	if err != nil {
		logs.ErrorContextf(ctx, "[admin] get user by id(%d) failed, %s", item.EmployeeID, err)
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

	// 更新用户信息
	if item.Username != "" {
		old.Username = item.Username
	}
	if item.Email != "" {
		old.Email = types.String(item.Email)
	}
	if item.Mobile != "" {
		old.Mobile = types.String(item.Mobile)
	}
	if item.RealName != "" {
		old.RealName = item.RealName
	}
	if item.Gender != 0 {
		old.Gender = item.Gender
	}

	old.FillSearchFilter()
	tx := db.Begin()

	if err := tx.Save(old).Error; err != nil {
		logs.ErrorContextf(ctx, "[admin] update user (%s) failed, %s", old.RealName, err)
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().
		Delete(&admintype.RelEmployeePosition{}, "employee_id = ?", item.EmployeeID).
		Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, positionID := range item.PositionIDs {
		rel := &admintype.RelEmployeePosition{
			EmployeeID: item.EmployeeID,
			PositionID: positionID,
		}
		if err := tx.Create(rel).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		logs.ErrorContextf(ctx, "[admin] commit transaction failed, %s", err)
		return err
	}

	return nil
}

// UpdateEmployeePassword 更新员工密码
func UpdateEmployeePassword(ctx context.Context, id uint, password string) (*admintype.Employee, error) {
	emp, err := GetEmployeeByID(ctx, id)
	if err != nil {
		logs.ErrorContextf(ctx, "[admin] get user by id(%d) failed, %s", id, err)
		return nil, err
	}
	passwd := user.EncryptPassword(ctx, password)
	if passwd == nil {
		logs.ErrorContextf(ctx, "[user][UpdateEmployeePassword] encrypt password failed")
		return nil, err
	}
	emp.Password = types.Password(*passwd)
	if err := dbutil.Account().Save(emp).Error; err != nil {
		logs.ErrorContextf(ctx, "[admin] update user (%s) password failed, %s", emp.RealName, err)
		return nil, err
	}
	return emp, nil
}

// DeleteEmployee 删除用户
func DeleteEmployee(ctx context.Context, userID uint) error {
	err := dbutil.Account().Table(admintype.TableNameEmployee).Where("id = ?", userID).Delete(&admintype.Employee{}).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[admin] delete user (%d) failed, %s", userID, err)
		return err
	}
	return nil
}
