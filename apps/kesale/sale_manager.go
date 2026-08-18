package kesale

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/insmtx/corekg/apps/kesale/callbacks"
	"github.com/insmtx/corekg/apps/kesale/client"
	"github.com/insmtx/corekg/apps/kesale/models"
	"github.com/insmtx/corekg/apps/kesale/models/saletype"
	"github.com/insmtx/corekg/apps/kesale/services"
	"github.com/insmtx/corekg/apps/kesale/utils/orderno"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type SaleManager struct {
	db         *gorm.DB
	payClients map[models.PaymentChannel]client.PaymentClient
	machineID  int
	bizKey     string
	orderTTL   int
}

// VerifyOrderStatus 验证订单状态，清理超时订单
func (m *SaleManager) VerifyOrderStatus(ctx context.Context) error {
	ctx = services.CtxWithDB(ctx, m.db)

	orders, err := services.ListPendingOrders(ctx, 72*time.Hour)
	if err != nil {
		logs.ErrorContextf(ctx, "VerifyOrderStatus: ListPendingOrders failed, err=%v", err)
		return fmt.Errorf("ListPendingOrders failed: %w", err)
	}

	logs.InfoContextf(ctx, "VerifyOrderStatus: found %d pending orders to verify", len(orders))

	// 遍历每个待处理订单
	for _, order := range orders {
		if err := m.verifyAndCloseOrder(ctx, &order); err != nil {
			logs.ErrorContextf(ctx, "VerifyOrderStatus: verify order failed, orderSN=%s, err=%v", order.OrderSn, err)
			continue
		}
	}
	return nil
}

// verifyAndCloseOrder 验证单个订单并在必要时关闭
func (m *SaleManager) verifyAndCloseOrder(ctx context.Context, order *saletype.SaleOrder) error {
	if order == nil || order.ID == 0 {
		logs.WarnContextf(ctx, "verifyAndCloseOrder: invalid order, orderSN=%s", order.OrderSn)
		return fmt.Errorf("invalid order, orderSN=%s", order.OrderSn)
	}
	records, err := services.ListPaymentRecords(ctx, &models.QueryPaymentParams{
		OrderSN: order.OrderSn,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "verifyAndCloseOrder: ListPaymentRecords failed, orderSN=%s, err=%v", order.OrderSn, err)
		return fmt.Errorf("ListPaymentRecords failed: %w", err)
	}

	// 判断是否超时
	isTimeout := time.Since(order.CreatedAt) > time.Duration(m.orderTTL)*time.Minute

	// 没有找到支付记录
	if len(records) == 0 {
		if isTimeout {
			logs.InfoContextf(ctx, "verifyAndCloseOrder: no payment record found and timeout, closing order, orderSN=%s", order.OrderSn)
			return services.CloseOrder(ctx, order.ID, "no payment record and timeout")
		}
		return nil
	}

	// 提取所有尝试过的支付渠道
	channels := make(map[string]interface{})
	for _, record := range records {
		channels[record.PaymentChannel] = struct{}{}
	}

	hasQueryError := false

	for channelStr, _ := range channels {
		paymentChannel := models.PaymentChannel(channelStr)
		payClient, ok := m.payClients[paymentChannel]
		if !ok {
			logs.WarnContextf(ctx, "verifyAndCloseOrder: payment client not found, orderSN=%s, channel=%s", order.OrderSn, channelStr)
			continue
		}

		// 查询支付平台的交易状态
		queryResult, err := payClient.QueryTrade(ctx, order.OrderSn)
		if err != nil {
			logs.ErrorContextf(ctx, "verifyAndCloseOrder: query trade failed, orderSN=%s, channel=%s, err=%v", order.OrderSn, channelStr, err)
			hasQueryError = true
			continue
		}

		if queryResult.Status == models.PaymentStatusSuccess {
			// 支付成功，触发回调处理
			logs.InfoContextf(ctx, "verifyAndCloseOrder: payment success, processing callback, orderSN=%s, channel=%s", order.OrderSn, channelStr)
			m.processPaymentSuccess(ctx, queryResult)
			return nil
		}
	}

	// (超时 && 无异常)，关闭订单
	if isTimeout && !hasQueryError {
		logs.InfoContextf(ctx, "verifyAndCloseOrder: order timeout and not success (no query errors), closing order, orderSN=%s", order.OrderSn)
		return services.CloseOrder(ctx, order.ID, "timeout and not success")
	}

	return nil
}

func (m *SaleManager) RegisterPaymentClient(channel models.PaymentChannel, payClient client.PaymentClient) (ok bool) {
	if m.payClients == nil {
		m.payClients = make(map[models.PaymentChannel]client.PaymentClient)
	}
	m.payClients[channel] = payClient
	return true
}

