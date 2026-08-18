# EinoNodes

## 定位

EinoNodes 是基于字节跳动 Eino 框架（cloudwego/eino）构建的 **AI Agent 节点库**。不是独立服务，而是提供可复用的 Agent/Graph 节点组件，用于构建 RAG 聊天流程图。

## 核心组件

### nodebase - 基础类型

- `RecordList` 基础类型
- 注册 Eino compose 自定义值合并函数

### qachatnodes - QA 聊天节点集合

| 节点 | 说明 |
|------|------|
| IntentRecognizer | 意图识别 |
| DataLoader | 数据加载（Gmail/Slack/Google Drive/Confluence） |
| Planner | 规划器 |
| Reporter | 报告生成 |
| Branch | 分支路由 |
| Executor | 节点执行器 |

提供完整的外部数据源搜索 Graph（ExternalDataToolsGraph），串联 Gmail -> Slack -> Google Drive -> Confluence。

### einodemo - 演示用例

5 个完整的 Eino 示例：基础 ChatModel -> Template -> Agent Chain -> Agent Graph -> 通用 Graph。

## 与其他服务的关系

- 被 `apps/kechat` 引用 - Eino 图节点用于聊天管线
- 依赖 `cloudwego/eino` - Eino AI Agent 框架
- 依赖 `cloudwego/eino-ext` - Eino 扩展组件
