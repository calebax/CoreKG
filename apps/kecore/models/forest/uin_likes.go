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

type UinLikesCond struct {
	BaseCond
	Filters       []apiobj.Filter
	ID            uint
	ResourceIDs   []uint
	ResourceType  foresttype.ResourceType
	NoResourceIDs []uint
}

type UinLikesDao struct {
	BaseModel
}

func NewUinLikesDao() *UinLikesDao {
	return &UinLikesDao{}
}

func (dao *UinLikesDao) TableName() string {
	return foresttype.TableNameUinLikes
}

func (dao *UinLikesDao) WithTx(db *gorm.DB) *UinLikesDao {
	return &UinLikesDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *UinLikesDao) Insert(ctx context.Context, entity *foresttype.UinLikes) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[UinLikesDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *UinLikesDao) BatchInsert(ctx context.Context, entityList foresttype.UinLikesList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[UinLikesDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[UinLikesDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *UinLikesDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.UinLikes) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[UinLikesDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *UinLikesDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[UinLikesDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *UinLikesDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[UinLikesDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *UinLikesDao) GetByID(ctx context.Context, id uint) (*foresttype.UinLikes, error) {
	var entity foresttype.UinLikes
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[UinLikesDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *UinLikesDao) GetByCond(ctx context.Context, cond *UinLikesCond) (*foresttype.UinLikes, error) {
	var entity foresttype.UinLikes
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[UinLikesDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *UinLikesDao) GetListByCond(ctx context.Context, cond *UinLikesCond) (foresttype.UinLikesList, error) {
	var entityList foresttype.UinLikesList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[UinLikesDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *UinLikesDao) GetPageListByCond(ctx context.Context, cond *UinLikesCond) (foresttype.UinLikesList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.UinLikes{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[UinLikesDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.UinLikesList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[UinLikesDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *UinLikesDao) CountByCond(ctx context.Context, cond *UinLikesCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.UinLikes{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[UinLikesDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *UinLikesDao) BuildCondition(db *gorm.DB, cond *UinLikesCond) {
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
	if len(cond.NoResourceIDs) > 0 {
		query := fmt.Sprintf("%s.resource_id NOT IN (?)", dao.TableName())
		db.Where(query, cond.NoResourceIDs)
	}
}
