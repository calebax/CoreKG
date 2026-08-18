# KEChat 服务

## 服务定位

KEChat 是 CoreKG 平台的 **AI 对话与问答服务**，是平台所有智能问答交互的核心引擎。它驱动知识库 RAG 问答、Agent 对话、直接模型对话、单文档问答、图谱问答、Excel 数据分析、Coze 集成、以及 OpenAI 兼容的 LLM 透传。

## 核心业务域

| 业务域 | 解决什么问题 | 关键能力 |
|--------|-------------|---------|
| 知识库 RAG 问答 | 基于文档的智能问答 | 检索增强生成、引用溯源、关键词提取 |
| Agent 对话 | 可配置的 AI Agent 交互 | Prompt/角色扮演/工作流 Agent、版本管理、权限发布 |
| 直接模型对话 | 不依赖知识库的纯 LLM 对话 | 工具调用、附件支持、ReAct 循环 |
| 单文档问答 (FileQA) | 针对单个文档的问答 | 上传文档、异步解析、对话 |
| 图谱问答 | 基于知识图谱的问答 | NebulaGraph 检索 + RAG 增强 |
| Excel 数据分析 | 表格数据的自然语言分析 | ReAct Agent + 工具调用 |
| Coze 集成 | 对接 Coze 平台 | 模型同步、Agent 映射、外部聊天流代理 |
| 外部 API | 第三方集成 | JWT Token 认证、OpenAI 兼容端点 |
| LLM 透传 | 统一模型调用 | 通过 kellm 代理到上游 LLM |
| 问题扩展 | 提升检索质量 | LLM 改写原始查询 |

## 核心业务概念

- **ChatSession** - 对话会话，BaseType 决定聊天模式（standard/agent/model/excel/mysql/graph/react_excel/forest_agent/graph_search）
- **ChatQuestion** - 单轮问答，存储在 ES（ke_chat_history）
- **ChatModel** - LLM 模型配置（API Key、URL、Provider、Function Call 支持）
- **ChatAgent** - Agent 定义（类型、发布范围、Coze 映射、版本）
- **ChatMode（接口）** - 可插拔的聊天策略，按 BaseType 分发
- **QueryReference** - RAG 检索引用，关联文件和 Chunk

## 多模式聊天架构

新版 `chat/` 包实现策略模式（`ChatMode` 接口），按会话 BaseType 分发：

| 模式 | 说明 |
|------|------|
| ForestChatMode | 知识库 RAG 问答，集成检索工具 |
| DirectModelChatMode | 直接 LLM 对话，支持工具调用/附件/ReAct |
| GraphSearchChatMode | 图谱搜索问答 |
| ExcelChatMode | Excel 数据分析 |
| ForestAgentChatMode | 知识库 + Agent 组合模式 |

旧版 `models/qachat/ChatWrapper` 仍在过渡期用于 agent/mysql/graph/external_data 模式。

## API 路由分组

路由定义在 `internal/apis/apis.go`，基础路径 `/v3`。

| 分组 | 代表接口 | 认证要求 |
|------|---------|---------|
| 会话管理 | NewChatSession, ListChatSession, RenameChatSession, RemoveChatSession, MoveSession | 需登录 |
| 流式问答 | SubmitChatQuestionStream, ChatQuestionStream, StopChat | 需登录 + QA配额 |
| Agent 管理 | CreateAgent, UpdateAgent, DeleteChatAgent, ListChatAgent, TestAgent | 需登录 |
| Agent 版本 | ListAgentVersion, ChooseAgentVersion | 需登录 |
| Agent 权限与发布 | GetAgentWithPerm, ModifyChatPermSet | 需登录 + 权限 |
| Agent 统计 | GetLatestAgents, GetAgentQuestionExcel, GetAgentQuestionCount | 需登录 |
| 模型管理 | ListModel, CreateModel, UpdateModel, DeleteModel, ModelTest, BindCozeModel | 需登录 |
| Coze/外部集成 | ExternalToken, NewExternalChatStream, ChatAgent/chat/completions | JWT/Coze Token |
| OpenAI 兼容 | Agent/chat/completions, LLM/chat/completions | 视接口而定 |
| 文件/附件 | UploadImage, UploadAttachment, GetFileSession | 需登录 |
| 单文档问答 | ListFileQA, FileChat, DeleteFileQA | 需登录 + QA配额 |
| 图表/画布 | SaveChartCanvas, GetChartCanvas | 公开 |

