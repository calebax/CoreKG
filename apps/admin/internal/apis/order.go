package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/internal/dto/dtoorder"
	"github.com/insmtx/corekg/apps/admin/services/svcorder"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// ListPaymentOrderRecord 查询支付订单列表
// @Tags 支付订单
// @Summary 查询支付订单列表
// @Description 查询支付订单列表
// @Router /admin.ListPaymentOrderRecord [post]
// @Param request body dtoorder.ListPaymentOrderRecordRequest true "request"
// @Success 200 {object} dtoorder.ListPaymentOrderRecordResponse "response"
func ListPaymentOrderRecord(ctx *gin.Context, req *dtoorder.ListPaymentOrderRecordRequest, resp *dtoorder.ListPaymentOrderRecordResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListPaymentOrderRecord] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcorder.ListPaymentOrderRecord(ctx, req)
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
