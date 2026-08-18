# CoreKG 核心业务流程技术文档

> 用户提问 - 后端处理 - 回答生成 完整技术流程

---

## 1. 系统架构概览

```
+---------------------------------------------------------------------+
|                            用户请求入口                              |
|  外部API: POST /keapi.chat/chat/completions (OpenAI兼容)              |
|  内部Web: /chat.ChatQuestionStream (流式问答)                         |
+-------------------------------+-------------------------------------+
                                |
                      +---------v---------+
                      |    API Handler     | 参数校验、鉴权、流式/非流式分发
                      +---------+---------+
                                |
                      +---------v---------+
                      |  Service Layer     | 会话管理、模型选择、消息解析、问题入库
                      +---------+---------+
                                |
                      +---------v---------+
                      |  ChatWrapper       | 根据 Session.BaseType 路由到对应模式
                      +---------+---------+
                                |
        +----------+-----+-----+----------+-----------+
        v          v     v                v           v
   ForestAgent  Forest  GraphSearch  DirectModel    Excel
   (ReAct+RAG) (RAG)   (NebulaGraph) (纯LLM)     (代码+图表)
        |          |        |            |           |
   +----v----------v--+     |            |           |
   | ForestSearchTool |     |            |           |
   |   (RAG管线)      |     |            |           |
   | +--------------+ |     |            |           |
   | |1.QA精确匹配   | |     |            |           |
   | |2.查询改写     | |     |            |           |
   | |3.向量生成     | |     |            |           |
   | |4.ES混合检索   | |     |            |           |
   | |5.Rerank#1    | |     |            |           |
   | |6.上下文扩展   | |     |            |           |
   | |7.Rerank#2    | |     |            |           |
   | |8.结果组装     | |     |            |           |
   | +--------------+ |     |            |           |
   +--------+---------+     |            |           |
            |               v            |           |
            |        NebulaGraph查询     |           |
            |        目录Agent筛选       |           |
            |        过滤Agent+分析Agent  |           |
            |               |            |           |
   +--------v---------------v------------v-----------v-+
   |           SummaryAgent / 直接返回                   |
   |         基于证据生成结构化答案                        |
   +------------------------+----------------------------+
                                |
                      +---------v---------+
                      |    响应返回        |
                      | JSON / SSE Stream  |
                      +--------------------+
```

### 存储依赖

| 存储 | 用途 | 访问方式 |
|------|------|----------|
| Elasticsearch | Chunk/QA索引、混合检索、向量搜索 | `elastic/go-elasticsearch/v8` |
| MySQL/MariaDB | 会话、问题、用户、知识库、文件等业务数据 | GORM |
| NebulaGraph | 知识图谱（汽车领域结构化数据） | 自定义 `NebulaCli` |
| Redis | SSE消息推送、分布式锁、缓存 | `redispool` |
| OSS/MinIO | 原始文档、解析后的Markdown、图片存储 | `fs.Forest` |

### 外部服务依赖

| 服务 | 用途 | 配置来源 |
|------|------|----------|
| LLM API (OpenAI兼容) | 推理、查询改写、摘要生成 | 数据库 `ChatModel` |
| Embedding API | 语义向量编码 | `knowledge.embedding` |
| Rerank API | 搜索结果重排序 | `knowledge.rerank` |

---

## 2. API入口层

### 2.1 外部API (keapi)

**文件**: `apps/keapi/internal/apis/forestctl/chat.go`

#### ChatCompletions 端点

```
POST /keapi.chat/chat/completions
Authorization: Bearer <api_key>
```

**请求体** (`dtokeapi.ChatCompletionsRequest`):

```json
{
  "session_id": 0,
  "forest_file_id": [101, 102],
  "messages": [
    {"role": "user", "content": "汽车发动机异响怎么办？"}
  ],
  "stream": true,
  "temperature": 0.2,
  "top_p": null,
  "presence_penalty": null,
  "extra_body": {
    "enable_reference": true
  }
}
```

**处理流程**:

1. 绑定并校验请求体 (`req.ValidChatCompletions()`)
2. 根据 `req.Stream` 分支:
   - **流式**: 创建 `OpenAIPrinter`，调用 `RunWithPrinter`
   - **非流式**: 使用 `NoopPrinter`，返回完整 JSON
