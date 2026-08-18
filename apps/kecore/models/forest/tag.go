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

type TagCond struct {
	BaseCond
	Filters  []apiobj.Filter
	ID       uint
	IDs      []uint
	GroupID  uint
	Name     string
	NameLike string
}

type TagDao struct {
	BaseModel
}

func NewTagDao() *TagDao {
	return &TagDao{}
}

func (dao *TagDao) TableName() string {
	return foresttype.TableNameTag
}

func (dao *TagDao) WithTx(db *gorm.DB) *TagDao {
	return &TagDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *TagDao) Insert(ctx context.Context, entity *foresttype.Tag) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[TagDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *TagDao) BatchInsert(ctx context.Context, entityList foresttype.TagList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[TagDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[TagDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *TagDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.Tag) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[TagDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *TagDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[TagDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *TagDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[TagDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *TagDao) DeleteByGroupID(ctx context.Context, groupID uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("group_id = ?", groupID).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[TagDao] DeleteByGroupID fail, groupID:%d, err: %v", groupID, err)
	}
	return nil
}

func (dao *TagDao) GetByID(ctx context.Context, id uint) (*foresttype.Tag, error) {
	var entity foresttype.Tag
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[TagDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *TagDao) GetByCond(ctx context.Context, cond *TagCond) (*foresttype.Tag, error) {
	var entity foresttype.Tag
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[TagDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *TagDao) GetListByCond(ctx context.Context, cond *TagCond) (foresttype.TagList, error) {
	var entityList foresttype.TagList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[TagDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *TagDao) GetPageListByCond(ctx context.Context, cond *TagCond) (foresttype.TagList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.Tag{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[TagDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.TagList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[TagDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *TagDao) CountByCond(ctx context.Context, cond *TagCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.Tag{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[TagDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *TagDao) BuildCondition(db *gorm.DB, cond *TagCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if len(cond.IDs) > 0 {
		query := fmt.Sprintf("%s.id IN (?)", dao.TableName())
		db.Where(query, cond.IDs)
	}
	if cond.GroupID > 0 {
		query := fmt.Sprintf("%s.group_id = ?", dao.TableName())
		db.Where(query, cond.GroupID)
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
