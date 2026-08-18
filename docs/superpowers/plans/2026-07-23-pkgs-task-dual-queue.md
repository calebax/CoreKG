# pkgs/task 双队列重构实施计划

> **给 agentic workers 的说明：** 必须使用子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 来逐任务实施此计划。步骤使用 checkbox（`- [ ]`）语法进行跟踪。

**目标：** 将 pkgs/task 重构为纯任务分发引擎，包含两个 JetStream 队列（dispatch + result），将回调/依赖逻辑从共享包中移除，下沉到业务消费方。

**架构：** pkgs/task 变为任务分发引擎：CreateTask → DB → JetStream dispatch 队列 → TypeScript Worker（PullSubscribe） → JetStream result 队列 → 业务消费方订阅并处理依赖推进。两个 JetStream streams：`CORE_TASK_DISPATCH`（workqueue retention）和 `CORE_TASK_RESULT`（limits retention）。

**技术栈：** Go 1.22+, nats.go v1.34.1 (JetStream), TypeScript (apps/worker pnpm monorepo), GORM, gin (via yg-go)

## 全局约束

- Go module 路径：`github.com/insmtx/corekg`
- 构建方式：`go build -mod=vendor`
- 依赖已 vendor；新增依赖需要执行 `go mod tidy && go mod vendor`
- NATS dispatch subject：`core.task.dispatch.<short_name>`
- NATS result subject：`core.task.result.<short_name>`
- Dispatch stream：`CORE_TASK_DISPATCH`，workqueue retention，MaxDeliver=3，AckWait=5min
- Result stream：`CORE_TASK_RESULT`，limits retention，MaxAge=24h
- 除非被要求，不要添加注释
- Struct tags：`json` = snake_case
- 测试命令：`go test ./pkgs/task/...`（需要 MySQL + NATS）

---

## 文件结构

```
pkgs/task/
├── task.go                  # 修改：移除 callBackMap、RegisterCallBack、GetCallBack、GetCallBackMap
├── biz.go                   # 修改：移除 GetNextStepTask，修改 SaveTask（纯 DB 更新）
├── nats_bridge.go           # 重写：双队列（dispatch + result），移除 SubscribeCallbacks
├── task_queue.go            # 修改：PushTaskQueue 改为发布到 dispatch JetStream
├── crud.go                  # 修改：移除 GetNextStepTask，简化 SaveTask
├── dao.go                   # 保留
├── dao_base.go              # 保留
├── task_info.go             # 修改：添加 ResultMessage struct
├── task_server.go           # 修改：移除已废弃的 handler 实现（保留 stub）
└── task_health_check.go     # 保留

apps/worker/packages/rpc/
├── src/server.ts            # 删除（被 JetStream consumer 替代）
├── src/client.ts            # 删除
├── src/schema.ts            # 保留（RPCResponse 复用于 result messages）
└── src/subjects.ts          # 修改：dispatch + result subjects

apps/worker/packages/workers/
├── src/types.ts             # 修改：添加 result publish 接口
├── src/registry.ts          # 修改：更新 subject 引用
└── src/index.ts             # 修改：导出 result 类型

apps/worker/apps/worker/
└── src/worker-main.ts       # 重写：JetStream PullSubscribe + result publish

apps/ketask/
├── cmd/init.go              # 修改：移除 SubscribeCallbacks，添加 result subscriptions
├── cmd/main.go              # 修改：接入 result consumers
└── internal/jobs/jobs.go    # 修改：添加 result consumer 协程

apps/kecore/models/coretask/
├── crud.go                  # 修改：将 GetNextStepTask 移入此处，添加 result consumer 逻辑
└── generate_task.go         # 保留（无变更）

apps/corekg/
└── cmd/init.go              # 修改：在 corekg 上下文中注册 result consumers
```

---

### 任务 1：添加 ResultMessage 类型并更新 task_info.go

**文件：**
- 修改：`pkgs/task/task_info.go`

**接口：**
- 产出：`ResultMessage` struct —— 被 nats_bridge（任务 2）、TS workers（任务 5）和业务消费方（任务 7）使用

- [ ] **步骤 1：在 task_info.go 中添加 ResultMessage struct**

在已有类型之后添加：

```go
type ResultMessage struct {
	TaskID       uint       `json:"task_id"`
	WorkerID     string     `json:"worker_id"`
	TaskType     string     `json:"task_type"`
	Status       TaskStatus `json:"status"`
	Result       string     `json:"result"`
	ErrorMessage string     `json:"error_message"`
}
```

- [ ] **步骤 2：验证编译**

