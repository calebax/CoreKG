package accountmds

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/encryptor"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

// decryptString 解密包装函数
func decryptString(ctx context.Context, s string) (string, error) {
	// Base64-decode the string to get the raw ciphertext bytes
	cipherText, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		logs.ErrorContextf(ctx, "Failed to decode Base64 string: %v", err)
		return "", err
	}

	secret, err := settings.GetSecret("account", "aeskey")
	if err != nil {
		logs.ErrorContextf(ctx, "Failed to get secret: %v", err)
		return "", err
	}

	// Pass the raw ciphertext bytes to the decryption function
	decrypt, err := encryptor.AesDecrypt([]byte(secret), cipherText)
	if err != nil {
		logs.ErrorContextf(ctx, "Failed to decrypt data: %v", err)
		return "", err
	}
	return string(decrypt), nil
}

const DecryptPrefix = "keep-enc-"

// findAndDecryptByPath 根据路径导航并解密字段
func findAndDecryptByPath(ctx context.Context, data map[string]interface{}, path string, decryptFunc func(context.Context, string) (string, error)) bool {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return false
	}

	// 遍历数据，逐级向下
	var currentData interface{} = data
	for i, part := range parts {
		if m, ok := currentData.(map[string]interface{}); ok {
			if i == len(parts)-1 {
				// 抵达路径的最后一级，进行解密
				if value, exists := m[part]; exists {
					if strValue, isString := value.(string); isString {
						if !strings.HasPrefix(strValue, DecryptPrefix) {
							logs.DebugContextf(ctx, "str value[%v] has no prefix(%v), ignored it", strValue, DecryptPrefix)
							continue
						}
						//cut prefix
						str := strings.TrimPrefix(strValue, DecryptPrefix)
						logs.DebugContextf(ctx, "prefix trimed str[%v] value[%v]", strValue, str)
						if decryptedValue, err := decryptFunc(ctx, str); err == nil {
							m[part] = decryptedValue
							return true
						} else {
							logs.ErrorContextf(ctx, "解密路径 '%s' 失败: %v\n", path, err)
							// 记录错误但不中断
						}
					}
				}
				return false
			}
			// 导航到下一级
			currentData = m[part]
		} else {
			// 路径不存在或类型不匹配
			return false
		}
	}
	return false
}

// DecryptMD 解密由完整 JSON 路径指定的字段
func DecryptMD(paths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil {
			logs.ErrorContextf(c, "request body is nil")
			runtime.InternalError(c, i18n.T(runtime.GetLanguage(c), "account_request_body_empty")) // 请求体为空
			c.Next()
			return
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logs.ErrorContextf(c, "Failed to read request body: %v", err)
			runtime.InternalError(c, i18n.T(runtime.GetLanguage(c), "account_read_request_body_failed")) // 获取请求体失败
			return
		}

		var data map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			logs.ErrorContextf(c, "failed to unmarshal request body: %v", err)
			//runtime.InternalError(c, "解析请求体失败")
			runtime.InternalError(c, i18n.T(runtime.GetLanguage(c), "account_parse_request_body_failed")) // 解析请求体失败
			return
		}

		changed := false
		for _, path := range paths {
			if findAndDecryptByPath(c, data, path, decryptString) {
				changed = true
			}
		}

		if changed {
			// 如果有字段被解密，重新编码 JSON 并替换请求体
			modifiedJSON, err := json.Marshal(data)
			if err != nil {
				logs.ErrorContextf(c, "failed to marshal request body: %v", err)
				runtime.InternalError(c, i18n.T(runtime.GetLanguage(c), "account_rewrite_request_body_failed")) // 回写请求体失败
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(modifiedJSON))
		} else {
			// 恢复原始请求体，以便后续处理器能够读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()
	}
}

const (
	KeyRefreshToken = "refresh_token"
	KeyUserID       = "user_id"
	KeyRequest      = "request"
)

// RequireRefreshToken 需要refresh token
func RequireRefreshToken(c *gin.Context) {
	if c.Request.Body == nil {
		logs.ErrorContextf(c, "request body is nil")
		runtime.InternalError(c, i18n.T(runtime.GetLanguage(c), "account_request_body_empty")) // 请求体为空
		c.Next()
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logs.ErrorContextf(c, "Failed to read request body: %v", err)
		runtime.InternalError(c, i18n.T(runtime.GetLanguage(c), "account_read_request_body_failed")) // 获取请求体失败
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		logs.ErrorContextf(c, "failed to unmarshal request body: %v", err)
		runtime.InternalError(c, i18n.T(runtime.GetLanguage(c), "account_parse_request_body_failed")) // 解析请求体失败
		return
	}

	var (
		userID       uint
		refreshToken string
	)

	userIDTmp, exists := data[KeyRequest].(map[string]interface{})[KeyUserID].(float64)
	if !exists {
		logs.ErrorContextf(c, "user_id not found %+v", data)
		runtime.BadRequest(c, i18n.T(runtime.GetLanguage(c), "kecore_user_id_empty"))
		return
	}
	userID = uint(userIDTmp)

	refreshToken, exists = data[KeyRequest].(map[string]interface{})[KeyRefreshToken].(string)
	if !exists {
		logs.ErrorContextf(c, "refresh_token not found %+v", data)
		runtime.BadRequest(c, i18n.T(runtime.GetLanguage(c), "account_refresh_token_invalid"))
		return
	}

	if len(refreshToken) != user.RefreshTokenLen {
		logs.ErrorContextf(c, "refresh_token invalid len: %v,acctually[%v]", len(refreshToken), user.RefreshTokenLen)
		runtime.BadRequest(c, i18n.T(runtime.GetLanguage(c), "account_refresh_token_invalid"))
		return
	}
	if userID <= 0 {
		logs.ErrorContextf(c, "user_id is lower than zero userID:%v", userID)
		runtime.BadRequest(c, i18n.T(runtime.GetLanguage(c), "kecore_user_id_empty"))
		return
	}

	// 获取用户refresh_token
	refToken, err := user.GetRefreshToken(userID)
	if err != nil {
		logs.ErrorContextf(c, "ChooseUin: get refresh token failed, %s", err)
		runtime.InternalError(c, i18n.T(runtime.GetLanguage(c), "account_refresh_token_fetch_failed"))
		return
	}
	if refToken != refreshToken {
		logs.ErrorContextf(c, "ChooseUin: refresh token not match, %s", err)
		runtime.InternalError(c, i18n.T(runtime.GetLanguage(c), "account_refresh_token_mismatch"))
		return
	}

	// 恢复原始请求体，以便后续处理器能够读取
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	c.Next()
}
