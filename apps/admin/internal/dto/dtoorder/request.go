package dtoorder

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type ListPaymentOrderRecordRequest struct {
	apiobj.BaseRequest
	Request ListPaymentOrderRecordEmbedRequest
}

type ListPaymentOrderRecordEmbedRequest struct {
	apiobj.PageQuery
}

func (opt *ListPaymentOrderRecordRequest) Validity(_ *ListPaymentOrderRecordResponse) {
}
