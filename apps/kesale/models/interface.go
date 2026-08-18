package models

type PaymentChannel string

const (
	ChannelWeChatPay PaymentChannel = "wechat"
	ChannelAliPay    PaymentChannel = "alipay"
)

func (p PaymentChannel) String() string {
	return string(p)
}

type BusinessType string

const (
	BusinessTypeOrder        BusinessType = "order"        // 普通订单支付
	BusinessTypeSubscription BusinessType = "subscription" // 订阅支付
)

func (b BusinessType) String() string {
	return string(b)
}