运行：`go build -mod=vendor ./pkgs/task/...`
预期：编译成功

- [ ] **步骤 3：提交**

```bash
git add pkgs/task/task_info.go
git commit -m "feat(task): add ResultMessage type for result queue"
```

---

### 任务 2：重写 nats_bridge.go 实现双队列

**文件：**
- 修改：`pkgs/task/nats_bridge.go`

**接口：**
- 消费：来自任务 1 的 `ResultMessage`
- 消费：来自 task.go 的 `Task` struct、`TaskStatus`
- 产出：`NewNATSBridge(nc)`、`PublishTaskRPC(taskType, payload)`、`PublishResult(result)`、`EnsureStreams(nc)`、`taskTypeToShort` map
- 移除：`SubscribeCallbacks()`、`RPCCallbackResponse`

- [ ] **步骤 1：重写 nats_bridge.go**

```go
package task

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/ygpkg/yg-go/logs"
)

type NATSBridge struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func NewNATSBridge(nc *nats.Conn) *NATSBridge {
	if nc == nil {
		return nil
	}
	js, err := nc.JetStream()
	if err != nil {
		logs.Errorf("create JetStream context failed: %v", err)
		return nil
	}
	return &NATSBridge{nc: nc, js: js}
}

var taskTypeToShort = map[string]string{
	"ke.copy_task":        "copy",
	"ke.doc_to_pdf_task":  "copy",
	"ke.prase_pdf_task":   "pdf_extract",
	"ke.prase_video_task": "video_extract",
	"ke.mind_map_task":    "mindmap",
	"ke.analysis_task":    "analysis",
	"ke.description_task": "desc",
	"ke.knowledge_task":   "split_text_chunk",
	"ke.insert_index":     "insert_index",
}

func DispatchSubject(taskType string) (string, error) {
	short, ok := taskTypeToShort[taskType]
	if !ok {
		return "", fmt.Errorf("no dispatch mapping for task type: %s", taskType)
	}
	return "core.task.dispatch." + short, nil
}

func ResultSubject(taskType string) (string, error) {
	short, ok := taskTypeToShort[taskType]
	if !ok {
		return "", fmt.Errorf("no result mapping for task type: %s", taskType)
	}
	return "core.task.result." + short, nil
}

func (b *NATSBridge) PublishTaskRPC(taskType string, payload []byte) error {
	if b == nil || b.js == nil {
		return fmt.Errorf("nats bridge not initialized")
	}
	subject, err := DispatchSubject(taskType)
	if err != nil {
		return err
	}
	_, err = b.js.Publish(subject, payload)
	if err != nil {
		return fmt.Errorf("nats dispatch publish %s: %w", subject, err)
	}
	return nil
}

func (b *NATSBridge) PublishResult(result *ResultMessage) error {
	if b == nil || b.js == nil {
		return fmt.Errorf("nats bridge not initialized")
	}
	subject, err := ResultSubject(result.TaskType)
	if err != nil {
		return err
	}
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	_, err = b.js.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("nats result publish %s: %w", subject, err)
	}
	return nil
}

func EnsureDispatchStream(js nats.JetStreamContext) error {
	_, err := js.StreamInfo("CORE_TASK_DISPATCH")
	if err == nil {
		return nil
	}
	if err != nats.ErrStreamNotFound {
		return fmt.Errorf("check dispatch stream: %w", err)
	}
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "CORE_TASK_DISPATCH",
		Subjects: []string{"core.task.dispatch.*"},
		Storage:  nats.FileStorage,
		Retention: nats.WorkQueuePolicy,
		MaxMsgs:  100000,
		MaxBytes: 256 * 1024 * 1024,
	})
	if err != nil {
		return fmt.Errorf("create dispatch stream: %w", err)
	}
	logs.Infof("created NATS JetStream stream: CORE_TASK_DISPATCH")
	return nil
}

func EnsureResultStream(js nats.JetStreamContext) error {
	_, err := js.StreamInfo("CORE_TASK_RESULT")
	if err == nil {
		return nil
	}
	if err != nats.ErrStreamNotFound {
		return fmt.Errorf("check result stream: %w", err)
	}
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "CORE_TASK_RESULT",
		Subjects: []string{"core.task.result.*"},
		Storage:  nats.FileStorage,
		Retention: nats.LimitsPolicy,
		MaxAge:   24 * time.Hour,
		MaxMsgs:  1000000,
		MaxBytes: 1024 * 1024 * 1024,
	})
	if err != nil {
		return fmt.Errorf("create result stream: %w", err)
	}
	logs.Infof("created NATS JetStream stream: CORE_TASK_RESULT")
	return nil
}

func (b *NATSBridge) EnsureStreams() error {
	if b == nil || b.js == nil {
		return fmt.Errorf("nats bridge not initialized")
	}
	if err := EnsureDispatchStream(b.js); err != nil {
		return err
	}
	return EnsureResultStream(b.js)
}
```

