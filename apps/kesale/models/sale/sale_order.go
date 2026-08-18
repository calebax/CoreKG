package sale

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kesale/models"
	dbtype "github.com/insmtx/corekg/apps/kesale/models/saletype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type SaleOrderCond struct {
	BaseCond
	Filters   []apiobj.Filter
	ID        uint
	CompanyID uint
	Uin       uint
	OrderSN   string
}

type SaleOrderDao struct {
	BaseModel
}

func NewSaleOrderDao() *SaleOrderDao {
	return &SaleOrderDao{}
}

func (dao *SaleOrderDao) TableName() string {
	return dbtype.TableNameSaleOrder
}

func (dao *SaleOrderDao) WithTx(db *gorm.DB) *SaleOrderDao {
	return &SaleOrderDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *SaleOrderDao) Insert(ctx context.Context, entity *dbtype.SaleOrder) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[SaleOrderDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *SaleOrderDao) BatchInsert(ctx context.Context, entityList dbtype.SaleOrderList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[SaleOrderDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[SaleOrderDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *SaleOrderDao) UpdateByID(ctx context.Context, id uint, entity *dbtype.SaleOrder) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[SaleOrderDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *SaleOrderDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[SaleOrderDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *SaleOrderDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[SaleOrderDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *SaleOrderDao) GetByID(ctx context.Context, id uint) (*dbtype.SaleOrder, error) {
	var entity dbtype.SaleOrder
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[SaleOrderDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *SaleOrderDao) GetByCond(ctx context.Context, cond *SaleOrderCond) (*dbtype.SaleOrder, error) {
	var entity dbtype.SaleOrder
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[SaleOrderDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *SaleOrderDao) GetListByCond(ctx context.Context, cond *SaleOrderCond) (dbtype.SaleOrderList, error) {
	var entityList dbtype.SaleOrderList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[SaleOrderDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *SaleOrderDao) GetPageListByCond(ctx context.Context, cond *SaleOrderCond) (dbtype.SaleOrderList, int64, error) {
	db := dao.DB(ctx).Model(&dbtype.SaleOrder{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	for _, v := range cond.Filters {
		switch v.Field {
		case "order_sn":
			db = db.Where(fmt.Sprintf("%s.order_sn LIKE ?", dao.TableName()), "%"+v.Value[0]+"%")
		case "status":
			db = db.Where(fmt.Sprintf("%s.order_status = ?", dao.TableName()), v.Value[0])
		case "company_id":
			db = db.Where(fmt.Sprintf("%s.company_id = ?", dao.TableName()), v.Value[0])
		default:
			logs.ErrorContextf(ctx, "unknown filter field :%v", v.Field)
			return nil, 0, fmt.Errorf("unknown filter field :%v", v.Field)
		}
	}

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[SaleOrderDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList dbtype.SaleOrderList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[SaleOrderDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *SaleOrderDao) CountByCond(ctx context.Context, cond *SaleOrderCond) (int64, error) {
	db := dao.DB(ctx).Model(&dbtype.SaleOrder{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[SaleOrderDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *SaleOrderDao) BuildCondition(db *gorm.DB, cond *SaleOrderCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.CompanyID > 0 {
		query := fmt.Sprintf("%s.company_id = ?", dao.TableName())
		db.Where(query, cond.CompanyID)
	}
	if cond.Uin > 0 {
		query := fmt.Sprintf("%s.uin = ?", dao.TableName())
		db.Where(query, cond.Uin)
	}
	if cond.OrderSN != "" {
		query := fmt.Sprintf("%s.order_sn = ?", dao.TableName())
		db.Where(query, cond.OrderSN)
	}
}

func (dao *SaleOrderDao) GetPendingOrders(ctx context.Context, startTime time.Time) (dbtype.SaleOrderList, error) {
	var entityList dbtype.SaleOrderList
	db := dao.DB(ctx).Table(dao.TableName())

	err := db.Where("created_at >= ?", startTime).
		Where("order_status NOT IN (?)", []string{
			models.OrderStatusSuccess.String(),
			models.OrderStatusClosed.String(),
		}).Find(&entityList).Error

	if err != nil {
		return nil, fmt.Errorf("[SaleOrderDao] GetPendingOrders fail, startTime:%v, err: %v", startTime, err)
	}
	return entityList, nil
}

// UpdateOrderStatusParams 订单状态更新参数
type UpdateOrderStatusParams struct {
	OrderID        uint
	TargetStatus   models.OrderStatus
	PaymentChannel *models.PaymentChannel // 可选：支付渠道（更新为成功时需要）
	PaidAt         *time.Time             // 可选：支付时间（更新为成功时需要）
}

// UpdateOrderStatus 更新订单状态
func (dao *SaleOrderDao) UpdateOrderStatus(ctx context.Context, params *UpdateOrderStatusParams) (int64, error) {
	db := dao.DB(ctx).Table(dao.TableName())

	updateMap := map[string]interface{}{
		"order_status": params.TargetStatus.String(),
		"updated_at":   time.Now(),
	}

	var query *gorm.DB

	switch params.TargetStatus {
	case models.OrderStatusSuccess:
		// 只能从 pending 或 paying 状态转换
		if params.PaymentChannel == nil || params.PaidAt == nil {
			return 0, fmt.Errorf("[SaleOrderDao] UpdateOrderStatus: payment channel and paid_at are required for success status")
		}
		updateMap["payment_channel"] = params.PaymentChannel.String()
		updateMap["paid_at"] = params.PaidAt

		query = db.Where("id = ? AND order_status IN (?)", params.OrderID,
			[]string{models.OrderStatusPending.String(), models.OrderStatusPaying.String()})

	case models.OrderStatusClosed:
		query = db.Where("id = ? AND order_status NOT IN (?)", params.OrderID, []string{
			models.OrderStatusSuccess.String(),
			models.OrderStatusClosed.String(),
		})

	default:
		return 0, fmt.Errorf("[SaleOrderDao] UpdateOrderStatus: unsupported target status: %s", params.TargetStatus.String())
	}

	result := query.Updates(updateMap)
	if result.Error != nil {
		return 0, fmt.Errorf("[SaleOrderDao] UpdateOrderStatus fail, params:%s, err: %v", logs.JSON(params), result.Error)
	}

	return result.RowsAffected, nil
}
