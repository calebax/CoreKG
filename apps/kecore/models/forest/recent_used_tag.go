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

type RecentUsedTagCond struct {
	BaseCond
	Filters []apiobj.Filter
	ID      uint
	TagIDs  []uint
}

type RecentUsedTagDao struct {
	BaseModel
}

func NewRecentUsedTagDao() *RecentUsedTagDao {
	return &RecentUsedTagDao{}
}

func (dao *RecentUsedTagDao) TableName() string {
	return foresttype.TableNameRecentUsedTag
}

func (dao *RecentUsedTagDao) WithTx(db *gorm.DB) *RecentUsedTagDao {
	return &RecentUsedTagDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *RecentUsedTagDao) Insert(ctx context.Context, entity *foresttype.RecentUsedTag) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[RecentUsedTagDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *RecentUsedTagDao) BatchInsert(ctx context.Context, entityList foresttype.RecentUsedTagList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[RecentUsedTagDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[RecentUsedTagDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *RecentUsedTagDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.RecentUsedTag) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[RecentUsedTagDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *RecentUsedTagDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[RecentUsedTagDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *RecentUsedTagDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[RecentUsedTagDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *RecentUsedTagDao) DeleteByTagIDs(ctx context.Context, tagIDs []uint) error {
	if len(tagIDs) == 0 {
		return nil
	}
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("tag_id IN ?", tagIDs).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[RecentUsedTagDao] DeleteByTagIDs fail, tagIDs:%s, err: %v", logs.JSON(tagIDs), err)
	}
	return nil
}

func (dao *RecentUsedTagDao) GetByID(ctx context.Context, id uint) (*foresttype.RecentUsedTag, error) {
	var entity foresttype.RecentUsedTag
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[RecentUsedTagDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *RecentUsedTagDao) GetByCond(ctx context.Context, cond *RecentUsedTagCond) (*foresttype.RecentUsedTag, error) {
	var entity foresttype.RecentUsedTag
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[RecentUsedTagDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *RecentUsedTagDao) GetListByCond(ctx context.Context, cond *RecentUsedTagCond) (foresttype.RecentUsedTagList, error) {
	var entityList foresttype.RecentUsedTagList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[RecentUsedTagDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *RecentUsedTagDao) GetPageListByCond(ctx context.Context, cond *RecentUsedTagCond) (foresttype.RecentUsedTagList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.RecentUsedTag{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[RecentUsedTagDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.RecentUsedTagList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[RecentUsedTagDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *RecentUsedTagDao) CountByCond(ctx context.Context, cond *RecentUsedTagCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.RecentUsedTag{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[RecentUsedTagDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *RecentUsedTagDao) BuildCondition(db *gorm.DB, cond *RecentUsedTagCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if len(cond.TagIDs) > 0 {
		query := fmt.Sprintf("%s.tag_id IN (?)", dao.TableName())
		db.Where(query, cond.TagIDs)
	}
}
