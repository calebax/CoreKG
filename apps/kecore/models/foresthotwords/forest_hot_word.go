package foresthotwords

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type ForestHotWordCond struct {
	BaseCond
	Filters []apiobj.Filter
	ID      uint
}

type ForestHotWordDao struct {
	BaseModel
}

func NewForestHotWordDao() *ForestHotWordDao {
	return &ForestHotWordDao{}
}

func (dao *ForestHotWordDao) TableName() string {
	return foresttype.TableNameForestHotWord
}

func (dao *ForestHotWordDao) WithTx(db *gorm.DB) *ForestHotWordDao {
	return &ForestHotWordDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ForestHotWordDao) Insert(ctx context.Context, entity *foresttype.ForestHotWord) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[ForestHotWordDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ForestHotWordDao) BatchInsert(ctx context.Context, entityList foresttype.ForestHotWordList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[ForestHotWordDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.CreateInBatches(entityList, 500).Error; err != nil {
		return fmt.Errorf("[ForestHotWordDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ForestHotWordDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.ForestHotWord) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[ForestHotWordDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ForestHotWordDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[ForestHotWordDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ForestHotWordDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ForestHotWordDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ForestHotWordDao) GetByID(ctx context.Context, id uint) (*foresttype.ForestHotWord, error) {
	var entity foresttype.ForestHotWord
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ForestHotWordDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ForestHotWordDao) GetByCond(ctx context.Context, cond *ForestHotWordCond) (*foresttype.ForestHotWord, error) {
	var entity foresttype.ForestHotWord
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ForestHotWordDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ForestHotWordDao) GetListByCond(ctx context.Context, cond *ForestHotWordCond) (foresttype.ForestHotWordList, error) {
	var entityList foresttype.ForestHotWordList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[ForestHotWordDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ForestHotWordDao) GetPageListByCond(ctx context.Context, cond *ForestHotWordCond) (foresttype.ForestHotWordList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.ForestHotWord{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[ForestHotWordDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.ForestHotWordList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[ForestHotWordDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ForestHotWordDao) CountByCond(ctx context.Context, cond *ForestHotWordCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.ForestHotWord{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[ForestHotWordDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ForestHotWordDao) BuildCondition(db *gorm.DB, cond *ForestHotWordCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
}
