
# 问题概述

`/chat.ChatQuestionStream` 接口 `base_type=standard` 分支，当 SummaryAgent 阶段 LLM 调用超时时，前端流式接收一部分内容后停止、接口不返回错误，刷新页面后流式结果未入库（DB 为 `status=answered`/`answer=""`）。

> **当前实际状态（已验证）**：requestID `74a971acfe6f43fa97ab40b2620cce0f` 复现，SummaryAgent 调 deepseek-v3（经 `https://yygu.cn/v3/llm.chat` 代理）Stream 请求 300.009s 无首 chunk，被 `DefaultChatModelTimeout` 击穿，错误被吞成 `err=nil`，`summaryLength=0`，最终落库 `answered`/空。

---

## 现象

1. 前端流式接收一部分内容（ReactAgent 阶段的 tool 调用/检索中间输出）后停止
2. 接口 HTTP 200 正常返回，不报错（`respsize` 可达 500KB+）
3. 刷新页面，DB 中该 question 为 `status=answered`、`answer=""`（非流式已发送内容，亦非错误提示）

---

## 触发条件

- `base_type=standard`（走 `ForestChatMode`）
- ReactAgent 触发 `compose.ErrExceedMaxSteps`（`maxStep=4`），强制走 SummaryAgent
- SummaryAgent 构造的请求体过大（`taskHistory` 含 2 轮 `forest_search_tool` 返回、最多 30 个 chunk 全文 JSON），外部 LLM prefill 超过 `DefaultChatModelTimeout` 无首 chunk

---

## 研发版本

release-2.13

## 根因分析

> 以下行号基于修复前代码。

### 完整调用链

```
svcchat.ChatQuestionStream (svcchat/chat.go:26)
  └─ agentwrapper.ChatWrapper.Run (chat/wrapper/wrapper.go:28)
       └─ ForestChatMode.Run (chat/modes/forest.go:39)
            └─ ReactAgentService.Handler (pkgs/einotools/service/react_agent_service.go:25)
                 ├─ reactAgent.Run → ErrExceedMaxSteps (react_agent_service.go:86)
                 │    └─ force enable summary
                 └─ summaryAgent.RunSummarizeResult (pkgs/einotools/agent/summary_agent.go:109)
                      └─ preBuildMsg 构造巨型 UserMessage (summary_agent.go:59-81)
                           └─ taskHistory = 5 条消息原样拼接（含 30 chunk 全文）
                           └─ history_dialogue = 全部历史 QA
                      └─ runnable.Stream → LLM 无首 chunk → Client.Timeout 击穿
                           └─ chat_model.go:608 sw.Send(nil, err)
                           └─ base_agent.go:186 handleStreamOutput err → return（不调 onEnd）
                           └─ summary_agent.go:125 循环靠 pipe close 转 EOF 退出（err 丢失）
                      └─ return "", nil ← 错误被吞
```

### 300s 无响应根因：请求体过大

`SummaryAgent.preBuildMsg`（`summary_agent.go:59-81`）将 ReactAgent 的 5 条历史消息（含 2 轮 `forest_search_tool` 返回，每轮至多 15 个 chunk 段落全文）**原样拼接**进 `taskHistory`，无截断。叠加 `history_dialogue`（全部历史 QA）和 4KB 的 `ForestKnowledgeSummarySystemPrompt`，合成单条巨型 UserMessage，可达数万到十万级 token。

model_id=1 是 deepseek-v3 经外部 API 代理（`chat.sql:34`），超大 prompt 的 prefill 极慢，300s 内无首 chunk。证据：`summaryLength=0`/`messages=0` 证明一个 chunk 都没到。

### 错误吞没机制：为什么 err=nil

两个缺陷叠加：

1. **`base_agent.go:186-189`**：`handleStreamOutput` 出错时 `fmt.Printf + return`，不调 `onEnd` → Memory 不 Flush（`messages=0`），不发 `isFinal=true`
2. **`summary_agent.go:125-130`**：stream 循环只处理 `io.EOF`，非 EOF 错误被忽略，靠 pipe close 转 EOF 正常 break → `err` 丢失，继续走 `summary_agent.go:139 agent.State = FINISHED` + `lastMsg == nil → return "", nil`

### 内容不入库机制：为什么 DB 是 answered/空

- `Handler` 返回 `("", nil)` → `react_agent_service.go:145 Handler success with summary: summaryLength=0`
- `forest.go:149-152 chatResult.Answer = ""`（空字符串）
- `wrapper.go:67 Question.Source.Answer = ""`
- `chat.go:140` err==nil **不进**错误兜底分支（不会用错误提示覆盖）
- defer 落库：`status=answered`（因 `chatResult.Status=Answered`，`forest.go:151`）、`answer=""`

