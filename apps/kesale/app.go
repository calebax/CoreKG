package kesale

import (
	"context"
	"sync"

	"github.com/insmtx/corekg/apps/kesale/callbacks"
	"github.com/insmtx/corekg/apps/kesale/client"
	"github.com/insmtx/corekg/apps/kesale/models"
	"github.com/insmtx/corekg/apps/kesale/models/sale"
	dbtype "github.com/insmtx/corekg/apps/kesale/models/saletype"
	"github.com/insmtx/corekg/apps/kesale/services"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"gorm.io/gorm"
)

func Init(db *gorm.DB, group, key string) error {
	cfg := &SaleConfig{}
	err := settings.GetYaml(group, key, cfg)
	if err != nil {
		logs.Errorf("[main] get sale config failed, %s", err)
		return err
	}

	// 初始化sale包的数据库连接
	sale.InitDB(db)

	// 初始化数据库
	if err := dbtype.InitDB(db); err != nil {
		logs.Errorf("[main] init sale database failed, %s", err)
		return err
	}

	ttl := cfg.OrderTTL
	if ttl == 0 {
		ttl = 15
	}
	saleManager = &SaleManager{
		db:        db,
		bizKey:    cfg.OrderPrefix,
		machineID: cfg.DeployMachineID,
		orderTTL:  ttl,
	}
	// 初始化支付
	for _, platform := range cfg.Platforms {
		var PaymentChannel models.PaymentChannel
		var payClient client.PaymentClient

		switch platform.PayType {
		case models.ChannelWeChatPay.String():
			PaymentChannel = models.ChannelWeChatPay
			payClient = &client.WeChatPaymentClient{}
		case models.ChannelAliPay.String():
			// TODO: AliPay
		}

		if payClient != nil {
			err := payClient.InitClient(platform.Params)
			if err != nil {
				logs.Errorf("[main] init payment client failed, %s", err)
				return err
			}
			saleManager.RegisterPaymentClient(PaymentChannel, payClient)
		}
	}
	// 初始化订单回调
	callbacks.AppendGlobalHandlers(&callbacks.PaymentCallbackHandler{
		SupportedBusinessType: models.BusinessTypeOrder,
		OnPaymentCompletedHandler: func(ctx context.Context, info *callbacks.PayInfo) context.Context {
			ctx = services.CtxWithDB(ctx, db)
			services.UpdateOrderPayment(ctx, info.OrderSN, info.Channel, info.PaymentAmount, info.PaidAt)
			return ctx
		},
	})

	RunJob()
	logs.Info("sale manager initialized")
	return nil
}

var saleManager *SaleManager

func Manager() *SaleManager {
	if saleManager == nil {
		logs.Warn("sale manager is nil")
		saleManager = &SaleManager{}
	}
	return saleManager
}

var onceStart sync.Once

func RunJob() error {
	onceStart.Do(func() {
		go runCornJob()
	})
	return nil
}
