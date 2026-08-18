package forest

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type KeForestGraphCond struct {
	BaseCond
	Filters []apiobj.Filter
	ID      uint
}

type KeForestGraphDao struct {
	BaseModel
}

func NewKeForestGraphDao() *KeForestGraphDao {
	return &KeForestGraphDao{}
}

func (dao *KeForestGraphDao) TableName() string {
	return foresttype.TableNameKeForestGraph
}

func (dao *KeForestGraphDao) WithTx(db *gorm.DB) *KeForestGraphDao {
	return &KeForestGraphDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *KeForestGraphDao) Insert(ctx context.Context, entity *foresttype.ForestGraph) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KeForestGraphDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeForestGraphDao) BatchInsert(ctx context.Context, entityList foresttype.ForestGraphList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KeForestGraphDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KeForestGraphDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *KeForestGraphDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.ForestGraph) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KeForestGraphDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeForestGraphDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeForestGraphDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *KeForestGraphDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeForestGraphDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *KeForestGraphDao) GetByID(ctx context.Context, id uint) (*foresttype.ForestGraph, error) {
	var entity foresttype.ForestGraph
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeForestGraphDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *KeForestGraphDao) GetByCond(ctx context.Context, cond *KeForestGraphCond) (*foresttype.ForestGraph, error) {
	var entity foresttype.ForestGraph
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeForestGraphDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *KeForestGraphDao) GetListByCond(ctx context.Context, cond *KeForestGraphCond) (foresttype.ForestGraphList, error) {
	var entityList foresttype.ForestGraphList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KeForestGraphDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *KeForestGraphDao) GetPageListByCond(ctx context.Context, cond *KeForestGraphCond) (foresttype.ForestGraphList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.ForestGraph{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeForestGraphDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.ForestGraphList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeForestGraphDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *KeForestGraphDao) CountByCond(ctx context.Context, cond *KeForestGraphCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.ForestGraph{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KeForestGraphDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *KeForestGraphDao) BuildCondition(db *gorm.DB, cond *KeForestGraphCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
}
