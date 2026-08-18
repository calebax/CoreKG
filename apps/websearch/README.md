# WebSearch API

独立部署的原子 Web Search 服务。它提供 Provider 路由、Profile Pool、缓存和 opaque cursor，不读取搜索结果正文，也不依赖 Read API。

## API Market 测试接口

```http
POST https://tapi.insmtx.com/v6/se/general/search
Authorization: Bearer <API_KEY>
Content-Type: application/json
```

```json
{"request":{"query":"golang","limit":10,"timeout":"20s","routing":{"providers":["brave","duckduckgo"]},"filters":{"region":"CN","include_domains":["go.dev"],"exclude_domains":["example.com"]},"query_options":{"exact_phrases":["context package"],"title_terms":["documentation"],"file_types":["html"]}}}
```

业务响应位于 API Market 返回的 `response` 字段。完整契约见 [OpenAPI](openapi/openapi.yaml)。公开接口只支持单个 query；`limit` 默认 10，范围 1–20。下一页使用 `response.page.next_cursor`，客户端不解析 cursor。

API Market Executor 调用 roc 内部路由 `POST /v3/websearch.Search`，使用专用固定 Token 的 `Authorization: Bearer <token>` 鉴权。该路由接收不带 `request` 外层的原始业务 JSON，并保持迁移前 Gin handler 的响应与错误格式。

## 运行

```bash
make run APP=websearch
```

默认监听 `:8080`。生产多副本必须在所有 Pod 挂载的配置文件中使用同一个 `api.cursor_key`；该值应由同一个 Secret 注入，确保任意 Pod 都能解析其他 Pod 生成的 cursor。

服务端鉴权 Key 默认从配置文件读取：

```yaml
auth:
  api_key: yg-se-replace-me
```

也可以在启动二进制时覆盖，命令行优先级高于配置文件：

```bash
make local APP=websearch
./bundles/websearch -c ./apps/websearch/conf/test/config.yaml --api-key yg-se-rotated-key
```

最终优先级为：配置文件 < `--api-key` 命令行参数。Key 必须使用 `yg-` 前缀，避免被 roc 全局登录中间件当成 JWT 解析并写入错误日志。

## Provider 配置

```yaml
api:
  enabled_providers: [baidu, bing, brave, duckduckgo]
  allow_request_providers: true
  provider_visibility: public
```

`routing.providers` 是有序、无重复的已启用 Provider 子集。服务严格按请求顺序降级；若请求未传则使用 YAML 中 `enabled_providers` 的默认顺序。`filters.include_domains` / `exclude_domains` 对返回 URL 做强制后置校验，并匹配目标域名及其子域名。`query_options` 使用严格能力策略：当前只有 Brave、DuckDuckGo 有官方明确的高级操作符契约，不兼容 Provider 会在执行前从候选链移除；若没有兼容项则返回参数错误。

搜索结果新增 `canonical_url` 和 `domain`。服务会移除 fragment、默认端口和常见跟踪参数，并按 canonical URL 去重。`url` 仍保留 Provider 返回的原始导航地址；`id` 仍基于该原始 URL 字符串生成。

第一页成功后，`cur_v2` cursor 会绑定规范化 query、region、filters、query options、limit 和兼容 Provider 顺序，并固定实际 Provider。需要上游 continuation state 的 Provider 会把 opaque state 一并加密进 cursor；后续页不跨 Provider 降级。

## 主要配置

```yaml
auth:
  api_key: yg-se-replace-me
server:
  default_timeout: 20s
  max_timeout: 60s
api:
  cursor_ttl: 15m
  cache_bypass: false
observability:
  diagnostics_enabled: false
  store_query: false
```

日志输出由 [`conf/test/config.yaml`](conf/test/config.yaml) 的 roc `logger` 配置控制。默认不记录完整 query。SQLite、spool 和共享 Profile 存储已移除；每个 Pod 使用自己的临时 Profile 和内存缓存。

配置优先级：代码默认值 < [`conf/test/config.yaml`](conf/test/config.yaml)。YAML 使用严格字段检查，未知字段或错误类型会阻止启动。请求中的 `timeout` 只接受整数 `ms`/`s`，范围 100ms–配置上限（上限不超过 60s），并覆盖 YAML 默认值。

## 文档与示例

- [OpenAPI](openapi/openapi.yaml)
- [curl](examples/curl/websearch.sh)
- [Python](examples/python/websearch.py)
- [Go](examples/go/main.go)

运行示例前：

```bash
export API_KEY='<测试环境 API Key>'
```

## 验证

```bash
go test ./apps/websearch/...
go test -race ./apps/websearch/...
go test -run '^$' -bench 'Benchmark(Parse|SearchHandlerLatency)$' ./apps/websearch/models/provider/duckduckgo ./apps/websearch/internal/apis
make local APP=websearch
```

Parser benchmark 覆盖空结果、20 条结果和 1000 条结果页面；Handler benchmark 通过真实 service/cache/provider 编排链路，以 1、8、32 并发输出 `min/avg/max`、`p50/p95/p99` 和 QPS。

逐 Provider 实测见 [provider smoke test](docs/provider-smoke-test.md)。
