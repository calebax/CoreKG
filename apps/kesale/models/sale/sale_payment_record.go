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

type SalePaymentRecordCond struct {
	BaseCond
	Filters  []apiobj.Filter
	ID       uint
	RecordSn string
	OrderSN  string
}

type SalePaymentRecordDao struct {
	BaseModel
}

func NewSalePaymentRecordDao() *SalePaymentRecordDao {
	return &SalePaymentRecordDao{}
}

func (dao *SalePaymentRecordDao) TableName() string {
	return dbtype.TableNameSalePaymentRecord
}

func (dao *SalePaymentRecordDao) WithTx(db *gorm.DB) *SalePaymentRecordDao {
	return &SalePaymentRecordDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *SalePaymentRecordDao) Insert(ctx context.Context, entity *dbtype.SalePaymentRecord) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[SalePaymentRecordDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *SalePaymentRecordDao) BatchInsert(ctx context.Context, entityList dbtype.SalePaymentRecordList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[SalePaymentRecordDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[SalePaymentRecordDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *SalePaymentRecordDao) UpdateByID(ctx context.Context, id uint, entity *dbtype.SalePaymentRecord) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[SalePaymentRecordDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *SalePaymentRecordDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[SalePaymentRecordDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *SalePaymentRecordDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[SalePaymentRecordDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *SalePaymentRecordDao) GetByID(ctx context.Context, id uint) (*dbtype.SalePaymentRecord, error) {
	var entity dbtype.SalePaymentRecord
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[SalePaymentRecordDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *SalePaymentRecordDao) GetByCond(ctx context.Context, cond *SalePaymentRecordCond) (*dbtype.SalePaymentRecord, error) {
	var entity dbtype.SalePaymentRecord
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[SalePaymentRecordDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *SalePaymentRecordDao) GetListByCond(ctx context.Context, cond *SalePaymentRecordCond) (dbtype.SalePaymentRecordList, error) {
	var entityList dbtype.SalePaymentRecordList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[SalePaymentRecordDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *SalePaymentRecordDao) GetPageListByCond(ctx context.Context, cond *SalePaymentRecordCond) (dbtype.SalePaymentRecordList, int64, error) {
	db := dao.DB(ctx).Model(&dbtype.SalePaymentRecord{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	for _, v := range cond.Filters {
		switch v.Field {
		case "order_sn":
			db = db.Where(fmt.Sprintf("%s.order_sn LIKE ?", dao.TableName()), "%"+v.Value[0]+"%")
		case "status":
			db = db.Where(fmt.Sprintf("%s.status = ?", dao.TableName()), v.Value[0])
		case "company_id":
			db = db.Where(fmt.Sprintf("%s.company_id = ?", dao.TableName()), v.Value[0])
		default:
			logs.ErrorContextf(ctx, "unknown filter field :%v", v.Field)
			return nil, 0, fmt.Errorf("unknown filter field :%v", v.Field)
		}
	}

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[SalePaymentRecordDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList dbtype.SalePaymentRecordList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[SalePaymentRecordDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *SalePaymentRecordDao) CountByCond(ctx context.Context, cond *SalePaymentRecordCond) (int64, error) {
	db := dao.DB(ctx).Model(&dbtype.SalePaymentRecord{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[SalePaymentRecordDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *SalePaymentRecordDao) BuildCondition(db *gorm.DB, cond *SalePaymentRecordCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.RecordSn != "" {
		query := fmt.Sprintf("%s.record_sn = ?", dao.TableName())
		db.Where(query, cond.RecordSn)
	}
	if cond.OrderSN != "" {
		query := fmt.Sprintf("%s.order_sn = ?", dao.TableName())
		db.Where(query, cond.OrderSN)
	}
}
