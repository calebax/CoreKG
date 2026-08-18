# 执行生命周期与模式

## 1. 工作流生命周期总览

```
创建 → 编辑草稿 → 测试运行 → 发布版本 → 生产执行 → 执行历史查询
  │         │           │            │           │           │
  ▼         ▼           ▼            ▼           ▼           ▼
meta+draft  save      test_run    publish     run/stream   get_process
            canvas    nodeDebug   version     resume       run_history
            updateMeta            release    cancel
            copy
            delete
```

## 2. 创建阶段

### 2.1 API

`POST /api/workflow_api/create`

### 2.2 流程

```
CreateWorkflow(request)
  → 校验 space 权限
  → IDGenerator 生成工作流 ID
  → repo.CreateMeta(entity.MetaCreate)
      → INSERT workflow_meta
      → INSERT workflow_draft（初始化空 canvas）
  → ChatFlow 模式额外创建会话模板
  → 发布资源事件（Created）到搜索索引
  → 返回 workflow_id
```

### 2.3 初始状态

- `workflow_meta.status = 0`（未发布）
- `workflow_draft.canvas = "{}"`（空画布）
- `workflow_draft.test_run_success = false`
- `workflow_draft.modified = false`

## 3. 编辑阶段

### 3.1 保存草稿

`POST /api/workflow_api/save`

```
SaveWorkflow(workflowID, canvas JSON)
  → 解析 Canvas → 提取 InputParams / OutputParams
  → repo.CreateOrUpdateDraft(workflow_draft)
      → UPSERT workflow_draft
  → 计算 test_run_success（canvas 变更后标记为 false）
  → 标记 modified = true
```

### 3.2 更新元数据

`POST /api/workflow_api/update_meta`

更新名称、描述、图标等元信息，异步发布资源事件。

### 3.3 获取画布信息

`POST /api/workflow_api/canvas`

返回完整画布数据 + VCS 信息（modified、test_run_success、latest_version）+ 开发状态。

## 4. 测试运行

### 4.1 整工作流测试

`POST /api/workflow_api/test_run`

```
TestRun(workflowID, input)
  → 加载草稿 Canvas
  → DomainSVC.AsyncExecute(config{Mode: debug})
      → parse canvas → WorkflowSchema → Build Graph
      → convert/validate inputs
      → prepare WorkflowRunner
      → create WorkflowExecution record (mode=1)
      → wf.AsyncRun(input)      ← goroutine
  → 立即返回 executeID
```

前端通过 `GET /api/workflow_api/get_process?execute_id=xxx` 轮询状态。

### 4.2 单节点调试

`POST /api/workflow_api/nodeDebug`

```
NodeDebug(workflowID, nodeID, input)
  → AsyncExecuteNode(config{Mode: node_debug})
  → 仅执行指定节点（fromNode=true，无 Entry/Exit）
  → 返回 executeID
```

### 4.3 测试恢复

`POST /api/workflow_api/test_resume`

恢复被中断的测试执行（Q&A、InputReceiver 节点触发中断）。

## 5. 发布阶段

### 5.1 发布

`POST /api/workflow_api/publish`

```
PublishWorkflow(workflowID, version, description)
  → 事务操作:
      → INSERT workflow_version（canvas + params + commit_id）
      → UPDATE workflow_draft.modified = false
      → UPDATE workflow_meta.status = 1, latest_version
  → 返回版本号
```

### 5.2 版本列表

`POST /api/workflow_api/list_publish_workflow`

分页查询 `workflow_version`，按 created_at 降序。

### 5.3 Canvas 历史

`GET /api/workflow_api/history_schema`

获取指定版本的 canvas schema。

## 6. 生产执行

### 6.1 同步执行

`POST /v1/workflow/run`（OpenAPI）

```
OpenAPIRunFlow(workflowID, input)
  → 验证 API Key 权限
  → 加载已发布版本 Canvas
  → DomainSVC.SyncExecute(config{Mode: release, SyncPattern: sync})
      → build graph
      → prepare runner
      → wf.SyncRun(input)
      → 等待完成
  → 返回 {code, data: {output, token_info, execute_id, ...}}
```

### 6.2 流式执行

`POST /v1/workflow/stream_run`（OpenAPI, SSE）

```
OpenAPIStreamRunFlow(workflowID, input)
  → StreamExecute(config{SyncPattern: stream})
      → create Pipe (StreamReader/Writer)
      → prepare runner with StreamWriter
      → wf.AsyncRun(input)     ← goroutine
  → 返回 StreamReader
  → SSE Writer 逐 chunk 发送
```

SSE 事件格式：
```
event: message
data: {"id":"...","event":"workflow.running","data":{...}}

event: message
data: {"id":"...","event":"node.start","data":{...}}

event: message
data: {"id":"...","event":"workflow.done","data":{...}}
```

### 6.3 流式恢复

`POST /v1/workflow/stream_resume`（OpenAPI, SSE）

```
OpenAPIStreamResumeFlow(workflowID, executeID, resumeData)
  → StreamResume(config, resumeRequest)
      → 加载已有执行记录
      → 验证状态为 Interrupted
      → 获取中断事件 & checkpoint
      → 重建工作流图
      → prepare runner with state modifier
      → wf.AsyncRun(nil)     ← 从检查点继续
  → SSE Writer 逐 chunk 发送
```

### 6.4 CoreKG 扩展执行

`POST /v1/workflow/ygrun`

CoreKG 特有的执行入口，附加 YYGU 认证和权限校验。

## 7. ChatFlow 执行

ChatFlow 是工作流的特殊模式（mode=3），增加了会话管理和消息持久化。

### 7.1 ChatFlow 运行

