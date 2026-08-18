package employee

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

func IsExistPositionIDs(db *gorm.DB, ids []uint) (bool, error) {
	var count int64
	if err := db.Table(accounttype.TableNamePosition).Where("deleted_at is null").
		Where("id in (?)", ids).Count(&count).Error; err != nil {
		return false, err
	}
	if count != int64(len(ids)) {
		return false, nil
	}
	return true, nil
}

// IsExistPositionName 查询职位名称是否存在
func IsExistPositionName(name string, exceptIDs ...uint) (bool, error) {
	p := &accounttype.Position{}
	db := dbutil.Account()
	query := db.Table(accounttype.TableNamePosition).Where("deleted_at is null").
		Where("name = ?", name)
	if len(exceptIDs) > 0 {
		query = query.Where("id not in (?)", exceptIDs)
	}

	err := query.First(p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type PositionListItem accounttype.Position

type QueryPositionListResponse struct {
	apiobj.QueryResponse
	Data []*PositionListItem
}

// CreatePositionOption 新增职位参数
type CreatePositionOption struct {
	accounttype.Position
	PrivilegeIDs []uint `json:"privilege_ids"`
}

// PositionDetail 职位及权限详情
type PositionDetail struct {
	accounttype.Position
	//Privileges   []*accounttype.Privilege `json:"privileges"`
	PrivilegeIDs []uint `json:"privilege_ids"`
}

// QueryPositionList 查询职位列表
func QueryPositionList(ctx context.Context, opt apiobj.PageQuery, positionList *QueryPositionListResponse) error {
	query := dbutil.Account().Table(accounttype.TableNamePosition).
		WithContext(ctx).
		Where("deleted_at is null").
		Where("company_id = ?", opt.CompanyID)
	// Joins("INNER JOIN rbac_rel_prole_binding ON rbac_rel_prole_binding.user_id = core_user.id")

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "name":
			query = query.Where(accounttype.TableNamePosition+".name LIKE ?", "%"+filter.Value[0]+"%")
		default:
			logs.WarnContextf(ctx, "[account][QueryPositionList] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	if err := query.Count(&positionList.Total).Error; err != nil {
		return err
	}
	if positionList.Total == 0 {
		return nil
	}

	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}

	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}

	err := query.Find(&positionList.Data).Error
	if err != nil {
		return err
	}
	return nil
}

