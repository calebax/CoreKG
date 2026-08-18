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

type UinCollectionsCond struct {
	BaseCond
	Filters        []apiobj.Filter
	ID             uint
	ResourceIDs    []uint
	ResourceType   foresttype.ResourceType
	NotResourceIDs []uint
}

type UinCollectionsDao struct {
	BaseModel
}

func NewUinCollectionsDao() *UinCollectionsDao {
	return &UinCollectionsDao{}
}

func (dao *UinCollectionsDao) TableName() string {
	return foresttype.TableNameUinCollections
}

func (dao *UinCollectionsDao) WithTx(db *gorm.DB) *UinCollectionsDao {
	return &UinCollectionsDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *UinCollectionsDao) Insert(ctx context.Context, entity *foresttype.UinCollections) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[UinCollectionsDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *UinCollectionsDao) BatchInsert(ctx context.Context, entityList foresttype.UinCollectionsList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[UinCollectionsDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[UinCollectionsDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *UinCollectionsDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.UinCollections) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[UinCollectionsDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *UinCollectionsDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[UinCollectionsDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *UinCollectionsDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[UinCollectionsDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *UinCollectionsDao) GetByID(ctx context.Context, id uint) (*foresttype.UinCollections, error) {
	var entity foresttype.UinCollections
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[UinCollectionsDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *UinCollectionsDao) GetByCond(ctx context.Context, cond *UinCollectionsCond) (*foresttype.UinCollections, error) {
	var entity foresttype.UinCollections
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[UinCollectionsDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *UinCollectionsDao) GetListByCond(ctx context.Context, cond *UinCollectionsCond) (foresttype.UinCollectionsList, error) {
	var entityList foresttype.UinCollectionsList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[UinCollectionsDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *UinCollectionsDao) GetPageListByCond(ctx context.Context, cond *UinCollectionsCond) (foresttype.UinCollectionsList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.UinCollections{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[UinCollectionsDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.UinCollectionsList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[UinCollectionsDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *UinCollectionsDao) CountByCond(ctx context.Context, cond *UinCollectionsCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.UinCollections{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[UinCollectionsDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *UinCollectionsDao) BuildCondition(db *gorm.DB, cond *UinCollectionsCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if len(cond.ResourceIDs) > 0 {
		query := fmt.Sprintf("%s.resource_id IN (?)", dao.TableName())
		db.Where(query, cond.ResourceIDs)
	}
	if cond.ResourceType != "" {
		query := fmt.Sprintf("%s.resource_type = ?", dao.TableName())
		db.Where(query, cond.ResourceType)
	}
	if len(cond.NotResourceIDs) > 0 {
		query := fmt.Sprintf("%s.resource_id NOT IN (?)", dao.TableName())
		db.Where(query, cond.NotResourceIDs)
	}
}
