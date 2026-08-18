# Rerank Token Overflow 当前状态

这是一份历史问题页。完整当前链路、日志关键字和本次 MR 修复请先看 [kechat-forestchat-rag-chain.md](./kechat-forestchat-rag-chain.md)。

## 旧结论里已经过时的部分

旧文档曾按“ES 召回 -> 第一次 rerank -> 查邻居 -> 拼接 -> 第二次 rerank”分析 token overflow。当前 MR 分支已经调整为：

```text
ES 召回 -> 查邻居 -> 当前 chunk + 上文 + 下文 -> 单次 rerank -> TopM -> 返回引用
```

因此，旧文档中这些描述不再代表当前行为：

1. “第一次 rerank 后 fallback topK 决定邻居集合”。
2. “第二次 rerank 才发生 token overflow”。
3. “JoinNeighborChunks 先拼左邻居，再拼当前 chunk，并改写 sequence/location”。

## 仍然有效的问题模型

Token overflow 的本质仍然是：多个长 chunk 被拼接后进入 rerank 或 SummaryAgent，上下文超过模型限制。

当前仍需关注两类风险：

1. `GetRerank` 的 `text_2` 是拼接后的 chunk description 列表，没有 token 预检或截断。
2. SummaryAgent 会把 ReAct 阶段的 ToolMessage 拼进 `taskHistory`，chunk 很多或文本很长时仍可能导致最终总结模型超时。

## 当前已完成的缓解

本次 MR 对检索侧做了这些缓解：

1. 去掉前置 chunk rerank，减少一次模型调用和一次不稳定过滤。
2. `JoinNeighborChunks` 改为“当前 chunk + 上文 + 下文”，降低开头邻居干扰。
3. 保留中心 chunk 的 `Sequence/Location`，避免引用定位被邻居覆盖。
4. rerank 阈值全空时先走 keyword fallback，再考虑旧 topK fallback。
5. keyword fallback 用中心 chunk 原文打分，避免邻居文本污染。

## 后续建议

1. 给 `GetRerank` 增加输入 token 预检和截断/分批。
2. 对 `SummaryAgent.taskHistory` 做 chunk 内容压缩或引用摘要化。
3. 对 `FallBackToTopK` 做更严格的开关或日志告警，避免 rerank 明显异常时返回不相关 chunk。
