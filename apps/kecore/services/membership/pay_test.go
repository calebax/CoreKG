package membership

import (
	"testing"
	"time"

	"github.com/insmtx/corekg/apps/kesale/callbacks"
	"github.com/insmtx/corekg/apps/kesale/models"
	"github.com/insmtx/corekg/pkgs/testutils"
)

func TestPayCallback(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))
	payInfo := &callbacks.PayInfo{
		UinID:           384,
		CompanyID:       2,
		OrderSN:         "KE25112708071040704",
		PaymentRecordSN: "PAY25112708071040704",
		Product: []*models.OrderProduct{
			{
				ProductID:   2, // 套餐ID，需要对应数据库中存在的套餐
				ProductName: "测试套餐",
				Quantity:    1,
				Price:       9900, // 99元，单位：分
			},
		},
		TradeNo:       "TRADE2025112708071040704",
		Channel:       models.ChannelWeChatPay,
		PaidAt:        time.Now(),
		PaymentAmount: 9900, // 支付金额，单位：分
		Metadata: map[string]string{
			"source": "test",
		},
	}
	NewPaymentHandler().OnPaymentCompleted(ctx, payInfo)
}
