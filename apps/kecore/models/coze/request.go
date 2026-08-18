package coze

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

// CozeRequest 发起请求到 coze 服务，默认使用登录态 token
func CozeRequest(ctx *gin.Context, path string, request map[string]interface{}, targetStruct interface{}) error {
	return cozeRequest(ctx, path, request, targetStruct, "")
}

// CozeRequestWithToken 发起请求到 coze 服务，显式指定 token
func CozeRequestWithToken(ctx *gin.Context, path string, request map[string]interface{}, targetStruct interface{}, token string) error {
	return cozeRequest(ctx, path, request, targetStruct, token)
}

func cozeRequest(ctx *gin.Context, path string, request map[string]interface{}, targetStruct interface{}, token string) error {
	baseURL, err := settings.GetText("corekg", "coze_url")
	if err != nil {
		logs.ErrorContextf(ctx, "get coze url error, %s", err.Error())
		return err
	}

	apiURL, err := url.JoinPath(baseURL, path)
	if err != nil {
		logs.ErrorContextf(ctx, "join coze url err %s", err.Error())
		return err
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		logs.ErrorContextf(ctx, "marshal request body error, %s", err.Error())
		return err
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %s", err.Error())
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	tokenToUse := token
	if tokenToUse == "" {
		tokenToUse = runtime.LoginStatus(ctx).Token
	}
	req.Header.Set("Authorization", "Bearer "+tokenToUse)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("coze request failed: %v, error: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "error reading body: %v", err)
		return err
	}

	var res CozeResponse
	if err := json.Unmarshal(body, &res); err != nil {
		logs.ErrorContextf(ctx, "unmarshal coze response error: %v", err)
		return err
	}
	if res.Code != 0 {
		logs.ErrorContextf(ctx, "request coze error code: %d, message: %s", res.Code, res.Msg)
		return fmt.Errorf("request coze error code: %d, message: %s", res.Code, res.Msg)
	}

	if targetStruct == nil || res.Data == nil {
		return nil
	}

	responseBytes, err := json.Marshal(res.Data)
	if err != nil {
		logs.ErrorContextf(ctx, "marshal coze response error: %v", err)
		return err
	}
	if err := json.Unmarshal(responseBytes, targetStruct); err != nil {
		logs.ErrorContextf(ctx, "unmarshal coze response error: %v", err)
		return err
	}
	return nil
}

type CozeResponse struct {
	Code       int         `json:"code"`
	Msg        string      `json:"msg"`
	Data       interface{} `json:"data"`
	SessionKey string      `json:"session_key"`
}