3. 错误映射:
   - `ErrInvalidChatMessages` / `ErrInvalidForestFiles` -> HTTP 400
   - `ErrChatSessionNotFound` -> HTTP 404
   - `ErrChatModelNotFound` -> HTTP 400
   - 其他错误 -> HTTP 500
4. 响应格式完全兼容 OpenAI Chat Completions API

#### OpenAI 流式输出 (`OpenAIPrinter`)

**文件**: `apps/keapi/internal/services/svcforestchat/printer.go`

SSE 输出遵循标准 `data: {json}\n\n` 格式:
- 首个 chunk 包含 `role: "assistant"`
- 增量内容通过 `resolveDelta` 计算差异
- 最终 chunk 包含 `finish_reason: "stop"`
- 以 `data: [DONE]\n\n` 结束
- 响应头: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`

### 2.2 内部Web API (kechat)

**文件**: `apps/kechat/internal/apis/chat_api.go`

| 端点 | Handler | 说明 |
|------|---------|------|
| `chat.ChatQuestionStream` | `ChatQuestionStream` | 流式问答 v2 (最新) |
| `chat.SubmitChatQuestion` | `SubmitChatQuestion` | 同步问答 (v2) |
| `chat.SubmitChatQuestionStream` | `SubmitChatQuestionStream` | 旧版流式问答 |
| `chat.Agent/chat/completions` | `AgentChat` | Agent/机器人API |
| `chat.GetMessage` | `ChatGetMessage` | 恢复获取流式消息 |
| `chat.StopChat` | `StopChat` | 停止活跃的聊天流 |

**内部请求体** (`SubmitChatQuestionEmbedRequest`):

```json
{
  "session_id": 0,
  "model_id": 5,
  "base_type": "standard",
  "resource_type": "forest",
  "resource_ids": [101, 102],
  "resource_names": ["发动机手册.pdf"],
  "question": "汽车发动机异响怎么办？"
}
```

**校验规则**:
- `SessionID == 0` 时要求 `BaseType`、`ResourceType`、资源列表均有效
- 始终要求 `ModelID != 0` 且 `Question` 非空

---

## 3. Service层 - 会话与消息准备

**文件**: `apps/keapi/internal/services/svcforestchat/chat.go`

### 3.1 核心入口函数

```go
func RunWithPrinter(ctx *gin.Context, req *ChatCompletionsRequest, msgPrinter printer.Printer) (*Result, error)
```

### 3.2 会话规划 (`prepareChatSessionPlan`)

```
if SessionID > 0:
    -> prepareExistingChatSessionPlan()
      -> 加载已有会话 + 关联模型
      -> BaseType 固定为 ForestAgent

else (SessionID == 0):
    -> createOneShotSession()
      -> selectChatModel()              // 自动选择第一个可用模型
      -> resolveForestFileScope(fileIDs) // 校验文件、解析forest_id和ES索引
      -> 创建临时会话 + 返回cleanup函数
