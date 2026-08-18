package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
	"github.com/insmtx/corekg/apps/kesale/models"
	"github.com/ygpkg/yg-go/logs"
)

type WeChatPaymentClient struct {
	client     *wechat.ClientV3
	appId      string
	apiV3Key   string
	timeExpire time.Duration
	notifyURL  string
	currency   string
}

type wechatPaymentOptions struct {
	AppId      string `json:"app_id"`
	MchId      string `json:"mch_id"`
	SerialNo   string `json:"serial_no"`
	ApiV3Key   string `json:"api_v3_key"`
	PrivateKey string `json:"private_key"`
	NotifyURL  string `json:"notify_url"`
	// 超时时间/分钟
	TimeExpire      uint   `json:"time_expire"`
	DefaultCurrency string `json:"default_currency"`
}

func (w *WeChatPaymentClient) InitClient(options map[string]any) error {
	clientOptions, ok := fillConfig(options)
	if !ok {
		return fmt.Errorf("invalid wechat payment options")
	}

	logs.InfoContextf(context.Background(), "[WeChatPaymentClient.InitClient] mch_id: %s", clientOptions.MchId)

	client, err := wechat.NewClientV3(clientOptions.MchId, clientOptions.SerialNo, clientOptions.ApiV3Key, clientOptions.PrivateKey)
	if err != nil {
		return err
	}
	err = client.AutoVerifySign()
	if err != nil {
		return err
	}
	w.client = client

	w.appId = clientOptions.AppId
	w.apiV3Key = clientOptions.ApiV3Key
	w.timeExpire = time.Duration(clientOptions.TimeExpire) * time.Minute
	w.notifyURL = clientOptions.NotifyURL
	w.currency = clientOptions.DefaultCurrency

	return nil
}

func (w *WeChatPaymentClient) CreateTrade(ctx context.Context, req *TradeRequest) (*PrepayResult, error) {
	bm := make(gopay.BodyMap)
	expireTime := time.Now().Add(w.timeExpire)

	bm.Set("appid", w.appId).
		Set("description", req.Subject).
		Set("out_trade_no", req.OutTradeNo).
		Set("notify_url", w.notifyURL).
		SetBodyMap("amount", func(bm gopay.BodyMap) {
			bm.Set("total", req.Amount)
			if len(req.Currency) > 0 {
				bm.Set("currency", req.Currency)
			} else {
				bm.Set("currency", w.currency)
			}
		})
	if req.Extra == nil {
		req.Extra = make(map[string]string)
	}
	req.Extra[ExtraKeyPaySN] = req.OutPayRecordSN
	attach, _ := json.Marshal(req.Extra)
	bm.Set("attach", string(attach))

	if w.timeExpire > 0 {
		expire := time.Now().Add(w.timeExpire).Format(time.RFC3339)
		bm.Set("time_expire", expire)
	}

	wxRsp, err := w.client.V3TransactionNative(ctx, bm)
	if err != nil {
		return nil, err
	}

	if wxRsp.Code == wechat.Success {
		return &PrepayResult{
			CodeUrl:        wxRsp.Response.CodeUrl,
			OutPayRecordSN: req.OutPayRecordSN,
			RequestRaw:     bm,
			ResponseRaw:    wxRsp.Response,
			ExpireTime:     expireTime,
		}, nil
	}

	return nil, fmt.Errorf("create trade failed: %s", wxRsp.Error)
}

func (w *WeChatPaymentClient) QueryTrade(ctx context.Context, outTradeNo string) (paymentResult *CallbackResult, err error) {
	paymentResult = &CallbackResult{
		PaymentChannel: models.ChannelWeChatPay,
		OutTradeNo:     outTradeNo,
		Status:         models.PaymentStatusPending,
	}
	wxRsp, err := w.client.V3TransactionQueryOrder(ctx, wechat.OutTradeNo, outTradeNo)
	if err != nil {
		return paymentResult, err
	}

	if wxRsp.Code == http.StatusNotFound {
		return paymentResult, nil
	}
	if wxRsp.Code != wechat.Success {
		paymentResult.Status = models.PaymentStatusUnknown
		return paymentResult, fmt.Errorf("query trade failed: %s", wxRsp.Error)
	}

	if wxRsp.Code == wechat.Success {
		wxPayResult := wxRsp.Response
		paymentResult, err = w.buildCallbackResult(wxPayResult, wxRsp)
		if err != nil {
			paymentResult.Status = models.PaymentStatusUnknown
			return paymentResult, err
		}
		return paymentResult, nil
	}

	return paymentResult, fmt.Errorf("query trade failed: %s", wxRsp.Error)
}