---

## 日志排查方式

### 关键陷阱

**最关键的证据 `读取失败: ...`（`base_agent.go:187`，修复前）用的是 `fmt.Printf` 而非 `logs.*`，不在 JSON 结构化日志里。用 `"lvl":"ERROR"` 过滤会漏掉它，必须用裸文本搜。**

> 修复后此日志改为 `logs.ErrorContextf` 输出 `[handleStreamOutput] read stream failed`，可被 `"lvl":"ERROR"` 搜到。但历史日志仍是裸文本。

### 关键字清单

| # | 关键字 | 来源 | 含义 |
|---|---|---|---|
| 1 | `读取失败: failed to receive stream chunk from OpenAI` | `base_agent.go:187`（修复前 printf 裸文本） | **LLM 底层超时铁证**，非结构化日志 |
| 1b | `[handleStreamOutput] read stream failed` | `base_agent.go:187`（修复后 logs） | 修复后的结构化版本 |
| 2 | `context deadline exceeded` / `Client.Timeout` / `reading body` | 同上 | 超时类型判定 |
| 3 | `exceed max steps, force enable summary` | `react_agent_service.go:88` | 触发 SummaryAgent 的前提条件 |
| 4 | `run SummaryAgent start:` | `react_agent_service.go:118` | Summary 阶段开始时间戳 |
| 5 | `run SummaryAgent finished: err=<nil> summaryLength=0` | `react_agent_service.go:120` | **错误被吞标志**（修复前）：err=nil + summaryLength=0 |
| 5b | `run SummaryAgent finished: err=<非nil>` | `react_agent_service.go:120`（修复后） | 修复后错误正确上抛 |
| 5c | `summary agent stream receive failed` | `summary_agent.go`（修复后） | 修复后 SummaryAgent 上抛的错误文本 |
| 6 | `Handler success with summary: summaryLength=0` | `react_agent_service.go:145` | 走成功路径（修复前，导致 status=answered） |
| 6b | `Handler error` / `[forestChatMode] Agent Handler error` | `forest.go:142`（修复后） | 修复后错误上抛到 forest 层 |
| 7 | `questionEntity:` | `svcchat/chat.go:74` | 最终落库 JSON（含 status/answer） |
| 8 | `chat.ChatQuestionStream` | `chat_api.go` | 请求入口（拿 requestID/latency/respsize） |

### 检索步骤

**Step 1 — 定位疑似请求（找 requestID）**

```bash
grep "chat.ChatQuestionStream" <logfile>
```

看每条的 `latency`（异常 > 300s）和 `respsize`（异常大，说明前期流式有数据）。记下 `reqid`。

**Step 2 — 验证 LLM 底层超时（铁证）**

```bash
# 修复前日志（裸文本）
grep -E "读取失败|context deadline exceeded|Client.Timeout|reading body" <logfile>

# 修复后日志（结构化）
grep -E "\[handleStreamOutput\] read stream failed" <logfile>
```

⚠️ 修复前不可用 `grep '"lvl":"ERROR"'` 搜此条——它是 `fmt.Printf` 裸文本，无 JSON 包装。

**Step 3 — 用 requestID 串全程，看走了哪条分支**

```bash
grep "<reqid>" <logfile> | grep -E "ReactAgentService|questionEntity|exceed max steps|Handler success|ChatWrapper"
```

**Step 4 — 看落库最终值（确认"内容没存"）**

```bash
grep "<reqid>" <logfile> | grep "questionEntity:"
```

看 JSON 里 `"status"` 和 `"answer"`：
- `status=answered, answer=""` → Summary 超时被吞路径（修复前的问题现象）
- `status=error, answer="...错误提示..."` → 修复后正确上抛错误，或 ReactAgent 阶段超时

**Step 5 — 确认超时阶段耗时**

对比两条日志时间差：
- `run SummaryAgent start:` → `run SummaryAgent finished:` 差 ≈ 超时时间 → Summary 阶段超时
- `run ReactAgent start:` → `run ReactAgent finished:` 差 ≈ 超时时间 → ReactAgent 阶段超时

**Step 6（可选）— 评估请求体大小**

```bash
grep "<reqid>" <logfile> | grep -iE "rerank|forest.*search|chunk"
```

看森林检索召回的 chunk 数量/大小，辅助判断是否有超大 prompt。

### 判定规则