```

### 3.3 消息解析 (`prepareChatMessages`)

根据 `SeedHistory` 标志分两种解析模式:
- **有历史种子**: `buildChatInput()` - 提取历史 user/assistant 对 + 当前问题 + 图片 + 系统提示
- **无历史种子**: `buildCurrentChatInput()` - 仅当前消息 + 图片

消息校验规则: 要求 user/assistant 交替出现，拒绝连续 user 消息。

### 3.4 文件作用域解析 (`resolveForestFileScope`)

- 校验所有文件存在、非目录、`KnowledgeStatus == Success`
- 提取所有文件的 `ForestID` 和 `EsIndex`
- 确保所有文件属于同一 ES 索引

### 3.5 问题入库

在调用 ChatWrapper 之前，创建一条 `ChatQuestion` 记录 (状态 `Pending`)，后续由 ChatWrapper 更新答案和引用。

### 3.6 结果处理

- 更新问题记录: 答案、Token用量、状态、引用列表、Agent服务标识
- 异步更新会话名称: 通过LLM生成会话摘要名称 (45s超时，goroutine + panic recovery)
- Token用量: 优先使用实际值，回退到 `字符数/4` 估算 (最少1 token)

---

## 4. ChatWrapper - 模式路由

**文件**: `apps/kechat/chat/wrapper/wrapper.go`

根据 `Session.BaseType` 分发到对应的 ChatMode:

| BaseType | Mode | 说明 |
|----------|------|------|
| `forest_agent` | `ForestAgentChatMode` | 默认模式，ReAct Agent + RAG 搜索工具 |
| `standard` | `ForestChatMode` | 关键词改写 + RAG 搜索 + ReAct Agent |
| `graph_search` | `GraphSearchChatMode` | 知识图谱 RAG (NebulaGraph) |
| `model` | `DirectModelChatMode` | 纯 LLM 对话 (无 RAG) |
| `react_excel` | `ExcelChatMode` | Excel/数据表分析 (代码执行 + 图表) |

执行完成后，将 `ChatResult` 的字段映射回数据库 `question.Source` 记录:
- Answer, Reasoning, ReasoningSeconds, CostSeconds
- OutToken, CacheHitToken, CacheMissToken, TotalTokens
- Status, AgentService, QueryReferences, ChatReferences

---

## 5. ChatMode 详解

### 5.1 ForestAgent 模式 (默认)

**文件**: `apps/kechat/chat/modes/forest_agent.go`

这是外部API和大多数内部对话的默认模式，采用 **ReAct Agent + RAG** 架构。

#### 执行流程

```
1. getForestChatHistory()          -> 加载历史对话
2. NewForestSearchTool()           -> 创建RAG搜索工具 (绑定共享引用池)
3. buildInputFiles(fileIDs)        -> 加载并解析输入文件
4. newToolCallingChatModel()       -> 创建LLM客户端
5. 构建 AgentContext:
   - MaxStep = 6
   - SystemPrompt = ForestAgentSystemPrompt
   - NextStepPrompt = ForestAgentNextStepPrompt
   - SummaryPrompt = BuildForestAgentSummarySystemPrompt()
   - Tools: ToolOptionCode, ToolOptionFile
   - AvailableTools: [forestSearchTool]
6. ReactAgentService.Handler()     -> 执行ReAct推理循环 + 摘要生成
7. 聚合引用 (按ChunkID去重, 按Sequence排序)
8. 映射统计信息到ChatResult
```

#### 输入文件处理

- 优先使用解析后的 Markdown (`file.ParsedMarkdownURL`)
- xlsx/xls/csv 文件使用原始预览文件而非解析后的Markdown
- 回退到文件预览URL (`fs.PreviewFile`)

#### System Prompt 核心原则

1. **证据优先**: 禁止猜测、编造数据或来源
2. **复用已有信息**: 优先利用已获取的信息，避免冗余搜索
3. **最小必要行动**: 最多9轮执行
4. **停止条件**: 信息充足 / 证据已获取 / 无更多来源
5. **状态输出**: `[状态] <阶段>\n[下一步] <操作>` (每项最多30字符)
6. Agent 只负责收集证据，**不生成最终答案**

#### Summary Prompt 设计

最终答案由 SummaryAgent 独立生成，遵循:
- 直接回答优先，按问题类型组织 (事实查询/总结/对比/数据分析/步骤/多对象/风险建议)
- 引用格式: `{Reference §fileID[sequence]}`
- 图片规则: 仅来自检索结果，每节1-3张
- 回退场景: A(模糊但可答) / B(结果有限) / C(无法回答)
- 禁止: "需要更多细节吗" / "基于..." / 内部过程泄露

### 5.2 Forest 模式 (Standard)

**文件**: `apps/kechat/chat/modes/forest.go`

与 ForestAgent 类似，但增加了**关键词改写**步骤。

#### 与 ForestAgent 的差异

| 特性 | ForestAgent | Forest |
|------|-------------|--------|
| MaxStep | 6 | 4 |
| 关键词改写 | 无 | 有 (同义词+行业术语替换) |
| Prompt | 固定模板 | 动态加载 (PromptMode) |
| 工具 | Code + File + Search | 仅 Search |

#### 关键词改写

在执行搜索前调用:
1. `devkeywords.ReplaceSynonymKeywords()` - 同义词替换
2. `devkeywords.ReplaceMajorKeywords()` - 行业术语替换

改写后的问题通过 SSE 发送给前端展示。

#### Prompt 风格选择

根据 `session.PromptMode` 动态选择 prompt 风格 (回退默认 `"normal"`)，支持 `[DYNAMIC_STYLE]` 占位符替换。

### 5.3 GraphSearch 模式

**文件**: `apps/kechat/chat/modes/graph_search.go`

基于 **NebulaGraph 知识图谱** 的结构化 RAG 模式，适用于汽车领域技术手册。

#### 执行流程

```
1. getForestChatHistory()           -> 加载历史
2. ExecuteCatalogueAgent(query)      -> 并行遍历3类目录树:
   - "诊断手册" / "维修手册" / "电路图"
   -> 返回相关目录节点ID列表
