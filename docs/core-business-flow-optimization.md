# CoreKG 核心业务流程优化方案

> 基于 `core-business-flow.md` 分析的问题清单与改造方案

---

## P0: 双轮 Rerank 延迟优化

### 问题

8步管线中 Step 2 和 Step 5 各调用一次外部 Rerank API，每次网络往返 200-800ms。流式场景下首字延迟 = 查询改写 + Embedding + ES检索 + Rerank#1 + 邻域扩展 + Rerank#2，累计可达 3-5 秒。

### 现状代码位置

- `apps/kesearch/models/reranksearch/essearch.go` - `SortRerankChunk()`
- `apps/kesearch/models/reranksearch/search.go` - `RerankSearchChunk()` Step 2 & Step 5

### 改造方案

**方案 A: 第2轮改用本地轻量 Cross-Encoder（推荐）**

1. 部署 `bge-reranker-v2-m3` 或 `bce-reranker-base_v1` 本地服务
2. 第1轮仍用外部 Rerank API（高精度粗筛）
3. 第2轮改用本地 Cross-Encoder（低延迟精排）
4. 配置项 `knowledge.rerank_local` 新增本地 rerank 地址

```go
// essearch.go - SortRerankChunk 增加 round 参数
func (w *RerankSearchWrapper) SortRerankChunk(chunks []*SearchType, round int) ([]*SearchType, error) {
    if round == 2 && w.localRerankClient != nil {
        return w.localRerank(chunks)
    }
    return w.remoteRerank(chunks)
}
```

预期收益: 第2轮延迟从 300ms 降至 30-50ms，总延迟降低 30-40%。

**方案 B: 条件触发第2轮 Rerank**

仅在邻域扩展后 chunk 数量 > Topn 时执行第2轮，否则直接截断:

```go
// search.go - RerankSearchChunk
if len(nbres) <= w.conf.Topn {
    nbsres = nbres[:min(len(nbres), w.conf.Topn)]
} else {
    nbsres, err = w.SortRerankChunk(nbres, 2)
}
```

预期收益: 简单查询跳过第2轮，延迟降低 20-30%。

**方案 C: 合并两轮为单轮 + 邻域后重排**

取消第2轮外部 Rerank，邻域扩展后用原始分数 + 位置权重重新排序:

```go
// JoinNeighborChunks 后直接按 centerScore * 0.7 + neighborBonus * 0.3 排序
```

预期收益: 节省一次完整外部调用，延迟降低 40-50%，但精度可能下降 2-5%。

### 推荐

优先实施方案 B（零成本、即时生效），同时推进方案 A 作为中期目标。

---

## P0: Embedding API 缓存与重试

### 问题

`GetEmbedding` 无缓存、无重试。相同查询重复调用浪费配额；API 瞬时故障直接导致整个问答失败。

### 现状代码位置

- `apps/kesearch/models/essearch/embedding.go` - `GetEmbedding()`

### 改造方案

```go
var embeddingCache = redispool.Redis()

func GetEmbedding(question string) (ragtypes.Embedding, error) {
    cacheKey := fmt.Sprintf("emb:%x", md5.Sum([]byte(question)))

    // 1. 查缓存
    if cached, err := embeddingCache.Get(cacheKey).Bytes(); err == nil {
        var emb ragtypes.Embedding
        if json.Unmarshal(cached, &emb) == nil {
            return emb, nil
        }
    }

    // 2. 带重试调用
    var emb ragtypes.Embedding
    var err error
    for i := 0; i < 3; i++ {
        emb, err = callEmbeddingAPI(question)
        if err == nil {
            break
        }
        time.Sleep(time.Duration(1<<uint(i)) * 100 * time.Millisecond)
    }
    if err != nil {
        return nil, fmt.Errorf("embedding API failed after 3 retries: %w", err)
    }

    // 3. 写缓存 (TTL 24h)
    if data, _ := json.Marshal(emb); data != nil {
        embeddingCache.Set(cacheKey, data, 24*time.Hour)
    }

    return emb, nil
}
```

### 注意事项

- 缓存 key 用 MD5 而非原文，避免 key 过长
- TTL 24h 足够（知识库更新频率低）
- 重试用指数退避，避免雪崩
- 可加配置开关 `knowledge.embedding.cache_enabled`

预期收益: 重复查询延迟降至 <1ms；API 抖动时可用性从 ~90% 提升至 ~99%。

---

