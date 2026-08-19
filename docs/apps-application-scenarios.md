# CoreKG apps 目录应用场景分析

> 基于 `apps/` 各应用 README 与源码整理。整个 monorepo 是知识库 / RAG 对话平台，共 21 个应用，可归纳为 **核心业务服务、内部运营服务、AI 引擎服务、基础设施/工具、独立 Algo 服务（Python/Node/Go）** 五类。

---

## 一、核心业务服务（知识平台主链路）

| 应用 | 定位 | 应用场景 |
|------|------|---------|
| **kecore** | 核心知识管理中枢 | 知识库（Forest）与文档全生命周期管理、文档解析/AI 分析、知识图谱、写作空间、资源级 RBAC 权限、配额订阅、支付订单、消息公告。是"知识长什么样、怎么存"的地方。 |
| **kechat** | AI 对话与问答引擎 | 所有智能问答的入口：知识库 RAG 问答、Agent 对话、直接模型对话、单文档问答（FileQA）、图谱问答、Excel 数据分析、Coze 集成、OpenAI 兼容透传。基于 Eino 框架、策略模式（ChatMode 按 BaseType 分发）。 |
| **kesearch** | 检索/搜索后端 | RAG 的核心搜索：知识库级语义搜索、全局搜索、多模态（文/图/视频）搜索、Chunk CRUD、Rerank 两阶段重排序、面向 Coze 的无登录搜索接口。数据来自 ES。 |
| **keapi** | 对外 API + MCP Server | 面向外部开发者的知识库 API（API Key 鉴权），同时暴露 **MCP StreamableHTTP Server**（21 个 Tool），让 LLM Agent 可以通过标准协议直接操作知识库。自定义端口 :8086，路径 /v3。 |
| **account** | 统一身份与组织管理 | 解决"谁在用、属于哪个组织、能做什么"：用户注册登录、多租户公司/组织、RBAC 权限、UIN 统一身份、API Key 管理、OAuth 绑定（微信）。是全站鉴权基础。 |
| **corekg** | 聚合单体（All-in-One） | 把 kecore + kechat + account + keapi + keparser + kesearch + wecom 打包进一个 HTTP 进程（共享 :8080），适用于不需要微服务拆分的部署场景；含 License 校验中间件和多种部署模式。 |

## 二、内部运营 / 后台类

| 应用 | 定位 | 应用场景 |
|------|------|---------|
| **admin** | 内部运营管理后台 | 面向运营人员的平台控制面板：员工/公司/终端用户管理、角色权限、License 管理、Prompt 模板、系统公告、登录配置、数据看板（跨库聚合统计）、HTTP 代理、LKX 销售线索收集。 |
| **keapp / keapp** | 知识应用 + Web 采集服务 | 提供"应用"化的知识体（Application 模板/资源/manifest），并附带 **网站爬虫 Worker**（NATS JetStream + 爬取规则 + 网页正文转 Markdown 入库），把网页变成可管理的知识资源。无 README，从源码归纳。 |
| **kesale** | 支付与订单库 | 不是独立服务，作为库被 corekg/kecore 初始化：订单创建、微信支付对接、支付回调（Redis 防重）、每小时订单验证、业务回调解耦。 |
| **ketask** | 任务调度 / 部署模式切换服务 | 提供 knowledge 命名空间路由：部署模式切换（PrivateEnv/DeployMode）、metrics 监控，并支撑多种 worker 任务（PDF 解析、文本切块、索引导入、视频提取等）。无 README，从源码归纳。 |
| **kecore 内部** | 异步文件处理管线 | CopyTask → Doc2PDF → Parse → MindMap/Analysis/AI提取 → Knowledge → Graph 的 8 步管线，通过 pkgs/task 队列调度。 |

## 三、AI / LLM 引擎库