3. NewFilterAgent(files) + ADK Run  -> 基于文件描述+内容过滤
   -> 返回匹配的file_id列表 (最多10个)
4. 并行执行:
   a. getGraph(fileID, analystInput) -> NebulaGraph 3跳子图查询
      -> 写入 GraphReference + GraphChatReference
   b. NewAnalystAgent(fileData)      -> 流式分析报告生成
5. 拼接所有流式消息块
6. 等待图谱查询完成
7. 返回 ChatResult (含图谱引用)
```

#### NebulaGraph 操作

```
# 目录遍历 (最大5跳)
GO 0 TO 5 STEPS FROM "{searchType}" OVER 包含 BIDIRECT YIELD DISTINCT dst(edge)

# 节点信息获取
MATCH (v:目录) WHERE id(v) IN ["title1","title2"] RETURN v;
```

- 空间: `a_car_test`
- 图谱查询深度: 3跳
- LLM Temperature: 0.4
- 超时: 300秒

#### Analyst Agent

- 输入: FilterAgent 筛选后的文件内容 + 分析摘要 (JSON序列化，截断至230,000字符)
- 输出: 严格Markdown格式的分析报告
- 规则: 仅使用提供的信息，信息不足时回复 "现有资料不足，无法回答"

### 5.4 DirectModel 模式

**文件**: `apps/kechat/chat/modes/direct_model_chat.go`

纯 LLM 对话模式，支持 Web 搜索和 ADK Agent 框架。

#### 执行流程

```
1. 加载历史，创建AgentRequest (可包含web_search选项)
2. 初始化SSE + Printer + BaseAgent (含FinalAnswerSignal)
3. 创建ChatModel，构建工具集:
   - 始终: FinalAnswerMarkerTool
   - 可选: WebSearchTool, FileAnalysisTool
4. 创建ADK ChatModelAgent:
   - Name: "atlas_agent"
   - BeforeChatModel中间件: 每轮动态替换系统提示
5. 运行Agent循环 (max 20轮)
6. 如果未生成最终输出或达到最大步数:
   -> 创建无工具Agent重新生成 (fallback summary)
7. 返回最后一条assistant消息
```

- 默认MaxIterations: 20
- 默认Temperature: 0.4
- LLM超时: 300秒

### 5.5 Excel 模式

**文件**: `apps/kechat/chat/modes/excel.go`

Excel/数据表分析模式，支持代码执行和图表生成。

#### 执行流程

```
1. 从 session.ExcelIDList 加载已解析的Excel文件
2. 构建输入文件 (预览URL)
3. 初始化SSE + Printer + ChatModel
4. 构建 AgentContext:
   - Tools: ToolOptionCode, ToolOptionFile, ToolOptionChart
   - SystemPrompt: excelprompt.SystemPrompt
   - SaveChartFunc: 图表持久化到数据库
