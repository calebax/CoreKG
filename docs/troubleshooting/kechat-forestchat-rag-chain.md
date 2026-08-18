# Kechat ForestChat RAG 调用链与 Rerank 排查手册

本文是当前 MR 分支的 source of truth，用于定位 `/chat.NewChatQuestionStream`、`/chat.ChatQuestionStream`、ForestChat、ReAct function call、ES 召回、rerank、fallback、引用返回之间的关系。

## 入口结论

`NewChatQuestionStream` 不会执行 ES 检索，也不会调用模型回答问题。它只做参数校验、读取 session、创建 `ChatQuestion`，并返回 `question_id`。

真正生成回答的是 `ChatQuestionStream`：

```text
/chat.NewChatQuestionStream
  -> apis.NewChatQuestionStream
  -> chatquestion.CreateQuestion
  -> 返回 question_id

/chat.ChatQuestionStream
  -> apis.ChatQuestionStream
  -> svcchat.ChatQuestionStream
  -> agentwrapper.ChatWrapper.Run
  -> ForestChatMode.Run             // base_type=standard
  -> ReactAgentService.Handler
  -> ReActAgent.Run
  -> forest_search_tool.invoke      // 仅当模型产出 tool_calls
  -> RerankSearchQuestionChunk
  -> RerankSearchWrapper.RerankSearchChunk
  -> SharedReferences 聚合
  -> SummaryAgent.RunSummarizeResult
  -> ChatWrapper.updateQuestionSource
  -> defer chatquestion.UpdateQuestion 落库
```

旧接口 `SubmitChatQuestionStream` 仍存在，并且 `standard` 分支会走旧的 `qachat.ChatWapper.ForestChat()`。当前 v2 主链路是 `/chat.ChatQuestionStream`。

## ForestChatMode 做什么

`ForestChatMode.Run` 的职责是把一个普通知识库 session 包装成 ReAct 检索任务：

1. 读取已回答的历史 Q&A，作为 `history_dialogue`。
2. 创建 `forest_search_tool`，同时把 `OriginalQuestion` 固定为用户原始问题。
3. 创建支持 tool calling 的 chat model。
4. 对用户问题做前置关键词处理：`ReplaceSynonymKeywords` 和 `ReplaceMajorKeywords`。
5. 发送一条 `question_rewrite` 展示消息。注意这只是把前置替换后的问题展示给前端，不等于 ES 层的 `UserQueryRewrite`。
6. 组装 `AgentContext`：`SystemPrompt`、`NextStepPrompt`、`SummarySystemPrompt`、`MaxStep=4`、可用工具。
7. 调用 `ReactAgentService.Handler(..., WithSummaryMode(true))`。
8. 从 `SharedReferences` 取回工具检索到的引用，写入 `QueryReferenceList` 和 `ChatReferenceList`。

## SummaryAgent 何时使用

`ReactAgentService.Handler` 会初始化 `ReactAgent` 和 `SummaryAgent`，但是否运行 SummaryAgent 由 `HandlerOptions.SummaryMode` 决定：

1. `SummaryMode=false`：ReactAgent 结束后直接返回最后一条模型消息。
2. `SummaryMode=true`：ReactAgent 结束后，把 ReactAgent 的消息历史交给 SummaryAgent 生成最终答案。
3. 如果 ReactAgent 抛出 `compose.ErrExceedMaxSteps`，即使原本没开 SummaryMode，也会强制开启 summary，用已有工具结果做总结。

`ForestChatMode` 当前固定传入 `models.WithSummaryMode(true)`，所以 `base_type=standard` 的主链路通常都会在 React 检索之后进入 SummaryAgent。SummaryAgent 的输入不是重新检索，而是 ReactAgent memory 中已经产生的 user/model/tool messages。

## Prompt 和上下文在哪里组装

当前链路里有三类 prompt/context：

| 阶段 | 组装位置 | 内容 |
|---|---|---|
| ReAct 首轮/后续轮模型输入 | `NewReactAgent` 的 `MessageModifier` | `ForestKnowledgeSystemPrompt + BasePrompt`、历史消息、当前 query、可用工具 schema；如果上一条不是 user，还会追加 `NextStepPrompt` |
| Tool schema | `NewForestSearchTool` | 工具名 `forest_search_tool`、工具描述、`SearchRequest.question/search_strategy` JSON schema |
| SummaryAgent 最终回答 | `SummaryAgent.preBuildMsg` | `ForestKnowledgeSummarySystemPrompt`、`query`、`taskHistory`、`history_dialogue` |

第二次 tool call 的“提示词”不是另一个手写 prompt，而是 ReAct 图在第一轮 tool 返回后再次进入 `chat` node：模型会看到原始 system prompt、历史消息、上一轮 assistant tool_calls、上一轮 tool observation，以及追加的 `NextStepPrompt`，再自行决定是否继续调用 `forest_search_tool`。

