package forest

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type KeCompanyDBCond struct {
	BaseCond
	Filters []apiobj.Filter
	ID      uint
}

type KeCompanyDBDao struct {
	BaseModel
}

func NewKeCompanyDBDao() *KeCompanyDBDao {
	return &KeCompanyDBDao{}
}

func (dao *KeCompanyDBDao) TableName() string {
	return foresttype.TableNameKeCompanyDb
}

func (dao *KeCompanyDBDao) WithTx(db *gorm.DB) *KeCompanyDBDao {
	return &KeCompanyDBDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *KeCompanyDBDao) Insert(ctx context.Context, entity *foresttype.KeCompanyDB) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KeCompanyDbDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeCompanyDBDao) BatchInsert(ctx context.Context, entityList foresttype.KeCompanyDBList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KeCompanyDbDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KeCompanyDbDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *KeCompanyDBDao) Upsert(ctx context.Context, entity *foresttype.KeCompanyDB) error {
	db := dao.DB(ctx).Table(dao.TableName())

	// 执行 upsert
	result := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "company_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
	}).Create(entity)

	if result.Error != nil {
		return fmt.Errorf("[KeCompanyDbDao] Upsert fail, entity:%s, err: %v", logs.JSON(entity), result.Error)
	}

	// 如果 RowsAffected = 0 或者主键为空，说明是更新操作，需要重新查询
	if result.RowsAffected == 0 || entity.ID == 0 {
		conditions := map[string]interface{}{
			"company_id": entity.CompanyID,
		}

		if err := db.Where(conditions).First(entity).Error; err != nil {
			return fmt.Errorf("[KeCompanyDbDao] Query after upsert fail, conditions:%v, err: %v", conditions, err)
		}
	}

	return nil
}

func (dao *KeCompanyDBDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.KeCompanyDB) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KeCompanyDbDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeCompanyDBDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeCompanyDbDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *KeCompanyDBDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_time": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeCompanyDbDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *KeCompanyDBDao) GetByID(ctx context.Context, id uint) (*foresttype.KeCompanyDB, error) {
	var entity foresttype.KeCompanyDB
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeCompanyDbDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *KeCompanyDBDao) GetByCond(ctx context.Context, cond *KeCompanyDBCond) (*foresttype.KeCompanyDB, error) {
	var entity foresttype.KeCompanyDB
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeCompanyDbDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *KeCompanyDBDao) GetListByCond(ctx context.Context, cond *KeCompanyDBCond) (foresttype.KeCompanyDBList, error) {
	var entityList foresttype.KeCompanyDBList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KeCompanyDbDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *KeCompanyDBDao) GetPageListByCond(ctx context.Context, cond *KeCompanyDBCond) (foresttype.KeCompanyDBList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeCompanyDB{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeCompanyDbDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.KeCompanyDBList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeCompanyDbDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *KeCompanyDBDao) CountByCond(ctx context.Context, cond *KeCompanyDBCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeCompanyDB{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KeCompanyDbDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *KeCompanyDBDao) BuildCondition(db *gorm.DB, cond *KeCompanyDBCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
}
