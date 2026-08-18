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

type ChatAgentCond struct {
	BaseCond
	Filters   []apiobj.Filter
	ID        uint
	CompanyID uint
}

type ChatAgentDao struct {
	BaseModel
}

func NewChatAgentDao() *ChatAgentDao {
	return &ChatAgentDao{}
}

func (dao *ChatAgentDao) TableName() string {
	return chattype.TableNameAgent
}

func (dao *ChatAgentDao) WithTx(db *gorm.DB) *ChatAgentDao {
	return &ChatAgentDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ChatAgentDao) Insert(ctx context.Context, entity *chattype.ChatAgent) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[ChatAgentDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ChatAgentDao) BatchInsert(ctx context.Context, entityList chattype.ChatAgentList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[ChatAgentDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[ChatAgentDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ChatAgentDao) UpdateByID(ctx context.Context, id uint, entity *chattype.ChatAgent) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[ChatAgentDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ChatAgentDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[ChatAgentDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ChatAgentDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ChatAgentDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ChatAgentDao) GetByID(ctx context.Context, id uint) (*chattype.ChatAgent, error) {
	var entity chattype.ChatAgent
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ChatAgentDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ChatAgentDao) GetByCond(ctx context.Context, cond *ChatAgentCond) (*chattype.ChatAgent, error) {
	var entity chattype.ChatAgent
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ChatAgentDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ChatAgentDao) GetListByCond(ctx context.Context, cond *ChatAgentCond) (chattype.ChatAgentList, error) {
	var entityList chattype.ChatAgentList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[ChatAgentDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ChatAgentDao) GetPageListByCond(ctx context.Context, cond *ChatAgentCond) (chattype.ChatAgentList, int64, error) {
	db := dao.DB(ctx).Model(&chattype.ChatAgent{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[ChatAgentDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList chattype.ChatAgentList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[ChatAgentDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ChatAgentDao) CountByCond(ctx context.Context, cond *ChatAgentCond) (int64, error) {
	db := dao.DB(ctx).Model(&chattype.ChatAgent{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[ChatAgentDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ChatAgentDao) BuildCondition(db *gorm.DB, cond *ChatAgentCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.CompanyID > 0 {
		query := fmt.Sprintf("%s.company_id = ?", dao.TableName())
		db.Where(query, cond.CompanyID)
	}
}