| 关键字组合 | 结论 |
|---|---|
| `exceed max steps` + `run SummaryAgent finished: err=<nil> summaryLength=0` + `Handler success with summary: summaryLength=0` + `questionEntity: status=answered answer=""` | **修复前的问题**：Summary 超时被吞 |
| `exceed max steps` + `summary agent stream receive failed` + `run SummaryAgent finished: err=<非nil>` + `questionEntity: status=error` | **修复后**：Summary 超时正确上抛 |
| `ReAct agent execution failed:`（无 `exceed max steps`） + `questionEntity: status=error` | ReactAgent 阶段超时，err 上抛 |
| `ReactAgentService ReactAgent canceled` / `SummaryAgent canceled` | 走 cancel 友好分支（ctx.Canceled，非超时） |

### 真实案例

requestID `74a971acfe6f43fa97ab40b2620cce0f`（日志 `corekg_last10min.log`）：

| 行 | 时间 | 关键内容 |
|---|---|---|
| 326 | 16:42:57.998 | `exceed max steps, force enable summary: maxStep=4 messages=5` |
| 327 | 16:42:57.998 | `run SummaryAgent start: isStream=true historyMessages=5` |
| 1003 | 16:47:54.x | `读取失败: failed to receive stream chunk from OpenAI: context deadline exceeded` |
| 1005 | 16:47:58.007 | `run SummaryAgent finished: err=<nil> summaryLength=0 messages=0` |
| 1009 | 16:47:58.007 | `Handler success with summary: summaryLength=0` |
| 1014 | 16:47:58.021 | `questionEntity: "status":"answered","answer":""` |

SummaryAgent 耗时 300.009s，正好等于修复前的 `DefaultChatModelTimeout=300s`。

---

## 修复方案

### 修复后行为对比

| 场景 | 修复前 | 修复后 |
|---|---|---|
| 接口返回 | HTTP 200 | HTTP 200（不变，`chat.go:151 return res,nil`） |
| 日志 err | `err=<nil>`（吞没） | `err=<非nil>`（`summary agent stream receive failed`） |
| DB status | `answered` | `error` |
| DB answer | `""`（空） | 本地化错误提示（`chat.go:143-146` 兜底） |
| 前端 | 卡住无结束标志 | 收到 `WriteContent` 推送的错误提示 |
| 诊断 | 搜 `读取失败` 裸文本 | 搜 `"lvl":"ERROR"` + `[handleStreamOutput]` |

### 改动清单

#### P0-1：修复 `handleStreamOutput` 错误处理

- **文件**：`pkgs/einotools/agent/base_agent.go`
- **改动**：出错时调 `onEnd(full)` 让 Memory Flush 已累积内容（若有 chunk 到达过），发 `isFinal=true`；用结构化日志替换 `fmt.Printf`
- **效果**：超时后 Memory 能保留已到达的 chunk 内容；前端能收到结束标志（若有 chunk 到达过）；日志可被 `"lvl":"ERROR"` 搜到

#### P0-2：修复 `summary_agent` stream 循环错误处理

- **文件**：`pkgs/einotools/agent/summary_agent.go`
- **改动**：stream 循环显式处理非 EOF 错误，上抛 `fmt.Errorf("summary agent stream receive failed: %w", err)`
- **效果**：超时错误上抛到 `react_agent_service.go:119`，走错误分支，而非伪装成成功

#### P0-3：修复 `react_agent` stream 循环错误处理（同款 bug）

- **文件**：`pkgs/einotools/agent/react_agent.go`
- **改动**：stream 循环显式处理非 EOF 错误，上抛 `fmt.Errorf("react agent stream receive failed: %w", err)`
- **效果**：防止 ReactAgent 阶段超时也被吞

#### P1-1：调长默认超时时间

- **文件**：`apps/kechat/chat/modelhelper/tool_calling_chat_model.go`
- **改动**：`DefaultChatModelTimeout` 从 `300s` 调整为 `600s`
- **效果**：给外部 LLM 更多 prefill 时间，减少超大 prompt 场景的超时概率

### 验证

- **编译验证**：`go build ./apps/kechat/... ./pkgs/einotools/...` 通过
- **复现验证**：用原 requestID 的 session（190）或等大 prompt 重现，确认：
  - 日志出现 `run SummaryAgent finished: err=<非nil>`
  - `questionEntity: "status":"error"`（非 `answered`）
  - 前端收到错误提示而非空白卡住

---

## 未覆盖项（已知，本次未修）

- **前端结束标志**：`SSEPrinter.Close`（`sse_printer.go:92`）/ `WriteContent` 仍不调 `sseClient.Stop`，前端 `BlockRead` 靠空闲超时结束。如需即时结束，需改 `SSEPrinter.Close` 调 `Stop`。
- **请求体过大治本**：`summary_agent.go:59-81` 的 `taskHistory` 仍原样拼接 30 个 chunk 全文。调长超时是缓解，不是根治。如需根治，需精简 tool 消息只留 fileID/sequence 摘要。
