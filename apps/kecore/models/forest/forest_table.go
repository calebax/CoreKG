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

type ForestTableCond struct {
	BaseCond
	Filters       []apiobj.Filter
	ID            uint
	IDs           []uint
	ForestIDs     []uint
	TableNameLike string
}

type ForestTableDao struct {
	BaseModel
}

func NewForestTableDao() *ForestTableDao {
	return &ForestTableDao{}
}

func (dao *ForestTableDao) TableName() string {
	return foresttype.TableNameKeForestTable
}

func (dao *ForestTableDao) WithTx(db *gorm.DB) *ForestTableDao {
	return &ForestTableDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ForestTableDao) Insert(ctx context.Context, entity *foresttype.ForestTable) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KeForestTableDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ForestTableDao) BatchInsert(ctx context.Context, entityList foresttype.ForestTableList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KeForestTableDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KeForestTableDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ForestTableDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.ForestTable) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KeForestTableDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ForestTableDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeForestTableDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ForestTableDao) UpdateForestIsDMap(ctx context.Context, id []uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("forest_id IN (?)", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeForestTableDao] UpdateForestIsDMap fail, id:%s, updateMap:%s, err: %v", logs.JSON(id), logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ForestTableDao) UpdateMapByIDs(ctx context.Context, id []uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id IN (?)", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeForestTableDao] UpdateMapByIDs fail, id:%s, updateMap:%s, err: %v", logs.JSON(id), logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ForestTableDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeForestTableDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ForestTableDao) DeleteByForestID(ctx context.Context, forestID uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("forest_id = ?", forestID).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeForestTableDao] DeleteByForestID fail, forestID:%d, err: %v", forestID, err)
	}
	return nil
}

func (dao *ForestTableDao) GetByID(ctx context.Context, id uint) (*foresttype.ForestTable, error) {
	var entity foresttype.ForestTable
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeForestTableDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ForestTableDao) GetByCond(ctx context.Context, cond *ForestTableCond) (*foresttype.ForestTable, error) {
	var entity foresttype.ForestTable
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeForestTableDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ForestTableDao) GetListByCond(ctx context.Context, cond *ForestTableCond) (foresttype.ForestTableList, error) {
	var entityList foresttype.ForestTableList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KeForestTableDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ForestTableDao) GetPageListByCond(ctx context.Context, cond *ForestTableCond) (foresttype.ForestTableList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.ForestTable{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeForestTableDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.ForestTableList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeForestTableDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ForestTableDao) CountByCond(ctx context.Context, cond *ForestTableCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.ForestTable{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KeForestTableDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ForestTableDao) BuildCondition(db *gorm.DB, cond *ForestTableCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if len(cond.IDs) > 0 {
		query := fmt.Sprintf("%s.id IN ?", dao.TableName())
		db.Where(query, cond.IDs)
	}
	if len(cond.ForestIDs) > 0 {
		query := fmt.Sprintf("%s.forest_id IN ?", dao.TableName())
		db.Where(query, cond.ForestIDs)
	}
	if cond.TableNameLike != "" {
		query := fmt.Sprintf("%s.table_name LIKE ?", dao.TableName())
		db.Where(query, fmt.Sprintf("%%%s%%", cond.TableNameLike))
	}
}