5. ReactAgentService.Handler() -> ReAct推理 + 摘要
6. 返回 ChatResult
```

与 ForestAgent 不同: 不包含 `forestSearchTool`，但包含 `ToolOptionChart` + 自定义图表保存函数。

---

## 6. RAG 搜索管线 (核心)

### 6.1 ForestSearchTool - Agent工具接口

**文件**: `apps/kesearch/pkg/ai/tools/forest_search_tool.go`

ReAct Agent 通过调用此工具执行知识检索。工具内部维护线程安全的共享引用池 (`SharedReferences`)。

#### 工具输入

```json
{
  "question": "用户的搜索问题",
  "search_strategy": "common_questions"
}
```

`search_strategy` 可选值:
- `common_questions` (默认): 混合文本+向量搜索 + Rerank
- `knowledge_summary`: 文件级聚合搜索

#### 执行优先级

```
1. 若 ForestIDs 和 FileIDs 均为空 -> 返回空结果
2. QA精确匹配: FindFQAByQuestion()
   -> 命中预置问答对 -> 直接返回 (短路)
3. knowledge_summary策略: SearchDescription()
   -> 搜索文件级描述/摘要
   -> 结果 > 150条时并行LLM摘要
4. 正常检索: RerankSearchQuestionChunk()
   -> 完整8步RAG管线
5. 所有结果追加到 SharedReferences
```

#### SharedReferences (线程安全引用聚合)

- `Append(refs...)`: 写锁追加
- `Get()`: 读锁返回原始列表
- `GetAggregated()`: 读锁 -> 按ChunkID去重 -> 按Sequence排序 -> 按FileID聚合

### 6.2 Rerank 搜索管线 (8步)

**文件**: `apps/kesearch/models/reranksearch/search.go`, `essearch.go`

#### 初始化 (`NewRerankSearchWrapper`)

```
1. InitESClient()                  -> 连接Elasticsearch
2. UserQueryRewrite(question)      -> LLM改写用户查询
3. GetEmbedding(rewrittenQuery)    -> 调用外部Embedding API获取向量
4. GetDefaultConfig()              -> 加载搜索配置
```

#### 查询改写 (`UserQueryRewrite`)

**文件**: `apps/kesearch/models/reranksearch/rerank.go`

通过内部 Agent (`ChatAgentUserQueryRewrite`) 非流式调用改写用户问题，支持国际化 Agent 名称。

#### Embedding 生成 (`GetEmbedding`)

**文件**: `apps/kesearch/models/essearch/embedding.go`

```
POST {embedding.url}
Authorization: Bearer {embedding.key}
Body: {"model": "{model_name}", "input": "查询文本"}
```

返回第一个 embedding 向量。

#### 8步管线详情

```
Step 1: SearchQuestionChunk (ES混合检索)
  Filter: is_disable != true, 有embedding字段
          type in {chunk, image, table, video, formula}
          forest_id/file_id terms过滤
  Should: multi_match on description^DescriptionWeight
  Scoring: script_score = BM25_score + cosineSimilarity * EmbeddingWeight
           image/video类型权重0.65
  Size: Topn * FetchFactor (默认 30 * 2 = 60)
  Excludes: embedding, references

Step 2: SortRerankChunk #1 (首轮重排序)
  调用外部Rerank API: POST {rerank.url}
    Body: {model, text_1: question, text_2: [docs]}
  按RerankScore降序排序
  过滤: score >= RerankThreshold (默认0.5)
  截取TopN (默认30)
  Fallback: 若全部被过滤且FallBackToTopK=true
    -> 按原始分数取TopK (默认5)

Step 3: SearchChunkSequence (邻域检索)
  对每个chunk取前后NeighborSize (默认1) 个相邻chunk
  基于 file_id + sequence 邻近查询
  minimum_should_match: 1

Step 4: JoinNeighborChunks (上下文拼接)
  构建 fileID:sequence -> SearchType 映射
  左邻居(逆序拼接) + 中心chunk + 右邻居
  更新description为拼接后的完整文本
  sequence/location更新为最左邻居

Step 5: SortRerankChunk #2 (二轮重排序)
  对拼接后的chunk重新调用Rerank
  同Step 2的过滤和截断逻辑

Step 6: TopM截断
  若结果数 > Topm (默认15)，截取前Topm条

Step 7: GroupByFileID
  按FileID聚合chunk列表

Step 8: 结果组装 (Result)
  获取文件元数据 (名称, UIN, 创建时间)
  获取用户元数据 (姓名, 头像)
  可选: SearchFilesAbstract + RerankAbstract (EnabelAbstract=true时)
    -> 搜索文件描述，Rerank后附加
  设置 DataSourceType = "DC"
  组装最终 QueryReferenceList
