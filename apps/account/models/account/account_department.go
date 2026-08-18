package account

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

const SortGap = 1000

type DepartmentCond struct {
	BaseCond
	Filters   []apiobj.Filter
	ID        uint
	ParentID  uint
	Name      string
	CompanyID uint
	IDs       []uint
}

type DepartmentDao struct {
	BaseModel
}

func NewAccountDepartmentDao() *DepartmentDao {
	return &DepartmentDao{}
}

func (dao *DepartmentDao) TableName() string {
	return accounttype.TableNameAccountDepartment
}

func (dao *DepartmentDao) WithTx(db *gorm.DB) *DepartmentDao {
	return &DepartmentDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *DepartmentDao) Insert(ctx context.Context, entity *accounttype.AccountDepartment) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[AccountDepartmentDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *DepartmentDao) BatchInsert(ctx context.Context, entityList accounttype.AccountDepartmentList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[AccountDepartmentDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[AccountDepartmentDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *DepartmentDao) UpdateByID(ctx context.Context, id uint, entity *accounttype.AccountDepartment) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[AccountDepartmentDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *DepartmentDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[AccountDepartmentDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *DepartmentDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[AccountDepartmentDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *DepartmentDao) GetByID(ctx context.Context, id uint) (*accounttype.AccountDepartment, error) {
	var entity accounttype.AccountDepartment
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, fmt.Errorf("[AccountDepartmentDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *DepartmentDao) GetByCond(ctx context.Context, cond *DepartmentCond) (*accounttype.AccountDepartment, error) {
	var entity accounttype.AccountDepartment
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[AccountDepartmentDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *DepartmentDao) GetListByCond(ctx context.Context, cond *DepartmentCond) (accounttype.AccountDepartmentList, error) {
	var entityList accounttype.AccountDepartmentList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[AccountDepartmentDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *DepartmentDao) GetPageListByCond(ctx context.Context, cond *DepartmentCond) (accounttype.AccountDepartmentList, int64, error) {
	db := dao.DB(ctx).Model(&accounttype.AccountDepartment{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[AccountDepartmentDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList accounttype.AccountDepartmentList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[AccountDepartmentDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *DepartmentDao) CountByCond(ctx context.Context, cond *DepartmentCond) (int64, error) {
	db := dao.DB(ctx).Model(&accounttype.AccountDepartment{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[AccountDepartmentDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *DepartmentDao) BuildCondition(db *gorm.DB, cond *DepartmentCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.ParentID > 0 {
		db = db.Where(fmt.Sprintf("%s.parent_id = ?", dao.TableName()), cond.ParentID)
	}
	if cond.CompanyID > 0 {
		db = db.Where(fmt.Sprintf("%s.company_id = ?", dao.TableName()), cond.CompanyID)
	}
	if len(cond.IDs) > 0 {
		db = db.Where(fmt.Sprintf("%s.id IN (?)", dao.TableName()), cond.IDs)
	}
	if len(cond.Name) > 0 {
		db = db.Where(fmt.Sprintf("%s.name = ?", dao.TableName()), cond.Name)
	}
}

// RebalanceSiblings 对指定父部门下的所有子部门进行排序值的重新分配。
// 它会将移动请求req一并处理，确保被移动的部门最终在正确的位置。
func (dao *DepartmentDao) RebalanceSiblings(ctx context.Context, tx *gorm.DB, parentID uint, departmentId, preID, postID uint) error {
	// 1. 获取所有同级部门，但不包括正在被移动的那个
	var siblings []accounttype.AccountDepartment
	err := tx.Where("parent_id = ? AND id != ?", parentID, departmentId).
		Order("sort asc").
		Find(&siblings).Error
	if err != nil {
		return err
	}

	// 2. 获取被移动的部门本身
	var targetDept accounttype.AccountDepartment
	if err := tx.First(&targetDept, departmentId).Error; err != nil {
		return err
	}

	// 3. 构建代表最终正确顺序的部门列表
	orderedDepts := make([]*accounttype.AccountDepartment, 0, len(siblings)+1)

	// 如果移动到顶部
	if preID == 0 {
		orderedDepts = append(orderedDepts, &targetDept)
	}

	// 遍历现有兄弟节点，在正确的位置插入被移动的节点
	for i := range siblings {
		orderedDepts = append(orderedDepts, &siblings[i])
		if siblings[i].ID == preID {
			orderedDepts = append(orderedDepts, &targetDept)
		}
	}

	// 如果移动到底部
	if postID == 0 && preID != 0 {
		// 如果PreID是最后一个元素，上面循环已经加过了。这里再次检查
		if orderedDepts[len(orderedDepts)-1].ID != targetDept.ID {
			orderedDepts = append(orderedDepts, &targetDept)
		}
	}

	// 4. 遍历最终排好序的列表，分配新的、有间隔的sort值并更新数据库
	for i, dept := range orderedDepts {
		newSort := uint(i+1) * SortGap
		if dept.Sort != newSort {
			logs.DebugContextf(ctx, "Rebalancing: Dept ID %d, Name '%s', Old Sort: %d -> New Sort: %d\n", dept.ID, dept.Name, dept.Sort, newSort)
			if err := tx.Model(dept).Update("sort", newSort).Error; err != nil {
				// 如果任何一次更新失败，整个事务将回滚
				return err
			}
		}
	}

	return nil
}
