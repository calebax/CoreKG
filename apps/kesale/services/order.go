package services

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kesale/models"
	"github.com/insmtx/corekg/apps/kesale/models/sale"
	"github.com/insmtx/corekg/apps/kesale/models/saletype"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

func CreateOrder(ctx context.Context, orderSN string, orderParams *models.OrderParams) (*saletype.SaleOrder, error) {
	order := &saletype.SaleOrder{
		CompanyID:      orderParams.CompanyID,
		Uin:            orderParams.Uin,
		OrderSn:        orderSN,
		TotalAmount:    orderParams.TotalAmount,
		PaymentAmount:  orderParams.PaymentAmount,
		PaymentChannel: orderParams.PaymentChannel.String(),
		OrderStatus:    models.OrderStatusPending.String(),
		BusinessSource: orderParams.BusinessSource,
		BusinessType:   orderParams.BusinessType.String(),
	}

	db := ManagerDBFromCtx(ctx)
	err := db.Transaction(func(db *gorm.DB) error {
		// Save order
		if err := sale.NewSaleOrderDao().WithTx(db).Insert(ctx, order); err != nil {
			return err
		}
		// Save order items
		orderItems := make([]saletype.SaleOrderItem, 0, len(orderParams.Products))
		for _, item := range orderParams.Products {
			orderItem := &saletype.SaleOrderItem{
				OrderID:        order.ID,
				BusinessSource: order.BusinessSource,
				ProductID:      item.ProductID,
				ProductName:    item.ProductName,
				Price:          item.Price,
				Quantity:       item.Quantity,
				TotalAmount:    item.Price * int64(item.Quantity),
				PaymentAmount:  item.Price * int64(item.Quantity),
			}
			orderItems = append(orderItems, *orderItem)
		}
		return sale.NewSaleOrderItemDao().WithTx(db).BatchInsert(ctx, orderItems)
	})

	if err != nil {
		return nil, err
	}
	return order, nil
}

func FindOrder(ctx context.Context, params *models.QueryOrderParams) (*saletype.SaleOrder, error) {
	db := ManagerDBFromCtx(ctx)
	return sale.NewSaleOrderDao().WithTx(db).GetByCond(ctx, &sale.SaleOrderCond{
		OrderSN:   params.OrderSN,
		CompanyID: params.CompanyID,
		Uin:       params.Uin,
	})
}

func ListOrderItems(ctx context.Context, orderID uint) ([]*models.OrderProduct, error) {
	db := ManagerDBFromCtx(ctx)
	items, err := sale.NewSaleOrderItemDao().WithTx(db).GetListByCond(ctx, &sale.SaleOrderItemCond{
		OrderID: orderID,
	})
	if err != nil {
		return nil, err
	}
	products := make([]*models.OrderProduct, 0, len(items))
	for _, item := range items {
		products = append(products, &models.OrderProduct{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}
	return products, nil
}

func UpdateOrderPayment(ctx context.Context, orderSN string, paymentChannel models.PaymentChannel, amount int64, paidAt time.Time) error {
	logs.InfoContextf(ctx, "UpdateOrderPayment: orderSN=%s, paymentChannel=%s, paidAt=%d", orderSN, paymentChannel.String(), paidAt)

	order, err := FindOrder(ctx, &models.QueryOrderParams{
		OrderSN: orderSN,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateOrderPayment: GetOrder err=%v", err)
		return err
	}
	if order == nil {
		logs.ErrorContextf(ctx, "UpdateOrderPayment: order not found, orderSN=%s", orderSN)
		return fmt.Errorf("order not found")
	}

	if order.OrderStatus == models.OrderStatusSuccess.String() {
		logs.InfoContextf(ctx, "UpdateOrderPayment: order already success, orderSN=%s", orderSN)
		return nil
	}

	if order.OrderStatus != models.OrderStatusPending.String() && order.OrderStatus != models.OrderStatusPaying.String() {
		logs.ErrorContextf(ctx, "UpdateOrderPayment: order status not matched, order=%v, orderStatus=%s", orderSN, order.OrderStatus)
		return fmt.Errorf("order status not matched: %s", order.OrderStatus)
	}

	if order.PaymentAmount != amount {
		logs.WarnContextf(ctx, "UpdateOrderPayment: payment amount not match, order=%v, expect=%d, actual=%d", orderSN, amount, order.PaymentAmount)
	}

	db := ManagerDBFromCtx(ctx)
	rowsAffected, err := sale.NewSaleOrderDao().WithTx(db).UpdateOrderStatus(ctx, &sale.UpdateOrderStatusParams{
		OrderID:        order.ID,
		TargetStatus:   models.OrderStatusSuccess,
		PaymentChannel: &paymentChannel,
		PaidAt:         &paidAt,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateOrderPayment: update failed, orderSN=%s, err=%v", orderSN, err)
		return err
	}

	if rowsAffected == 0 {
		logs.WarnContextf(ctx, "UpdateOrderPayment: no rows affected (order already processed or status changed), orderSN=%s", orderSN)
		return nil
	}

	logs.InfoContextf(ctx, "UpdateOrderPayment: order updated successfully, orderSN=%s, rowsAffected=%d", orderSN, rowsAffected)
	return nil
}

func ListPendingOrders(ctx context.Context, duration time.Duration) ([]saletype.SaleOrder, error) {
	db := ManagerDBFromCtx(ctx)
	startTime := time.Now().Add(-duration)
	return sale.NewSaleOrderDao().WithTx(db).GetPendingOrders(ctx, startTime)
}

func CloseOrder(ctx context.Context, orderID uint, reason string) error {
	db := ManagerDBFromCtx(ctx)
	rowsAffected, err := sale.NewSaleOrderDao().WithTx(db).UpdateOrderStatus(ctx, &sale.UpdateOrderStatusParams{
		OrderID:      orderID,
		TargetStatus: models.OrderStatusClosed,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "CloseOrder: update failed, orderID=%d, reason=%s, err=%v", orderID, reason, err)
		return err
	}

	if rowsAffected > 0 {
		logs.InfoContextf(ctx, "CloseOrder: order closed successfully, orderID=%d, reason=%s", orderID, reason)
	} else {
		logs.WarnContextf(ctx, "CloseOrder: order not closed (status mismatch or already processed), orderID=%d, reason=%s", orderID, reason)
	}
	return nil
}