```

### 6.3 搜索配置

**文件**: `apps/kesearch/models/reranksearch/type.go`

#### 标准配置 (默认)

| 参数 | 默认值 | 说明 |
|------|--------|------|
| DescriptionWeight | 0.3 | BM25文本分权重 |
| EmbeddingWeight | 0.7 | 向量相似度权重 |
| EnabelAbstract | true | 是否检索文件摘要 |
| EnableRerank | true | 是否启用重排序 |
| Topn | 30 | 每轮Rerank保留数 |
| Topm | 15 | 最终最大结果数 |
| Topk | 5 | Rerank全过滤时的回退数 |
| NeighborSize | 1 | 邻域扩展窗口 (每侧) |
| RerankThreshold | 0.5 | Rerank最低分数阈值 |
| FetchFactor | 2 | 初始检索倍数 (Topn * FetchFactor) |
| FallBackToTopK | true | 是否启用TopK回退 |

#### GraphSearch专用配置

| 参数 | 值 |
|------|-----|
| Topn | 100 |
| Topm | 20 |
| Topk | 50 |
| RerankThreshold | 0.4 |
| FetchFactor | 3 |

#### 配置验证

- DescriptionWeight + EmbeddingWeight == 1
- 所有值为正数
- RerankThreshold 在 0~1 之间

### 6.4 ES索引设计

- 索引模式: `ke_0`, `ke_1`, ... (按公司/知识库隔离)
- Chunk类型: `chunk`, `image`, `table`, `video`, `formula`, `file_description`
- 默认结果大小: 50 (`EsResultSize`)
- 中文分词器: `ik_max_word`

---

## 7. Agent 框架 (Eino)

### 7.1 ReAct Agent Service

**文件**: `pkgs/einotools/service/react_agent_service.go`

Agent编排核心，协调 ReAct 推理循环和 Summary 摘要生成。

#### Handler 流程

```
1. 解析角色名称 (从settings加载，有默认值)
2. 解析HandlerOptions (SummaryMode, ReactStream, SummaryStream, Debug)
3. NewReactAgent()  -> 构建Eino ReAct Agent
4. NewSummaryAgent() -> 构建摘要Agent
5. reactAgent.Run(query, options) -> 执行推理循环
6. 错误处理:
   - ContextCanceled -> 返回空字符串, nil error
   - ErrExceedMaxSteps -> 强制进入Summary模式
   - 其他错误 -> 返回error
7. 若SummaryMode启用: summaryAgent.RunSummarizeResult(historyMsgs)
8. 合并两个Agent的统计信息 -> TotalUsage, DurationMs
```

#### HandlerOptions

| 选项 | 说明 |
|------|------|
| SummaryMode | 是否在ReAct后运行Summary |
| ReactStream | 覆盖ReAct阶段的流式设置 |
| SummaryStream | 覆盖Summary阶段的流式设置 |
| Debug | 调试模式 |

### 7.2 ReAct Agent

**文件**: `pkgs/einotools/agent/react_agent.go`

基于 CloudWeGo Eino 框架的 ReAct (Reasoning + Acting) Agent。

#### 构建参数

- SystemPrompt: 默认 `ReactSystemPrompt`
- NextStepPrompt: 默认 `ReactNextStepPrompt`
- 模板变量: `roleName`, `date`, `files` (JSON), `history_dialogue`, `query`
- 工具集: 基础工具 + `agentContext.AvailableTools`
- UnknownToolsHandler: 返回友好错误提示
- MessageModifier: 前置系统提示 + 条件追加下一步提示
- DefaultMaxStep: 20 (各Mode可覆盖)
- StreamToolCallChecker: 读取完整流检测工具调用

#### 执行

- 支持流式 (`executor.Stream`) 和非流式 (`executor.Generate`)
- 通过 `compose.WithCallbacks + DesignateNode` 注册ChatModel和Tool节点回调
- 返回最后一条消息内容

### 7.3 Summary Agent

**文件**: `pkgs/einotools/agent/summary_agent.go`

基于Eino Graph的摘要生成Agent，固定2步 (`MaxStep = 2`)。

#### Graph结构

```
START -> prepare_input -> summarizer_model -> END
```

- `prepare_input`: 传递输入消息 (无内容时创建空user消息)
- `summarizer_model`: ChatModel节点
  - StatePreHandler 格式化prompt模板:
    - `{{.date}}` - 当前日期
    - `{{.roleName}}` - 角色名称
    - `{{.query}}` - 用户原始问题
    - `{{.taskHistory}}` - ReAct推理过程 (role/content对)
    - `{{.history_dialogue}}` - 历史对话
- 触发模式: `AnyPredecessor`

---

## 8. LLM 客户端

**文件**: `apps/kechat/chat/modelhelper/tool_calling_chat_model.go`

#### 创建流程

```
NewToolCallingChatModel(ctx, chatModel, options)
  BaseURL = chatModel.ModelUrl (去除 /chat/completions 后缀)
  Timeout = options.Timeout (默认300秒)
  Temperature, TopP, PresencePenalty -> 透传
  ResponseFormat -> 透传
  私有部署特殊处理:
    DeployMode == OnPremise 且 模型名匹配 DisableThinkingModelKeywords
    -> ExtraFields["chat_template_kwargs"]["enable_thinking"] = false