## 代码架构

```
apps/kechat/
├── app.go                   # Routers/Migrates/RunJob
├── cmd/
│   ├── main.go              # Cobra 入口，初始化 DB/ES/Nebula/Storage/Providers
│   └── init.go              # 数据库 + Redis 初始化
├── internal/
│   ├── apis/                # Handler 层（32个文件，*_api.go + *_biz.go 配对）
│   │   └── apis.go          # 统一路由注册
│   ├── dto/                 # 请求/响应 DTO
│   └── docs/                # Swagger 自动生成
├── mds/                     # 中间件
│   ├── chatmds.go           # QA 配额检查、Agent 权限、外部 JWT 认证
│   └── cozemds.go           # Coze Token 解析
├── services/                # 业务逻辑层
│   ├── svcchat/             # 核心聊天流式处理、历史、问题扩展
│   ├── svcsession/          # 会话操作
│   ├── svcagent/            # Agent 列表
│   ├── svcmodel/            # 模型 CRUD、Coze 模型同步
│   ├── svccoze/             # Coze 平台集成
│   ├── svcfile/             # 附件解析
│   └── svcchart/            # 图表画布
├── models/                  # 数据访问层
│   ├── chat/                # 核心 DAO
│   ├── chattype/            # 类型定义
│   ├── chatquestion/        # 问题 CRUD + ES 历史
│   ├── chatsession/         # 会话 CRUD + LLM 命名
│   ├── chatagent/           # Agent CRUD + 版本 + 权限
│   ├── qachat/              # 旧版 ChatWrapper 编排
│   ├── llmchat/             # LLM 流式、工具调用、SSE 写入
│   └── coze/                # Coze API 客户端
└── chat/                    # 新版聊天架构
    ├── core/                # ChatMode 接口、ChatContext、ChatResult
    ├── modes/               # 可插拔聊天模式实现
    ├── wrapper/             # ChatWrapper：按 BaseType 分发
    ├── prompt/              # 各模式的 Prompt 模板
    └── modelhelper/         # 模型配置辅助
```

## 技术要点

- **SSE 流式响应**：所有主要聊天端点使用 Server-Sent Events
- **Eino ADK 集成**：使用 CloudWeGo Eino 框架做 Agent 编排、工具调用循环
- **Coze 双向同步**：模型同步到 Coze，Agent 映射到 Coze 工作流
- **异步附件解析**：PDF/DOC/PPT/OFD 转 Markdown 通过任务管线
- **QA 配额**：SaaS 模式下按公司检查问答配额，私有化跳过
- **ES 存储聊天记录**：问题和答案存储在 ES，支持搜索、分析和 Excel 导出
- **会话名称自动生成**：问答后异步调用 LLM 摘要更新会话名称
- **多库**：Chat、Core、Coze、Knownow

## 本地开发

```bash
make local APP=kechat
make run APP=kechat ENV=test
make generate-docs APP=kechat
```

## 与其他服务的关系

- 依赖 `apps/account/accountmds` - 认证中间件
- 依赖 `apps/kecore` - 部署模式、文件存储、NebulaGraph、配额检查、关键词提取、异步任务
- 依赖 `apps/kesearch` - ES 搜索、AI 搜索工具
- 依赖 `apps/keparser` - RAG 任务载荷
- 依赖 `apps/kellm` - LLM 透传
- 依赖 `apps/einonodes` - Eino 图节点
- 依赖 `pkgs/einotools` - Eino Agent 框架、工具、SSE Printer
- 依赖 `pkgs/task` - 异步任务
- 依赖 `pkgs/connectors` - 外部 Provider 初始化
