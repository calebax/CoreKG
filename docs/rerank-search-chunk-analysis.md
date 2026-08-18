# RerankSearchChunk 实现逻辑分析

## 1. 概述

`RerankSearchChunk` 是知识库检索的 rerank 版本 API，注册在 `apps/kesearch/internal/apis/apis.go:46`：

```go
eng.PRequireLogin("kesearch.RerankSearchChunk", RerankSearchChunk)
```

该 API 需要登录态，接收用户问题、森林/文件 ID 过滤条件及可选搜索配置，通过 ES 混合检索（全文 + 向量）+ 外部 Rerank 模型重排序，返回按文件聚合的 Chunk 检索结果。

## 2. 三层架构

```
API Handler                  Service                        Model
─────────────────────────────────────────────────────────────────────
reranksearch.go          svcreranksearch/              reranksearch/
RerankSearchChunk  ───▶  RerankSearchChunk()  ───▶  RerankSearchWrapper
  │ 参数校验                   │                          │
  │                           ├─ FindFQAByQuestion()     ├─ essearch.go
  │                           │  (FQA 问答对命中)         │  (ES 检索)
  │                           │                          │
  │                           └─ wrapper.Rerank-         ├─ rerank.go
  │                              SearchChunk()           │  (Rerank & 问题改写)
  │                                                      │
  │                                                      └─ type.go
  │                                                         (配置/模型定义)
```

### 2.1 API Handler 层

`apps/kesearch/internal/apis/reranksearch.go:18` — 职责：

1. 调用 `req.Validity(resp)` 校验参数
2. 调用 `svcreranksearch.RerankSearchChunk(ctx, req)` 执行业务逻辑
3. 错误处理：统一返回 `errcode.ErrCode_InternalError`（code=500）

### 2.2 Service 层

`apps/kesearch/services/svcreranksearch/reranksearch.go:16` — 职责：

1. 优先调用 `svcessearch.FindFQAByQuestion()` 搜索 FQA 问答对，命中则直接返回
2. 未命中则创建 `RerankSearchWrapper`，执行核心 8 步检索流程
3. 将 `chattype.QueryReferenceList` 组装为响应

### 2.3 Model 层（核心）

`apps/kesearch/models/reranksearch/search.go:9` — `RerankSearchWrapper.RerankSearchChunk()` 是核心方法，包含完整的 8 步检索流水线。

## 3. 完整执行流程

### 3.0 前置处理

#### 参数校验

`apps/kesearch/internal/dto/dtoreranksearch/request.go:21` — `Validity()` 方法检查：

- `question` 不能为空
- `forest_ids` 和 `file_ids` 至少提供一个
- 如果提供了 `config`，校验 `SearchConfig` 的合法性（权重之和为 1、TopN/TopM/TopK > 0、阈值在 0~1 之间等）

#### FQA 问答对命中检测

Service 层首先调用 `FindFQAByQuestion(ctx, "ke_0", question, forestIDs, fileIDs)` 在 ES 中搜索 FQA 类型的文档。如果命中，直接构建 `QueryReferenceList` 返回，**跳过后续全部检索流程**。

FQA 结果构建逻辑（`buildFQAResult`）：
- 取 ES 命中结果的第一条
- 如果有 `QAAnswerID`，使用答案 ID 作为 ChunkID；否则使用 Q 的 ID
- 返回类型标记为 `ragtypes.ChunkTypeFQA`

#### Wrapper 初始化

`apps/kesearch/models/reranksearch/wrapper.go:26` — `NewRerankSearchWrapper()` 按顺序初始化：

1. **ES 客户端** — `essearch.InitESClient(ctx)`，单例模式，配置 key：`knowledge.es`
2. **问题改写** — `UserQueryRewrite(ctx, question)`，通过内部 Chat Agent 将原始问题改写为更利于检索的查询语句
3. **向量嵌入** — `essearch.GetEmbedding(quest)`，将改写后的问题转为向量，配置 key：`knowledge.embedding`
4. **搜索配置** — 若传入 `config` 为 nil，使用 `GetDefaultConfig()` 获取默认配置，配置 key：`knowledge.reranksearchcfg`

---

### Step 1: ES 混合检索（全文 + 向量）

`apps/kesearch/models/reranksearch/essearch.go:24` — `SearchQuestionChunk()`

#### ES 查询结构