```

通过 `openai.NewChatModel` 创建 OpenAI 兼容客户端。

---

## 9. 完整请求链路 (外部API示例)

以 `POST /keapi.chat/chat/completions` 流式请求为例的完整调用链:

```
[HTTP] gin.Context
  -> forestctl.ChatCompletions()
      |-> req.ValidChatCompletions()
      |-> NewOpenAIPrinter() / NoopPrinter
      -> svcforestchat.RunWithPrinter()
          |-> prepareChatSessionPlan()
          |    |-> selectChatModel()
          |    |-> resolveForestFileScope()
          |    -> createOneShotSession() / loadExisting
          |-> prepareChatMessages()
          |    -> buildChatInput() / buildCurrentChatInput()
          |-> createCurrentQuestion() -> DB写入(Pending)
          -> chatwrapper.NewChatWrapper().Run()
               -> (以ForestAgent为例)
                    |-> getForestChatHistory()
                    |-> NewForestSearchTool()
                    |-> buildInputFiles()
                    |-> newToolCallingChatModel()
                    |-> ReactAgentService.Handler()
                    |    |-> NewReactAgent().Run()
                    |    |    -> [迭代推理循环, MaxStep=6]
                    |    |         |-> LLM推理 -> 决定调用工具
                    |    |         -> forestSearchTool.invoke()
                    |    |              |-> FindFQAByQuestion() [QA匹配]
                    |    |              -> RerankSearchQuestionChunk() [8步管线]
                    |    |                   |-> EsQueryRewrite() [LLM改写]
                    |    |                   |-> GetEmbedding() [向量生成]
                    |    |                   |-> SearchQuestionChunk() [ES混合检索]
                    |    |                   |-> SortRerankChunk() #1
                    |    |                   |-> SearchChunkSequence() [邻域]
                    |    |                   |-> JoinNeighborChunks() [拼接]
                    |    |                   |-> SortRerankChunk() #2
                    |    |                   |-> TopM截断
                    |    |                   |-> GroupByFileID()
                    |    |                   -> Result() [组装]
                    |    -> NewSummaryAgent().RunSummarizeResult()
                    |         -> 基于证据生成最终答案
                    |-> AggregateReferences() [去重+排序]
                    -> finalizeChatQuestion() [更新DB]
                         -> updateSessionNameWithLLMAsync() [goroutine]
