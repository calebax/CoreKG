# 数据模型与表结构

## 1. 工作流核心表（15张）

### 1.1 workflow_meta - 工作流元数据

Path: `domain/workflow/internal/repo/dal/model/workflow_meta.gen.go`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 工作流 ID（非自增，ID 生成器分配） |
| name | string | 工作流名称 |
| description | string | 描述 |
| icon_uri | string | 图标 URI |
| status | int32 | 0: 未发布, 1: 已发布 |
| content_type | int32 | 0: 用户, 1: 官方 |
| mode | int32 | 0: workflow, 3: chat_flow |
| created_at | int64 | 创建时间（毫秒时间戳） |
| updated_at | int64 | 更新时间 |
| deleted_at | gorm.DeletedAt | 软删除 |
| creator_id | int64 | 创建者用户 ID |
| tag | int32 | 模板标签（1=All, 2=Hot ... 18=Finance, 100=Hidden） |
| author_id | int64 | 原始作者 ID |
| space_id | int64 | 空间 ID |
| updater_id | int64 | 更新者 ID |
| source_id | int64 | 源工作流 ID |
| app_id | int64 | 应用 ID |
| latest_version | string | 最新版本号 |
| latest_version_ts | int64 | 最新版本时间 |

### 1.2 workflow_draft - 工作流草稿

Path: `domain/workflow/internal/repo/dal/model/workflow_draft.gen.go`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 工作流 ID（与 workflow_meta.id 相同，1:1） |
| canvas | string | 前端画布 Schema（JSON） |
| input_params | string | 输入参数 Schema（JSON） |
| output_params | string | 输出参数 Schema（JSON） |
| test_run_success | bool | 测试运行是否成功 |
| modified | bool | 是否有未发布修改 |
| updated_at | int64 | 更新时间 |
| deleted_at | gorm.DeletedAt | 软删除 |
| commit_id | string | 草稿快照唯一标识 |

### 1.3 workflow_version - 工作流版本

Path: `domain/workflow/internal/repo/dal/model/workflow_version.gen.go`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 自增主键 |
| workflow_id | int64 | FK → workflow_meta.id |
| version | string | 版本号 |
| version_description | string | 版本描述 |
| canvas | string | 画布 Schema |
| input_params | string | 输入参数 |
| output_params | string | 输出参数 |
| creator_id | int64 | 创建者 |
| created_at | int64 | 创建时间 |
| deleted_at | gorm.DeletedAt | 软删除 |
| commit_id | string | 对应 commit ID |

### 1.4 workflow_snapshot - 工作流快照

Path: `domain/workflow/internal/repo/dal/model/workflow_snapshot.gen.go`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 自增主键 |
| workflow_id | int64 | FK → workflow |
| commit_id | string | 草稿 commit ID |
| canvas | string | 前端 Schema |
| input_params | string | 输入参数 |
| output_params | string | 输出参数 |
| created_at | int64 | 时间戳 |

用途：存储草稿在特定 commit ID 的历史快照，用于调试回放。

### 1.5 workflow_execution - 工作流执行记录

Path: `domain/workflow/internal/repo/dal/model/workflow_execution.gen.go`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 执行 ID |
| workflow_id | int64 | 工作流 ID |
| version | string | 工作流版本（草稿为空） |
| space_id | int64 | 空间 ID |
| mode | int32 | 1: 调试运行, 2: 发布运行, 3: 节点调试 |
| operator_id | int64 | 执行者用户 ID |
| connector_id | int64 | 连接器 ID |
| connector_uid | string | 连接器用户 ID |
| created_at | int64 | 创建时间 |
| log_id | string | 日志 ID |
| status | int32 | 1=运行中, 2=成功, 3=失败, 4=已取消, 5=中断 |
| duration | int64 | 执行时长（毫秒） |
| input | string | 实际输入（JSON） |
| output | string | 实际输出（JSON） |
| error_code | string | 错误码 |
| fail_reason | string | 失败原因 |
| input_tokens | int64 | 输入 token 数 |
| output_tokens | int64 | 输出 token 数 |
| updated_at | int64 | 更新时间 |
| root_execution_id | int64 | 顶层执行 ID（子工作流时有值） |
| parent_node_id | string | 父节点 key（SubWorkflow） |
| app_id | int64 | 应用 ID |
| node_count | int32 | 总节点数 |
| resume_event_id | int64 | 当前恢复事件 ID |
| agent_id | int64 | 关联 Agent ID |
| sync_pattern | int32 | 1: 同步, 2: 异步, 3: 流式 |
| commit_id | string | 草稿 commit ID |

