package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ChatRecentUsedModelCond struct {
	BaseCond
	Filters []apiobj.Filter
	ID      uint
	Uin     uint
}

type ChatRecentUsedModelDao struct {
	BaseModel
}

func NewChatRecentUsedModelDao() *ChatRecentUsedModelDao {
	return &ChatRecentUsedModelDao{}
}

func (dao *ChatRecentUsedModelDao) TableName() string {
	return chattype.TableNameChatRecentUsedModel
}

func (dao *ChatRecentUsedModelDao) WithTx(db *gorm.DB) *ChatRecentUsedModelDao {
	return &ChatRecentUsedModelDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ChatRecentUsedModelDao) Insert(ctx context.Context, entity *chattype.ChatRecentUsedModel) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[ChatRecentUsedModelDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ChatRecentUsedModelDao) BatchInsert(ctx context.Context, entityList chattype.ChatRecentUsedModelList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[ChatRecentUsedModelDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[ChatRecentUsedModelDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ChatRecentUsedModelDao) Upsert(ctx context.Context, entity *chattype.ChatRecentUsedModel) error {
	now := time.Now()
	// 如果 LastUsedAt 为空，设置为当前时间
	if entity.LastUsedAt == nil {
		entity.LastUsedAt = &now
	}
	// 如果 UsageCount 为 0，设置为 1（新记录）
	if entity.UsageCount == 0 {
		entity.UsageCount = 1
	}

	db := dao.DB(ctx).Table(dao.TableName())

	// 使用唯一索引 (uin, company_id, model_id) 作为冲突检测
	result := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "uin"},
			{Name: "company_id"},
			{Name: "model_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_used_at": now,
			"usage_count":  gorm.Expr("usage_count + 1"),
		}),
	}).Create(entity)

	if result.Error != nil {
		return fmt.Errorf("[ChatRecentUsedModelDao] Upsert fail, entity:%s, err: %v", logs.JSON(entity), result.Error)
	}

	// 由于使用了 gorm.Expr 自增，需要重新查询以获取最新的 usage_count 和其他字段
	var updatedEntity chattype.ChatRecentUsedModel
	if err := db.Where("uin = ? AND company_id = ? AND model_id = ?", entity.Uin, entity.CompanyID, entity.ModelID).
		First(&updatedEntity).Error; err != nil {
		return fmt.Errorf("[ChatRecentUsedModelDao] Query after upsert fail, err: %v", err)
	}
	// 更新 entity 的 ID 和其他字段
	entity.ID = updatedEntity.ID
	entity.UsageCount = updatedEntity.UsageCount
	entity.LastUsedAt = updatedEntity.LastUsedAt
	entity.CreatedAt = updatedEntity.CreatedAt
	entity.UpdatedAt = updatedEntity.UpdatedAt

	return nil
}

func (dao *ChatRecentUsedModelDao) UpdateByID(ctx context.Context, id uint, entity *chattype.ChatRecentUsedModel) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[ChatRecentUsedModelDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ChatRecentUsedModelDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[ChatRecentUsedModelDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ChatRecentUsedModelDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ChatRecentUsedModelDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ChatRecentUsedModelDao) GetByID(ctx context.Context, id uint) (*chattype.ChatRecentUsedModel, error) {
	var entity chattype.ChatRecentUsedModel
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ChatRecentUsedModelDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ChatRecentUsedModelDao) GetByCond(ctx context.Context, cond *ChatRecentUsedModelCond) (*chattype.ChatRecentUsedModel, error) {
	var entity chattype.ChatRecentUsedModel
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ChatRecentUsedModelDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ChatRecentUsedModelDao) GetListByCond(ctx context.Context, cond *ChatRecentUsedModelCond) (chattype.ChatRecentUsedModelList, error) {
	var entityList chattype.ChatRecentUsedModelList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[ChatRecentUsedModelDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ChatRecentUsedModelDao) GetPageListByCond(ctx context.Context, cond *ChatRecentUsedModelCond) (chattype.ChatRecentUsedModelList, int64, error) {
	db := dao.DB(ctx).Model(&chattype.ChatRecentUsedModel{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[ChatRecentUsedModelDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList chattype.ChatRecentUsedModelList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[ChatRecentUsedModelDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ChatRecentUsedModelDao) CountByCond(ctx context.Context, cond *ChatRecentUsedModelCond) (int64, error) {
	db := dao.DB(ctx).Model(&chattype.ChatRecentUsedModel{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[ChatRecentUsedModelDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ChatRecentUsedModelDao) BuildCondition(db *gorm.DB, cond *ChatRecentUsedModelCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.Uin > 0 {
		query := fmt.Sprintf("%s.uin = ?", dao.TableName())
		db.Where(query, cond.Uin)
	}
}
