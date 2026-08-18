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

type ResourceTagCond struct {
	BaseCond
	Filters      []apiobj.Filter
	ID           uint
	TagIDs       []uint
	ResourceType foresttype.TagResourceType
	ResourceIDs  []uint
}

type ResourceTagDao struct {
	BaseModel
}

func NewResourceTagDao() *ResourceTagDao {
	return &ResourceTagDao{}
}

func (dao *ResourceTagDao) TableName() string {
	return foresttype.TableNameResourceTag
}

func (dao *ResourceTagDao) WithTx(db *gorm.DB) *ResourceTagDao {
	return &ResourceTagDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ResourceTagDao) Insert(ctx context.Context, entity *foresttype.ResourceTag) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[ResourceTagDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ResourceTagDao) BatchInsert(ctx context.Context, entityList foresttype.ResourceTagList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[ResourceTagDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[ResourceTagDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ResourceTagDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.ResourceTag) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[ResourceTagDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ResourceTagDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[ResourceTagDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ResourceTagDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ResourceTagDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ResourceTagDao) DeleteByTagIDs(ctx context.Context, tagIDs []uint) error {
	if len(tagIDs) == 0 {
		return nil
	}
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("tag_id IN ?", tagIDs).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ResourceTagDao] DeleteByTagIDs fail, tagIDs:%s, err: %v", logs.JSON(tagIDs), err)
	}
	return nil
}

func (dao *ResourceTagDao) DeleteByResource(ctx context.Context, resourceType foresttype.TagResourceType, resourceIDs []uint) error {
	if len(resourceIDs) == 0 {
		return nil
	}
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("resource_type = ? AND resource_id IN ?", resourceType, resourceIDs).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ResourceTagDao] DeleteByResource fail, resourceType:%s, resourceIDs:%s, err: %v", resourceType, logs.JSON(resourceIDs), err)
	}
	return nil
}

func (dao *ResourceTagDao) GetByID(ctx context.Context, id uint) (*foresttype.ResourceTag, error) {
	var entity foresttype.ResourceTag
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ResourceTagDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ResourceTagDao) GetByCond(ctx context.Context, cond *ResourceTagCond) (*foresttype.ResourceTag, error) {
	var entity foresttype.ResourceTag
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ResourceTagDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ResourceTagDao) GetListByCond(ctx context.Context, cond *ResourceTagCond) (foresttype.ResourceTagList, error) {
	var entityList foresttype.ResourceTagList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[ResourceTagDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ResourceTagDao) GetPageListByCond(ctx context.Context, cond *ResourceTagCond) (foresttype.ResourceTagList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.ResourceTag{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[ResourceTagDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.ResourceTagList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[ResourceTagDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ResourceTagDao) CountByCond(ctx context.Context, cond *ResourceTagCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.ResourceTag{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[ResourceTagDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ResourceTagDao) BuildCondition(db *gorm.DB, cond *ResourceTagCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.ResourceType != "" {
		query := fmt.Sprintf("%s.resource_type = ?", dao.TableName())
		db.Where(query, cond.ResourceType)
	}
	if len(cond.ResourceIDs) > 0 {
		query := fmt.Sprintf("%s.resource_id IN ?", dao.TableName())
		db.Where(query, cond.ResourceIDs)
	}
	if len(cond.TagIDs) > 0 {
		query := fmt.Sprintf("%s.tag_id IN ?", dao.TableName())
		db.Where(query, cond.TagIDs)
	}
}
