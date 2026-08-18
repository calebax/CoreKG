package sale

import (
	"context"
	"fmt"
	"time"

	dbtype "github.com/insmtx/corekg/apps/kesale/models/saletype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type SaleOrderItemCond struct {
	BaseCond
	Filters []apiobj.Filter
	ID      uint
	OrderID uint
}

type SaleOrderItemDao struct {
	BaseModel
}

func NewSaleOrderItemDao() *SaleOrderItemDao {
	return &SaleOrderItemDao{}
}

func (dao *SaleOrderItemDao) TableName() string {
	return dbtype.TableNameSaleOrderItem
}

func (dao *SaleOrderItemDao) WithTx(db *gorm.DB) *SaleOrderItemDao {
	return &SaleOrderItemDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *SaleOrderItemDao) Insert(ctx context.Context, entity *dbtype.SaleOrderItem) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[SaleOrderItemDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *SaleOrderItemDao) BatchInsert(ctx context.Context, entityList dbtype.SaleOrderItemList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[SaleOrderItemDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[SaleOrderItemDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *SaleOrderItemDao) UpdateByID(ctx context.Context, id uint, entity *dbtype.SaleOrderItem) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[SaleOrderItemDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *SaleOrderItemDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[SaleOrderItemDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *SaleOrderItemDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[SaleOrderItemDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *SaleOrderItemDao) GetByID(ctx context.Context, id uint) (*dbtype.SaleOrderItem, error) {
	var entity dbtype.SaleOrderItem
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[SaleOrderItemDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *SaleOrderItemDao) GetByCond(ctx context.Context, cond *SaleOrderItemCond) (*dbtype.SaleOrderItem, error) {
	var entity dbtype.SaleOrderItem
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[SaleOrderItemDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *SaleOrderItemDao) GetListByCond(ctx context.Context, cond *SaleOrderItemCond) (dbtype.SaleOrderItemList, error) {
	var entityList dbtype.SaleOrderItemList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[SaleOrderItemDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *SaleOrderItemDao) GetPageListByCond(ctx context.Context, cond *SaleOrderItemCond) (dbtype.SaleOrderItemList, int64, error) {
	db := dao.DB(ctx).Model(&dbtype.SaleOrderItem{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[SaleOrderItemDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList dbtype.SaleOrderItemList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[SaleOrderItemDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *SaleOrderItemDao) CountByCond(ctx context.Context, cond *SaleOrderItemCond) (int64, error) {
	db := dao.DB(ctx).Model(&dbtype.SaleOrderItem{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[SaleOrderItemDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *SaleOrderItemDao) BuildCondition(db *gorm.DB, cond *SaleOrderItemCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.OrderID > 0 {
		query := fmt.Sprintf("%s.order_id = ?", dao.TableName())
		db.Where(query, cond.OrderID)
	}
}