ES 层的 `UserQueryRewrite` 是另一条内部 chat-agent 调用：Go 代码里只指定 `sys_agent_user_query_rewrite` 这个 agent 名和 `input1=question`，具体提示词来自 chat agent 配置，不在 `rerank.go` 中硬编码。

## 三种 question 的来源

排查日志时最容易混淆的是这几个字段：

| 字段 | 来源 | 用途 |
|---|---|---|
| `OriginalQuestion` | 用户原始问题 `ctxData.Question.Source.Question` | 当前 MR 用它作为 rerank 的 `text_1`，保持用户原始意图 |
| `agentReq.Query` | 原始问题经过同义词/专业词替换 | 传给 ReActAgent，模型基于它决定是否调用工具以及工具参数 |
| `SearchRequest.question` | 模型 function call 自己生成的参数 | 传入 `forest_search_tool.invoke`，继续给 ES 层做改写和召回 |
| `userQuery` | `UserQueryRewrite(SearchRequest.question)` 的输出 | 用于 ES `multi_match` 和 embedding |
| `rerankQuestion` | 优先取 `OriginalQuestion`，否则取工具参数 question | 用于 rerank 请求的 `text_1` |

所以，日志里看到的“未按规定确认收入 企业审计 常见表现形式”这类词，通常不是用户直接输入，而是 ReAct 模型在 tool call 里生成的 `SearchRequest.question`。进入 `NewRerankSearchWrapper` 后还会再走一次 `UserQueryRewrite`，这个结果才是 ES 检索 query 和 embedding 的输入。

## Function Call 时机

ReAct agent 里只有一个知识库工具：`forest_search_tool`。工具是否调用、调用几次、参数是什么，取决于模型在 `chat` node 的输出里是否包含 `tool_calls`。

当前 Eino 封装可以理解成一个小图：

```text
chat node
  -> 模型读取 SystemPrompt + history + user query + tool schema
  -> 如果输出普通 content：进入下一步或结束
  -> 如果输出 tool_calls：交给 tools node

tools node
  -> 执行 forest_search_tool.invoke
  -> 返回 ToolMessage/observation
  -> observation 回到 chat node
```

`newModelHandler` 和 `newToolHandler` 是注册到 Eino graph node 的生命周期 callback：

| callback | 触发时机 | 作用 |
|---|---|---|
| model `OnEnd` | 非流式模型调用结束 | 打印/保存模型输出 |
| model `OnEndWithStreamOutput` | 流式模型持续输出和结束 | 边收边发 SSE，结束时 flush |
| tool `OnStart` | tool 开始执行 | 生成 tool shell 展示消息，记录 `tool_call_id -> msg_id` |
| tool `OnEnd` | 非流式 tool 结束 | 保存 ToolMessage，发送最终 tool 结果 |
| tool `OnEndWithStreamOutput` | 流式 tool 输出和结束 | 分片发送，结束时保存 ToolMessage |

这段配置：

```go
compose.WithCallbacks(modelHandlerCalls).DesignateNode("chat")
compose.WithCallbacks(toolHandlerCalls).DesignateNode("tools")
```

表示 model callback 只挂到 ReAct 内部的 `chat` node，tool callback 只挂到 `tools` node。它不是业务判断逻辑，而是“在固定生命周期点记录和发送消息”。

## 工具策略怎么决定

`forest_search_tool` 的 schema 只有两个输入：

```go
type SearchRequest struct {
    Question       string `json:"question"`
    SearchStrategy string `json:"search_strategy"` // common_questions / knowledge_summary
}
```

`search_strategy` 是模型 function call 参数决定的，不是后端根据某个 if 强制改的。后端只在 `invoke` 里消费这个值：

1. 先查 FQA：`FindFQAByQuestion`。如果命中，直接返回 `search_qa_result`。
2. 未命中 FQA 且 `search_strategy == "knowledge_summary"`，走 `SearchDescription` 摘要检索。
3. 未命中 FQA 且策略不是 `knowledge_summary`，走普通 `RerankSearchQuestionChunk`。

目前没有后端逻辑把 `common_questions` 强制 fallback 成 `knowledge_summary`。所谓“策略变化”通常来自模型下一轮 function call 重新选择了 `search_strategy`。

## 当前 ES + Rerank 流程

当前 MR 分支已经不是旧文档里的“两次 rerank”。当前顺序是：

