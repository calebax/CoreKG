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

type ForestCond struct {
	BaseCond
	Filters        []apiobj.Filter
	ID             uint
	IDs            []uint
	ForestTypeList []foresttype.ForestType
	NameLike       string
	GraphStatus    foresttype.GraphStatus
}

type ForestDao struct {
	BaseModel
}

func NewForestDao() *ForestDao {
	return &ForestDao{}
}

func (dao *ForestDao) TableName() string {
	return foresttype.TableNameKnownowForest
}

func (dao *ForestDao) WithTx(db *gorm.DB) *ForestDao {
	return &ForestDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ForestDao) Insert(ctx context.Context, entity *foresttype.KnownowForest) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KeForestDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ForestDao) BatchInsert(ctx context.Context, entityList foresttype.KnownowForestList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KeForestDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KeForestDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ForestDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.KnownowForest) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KeForestDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ForestDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeForestDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ForestDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_time": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeForestDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ForestDao) GetByID(ctx context.Context, id uint) (*foresttype.KnownowForest, error) {
	var entity foresttype.KnownowForest
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeForestDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ForestDao) GetByCond(ctx context.Context, cond *ForestCond) (*foresttype.KnownowForest, error) {
	var entity foresttype.KnownowForest
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeForestDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ForestDao) GetListByCond(ctx context.Context, cond *ForestCond) (foresttype.KnownowForestList, error) {
	var entityList foresttype.KnownowForestList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KeForestDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ForestDao) GetPageListByCond(ctx context.Context, cond *ForestCond) (foresttype.KnownowForestList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KnownowForest{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeForestDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.KnownowForestList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeForestDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ForestDao) CountByCond(ctx context.Context, cond *ForestCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KnownowForest{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KeForestDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ForestDao) BuildCondition(db *gorm.DB, cond *ForestCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if len(cond.IDs) > 0 {
		query := fmt.Sprintf("%s.id in (?)", dao.TableName())
		db.Where(query, cond.IDs)
	}
	if cond.NameLike != "" {
		query := fmt.Sprintf("%s.name LIKE ?", dao.TableName())
		db.Where(query, fmt.Sprintf("%%%s%%", cond.NameLike))
	}
	if len(cond.ForestTypeList) > 0 {
		query := fmt.Sprintf("%s.forest_type in (?)", dao.TableName())
		db.Where(query, cond.ForestTypeList)
	}
	if cond.GraphStatus != "" {
		query := fmt.Sprintf("%s.graph_status = ?", dao.TableName())
		db.Where(query, cond.GraphStatus)
	}
}
