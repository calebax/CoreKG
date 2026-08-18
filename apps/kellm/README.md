# KELLM

## 定位

KELLM 是 CoreKG 平台的 **LLM 代理服务（LLM Proxy/Gateway）库**。提供 OpenAI 兼容的 `/chat/completions` 接口，根据模型名路由到不同的上游 LLM 服务。支持流式响应（SSE）和多驱动架构。

## 核心能力

- OpenAI 兼容的 ChatCompletions 接口
- 多驱动架构（通过 RegisterDriver/GetDriver 注册不同模型类型的 Driver）
- 当前实现了 OpenAI 驱动
- 完整的 OpenAI 错误格式兼容
- 流式响应逐行 flush
- 中间件链处理（middleware.Chain）

## 代码架构

```
apps/kellm/
├── chat_api.go              # HTTP handler: ChatCompletions()
├── drivers/
│   ├── driver.go            # Driver 接口 + 注册/获取
│   ├── errors.go            # 驱动错误定义
│   └── openai/              # OpenAI 驱动实现
├── models/
│   ├── kellmtype/           # 模型配置、请求/响应类型
│   └── functioncall/        # Function Call 类型
└── services/
    └── svckellm/
        ├── service.go       # ProxyChatCompletions 核心逻辑
        ├── result.go        # ProxyResult / OpenAIErrorResult
        ├── errors.go        # 业务错误
        └── middleware/      # 中间件链
```

## 与其他服务的关系

- 被 `apps/kechat` 引用 - LLM 透传代理
- 被 `apps/keapi` 引用 - LLM 消息格式
- 依赖 `yg-go/settings` - 从配置中心获取模型配置
