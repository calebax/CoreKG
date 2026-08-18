package client

import (
	"context"
	"net/http"
	"time"

	"github.com/insmtx/corekg/apps/kesale/models"
)

type PaymentClient interface {
	InitClient(options map[string]any) error

	// 发起第三方支付交易 & 返回支付凭证
	CreateTrade(ctx context.Context, req *TradeRequest) (*PrepayResult, error)

	// 查询支付交易状态 (基于渠道交易号)
	QueryTrade(ctx context.Context, outTradeNo string) (paymentResult *CallbackResult, err error)

	// 解析支付回调信息
	HandleCallback(ctx context.Context, req *http.Request) (paymentResult *CallbackResult, callbackResponse any, err error)
}

type (
	TradeRequest struct {
		OutTradeNo     string            // 平台定义的外部交易号
		OutPayRecordSN string            // 平台定义的外部支付单号
		Amount         int64             // 单位：分
		Currency       string            // 货币类型; 默认CNY
		Subject        string            // 标题/商品名
		Extra          map[string]string // 各渠道扩展参数
	}

	PrepayResult struct {
		OutPayRecordSN string `json:"out_pay_record_sn"`
		// 支付跳转URL
		PayURL  string `json:"pay_url"`
		CodeUrl string `json:"code_url"`
		// 有效截止时间
		ExpireTime time.Time `json:"expire_time"`

		RequestRaw  any `json:"-"`
		ResponseRaw any `json:"-"`
	}

	CallbackResult struct {
		Status models.PaymentStatus `json:"status"`
		// 交易单号
		OutTradeNo     string `json:"out_trade_no"`
		OutPayRecordSN string `json:"out_pay_record_sn"`
		TradeNo        string `json:"trade_no"`
		// 支付渠道信息
		PaymentChannel models.PaymentChannel `json:"payment_channel"`
		TradeType      string                `json:"trade_type"`
		BankType       string                `json:"bank_type"`
		//订单金额
		PaymentAmount int64  `json:"payment_amount"`
		Currency      string `json:"currency"`
		//支付时间
		PaidAt *time.Time `json:"paid_at"`

		Extra map[string]string `json:"extra"`

		ResponseRaw any `json:"-"`
	}
)

type TransactionStatus int

const (
	StatusUnknown TransactionStatus = iota
	StatusPending
	StatusSuccess
	StatusFailed
	StatusCancelled
	StatusRefunded
)
