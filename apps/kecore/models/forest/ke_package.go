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

type KePackageCond struct {
	BaseCond
	Filters    []apiobj.Filter
	ID         uint
	IDs        []uint
	SourceType foresttype.PackageSourceType
	Status     foresttype.PackageStatus
}

type KePackageDao struct {
	BaseModel
}

func NewKePackageDao() *KePackageDao {
	return &KePackageDao{}
}

func (dao *KePackageDao) TableName() string {
	return foresttype.TableNameKePackage
}

func (dao *KePackageDao) WithTx(db *gorm.DB) *KePackageDao {
	return &KePackageDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *KePackageDao) Insert(ctx context.Context, entity *foresttype.KePackage) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KePackageDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *KePackageDao) BatchInsert(ctx context.Context, entityList foresttype.KePackageList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KePackageDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KePackageDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *KePackageDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.KePackage) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KePackageDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *KePackageDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KePackageDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *KePackageDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KePackageDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *KePackageDao) GetByID(ctx context.Context, id uint) (*foresttype.KePackage, error) {
	var entity foresttype.KePackage
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KePackageDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *KePackageDao) GetByCond(ctx context.Context, cond *KePackageCond) (*foresttype.KePackage, error) {
	var entity foresttype.KePackage
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KePackageDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *KePackageDao) GetListByCond(ctx context.Context, cond *KePackageCond) (foresttype.KePackageList, error) {
	var entityList foresttype.KePackageList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KePackageDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *KePackageDao) GetPageListByCond(ctx context.Context, cond *KePackageCond) (foresttype.KePackageList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KePackage{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KePackageDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.KePackageList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KePackageDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *KePackageDao) CountByCond(ctx context.Context, cond *KePackageCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KePackage{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KePackageDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *KePackageDao) BuildCondition(db *gorm.DB, cond *KePackageCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if len(cond.IDs) > 0 {
		query := fmt.Sprintf("%s.id in ?", dao.TableName())
		db.Where(query, cond.IDs)
	}
	if cond.SourceType != "" {
		query := fmt.Sprintf("%s.source_type = ?", dao.TableName())
		db.Where(query, cond.SourceType)
	}
	if cond.Status != "" {
		query := fmt.Sprintf("%s.status = ?", dao.TableName())
		db.Where(query, cond.Status)
	}
}