- [ ] **步骤 2：验证编译**

运行：`go build -mod=vendor ./pkgs/task/...`
预期：其他文件出现编译错误（SubscribeCallbacks 已移除）—— 此阶段属预期情况

- [ ] **步骤 3：提交**

```bash
git add pkgs/task/nats_bridge.go
git commit -m "refactor(task): rewrite nats_bridge for dual queue architecture"
```

---

### 任务 3：从 pkgs/task 中移除回调/依赖逻辑

**文件：**
- 修改：`pkgs/task/task.go` —— 移除 callBackMap、RegisterCallBack、GetCallBack、GetCallBackMap、callBack interface
- 修改：`pkgs/task/biz.go` —— 移除 GetNextStepTask；简化 SaveTask（移除自动重试重新推送）
- 修改：`pkgs/task/task_queue.go` —— 更新 PushTaskQueue 使用 JetStream dispatch
- 修改：`pkgs/task/task_server.go` —— 移除 TaskCallBack 实现（保留 stub），移除 GetPendingTask 实现
- 修改：`pkgs/task/crud.go` —— 如果存在则移除 GetNextStepTask

**接口：**
- 消费：来自任务 2 的 `NATSBridge.PublishTaskRPC`
- 产出：清理后的 pkgs/task，不再包含回调/依赖逻辑

- [ ] **步骤 1：修改 task.go —— 移除回调基础设施**

移除：`callBack` interface、`callBackMap` var、`RegisterCallBack()`、`GetCallBack()`、`GetCallBackMap()`、`InitTask()` 实现（保留函数，仅调用 `InitTaskDBStauts`）。

保留：`TaskStatus` 常量、`Task` struct、`InitDB()`。

- [ ] **步骤 2：修改 biz.go —— 移除 GetNextStepTask，简化 SaveTask**

完全移除 `GetNextStepTask` 函数。

简化 `SaveTask`：

```go
func SaveTask(tsk *Task) error {
	if tsk.StartAt != nil && tsk.EndAt != nil {
		if tsk.StartAt.Before(*tsk.EndAt) {
			tsk.Cost = int64(tsk.EndAt.Sub(*tsk.StartAt).Seconds())
		}
	}
	if tsk.TaskStatus == TaskStatusFail {
		tsk.Redo++
	}
	return dbutil.Core().Save(tsk).Error
}
```

- [ ] **步骤 3：修改 task_queue.go —— JetStream dispatch**

```go
package task

import (
	"context"
	"fmt"

	"github.com/ygpkg/yg-go/logs"
)

var natsBridge *NATSBridge

func SetNATSBridge(b *NATSBridge) {
	natsBridge = b
}

func PushTaskQueue(ctx context.Context, taskType string) error {
	if natsBridge == nil {
		return fmt.Errorf("nats bridge not initialized")
	}
	tsk, err := GetOnePendingTask(taskType, "nats-dispatcher")
	if err != nil {
		return fmt.Errorf("get pending task: %w", err)
	}
	if tsk == nil {
		logs.InfoContextf(ctx, "no pending task for type %s", taskType)
		return nil
	}
	err = natsBridge.PublishTaskRPC(tsk.TaskType, []byte(tsk.Payload))
	if err != nil {
		logs.ErrorContextf(ctx, "nats dispatch publish failed: %v, task_id: %d", err, tsk.ID)
		return err
	}
	logs.InfoContextf(ctx, "nats dispatched task: id=%d type=%s subject=%s", tsk.ID, tsk.TaskType, "core.task.dispatch.*")
	return nil
}
```

- [ ] **步骤 4：修改 task_server.go —— 移除已废弃的 handler 实现**

保留函数签名，但返回 deprecated/error 响应。移除对回调函数的 import。

- [ ] **步骤 5：验证编译**

运行：`go build -mod=vendor ./pkgs/task/...`
预期：编译成功

- [ ] **步骤 6：验证 core_task 包仍可编译**

运行：`go build -mod=vendor ./apps/kecore/models/coretask/...`
预期：可能因已移除的函数而编译失败 —— 属预期情况，将在任务 7 中修复

- [ ] **步骤 7：提交**

