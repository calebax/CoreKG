# 工作流引擎核心

## 1. 引擎概述

工作流引擎基于 **CloudWeGo Eino compose** 框架构建。核心代码位于 `apps/workflow/domain/workflow/internal/compose/` 目录。

Eino compose 是一个通用的有向图编排框架，提供：

- 图构建（Workflow）
- 节点编排（AddNode / AddCompositeNode）
- 依赖解析（FieldMapping）
- 检查点/恢复（Checkpoint）
- 流式传输（Stream）
- 回调系统（Callbacks）

## 2. Workflow 结构体

Path: `domain/workflow/internal/compose/workflow.go`

```go
type Workflow struct {
    *workflow                              // Eino compose.Workflow[map[string]any, map[string]any]
    hierarchy         map[vo.NodeKey]vo.NodeKey  // 父子节点关系
    connections       []*schema.Connection        // 连接线
    requireCheckpoint bool                        // 是否需要检查点
    entry             *compose.WorkflowNode         // 入口节点
    inner             bool                        // 是否内部工作流（Loop/Batch）
    fromNode          bool                        // 是否从单节点构建（无 Entry/Exit）
    streamRun         bool                        // 是否流式运行
    Runner            compose.Runnable[map[string]any, map[string]any]
    input             map[string]*vo.TypeInfo     // 输入参数定义
    output            map[string]*vo.TypeInfo     // 输出参数定义
    terminatePlan     vo.TerminatePlan            // 终止计划
    schema            *schema.WorkflowSchema        // 工作流 Schema
}
```

## 3. 工作流构建流程

### 3.1 NewWorkflow

```
NewWorkflow(ctx, schema, ...opts)
  → sc.Init()                    // 初始化 Schema
  → 创建 Eino workflow           // compose.NewWorkflow
  → 遍历 schema.Nodes
      → 普通节点: AddNode()
      → 复合节点(Loop/Batch): AddCompositeNode()
  → Compile()                    // 连接 START→Entry, Exit→END, 编译图
```

### 3.2 AddNode 流程

每个节点添加时：

1. 根据 NodeType 查找 Adaptor（NodeAdaptor 接口）
2. Adaptor 将前端 Canvas Node 转换为后端 NodeSchema
3. 创建 Eino 节点（含 Invoke/Stream/Collect/Transform 包装）
4. 注册到 Eino workflow

### 3.3 AddCompositeNode 流程

复合节点（Loop/Batch）：

1. 创建内部子工作流（getInnerWorkflow）
2. 子工作流继承外层节点，通过 carry-over 字段映射传递数据
3. 设置循环/批处理特定的 pre/post processor
4. 注册为 Eino ComposeNode

### 3.4 Compile 流程

```
Compile()
  → 连接 START → entry node key
  → 连接 exit node key → END
  → 如果有检查点需求: compose.WithCheckpointStore
  → wf.Compile(compileOpts...)
```

## 4. 依赖解析

工作流引擎通过 **FieldMapping** 机制解析节点间的数据依赖：

### 4.1 依赖类型

| 类型 | 说明 |
|------|------|
| 直接依赖 | 同一层级中已连接的节点 |
| 间接依赖 | 同一层级中未连接但存在的节点 |
| 跨层级引用 | 子节点引用父工作流的输出 |
| 静态值 | 输入项上设置的字面值 |
| 变量 | 全局变量/应用变量（在 state pre-handler 中处理） |

### 4.2 字段映射路径

使用 Eino 的 `compose.FieldMapping`，路径格式如：

- `nodeOutput.fieldName` — 节点输出字段
- `nodeOutput.items[*].fieldName` — 数组钻取
- `globalVars.appVarName` — 全局变量

### 4.3 arrayDrillDown

当字段映射穿越数组类型时，自动启用数组钻取模式：

- 识别 `items[*]` 或 list 类型字段
- 为每个数组元素创建独立的映射路径

## 5. 节点执行引擎

Path: `domain/workflow/internal/compose/node_runner.go`

### 5.1 节点运行配置

```go
type nodeRunConfig[O any] struct {
    nodeKey, nodeName, nodeType    // 节点标识
    timeoutMS, maxRetry           // 超时和重试
    errProcessType                // 错误处理策略
    dataOnErr                     // 错误时默认数据
    preProcessors                 // 输入预处理
    postProcessors                // 输出后处理
    streamPreProcessors           // 流式预处理
    callbackInputConverter        // 回调输入转换
    callbackOutputConverter       // 回调输出转换
    init                          // 初始化函数
    i, s, c, t                    // Invoke/Stream/Collect/Transform
}
```

