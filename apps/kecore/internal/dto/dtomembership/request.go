package dtomembership

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type ListPackageRequest struct {
	apiobj.BaseRequest
	Request ListPackageEmbedRequest
}

type ListPackageEmbedRequest struct {
}

func (opt *ListPackageRequest) Validity(resp *ListPackageResponse) {
}

type CreateOrderRequest struct {
	apiobj.BaseRequest
	Request CreateOrderEmbedRequest
}
type CreateOrderEmbedRequest struct {
	// PackageID 套餐ID
	PackageID uint `json:"package_id" validate:"required"`
}

func (opt *CreateOrderRequest) Validity(resp *CreateOrderResponse) {
	if opt.Request.PackageID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_package_id_required"
		return
	}
}

type QueryOrderStatusRequest struct {
	apiobj.BaseRequest
	Request QueryOrderStatusEmbedRequest
}
type QueryOrderStatusEmbedRequest struct {
	// OrderSN 订单号
	OrderSN string `json:"order_sn" validate:"required"`
}

func (opt *QueryOrderStatusRequest) Validity(resp *QueryOrderStatusResponse) {
	if opt.Request.OrderSN == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_order_sn_required"
		return
	}
}

type ListPaymentOrderRecordRequest struct {
	apiobj.BaseRequest
	Request ListPaymentOrderRecordEmbedRequest
}

type ListPaymentOrderRecordEmbedRequest struct {
	apiobj.PageQuery
}

func (opt *ListPaymentOrderRecordRequest) Validity(_ *ListPaymentOrderRecordResponse) {
}