// getAllPrivilege 获取全部权限
func getAllPrivilege() ([]*accounttype.Privilege, error) {
	out := []*accounttype.Privilege{}
	if err := dbutil.Account().Table(accounttype.TableNamePrivilege).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// GetPositionUser 获取指定职位的用户信息
func GetPositionUser(positionIDs uint) ([]*accounttype.Employee, error) {
	var employees []*accounttype.Employee

	err := dbutil.Account().Table(accounttype.TableNameRelEmployeePosition).
		Select(accounttype.TableNameEmployee+".*").
		Joins("INNER JOIN "+accounttype.TableNameEmployee+" ON "+accounttype.TableNameEmployee+".id = "+accounttype.TableNameRelEmployeePosition+".employee_id").
		Where(accounttype.TableNameRelEmployeePosition+".position_id = ?", positionIDs).
		Where(accounttype.TableNameEmployee + ".deleted_at IS NULL").
		Where(accounttype.TableNameRelEmployeePosition + ".deleted_at IS NULL").
		Find(&employees).Error
	if err != nil {
		return nil, err
	}

	return employees, nil
}

// GetEmployeePositions 获取员工职位列表
func GetEmployeePositions(empID uint) ([]*accounttype.Position, error) {
	out := make([]*accounttype.Position, 0)
	joinstr := fmt.Sprintf("INNER JOIN `%s` ON `%s`.id = `%s`.position_id AND `%s`.deleted_at IS NULL",
		accounttype.TableNameRelEmployeePosition,
		accounttype.TableNamePosition,
		accounttype.TableNameRelEmployeePosition,
		accounttype.TableNameRelEmployeePosition,
	)
	err := dbutil.Account().Table(accounttype.TableNamePosition).
		Joins(joinstr).
		Where(accounttype.TableNameRelEmployeePosition+".employee_id = ?", empID).
		Select(accounttype.TableNamePosition + ".*").
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetPositionPrivileges 获取职位的权限
func GetPositionPrivileges(positionIDs ...uint) ([]*accounttype.Privilege, error) {
	out := make([]*accounttype.Privilege, 0)
	joinstr := fmt.Sprintf("INNER JOIN `%s` ON `%s`.id = `%s`.privilege_id AND `%s`.deleted_at IS NULL",
		accounttype.TableNameRelPositionPrivilege,
		accounttype.TableNamePrivilege,
		accounttype.TableNameRelPositionPrivilege,
		accounttype.TableNameRelPositionPrivilege,
	)
	err := dbutil.Account().Table(accounttype.TableNamePrivilege).
		Joins(joinstr).
		Where(accounttype.TableNamePrivilege+".deleted_at is null").
		Where(accounttype.TableNameRelPositionPrivilege+".position_id in (?)", positionIDs).
		Select(accounttype.TableNamePrivilege + ".*").
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetPositionPrivilegeIDs 获取职位的权限id数组
func GetPositionPrivilegeIDs(ctx context.Context, positionIDs ...uint) ([]uint, error) {
	out := make([]uint, 0)
	joinStr := fmt.Sprintf("INNER JOIN `%s` ON `%s`.id = `%s`.privilege_id AND `%s`.deleted_at IS NULL",
		accounttype.TableNameRelPositionPrivilege,
		accounttype.TableNamePrivilege,
		accounttype.TableNameRelPositionPrivilege,
		accounttype.TableNameRelPositionPrivilege,
	)
	err := dbutil.Account().Table(accounttype.TableNamePrivilege).
		WithContext(ctx).
		Joins(joinStr).
		Where(accounttype.TableNamePrivilege+".deleted_at is null").
		Where(accounttype.TableNameRelPositionPrivilege+".position_id in (?)", positionIDs).
		Pluck(accounttype.TableNamePrivilege+".id", &out).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetPositionPrivilegeIDs faild err: %v", err)
		return nil, err
	}
	return out, nil
}

// GetEmployeePrivileges 获取员工权限
func GetEmployeePrivileges(empID uint) ([]*accounttype.Position, []*accounttype.Privilege, error) {
	positions, err := GetEmployeePositions(empID)
	if err != nil {
		return nil, nil, fmt.Errorf("获取用户职位信息失败: %v", err)
	}
	ps := []*accounttype.Privilege{}
	for _, p := range positions {
		privileges, err := GetPositionPrivileges(p.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("获取用户职位信息失败: %v", err)
		}
		ps = append(ps, privileges...)
	}
	return positions, ps, nil
}

// HasEmployeeApiPrivilege  员工是否有调用API的权限
func HasEmployeeApiPrivilege(empID uint, uri string) (bool, error) {
	positions, err := GetEmployeePositions(empID)
	if err != nil {
		return false, fmt.Errorf("获取用户职位信息失败: %v", err)
	}
	if len(positions) == 0 {
		return false, nil
	}

	var positionIDs []uint
	for _, p := range positions {
		positionIDs = append(positionIDs, p.ID)
	}

	joinstr := fmt.Sprintf("INNER JOIN `%s` ON `%s`.id = `%s`.privilege_id AND `%s`.deleted_at IS NULL",
		accounttype.TableNameRelPositionPrivilege,
		accounttype.TableNamePrivilege,
		accounttype.TableNameRelPositionPrivilege,
		accounttype.TableNameRelPositionPrivilege,
	)

	var count int64
	err = dbutil.Account().Table(accounttype.TableNamePrivilege).
		Where(accounttype.TableNamePrivilege+".deleted_at is null").
		Joins(joinstr).
		Where(accounttype.TableNameRelPositionPrivilege+".position_id in (?)", positionIDs).
		Where(accounttype.TableNamePrivilege+".api = (?)", uri).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

// GetEmployeeRbac 返回员工职位信息、权限信息、action信息
func GetEmployeeRbac(emp *accounttype.Employee) ([]*accounttype.Position, []*accounttype.Privilege, []string, error) {
	var (
		positions   = make([]*accounttype.Position, 0)
		privileges  = make([]*accounttype.Privilege, 0)
		actionPaths = make([]string, 0)
		err         error
	)
	if emp.SysRole == accounttype.SysRoleSysAdmin {
		// positions, err = GetEmployeePositions(emp.ID)
		// if err != nil {
		// 	return nil, nil, nil, err
		// }
		// privileges, err = getAllPrivilege()
		// if err != nil {
		// 	return nil, nil, nil, err
		// }
		positions, privileges, err = GetEmployeePrivileges(emp.ID)
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		positions, privileges, err = GetEmployeePrivileges(emp.ID)
		if err != nil {

			return nil, nil, nil, err
		}
	}

	for _, p := range privileges {
		actionPaths = append(actionPaths, p.ActionPath)
	}
	return positions, privileges, actionPaths, nil
}

type EmployeeLoginInfo struct {
	EmployeeDetail
}

type BindEmployeeWechatResponse struct {
	apiobj.BaseResponse

	Response struct {
		JwtToken string            `json:"jwt_token,omitempty"`
		UserInfo EmployeeLoginInfo `json:"user_info"`
	}
}

// GetEmployeeByUnionID 根据UnionID获取员工信息
func GetEmployeeByUnionID(unionID string) (*accounttype.Employee, error) {
	employee := &accounttype.Employee{}
	err := dbutil.Account().Table(accounttype.TableNameEmployee).Where("union_id = ?", unionID).First(employee).Error
	if err != nil {
		return nil, err
	}
	return employee, nil
}

func GetEmployeeLoginInfo(unionId string, info *EmployeeLoginInfo, resp *apiobj.BaseResponse) error {
	emp, err := GetEmployeeByUnionID(unionId)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("获取用户信息失败: %v", err)
		return err
	}
	positions, _, actionPaths, err := GetEmployeeRbac(emp)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("获取用户权限失败: %v", err)
		return err
	}
	info.EmployeeDetail.Employee = *emp
	info.EmployeeDetail.Positions = positions
	info.ActionPaths = actionPaths
	return nil
}

func IsExistPrivilegeIDs(ids []uint) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	var count int64
	db := dbutil.Account()
	if err := db.Table(accounttype.TableNamePrivilege).Where("deleted_at is null").
		Where("id in (?)", ids).Count(&count).Error; err != nil {
		return false, err
	}
	if count != int64(len(ids)) {
		return false, nil
	}
	return true, nil
}

