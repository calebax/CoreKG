package kesale

import (
	"testing"

	"github.com/insmtx/corekg/apps/kesale/models"
	"github.com/insmtx/corekg/apps/kesale/services"
	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/ygpkg/yg-go/apis/runtime"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
)

func TestCreateOrder2Pay(t *testing.T) {

	testutils.Initialize(testutils.AppNameKesale)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(810))

	companyID := runtime.CompanyID(ctx)
	uin := runtime.Uin(ctx)
	t.Log(companyID)
	t.Log(uin)

	Init(dbtools.Core(), "core", "kesale_pay")

	var orderSN string = "KE25112706461745257"
	orderSN = "KE25112708071040704"
	t.Run("CreateOrder", func(t *testing.T) {
		orderInfo, err := Manager().CreateOrder(ctx, models.BusinessTypeSubscription, &models.OrderParams{
			CompanyID:      companyID,
			Uin:            uin,
			PaymentChannel: models.ChannelWeChatPay,
			Description:    "test pay",
			// 0.01元
			TotalAmount:   1,
			PaymentAmount: 1,
			Products: []models.OrderProduct{
				{
					ProductID: 1,
					Quantity:  1,
					Price:     1,
				},
			},
		})

		if err != nil {
			t.Fatal(err)
		}

		logs.Infof("orderInfo: %s", orderInfo)
		orderSN = orderInfo.OrderSN
	})

	t.Run("QueryOrderStatus", func(t *testing.T) {
		orderStatus, err := Manager().QueryOrderStatus(ctx, &models.QueryOrderParams{
			OrderSN: orderSN,
		})
		if err != nil {
			t.Fatal(err)
		}
		logs.Infof("orderStatus: %s", orderStatus.String())
	})

	t.Run("VerifySingleOrderPayStatus", func(t *testing.T) {
		orderSN := "KE25112611281807241"
		order, err := services.FindOrder(ctx, &models.QueryOrderParams{
			OrderSN: orderSN,
		})
		if err != nil || order.ID == 0 {
			t.Fatal(err)
		}
		err = Manager().verifyAndCloseOrder(ctx, order)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("VerifyOrderPayStatus", func(t *testing.T) {
		Manager().VerifyOrderStatus(ctx)
	})

	t.Run("VerifyNotFoundOrderPayStatus", func(t *testing.T) {
		orderSN := "KE2511270807104070400"
		orderSN = "KE25112706461745257"
		// orderSN = "KE25112708071040704"
		result, err := Manager().payClients[models.ChannelWeChatPay].QueryTrade(ctx, orderSN)
		if err != nil {
			t.Fatal(err)
		}
		logs.Infof("result: %s", result)
	})

}