`POST /v1/workflows/chat`（SSE）

```
OpenAPIChatFlowRun(workflowID, conversationID, message, ...)
  1. 校验权限
  2. GetOrCreateConversation
  3. 创建 Agent Run 记录（crossagentrun）
  4. 创建用户消息（crossmessage）
  5. 检查是否有待处理的中断
      → 有: StreamResume
      → 无: StreamExecute
  6. 转换流式输出为 ChatFlow 事件序列:
      → WorkflowRunning → Created + InProgress
      → NodeStreamingOutput → 文本增量
      → WorkflowSuccess → message completed + suggest replies + Done
      → WorkflowFailed → Error 事件
      → WorkflowInterrupted → Input/QA card DSL + RequiresAction
  7. SSE 发送
```

### 7.2 会话管理

- `POST /v1/workflow/conversation/create` — 创建或获取会话
- 会话绑定 workflow_id + connector_id + user_id
- 支持静态/动态/模板三种会话类型

## 8. 执行状态查询

### 8.1 获取执行进度

`GET /api/workflow_api/get_process`

```
GetProcess(executeID)
  → DomainSVC.GetExecution(executeID)
      → query workflow_execution
      → query node_executions by execute_id
      → 转换节点执行状态
  → 返回:
      - 工作流状态（running/success/fail/interrupted）
      - 所有节点执行详情（input/output/status/duration/tokens）
      - 中断事件列表（如有）
      - Token 使用汇总
```

### 8.2 节点执行历史

`GET /api/workflow_api/get_node_execute_history`

获取指定节点的执行历史（测试运行输入或特定执行的节点详情）。

### 8.3 OpenAPI 执行历史

`GET /v1/workflow/get_run_history`

OpenAPI 版本的执行历史查询。

## 9. 取消执行

`POST /api/workflow_api/cancel`

```
CancelWorkFlow(executeID)
  → 更新 workflow_execution.status = Cancel
  → 取消所有 running 状态的 node_execution
  → 设置 Redis cancel flag
  → 上下文 cancel 传播到所有 goroutine
```

## 10. 中断与恢复机制

### 10.1 触发中断的节点

| 节点 | 中断原因 |
|------|----------|
| QuestionAnswer | 等待用户回答问题 |
| InputReceiver | 等待用户输入 |
| Loop | 循环迭代间暂停（可选） |
| Batch | 批次间暂停（可选） |

### 10.2 中断流程

```
节点执行 → 检测到需要用户交互
  → 创建 InterruptEvent
  → CheckpointStore.Save(state)
  → InterruptEventStore.Save(events)
  → 更新 workflow_execution.status = Interrupted
  → 发射 WorkflowInterrupt 事件
  → 停止后续节点执行
```

### 10.3 恢复流程

```
API 收到 resume 请求
  → GetWorkflowExecution（验证状态=Interrupted）
  → TryLockWorkflowExecution（CAS，防止并发恢复）
  → PopFirstInterruptEvent
  → CheckpointStore.Load(state)
  → 构建 state modifier（注入用户输入到中断节点）
  → 重建工作流图（同一版本）
  → AsyncRun(nil, compose.WithStateModifier(...))
      → Eino 从检查点恢复执行
```

### 10.4 检查点存储

两种实现：
- **Memory**: 内存存储（调试用，单实例）
- **Redis**: 持久化存储（生产用，支持多实例）

存储内容：
- 全局状态（GlobalState）
- 节点局部状态（NodeState）
- 中断事件（InterruptEvents）

## 11. 执行模式汇总

| 模式 | 触发 API | 用途 | 状态记录 |
|------|----------|------|----------|
| debug | test_run | 开发测试 | workflow_execution.mode=1 |
| release | /v1/workflow/run, stream_run | 生产执行 | workflow_execution.mode=2 |
| node_debug | nodeDebug | 单节点调试 | workflow_execution.mode=3 |

| 同步模式 | 说明 | 返回值 |
|----------|------|--------|
| sync | 同步等待完成 | JSON 结果 |
| async | 立即返回 executeID | {execute_id} |
| stream | SSE 流式 | StreamReader |

## 12. 错误处理策略

节点级别的错误处理由 `ErrorProcessType` 控制：

| 策略 | 行为 |
|------|------|
| Throw (1) | 直接传播错误，工作流失败 |
| ReturnDefaultData (2) | 返回节点配置的默认数据，工作流继续 |
| ExceptionBranch (3) | 路由到异常分支，携带错误信息 |

工作流级错误：
- 更新 workflow_execution.status = Failed
- 记录 error_code 和 fail_reason
- 发射 WorkflowFailed 事件

## 13. 复制与迁移

### 13.1 复制工作流

`POST /api/workflow_api/copy`

浅复制：仅复制 meta + draft，不复制依赖资源。

### 13.2 从模板复制

`POST /api/workflow_api/copy_wk_template`

从官方模板创建工作流。

### 13.3 跨应用复制

```
CopyWorkflowFromAppToLibrary(workflowID, spaceID, appID)
  → 深度复制工作流
  → 复制依赖的插件、知识库、数据库
  → 更新引用关系
  → 验证并返回问题列表
```

### 13.4 批量复制

```
DuplicateWorkflowsByAppID(sourceAppID, targetAppID, externalResource)
  → 遍历源应用所有工作流
  → 逐个深度复制到目标应用
```

### 13.5 依赖资源追踪

`GetWorkflowDependenceResource` 解析 canvas 中的：
- Plugin IDs（UsePlugin 节点）
- Knowledge IDs（UseKnowledge 节点）
- Database IDs（UseDatabase 节点）

复制/移动时自动携带这些依赖资源。
