package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type ChatChartCond struct {
	BaseCond
	Filters []apiobj.Filter
	ID      uint
}

type ChatChartDao struct {
	BaseModel
}

func NewChatChartDao() *ChatChartDao {
	return &ChatChartDao{}
}

func (dao *ChatChartDao) TableName() string {
	return chattype.TableNameChatChart
}

func (dao *ChatChartDao) WithTx(db *gorm.DB) *ChatChartDao {
	return &ChatChartDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ChatChartDao) Insert(ctx context.Context, entity *chattype.ChatChart) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[ChatChartDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ChatChartDao) BatchInsert(ctx context.Context, entityList chattype.ChatChartList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[ChatChartDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[ChatChartDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ChatChartDao) UpdateByID(ctx context.Context, id uint, entity *chattype.ChatChart) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[ChatChartDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ChatChartDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[ChatChartDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ChatChartDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ChatChartDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ChatChartDao) DeleteByIDs(ctx context.Context, ids []uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id in (?)", ids).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ChatChartDao] DeleteByIDs fail, ids:%s, err: %v", logs.JSON(ids), err)
	}
	return nil
}

func (dao *ChatChartDao) GetByID(ctx context.Context, id uint) (*chattype.ChatChart, error) {
	var entity chattype.ChatChart
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ChatChartDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ChatChartDao) GetByCond(ctx context.Context, cond *ChatChartCond) (*chattype.ChatChart, error) {
	var entity chattype.ChatChart
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ChatChartDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ChatChartDao) GetListByCond(ctx context.Context, cond *ChatChartCond) (chattype.ChatChartList, error) {
	var entityList chattype.ChatChartList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[ChatChartDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ChatChartDao) GetPageListByCond(ctx context.Context, cond *ChatChartCond) (chattype.ChatChartList, int64, error) {
	db := dao.DB(ctx).Model(&chattype.ChatChart{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[ChatChartDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList chattype.ChatChartList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[ChatChartDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ChatChartDao) CountByCond(ctx context.Context, cond *ChatChartCond) (int64, error) {
	db := dao.DB(ctx).Model(&chattype.ChatChart{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[ChatChartDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ChatChartDao) BuildCondition(db *gorm.DB, cond *ChatChartCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
}