```
bool
  ├── filter (必须匹配)
  │   ├── is_disable = false OR is_disable 不存在
  │   ├── embedding 字段必须存在
  │   ├── type IN ["chunk", "image", "table", "video", "formula"]
  │   ├── forest_id IN forestIds（若提供）
  │   └── file_id IN fileIds（若提供）
  └── should
       └── multi_match on description（BM25 全文匹配）
  └── script_score（自定义打分）
       └── textScore + vectorScore * embedding_weight
```

#### 自定义打分脚本

```painless
double textScore = _score;
double vectorScore = 0.0;
if (doc['embedding'].size() > 0) {
    vectorScore = (cosineSimilarity(params.query_vector, 'embedding') + 1.0) / 2.0;
}
return textScore + vectorScore * params.embedding_weight;
```

- `textScore`：BM25 全文检索分数（description 字段）
- `vectorScore`：余弦相似度归一化到 [0,1] 区间，公式 `(cos + 1) / 2`
- 最终分数：`textScore + vectorScore * embedding_weight`

#### 检索数量

`TopN * FetchFactor`，确保第一次检索召回足够多的候选（默认 30 * 2 = 60 条）。

#### 结果转换

通过 `getSearchType()` 将 ES 结果转为 `[]*SearchType`，特殊处理：
- **image 类型**：描述格式 `![图片描述：xxx] url`
- **video 类型**：描述格式 `![视频帧描述：xxx] url`
- **table 类型**：描述格式 `表格数据：xxx\n 表格描述：xxx`

---

### Step 1.1: Chunk 类型枚举及其处理机制

检索过程中涉及以下 6 种 Chunk 类型常量（定义在 `apps/keparser/models/ragtypes/chunks.go:13-22`）：

| ChunkType | 常量值 | ES 主检索 type 过滤 | Embedding 要求 | Description 格式化规则 | 邻居搜索 type 过滤 |
|-----------|--------|-------------------|----------------|----------------------|-------------------|
| `chunk` | `"chunk"` | 纳入 | 必须有 | 原文 | 纳入 |
| `image` | `"image"` | 纳入 | 必须有 | `![图片描述：{desc}] {url}` | 纳入 |
| `table` | `"table"` | 纳入 | 必须有 | `表格数据：{table}\n 表格描述：{desc}` | 纳入 |
| `video` | `"video"` | 纳入 | 必须有 | `![视频帧描述：{desc}] {url}` | 纳入 |
| `formula` | `"formula"` | 纳入 | 必须有 | 原文 | **不纳入** |
| `FQA` | `"FQA"` | **不纳入主检索** | 不需要 | — | **不纳入** |

**关键差异**：

1. **ES 主检索**（`SearchQuestionChunk`）的 type filter 为 `["chunk", "image", "table", "video", "formula"]`，明确排除 `FQA` 和 `entity`/`relationship` 等图谱类型
2. **Embedding 要求**：主检索要求 `embedding` 字段必须存在（`exists` 查询），因此 Chunk 入库时必须生成向量
3. **Description 格式化**：在 `getSearchType()` 函数（`type.go:128-161`）中完成，将不同类型 Chunk 的原始字段拼接为可被 Rerank 模型理解的统一文本
4. **邻居搜索**（`SearchChunkSequence`）的 type filter 为 `["chunk", "image", "table", "video"]`，排除 `formula` 和 `FQA`——公式类型的上下文拼接被认为无意义

所有类型的 ES Source 字段均来自 `ragtypes.Chunk` 结构体（`apps/keparser/models/ragtypes/chunks.go:44-98`），关键字段：

| 字段 | 类型 | JSON Tag | 说明 |
|------|------|----------|------|
| `Type` | `ChunkType` | `type` | Chunk 类型标识 |
| `Content` | `string` | `content` | 原始文本内容 |
| `Description` | `string` | `description` | LLM 生成的语义描述 |
| `Table` | `string` | `table` | JSON 格式的表格数据 |
| `ImageUrl` | `string` | `image_url` | 图片/视频帧 URL |
| `Embedding` | `Embedding` | `embedding` | 向量表示 |
| `Sequence` | `int` | `sequence` | 在文件中的序号 |
| `Location` | `[5]int` | `location` | 在文件中的坐标位置 |

---

### Step 1.2: Table 类型专项分析

`table` 类型的 Chunk 在检索全流程中的处理链路如下：

#### ES 存储结构

