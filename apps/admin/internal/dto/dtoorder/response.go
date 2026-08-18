package dtoorder

import (
	"time"

	"github.com/ygpkg/yg-go/apis/apiobj"
)

type ListPaymentOrderRecordResponse struct {
	apiobj.BaseResponse
	Response ListPaymentOrderRecordEmbedResponse
}

type ListPaymentOrderRecordEmbedResponse struct {
	Data []RecordItem `json:"data"`
	apiobj.QueryResponse
}

type RecordItem struct {
	ID          uint   `json:"id"`
	Uin         uint   `json:"uin"`
	UserName    string `json:"username"`
	CompanyID   uint   `json:"company_id"`
	CompanyName string `json:"company_name"`
	//金额(分)
	Amount  int64  `json:"amount"`
	OrderSn string `json:"order_sn"`
	//流水状态：pending-处理中，success-成功，failed-失败
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	PaidAt    *time.Time `json:"paid_at"`
}
