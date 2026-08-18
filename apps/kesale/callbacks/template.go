package callbacks

import (
	"context"

	"github.com/insmtx/corekg/apps/kesale/models"
)

type PaymentCallbackHandler struct {
	SupportedBusinessType     models.BusinessType
	OnPaymentCompletedHandler func(ctx context.Context, info *PayInfo) context.Context
}

func (h *PaymentCallbackHandler) BusinessType() models.BusinessType {
	return h.SupportedBusinessType
}

func (h *PaymentCallbackHandler) OnPaymentCompleted(ctx context.Context, info *PayInfo) context.Context {
	return h.OnPaymentCompletedHandler(ctx, info)
}