## P1: 邻域拼接后 Rerank 失真修复

### 问题

`JoinNeighborChunks` 将左右邻居拼接到 description，拼接后文本可能从 512 token 膨胀到 1500+ token。Rerank 模型对长文本打分偏移，导致高质量短 chunk 被低质量长 chunk 挤掉。

### 现状代码位置

- `apps/kesearch/models/reranksearch/essearch.go` - `JoinNeighborChunks()`
- `apps/kesearch/models/reranksearch/essearch.go` - `SortRerankChunk()`

### 改造方案

**Rerank 时使用中心 chunk 原文打分，邻居仅作为上下文附加:**

```go
type SearchType struct {
    // ... existing fields
    OriginalDescription string  // 新增: 保存拼接前的原始描述
}

// JoinNeighborChunks 中保存原始描述
func (w *RerankSearchWrapper) JoinNeighborChunks(chunks, nbchunks []*SearchType) []*SearchType {
    for _, chunk := range chunks {
        chunk.OriginalDescription = chunk.Description
        // ... 拼接逻辑不变，Description 变为拼接后文本
    }
}

// SortRerankChunk 中使用 OriginalDescription 送 Rerank
func (w *RerankSearchWrapper) SortRerankChunk(chunks []*SearchType, round int) ([]*SearchType, error) {
    docs := make([]string, len(chunks))
    for i, c := range chunks {
        if c.OriginalDescription != "" {
            docs[i] = c.OriginalDescription
        } else {
            docs[i] = c.Description
        }
    }
    rerankResp, err := GetRerank(w.ctx, w.userQuery, docs)
    // ... 打分后仍返回完整 Description（含邻居）
}
```

预期收益: Rerank 打分准确性提升，回答相关性改善 5-10%。

---

## P1: 查询改写条件化

### 问题

`UserQueryRewrite` 每次搜索前都调用 LLM，对于精确术语查询（如 "P0300故障码含义"）改写反而引入噪声，且增加 500-1000ms 延迟。

### 现状代码位置

- `apps/kesearch/models/reranksearch/wrapper.go` - `NewRerankSearchWrapper()`
- `apps/kesearch/models/reranksearch/rerank.go` - `UserQueryRewrite()`

### 改造方案

```go
func shouldRewriteQuery(question string) bool {
    // 1. 过短不改写
    if len([]rune(question)) < 5 {
        return false
    }
    // 2. 包含明确专有名词模式不改写 (如 P+数字故障码、零件号等)
    if regexp.MustCompile(`[A-Z]\d{3,}|[A-Z]{2,}\d+`).MatchString(question) {
        return false
    }
    // 3. 已经是完整问句且 < 30字 不改写
    if len([]rune(question)) < 30 && strings.Contains(question, "？") {
        return false
    }
    return true
}

func NewRerankSearchWrapper(...) (*RerankSearchWrapper, error) {
    // ...
    var rewrittenQuery string
    if shouldRewrite(question) {
        rewrittenQuery, err = UserQueryRewrite(ctx, question)
        if err != nil {
            rewrittenQuery = question // 改写失败回退原文
        }
    } else {
        rewrittenQuery = question
    }
    // ...
}
```

预期收益: 30-40% 的查询跳过改写，平均延迟降低 300-500ms。

---

## P1: Summary 引用白名单校验

### 问题

Summary prompt 要求引用严格来自 taskHistory 中的 chunks，但 ReAct 过程中引用信息可能丢失或格式变化，导致 Summary 生成无效引用 `{Reference §0[0]}` 或遗漏有效引用。

### 现状代码位置

- `apps/kechat/chat/modes/forest_agent.go` - `Run()` 中引用聚合
- `apps/kechat/chat/prompt/forest_agent/summary_prompt.go` - 引用规则

### 改造方案

在 ReAct 结束后、Summary 之前，将 `SharedReferences.GetAggregated()` 的结果显式序列化注入 Summary prompt:

```go
// forest_agent.go - Run() 中
refs := sharedRefs.GetAggregated()

// 构建引用白名单 JSON
type RefWhitelistEntry struct {
    FileID   uint   `json:"file_id"`
    FileName string `json:"file_name"`
    Chunks   []struct {
        Sequence    int    `json:"sequence"`
        Description string `json:"description"`
    } `json:"chunks"`
}

whitelist := buildRefWhitelist(refs)
whitelistJSON, _ := json.Marshal(whitelist)

// 注入到 Summary prompt 的 extra 参数中
summaryOpts.ExtraPrompt = fmt.Sprintf(
    "\n\n<referenceWhitelist>\n%s\n</referenceWhitelist>\n"+
    "你只能使用以上 whitelist 中的 file_id 和 sequence 生成引用标签。",
    string(whitelistJSON),
)
```