```

---

## 10. 关键设计决策

### 10.1 ReAct vs 单次检索

系统采用 ReAct Agent 而非传统的单次检索+生成模式:
- Agent 可多轮调用搜索工具，逐步收集证据
- 根据中间结果动态调整搜索策略
- 避免一次性检索不足或检索过度

### 10.2 双轮 Rerank

先筛后扩再筛的策略:
- 第1轮: 从大量候选中筛选高质量chunk
- 邻域扩展: 补充上下文信息
- 第2轮: 对扩展后的内容重新评估相关性
- 平衡召回率和精确率

### 10.3 QA 短路机制

预置问答优先匹配，命中则跳过完整RAG管线:
- 降低延迟
- 保证高频问题的答案一致性
- 作为知识兜底

### 10.4 私有部署兼容

`OnPremise` 模式下自动检测并关闭不支持的 thinking 功能，通过 `ExtraFields` 传递模型特定参数。

### 10.5 流式输出设计

- 统一的 `printer.Printer` 接口
- `OpenAIPrinter` 实现增量差异计算
- 消息通过 Redis 中转，支持断线恢复 (`GetMessage`)
- SSE 过期时间: 5分钟

---

## 11. 关键文件索引

| 层级 | 文件路径 | 职责 |
|------|----------|------|
| API路由 | `apps/keapi/internal/apis/apis.go` | 外部API路由注册 |
| API路由 | `apps/kechat/internal/apis/apis.go` | 内部API路由注册 |
| Handler | `apps/keapi/internal/apis/forestctl/chat.go` | ChatCompletions端点 |
| Handler | `apps/kechat/internal/apis/chat_api.go` | 内部聊天端点 |
| DTO | `apps/keapi/internal/dto/dtokeapi/chat.go` | 外部API请求/响应 |
| DTO | `apps/kechat/internal/apis/chat_biz.go` | 内部请求/响应 |
| Service | `apps/keapi/internal/services/svcforestchat/chat.go` | 会话/消息编排 |
| Service | `apps/keapi/internal/services/svcforestchat/printer.go` | OpenAI流式输出 |
| Service | `apps/keapi/internal/services/svcforestchat/response_filter.go` | 答案过滤 |
| Mode | `apps/kechat/chat/wrapper/wrapper.go` | 模式路由 |
| Mode | `apps/kechat/chat/modes/forest_agent.go` | ForestAgent模式 |
| Mode | `apps/kechat/chat/modes/forest.go` | Forest模式 |
| Mode | `apps/kechat/chat/modes/graph_search.go` | GraphSearch模式 |
| Mode | `apps/kechat/chat/modes/direct_model_chat.go` | DirectModel模式 |
| Mode | `apps/kechat/chat/modes/excel.go` | Excel模式 |
| Mode | `apps/kechat/chat/modes/mode_common.go` | 共享函数 |
| Prompt | `apps/kechat/chat/prompt/forest_agent/system_prompt.go` | ReAct系统提示 |
| Prompt | `apps/kechat/chat/prompt/forest_agent/summary_prompt.go` | 摘要提示 |
| Agent | `pkgs/einotools/service/react_agent_service.go` | ReAct编排 |
| Agent | `pkgs/einotools/agent/react_agent.go` | ReAct实现 |
| Agent | `pkgs/einotools/agent/summary_agent.go` | Summary实现 |
| Tool | `apps/kesearch/pkg/ai/tools/forest_search_tool.go` | RAG搜索工具 |
| Search | `apps/kesearch/models/reranksearch/search.go` | 8步管线 |
| Search | `apps/kesearch/models/reranksearch/essearch.go` | ES操作 |
| Search | `apps/kesearch/models/reranksearch/wrapper.go` | 初始化(改写+向量) |
| Search | `apps/kesearch/models/reranksearch/type.go` | 搜索配置 |
| Search | `apps/kesearch/models/reranksearch/rerank.go` | Rerank API + 查询改写 |
| Search | `apps/kesearch/models/essearch/embedding.go` | Embedding API |
| Search | `apps/kesearch/models/essearch/search.go` | ES原始搜索 |
| Search | `apps/kesearch/services/svcessearch/search_question_chunk.go` | 搜索服务层 |
| GraphRAG | `apps/kecore/services/graphragsearch/agent/catalogue.go` | 目录Agent |
| GraphRAG | `apps/kecore/services/graphragsearch/agent/filteragent.go` | 过滤Agent |
| GraphRAG | `apps/kecore/services/graphragsearch/agent/mdanalyst.go` | 分析Agent |
| GraphRAG | `apps/kecore/services/graphragsearch/search/graphsearch.go` | NebulaGraph查询 |
| LLM | `apps/kechat/chat/modelhelper/tool_calling_chat_model.go` | LLM客户端 |
