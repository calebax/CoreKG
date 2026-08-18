package callbacks

import (
	"context"
	"time"

	"github.com/insmtx/corekg/apps/kesale/models"
)

var GlobalHandlers = make(map[string]Handler)

type Handler interface {
	BusinessType() models.BusinessType

	OnPaymentCompleted(ctx context.Context, info *PayInfo) context.Context
}

func AppendGlobalHandlers(handlers ...Handler) {
	for _, h := range handlers {
		bt := h.BusinessType().String()
		if h.BusinessType() == models.BusinessTypeOrder && GlobalHandlers[bt] != nil {
			continue
		}
		GlobalHandlers[bt] = h
	}
}

type PayInfo struct {
	UinID     uint
	CompanyID uint

	// 内部交易信息
	OrderSN         string
	PaymentRecordSN string
	Product         []*models.OrderProduct

	TradeNo       string                // 三方支付订单号
	Channel       models.PaymentChannel // 支付渠道
	PaidAt        time.Time             // 支付时间
	PaymentAmount int64                 // 支付金额（单位：分）

	Metadata map[string]string // 业务自定义信息，例如优惠券ID、商品ID
}
