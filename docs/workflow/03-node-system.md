# 节点系统详解

## 1. 节点执行接口

Path: `domain/workflow/internal/nodes/node.go`

节点通过四种执行范式定义其行为：

| 接口 | 输入 | 输出 | 说明 |
|------|------|------|------|
| InvokableNode / InvokableNodeWOpt | 非流式 | 非流式 | 标准同步调用 |
| StreamableNode / StreamableNodeWOpt | 非流式 | 流式 | LLM 生成等 |
| CollectableNode / CollectableNodeWOpt | 流式 | 非流式 | 汇总流式数据 |
| TransformableNode / TransformableNodeWOpt | 流式 | 流式 | 流式转换 |

WOpt 版本接收 `NodeOption` 参数，可访问执行上下文。

辅助接口：
- **Initializer**: 每次调用前初始化上下文
- **NodeAdaptor**: 前端 Canvas Node → 后端 NodeSchema 转换
- **BranchAdaptor**: 分支条件转换
- **StreamGenerator**: 流类型推断
- **CallbackInputConverted / CallbackOutputConverted**: UI 友好的执行追踪转换

## 2. 节点类型全景表

Path: `domain/workflow/entity/node_meta.go`

总计 32 种节点类型 + 1 个内部节点（Lambda）：

### 2.1 输入 & 输出

| ID | 类型 | 名称 | 特性 |
|----|------|------|------|
| 1 | Entry | Start | PostFillNil, 默认值填充 |
| 2 | Exit | End | IncrementalOutput, InputSourceAware |
| 13 | OutputEmitter | Message | IncrementalOutput, BlockEndStream |
| 30 | InputReceiver | Input | 中间用户输入 |

### 2.2 核心节点

| ID | 类型 | 名称 | 特性 |
|----|------|------|------|
| 3 | LLM | LLM | MayUseChatModel, UseCtxCache, SupportBatch |
| 4 | Plugin | Api | UsePlugin, SupportBatch |
| 5 | CodeRunner | Code | UseCtxCache |
| 9 | SubWorkflow | SubWorkflow | BlockEndStream, SupportBatch |

### 2.3 逻辑控制

| ID | 类型 | 名称 | 特性 |
|----|------|------|------|
| 8 | Selector | If | 条件分支 |
| 19 | Break | Break | 循环终止 |
| 20 | VariableAssignerWithinLoop | LoopSetVariable | 循环内变量赋值 |
| 21 | Loop | Loop | IsComposite, PersistInputOnInterrupt |
| 28 | Batch | Batch | IsComposite, PersistInputOnInterrupt |
| 29 | Continue | Continue | 循环继续 |
| 22 | IntentDetector | Intent | MayUseChatModel, UseCtxCache |
| 32 | VariableAggregator | VariableMerge | InputSourceAware, IncrementalOutput |

### 2.4 知识库 & 数据

| ID | 类型 | 名称 | 特性 |
|----|------|------|------|
| 6 | KnowledgeRetriever | Dataset | UseKnowledge, UseCtxCache |
| 27 | KnowledgeIndexer | DatasetWrite | UseKnowledge |
| 40 | VariableAssigner | AssignVariable | 变量赋值 |
| 60 | KnowledgeDeleter | KnowledgeDelete | UseKnowledge |

### 2.5 数据库

| ID | 类型 | 名称 | 特性 |
|----|------|------|------|
| 12 | DatabaseCustomSQL | Database | UseDatabase |
| 42 | DatabaseUpdate | DatabaseUpdate | UseDatabase |
| 43 | DatabaseQuery | DatabaseSelect | UseDatabase |
| 44 | DatabaseDelete | DatabaseDelete | UseDatabase |
| 46 | DatabaseInsert | DatabaseInsert | UseDatabase |

### 2.6 组件工具

| ID | 类型 | 名称 | 特性 |
|----|------|------|------|
| 15 | TextProcessor | Text | InputSourceAware |
| 18 | QuestionAnswer | Question | MayUseChatModel, PersistInputOnInterrupt |
| 45 | HTTPRequester | Http | HTTP API 调用 |
| 58 | JsonSerialization | ToJSON | JSON 序列化 |
| 59 | JsonDeserialization | FromJSON | JSON 反序列化, UseCtxCache |

