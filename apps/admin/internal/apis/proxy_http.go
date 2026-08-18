package apis

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

const (
	settingGroupAdmin        = "admin"
	settingKeyProxyWhitelist = "proxy_whitelist"
	settingKeyProxyAuth      = "proxy_auth"
	maxProxyBodyBytes        = 32 << 20 // 32MiB
)

// ProxyHTTP 纯 HTTP 透传：依赖 settings admin/proxy_whitelist 与 admin/proxy_auth。
// 请求体须为 JSON 对象，最外层含 proxy_base、proxy_path；二者会从转发给下游的 body 中剔除，其余字段原样 JSON 转发。
func ProxyHTTP(ctx *gin.Context) {
	r := ctx.Request
	w := ctx.Writer
	reqCtx := r.Context()

	var wl proxyWhitelist
	if err := settings.GetYaml(settingGroupAdmin, settingKeyProxyWhitelist, &wl); err != nil {
		logs.ErrorContextf(ctx, "[ProxyHTTP] load proxy_whitelist: %v", err)
		http.Error(w, "proxy whitelist not configured", http.StatusServiceUnavailable)
		return
	}
	if len(wl.BaseURLs) == 0 || len(wl.Routers) == 0 {
		http.Error(w, "proxy whitelist empty", http.StatusServiceUnavailable)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxProxyBodyBytes+1))
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}
	if len(body) > maxProxyBodyBytes {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(body, &envelope); err != nil {
		http.Error(w, "body must be JSON object with proxy_base and proxy_path", http.StatusBadRequest)
		return
	}

	baseRaw, okBase := envelope["proxy_base"]
	pathRaw, okPath := envelope["proxy_path"]
	if !okBase || !okPath {
		http.Error(w, "missing proxy_base or proxy_path", http.StatusBadRequest)
		return
	}
	var baseStr, pathStr string
	if err := json.Unmarshal(baseRaw, &baseStr); err != nil || strings.TrimSpace(baseStr) == "" {
		http.Error(w, "invalid proxy_base", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(pathRaw, &pathStr); err != nil || strings.TrimSpace(pathStr) == "" {
		http.Error(w, "invalid proxy_path", http.StatusBadRequest)
		return
	}

	delete(envelope, "proxy_base")
	delete(envelope, "proxy_path")

	normalizedBase, err := normalizeProxyBaseURL(baseStr)
	if err != nil {
		http.Error(w, "invalid proxy_base", http.StatusBadRequest)
		return
	}
	if !wl.baseAllowed(normalizedBase) {
		http.Error(w, "base not allowed", http.StatusForbidden)
		return
	}
	if !wl.pathAllowed(pathStr) {
		http.Error(w, "path not allowed", http.StatusForbidden)
		return
	}

	var downBody io.Reader
	var downLen int64
	if len(envelope) == 0 {
		downBody = http.NoBody
		downLen = 0
	} else {
		rest, err := json.Marshal(envelope)
		if err != nil {
			http.Error(w, "marshal downstream body failed", http.StatusBadRequest)
			return
		}
		downBody = bytes.NewReader(rest)
		downLen = int64(len(rest))
	}

	dest := joinProxyBaseAndPath(normalizedBase, pathStr)
	outReq, err := http.NewRequestWithContext(reqCtx, r.Method, dest, downBody)
	if err != nil {
		http.Error(w, "bad target url", http.StatusBadRequest)
		return
	}
	outReq.ContentLength = downLen

	for k, vals := range r.Header {
		if strings.EqualFold(k, "Host") {
			continue
		}
		if strings.EqualFold(k, "Content-Length") || strings.EqualFold(k, "Transfer-Encoding") {
			continue
		}
		for _, v := range vals {
			outReq.Header.Add(k, v)
		}
	}
	if downLen > 0 {
		outReq.Header.Set("Content-Length", strconv.FormatInt(downLen, 10))
		if outReq.Header.Get("Content-Type") == "" {
			outReq.Header.Set("Content-Type", "application/json")
		}
	}

	applyProxyDownstreamAuth(outReq, normalizedBase)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(outReq)
	if err != nil {
		logs.ErrorContextf(ctx, "[ProxyHTTP] downstream: %v", err)
		http.Error(w, "downstream request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vals := range resp.Header {
		if skipProxyResponseHeader(k) {
			continue
		}
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		logs.ErrorContextf(ctx, "[ProxyHTTP] copy body: %v", err)
	}
}

// 不把下游的 CORS / hop-by-hop 头回写给浏览器；CORS 仅由本服务中间件决定，避免与 dotpen 等响应头冲突。
func skipProxyResponseHeader(key string) bool {
	lk := strings.ToLower(key)
	switch lk {
	case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization",
		"te", "trailer", "transfer-encoding", "upgrade":
		return true
	}
	return strings.HasPrefix(lk, "access-control-")
}

func applyProxyDownstreamAuth(req *http.Request, normalizedBase string) {
	req.Header.Del("Authorization")

	var authMap map[string]string
	if err := settings.GetYaml(settingGroupAdmin, settingKeyProxyAuth, &authMap); err != nil || len(authMap) == 0 {
		return
	}
	for k, v := range authMap {
		nb, err := normalizeProxyBaseURL(k)
		if err != nil {
			continue
		}
		if nb == normalizedBase && strings.TrimSpace(v) != "" {
			req.Header.Set("Authorization", strings.TrimSpace(v))
			return
		}
	}
}
