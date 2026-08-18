# 工作流服务技术文档

## 1. 项目概述

apps/workflow 是基于 [Coze Studio](https://github.com/coze-dev/coze-studio) (Apache 2.0) 二次开发的可视化工作流引擎，集成在 CoreKG 单体仓库中。提供拖拽式工作流编排、32 种节点类型、实时调试、版本发布等能力。

### 1.1 与 CoreKG 的关系

- 独立应用，有自己的 Hertz HTTP 框架（非 CoreKG 的 yg-go）
- 有自己的 go.mod 风格独立包层次
- 通过 `utils/yyguauth`、`utils/yygudb`、`utils/requestyygu` 与 CoreKG 集成
- 无 `app.go` 入口（其他 CoreKG app 有），入口为 `cmd/main.go`

### 1.2 技术栈

| 类别 | 技术选型 |
|------|----------|
| HTTP 框架 | CloudWeGo Hertz |
| 工作流引擎 | CloudWeGo Eino compose |
| ORM | GORM + GORM Gen |
| 缓存 | Redis |
| 搜索 | Elasticsearch 8 |
| 对象存储 | MinIO / S3 / TOS |
| 消息队列 | NATS / Pulsar / RocketMQ / NSQ / Kafka |
| LLM 提供商 | OpenAI / Claude / Gemini / Qwen / DeepSeek / Ollama / ARK |

## 2. 架构总览

### 2.1 DDD 分层架构

```
apps/workflow/
├── api/                    # HTTP 层 (Hertz handlers + routes)
│   ├── handler/coze/       # 26+ handler 文件
│   ├── model/              # 请求/响应 DTO
│   ├── router/coze/        # 路由注册（自动生成 + 自定义）
│   └── middleware/         # 中间件（8个）
├── application/            # 应用服务层（编排）
│   ├── workflow/           # 工作流应用服务
│   ├── singleagent/        # Agent 管理
│   ├── conversation/       # 会话管理
│   ├── knowledge/          # 知识库管理
│   ├── plugin/             # 插件管理
│   └── ...                 # 其他子域
├── domain/                 # 领域层
│   └── workflow/           # 工作流核心域
│       ├── entity/         # 领域实体 & 值对象
│       ├── service/        # 领域服务
│       ├── internal/
│       │   ├── compose/    # 核心执行引擎（Eino）
│       │   ├── nodes/      # 32种节点实现
│       │   ├── execute/    # 执行上下文 & 回调
│       │   ├── schema/     # 内部 Schema
│       │   └── repo/       # 仓储实现（GORM Gen）
│       └── component_interface.go  # 领域接口定义
├── crossdomain/            # 跨域接口（15个）
├── infra/                  # 基础设施层
│   ├── cache/              # Redis
│   ├── checkpoint/         # 检查点（内存/Redis）
│   ├── coderunner/         # 代码执行器
│   ├── document/           # 文档处理管线
│   ├── embedding/          # 向量嵌入
│   ├── es/                 # Elasticsearch
│   ├── eventbus/           # 事件总线
│   ├── idgen/              # ID 生成器
│   ├── storage/            # 对象存储
│   └── sse/                # Server-Sent Events
├── bizpkg/                 # 业务支撑包
│   ├── llm/modelbuilder/   # LLM 模型构建器（7个提供商）
│   └── config/             # 运行时配置
├── conf/                   # 配置文件
├── types/                  # 共享类型
│   ├── consts/             # 常量
│   ├── ddl/                # DDL 生成器
│   └── errno/              # 错误码
└── utils/                  # CoreKG 集成工具
    ├── yyguauth/           # YYGU 认证
    ├── yygudb/             # YYGU 数据库
    └── requestyygu/        # CoreKG HTTP 客户端
```

### 2.2 分层调用链

```
HTTP Request
    → Middleware Chain (Auth/Log/i18n/CORS)
        → Handler (api/handler/coze/)
            → Application Service (application/workflow/)
                → Domain Service (domain/workflow/service/)
                    → Compose Engine (domain/workflow/internal/compose/)
                        → Node Implementation (domain/workflow/internal/nodes/)
                    → Repository (domain/workflow/internal/repo/)
                        → GORM Gen Query Builder → MySQL
                → Cross-Domain Service (crossdomain/*)
            → SSE Response / JSON Response
```

### 2.3 应用初始化流程

三阶段初始化（`application/application.go` `Init` 函数）：

**Phase 1 — 基础服务**（仅依赖 infra）：

UploadService, OpenAuthService, PromptService, ModelMgrService, ConnectorService, UserService, TemplateService, PermissionService

**Phase 2 — 核心服务**（依赖 Phase 1）：

PluginService, MemoryService, KnowledgeService, **WorkflowService**, ShortcutCmdService

**Phase 3 — 复合服务**（依赖 Phase 1+2）：

SingleAgentService, AppService, SearchService, ConversationService

初始化后设置 15 个跨域单例（crossdomain），实现域间解耦调用。

### 2.4 中间件链（按执行顺序）

| 序号 | 中间件 | 作用 |
|------|--------|------|
| 1 | ContextCacheMW | 初始化请求级上下文缓存 |
| 2 | RequestInspectorMW | 分类请求类型（WebAPI/OpenAPI/Static） |
| 3 | SetHostMW | 存储 host/scheme 到上下文 |
| 4 | SetLogIDMW | 生成 X-Request-Id |
| 5 | CORS | 跨域允许 |
| 6 | AccessLogMW | 结构化请求/响应日志 |
| 7 | OpenapiAuthMW | OpenAPI Bearer Token 认证 |
| 8 | SessionAuthMW | WebAPI YYGU Token 认证 |
| 9 | I18nMW | 国际化语言检测 |

CoreKG 特有中间件：

- **corekgCreateResourcePermissionMw**：资源创建后注册到 CoreKG 权限系统
- **corekgLibraryResourcePermissionFilterMw**：列表响应按 CoreKG 权限过滤
- **SpaceSyncMW**：空间和成员同步
- **AdminAuthMW**：管理员 UIN 白名单

## 3. 子域概览

| 子域 | 包路径 | 职责 |
|------|--------|------|
| Workflow | `application/workflow`, `domain/workflow` | 工作流编排核心 |
| Agent | `application/singleagent`, `domain/agent` | Bot/Agent 管理 |
| Conversation | `application/conversation`, `domain/conversation` | 会话与消息 |
| Knowledge | `application/knowledge`, `domain/knowledge` | 知识库管理 |
| Plugin | `application/plugin`, `domain/plugin` | 插件/工具管理 |
| Memory | `application/memory`, `domain/memory` | 数据库记忆 & 变量 |
| User | `application/user`, `domain/user` | 用户管理 |
| Connector | `application/connector`, `domain/connector` | 部署渠道 |
| Permission | `application/permission`, `domain/permission` | 权限管理 |
| Search | `application/search`, `domain/search` | 搜索 |
| Upload | `application/upload`, `domain/upload` | 文件上传 |
| Prompt | `application/prompt`, `domain/prompt` | 提示词资源 |
| Template | `application/template`, `domain/template` | 模板管理 |
| OpenAuth | `application/openauth`, `domain/openauth` | API Key 认证 |
| DataCopy | `domain/datacopy` | 数据复制 |