```text
RerankSearchChunk
  step1 SearchQuestionChunk      // ES 混合召回
  step2 SearchChunkSequence      // 查当前 chunk 的上下文窗口
  step3 JoinNeighborChunks       // 当前 chunk + 上文 + 下文
  step4 SortRerankChunk          // 唯一一次 chunk rerank
  step5 TopM 截断
  step6 GroupByFileID
  step7 Resault                  // 生成 QueryReferenceList
```

`SearchQuestionChunk` 的 ES DSL 核心结构：

```text
filter:
  - is_disable=false 或字段不存在
  - exists embedding
  - type in chunk/image/table/video/formula
  - forest_id/file_id 过滤
should:
  - multi_match(description^description_weight, query=userQuery)
script_score:
  textScore = _score
  vectorScore = (cosineSimilarity(query_vector, embedding) + 1.0) / 2.0
  return textScore + vectorScore * embedding_weight
size:
  topn * fetch_factor
```

可以把 ES 第一阶段分理解为：

$$
\text{score}_{es}=\text{score}_{text}+\text{weight}_{embedding}\times\frac{\cos(q,d)+1}{2}
$$

其中：

| 参数 | 含义 |
|---|---|
| `score_text` | ES `multi_match` 在 `description` 字段上的文本相关性 |
| `weight_embedding` | 配置里的 `EmbeddingWeight` |
| `q` | `UserQueryRewrite` 后问题的 embedding |
| `d` | chunk 的 embedding |

`GetRerank` 请求只有：

```json
{
  "model": "...",
  "text_1": "rerankQuestion",
  "text_2": ["chunk description 1", "chunk description 2"]
}
```

Rerank 模型本身不接收 ES score。ES score 只在 fallback/keyword fallback 的融合排序里参与。

## Neighbor 拼接的当前语义

旧实现曾经在拼左邻居时把当前 chunk 的 `Sequence` 和 `Location` 改成左邻居，导致引用定位偏移。当前实现不再改写这些定位字段。

当前 `JoinNeighborChunks` 的拼接顺序是：

```text
当前 chunk -> 上文 chunk -> 下文 chunk
```

这样做的原因是 rerank 应优先看到被 ES 命中的正文片段，再看到上下文补充，避免开头的目录、标题、邻居内容把相关性判断带偏。

## Rerank 为空时的 fallback

`SortRerankChunk` 正常逻辑：

1. 对拼接后的 chunks 调 `GetRerank(rerankQuestion, descriptions)`。
2. 写入每个 chunk 的 `RerankScore`。
3. 保留 `RerankScore >= RerankThreshold` 的 chunk。
4. 按 `RerankScore` 降序取 `Topn`。

如果阈值过滤后为空，当前 MR 先走 keyword fallback：

1. 使用 IK analyzer 对 `rerankQuestion` 分词。
2. 过滤过短词、去重、优先保留长词，最多 12 个。
3. 在候选 chunk 的中心原文上算关键词命中，不用已拼接邻居后的 `Description`。
4. 要求命中候选集中最稀有关键词，并且关键词分达到阈值。
5. 用关键词分和 ES 分做融合排序。

融合分为：

$$
\text{score}_{final}=0.7\times\text{score}_{keyword}+0.3\times\text{score}_{es\_normalized}
$$

其中：

| 参数 | 含义 |
|---|---|
| `score_keyword` | 按关键词长度和 IDF 加权后的命中比例 |
| `score_es_normalized` | 当前候选 chunk 的 ES score 除以候选集最大 ES score |

这样修复了一个关键问题：如果用拼接后的 `Description` 做 keyword fallback，某个不相关 chunk 可能因为邻居里带有目标词而被误选。当前只评价 ES 原始命中的中心 chunk 原文。

仍需注意：如果 keyword fallback 也没有候选，当前代码仍可能在 `FallBackToTopK=true` 时回到旧的 rerank topK 兜底。这是保守兼容行为，但对“rerank 明显坏掉”的场景仍有误召回风险。

## 本次排查的问题与修复

| 问题 | 表现 | 修复 |
|---|---|---|
| rerank 输入使用工具改写句 | rerank `question` 变成模型生成的陈述/关键词句，偏离用户原始意图 | `NewRerankSearchWrapper` 增加 `RerankSearchOptions.OriginalQuestion`，`GetRerank` 使用 `rerankQuestion` |
| 看不到 tool call 问题来源 | 日志只看到工具入参，不知道是哪轮模型生成 | 增加 `[DEBUG][tool-question-source]`：可用工具、每轮模型输入、模型输出 tool_calls |
| 两次 rerank 放大不稳定 | 第一次 rerank 就可能把正确 chunk 排掉 | 去掉前置 rerank，改为 ES 召回后先扩邻居，再统一 rerank |
| 邻居拼接导致引用定位偏移 | 命中 seq44，但返回引用可能对到邻居 seq43/46 | `JoinNeighborChunks` 不再改写中心 chunk 的 `Sequence/Location` |
| 邻居开头干扰 rerank | 拼接文本开头是上文或目录，rerank 更容易关注无关信息 | 拼接顺序改为“当前、上文、下文” |
| rerank 全部低分后 topK 兜底误召回 | 无关 chunk 的 rerank 分高，正确 chunk 分低 | 增加 keyword fallback，先用原始问题关键词证据过滤 |
| keyword fallback 被邻居污染 | seq46 因邻居含目标词被选，真正 seq44 反而落后 | fallback 只评价中心 chunk 原文 |

