# WebFetch API

独立部署的单 URL 正文读取服务。它对公网 HTTP/HTTPS URL 执行 SSRF 校验，优先 HTTP 获取，必要时使用 Chromium 渲染，然后提取 HTML 或纯文本正文。

## API Market 测试接口

```http
POST https://tapi.insmtx.com/v6/se/general/fetch
Authorization: Bearer <API_KEY>
Content-Type: application/json
```

```json
{"request":{"url":"https://example.com/article","timeout":"20s","output":{"format":"markdown","max_chars":30000}}}
```

业务响应位于 API Market 返回的 `response` 字段。完整契约见 [OpenAPI](openapi/openapi.yaml)。第一版只支持单个 URL、`text/html` 和 `text/plain`，不支持 PDF、Office、图片、OCR、批量读取、登录、付费墙或验证码求解。

API Market Executor 调用 roc 内部路由 `POST /v3/webfetch.Fetch`，使用专用固定 Token 的 `Authorization: Bearer <token>` 鉴权。该路由接收不带 `request` 外层的原始业务 JSON，并保持迁移前 Gin handler 的响应与错误格式。

## 运行

```bash
make run APP=webfetch
```

默认监听 `:8080`。

服务端鉴权 Key 默认从配置文件读取：

```yaml
auth:
  api_key: yg-se-replace-me
```

也可以在启动二进制时覆盖，命令行优先级高于配置文件：

```bash
make local APP=webfetch
./bundles/webfetch -c ./apps/webfetch/conf/test/config.yaml --api-key yg-se-rotated-key
```

最终优先级为：配置文件 < `--api-key` 命令行参数。Key 必须使用 `yg-` 前缀，避免被 roc 全局登录中间件当成 JWT 解析并写入错误日志。

## 主要配置

```yaml
auth:
  api_key: yg-se-replace-me
server:
  default_timeout: 20s
  max_timeout: 60s
http:
  timeout: 6s
  max_body_bytes: 5242880
  robots_policy: ignore
browser:
  enabled: true
  timeout: 12s
  slots: 4
cache:
  bypass: false
observability:
  diagnostics_enabled: false
  store_url_query: false
```

`http.robots_policy: respect` 是后续扩展点，当前配置该值会拒绝启动，避免产生已经执行 robots 检查的错误预期。日志默认移除 URL query、fragment 和凭据；只有配置 `observability.store_url_query: true` 才保留 query，外部请求无法覆盖。

域名特化由两层接口预留：`SiteStrategyResolver` 按最长域名后缀及路径前缀选择策略，`SiteStrategy` 负责准备具体读取行为。当前仅启用不改变流程的 `GenericStrategy`；验证码只返回 `captcha_required`，不尝试求解。

日志输出由 [`conf/test/config.yaml`](conf/test/config.yaml) 的 roc `logger` 配置控制。配置优先级：代码默认值 < 该 YAML。YAML 未知字段或错误类型会阻止启动。请求 `timeout` 只接受整数 `ms`/`s`，范围 100ms–配置上限（上限不超过 60s）。`max_chars` 与响应 `content_length` 都以 Unicode code point 计数；成功响应会返回上游 `content_type`、`status_code` 和固定 `usage.units=1`。

## 文档与示例

- [OpenAPI](openapi/openapi.yaml)
- [curl](examples/curl/webfetch.sh)
- [Python](examples/python/webfetch.py)
- [Go](examples/go/main.go)

运行示例前：

```bash
export API_KEY='<测试环境 API Key>'
```

## 验证

```bash
go test ./apps/webfetch/...
go test -race ./apps/webfetch/...
go test -run '^$' -bench 'Benchmark(HTMLExtractor|FetchHandlerLatency)$' ./apps/webfetch/models/fetch/extractor ./apps/webfetch/internal/apis
make local APP=webfetch
```

Extractor benchmark 覆盖 4 KiB、64 KiB、1 MiB 页面；Handler benchmark 通过真实 service/cache/detect/extract/convert 编排链路，以 1、8、32 并发输出 `min/avg/max`、`p50/p95/p99` 和 QPS。
