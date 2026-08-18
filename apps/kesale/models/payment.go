package models

import "time"

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusSuccess PaymentStatus = "success"
	PaymentStatusFailed  PaymentStatus = "failed"
	PaymentStatusUnknown PaymentStatus = "unknown"
)

func (ps PaymentStatus) String() string {
	return string(ps)
}

type PaymentType string

const (
	PaymentTypeURL    PaymentType = "url"
	PaymentTypeQRCode PaymentType = "qrcode"
)

func (pt PaymentType) String() string {
	return string(pt)
}

type PaymentCredentials struct {
	Type PaymentType `json:"type"`
	// 支付跳转URL
	PayURL string `json:"pay_url"`
	// 有效截止时间
	ExpireTime time.Time `json:"expire_time"`
}

type QueryPaymentParams struct {
	Uin       uint   `json:"uin"`
	CompanyID uint   `json:"company_id"`
	RecordSn  string `json:"record_sn"`
	OrderSN   string `json:"order_sn"`
}
