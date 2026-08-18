package requestyygu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/insmtx/corekg/apps/workflow/utils/mysql/coresettings"
)

const (
	tokenExchangeTTLSeconds    = int64(900) // 15 minutes
	tokenExchangeRefreshBuffer = 60 * time.Second
	tokenExchangeHTTPTimeout   = 8 * time.Second
	defaultTokenExchangePath   = "/v2/account.GetOBOToken"
	oboAudience                = "kg_open_coze"
	oboGrantType               = "token_exchange"
)

type exchangedTokenCache struct {
	token    string
	expireAt time.Time
}

var apiTokenCache sync.Map // key: userID(string) -> exchangedTokenCache

func getOrExchangeCoreKGToken(ctx context.Context, userID int64) (string, error) {
	if userID <= 0 {
		return "", errors.New("invalid user id for corekg token exchange")
	}

	key := strconv.FormatInt(userID, 10)
	if cacheV, ok := apiTokenCache.Load(key); ok {
		if cacheItem, ok := cacheV.(exchangedTokenCache); ok &&
			cacheItem.token != "" &&
			time.Until(cacheItem.expireAt) > tokenExchangeRefreshBuffer {
			return cacheItem.token, nil
		}
	}

	token, expireAt, err := exchangeCoreKGToken(ctx, userID)
	if err != nil {
		return "", err
	}

	apiTokenCache.Store(key, exchangedTokenCache{
		token:    token,
		expireAt: expireAt,
	})
	return token, nil
}

func exchangeCoreKGToken(ctx context.Context, userID int64) (string, time.Time, error) {
	baseURL, err := coresettings.GetCoreKGUrl()
	if err != nil {
		return "", time.Time{}, err
	}

	apiURL, err := url.JoinPath(baseURL, defaultTokenExchangePath)
	if err != nil {
		return "", time.Time{}, err
	}

	payload := map[string]any{
		"request": map[string]any{
			"uin":        userID,
			"audience":   oboAudience,
			"grant_type": oboGrantType,
			"scope":      "",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", time.Time{}, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return "", time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: tokenExchangeHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("corekg token exchange failed, status=%d, body=%s", resp.StatusCode, string(respBytes))
	}

	token, expireAt, err := parseExchangeResponse(respBytes)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expireAt, nil
}

func parseExchangeResponse(raw []byte) (string, time.Time, error) {
	type baseResp struct {
		Code     int             `json:"code"`
		Message  string          `json:"message"`
		Response json.RawMessage `json:"Response"`
		Data     json.RawMessage `json:"data"`
	}

	var res baseResp
	if err := json.Unmarshal(raw, &res); err != nil {
		return "", time.Time{}, err
	}
	if res.Code != 0 {
		return "", time.Time{}, fmt.Errorf("corekg token exchange business error: code=%d message=%s", res.Code, res.Message)
	}

	payload := res.Response
	if len(payload) == 0 {
		payload = res.Data
	}
	if len(payload) == 0 {
		return "", time.Time{}, errors.New("corekg token exchange response payload is empty")
	}

	var tokenMap map[string]any
	if err := json.Unmarshal(payload, &tokenMap); err != nil {
		return "", time.Time{}, err
	}

	token := firstNonEmptyString(
		anyToString(tokenMap["jwt_token"]),
		anyToString(tokenMap["access_token"]),
		anyToString(tokenMap["token"]),
		anyToString(tokenMap["corekg_token"]),
	)
	if token == "" {
		return "", time.Time{}, fmt.Errorf("corekg token exchange payload missing token: %s", string(payload))
	}

	expireAt := time.Now().Add(time.Duration(tokenExchangeTTLSeconds) * time.Second)
	if ts := anyToInt64(tokenMap["expired_at"]); ts > 0 {
		expireAt = time.Unix(ts, 0)
	} else if sec := anyToInt64(tokenMap["expires_in"]); sec > 0 {
		expireAt = time.Now().Add(time.Duration(sec) * time.Second)
	} else if sec := anyToInt64(tokenMap["expire_in"]); sec > 0 {
		expireAt = time.Now().Add(time.Duration(sec) * time.Second)
	} else if ts := anyToInt64(tokenMap["expire_at"]); ts > 0 {
		expireAt = time.Unix(ts, 0)
	} else if ts := anyToInt64(tokenMap["expires_at"]); ts > 0 {
		expireAt = time.Unix(ts, 0)
	}

	return sanitizeToken(token), expireAt, nil
}

func anyToInt64(v any) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	case float32:
		return int64(val)
	case json.Number:
		n, _ := val.Int64()
		return n
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(val), 10, 64)
		return n
	default:
		return 0
	}
}

func anyToString(v any) string {
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val)
	case json.Number:
		return val.String()
	default:
		return ""
	}
}

func firstNonEmptyString(items ...string) string {
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			return item
		}
	}
	return ""
}