```bash
git add pkgs/task/
git commit -m "refactor(task): remove callback/dependency logic, simplify to dispatch engine"
```

---

### 任务 4：更新 TypeScript subjects 和 result 类型

**文件：**
- 修改：`apps/worker/packages/rpc/src/subjects.ts`
- 修改：`apps/worker/packages/workers/src/types.ts`
- 修改：`apps/worker/packages/workers/src/registry.ts`

**接口：**
- 产出：为任务 5 准备更新后的 subject 常量和 result 类型

- [ ] **步骤 1：更新 subjects.ts**

```typescript
export const DISPATCH_SUBJECTS = {
  analysis: "core.task.dispatch.analysis",
  copy: "core.task.dispatch.copy",
  desc: "core.task.dispatch.desc",
  mindmap: "core.task.dispatch.mindmap",
  pdfExtract: "core.task.dispatch.pdf_extract",
  videoExtract: "core.task.dispatch.video_extract",
  splitTextChunk: "core.task.dispatch.split_text_chunk",
  insertIndex: "core.task.dispatch.insert_index",
} as const;

export const RESULT_SUBJECTS = {
  analysis: "core.task.result.analysis",
  copy: "core.task.result.copy",
  desc: "core.task.result.desc",
  mindmap: "core.task.result.mindmap",
  pdfExtract: "core.task.result.pdf_extract",
  videoExtract: "core.task.result.video_extract",
  splitTextChunk: "core.task.result.split_text_chunk",
  insertIndex: "core.task.result.insert_index",
} as const;

// Keep backward compat alias
export const RPC_SUBJECTS = DISPATCH_SUBJECTS;
```

- [ ] **步骤 2：更新 types.ts —— 添加 result publish 接口**

在已有类型中添加：

```typescript
export interface TaskResultMessage {
  task_id: number;
  worker_id: string;
  task_type: string;
  status: "success" | "fail";
  result?: string;
  error_message?: string;
}

export interface TaskHandlerDef {
  name: string;
  dispatchSubject: string;
  resultSubject: string;
  handler: TaskHandlerFn;
}
```

- [ ] **步骤 3：更新 registry.ts**

更新每个条目，使用新常量中的 `dispatchSubject` 和 `resultSubject`。

- [ ] **步骤 4：验证 TypeScript 编译**

运行：`cd apps/worker && pnpm build`（或等效命令）
预期：编译成功

- [ ] **步骤 5：提交**

```bash
git add apps/worker/
git commit -m "feat(worker): update subjects for dual queue dispatch/result"
```

---

### 任务 5：重写 TypeScript worker 为 JetStream PullSubscribe + result publish

**文件：**
- 创建：`apps/worker/packages/nats/src/dispatch-consumer.ts`
- 创建：`apps/worker/packages/nats/src/result-publisher.ts`
- 修改：`apps/worker/apps/worker/src/worker-main.ts`
- 修改：`apps/worker/packages/rpc/src/server.ts` —— 可以删除或保留为未使用状态

**接口：**
- 消费：来自任务 4 的 `TaskHandlerDef`，来自 nats.go 的 JetStream
- 产出：Worker 从 dispatch 队列消费消息并发布结果到 result 队列

- [ ] **步骤 1：创建 dispatch-consumer.ts**

JetStream PullSubscribe consumer，支持可配置的并发度。对每条消息：解析 TaskPayload，调用 handler，成功/失败时发布 result，确认消息。

- [ ] **步骤 2：创建 result-publisher.ts**

封装 JetStream publish，向 `core.task.result.<type>` subjects 发布消息。

- [ ] **步骤 3：重写 worker-main.ts**

移除 RPCServer 初始化。替换为每个已注册 handler 对应的 dispatch consumer。每个 handler 拥有自己的 PullSubscribe，使用 durable consumer 名称。handler 完成后，通过 result-publisher 发布 ResultMessage。

移除 `chunkConsumer` JetStream 路径（现已统一）。

- [ ] **步骤 4：验证 TypeScript 编译**

运行：`cd apps/worker && pnpm build`
预期：编译成功

- [ ] **步骤 5：提交**

```bash
git add apps/worker/
git commit -m "refactor(worker): JetStream PullSubscribe dispatch + result publish"
```

---

### 任务 6：将 GetNextStepTask 迁移到 kecore

**文件：**
- 修改：`apps/kecore/models/coretask/crud.go` —— 添加 GetNextStepTask（从 pkgs/task/biz.go 移入）

