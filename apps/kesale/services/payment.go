package services

import (
	"context"
	"encoding/json"

	"github.com/insmtx/corekg/apps/kesale/client"
	"github.com/insmtx/corekg/apps/kesale/models"
	"github.com/insmtx/corekg/apps/kesale/models/sale"
	"github.com/insmtx/corekg/apps/kesale/models/saletype"
	"github.com/ygpkg/yg-go/logs"
)

// AddPaymentRecord 支付调用记录
func AddPaymentRecord(ctx context.Context, order *saletype.SaleOrder, payResult *client.PrepayResult) error {
	requestJSON, _ := json.Marshal(payResult.RequestRaw)
	responseJSON, _ := json.Marshal(payResult.ResponseRaw)

	record := &saletype.SalePaymentRecord{
		OrderID:        order.ID,
		OrderSn:        order.OrderSn,
		CompanyID:      order.CompanyID,
		Uin:            order.Uin,
		RecordSn:       payResult.OutPayRecordSN,
		RecordType:     "payment",
		BusinessSource: order.BusinessSource,
		Amount:         order.PaymentAmount,
		PaymentChannel: order.PaymentChannel,
		Status:         models.PaymentStatusPending.String(),
		RequestParams:  string(requestJSON),
		ResponseData:   string(responseJSON),
	}

	db := ManagerDBFromCtx(ctx)
	return sale.NewSalePaymentRecordDao().WithTx(db).Insert(ctx, record)
}

func FindPaymentRecord(ctx context.Context, params *models.QueryPaymentParams) (*saletype.SalePaymentRecord, error) {
	db := ManagerDBFromCtx(ctx)
	record, err := sale.NewSalePaymentRecordDao().WithTx(db).GetByCond(ctx, &sale.SalePaymentRecordCond{
		OrderSN:  params.OrderSN,
		RecordSn: params.RecordSn,
	})
	if err != nil {
		return nil, err
	}
	return record, nil
}

func ListPaymentRecords(ctx context.Context, params *models.QueryPaymentParams) ([]saletype.SalePaymentRecord, error) {
	db := ManagerDBFromCtx(ctx)
	return sale.NewSalePaymentRecordDao().WithTx(db).GetListByCond(ctx, &sale.SalePaymentRecordCond{
		OrderSN:  params.OrderSN,
		RecordSn: params.RecordSn,
	})
}

func QueryPaymentStatus(ctx context.Context, params *models.QueryPaymentParams) (models.PaymentStatus, error) {
	db := ManagerDBFromCtx(ctx)
	recordList, err := sale.NewSalePaymentRecordDao().WithTx(db).GetListByCond(ctx, &sale.SalePaymentRecordCond{
		OrderSN: params.OrderSN,
	})
	if err != nil {
		return models.PaymentStatusUnknown, err
	}

	successStatus := models.PaymentStatusSuccess.String()
	for _, record := range recordList {
		if record.Status == successStatus {
			return models.PaymentStatusSuccess, nil
		}
	}
	return models.PaymentStatusPending, nil
}

func UpdatePaymentRecord(ctx context.Context, order *saletype.SaleOrder, payResult *client.CallbackResult) (ok bool, err error) {
	db := ManagerDBFromCtx(ctx)
	record, err := FindPaymentRecord(ctx, &models.QueryPaymentParams{
		OrderSN:  payResult.OutTradeNo,
		RecordSn: payResult.OutPayRecordSN,
	})
	if err != nil {
		return false, err
	}

	if record.Status == models.PaymentStatusSuccess.String() {
		logs.InfoContextf(ctx, "UpdatePaymentRecord: record already exists, orderSN: %s, recordSN: %s, oldStatus: %s", payResult.OutTradeNo, payResult.OutPayRecordSN, record.Status)
		return false, nil
	}

	if len(record.CallbackData) > 0 {
		logs.InfoContextf(ctx, "UpdatePaymentRecord: record already exists, orderSN: %s, recordSN: %s, oldCallbackData: %s", payResult.OutTradeNo, payResult.OutPayRecordSN, record.CallbackData)
	}

	// 更新支付记录
	responseJSON, _ := json.Marshal(payResult.ResponseRaw)
	record.Status = payResult.Status.String()
	record.TradeNo = payResult.TradeNo
	record.PaidAt = payResult.PaidAt
	record.CallbackData = string(responseJSON)

	// 执行数据库更新
	err = sale.NewSalePaymentRecordDao().WithTx(db).UpdateByID(ctx, record.ID, record)
	if err != nil {
		return false, err
	}

	return true, nil
}
