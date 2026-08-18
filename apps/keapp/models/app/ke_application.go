package app

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/keapp/models/apptype"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type ApplicationCond struct {
	BaseCond
	NameLike   string
	TypeList   []apptype.AppTemplateType
	StatusList []apptype.AppStatus
}

type ApplicationDao struct {
	BaseModel
}

func NewApplicationDao() *ApplicationDao {
	return &ApplicationDao{}
}

func (dao *ApplicationDao) TableName() string {
	return apptype.TableNameKeApplication
}

func (dao *ApplicationDao) WithTx(db *gorm.DB) *ApplicationDao {
	return &ApplicationDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ApplicationDao) Insert(ctx context.Context, entity *apptype.KeApplication) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[ApplicationDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ApplicationDao) GetByID(ctx context.Context, id uint) (*apptype.KeApplication, error) {
	var entity apptype.KeApplication
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ApplicationDao] GetByID fail, id:%d, err: %v", id, err)
	}
	if entity.ID == 0 {
		return nil, nil
	}
	return &entity, nil
}

func (dao *ApplicationDao) UpdateByID(ctx context.Context, id uint, entity *apptype.KeApplication) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[ApplicationDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ApplicationDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[ApplicationDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ApplicationDao) SoftDelete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ApplicationDao] SoftDelete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ApplicationDao) CheckNameExists(ctx context.Context, excludeID uint, name string, companyID uint) (bool, error) {
	var count int64
	db := dao.DB(ctx).Table(dao.TableName()).Where("name = ? AND company_id = ?", name, companyID)
	if excludeID > 0 {
		db = db.Where("id != ?", excludeID)
	}
	if err := db.Count(&count).Error; err != nil {
		return false, fmt.Errorf("[ApplicationDao] CheckNameExists fail, name:%s, companyID:%d, err: %v", name, companyID, err)
	}
	return count > 0, nil
}

func (dao *ApplicationDao) GetPageListByCond(ctx context.Context, cond *ApplicationCond) (apptype.KeApplicationList, int64, error) {
	db := dao.DB(ctx).Model(&apptype.KeApplication{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[ApplicationDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList apptype.KeApplicationList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[ApplicationDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ApplicationDao) BuildCondition(db *gorm.DB, cond *ApplicationCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.NameLike != "" {
		query := fmt.Sprintf("%s.name LIKE ?", dao.TableName())
		db.Where(query, fmt.Sprintf("%%%s%%", cond.NameLike))
	}
	if len(cond.TypeList) > 0 {
		query := fmt.Sprintf("%s.type in (?)", dao.TableName())
		db.Where(query, cond.TypeList)
	}
	if len(cond.StatusList) > 0 {
		query := fmt.Sprintf("%s.status in (?)", dao.TableName())
		db.Where(query, cond.StatusList)
	}
}
