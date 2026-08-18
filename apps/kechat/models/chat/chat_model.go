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

type ChatModelCond struct {
	BaseCond
	Filters    []apiobj.Filter
	ID         uint
	PublicType chattype.PublecType
}

type ChatModelDao struct {
	BaseModel
}

func NewChatModelDao() *ChatModelDao {
	return &ChatModelDao{}
}

func (dao *ChatModelDao) TableName() string {
	return chattype.TableNameChatModel
}

func (dao *ChatModelDao) WithTx(db *gorm.DB) *ChatModelDao {
	return &ChatModelDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ChatModelDao) Insert(ctx context.Context, entity *chattype.ChatModel) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[ChatModelDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ChatModelDao) BatchInsert(ctx context.Context, entityList chattype.ChatModelList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[ChatModelDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[ChatModelDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ChatModelDao) UpdateByID(ctx context.Context, id uint, entity *chattype.ChatModel) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[ChatModelDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ChatModelDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[ChatModelDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ChatModelDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ChatModelDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ChatModelDao) GetByID(ctx context.Context, id uint) (*chattype.ChatModel, error) {
	var entity chattype.ChatModel
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ChatModelDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ChatModelDao) GetByCond(ctx context.Context, cond *ChatModelCond) (*chattype.ChatModel, error) {
	var entity chattype.ChatModel
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ChatModelDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ChatModelDao) GetListByCond(ctx context.Context, cond *ChatModelCond) (chattype.ChatModelList, error) {
	var entityList chattype.ChatModelList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[ChatModelDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ChatModelDao) GetPageListByCond(ctx context.Context, cond *ChatModelCond) (chattype.ChatModelList, int64, error) {
	db := dao.DB(ctx).Model(&chattype.ChatModel{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[ChatModelDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList chattype.ChatModelList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[ChatModelDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ChatModelDao) CountByCond(ctx context.Context, cond *ChatModelCond) (int64, error) {
	db := dao.DB(ctx).Model(&chattype.ChatModel{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[ChatModelDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ChatModelDao) BuildCondition(db *gorm.DB, cond *ChatModelCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.PublicType != "" {
		db.Where("public_type = ?", cond.PublicType)
	}
}