| 应用 | 定位 | 应用场景 |
|------|------|---------|
| **kellm** | LLM 代理库（Gateway） | OpenAI 兼容的 /chat/completions 接口，按模型名路由到上游 LLM，支持 SSE 流式、多 Driver 架构。被 kechat（透传）、keapi（消息格式）引用，不独立运行。 |
| **einonodes** | Eino Agent 节点库 | 基于字节跳动 Eino 框架的节点组件（Intent/Dataloader/Planner/Reporter/Executor/Branch），含外部数据源搜索 Graph，被 kechat 用于聊天管线。 |
| **nodes** | Eino ChatModel 工厂 | 统一的 LLM ChatModel 创建工厂，被 kechat 引用。 |

## 四、基础设施 / 初始化 / 工具

| 应用 | 定位 | 应用场景 |
|------|------|---------|
| **keinit** | 部署初始化 CLI | 平台的"安装向导"：初始化 MySQL 表、ES 索引、MinIO 桶、聊天模型配置、API Key、系统设置，并执行数据迁移脚本、返回 K8s 集群 UID（License 绑定用）。 |
| **webfetch** | 单 URL 正文读取服务 | 独立部署，对单 URL 做 SSRF 校验 + Chromium 渲染 + 正文提取（markdown/纯文本），通过 API Market 暴露，供外部调用。 |
| **websearch** | 原子 Web 搜索服务 | 独立部署，多 Provider 路由、Profile Pool、缓存、opaque cursor，通过 API Market 暴露（配合 webfetch 构成"先搜后读"）。 |
| **worker** (Node) | TS 算法 Worker | TypeScript worker 二进制（commander），包含 PDF 提取、文本切块、分析、思维导图、上下文构建、视频提取、索引、拷贝等命令，批量建包的 kealgo worker。 |
| **workflow** | Coze Studio 工作流服务 | 基于开源 [coze-dev/coze-studio](https://github.com/coze-dev/coze-studio) 的 LLM 工作流/Agent 编排后端（Hertz 框架），含 agent/workflow/knowledge/memory/plugin 等领域模块，对接到 CoreKG 知识库。 |

## 五、Python 数据管道（独立的摄入管线）

| 应用 | 定位 | 应用场景 |
|------|------|---------|
| **pipeline** | 文档知识摄入/切块管线 | 两个 Python worker 微服务：`corekg_analyser`（S3 下载 → MinerU 解析 PDF/图片 → Markdown 上传回 S3）、`corekg_chunk`（拉取 MD → 多种策略切块 → 图片/表格大模型增强 → Embedding → 写入 ES）。以任务队列 Worker 模式运行，支持 Docker/Nuitka 加密私有化交付。 |

---

## 关键架构关系

```
                  ┌────────────── 对外（开发者/第三方/Agent）──────────────┐
                  │   keapi (REST/OpenAI兼容/MCP)                        │
                  └──────────────────────┬───────────────────────────────┘
                                          │
  身份/权限  account ◄──── 鉴权基础 ─────┼─────────►  admin (运营后台)
                                          ▼
        kecore (知识库)◄──RAG──► kesearch (检索) ◄──► kechat (对话引擎)
              ▲                                │
        async 管线 (ketask / pipeline Py /      │
        worker Node / webfetch 采集...)          ▼
                                        kellm / einonodes / nodes / workflow
```

**核心关系：**
- **account / admin** 负责身份权限与平台运营；
- **kecore** 管知识资产的存储与生命周期；
- **kesearch** 提供检索后端；
- **kechat** 负责对话编排，通过 kellm / einonodes / nodes 完成 LLM 调用与 Agent 流程；
- **keapi / corekg** 是两种对外/部署形态（独立微服务对外 API 与聚合单体）；
- **keinit、ketask、pipeline** 是初始化与异步处理承载；
- **websearch / webfetch** 为外部数据接入；
- **workflow** 是一个独立的 Coze 风格工作流编排服务。

---

## 附：补充说明

- **keapp / ketask 无 README**，以上分析基于源码归纳，若需精确接口清单可进一步补充。
- 应用分层约定统一为：`internal/apis/apis.go`（路由注册）→ `internal/apis/<ctl>` handler → `services/svc*` → `models/*`（GORM），DTO 在 `internal/dto`。
- 构建/运行始终传入 `APP=<name>`，例如 `make run APP=keapi ENV=test`。