同时在 Summary prompt 模板中增加:

```
引用校验规则:
- 生成答案后，逐条检查 {Reference §X[Y]} 中的 X 和 Y 是否存在于 <referenceWhitelist>
- 不存在则删除该引用标签
- 答案中涉及的知识点若有对应 whitelist 条目但未引用，必须补上
```

预期收益: 无效引用减少 80%+，引用覆盖率提升 10-20%。

---

## P2: QA 匹配与检索并行化

### 问题

`ForestSearchTool.invoke()` 中 QA 匹配、Description 搜索、Chunk 搜索串行执行。QA 匹配耗时 50-200ms，即使命中也要等后续逻辑判断。

### 现状代码位置

- `apps/kesearch/pkg/ai/tools/forest_search_tool.go` - `invoke()`

### 改造方案

```go
func (t *forestSearchTool) invoke(ctx context.Context, req SearchRequest) (*SearchResult, error) {
    // QA 匹配和正常检索并行
    type qaResult struct {
        answer string
        found  bool
        err    error
    }
    qaCh := make(chan qaResult, 1)

    go func() {
        answer, found, err := svcessearch.FindFQAByQuestion(ctx, ...)
        qaCh <- qaResult{answer, found, err}
    }()

    // 同时启动正常检索
    normalCh := make(chan normalResult, 1)
    go func() {
        refs, err := svcessearch.RerankSearchQuestionChunk(ctx, ...)
        normalCh <- normalResult{refs, err}
    }()

    // 先等 QA 结果
    qa := <-qaCh
    if qa.found && qa.err == nil {
        // QA 命中，取消正常检索（context cancel）
        cancelNormal()
        return buildQAResult(qa.answer), nil
    }

    // QA 未命中，等正常检索
    normal := <-normalCh
    return buildNormalResult(normal.refs, normal.err), nil
}
```

预期收益: QA 命中时延迟从 200ms+2s 降至 200ms；未命中时无额外开销。

---

## P2: MaxStep 自适应

### 问题

MaxStep 硬编码（ForestAgent=6, Forest=4），简单问题浪费步数（多消耗 Token），复杂问题不够用（被迫提前 Summary）。

### 现状代码位置

- `apps/kechat/chat/modes/forest_agent.go` - `MaxStep = 6`
- `apps/kechat/chat/modes/forest.go` - `MaxStep = 4`

### 改造方案

```go
func adaptiveMaxStep(question string, historyLen int) int {
    base := 4
    qLen := len([]rune(question))

    // 问题越长越复杂
    if qLen > 50 {
        base += 2
    }
    if qLen > 100 {
        base += 2
    }

    // 多轮对话可能需要更多步数
    if historyLen > 4 {
        base += 2
    }

    // 上限保护
    if base > 10 {
        base = 10
    }
    return base
}

// forest_agent.go
agentContext.MaxStep = adaptiveMaxStep(question, len(history))
```

预期收益: 简单问题减少 1-3 步无效推理（节省 Token 15-30%）；复杂问题多获得 2-4 步探索空间。

---

## P2: DirectModel Fallback Summary 优化

### 问题

Agent 达到 maxStep 或无最终输出时，创建全新无工具 Agent 重新生成答案，重复消耗整个上下文的 Token。

### 现状代码位置

- `apps/kechat/chat/modes/direct_model_chat.go` - `Run()` 中 fallback 逻辑

### 改造方案

复用 SummaryAgent 而非重建 Agent:

```go
// 替换原有的 rebuild agent 逻辑
if !hasFinalOutput || reachedMaxStep {
    summaryAgent, err := agent.NewSummaryAgent(ctx, agentContext)
    if err == nil {
        answer, err = summaryAgent.RunSummarizeResult(ctx, allMessages)
        if err == nil && answer != "" {
            finalAnswer = answer
        }
    }
    // SummaryAgent 也失败才回退到最后一条 assistant 消息
}
```

预期收益: fallback 时 Token 消耗减少 50-70%（SummaryAgent MaxStep=2 vs 完整 Agent MaxStep=20）。

---

