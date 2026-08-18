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

type ForestDBCond struct {
	BaseCond
	Filters            []apiobj.Filter
	ID                 uint
	ForestID           uint
	ForestIDs          []uint
	ForestDBInstanceID uint
	ForestDBName       string
	ForestDBNameLike   string
	Enable             types.Bool
}

type ForestDBDao struct {
	BaseModel
}

func NewForestDBDao() *ForestDBDao {
	return &ForestDBDao{}
}

func (dao *ForestDBDao) TableName() string {
	return foresttype.TableNameKeForestDB
}

func (dao *ForestDBDao) WithTx(db *gorm.DB) *ForestDBDao {
	return &ForestDBDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ForestDBDao) Insert(ctx context.Context, entity *foresttype.ForestDB) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KeForestDbDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ForestDBDao) BatchInsert(ctx context.Context, entityList foresttype.ForestDBList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KeForestDbDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KeForestDbDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ForestDBDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.ForestDB) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KeForestDbDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ForestDBDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeForestDbDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ForestDBDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeForestDbDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ForestDBDao) DeleteByForestID(ctx context.Context, forestID uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("forest_id = ?", forestID).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeForestDbDao] DeleteByForestID fail, forestID:%d, err: %v", forestID, err)
	}
	return nil
}

func (dao *ForestDBDao) GetByID(ctx context.Context, id uint) (*foresttype.ForestDB, error) {
	var entity foresttype.ForestDB
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeForestDbDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ForestDBDao) GetByCond(ctx context.Context, cond *ForestDBCond) (*foresttype.ForestDB, error) {
	var entity foresttype.ForestDB
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeForestDbDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ForestDBDao) GetListByCond(ctx context.Context, cond *ForestDBCond) (foresttype.ForestDBList, error) {
	var entityList foresttype.ForestDBList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KeForestDbDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ForestDBDao) GetPageListByCond(ctx context.Context, cond *ForestDBCond) (foresttype.ForestDBList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.ForestDB{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeForestDbDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.ForestDBList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeForestDbDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ForestDBDao) CountByCond(ctx context.Context, cond *ForestDBCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.ForestDB{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KeForestDbDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ForestDBDao) BuildCondition(db *gorm.DB, cond *ForestDBCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}

	if cond.ForestID > 0 {
		query := fmt.Sprintf("%s.forest_id = ?", dao.TableName())
		db.Where(query, cond.ForestID)
	}
	if len(cond.ForestIDs) > 0 {
		query := fmt.Sprintf("%s.forest_id in (?)", dao.TableName())
		db.Where(query, cond.ForestIDs)
	}

	if cond.ForestDBInstanceID > 0 {
		query := fmt.Sprintf("%s.db_instance_id = ?", dao.TableName())
		db.Where(query, cond.ForestDBInstanceID)
	}

	if cond.ForestDBName != "" {
		query := fmt.Sprintf("%s.db_name = ?", dao.TableName())
		db.Where(query, cond.ForestDBName)
	}
	if cond.ForestDBNameLike != "" {
		query := fmt.Sprintf("%s.db_name LIKE ?", dao.TableName())
		db.Where(query, fmt.Sprintf("%%%s%%", cond.ForestDBNameLike))
	}
	if cond.Enable != 0 {
		query := fmt.Sprintf("%s.enable = ?", dao.TableName())
		db.Where(query, cond.Enable)
	}
}

func (dao *ForestDBDao) UpdateIDsMap(ctx context.Context, id []uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeForestDbDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}