## 日志怎么读

推荐按 requestID 过滤，再看这几组关键日志。

### 工具参数来源

```text
[DEBUG][tool-question-source] ReActAgent available tools
[DEBUG][tool-question-source] ReActAgent model input begin
[DEBUG][tool-question-source] ReActAgent model input message content chunk
[DEBUG][tool-question-source] ReActAgent model output tool_calls
```

用途：

1. 看模型实际收到的新增/变化 prompt、history、tool observation。日志会跳过与上一轮相同的前缀消息，避免每轮重放全部历史。
2. 看 `forest_search_tool` 参数是模型哪一轮生成的。
3. 判断第二次 tool call 是否因为上一轮 observation 触发。

### 检索链路

```text
[DEBUG][chunk-empty] ForestSearchTool.invoke start
[DEBUG][chunk-empty] NewRerankSearchWrapper start
[DEBUG][chunk-empty] NewRerankSearchWrapper UserQueryRewrite
[DEBUG][chunk-empty] SearchQuestionChunk ES result
[DEBUG][chunk-empty] RerankSearchChunk step1_ES_search
[DEBUG][chunk-empty] RerankSearchChunk step2_neighbor_search
[DEBUG][chunk-empty] RerankSearchChunk step3_join_neighbor
[DEBUG][chunk-empty] GetRerank request stats
[DEBUG][chunk-empty] SortRerankChunk after threshold filter
[DEBUG][chunk-empty] RerankSearchChunk step7_final_result
```

用途：

1. 对比 `question`、`original_question`、`rerank_question`、`UserQueryRewrite rewritten`。
2. 确认目标 chunk 是否在 ES 第一阶段已经召回。
3. 确认目标 chunk 是否在邻居拼接后保留了自己的 sequence/location。
4. 确认目标 chunk 是被 rerank 阈值丢掉、TopM 丢掉，还是最终结果聚合丢掉。

### chunk 留存与得分

```text
[DEBUG][chunk-rank] stage=...
[DEBUG][chunk-selection] stage=...
[DEBUG][chunk-empty] SortRerankChunk keyword fallback candidate
```

关注字段：

| 字段 | 含义 |
|---|---|
| `rank` | 当前阶段排序 |
| `outcome` | kept/dropped 以及原因 |
| `file_id`/`sequence`/`chunk_id` | 定位 chunk |
| `score` | ES score |
| `rerank_score` | rerank 模型分 |
| `final_score` | keyword fallback 融合分 |
| `keyword_score` | 关键词证据分 |
| `hit_keywords` | 命中的关键词 |
| `rare_keywords` | 命中的稀有关键词 |

## 常见判定

| 现象 | 优先判断 |
|---|---|
| 日志里 `question` 不是用户原话 | 正常，可能是 ReAct 模型生成的 tool call 参数 |
| `UserQueryRewrite` 后又出现一轮改写 | 正常，这是 ES 层 query rewrite，不是前端展示的 `question_rewrite` |
| ES 第一阶段没有目标 chunk | 先看 `UserQueryRewrite rewritten`、file/forest filter、type/is_disable/embedding |
| ES 有目标 chunk，rerank 后没了 | 看 `rerank_score` 和阈值；当前 fallback 应优先看 keyword evidence |
| 命中文件正确但 sequence 不对 | 看 `JoinNeighborChunks`、中心 chunk 原文、`sequence/location` 是否保留中心 chunk |
| keyword fallback 选了邻居相关 chunk | 检查是否部署到包含中心 chunk 原文修复的版本 |
| SummaryAgent 超时后空答案 | 看 `summary-agent-timeout-swallowed.md` |

## 当前遗留风险

1. `FallBackToTopK=true` 时，如果 keyword fallback 没候选，仍可能回到旧的 rerank topK 兜底。
2. keyword fallback 的“必须命中最稀有关键词”规则偏硬，IK analyzer 偶发词可能影响召回。
3. `UserQueryRewrite` 仍会影响 ES 第一阶段召回；如果它把 query 改窄，正确 chunk 可能根本进不了候选集。
4. `GetRerank` 失败时目前仍直接返回 error，token overflow 相关降级还应独立治理。