### 2.7 会话管理

| ID | 类型 | 名称 | 分类 |
|----|------|------|------|
| 37 | MessageList | - | message |
| 38 | ClearConversationHistory | - | conversation_history |
| 39 | CreateConversation | - | conversation_management |
| 51 | ConversationUpdate | - | conversation_management |
| 52 | ConversationDelete | - | conversation_management |
| 53 | ConversationList | - | conversation_management |
| 54 | ConversationHistory | - | conversation_history |
| 55 | CreateMessage | - | message |
| 56 | EditMessage | - | message |
| 57 | DeleteMessage | - | message |

### 2.8 其他

| ID | 类型 | 名称 | 说明 |
|----|------|------|------|
| 31 | Comment | Comment | 注释（不可执行） |
| 1000 | Lambda | - | 内部使用 |

## 3. ExecutableMeta 节点特性详解

每种节点类型通过 `ExecutableMeta` 声明其运行时特性：

| 特性 | 含义 | 使用场景 |
|------|------|----------|
| IsComposite | 复合节点，包含内部子工作流 | Loop, Batch |
| DefaultTimeoutMS | 默认超时（毫秒），0 不限 | 按需配置 |
| PreFillZero | 执行前填充零值 | Entry |
| PostFillNil | 执行后缺失字段填 nil | Entry |
| MayUseChatModel | 可能使用 Chat Model | LLM, IntentDetector, QuestionAnswer |
| InputSourceAware | 需要知道输入源的运行时状态 | Exit, TextProcessor, VariableAggregator |
| IncrementalOutput | 输出为用户面向的增量流式 | Exit, OutputEmitter, VariableAggregator |
| UseCtxCache | 每次调用初始化新 ctx cache | LLM, CodeRunner, KnowledgeRetriever, IntentDetector, JsonDeserialization |
| PersistInputOnInterrupt | 中断时持久化输入，恢复时还原 | QuestionAnswer, InputReceiver, Loop, Batch |
| BlockEndStream | 阻塞直到所有流式 chunk 接收完 | SubWorkflow, OutputEmitter |
| UseDatabase | 需要数据库配置 | 所有 Database* 节点 |
| UseKnowledge | 需要知识库配置 | KnowledgeRetriever, KnowledgeIndexer, KnowledgeDeleter |
| UsePlugin | 需要插件配置 | Plugin |

## 4. 核心节点实现详解

### 4.1 Entry（开始节点）

Path: `domain/workflow/internal/nodes/entry/entry.go`

- 实现 `InvokableNode`
- 合并默认值与提供的输入
- 空值使用默认值填充
- 定义工作流的输入参数 Schema

### 4.2 Exit（结束节点）

Path: `domain/workflow/internal/nodes/exit/exit.go`

- 实现 `CallbackOutputConverted`
- 如果 `TerminatePlan == ReturnVariables`：透传输入作为输出
- 否则：委托给 `OutputEmitter` 进行模板化流式输出
- 定义工作流的输出参数 Schema

### 4.3 LLM 节点

- 实现 `StreamableNodeWOpt`（非流式输入，流式输出）
- 使用 `modelbuilder.BaseChatModel` 调用 LLM
- 支持变量替换和提示词模板
- 支持 Function Calling
- Token 使用通过回调系统追踪
- 支持批量执行

### 4.4 Plugin（插件/API 节点）

- 实现 `InvokableNodeWOpt`
- 调用外部 API/工具
- 支持 OAuth 认证
- 支持批量执行

### 4.5 CodeRunner（代码执行节点）

- 实现 `InvokableNode`
- 支持 Python / JavaScript 代码执行
- 三种执行模式：直接执行、沙箱执行、Python 脚本
- 使用 ctx cache 记录执行警告

### 4.6 Selector（条件分支节点）

- 实现 `InvokableNode`
- 支持多条分支条件
- 条件类型：OR / AND
- 操作符：Equal, NotEqual, Contains, StartsWith, EndsWith, GreaterThan, LessThan, IsEmpty, In, MatchRegex 等 16 种

### 4.7 Loop（循环节环节点）

