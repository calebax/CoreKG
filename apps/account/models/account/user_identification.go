package account

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type UserIdentificationCond struct {
	BaseCond
	Filters     []apiobj.Filter
	ID          uint
	IDs         []uint
	SubjectType accounttype.SubjectType
	SubjectID   uint
	SubjectIDs  []uint
}

type UserIdentificationDao struct {
	BaseModel
}

func NewUserIdentificationDao() *UserIdentificationDao {
	return &UserIdentificationDao{}
}

func (dao *UserIdentificationDao) TableName() string {
	return accounttype.TableNameUserIdentification
}

func (dao *UserIdentificationDao) WithTx(db *gorm.DB) *UserIdentificationDao {
	return &UserIdentificationDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *UserIdentificationDao) Insert(ctx context.Context, entity *accounttype.UserIdentification) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[UserIdentificationDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *UserIdentificationDao) BatchInsert(ctx context.Context, entityList accounttype.UserIdentificationList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[UserIdentificationDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[UserIdentificationDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *UserIdentificationDao) UpdateByID(ctx context.Context, id uint, entity *accounttype.UserIdentification) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[UserIdentificationDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *UserIdentificationDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[UserIdentificationDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *UserIdentificationDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[UserIdentificationDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *UserIdentificationDao) GetByID(ctx context.Context, id uint) (*accounttype.UserIdentification, error) {
	var entity accounttype.UserIdentification
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[UserIdentificationDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *UserIdentificationDao) GetByCond(ctx context.Context, cond *UserIdentificationCond) (*accounttype.UserIdentification, error) {
	var entity accounttype.UserIdentification
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[UserIdentificationDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *UserIdentificationDao) GetListByCond(ctx context.Context, cond *UserIdentificationCond) (accounttype.UserIdentificationList, error) {
	var entityList accounttype.UserIdentificationList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[UserIdentificationDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *UserIdentificationDao) GetPageListByCond(ctx context.Context, cond *UserIdentificationCond) (accounttype.UserIdentificationList, int64, error) {
	db := dao.DB(ctx).Model(&accounttype.UserIdentification{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[UserIdentificationDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList accounttype.UserIdentificationList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[UserIdentificationDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *UserIdentificationDao) CountByCond(ctx context.Context, cond *UserIdentificationCond) (int64, error) {
	db := dao.DB(ctx).Model(&accounttype.UserIdentification{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[UserIdentificationDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *UserIdentificationDao) BuildCondition(db *gorm.DB, cond *UserIdentificationCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.SubjectType != "" {
		query := fmt.Sprintf("%s.subject_type = ?", dao.TableName())
		db.Where(query, cond.SubjectType)
	}
	if cond.SubjectID > 0 {
		query := fmt.Sprintf("%s.subject_id = ?", dao.TableName())
		db.Where(query, cond.SubjectID)
	}
	if len(cond.SubjectIDs) > 0 {
		query := fmt.Sprintf("%s.subject_id IN (?)", dao.TableName())
		db.Where(query, cond.SubjectIDs)
	}
	if len(cond.IDs) > 0 {
		query := fmt.Sprintf("%s.id IN (?)", dao.TableName())
		db.Where(query, cond.IDs)
	}
}