`ragtypes.Chunk.Table` 字段为 `string` 类型，存储 JSON 格式的表格原始数据。`Description` 字段存储 LLM 对表格内容生成的自然语言描述。两个字段在入库时分别生成。

#### Description 格式化

`type.go:141-142`：

```go
if v.Source.Type == "table" {
    desc = fmt.Sprintf("表格数据：%s\n 表格描述：%s", v.Source.Table, v.Source.Description)
}
```

格式化后的 Description 同时包含表格原始数据和语义描述，例如：
```
表格数据：[{"列1":"值A","列2":"值B"},...]
 表格描述：该表格展示了2024年各季度销售额数据，Q1至Q4呈增长趋势...
```

#### 向量检索行为

table 类型 Chunk 的 embedding 向量基于其 `description + table` 内容生成。在 ES 混合检索的 `script_score` 阶段，与其他类型（chunk/image/video/formula）使用相同的余弦相似度计算公式，无差别竞争：

```
vectorScore = (cosineSimilarity(params.query_vector, 'embedding') + 1.0) / 2.0
```

同时，table 的 `description` 字段参与 BM25 全文检索（`multi_match`），权重为 `DescriptionWeight`（默认 0.3）。

#### Rerank 混合处理

格式化后的 Description（即 `"表格数据：...\n 表格描述：..."`）作为纯文本送入外部 Rerank 模型的 `text_2` 参数。Rerank 模型基于完整文本内容（包括表格数据字符串和自然语言描述）评估其与用户问题的相关性，**无需区分 Chunk 类型**。

与纯文本 `chunk` 类型相比，table 的优势是格式化文本中同时包含结构化数据和语义描述，在 Rerank 时比仅有描述的 chunk 能提供更丰富的匹配信号。

#### 邻居搜索

table 类型在邻居搜索中被纳入（type filter 包含 `table`）。当排序后的 table Chunk 进入 Step 3 邻居搜索时，系统会查找同一文件的相邻 Chunk（可能是 chunk、image 或其他 table），并在 Step 4 中将其 Description 拼接进来。

例如，一个 table Chunk 的左右邻居是普通 chunk，拼接后的 Description 变为：
```
{左邻居chunk文本} 表格数据：{table} 表格描述：{desc} {右邻居chunk文本}
```

这使得 Rerank 模型在第二步重排序时能看到表格的上下文。

#### 最终返回

在 `Resault()` 组装最终结果时，table 的 `Type = "table"`，`Content` 字段即格式化后的 Description（包含表格数据和描述），存入 `QueryReferenceChunk.Content`，客户端可以根据 `Type` 字段区别渲染。

---

### Step 1.3: 去重缺失与重复数据风险

整个检索流程（Step 1 ~ Step 8）**不包含任何基于 ChunkID、Content 或 Description 的去重逻辑**。

| 步骤 | 方法 | 去重检查 |
|------|------|----------|
| Step 1 | `SearchQuestionChunk` | ES 保证单次搜索 `_id` 不重复，但**不保证内容不重复** |
| Step 2 | `SortRerankChunk` | 仅排序过滤，无去重 |
| Step 3 | `SearchChunkSequence` | 无去重；`nbchunksMap` 使用 `FileID:Sequence` 为 key，相同 key 后写覆盖 |
| Step 4 | `JoinNeighborChunks` | 无去重；`nbchunksMap` 仅用于查邻居，不对 `chunks` 去重 |
| Step 5 | `SortRerankChunk` | 同 Step 2 |
| Step 6 | TopM 截断 | 仅控制数量，无去重 |
| Step 7 | `GroupByFileID` | 仅按文件分组，无去重 |
| Step 8 | `Resault` | 仅遍历组装，无去重 |

#### 重复数据的产生路径

**入库层（外部算法服务）**是关键源头：

- Chunk 切分和 ES 写入由外部算法服务（`workerServerURL + "/split"`）完成，本仓库中的 `apps/keworker/cmd/worker_split_text_chunk.go` 仅通过 HTTP 转发请求
- 入库前虽有清理逻辑（`DeleteFileReferencesFileChunk`，按 `file_id` 删除），但若清理与入库存在并发竞争，或清理失败，会产生残留
- 若入库时 ES 自动生成 `_id`（而非使用稳定的 `FileID:Sequence:Type`），同一表格重新切分会生成内容相同但 `_id` 不同的重复文档

**检索层完全透传重复数据**：

