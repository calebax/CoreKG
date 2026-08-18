package account

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/account/internal/dto/dtoorganize"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type RelEmployeeDepartmentCond struct {
	BaseCond
	Filters      []apiobj.Filter
	ID           uint
	Uin          uint
	DepartmentID uint
	CompanyID    uint
}

type RelEmployeeDepartmentDao struct {
	BaseModel
}

func NewAccountRelEmployeeDepartmentDao() *RelEmployeeDepartmentDao {
	return &RelEmployeeDepartmentDao{}
}

func (dao *RelEmployeeDepartmentDao) TableName() string {
	return accounttype.TableNameAccountRelEmployeeDepartment
}

func (dao *RelEmployeeDepartmentDao) WithTx(db *gorm.DB) *RelEmployeeDepartmentDao {
	return &RelEmployeeDepartmentDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *RelEmployeeDepartmentDao) Insert(ctx context.Context, entity *accounttype.AccountRelEmployeeDepartment) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[AccountRelEmployeeDepartmentDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *RelEmployeeDepartmentDao) BatchInsert(ctx context.Context, entityList accounttype.AccountRelEmployeeDepartmentList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[AccountRelEmployeeDepartmentDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[AccountRelEmployeeDepartmentDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *RelEmployeeDepartmentDao) UpdateByID(ctx context.Context, id uint, entity *accounttype.AccountRelEmployeeDepartment) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[AccountRelEmployeeDepartmentDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *RelEmployeeDepartmentDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[AccountRelEmployeeDepartmentDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *RelEmployeeDepartmentDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[AccountRelEmployeeDepartmentDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *RelEmployeeDepartmentDao) GetByID(ctx context.Context, id uint) (*accounttype.AccountRelEmployeeDepartment, error) {
	var entity accounttype.AccountRelEmployeeDepartment
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[AccountRelEmployeeDepartmentDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *RelEmployeeDepartmentDao) GetByCond(ctx context.Context, cond *RelEmployeeDepartmentCond) (*accounttype.AccountRelEmployeeDepartment, error) {
	var entity accounttype.AccountRelEmployeeDepartment
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[AccountRelEmployeeDepartmentDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *RelEmployeeDepartmentDao) GetListByCond(ctx context.Context, cond *RelEmployeeDepartmentCond) (accounttype.AccountRelEmployeeDepartmentList, error) {
	var entityList accounttype.AccountRelEmployeeDepartmentList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[AccountRelEmployeeDepartmentDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *RelEmployeeDepartmentDao) GetPageListByCond(ctx context.Context, cond *RelEmployeeDepartmentCond) (accounttype.AccountRelEmployeeDepartmentList, int64, error) {
	db := dao.DB(ctx).Model(&accounttype.AccountRelEmployeeDepartment{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[AccountRelEmployeeDepartmentDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList accounttype.AccountRelEmployeeDepartmentList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[AccountRelEmployeeDepartmentDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *RelEmployeeDepartmentDao) CountByCond(ctx context.Context, cond *RelEmployeeDepartmentCond) (int64, error) {
	db := dao.DB(ctx).Model(&accounttype.AccountRelEmployeeDepartment{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[AccountRelEmployeeDepartmentDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *RelEmployeeDepartmentDao) BuildCondition(db *gorm.DB, cond *RelEmployeeDepartmentCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.CompanyID > 0 {
		db = db.Where(fmt.Sprintf("%s.company_id = ?", dao.TableName()), cond.CompanyID)
	}
	if cond.DepartmentID > 0 {
		db = db.Where(fmt.Sprintf("%s.department_id = ?", dao.TableName()), cond.DepartmentID)
	}
	if cond.Uin > 0 {
		db = db.Where(fmt.Sprintf("%s.uin = ?", dao.TableName()), cond.Uin)
	}

}

// GetCompanyEmployInfo will return current company all employees' base info
func (dao *RelEmployeeDepartmentDao) GetCompanyEmployInfo(ctx context.Context, companyID uint) (res []dtoorganize.EmployeeInfo, err error) {
	if err = dao.DB(ctx).Table(accounttype.TableNameEmployee+" e").
		Select("e.uin as uin, e.id as employee_id,us.name as user_name, u.name AS name,us.phone as phone,us.email as email, e.created_at as created_at,e.sys_role as sys_role").
		Joins("LEFT JOIN user_identification u ON e.user_id = u.user_id "+
			"AND (u.subject_type = ? AND u.subject_id = ?) AND u.deleted_at IS NULL AND e.uin = u.id", accounttype.SubjectTypeCompany, companyID).
		Joins("INNER JOIN user us ON us.id = u.user_id AND us.deleted_at IS NULL").
		Where("e.company_id = ?", companyID).
		Where("e.deleted_at IS NULL").
		Find(&res).
		Error; err != nil {
		logs.ErrorContextf(ctx, "get employee info fail, company_id:%d, err:%v", companyID, err)
		return nil, err
	}
	return res, nil
}
