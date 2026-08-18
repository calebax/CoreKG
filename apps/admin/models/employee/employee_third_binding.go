package employee

import (
	"context"

	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
)

// // GetUserByBindingInfo 通过绑定信息获取用户
// func GetUserByBindingInfo(bindtype, bindvalue string) (*admintype.Employee, error) {
// 	user := &admintype.Employee{}
// 	joinstr := fmt.Sprintf("LEFT JOIN `%s` ON `%s`.employee_id = `%s`.id",
// 		admintype.TableNameEmployeeThirdBinding,
// 		admintype.TableNameEmployeeThirdBinding,
// 		admintype.TableNameEmployee,
// 	)
// 	err := dbutil.Account().Table(admintype.TableNameEmployee).
// 		Joins(joinstr).
// 		Where("bind_type = ? AND bind_value = ? ", bindtype, bindvalue).
// 		Find(user).Error
// 	if err != nil {
// 		logs.Errorf("[admin] found %s:%s failed, %s", bindtype, bindvalue, err)
// 		return nil, err
// 	}
// 	return user, nil
// }

// GetUserByWechatUnionID 通过微信用户ID获取用户
func GetUserByWechatUnionID(ctx context.Context, unionID string) (*admintype.Employee, error) {
	emp := &admintype.Employee{}
	err := dbutil.Account().Table(admintype.TableNameEmployee).
		Where("union_id = ? AND deleted_at IS NULL", unionID).
		First(emp).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[admin] found wechat unionid %s failed, %s", unionID, err)
		return nil, err
	}
	return emp, nil
}

// GetEmployeeByWechatComUserID 通过企业微信用户ID获取用户
func GetEmployeeByWechatComUserID(ctx context.Context, userid string) (*admintype.Employee, error) {
	emp := &admintype.Employee{}
	err := dbutil.Account().Table(admintype.TableNameEmployee).
		Where("wecom_user_id = ? AND deleted_at IS NULL", userid).
		First(emp).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[admin] found wechatcom userid(%s), failed %s", userid, err)
		return nil, err
	}
	return emp, nil
}