### 1.6 node_execution - 节点执行记录

Path: `domain/workflow/internal/repo/dal/model/node_execution.gen.go`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 节点执行 ID |
| execute_id | int64 | FK → workflow_execution.id |
| node_id | string | 节点 key |
| node_name | string | 节点名称 |
| node_type | string | 节点类型 |
| created_at | int64 | 创建时间 |
| status | int32 | 1=等待, 2=运行中, 3=成功, 4=失败 |
| duration | int64 | 执行时长 |
| input | string | 实际输入（JSON） |
| output | string | 实际输出（JSON） |
| raw_output | string | 原始输出 |
| error_info | string | 错误信息 |
| error_level | string | 错误级别 |
| input_tokens | int64 | 输入 tokens |
| output_tokens | int64 | 输出 tokens |
| updated_at | int64 | 更新时间 |
| composite_node_index | int64 | 循环/批处理执行索引 |
| composite_node_items | string | 父复合节点的 items |
| parent_node_id | string | 父节点 key（循环/批处理内节点） |
| sub_execute_id | int64 | 子工作流执行 ID |
| extra | string | 额外信息（JSON） |

### 1.7 workflow_reference - 工作流引用关系

Path: `domain/workflow/internal/repo/dal/model/workflow_reference.gen.go`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 主键 |
| referred_id | int64 | 被引用的工作流 ID |
| referring_id | int64 | 引用方 ID |
| refer_type | int32 | 1: 子工作流, 2: 工具 |
| referring_biz_type | int32 | 1: 工作流, 2: Agent |
| created_at | int64 | 创建时间 |
| status | int32 | 0: 禁用, 1: 启用 |
| deleted_at | gorm.DeletedAt | 软删除 |

### 1.8 connector_workflow_version - 连接器-工作流版本映射

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 自增主键 |
| app_id | int64 | 应用 ID |
| connector_id | int64 | 连接器 ID |
| workflow_id | int64 | 工作流 ID |
| version | string | 版本 |
| created_at | int64 | 创建时间 |

### 1.9 chat_flow_role_config - ChatFlow 角色配置

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 主键 |
| workflow_id | int64 | 工作流 ID |
| name | string | 角色名 |
| description | string | 描述 |
| version | string | 版本 |
| avatar | string | 头像 URI |
| background_image_info | string | JSON |
| onboarding_info | string | JSON |
| suggest_reply_info | string | JSON |
| audio_config | string | JSON |
| user_input_config | string | JSON |
| creator_id | int64 | 创建者 |
| connector_id | int64 | 连接器 ID |

### 1.10 会话模板相关表（6张）

- `app_conversation_template_draft`: 草稿会话模板
- `app_conversation_template_online`: 在线会话模板
- `app_static_conversation_draft`: 草稿静态会话
- `app_static_conversation_online`: 在线静态会话
- `app_dynamic_conversation_draft`: 草稿动态会话
- `app_dynamic_conversation_online`: 在线动态会话

## 2. ER 关系图

```text
workflow_meta (1) ──── (1) workflow_draft         [同一 PK]
workflow_meta (1) ──── (*) workflow_version        [FK: workflow_id]
workflow_meta (1) ──── (*) workflow_reference      [referred_id / referring_id]
workflow_meta (1) ──── (*) workflow_snapshot        [FK: workflow_id]
workflow_meta (1) ──── (*) workflow_execution       [FK: workflow_id]
workflow_execution (1) ──── (*) node_execution      [FK: execute_id]
workflow_meta (1) ──── (*) chat_flow_role_config    [FK: workflow_id]
connector_workflow_version ──── (*)                  [FK: workflow_id, connector_id]
```

