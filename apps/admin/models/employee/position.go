package employee

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

func IsExistPositionIDs(db *gorm.DB, ids []uint) (bool, error) {
	var count int64
	if err := db.Table(admintype.TableNamePosition).Where("deleted_at is null").
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
	p := &admintype.Position{}
	db := dbutil.Account()
	query := db.Table(admintype.TableNamePosition).Where("deleted_at is null").
		Where("name = ?", name)
	if len(exceptIDs) > 0 {
		query = query.Where("id not in (?)", exceptIDs)
	}

	err := query.First(p).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type PositionListItem admintype.Position

type QueryPositionListResponse struct {
	apiobj.QueryResponse
	Data []*PositionListItem
}

// CreatePositionOption 新增职位参数
type CreatePositionOption struct {
	admintype.Position
	PrivilegeIDs []uint `json:"privilege_ids"`
}

// PositionDetail 职位及权限详情
type PositionDetail struct {
	admintype.Position
	//Privileges   []*admintype.Privilege `json:"privileges"`
	PrivilegeIDs []uint `json:"privilege_ids"`
}

// QueryPositionList 查询职位列表
func QueryPositionList(ctx context.Context, opt apiobj.PageQuery, positionList *QueryPositionListResponse) error {
	query := dbutil.Account().Table(admintype.TableNamePosition).
		WithContext(ctx).
		Where("deleted_at is null")
	// Joins("INNER JOIN rbac_rel_prole_binding ON rbac_rel_prole_binding.user_id = core_user.id")

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "name":
			query = query.Where(admintype.TableNamePosition+".name LIKE ?", "%"+filter.Value[0]+"%")
		default:
			logs.WarnContextf(ctx, "[admin][QueryPositionList] invalid filter field: %s", filter.Field)
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
func getAllPrivilege() ([]*admintype.Privilege, error) {
	out := []*admintype.Privilege{}
	if err := dbutil.Account().Table(admintype.TableNamePrivilege).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// GetPositionUser 获取指定职位的用户信息
func GetPositionUser(positionIDs uint) ([]*admintype.Employee, error) {
	var employees []*admintype.Employee

	err := dbutil.Account().Table(admintype.TableNameRelEmployeePosition).
		Select(admintype.TableNameEmployee+".*").
		Joins("INNER JOIN "+admintype.TableNameEmployee+" ON "+admintype.TableNameEmployee+".id = "+admintype.TableNameRelEmployeePosition+".employee_id").
		Where(admintype.TableNameRelEmployeePosition+".position_id = ?", positionIDs).
		Where(admintype.TableNameEmployee + ".deleted_at IS NULL").
		Where(admintype.TableNameRelEmployeePosition + ".deleted_at IS NULL").
		Find(&employees).Error
	if err != nil {
		return nil, err
	}

	return employees, nil
}

// GetEmployeePositions 获取员工职位列表
func GetEmployeePositions(ctx context.Context, empID uint) ([]*admintype.Position, error) {
	out := make([]*admintype.Position, 0)
	joinStr := fmt.Sprintf("INNER JOIN `%s` ON `%s`.id = `%s`.position_id AND `%s`.deleted_at IS NULL",
		admintype.TableNameRelEmployeePosition,
		admintype.TableNamePosition,
		admintype.TableNameRelEmployeePosition,
		admintype.TableNameRelEmployeePosition,
	)
	err := dbutil.Account().Table(admintype.TableNamePosition).
		WithContext(ctx).
		Joins(joinStr).
		Where(admintype.TableNameRelEmployeePosition+".employee_id = ?", empID).
		Select(admintype.TableNamePosition + ".*").
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetPositionPrivileges 获取职位的权限
func GetPositionPrivileges(positionIDs ...uint) ([]*admintype.Privilege, error) {
	out := make([]*admintype.Privilege, 0)
	joinstr := fmt.Sprintf("INNER JOIN `%s` ON `%s`.id = `%s`.privilege_id AND `%s`.deleted_at IS NULL",
		admintype.TableNameRelPositionPrivilege,
		admintype.TableNamePrivilege,
		admintype.TableNameRelPositionPrivilege,
		admintype.TableNameRelPositionPrivilege,
	)
	err := dbutil.Account().Table(admintype.TableNamePrivilege).
		Joins(joinstr).
		Where(admintype.TableNamePrivilege+".deleted_at is null").
		Where(admintype.TableNameRelPositionPrivilege+".position_id in (?)", positionIDs).
		Select(admintype.TableNamePrivilege + ".*").
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetPositionPrivilegeIDs 获取职位的权限id数组
func GetPositionPrivilegeIDs(positionIDs ...uint) ([]uint, error) {
	out := make([]uint, 0)
	joinstr := fmt.Sprintf("INNER JOIN `%s` ON `%s`.id = `%s`.privilege_id AND `%s`.deleted_at IS NULL",
		admintype.TableNameRelPositionPrivilege,
		admintype.TableNamePrivilege,
		admintype.TableNameRelPositionPrivilege,
		admintype.TableNameRelPositionPrivilege,
	)
	err := dbutil.Account().Table(admintype.TableNamePrivilege).
		Joins(joinstr).
		Where(admintype.TableNamePrivilege+".deleted_at is null").
		Where(admintype.TableNameRelPositionPrivilege+".position_id in (?)", positionIDs).
		Pluck(admintype.TableNamePrivilege+".id", &out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetEmployeePrivileges 获取员工权限
func GetEmployeePrivileges(ctx context.Context, empID uint) ([]*admintype.Position, []*admintype.Privilege, error) {
	positions, err := GetEmployeePositions(ctx, empID)
	if err != nil {
		return nil, nil, fmt.Errorf("获取用户职位信息失败: %v", err)
	}
	ps := []*admintype.Privilege{}
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
func HasEmployeeApiPrivilege(ctx context.Context, empID uint, uri string) (bool, error) {
	positions, err := GetEmployeePositions(ctx, empID)
	if err != nil {
		return false, fmt.Errorf("获取用户职位信息失败: %v", err)
	}
	if len(positions) == 0 {
		return false, nil
	}

	positionIDs := []uint{}
	for _, p := range positions {
		positionIDs = append(positionIDs, p.ID)
	}

	joinstr := fmt.Sprintf("INNER JOIN `%s` ON `%s`.id = `%s`.privilege_id AND `%s`.deleted_at IS NULL",
		admintype.TableNameRelPositionPrivilege,
		admintype.TableNamePrivilege,
		admintype.TableNameRelPositionPrivilege,
		admintype.TableNameRelPositionPrivilege,
	)

	var count int64
	err = dbutil.Account().Table(admintype.TableNamePrivilege).
		WithContext(ctx).
		Where(admintype.TableNamePrivilege+".deleted_at is null").
		Joins(joinstr).
		Where(admintype.TableNameRelPositionPrivilege+".position_id in (?)", positionIDs).
		Where(admintype.TableNamePrivilege+".api = (?)", uri).
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
func GetEmployeeRbac(ctx context.Context, emp *admintype.Employee) ([]*admintype.Position, []*admintype.Privilege, []string, error) {
	var (
		positions   = make([]*admintype.Position, 0)
		privileges  = make([]*admintype.Privilege, 0)
		actionPaths = make([]string, 0)
		err         error
	)
	if emp.SysRole == admintype.SysRoleSysAdmin {
		positions, err = GetEmployeePositions(ctx, emp.ID)
		if err != nil {
			return nil, nil, nil, err
		}
		privileges, err = getAllPrivilege()
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		positions, privileges, err = GetEmployeePrivileges(ctx, emp.ID)
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
func GetEmployeeByUnionID(unionID string) (*admintype.Employee, error) {
	employee := &admintype.Employee{}
	err := dbutil.Account().Table(admintype.TableNameEmployee).Where("union_id = ?", unionID).First(employee).Error
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
	positions, _, actionPaths, err := GetEmployeeRbac(context.TODO(), emp)
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
	if err := db.Table(admintype.TableNamePrivilege).Where("deleted_at is null").
		Where("id in (?)", ids).Count(&count).Error; err != nil {
		return false, err
	}
	if count != int64(len(ids)) {
		return false, nil
	}
	return true, nil
}

// CreatePosition 添加职位及权限
func CreatePosition(opt *CreatePositionOption) (*admintype.Position, error) {
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
			rel := make([]*admintype.RelPositionPrivilege, 0)
			for _, p := range opt.PrivilegeIDs {
				rel = append(rel, &admintype.RelPositionPrivilege{
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
func GetPositionByID(id uint) (*admintype.Position, error) {
	position := &admintype.Position{}
	err := dbutil.Account().Table(admintype.TableNamePosition).Where("id = ?", id).First(position).Error
	if err != nil {
		return nil, err
	}
	return position, nil
}

// GetPositionDetail 获取职位及权限详情
func GetPositionDetailByID(id uint) (*PositionDetail, error) {
	out := &PositionDetail{}
	p, err := GetPositionByID(id)
	if err != nil {
		return nil, err
	}
	pris, err := GetPositionPrivilegeIDs(id)
	if err != nil {
		return nil, err
	}
	out.Position = *p
	out.PrivilegeIDs = pris

	return out, nil
}

// ModifyPosition 修改职位信息
func ModifyPosition(id uint, p *admintype.Position) (*admintype.Position, error) {
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
func ModifyPositionPrivilege(id uint, privilegeIDs []uint) (*admintype.Position, error) {
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
	rel := make([]*admintype.RelPositionPrivilege, 0)
	for _, privilegeID := range privilegeIDs {
		rel = append(rel, &admintype.RelPositionPrivilege{
			PositionID:  position.ID,
			PrivilegeID: privilegeID,
		})
	}

	err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
		err := tx.Table(admintype.TableNameRelPositionPrivilege).Unscoped().
			Where("position_id = ?", position.ID).
			Delete(&admintype.RelPositionPrivilege{}).Error
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

		err := tx.Table(admintype.TableNameRelPositionPrivilege).
			Where("position_id = ?", position.ID).
			Delete(&admintype.RelPositionPrivilege{}).Error
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
