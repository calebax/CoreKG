# Nodes

## 定位

Nodes 是一个轻量的 **Eino ChatModel 工厂库**，提供创建基础 LLM ChatModel 实例的工厂函数。作为共享组件被 kechat 等服务引用。

## 核心函数

| 函数 | 说明 |
|------|------|
| `BaseModel()` | 硬编码默认模型（deepseek-v3） |
| `GetBaseModelFromSetting()` | 从 core.settings 配置动态创建模型实例 |

## 代码结构

```
apps/nodes/
└── chatmodel/
    ├── std.go            # BaseModel + GetBaseModelFromSetting
    └── yg_agent.go       # 预留（空文件）
```

## 与其他服务的关系

- 被 `apps/kechat` 等服务引用 - 统一创建 AI 模型客户端
- 依赖 `cloudwego/eino-ext/components/model/openai` - Eino OpenAI ChatModel
- 依赖 `yg-go/settings` - 配置读取