- ES 单次 search 保证返回结果 `_id` 不重复，但两个 `_id` 不同、内容和分数相同的 chunk 会被同时召回
- 后续 Step 2~8 没有任何去重，导致它们一路透传至最终结果

#### 现象

若一个 table Chunk 在 ES 中存在两条 `_id` 不同但 `Table`、`Description` 完全相同的数据，检索后会返回"完全相同"的两条 Chunk 记录，ChunkID 不同但 Content 值一致。

#### 排查建议

1. 在 ES 中按 `file_id` + `type = "table"` 查询，验证是否存在多条 `Table` 字段相同但 `_id` 不同的记录
2. 检查外部算法服务的 Chunk 切分逻辑和 ES 写入的 `_id` 生成策略
3. 考虑在检索结果组装（`Resault`）或 GroupByFileID 后增加基于 Content 或 FileID+Sequence+Type 的去重

---

`apps/kesearch/models/reranksearch/essearch.go:108` — `SortRerankChunk()`

#### 流程

1. 如果 `EnableRerank = false`，直接截断到 TopN 返回
2. 提取所有 Chunk 的 Description 文本
3. 调用外部 Rerank 服务 `GetRerank(ctx, question, descriptions)` 获取重排序分数
4. 更新每个 Chunk 的 `RerankScore`
5. **阈值过滤**：只保留 `RerankScore >= RerankThreshold`（默认 0.5）的 Chunk
6. 按 RerankScore 降序排列
7. 截断到 TopN（默认 30）

#### 兜底策略

如果阈值过滤后结果为空且 `FallBackToTopK = true`（默认开启）：
- 跳过阈值，直接按 RerankScore 降序取 TopK（默认 5）

#### 外部 Rerank 服务

`apps/kesearch/models/reranksearch/rerank.go:23` — `GetRerank()`

配置 key：`knowledge.rerank`，包含 `url`、`key`、`model_name`。

请求格式：
```json
{
  "model": "model_name",
  "text_1": "用户问题",
  "text_2": ["chunk1描述", "chunk2描述", ...]
}
```

响应：
```json
{
  "data": [
    {"index": 0, "score": 0.85, "object": "relevance_score"},
    {"index": 1, "score": 0.42, "object": "relevance_score"}
  ]
}
```

---

### Step 3: 邻居 Chunk 搜索

`apps/kesearch/models/reranksearch/essearch.go:169` — `SearchChunkSequence()`

对 Step 2 得到的每个 Chunk，在 ES 中搜索其上下相邻的 Chunk。

#### 查询逻辑

对每个输入 Chunk，构建 bool 查询：
- 同一 `file_id`
- `sequence` 在 `[Sequence - NeighborSize, Sequence + NeighborSize]` 范围内（默认 ±1）
- 排除当前 Chunk 自己的 sequence
- 排除 FQA 类型，只搜索 `["chunk", "image", "table", "video"]`
- 过滤已禁用的 Chunk

搜索总数为 `len(chunks) * (2 * NeighborSize + 1)`。

---

### Step 4: 邻居拼接

`apps/kesearch/models/reranksearch/essearch.go:245` — `JoinNeighborChunks()`

将原始 Chunk 与其找到的邻居 Chunk 按 Description 拼接：

```
左邻居(N) → 左邻居(1) → 当前Chunk → 右邻居(1) → 右邻居(N)
```

拼接后当前 Chunk 的 `Description` 包含上下文信息，`Sequence` 和 `Location` 更新为最左边邻居的值。

#### 已知问题：side effect 导致 Sequence 被篡改

`JoinNeighborChunks`（`essearch.go:259`）在处理左邻居时存在 side effect：

```go
c.Sequence = left.Sequence
c.Location = left.Location
```

这会**直接修改原始 Chunk 指针的 Sequence 和 Location**。由于 Go 中 `[]*SearchType` 是引用语义，此修改会影响调用方持有的原始数据。

**影响**：

1. **右邻居搜索基于错误的 Sequence**：左邻居处理后 `c.Sequence` 已被篡改，后续右邻居查找 `c.Sequence + offset` 计算的是被篡改后的偏移，可能拼接错误的邻居内容
2. **Sequence 传递**：下一次遍历 `chunks` 的其他元素时，若该元素也在 `nbchunksMap` 中作为其他 chunk 的邻居被引用，读取到的是被篡改过的 Sequence
3. **最终结果中 Sequence 不准确**：`Resault()` 返回的 `QueryReferenceChunk.Sequence` 可能不是该 Chunk 的真实序号

