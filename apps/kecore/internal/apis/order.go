package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtomembership"
	"github.com/insmtx/corekg/apps/kecore/services/svcmembership"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// CreateOrder 创建订单
// @Tags 支付管理
// @Summary 创建订单
// @Description 创建订单
// @Router /forest.CreateOrder [post]
// @Param request body dtomembership.CreateOrderRequest true "request"
// @Success 200 {object} dtomembership.CreateOrderResponse "response"
func CreateOrder(ctx *gin.Context, req *dtomembership.CreateOrderRequest, resp *dtomembership.CreateOrderResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreateOrder] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svcmembership.CreateOrder(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateOrder] svcmembership.CreateOrder failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_order_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// QueryOrderStatus 订单状态查询
// @Tags 支付管理
// @Summary 订单状态查询
// @Description 订单状态查询
// @Router /forest.QueryOrderStatus [post]
// @Param request body dtomembership.QueryOrderStatusRequest true "request"
// @Success 200 {object} dtomembership.QueryOrderStatusResponse "response"
func QueryOrderStatus(ctx *gin.Context, req *dtomembership.QueryOrderStatusRequest, resp *dtomembership.QueryOrderStatusResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[QueryOrderStatus] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svcmembership.QueryOrderStatus(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[QueryOrderStatus] svcmembership.QueryOrderStatus failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_order_status_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ListPaymentOrderRecord 获取支付记录
// @Tags 支付
// @Summary 获取支付记录
// @Description 获取支付记录
// @Router /forest.ListPaymentOrderRecord [post]
// @Param request body dtomembership.ListPaymentOrderRecordRequest true "request"
// @Success 200 {object} dtomembership.ListPaymentOrderRecordResponse "response"
func ListPaymentOrderRecord(ctx *gin.Context, req *dtomembership.ListPaymentOrderRecordRequest, resp *dtomembership.ListPaymentOrderRecordResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListPaymentOrderRecord] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcmembership.ListPaymentOrderRecord(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListPaymentOrderRecord] svcorder.ListPaymentOrderRecord failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "admin_list_payment_record_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