// CreateOrder 创建支付订单并返回预支付信息
func (m *SaleManager) CreateOrder(ctx context.Context,
	businessType models.BusinessType, orderParams *models.OrderParams) (*models.PayOrderInfo, error) {
	ctx = services.CtxWithDB(ctx, m.db)

	sn := orderno.Generate(m.bizKey, m.machineID)

	payClient, ok := m.payClients[orderParams.PaymentChannel]
	if !ok {
		return nil, fmt.Errorf("payment channel %s not supported", orderParams.PaymentChannel)
	}

	logs.InfoContextf(ctx, "CreateOrder: %v %v", sn, orderParams.PaymentChannel.String())
	orderParams.BusinessType = businessType
	orderInfo, err := services.CreateOrder(ctx, sn, orderParams)
	if err != nil {
		logs.ErrorContextf(ctx, "create order failed: %v", err)
		return nil, fmt.Errorf("create order failed: %w", err)
	}

	// 调用支付渠道
	logs.InfoContextf(ctx, "CreatePaymentTrade: %v %v", orderInfo.OrderSn, orderParams.PaymentChannel.String())
	payRecordSN := orderno.GeneratePaymentTradeNo()
	payResult, err := payClient.CreateTrade(ctx, &client.TradeRequest{
		OutTradeNo:     sn,
		OutPayRecordSN: payRecordSN,
		Amount:         orderInfo.PaymentAmount,
		Subject:        orderParams.Description,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "OrderSn: %v CreatePaymentTrade failed: %v", orderInfo.OrderSn, err)
		return nil, fmt.Errorf("CreatePaymentTrade failed: %w", err)
	}

	// 支付调用记录
	err = services.AddPaymentRecord(ctx, orderInfo, payResult)
	if err != nil {
		// 优先返回支付信息，先支付。
		logs.ErrorContextf(ctx, "AddPaymentRecord failed: %v", err)
	}

	payType := models.PaymentTypeURL
	payURL := payResult.PayURL
	if qrcode := payResult.CodeUrl; qrcode != "" {
		payType = models.PaymentTypeQRCode
		payURL = qrcode
	}
	return &models.PayOrderInfo{
		OrderSN: orderInfo.OrderSn,
		Amount:  orderInfo.PaymentAmount,
		Subject: orderParams.Description,
		PayInfo: models.PaymentCredentials{
			Type:       payType,
			PayURL:     payURL,
			ExpireTime: payResult.ExpireTime,
		},
	}, nil
}

// QueryOrderStatus 查询订单支付状态
func (m *SaleManager) QueryOrderStatus(ctx context.Context, params *models.QueryOrderParams) (models.PaymentStatus, error) {
	ctx = services.CtxWithDB(ctx, m.db)
	order, err := services.FindOrder(ctx, params)
	if err != nil {
		return models.PaymentStatusUnknown, err
	}
	if order == nil {
		return models.PaymentStatusUnknown, fmt.Errorf("order not found")
	}

	if order.OrderStatus == models.OrderStatusSuccess.String() {
		return models.PaymentStatusSuccess, nil
	} else if order.OrderStatus == models.OrderStatusClosed.String() {
		return models.PaymentStatusFailed, nil
	}
	return models.PaymentStatusPending, nil
}

func (m *SaleManager) HandlePaymentCallback(ctx context.Context, channel models.PaymentChannel, req *http.Request) (callbackResponse any, err error) {
	ctx = services.CtxWithDB(ctx, m.db)
	payClient, ok := m.payClients[channel]
	if !ok {
		return nil, fmt.Errorf("HandlePaymentCallback: payment channel %s not supported", channel)
	}
	// 解析支付回调
	paymentResult, callbackResponse, err := payClient.HandleCallback(ctx, req)
	if err != nil {
		return callbackResponse, fmt.Errorf("HandlePaymentCallback: handle payment callback failed: %w, paymentResult=%v", err, paymentResult)
	}
	if paymentResult == nil {
		logs.ErrorContextf(ctx, "HandlePaymentCallback: payment result is nil, channel=%s", channel)
	} else {
		logs.InfoContextf(ctx, "HandlePaymentCallback: payment status=%s, OrderSN=%s", paymentResult.Status, paymentResult.OutTradeNo)
	}

	if paymentResult != nil && paymentResult.Status == models.PaymentStatusSuccess {
		// 支付完成
		// 使用 WithoutCancel 创建新 context，保留原有值（如 trace ID、logger）但不受父 context 取消影响
		asyncCtx := context.WithoutCancel(ctx)
		go func(pr *client.CallbackResult, actx context.Context) {
			defer func() {
				if r := recover(); r != nil {
					data, _ := json.Marshal(pr)
					logs.ErrorContextf(actx, "HandlePaymentCallback: [payment] panic in callback: %v\npaymentResult=%s\n", r, string(data))
				}
			}()
			m.processPaymentSuccess(actx, pr)
		}(paymentResult, asyncCtx)
	}

	return callbackResponse, nil
}

