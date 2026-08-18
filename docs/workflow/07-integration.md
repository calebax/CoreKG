# 集成与扩展

## 1. CoreKG 集成

apps/workflow 作为独立应用，通过以下机制与 CoreKG 主系统深度集成。

### 1.1 认证集成

Path: `utils/yyguauth/`

- 验证 YYGU Token，对接 CoreKG 认证服务
- SessionAuthMW 中间件自动提取用户信息
- 支持 API Key（PAT）认证用于 OpenAPI

### 1.2 数据库集成

Path: `utils/yygudb/`

- 连接 CoreKG 共享 MySQL 实例
- 读取核心设置（core_settings 表）
- 初始化共享数据库连接

### 1.3 CoreKG API 客户端

Path: `utils/requestyygu/`

HTTP 客户端调用 CoreKG 后端 API：

| 功能 | 说明 |
|------|------|
| 账号信息获取 | 获取用户详细信息 |
| 权限校验 | 检查用户对资源的权限 |
| 知识库操作 | 跨应用知识库访问 |
| 空间同步 | 同步空间和成员信息 |
| Token 交换 | 内部 Token 互信 |
| 权限范围管理 | SetCoreKGResourceScope, FilterCoreKGResourceIDsByScopePermission |

### 1.4 CoreKG 权限同步

Path: `api/middleware/middleware_corekg.go`

两个 CoreKG 特有中间件：

**corekgCreateResourcePermissionMw**（后置过滤器）：
- 拦截资源创建端点的响应
- 自动将新创建的资源注册到 CoreKG 权限系统
- 支持的资源类型：Agent, Plugin, Workflow, Prompt, Database, Knowledge, Project, UI

**corekgLibraryResourcePermissionFilterMw**（后置过滤器）：
- 拦截列表响应
- 按 CoreKG 权限范围过滤结果
- 确保用户只能看到有权限的资源

### 1.5 迁移路由

Path: `api/router/coze/migration.go`

`/api/internal/space_id_migration` — 资源空间 ID 迁移接口。

## 2. 跨域服务（CrossDomain）

Path: `crossdomain/`

15 个跨域接口提供域间解耦通信：

| 包 | 接口 | 职责 |
|----|------|------|
| crosspermission | PermissionService | 权限校验 |
| crossconnector | ConnectorService | 连接器管理 |
| crossdatabase | DatabaseService | 数据库记忆 |
| crossknowledge | KnowledgeService | 知识库 |
| crossplugin | PluginService | 插件/工具 |
| crossvariables | VariablesService | 变量管理 |
| crossworkflow | WorkflowService | 工作流执行 |
| crossconversation | ConversationService | 会话管理 |
| crossmessage | MessageService | 消息管理 |
| crossagentrun | AgentRunService | Agent 运行记录 |
| crossagent | AgentService | Agent/Bot |
| crossuser | UserService | 用户 |
| crossdatacopy | DataCopyService | 数据复制 |
| crosssearch | SearchService | 搜索 |
| crossupload | UploadService | 文件上传 |
| crossapp | AppService | 应用管理 |

每个跨域包包含：
- 接口定义（`crossdomain/<name>/service.go`）
- 默认单例（`DefaultSVC()`）
- 设置函数（`SetDefaultSVC()`）
- 实现（`crossdomain/<name>/impl/`）

初始化时在 `application.Init()` 中通过 `SetDefaultSVC` 注入实际实现。

## 3. 与 Agent 系统的集成

### 3.1 工作流作为 Agent 工具

Path: `domain/workflow/component_interface.go` → `AsTool` 接口

工作流可以暴露为 Eino Tool，供 Agent 在对话中调用：

```go
type AsTool interface {
    WorkflowAsModelTool(ctx, policies) ([]ToolFromWorkflow, error)
    WithMessagePipe() (compose.Option, StreamReader, cleanup)
    WithExecuteConfig(cfg) compose.Option
    WithResumeToolWorkflow(events, data, allEvents) compose.Option
}
```

Agent 对话流程：
```
用户消息 → Agent → Eino Graph → Tool Call → WorkflowAsModelTool
  → 选择工作流 → 执行 → 返回结果 → Agent 生成回复
```

### 3.2 WorkflowReference 追踪

`workflow_reference` 表记录工作流被 Agent 引用的关系：
- `refer_type = 2`（tool）
- `referring_biz_type = 2`（agent）

### 3.3 ConnectorWorkflowVersion

`connector_workflow_version` 表记录连接器与工作流版本的绑定关系，用于：
- Agent 发布时绑定特定工作流版本
- 多渠道部署（同一 Agent 不同渠道可用不同工作流版本）

## 4. 与知识库的集成

### 4.1 KnowledgeRetriever 节点

- 从 Elasticsearch 检索文档切片
- 支持 rerank 重排序
- 依赖 `crossknowledge` 获取知识库配置

### 4.2 KnowledgeIndexer 节点

- 向知识库写入新文档
- 执行完整的文档处理管线

### 4.3 KnowledgeDeleter 节点

- 从知识库删除文档

### 4.4 工作流复制时的知识库依赖

复制工作流时，自动检测并复制引用的知识库数据集。

