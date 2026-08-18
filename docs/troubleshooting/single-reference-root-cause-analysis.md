# 单引用/引用 Chunk 不准 当前状态

这是一份历史问题页。完整当前链路、日志关键字和本次 MR 修复请先看 [kechat-forestchat-rag-chain.md](./kechat-forestchat-rag-chain.md)。

## 旧结论里已经过时的部分

旧文档把“只引用一篇资源”主要归因到 TopM 被调小、两次 rerank、邻居拼接放大等历史状态。当前 MR 分支里，chunk 检索链路已经调整：

```text
ES 召回 -> 查邻居 -> 当前 chunk + 上文 + 下文 -> 单次 rerank -> TopM -> GroupByFileID -> Resault
```

旧文档中以下描述不再代表当前代码：

1. `JoinNeighborChunks` 会用左邻居覆盖当前 chunk 的 `Sequence/Location`。
2. 当前 chunk 前面一定先拼左邻居。
3. 先 rerank 一次再查邻居。
4. 引用不准只能通过降低 `NeighborSize` 或直接截断解决。

## 本次真实排查到的问题

用户问题“未按规定确认收入的常见表现形式有哪些”里，目标 chunk 是 `sequence=44`。日志显示：

1. ES 第一阶段能召回目标文件和目标 chunk。
2. rerank 对目标 `sequence=44` 打分很低，而一些目录/不相关 chunk 分更高。
3. 仅依赖 rerank topK fallback 会把不相关 chunk 返回。
4. 初版 keyword fallback 如果用拼接后的 `Description` 打分，会因为邻居文本含关键词，把相邻 `sequence=46` 误判得比中心 `sequence=44` 更相关。

## 当前已完成的修复

1. rerank 的 `text_1` 改用用户原始问题，而不是模型 function call 生成的查询句。
2. ES 召回后先查邻居并拼接，再做唯一一次 chunk rerank。
3. `JoinNeighborChunks` 改为“当前 chunk + 上文 + 下文”，并保留中心 chunk 的 `Sequence/Location`。
4. 增加 chunk 排序、保留、丢弃、rerank 分、ES 分、keyword fallback 分的日志。
5. rerank 阈值全空时增加 keyword fallback。
6. keyword fallback 改为使用中心 chunk 原文，避免邻居内容把相邻 chunk 错判为命中。

## 当前排查方式

按 requestID 搜索：

```bash
grep "<request_id>" <logfile> | grep -E "chunk-rank|chunk-selection|keyword fallback|GetRerank|ForestSearchTool.invoke|UserQueryRewrite"
```

优先确认四件事：

1. `ForestSearchTool.invoke` 的 `question` 和 `original_question` 分别是什么。
2. `NewRerankSearchWrapper UserQueryRewrite` 生成了什么 ES query。
3. 目标 chunk 在 `step1_ES_search_order`、`step3_join_neighbor_order`、`rerank_threshold_selection` 中分别处于什么位置。
4. 如果触发 keyword fallback，目标 chunk 的 `hit_keywords`、`rare_keywords`、`keyword_score` 是否合理。

## 当前遗留风险

1. 如果目标 chunk 没进入 ES 第一阶段候选，后续 rerank/fallback 都救不回来。
2. 如果 keyword fallback 没有候选，`FallBackToTopK=true` 时仍可能回到旧 topK 兜底。
3. keyword fallback 的“最稀有关键词”约束较硬，后续可改成多关键词覆盖率、业务关键词白名单或单独关键词提取 agent。
