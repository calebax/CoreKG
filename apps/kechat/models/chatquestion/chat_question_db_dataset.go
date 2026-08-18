package chatquestion

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kechat/models"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type ChatQuestionDbDatasetCond struct {
	models.BaseCond
	Filters []apiobj.Filter
	ID      uint
}

type ChatQuestionDbDatasetDao struct {
	models.BaseModel
}

func NewChatQuestionDbDatasetDao() *ChatQuestionDbDatasetDao {
	return &ChatQuestionDbDatasetDao{}
}

func (dao *ChatQuestionDbDatasetDao) TableName() string {
	return chattype.TableNameChatQuestionDbDataset
}

func (dao *ChatQuestionDbDatasetDao) WithTx(db *gorm.DB) *ChatQuestionDbDatasetDao {
	return &ChatQuestionDbDatasetDao{
		BaseModel: models.BaseModel{DBClient: db},
	}
}

func (dao *ChatQuestionDbDatasetDao) Insert(ctx context.Context, entity *chattype.ChatQuestionDbDataset) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[ChatQuestionDbDatasetDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ChatQuestionDbDatasetDao) BatchInsert(ctx context.Context, entityList chattype.ChatQuestionDbDatasetList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[ChatQuestionDbDatasetDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[ChatQuestionDbDatasetDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ChatQuestionDbDatasetDao) UpdateByID(ctx context.Context, id uint, entity *chattype.ChatQuestionDbDataset) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[ChatQuestionDbDatasetDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ChatQuestionDbDatasetDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[ChatQuestionDbDatasetDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ChatQuestionDbDatasetDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_time": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ChatQuestionDbDatasetDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ChatQuestionDbDatasetDao) GetByID(ctx context.Context, id uint) (*chattype.ChatQuestionDbDataset, error) {
	var entity chattype.ChatQuestionDbDataset
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ChatQuestionDbDatasetDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ChatQuestionDbDatasetDao) GetByCond(ctx context.Context, cond *ChatQuestionDbDatasetCond) (*chattype.ChatQuestionDbDataset, error) {
	var entity chattype.ChatQuestionDbDataset
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ChatQuestionDbDatasetDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ChatQuestionDbDatasetDao) GetListByCond(ctx context.Context, cond *ChatQuestionDbDatasetCond) (chattype.ChatQuestionDbDatasetList, error) {
	var entityList chattype.ChatQuestionDbDatasetList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[ChatQuestionDbDatasetDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ChatQuestionDbDatasetDao) GetPageListByCond(ctx context.Context, cond *ChatQuestionDbDatasetCond) (chattype.ChatQuestionDbDatasetList, int64, error) {
	db := dao.DB(ctx).Model(&chattype.ChatQuestionDbDataset{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[ChatQuestionDbDatasetDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList chattype.ChatQuestionDbDatasetList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[ChatQuestionDbDatasetDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ChatQuestionDbDatasetDao) CountByCond(ctx context.Context, cond *ChatQuestionDbDatasetCond) (int64, error) {
	db := dao.DB(ctx).Model(&chattype.ChatQuestionDbDataset{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[ChatQuestionDbDatasetDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ChatQuestionDbDatasetDao) BuildCondition(db *gorm.DB, cond *ChatQuestionDbDatasetCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
}