### 5.2 节点生命周期

```
init → preProcess → onStart → invoke/stream/collect/transform → postProcess → onEnd
                                                              ↘ onError
```

每个阶段说明：

1. **init** — 初始化上下文缓存、检查最大节点数
2. **preProcess** — 类型转换、填充零值、去除完成标记
3. **onStart** — 发射回调开始事件
4. **invoke/stream/collect/transform** — 执行实际节点逻辑（含重试循环）
5. **postProcess** — 填充缺失输出字段的 nil 值
6. **onEnd** — 发射回调结束事件
7. **onError** — 根据 ErrorProcessType 处理：
   - **Throw**: 传播错误
   - **ReturnDefaultData**: 返回配置的默认数据
   - **ExceptionBranch**: 路由到异常分支

### 5.3 重试机制

每个执行方法（invoke/stream/collect/transform）内部有重试循环：

- 最多重试 `maxRetry` 次
- 遇到 `InterruptRerunError` 立即中断不重试

## 6. 执行上下文与回调

Path: `domain/workflow/internal/execute/`

### 6.1 Context 结构

```go
type Context struct {
    RootCtx          // 根工作流信息、执行 ID、恢复事件、执行配置
    SubWorkflowCtx   // 子工作流信息
    NodeCtx          // 当前节点 key、执行 ID、类型、路径、重试次数
    BatchInfo        // 批处理迭代信息
    TokenCollector   // LLM token 使用追踪
    AppVarStore      // 共享应用变量（线程安全）
    CheckPointID     // 检查点标识
}
```

### 6.2 回调系统

实现 Eino 的 `callbacks.Handler` 接口：

| Handler | 事件 |
|---------|------|
| WorkflowHandler | WorkflowStart/Resume/Success/Failed/Cancel/Interrupt |
| NodeHandler | NodeStart/End/Error/StreamingInput/StreamingOutput/NodeEndStreaming |
| ToolHandler | FunctionCall/ToolResponse/ToolStreamingResponse/ToolError |

事件通过 channel 发送到中央事件处理器，由 `HandleExecuteEvent` 消费。

### 6.3 事件类型

```go
// 工作流事件
WorkflowStart, WorkflowResume, WorkflowSuccess,
WorkflowFailed, WorkflowCancel, WorkflowInterrupt

// 节点事件
NodeStart, NodeEnd, NodeError,
NodeStreamingInput, NodeStreamingOutput, NodeEndStreaming

// 工具事件
FunctionCall, ToolResponse, ToolStreamingResponse, ToolError
```

## 7. WorkflowRunner 编排器

Path: `domain/workflow/internal/compose/workflow_run.go`

WorkflowRunner 是执行生命周期的核心编排器：

```go
type WorkflowRunner struct {
    basic          *entity.WorkflowBasic
    input          string
    resumeReq      *entity.ResumeRequest
    schema         *schema2.WorkflowSchema
    sw             *schema.StreamWriter[*entity.Message]
    container      *execute.StreamContainer
    config         model.ExecuteConfig
    executeID      int64
    eventChan      chan *execute.Event
    interruptEvent *entity.InterruptEvent
}
```

### 7.1 Prepare 方法（核心编排）

```
Prepare(ctx)
  1. 生成或复用执行 ID
  2. 恢复场景：获取中断事件，验证事件 ID
  3. 构建 compose options（检查点、全局状态、回调）
  4. 恢复场景：构建状态修改器（嵌套路径解析）
  5. 新执行：创建 WorkflowExecution DB 记录
  6. 设置上下文超时（前台 vs 后台）
  7. 启动 goroutine: execute.HandleExecuteEvent 消费事件
  8. 返回 cancelCtx, executeID, composeOpts, lastEventChan
```

## 8. 检查点与恢复

检查点机制支持工作流在中断点暂停并恢复：

- **Checkpoint Store**: 支持内存和 Redis 两种实现
- **中断触发**: QuestionAnswer、InputReceiver、Batch、Loop 节点可触发中断
- **恢复数据**: 中断时持久化到 checkpoint，恢复时重建上下文
- **状态修改器**: 恢复时通过 state modifier 将用户输入注入到中断节点

检查点由 Schema 自动判断是否需要（`RequireCheckpoint`），以下情况需要：

- 包含 QuestionAnswer 节点
- 包含 InputReceiver 节点
- 包含 Loop/Batch 复合节点