## P3: 代码质量修复

### 3.1 拼写错误修正

| 文件 | 错误 | 修正 | 影响范围 |
|------|------|------|----------|
| `reranksearch/essearch.go` | `Resault()` | `Result()` | 仅内部调用 |
| `reranksearch/type.go` | `EnabelAbstract` | `EnableAbstract` | 配置字段(需迁移) |
| `graphragsearch/agent/catalogue.go` | `NewCatalpgueAgent` | `NewCatalogueAgent` | 仅内部调用 |

注意: `EnabelAbstract` 如果已持久化到 YAML 配置，需同时更新配置模板并加兼容读取。

### 3.2 硬编码提取

**NebulaGraph space:**

```yaml
# config.yaml
knowledge:
  graphsearch:
    nebula_space: "a_car_test"
    graph_step_depth: 3
```

```go
// graph_search.go
space := settings.GetString("knowledge.graphsearch.nebula_space", "a_car_test")
stepDepth := settings.GetInt("knowledge.graphsearch.graph_step_depth", 3)
```

**Analyst Agent fileIDS:**

移除硬编码的 `fileIDS` 变量，改为从 session 上下文或配置动态获取。

### 3.3 Token 估算改进

```go
// 替换 approxUsage 中的 chars/4
func approxUsage(messages []kellmtype.Message, answer string) Usage {
    totalChars := 0
    for _, m := range messages {
        totalChars += len([]rune(m.Content))
    }
    totalChars += len([]rune(answer))

    // 中文约 1.5 字符/token，英文约 4 字符/token
    // 混合文本取 2.5 字符/token 作为折中
    tokens := int(math.Ceil(float64(totalChars) / 2.5))
    if tokens < 1 && totalChars > 0 {
        tokens = 1
    }
    return Usage{TotalTokens: tokens}
}
```

更优方案: 引入 `tiktoken-go` 库按模型精确计算，但需评估性能影响。

### 3.4 异步会话名称更新兜底

```go
func updateSessionNameWithLLMAsync(ctx context.Context, session, question, answer string) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                log.Errorf("updateSessionName panic: %v", r)
                // 兜底: 用问题前20字作为会话名称
                fallbackName := truncateRunes(question.Question, 20)
                chatdao.UpdateSessionName(session.ID, fallbackName)
            }
        }()

        // ... 原有 LLM 调用逻辑
        name, err := generateSessionName(ctx, question, answer)
        if err != nil {
            // LLM 失败兜底
            fallbackName := truncateRunes(question.Question, 20)
            chatdao.UpdateSessionName(session.ID, fallbackName)
            return
        }
        chatdao.UpdateSessionName(session.ID, name)
    }()
}
```

---

## 实施路线图

```
Phase 1 (1-2周) - 零风险快速收益
  [x] P0-B: 条件触发第2轮 Rerank
  [x] P0:   Embedding 缓存 + 重试
  [x] P1:   查询改写条件化
  [x] P3.1: 拼写错误修正

Phase 2 (2-4周) - 核心质量提升
  [ ] P1:   邻域拼接 Rerank 失真修复
  [ ] P1:   Summary 引用白名单校验
  [ ] P2:   QA 匹配并行化
  [ ] P3.2: 硬编码提取

Phase 3 (4-6周) - 架构优化
  [ ] P0-A: 本地 Cross-Encoder 第2轮 Rerank
  [ ] P2:   MaxStep 自适应
  [ ] P2:   DirectModel Fallback 优化
  [ ] P3.3: Token 估算改进
  [ ] P3.4: 会话名称兜底
```

---

## 监控建议

每个优化项上线后需监控:

| 指标 | 说明 | 告警阈值 |
|------|------|----------|
| `chat.first_token_latency_p99` | 首字延迟 | > 5s |
| `search.rerank.latency_ms` | Rerank API 延迟 | > 1000ms |
| `search.embedding.cache_hit_rate` | Embedding 缓存命中率 | < 20% |
| `search.rerank.fallback_rate` | Rerank 全过滤回退率 | > 30% |
| `agent.react.steps_avg` | 平均推理步数 | > MaxStep * 0.8 |
| `agent.summary.invalid_ref_rate` | 无效引用比例 | > 10% |
| `chat.session_name_empty_rate` | 会话名称为空比例 | > 5% |
| `search.qa.hit_rate` | QA 匹配命中率 | < 5% (过低说明 QA 库需补充) |