所有模型使用 GORM Gen 生成的 Query Builder，位于 `domain/workflow/internal/repo/dal/query/`。

## 3. 枚举常量

### 3.1 WorkflowMode（API 层）

```go
WorkflowMode_Workflow  = 0   // 普通工作流
WorkflowMode_Imageflow = 1   // 图像流
WorkflowMode_SceneFlow = 2   // 场景流
WorkflowMode_ChatFlow  = 3   // 对话流
WorkflowMode_All       = 100 // 查询用
```

### 3.2 执行状态

```go
// 工作流执行状态
WorkflowRunning     = 1
WorkflowSuccess     = 2
WorkflowFailed      = 3
WorkflowCancel      = 4
WorkflowInterrupted = 5

// 节点执行状态
NodeWaiting = 1
NodeRunning = 2
NodeSuccess = 3
NodeFailed  = 4
```

### 3.3 执行模式与同步模式

```go
ExecuteMode: debug / release / node_debug
SyncPattern: sync / async / stream
TaskType: foreground / background
BizType: agent / workflow
```

### 3.4 数据类型（节点变量）

```go
DataTypeString  = "string"
DataTypeInteger = "integer"
DataTypeNumber  = "number"
DataTypeBoolean = "boolean"
DataTypeTime    = "time"
DataTypeObject  = "object"
DataTypeArray   = "list"
DataTypeFile    = "file"
```

### 3.5 循环类型

```go
LoopType: array / count / infinite
```

### 3.6 错误处理类型

```go
ErrorProcessType: 1=Throw, 2=ReturnDefaultData, 3=ExceptionBranch
```

### 3.7 操作符类型（条件判断）

| 枚举 | 值 | 说明 |
|------|----|------|
| Equal | 1 | 等于 |
| NotEqual | 2 | 不等于 |
| Contains | 3 | 包含 |
| NotContains | 4 | 不包含 |
| StartsWith | 5 | 前缀匹配 |
| EndsWith | 6 | 后缀匹配 |
| GreaterThan | 7 | 大于 |
| GreaterThanEqual | 8 | 大于等于 |
| LessThan | 9 | 小于 |
| LessThanEqual | 10 | 小于等于 |
| IsEmpty | 11 | 为空 |
| IsNotEmpty | 12 | 非空 |
| In | 13 | 在集合中 |
| NotIn | 14 | 不在集合中 |
| MatchRegex | 15 | 正则匹配 |
| NotMatchRegex | 16 | 正则不匹配 |

### 3.8 引用源类型

```go
RefSourceType: block-output / global_variable_app / system / user
```

## 4. 仓储模式

`RepositoryImpl` 位于 `domain/workflow/internal/repo/repository.go`，实现 `workflow.Repository` 接口。使用 GORM Gen 生成的 typed query builder。

主要方法：`CreateMeta`, `CreateOrUpdateDraft`, `CreateVersion`, `Delete`（事务级联）, `GetMeta`, `GetEntity`, `DraftV2`, `GetVersion`, `GetLatestVersion`, `MGetDrafts`, `MGetLatestVersion`, `MGetMetas`, `MGetReferences`, `CreateSnapshotIfNeeded`, `CopyWorkflow`, `CreateChatFlowRoleConfig` 等。

`ExecuteHistoryStore` 位于 `execute_history_store.go`，负责：`CreateWorkflowExecution`, `UpdateWorkflowExecution`, `GetWorkflowExecution`, `CreateNodeExecution`, `UpdateNodeExecution`, `GetNodeExecutionsByWfExeID`, `TryLockWorkflowExecution`（CAS）, `CancelAllRunningNodes`, Redis 流式输出缓存。
