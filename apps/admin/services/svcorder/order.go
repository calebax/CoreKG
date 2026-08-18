package svcorder

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/admin/internal/dto/dtoorder"
	"github.com/insmtx/corekg/apps/kesale/models/sale"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

func ListPaymentOrderRecord(ctx *gin.Context, req *dtoorder.ListPaymentOrderRecordRequest) (res *dtoorder.ListPaymentOrderRecordResponse, err error) {
	res = &dtoorder.ListPaymentOrderRecordResponse{}

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

	uinIDs := make([]uint, 0, len(ss))
	for _, v := range ss {
		uinIDs = append(uinIDs, v.Uin)
	}

	uins, err := user.GetUinCompanyByUins(ctx, uinIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "GetUinCompanyByUins(ids:%v)", uinIDs)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "account_get_uin_failed"
		return res, nil
	}

	uinMaps := utils.ToMap(uins, func(v *user.UinCompanyItem) uint {
		return v.ID
	})

	data := make([]dtoorder.RecordItem, 0, len(ss))
	for _, v := range ss {
		data = append(data, dtoorder.RecordItem{
			ID:          v.ID,
			Uin:         v.Uin,
			UserName:    uinMaps[v.Uin].Name,
			CompanyID:   v.CompanyID,
			CompanyName: uinMaps[v.Uin].CompanyName,
			Amount:      v.PaymentAmount,
			OrderSn:     v.OrderSn,
			Status:      v.OrderStatus,
			CreatedAt:   v.CreatedAt,
			PaidAt:      v.PaidAt,
		})
	}
	res.Response.Total = i
	res.Response.Offset = req.Request.Offset
	res.Response.Limit = req.Request.Limit
	res.Response.Data = data

	return res, nil
}