**接口：**
- 消费：来自 pkgs/task 的 `task.Task`、`task.TaskStatus`
- 产出：`coretask.GetNextStepTask(tsk *task.Task) ([]*task.Task, error)` —— 被任务 7 使用

- [ ] **步骤 1：将 GetNextStepTask 复制到 coretask/crud.go**

从旧的 `pkgs/task/biz.go` 实现中复制该函数。调整 imports 以使用 `pkgs/task.Task` 类型。该函数直接查询 core_task 表（同一个 DB）。

- [ ] **步骤 2：验证编译**

运行：`go build -mod=vendor ./apps/kecore/models/coretask/...`
预期：编译成功

- [ ] **步骤 3：提交**

```bash
git add apps/kecore/models/coretask/crud.go
git commit -m "feat(coretask): add GetNextStepTask (moved from pkgs/task)"
```

---

### 任务 7：为 ketask 添加 result 队列消费方

**文件：**
- 修改：`apps/ketask/cmd/init.go` —— 移除 SubscribeCallbacks，初始化 result consumers
- 修改：`apps/ketask/cmd/main.go` —— 接入 result consumers
- 修改：`apps/ketask/internal/jobs/jobs.go` —— 添加 result consumer 协程
- 创建：`apps/ketask/internal/jobs/result_consumer.go` —— result 队列订阅 + 回调 + 下一步骤逻辑

**接口：**
- 消费：来自任务 2 的 `NATSBridge`，来自任务 6 的 `coretask.GetNextStepTask`，`RegisterCallBack` 函数（现在位于 ketask 本地）
- 产出：ketask 订阅 result 队列，处理任务完成 + 依赖推进

- [ ] **步骤 1：创建 result_consumer.go**

该文件包含：
1. 本地 `callBackMap`（从 pkgs/task 移入）
2. ketask 任务类型的 `RegisterCallBack()`
3. `StartResultConsumer(nc, js)` —— 通过 JetStream 订阅 `core.task.result.>`
4. Handler：反序列化 ResultMessage → 加载 task → 执行回调 → SaveTask → GetNextStepTask → PushTaskQueue
5. 失败处理：检查 redo < task_config_redo → 如果可重试则重新推送

- [ ] **步骤 2：修改 init.go**

移除 `task.SubscribeCallbacks(nc)` 调用。添加 result consumer 初始化。将回调注册从 `initTask()` 移到 result_consumer 初始化逻辑中。

- [ ] **步骤 3：修改 main.go**

将 result consumer 启动接入启动流程。

- [ ] **步骤 4：修改 jobs.go**

将 result consumer 添加为一个协程。

- [ ] **步骤 5：验证编译**

运行：`go build -mod=vendor ./apps/ketask/...`
预期：编译成功

- [ ] **步骤 6：提交**

```bash
git add apps/ketask/
git commit -m "feat(ketask): add result queue consumer with callback + next-step"
```

---

### 任务 8：迁移 SuccessStatusRoutine 及剩余的 kecore 回调

**文件：**
- 修改：`apps/kecore/models/coretask/crud.go` —— 保留 SuccessStatusRoutine（轮询仍可作为安全网保留）
- 修改：`apps/corekg/cmd/init.go` —— 如有需要，在 corekg 上下文中注册 result consumers

**接口：**
- 消费：来自任务 7 的 result consumer 模式
- 产出：corekg 也订阅相关的 result subjects

- [ ] **步骤 1：审查 corekg init.go 中的回调注册**

检查 corekg 是否注册了回调。如果有，参照任务 7 在 corekg 上下文中创建 result consumer。

- [ ] **步骤 2：验证完整构建**

运行：`go build -mod=vendor ./apps/corekg/...`
预期：编译成功

- [ ] **步骤 3：提交**

```bash
git add apps/corekg/ apps/kecore/
git commit -m "feat(corekg): migrate result consumers from pkgs/task callbacks"
```

---

### 任务 9：清理并验证端到端流程

- [ ] **步骤 1：移除死代码引用**

搜索对已移除函数的剩余引用（`SubscribeCallbacks`、`RegisterCallBack`、`GetCallBack`、pkgs/task 中的 `GetNextStepTask`、TS 中的 `RPCServer`）。清理 imports。

- [ ] **步骤 2：验证完整单体仓库构建**

运行：`go build -mod=vendor ./...`
预期：编译成功

- [ ] **步骤 3：验证 TypeScript 构建**

运行：`cd apps/worker && pnpm build`
预期：编译成功

- [ ] **步骤 4：提交**

```bash
git add -A
git commit -m "chore(task): cleanup dead references after dual queue migration"
```
