package svcmembership

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtomembership"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kesale"
	"github.com/insmtx/corekg/apps/kesale/models"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

func CreateOrder(ctx *gin.Context, req *dtomembership.CreateOrderRequest) (res *dtomembership.CreateOrderResponse, err error) {
	companyID := runtime.CompanyID(ctx)
	uin := runtime.Uin(ctx)
	// 统计已购买的配额记录
	now := time.Now()
	purchasedCount, err := forest.NewKeCompanyQuotaDao().CountByCond(ctx, &forest.KeCompanyQuotaCond{
		CompanyID:     companyID,
		SourceType:    foresttype.CompanyQuotaSourceTypeOrder,
		ExpireAtStart: &now,
	})
	if err != nil {
		return nil, err
	}

	res = &dtomembership.CreateOrderResponse{}

	if purchasedCount > 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_package_already_purchased"
		return res, nil
	}

	packageEntity, err := forest.NewKePackageDao().GetByID(ctx, req.Request.PackageID)
	if err != nil {
		return nil, err
	}
	if packageEntity == nil || packageEntity.ID == 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_package_not_found"
		return res, nil
	}

	var orderProducts []models.OrderProduct
	orderProducts = append(orderProducts, models.OrderProduct{
		ProductID:   packageEntity.ID,
		ProductName: packageEntity.Name,
		Quantity:    1,
		Price:       packageEntity.SalePrice,
	})
	payOrderInfo, err := kesale.Manager().CreateOrder(ctx, models.BusinessTypeSubscription, &models.OrderParams{
		CompanyID:      companyID,
		Uin:            uin,
		Description:    packageEntity.Name,
		BusinessSource: global.SaleBusinessSourceKnowledge,
		BusinessType:   models.BusinessTypeSubscription,
		TotalAmount:    packageEntity.Price,
		PaymentAmount:  packageEntity.SalePrice,
		PaymentChannel: models.ChannelWeChatPay,
		Products:       orderProducts,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateOrder] kesale CreateOrder fail, err: %v", err)
		return nil, err
	}
	res.Response = dtomembership.CreateOrderEmbedResponse{
		OrderSN:    payOrderInfo.OrderSN,
		PayURL:     payOrderInfo.PayInfo.PayURL,
		ExpireTime: payOrderInfo.PayInfo.ExpireTime,
	}
	return res, nil
}

func QueryOrderStatus(ctx *gin.Context, req *dtomembership.QueryOrderStatusRequest) (res *dtomembership.QueryOrderStatusResponse, err error) {
	status, err := kesale.Manager().QueryOrderStatus(ctx, &models.QueryOrderParams{
		OrderSN: req.Request.OrderSN,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[QueryOrderStatus] kesale QueryOrderStatus fail, orderSn: %s, err: %v", req.Request.OrderSN, err)
		return nil, err
	}
	res = &dtomembership.QueryOrderStatusResponse{
		Response: dtomembership.QueryOrderStatusEmbedResponse{
			Status: status.String(),
		},
	}
	return res, nil
}
