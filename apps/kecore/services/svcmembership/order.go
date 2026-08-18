package svcmembership

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtomembership"
	"github.com/insmtx/corekg/apps/kesale/models/sale"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

func ListPaymentOrderRecord(ctx *gin.Context, req *dtomembership.ListPaymentOrderRecordRequest) (res *dtomembership.ListPaymentOrderRecordResponse, err error) {
	res = &dtomembership.ListPaymentOrderRecordResponse{}
	req.Request.Uin = runtime.Uin(ctx)

	ss, i, err := sale.NewSaleOrderDao().GetPageListByCond(ctx, &sale.SaleOrderCond{
		Filters: req.Request.Filters,
		BaseCond: sale.BaseCond{
			Limit:     req.Request.Limit,
			Offset:    req.Request.Offset,
			OrderBy:   req.Request.OrderBy,
			BeginTime: req.Request.BeginTime,
			EndTime:   req.Request.EndTime,
			Uin:       req.Request.Uin,
		},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "GetPageListByCond(%+v) failed err: %v", req, err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "admin_list_payment_record_failed"
		return res, nil
	}

	data := make([]dtomembership.RecordItem, 0, len(ss))
	for _, v := range ss {
		data = append(data, dtomembership.RecordItem{
			ID:        v.ID,
			Uin:       v.Uin,
			CompanyID: v.CompanyID,
			Amount:    v.PaymentAmount,
			OrderSn:   v.OrderSn,
			Status:    v.OrderStatus,
			CreatedAt: v.CreatedAt,
			PaidAt:    v.PaidAt,
		})
	}

	res.Response.Total = i
	res.Response.Offset = req.Request.Offset
	res.Response.Limit = req.Request.Limit
	res.Response.Data = data
	return res, nil
}