// 支付回调信息解析后，业务处理
func (m *SaleManager) processPaymentSuccess(ctx context.Context, result *client.CallbackResult) {
	start := time.Now()
	printLog := func(level string, format string, args ...interface{}) {
		resultData, _ := json.Marshal(result)
		msg := fmt.Sprintf(format, args...)
		msg = fmt.Sprintf("HandlePaymentCallback: OrderSN=%s processPaymentSuccess: %s, paymentResult=%s", result.OutTradeNo, msg, string(resultData))
		switch level {
		case "info":
			logs.InfoContextf(ctx, msg)
		case "warn":
			logs.WarnContextf(ctx, msg)
		case "error":
			logs.ErrorContextf(ctx, msg)
		default:
			logs.InfoContextf(ctx, msg)
		}
	}
	printLog("info", "Start processing payment success callback")
	ctx = services.CtxWithDB(ctx, m.db)

	lockKey := fmt.Sprintf("kesale:payment:callback:lock:%s", result.OutTradeNo)
	ok := redispool.LockWithTimeout(
		lockKey,
		30*time.Second,       // 锁过期时间：30s
		100*time.Millisecond, // 重试间隔：100ms
		5*time.Second,        // 获取锁超时时间：5s
	)
	if !ok {
		printLog("warn", "failed to acquire lock, another callback is processing")
		return
	}
	defer redispool.UnLock(lockKey)

	order, err := services.FindOrder(ctx, &models.QueryOrderParams{OrderSN: result.OutTradeNo})
	if err != nil {
		printLog("error", "FindOrder failed, err=%v", err)
		return
	} else if order == nil {
		printLog("warn", "Order not found")
		return
	}

	// 优先更新支付单状态
	ok, err = services.UpdatePaymentRecord(ctx, order, result)
	if err != nil {
		printLog("error", "UpdatePaymentRecord failed (internal error), err=%v", err)
		return
	} else if !ok {
		printLog("warn", "UpdatePaymentRecord skipped (status not matched or already processed)")
		return
	}

	if order.OrderStatus == models.OrderStatusSuccess.String() {
		// 已处理
		printLog("info", "order already success")
		return
	}

	// 触发业务逻辑回调
	handles := []callbacks.Handler{}
	if handler, ok := callbacks.GlobalHandlers[models.BusinessTypeOrder.String()]; ok {
		handles = append(handles, handler)
	}
	if handler, ok := callbacks.GlobalHandlers[order.BusinessType]; ok {
		handles = append(handles, handler)
	} else {
		printLog("warn", "No handler found for business type %s", order.BusinessType)
	}
	if len(handles) == 0 {
		printLog("warn", "No handlers found")
		return
	}

	items, err := services.ListOrderItems(ctx, order.ID)
	if err != nil {
		printLog("error", "ListOrderItems failed, err=%v", err)
		return
	}

	payInfo := &callbacks.PayInfo{
		UinID:         order.Uin,
		CompanyID:     order.CompanyID,
		OrderSN:       result.OutTradeNo,
		Product:       items,
		TradeNo:       result.TradeNo,
		Channel:       result.PaymentChannel,
		PaidAt:        *result.PaidAt,
		PaymentAmount: result.PaymentAmount,
	}
	callbacks.OnPaymentCompletedHandler(ctx, payInfo, handles)

	elapsed := time.Since(start)
	printLog("info", "callback processed successfully, elapsed=%s", elapsed)
}

type SaleConfig struct {
	DeployMachineID int `json:"deploy_machine_id" yaml:"deploy_machine_id"`
	// 订单前缀 默认KE
	OrderPrefix string `json:"order_prefix" yaml:"order_prefix"`
	// 支付平台配置
	Platforms []PlatformConfig `json:"platforms" yaml:"platforms"`
	// 订单超时时间
	OrderTTL int `json:"order_ttl" yaml:"order_ttl"`
}

type PlatformConfig struct {
	// 支付类型 models.PaymentChannel
	PayType string         `json:"pay_type" yaml:"pay_type"`
	Params  map[string]any `json:"params" yaml:"params"`
}