但没有直接的修复方式是将 Sequence 的修改作用于一份 copy，而不是原始 `SearchType`。

注意：**此 bug 不会导致产生完全相同的两组数据**——`JoinNeighborChunks` 只遍历 `chunks` 一次，`expandedChunks` 长度等于 `len(chunks)`，不会凭空新增 chunk 项。它只会导致 Description 拼接异常和 Sequence 错误。

---

### Step 5: 第二次 Rerank

再次调用 `SortRerankChunk()`，对拼接后的 Chunk 进行重排序。

意义：拼接邻居后 Description 包含更多上下文，Rerank 模型能更准确地评估相关性。阈值过滤和 TopN 截断逻辑与 Step 2 相同。

---

### Step 6: TopM 截断

```go
if w.conf.Topm < len(nbsres) {
    nbsres = nbsres[:w.conf.Topm]
}
```

将结果截断到 TopM（默认 15），控制最终返回的 Chunk 总数。

---

### Step 7: 按文件聚合

`apps/kesearch/models/reranksearch/essearch.go:279` — `GroupByFileID()`

将 Chunk 按 `file_id` 分组，存入 `map[uint][]*SearchType`。

---

### Step 8: 摘要获取 + Rerank + 结果组装

#### 8.1 搜索文件摘要

`apps/kesearch/models/reranksearch/essearch.go:288` — `SearchFilesAbstract()`

当 `EnabelAbstract = true` 时（默认开启），根据聚合后的 `file_id` 列表，在 ES 中搜索 `type = "file_description"` 的文档，获取每个文件的摘要。

#### 8.2 摘要 Rerank

`apps/kesearch/models/reranksearch/essearch.go:344` — `RerankAbstract()`

调用 `GetRerank()` 对文件摘要进行重排序，只保留 `score >= RerankThreshold` 的摘要。

#### 8.3 最终组装

`apps/kesearch/models/reranksearch/essearch.go:373` — `Resault()`

遍历每个文件的 Chunk，组装 `chattype.QueryReference`：
- 通过 `forest.GetForestFileByID()` 查询文件信息（名称、创建时间、Uin）
- 通过 `user.GetUserByUin()` 查询用户信息（用户名、头像）
- 设置 `DataSourceType = "DC"`
- 组装 Chunk 列表（ChunkID、Sequence、Content/Description、Score、Location、Type）

---

## 4. 数据模型

### 4.1 请求

```
RerankSearchChunkRequest
  ├── BaseRequest
  └── Request: RerankSearchChunkEmbedRequest
        ├── Question  string           // 用户问题
        ├── ForestIDs []uint            // 知识森林 ID 列表
        ├── FileIDs   []uint            // 文件 ID 列表
        └── Config    *SearchConfig     // 可选搜索配置
```

### 4.2 响应

```
RerankSearchChunkResponse
  ├── BaseResponse
  └── Response: RerankSearchChunkEmbedResponse
        └── SearchResult: QueryReferenceList
              └── []*QueryReference
                    ├── FileID, FileName, ForestID, Uin
                    ├── Abstract, DataSourceType
                    ├── UserName, AvatarURL, CreatedAt
                    └── ChunkList: []QueryReferenceChunk
                          ├── ChunkID, Sequence, Type
                          ├── Content, ImageURL, Score, Location
```

### 4.3 SearchConfig

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `DescriptionWeight` | float32 | 0.3 | BM25 全文检索权重 |
| `EmbeddingWeight` | float32 | 0.7 | 向量相似度权重 |
| `EnabelAbstract` | bool | true | 是否获取文件摘要 |
| `EnableRerank` | bool | true | 是否启用 Rerank 重排序 |
| `Topn` | int | 30 | 每次 Rerank 后保留数 |
| `Topm` | int | 15 | 最终返回的 Chunk 上限 |
| `Topk` | int | 5 | 兜底 TopK |
| `NeighborSize` | int | 1 | 邻居 Chunk 数量 |
| `RerankThreshold` | float64 | 0.5 | Rerank 过滤阈值 |
| `FetchFactor` | int | 2 | ES 检索放大因子 |
| `FallBackToTopK` | bool | true | 过滤后为空是否兜底 |

### 4.4 SearchType（内部模型）