## 5. 与插件系统的集成

### 5.1 Plugin 节点

- 调用外部 API/工具
- 支持 OAuth 认证流
- 依赖 `crossplugin` 获取插件 API 定义

### 5.2 内置插件

16 个内置插件定义在 `conf/plugin/pluginproduct/`，涵盖飞书、高德地图、博查搜索等。

### 5.3 自定义插件

通过 `/api/plugin_api/` 注册自定义插件：
- 定义 API Schema
- 配置认证方式
- 调试测试
- 发布上线

### 5.4 工作流复制时的插件依赖

复制工作流时，自动检测并复制引用的插件。

## 6. 与数据库记忆的集成

### 6.1 Database 系列节点

5 种数据库操作节点，支持结构化数据的 CRUD 操作。

### 6.2 数据库 Schema

通过 Memory 子域管理数据库表 Schema，支持：
- 动态创建表
- Schema 验证
- NL2SQL 自然语言查询

### 6.3 工作流复制时的数据库依赖

复制工作流时，自动检测并复制引用的数据库配置。

## 7. 与变量系统的集成

### 7.1 全局变量

- `global_variable_app` — 应用级变量
- `user` — 用户级变量
- `system` — 系统级变量

### 7.2 VariableAssigner 节点

- 写入应用变量或用户变量
- 循环内专用版本：VariableAssignerWithinLoop

### 7.3 VariableAggregator 节点

- 聚合多分支输出到单一变量

## 8. 与会话系统的集成

### 8.1 ChatFlow 模式

工作流 mode=3 时为 ChatFlow，自动集成：
- 会话创建/管理
- 消息持久化
- 建议回复生成
- 中断恢复

### 8.2 会话管理节点

10 个节点提供完整的会话 CRUD 能力，可在工作流中动态操作会话。

### 8.3 会话模板

- 静态会话模板
- 动态会话模板
- 草稿/在线双版本

## 9. 与搜索系统的集成

### 9.1 资源事件

工作流的创建/更新/删除自动发布到搜索索引：

```
ApplicationService.CreateWorkflow
  → repo.CreateMeta
  → PublishWorkflowResource(Created)
      → search.ResourceEventBus.Publish
          → EventBus → 搜索服务 → ES 索引更新
```

### 9.2 项目搜索

工作流作为搜索资源参与项目级搜索。

## 10. 外部集成配置

### 10.1 LLM 模型配置

通过 `conf/model/template/` 配置 LLM 模型，支持 7 个提供商。

### 10.2 消息队列配置

通过 `WorkflowConfig.MQ` 配置事件总线后端。

### 10.3 存储配置

通过 `WorkflowConfig.Storage` 配置对象存储（MinIO/S3/TOS）。

### 10.4 Elasticsearch 配置

通过 `WorkflowConfig.Elasticsearch` 配置 ES 连接。

## 11. 错误码体系

Path: `types/errno/workflow.go`

30+ 工作流专用错误码：

| 错误码 | 说明 |
|--------|------|
| 720702011 | ErrWorkflowNotPublished |
| 720701013 | ErrWorkflowExecuteFail |
| 720702004 | ErrWorkflowNotFound |
| 720702085 | ErrWorkflowTimeout |
| 777777777 | ErrWorkflowCanceledByUser |
| 777777776 | ErrNodeTimeout |

所有错误码通过 `code.Register()` 注册人类可读消息。`errnoMap` 映射内部错误码到 OpenAPI 错误码。

## 12. 国际化（i18n）

- 节点名称和描述支持中英文
- `NodeTypeMeta` 包含 `Name` / `EnUSName` / `Desc` / `EnUSDescription`
- I18nMW 从 Session 或 Accept-Language 检测语言
- 错误消息支持多语言

## 13. 安全机制

### 13.1 认证

| 路径前缀 | 认证方式 |
|----------|----------|
| `/api/*` | YYGU Session Token |
| `/open_api/*` | Bearer Token (API Key) |
| `/v1/*` | Bearer Token (API Key) |
| `/v3/*` | Bearer Token (API Key) |
| `/api/admin/*` | Admin UIN 白名单 |
| `/api/internal/*` | 内部服务间调用 |

### 13.2 权限

- 空间级权限：用户必须属于目标空间
- 资源级权限：通过 crosspermission 校验读取/写入权限
- CoreKG 权限同步：资源创建自动注册，列表自动过滤
- API Key 权限：PAT 绑定空间和资源范围

### 13.3 SQL 安全

DatabaseCustomSQL 节点通过 SQL Parser 验证：
- 语法合法性
- 注入风险检测
- 操作权限范围

### 13.4 代码执行安全

CodeRunner 节点提供沙箱模式，隔离代码执行环境。

## 14. 上游同步

项目从 [Coze Studio](https://github.com/coze-dev/coze-studio) 上游同步。

同步说明文件：`apps/workflow/coze同步说明.md`

注意事项：
- 大部分文件保留上游 Apache 2.0 Copyright
- YYGU 自定义修改集中在：utils/、api/middleware/middleware_corekg.go、api/router/coze/custom_corekg.go
- 上游更新时需手动合并冲突