func (w *WeChatPaymentClient) HandleCallback(ctx context.Context, req *http.Request) (paymentResult *CallbackResult, callbackResponse any, err error) {
	callbackResponse = &wechat.V3NotifyRsp{Code: gopay.FAIL, Message: "失败"}

	notifyReq, err := wechat.V3ParseNotify(req)
	if err != nil {
		return &CallbackResult{ResponseRaw: notifyReq}, callbackResponse, err
	}

	certMap := w.client.WxPublicKeyMap()
	err = notifyReq.VerifySignByPKMap(certMap)
	if err != nil {
		return &CallbackResult{ResponseRaw: notifyReq}, callbackResponse, err
	}

	// TODO 当前仅当处理普通支付回调
	wxPayResult, err := notifyReq.DecryptPayCipherText(w.apiV3Key)
	if err != nil {
		return &CallbackResult{ResponseRaw: notifyReq}, callbackResponse, err
	}

	paymentResult, err = w.buildCallbackResult(wxPayResult, notifyReq)
	if err != nil {
		return &CallbackResult{ResponseRaw: notifyReq}, callbackResponse, err
	}

	callbackResponse = &wechat.V3NotifyRsp{Code: gopay.SUCCESS, Message: "成功"}
	return paymentResult, callbackResponse, nil
}

const (
	ExtraKeyPaySN = "paysn" // 内部支付流水号
)

func fillConfig(params map[string]any) (options *wechatPaymentOptions, ok bool) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, false
	}
	var clientOptions wechatPaymentOptions
	if err := json.Unmarshal(data, &clientOptions); err != nil {
		return nil, false
	}
	if len(clientOptions.DefaultCurrency) == 0 {
		clientOptions.DefaultCurrency = "CNY"
	}
	if clientOptions.TimeExpire < 1 {
		clientOptions.TimeExpire = 60
	}
	// TODO 参数校验
	return &clientOptions, true
}

// parsePaymentStatus 将微信支付状态转换为平台支付状态
func parsePaymentStatus(tradeState string) models.PaymentStatus {
	switch tradeState {
	case "SUCCESS":
		return models.PaymentStatusSuccess
	case "REFUND":
		// 退款
		return models.PaymentStatusSuccess
	case "NOTPAY":
		return models.PaymentStatusPending
	case "CLOSED":
		return models.PaymentStatusFailed
	case "REVOKED":
		return models.PaymentStatusFailed
	case "USERPAYING":
		return models.PaymentStatusPending
	case "PAYERROR":
		return models.PaymentStatusFailed
	default:
		return models.PaymentStatusUnknown
	}
}

func (w *WeChatPaymentClient) buildCallbackResult(obj any, responseRaw any) (*CallbackResult, error) {
	v := reflect.ValueOf(obj)
	if !v.IsValid() {
		return nil, fmt.Errorf("invalid object")
	}
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if !v.IsValid() {
		return nil, fmt.Errorf("invalid dereferenced object")
	}

	// 提取金额信息
	payerTotal, currency := extractAmount(v)

	// 解析 Attach 字段中的扩展信息
	var extra map[string]string
	if attach := getReflectString(v, "Attach"); attach != "" {
		if err := json.Unmarshal([]byte(attach), &extra); err != nil {
			return nil, fmt.Errorf("unmarshal attach failed: %w", err)
		}
	}
	// 确保 extra 不为 nil
	if extra == nil {
		extra = make(map[string]string)
	}

	// 解析支付时间
	var paidAt *time.Time
	if successTime := getReflectString(v, "SuccessTime"); successTime != "" {
		if t, err := time.Parse(time.RFC3339, successTime); err == nil {
			paidAt = &t
		}
	}

	// 构建结果对象
	res := &CallbackResult{
		Status:         parsePaymentStatus(getReflectString(v, "TradeState")),
		OutTradeNo:     getReflectString(v, "OutTradeNo"),
		OutPayRecordSN: extra[ExtraKeyPaySN],
		TradeNo:        getReflectString(v, "TransactionId"),
		PaymentChannel: models.ChannelWeChatPay,
		TradeType:      getReflectString(v, "TradeType"),
		BankType:       getReflectString(v, "BankType"),
		PaymentAmount:  payerTotal,
		Currency:       currency,
		PaidAt:         paidAt,
		Extra:          extra,
		ResponseRaw:    responseRaw,
	}
	return res, nil
}

func getReflectString(v reflect.Value, fieldName string) string {
	if !v.IsValid() {
		return ""
	}
	f := v.FieldByName(fieldName)
	if !f.IsValid() {
		return ""
	}

	if f.Kind() == reflect.Pointer {
		if f.IsNil() {
			return ""
		}
		f = f.Elem()
	}

	if f.Kind() == reflect.String {
		return f.String()
	}
	return ""
}

// extractAmount 提取金额信息
func extractAmount(v reflect.Value) (payerTotal int64, currency string) {
	amount := v.FieldByName("Amount")
	if !amount.IsValid() {
		return 0, ""
	}

	if amount.Kind() == reflect.Pointer {
		if amount.IsNil() {
			return 0, ""
		}
		amount = amount.Elem()
	}

	if !amount.IsValid() {
		return 0, ""
	}

	// 提取 PayerTotal
	if pt := amount.FieldByName("PayerTotal"); pt.IsValid() {
		switch pt.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			payerTotal = pt.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			payerTotal = int64(pt.Uint())
		case reflect.Float32, reflect.Float64:
			payerTotal = int64(pt.Float())
		}
	}

	// 提取 Currency
	if c := amount.FieldByName("Currency"); c.IsValid() {
		if c.Kind() == reflect.String {
			currency = c.String()
		} else if c.Kind() == reflect.Pointer && !c.IsNil() {
			if elem := c.Elem(); elem.Kind() == reflect.String {
				currency = elem.String()
			}
		}
	}

	return payerTotal, currency
}