- 复合节点（IsComposite），包含内部子工作流
- 三种循环类型：array（数组遍历）、count（计数循环）、infinite（无限循环）
- 支持 Break / Continue 控制流
- 每次迭代执行内部子工作流
- 中断时持久化输入数组

### 4.8 Batch（批处理节点）

- 复合节点，类似 Loop 但针对批量数据处理
- 将数组输入分割为批次
- 每批并发执行内部子工作流
- 汇总所有批次结果

### 4.9 SubWorkflow（子工作流节点）

- 创建独立的执行记录（root_execution_id 指向父执行）
- 支持参数传递和结果返回
- BlockEndStream：等待子工作流流式输出完成
- 支持批量执行

### 4.10 KnowledgeRetriever（知识检索节点）

- 实现 `InvokableNode`
- 从 Elasticsearch 检索知识库文档
- 支持 Rerank（RRF）
- 查询转换：Messages2Query

### 4.11 Database 系列节点

五种数据库节点均实现 `InvokableNode`：
- **DatabaseQuery**: SELECT 查询
- **DatabaseInsert**: INSERT 记录
- **DatabaseUpdate**: UPDATE 记录
- **DatabaseDelete**: DELETE 记录
- **DatabaseCustomSQL**: 自定义 SQL（使用 SQL Parser 验证安全性）

### 4.12 QuestionAnswer（问答节点）

- 实现中断机制
- 暂停工作流执行，等待用户回答
- 中断时将输入持久化到 checkpoint
- 恢复时重建上下文

### 4.13 TextProcessor（文本处理节点）

- 支持模板字符串处理
- 支持 concat / split 操作
- InputSourceAware：需要知道输入源运行时状态

### 4.14 HTTPRequester（HTTP 请求节点）

- 支持 GET/POST/PUT/DELETE 等方法
- 支持自定义 Headers、Body
- 支持认证配置

### 4.15 VariableAggregator（变量聚合节点）

- 聚合多条分支的输出
- 支持流式增量输出
- InputSourceAware

### 4.16 IntentDetector（意图识别节点）

- 使用 LLM 进行意图分类
- 预设意图类别
- MayUseChatModel

### 4.17 JsonSerialization / JsonDeserialization

- 序列化：对象 → JSON 字符串
- 反序列化：JSON 字符串 → 对象

### 4.18 会话管理节点

完整的会话 CRUD：
- CreateConversation / ConversationList / ConversationUpdate / ConversationDelete
- ConversationHistory / ClearConversationHistory
- CreateMessage / EditMessage / DeleteMessage / MessageList

## 5. 节点目录结构

```
domain/workflow/internal/nodes/
├── node.go                 # 接口定义
├── convert.go              # JSON 序列化/反序列化实现
├── interrupt.go            # 中断事件存储接口
├── entry/                  # Entry 节点
├── exit/                   # Exit 节点
├── llm/                    # LLM 节点
├── selector/               # Selector 条件分支
├── loop/                   # Loop / Break / Continue
├── batch/                  # Batch 批处理
├── code/                   # CodeRunner
├── qa/                     # QuestionAnswer
├── receiver/               # InputReceiver
├── emitter/                # OutputEmitter
├── textprocessor/          # TextProcessor
├── httprequester/          # HTTPRequester
├── intentdetector/         # IntentDetector
├── variableassigner/       # VariableAssigner / WithinLoop
├── variableaggregator/     # VariableAggregator
├── database/               # 5 种 Database 节点
├── knowledge/              # 3 种 Knowledge 节点
└── conversation/           # 10 种 Conversation/Message 节点
```

## 6. 节点类型转换

### 6.1 前端 → 后端

每个节点类型的 `NodeAdaptor` 将前端 Canvas JSON 转换为后端 `NodeSchema`：
- 解析节点配置（Inputs/Outputs）
- 构建 FieldMapping
- 设置执行参数（超时、重试、错误处理）

### 6.2 ID ↔ Type 映射

```go
NodeType("LLM") → ID 3
IDStr("3") → NodeType("LLM")
```

定义在 `entity/node_meta.go` 的 `NodeTypeMetas` 变量中，运行时 O(n) 查找。