```go
type SearchType struct {
    ChunkID     string
    Type        string    // chunk/image/table/video/formula
    Score       float64   // ES 原始分数
    FileID      uint
    FileName    string
    ForestID    uint
    Sequence    int
    Description string    // 经过类型特殊格式化的描述文本
    ImageURL    string
    Location    [5]int    // 文档中的位置信息
    RerankScore float64   // Rerank 后的分数
}
```

## 5. 外部依赖

| 依赖 | 配置 Key | 说明 |
|------|----------|------|
| Elasticsearch | `knowledge.es` | 存储 Chunk 和文件摘要，提供全文+向量检索 |
| Embedding 服务 | `knowledge.embedding` | 将问题文本转为向量（Url、Key、ModelName） |
| Rerank 服务 | `knowledge.rerank` | 外部 Rerank 模型，评估问题与文档的相关性 |
| Chat Agent | — | `UserQueryRewrite` 调用，用于问题改写 |
| 文件服务 | — | `forest.GetForestFileByID()` 获取文件元信息 |
| 用户服务 | — | `user.GetUserByUin()` 获取用户信息 |

## 6. 错误处理与兜底策略

### 错误码

| 场景 | Code | Message |
|------|------|---------|
| 参数无效 | 400 | `kechat_invalid_params` |
| 配置无效 | 400 | `kesearch_reranksearch_config_error` |
| FQA 搜索失败 | 500 | `kechat_internal_error` |
| Wrapper 创建失败 | 500 | `kechat_internal_error` |
| 检索/Rerank 失败 | 500 | `kesearch_search_failed` |

### 兜底策略

1. **FQA 优先**：如果 FQA 问答对命中，直接返回，不走任何后续步骤
2. **配置为 nil**：自动使用 `GetDefaultConfig()` 默认配置
3. **Rerank 过滤后为空**：`FallBackToTopK=true` 时跳过阈值，直接取 TopK
4. **Rerank 服务返回 nil Data**：返回错误而非静默跳过
5. **默认配置兜底**：YAML 配置读取失败时使用硬编码默认值

## 7. 配置管理

两个搜索配置预设（均从 YAML 读取，失败时使用硬编码默认值）：

| 字段 | 默认搜索 (`reranksearchcfg`) | 图谱搜索 (`graphsearchcfg`) |
|------|------|------|
| DescriptionWeight | 0.3 | 0.3 |
| EmbeddingWeight | 0.7 | 0.7 |
| EnabelAbstract | true | true |
| EnableRerank | true | true |
| Topn | 30 | 100 |
| Topm | 15 | 20 |
| Topk | 5 | 50 |
| NeighborSize | 1 | 1 |
| RerankThreshold | 0.5 | 0.4 |
| FetchFactor | 2 | 3 |
| FallBackToTopK | true | true |

图谱搜索配置（`GraphSearchConf()`）用于知识图谱场景，在 `kechat/models/keqa/es_qa.go` 的 `GraphRerankChat()` 中调用，特点是更大的召回量（Topn=100）和更宽松的阈值（0.4）。

## 8. 已知问题

### 8.1 JoinNeighborChunks 的 Sequence 篡改（side effect bug）

**位置**：`essearch.go:259`

```go
c.Sequence = left.Sequence
c.Location = left.Location
```

直接修改了原始 `*SearchType` 指针的字段，导致：

- 右邻居搜索基于被篡改的 Sequence 进行偏移计算
- 最终结果中 Chunk 的 Sequence 可能不准确
- 不影响最终 Chunk 数量，不会凭空产生重复项

**修复方向**：操作 `SearchType` 的副本而非原始对象。

### 8.2 检索结果可能包含重复数据

**原因**：
- 外部算法服务入库时未使用稳定 `_id`，同一 table 重复切分/入库产生内容相同但 `_id` 不同的重复文档
- 整个检索流程（Step 1~8）无任何去重逻辑

**现象**：最终返回结果中出现两个 ChunkID 不同但 Content（表格数据+表格描述）完全相同的记录。

**排查路径**：
1. 查 ES 中该文件是否存在多条 `type = "table"` 且 `Table` 字段相同的记录
2. 检查外部算法服务（`workerServerURL + "/split"`）的 `_id` 生成策略
3. 考虑在 `Resault()` 或 `GroupByFileID()` 后增加基于 `FileID + Sequence + Type` 或 `Content` 的去重
