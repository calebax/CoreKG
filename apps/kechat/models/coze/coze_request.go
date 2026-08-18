package coze

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

const defaultCozeUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"

type cozeRequestOptions struct {
	method        string
	token         string
	baseURL       string
	headers       map[string]string
	query         map[string]string
	skipCodeCheck bool
	dataOnly      bool
}

// CozeRequestOption 用于定制 CozeRequest 的行为
type CozeRequestOption func(*cozeRequestOptions)

// WithCozeMethod 自定义请求方法，默认 POST
func WithCozeMethod(method string) CozeRequestOption {
	return func(opt *cozeRequestOptions) {
		opt.method = method
	}
}

// WithCozeToken 显式指定 token，不传则使用当前登录态
func WithCozeToken(token string) CozeRequestOption {
	return func(opt *cozeRequestOptions) {
		opt.token = token
	}
}

// WithCozeBaseURL 显式指定 coze 基础地址，不传则读取配置
func WithCozeBaseURL(baseURL string) CozeRequestOption {
	return func(opt *cozeRequestOptions) {
		opt.baseURL = baseURL
	}
}

// WithCozeHeaders 追加自定义 header，后设置的值会覆盖默认值
func WithCozeHeaders(headers map[string]string) CozeRequestOption {
	return func(opt *cozeRequestOptions) {
		if opt.headers == nil {
			opt.headers = make(map[string]string)
		}
		for k, v := range headers {
			opt.headers[k] = v
		}
	}
}

// WithCozeQuery 为请求追加 query 参数
func WithCozeQuery(query map[string]string) CozeRequestOption {
	return func(opt *cozeRequestOptions) {
		if opt.query == nil {
			opt.query = make(map[string]string)
		}
		for k, v := range query {
			opt.query[k] = v
		}
	}
}

// WithCozeDataOnly 仅将响应的 data 字段解包到目标结构体
func WithCozeDataOnly() CozeRequestOption {
	return func(opt *cozeRequestOptions) {
		opt.dataOnly = true
	}
}

// SkipCozeCodeCheck 跳过对 code 字段的校验
func SkipCozeCodeCheck() CozeRequestOption {
	return func(opt *cozeRequestOptions) {
		opt.skipCodeCheck = true
	}
}

// CozeRequest 统一的 coze 请求封装
func CozeRequest(ctx *gin.Context, path string, payload interface{}, target interface{}, opts ...CozeRequestOption) error {
	options := cozeRequestOptions{
		method:  http.MethodPost,
		headers: map[string]string{},
	}

	for _, opt := range opts {
		opt(&options)
	}

	baseURL := options.baseURL
	if baseURL == "" {
		var err error
		baseURL, err = settings.GetText("corekg", "coze_url")
		if err != nil {
			logs.ErrorContextf(ctx, "get coze url error, %s", err.Error())
			return err
		}
	}

	requestURL, err := url.JoinPath(baseURL, path)
	if err != nil {
		logs.ErrorContextf(ctx, "join coze url err %s", err.Error())
		return err
	}

	body, err := buildCozeRequestBody(options.method, payload)
	if err != nil {
		logs.ErrorContextf(ctx, "marshal request body error, %s", err.Error())
		return err
	}

	req, err := http.NewRequest(options.method, requestURL, body)
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %s", err.Error())
		return err
	}

	if len(options.query) > 0 {
		query := req.URL.Query()
		for k, v := range options.query {
			query.Set(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", defaultCozeUserAgent)

	token := options.token
	if token == "" {
		token = runtime.LoginStatus(ctx).Token
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for k, v := range options.headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "error reading body: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		logs.ErrorContextf(ctx, "coze request status %d body: %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("coze request failed with status: %d", resp.StatusCode)
	}

	var baseResp CozeResponse
	if err := json.Unmarshal(respBody, &baseResp); err != nil {
		logs.ErrorContextf(ctx, "unmarshal coze base response error: %v, body: %s", err, string(respBody))
		return err
	}

	if !options.skipCodeCheck && baseResp.Code != 0 {
		logs.ErrorContextf(ctx, "coze request code %d, msg: %s", baseResp.Code, baseResp.Msg)
		return fmt.Errorf("coze request code: %d, msg: %s", baseResp.Code, baseResp.Msg)
	}

	if target == nil {
		return nil
	}

	if options.dataOnly && baseResp.Data != nil {
		dataBytes, err := json.Marshal(baseResp.Data)
		if err != nil {
			logs.ErrorContextf(ctx, "marshal coze response data error: %v", err)
			return err
		}
		if err := json.Unmarshal(dataBytes, target); err != nil {
			logs.ErrorContextf(ctx, "unmarshal coze response data error: %v", err)
			return err
		}
		return nil
	}

	if err := json.Unmarshal(respBody, target); err != nil {
		logs.ErrorContextf(ctx, "unmarshal coze response error: %v, body: %s", err, string(respBody))
		return err
	}
	return nil
}

func buildCozeRequestBody(method string, payload interface{}) (io.Reader, error) {
	if payload == nil || strings.EqualFold(method, http.MethodGet) {
		return nil, nil
	}

	switch v := payload.(type) {
	case io.Reader:
		return v, nil
	case []byte:
		return bytes.NewReader(v), nil
	case string:
		return strings.NewReader(v), nil
	default:
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(data), nil
	}
}