// CreatePosition 添加职位及权限
func CreatePosition(opt *CreatePositionOption) (*accounttype.Position, error) {
	db := dbutil.Account()
	if opt.Name == "" {
		return nil, errors.New("请先填写职位名称")
	}
	isExist, err := IsExistPositionName(opt.Name)
	if err != nil {
		return nil, err
	}
	if isExist {
		return nil, errors.New("该职位名称已经存在")
	}

	if len(opt.PrivilegeIDs) > 0 {
		isExist, err := IsExistPrivilegeIDs(opt.PrivilegeIDs)
		if err != nil {
			return nil, err
		}
		if !isExist {
			return nil, errors.New("权限有误")
		}
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&opt.Position).Error; err != nil {
			return err
		}
		if len(opt.PrivilegeIDs) > 0 {
			rel := make([]*accounttype.RelPositionPrivilege, 0)
			for _, p := range opt.PrivilegeIDs {
				rel = append(rel, &accounttype.RelPositionPrivilege{
					PositionID:  opt.Position.ID,
					PrivilegeID: p,
				})
			}
			if err := tx.Create(rel).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &opt.Position, nil
}

// GetPositionByID 根据ID获取职位信息
func GetPositionByID(id uint) (*accounttype.Position, error) {
	position := &accounttype.Position{}
	err := dbutil.Account().Table(accounttype.TableNamePosition).Where("id = ?", id).First(position).Error
	if err != nil {
		return nil, err
	}
	return position, nil
}

// GetPositionDetailByID 获取职位及权限详情
func GetPositionDetailByID(ctx context.Context, id uint) (*PositionDetail, error) {
	out := &PositionDetail{}
	p, err := GetPositionByID(id)
	if err != nil {
		return nil, err
	}
	pris, err := GetPositionPrivilegeIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	out.Position = *p
	out.PrivilegeIDs = pris

	return out, nil
}

// ModifyPosition 修改职位信息
func ModifyPosition(id uint, p *accounttype.Position) (*accounttype.Position, error) {
	isExist, err := IsExistPositionName(p.Name, id)
	if err != nil {
		return nil, err
	}
	if isExist {
		return nil, errors.New("该职位名称已经存在")
	}

	position, err := GetPositionByID(id)
	if err != nil {
		return nil, err
	}
	position.Name = p.Name
	position.Description = p.Description

	if err := dbutil.Account().Save(position).Error; err != nil {
		return nil, err
	}
	return position, nil
}

// ModifyPositionPrivilege 修改职位权限信息
func ModifyPositionPrivilege(id uint, privilegeIDs []uint) (*accounttype.Position, error) {
	position, err := GetPositionByID(id)
	if err != nil {
		return nil, err
	}

	isExist, err := IsExistPrivilegeIDs(privilegeIDs)
	if err != nil {
		return nil, err
	}
	if !isExist {
		return nil, errors.New("权限有误")
	}
	rel := make([]*accounttype.RelPositionPrivilege, 0)
	for _, privilegeID := range privilegeIDs {
		rel = append(rel, &accounttype.RelPositionPrivilege{
			PositionID:  position.ID,
			PrivilegeID: privilegeID,
		})
	}

	err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
		err := tx.Table(accounttype.TableNameRelPositionPrivilege).Unscoped().
			Where("position_id = ?", position.ID).
			Delete(&accounttype.RelPositionPrivilege{}).Error
		if err != nil {
			return err
		}
		if len(rel) > 0 {
			if err := tx.Create(rel).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return position, nil
}

// DeletePosition 删除职位及职位权限信息
func DeletePosition(id uint) error {
	position, err := GetPositionByID(id)
	if err != nil {
		return err
	}

	err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(position).Error; err != nil {
			return err
		}

		err := tx.Table(accounttype.TableNameRelPositionPrivilege).
			Where("position_id = ?", position.ID).
			Delete(&accounttype.RelPositionPrivilege{}).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
