# Troubleshooting Index

这里放 kechat/ForestChat/RAG 链路的排查材料。当前排查时优先看：

1. [kechat-forestchat-rag-chain.md](./kechat-forestchat-rag-chain.md)：当前版本的完整链路、function call 时机、ES/rerank/fallback 逻辑、日志定位方式。
2. [summary-agent-timeout-swallowed.md](./summary-agent-timeout-swallowed.md)：SummaryAgent 流式超时和空答案落库问题。
3. [rerank-token-overflow.md](./rerank-token-overflow.md)：历史 token overflow 问题的当前状态摘要。
4. [single-reference-root-cause-analysis.md](./single-reference-root-cause-analysis.md)：历史“只引用一篇资源”问题的当前状态摘要。

旧排查材料里出现的“双 rerank 流程”“左邻居先拼接并改写 sequence/location”等描述已经不再代表当前 MR 分支行为。当前链路以 `kechat-forestchat-rag-chain.md` 为准。
