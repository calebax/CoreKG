package forest

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
)

type ForestDBInstanceCond struct {
	BaseCond
	Filters  []apiobj.Filter
	ID       uint
	ForestID uint
	Enable   types.Bool
}

type ForestDBInstanceDao struct {
	BaseModel
}

func NewForestDBInstanceDao() *ForestDBInstanceDao {
	return &ForestDBInstanceDao{}
}

func (dao *ForestDBInstanceDao) TableName() string {
	return foresttype.TableNameKeForestDBInstance
}

func (dao *ForestDBInstanceDao) WithTx(db *gorm.DB) *ForestDBInstanceDao {
	return &ForestDBInstanceDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ForestDBInstanceDao) Insert(ctx context.Context, entity *foresttype.ForestDBInstance) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KeForestDbInstanceDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ForestDBInstanceDao) BatchInsert(ctx context.Context, entityList foresttype.ForestDbInstanceList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KeForestDbInstanceDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KeForestDbInstanceDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ForestDBInstanceDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.ForestDBInstance) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KeForestDbInstanceDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ForestDBInstanceDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeForestDbInstanceDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ForestDBInstanceDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_time": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeForestDbInstanceDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ForestDBInstanceDao) GetByID(ctx context.Context, id uint) (*foresttype.ForestDBInstance, error) {
	var entity foresttype.ForestDBInstance
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeForestDbInstanceDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ForestDBInstanceDao) GetByCond(ctx context.Context, cond *ForestDBInstanceCond) (*foresttype.ForestDBInstance, error) {
	var entity foresttype.ForestDBInstance
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeForestDbInstanceDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ForestDBInstanceDao) GetListByCond(ctx context.Context, cond *ForestDBInstanceCond) (foresttype.ForestDbInstanceList, error) {
	var entityList foresttype.ForestDbInstanceList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KeForestDbInstanceDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ForestDBInstanceDao) GetPageListByCond(ctx context.Context, cond *ForestDBInstanceCond) (foresttype.ForestDbInstanceList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.ForestDBInstance{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeForestDbInstanceDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.ForestDbInstanceList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeForestDbInstanceDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ForestDBInstanceDao) CountByCond(ctx context.Context, cond *ForestDBInstanceCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.ForestDBInstance{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KeForestDbInstanceDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ForestDBInstanceDao) BuildCondition(db *gorm.DB, cond *ForestDBInstanceCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.ForestID > 0 {
		query := fmt.Sprintf("%s.forest_id = ?", dao.TableName())
		db.Where(query, cond.ForestID)
	}
	if cond.Enable != 0 {
		db = db.Where(fmt.Sprintf("%s.enable = ?", dao.TableName()), cond.Enable)
	}
}
