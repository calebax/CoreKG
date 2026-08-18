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

type KeMessageTemplateCond struct {
	BaseCond
	Filters  []apiobj.Filter
	ID       uint
	Name     foresttype.MessageTemplateName
	NameList []foresttype.MessageTemplateName
}

type KeMessageTemplateDao struct {
	BaseModel
}

func NewKeMessageTemplateDao() *KeMessageTemplateDao {
	return &KeMessageTemplateDao{}
}

func (dao *KeMessageTemplateDao) TableName() string {
	return foresttype.TableNameKeMessageTemplate
}

func (dao *KeMessageTemplateDao) WithTx(db *gorm.DB) *KeMessageTemplateDao {
	return &KeMessageTemplateDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *KeMessageTemplateDao) Insert(ctx context.Context, entity *foresttype.KeMessageTemplate) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KeMessageTemplateDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeMessageTemplateDao) BatchInsert(ctx context.Context, entityList foresttype.KeMessageTemplateList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KeMessageTemplateDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KeMessageTemplateDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *KeMessageTemplateDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.KeMessageTemplate) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KeMessageTemplateDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeMessageTemplateDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeMessageTemplateDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *KeMessageTemplateDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeMessageTemplateDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *KeMessageTemplateDao) GetByID(ctx context.Context, id uint) (*foresttype.KeMessageTemplate, error) {
	var entity foresttype.KeMessageTemplate
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeMessageTemplateDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *KeMessageTemplateDao) GetByCond(ctx context.Context, cond *KeMessageTemplateCond) (*foresttype.KeMessageTemplate, error) {
	var entity foresttype.KeMessageTemplate
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeMessageTemplateDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *KeMessageTemplateDao) GetListByCond(ctx context.Context, cond *KeMessageTemplateCond) (foresttype.KeMessageTemplateList, error) {
	var entityList foresttype.KeMessageTemplateList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KeMessageTemplateDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *KeMessageTemplateDao) GetPageListByCond(ctx context.Context, cond *KeMessageTemplateCond) (foresttype.KeMessageTemplateList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeMessageTemplate{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeMessageTemplateDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.KeMessageTemplateList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeMessageTemplateDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *KeMessageTemplateDao) CountByCond(ctx context.Context, cond *KeMessageTemplateCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeMessageTemplate{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KeMessageTemplateDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *KeMessageTemplateDao) BuildCondition(db *gorm.DB, cond *KeMessageTemplateCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.Name != "" {
		query := fmt.Sprintf("%s.name = ?", dao.TableName())
		db.Where(query, cond.Name)
	}
	if len(cond.NameList) > 0 {
		query := fmt.Sprintf("%s.name IN ?", dao.TableName())
		db.Where(query, cond.NameList)
	}
}
