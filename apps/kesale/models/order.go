package models

type OrderStatus int

const (
	OrderStatusUnknown OrderStatus = iota
	OrderStatusPending
	OrderStatusPaying
	OrderStatusSuccess
	OrderStatusFailed
	OrderStatusCancelled
	OrderStatusClosed
)

func (s OrderStatus) String() string {
	return []string{
		"unknown",
		"pending",
		"paying",
		"success",
		"failed",
		"cancelled",
		"closed",
	}[s]
}

type OrderParams struct {
	Uin uint `json:"uin"`
	// CompanyID 公司ID
	CompanyID uint `json:"company_id"`
	// 订单描述(用于付款显示)
	Description string `json:"description"`
	// 订单金额，单位分
	TotalAmount   int64 `json:"total_amount"`
	PaymentAmount int64 `json:"payment_amount"`

	// 业务类型
	BusinessSource string       `json:"business_source"`
	BusinessType   BusinessType `json:"business_type"`
	// 订单商品列表
	Products []OrderProduct `json:"products"`

	// 支付渠道
	PaymentChannel PaymentChannel `json:"payment_channel"`
	// 订单备注
	Remark string `json:"remark"`

	Metadata map[string]string `json:"metadata"`
}

type OrderProduct struct {
	// 商品ID
	ProductID uint `json:"product_id"`
	// 商品名称
	ProductName string `json:"product_name"`
	// 商品数量
	Quantity uint `json:"quantity"`
	// 商品单价，单位分
	Price int64 `json:"price"`
}

type QueryOrderParams struct {
	Uin       uint   `json:"uin"`
	CompanyID uint   `json:"company_id"`
	OrderSN   string `json:"order_sn"`
}

type PayOrderInfo struct {
	// 订单号
	OrderSN string `json:"order_sn"`
	Amount  int64  `json:"amount"`
	Subject string `json:"subject"`
	// 支付跳转URL
	PayInfo PaymentCredentials `json:"pay_info"`
}
