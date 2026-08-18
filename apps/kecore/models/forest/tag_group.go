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

type TagGroupCond struct {
	BaseCond
	Filters  []apiobj.Filter
	ID       uint
	IDs      []uint
	Name     string
	NameLike string
}

type TagGroupDao struct {
	BaseModel
}

func NewTagGroupDao() *TagGroupDao {
	return &TagGroupDao{}
}

func (dao *TagGroupDao) TableName() string {
	return foresttype.TableNameTagGroup
}

func (dao *TagGroupDao) WithTx(db *gorm.DB) *TagGroupDao {
	return &TagGroupDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *TagGroupDao) Insert(ctx context.Context, entity *foresttype.TagGroup) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[TagGroupDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *TagGroupDao) BatchInsert(ctx context.Context, entityList foresttype.TagGroupList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[TagGroupDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[TagGroupDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *TagGroupDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.TagGroup) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[TagGroupDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *TagGroupDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[TagGroupDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *TagGroupDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[TagGroupDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *TagGroupDao) GetByID(ctx context.Context, id uint) (*foresttype.TagGroup, error) {
	var entity foresttype.TagGroup
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[TagGroupDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *TagGroupDao) GetByCond(ctx context.Context, cond *TagGroupCond) (*foresttype.TagGroup, error) {
	var entity foresttype.TagGroup
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[TagGroupDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *TagGroupDao) GetListByCond(ctx context.Context, cond *TagGroupCond) (foresttype.TagGroupList, error) {
	var entityList foresttype.TagGroupList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[TagGroupDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *TagGroupDao) GetPageListByCond(ctx context.Context, cond *TagGroupCond) (foresttype.TagGroupList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.TagGroup{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[TagGroupDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.TagGroupList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[TagGroupDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *TagGroupDao) CountByCond(ctx context.Context, cond *TagGroupCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.TagGroup{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[TagGroupDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *TagGroupDao) BuildCondition(db *gorm.DB, cond *TagGroupCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if len(cond.IDs) > 0 {
		query := fmt.Sprintf("%s.id IN (?)", dao.TableName())
		db.Where(query, cond.IDs)
	}
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.Name != "" {
		query := fmt.Sprintf("%s.name = ?", dao.TableName())
		db.Where(query, cond.Name)
	}
	if cond.NameLike != "" {
		query := fmt.Sprintf("%s.name LIKE ?", dao.TableName())
		db.Where(query, fmt.Sprintf("%%%s%%", cond.NameLike))
	}
}
