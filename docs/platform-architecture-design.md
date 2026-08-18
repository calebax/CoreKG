# CoreKG Platform 架构设计

> **文档状态：** 规划提案
>
> **定位说明：** 本文档描述 CoreKG Platform 的目标架构愿景，是面向未来的演进方向，**非当前实现现状**。当前产品以 CoreKG（仓库 `roc`，Go module `github.com/insmtx/corekg`）单体仓库形式迭代，已有 kecore、kechat、keapi、kesearch、keparser、keworker 等应用模块。本文档描述的是 CoreKG 的架构升级目标与演进路径。

---

我认为你的产品已经到了一个节点，**不能按照"一个前端 + 一个后端"去设计了，而应该按照"平台（Platform）+ 能力（Capability）+ 应用（Application）"来设计。**

这样最终可以实现：

- 一个安装包即可部署（社区版）
- 企业版按模块部署
- 每个模块可以独立升级
- 后续还能开放 SaaS、私有化、OEM

这也是现在很多 AI 平台（Dify、Flowise、OpenWebUI、Cursor Backend、Confluence）的发展方向。

---

## 三层架构

不是传统 MVC。

而是：

```
                    Applications
            ┌───────────────────────────┐
            │ 网站助手   维修助手  AI客服 │
            │ 培训助手   招标助手 ...... │
            └────────────▲──────────────┘
                         │
                   Capability Layer
┌────────────────────────────────────────────────────┐
│ Knowledge │ OCR │ Parser │ Search │ RAG │ Workflow │
│ Embed     │ ES  │ Chunk  │ Crawl  │ Auth│ Storage  │
└────────────────────────────────────────────────────┘
                         ▲
                         │
                    Platform Layer
┌────────────────────────────────────────────────────┐
│ 用户 │ 权限 │ 项目 │ 文件 │ 组织 │ API │ Billing │
└────────────────────────────────────────────────────┘
```

这样三层职责非常明确。

---

## 前端不要做成一个系统

很多产品都是：

```
Dashboard

知识库

工具

网站助手

维修助手

设置
```

最后越来越大。

我建议做成：

```
Portal

↓

加载不同的 App
```

类似于：

```
Notion

↓

Database
Calendar
Mail
AI
```

或者

```
Figma

↓

Design

Slides

Whiteboard

Dev Mode
```

其实都是不同应用。

---

### 前端建议采用 Micro Frontend（微前端）

例如：

```
Portal

├── Dashboard

├── Knowledge App

├── Tools App

├── Website Assistant

├── Maintenance Assistant

├── Admin

├── Developer Center
```

每一个：

都是一个独立 Vue 项目。

例如：

```
apps/

knowledge/

tools/

website/

maintenance/

admin/

developer/
```

每个：

```
npm run build
```

都能独立部署。

Portal：

负责

```
菜单

登录

权限

路由

主题
```

其它：

全部独立。

这样以后：

网站助手

甚至可以卖独立版本。

---

## 后端不要做成一个 Server

建议：

```
gateway
```

下面挂：

```
knowledge-service

tool-service

website-service

maintenance-service

user-service

auth-service

storage-service

workflow-service

agent-service
```

注意：

不是微服务。

而是：

Modular Monolith（模块化单体）。

---

### 第一阶段：模块化单体（✅ 当前 CoreKG 已实现）

> **当前状态**：CoreKG 已处于此阶段。`apps/corekg` 为聚合入口，通过 `Routers()` 挂载 kecore/kechat/account/keapi/keparser/kesearch；每个子应用有独立 `app.go` 暴露 `Routers()`/`Migrates()`/`RunJob()`，也可独立构建二进制（`make local APP=keapi`）。

一个程序：

```
corekg-server
```

里面：

```
internal/

knowledge/

tool/

website/

maintenance/

workflow/

user/

auth/

storage/
```

全部都是：

Module。

例如：

```
knowledge.Register()

tool.Register()

website.Register()
```

最终：

```
main.go

↓

注册Module
```

即可。

这样：

部署：

```
docker run
```

一个容器。

---

### 第二阶段

发现：

网站助手越来越大。

直接：

```
website/

↓

独立仓库

↓

website-service
```

其它：

不用改。

因为：

所有调用：

都是：

```
interface
```

不是：

```
直接调用数据库
```

---

## 每个模块都应该有自己的 API

例如：

Knowledge

```
POST /knowledge/create

GET /knowledge/query

DELETE /knowledge
```

Tools

```
POST /tools/pdf

POST /tools/ocr

POST /tools/extract
```

Website

```
POST /website/build

GET /website/status

POST /website/widget
```

Maintenance

```
POST /maintenance/build

POST /maintenance/query
```

它们互相不能访问数据库。

只能：

```
HTTP

或者

Go Interface

或者

Event
```

---

## Capability 一定不要依赖 UI

例如：

OCR

今天：

```
工具
```

调用：

```
OCR
```

明天：

```
网站助手
```

也要调用。

后天：

```
维修助手
```

也要调用。

Agent：

也要调用。

因此：

```
OCR

↓

独立 Capability
```

而不是：

```
Tools

↓

OCR
```

否则以后全是循环依赖。

---

## 整个后端建议按 Capability 划分

例如：

```
core/

auth

project

storage

user

org

permission

--------------------------------

knowledge/

parser/

ocr/

crawler/

embedding/

vector/

retrieval/

rerank/

workflow/

agent/

search/

--------------------------------

application/

website/

maintenance/

customer/

training/

```

注意：

Application

不能反向依赖。

例如：

```
Website

↓

Knowledge

↓

Parser

↓

OCR

↓

Storage
```

而不是：

```
Knowledge

↓

Website
```

---

## 数据库建议也拆层

很多产品：

一套数据库。

以后越来越乱。

建议：

```
Core

user

project

permission

organization

----------------

Knowledge

knowledge

chunk

embedding

datasource

crawler

----------------

Application

website

maintenance

customer

```

以后：

Website

独立。

数据库直接迁走。

Knowledge：

完全不用动。

### 当前数据库多库现状

当前 `pkgs/utils/dbutil` 已提供多个命名数据库连接，说明数据库层面已做了初步拆分：

| 连接名 | 对应目标层级 | 说明 |
|---|---|---|
| `Core()` | Core | 核心业务表（Forest、File、Project 等） |
| `Account()` | Core | 用户、组织、权限、API Key |
| `Chat()` | Knowledge | 对话、Session、Agent、消息 |
| `AIGC()` | Knowledge | AI 生成相关数据 |
| `LLM()` | Capability | LLM 模型配置、API Key 管理 |
| `Cook()` | Application | 业务定制数据 |
| `Cluster()` | Platform | 集群级配置、License |

> **演进方向**：当前多库拆分已为分层打下基础，后续需将 `Chat()`/`AIGC()` 中的共享模型与 kecore/kechat 解耦后，才能按 Knowledge/Application 层完全归位。

---

## 前后端之间建议采用 BFF（Backend For Frontend）

不要：

```
Vue

↓

Knowledge API
```

而是：

```
Vue

↓

Portal API

↓

Knowledge

↓

Tool

↓

Workflow
```

例如：

网站助手首页：

其实：

需要：

```
Knowledge

Tool

Crawler

Workflow

```

四个接口。

如果：

Vue：

自己调。

以后很难维护。

应该：

```
Website BFF

↓

聚合

↓

返回
```

这样：

以后：

APP

Web

Electron

都能复用。

> **落地前置条件**：BFF 层的引入与前端微前端 Portal 强绑定。建议当 Portal 门户搭建完成后同步引入 BFF，而非在后端独立先行建设。当前阶段前端仍为单体 Vue 项目，直接调用后端 API 即可。

---

## CoreKG Platform 完整模块规划

CoreKG Platform 的完整模块规划如下：

```
corekg-platform
│
├── Portal（门户）
│
├── IAM（认证权限）
│
├── Workspace（项目）
│
├── Knowledge Engine（知识引擎）
│
├── Tool Engine（工具引擎）
│
├── Workflow Engine（流程引擎）
│
├── Agent Engine（智能体引擎）
│
├── Search Engine（检索引擎）
│
├── Widget Engine（嵌入组件）
│
├── Web Fetch Engine（网页内容抓取）
│
├── Web Search Engine（外部搜索）
│
├── Application Center（应用中心）
│      ├── Website Assistant
│      ├── Maintenance Assistant
│      ├── Customer Assistant
│      ├── Training Assistant
│
└── Developer Center
       API
       SDK
       MCP
```

### 把"应用"和"能力"彻底分离

结合 Agent 平台的规划，最值得长期坚持的一条原则是：

> **任何业务应用都不能直接依赖另一个业务应用，只能依赖平台能力。**

例如：

```
Website Assistant
        │
        ├── Knowledge Engine
        ├── Crawl Engine
        ├── Search Engine
        ├── Widget Engine
        └── Agent Engine

Maintenance Assistant
        │
        ├── Knowledge Engine
        ├── OCR Engine
        ├── Search Engine
        ├── Workflow Engine
        └── Agent Engine
```

这样带来的收益非常大：

- **部署方式**：开发版可以把所有 Engine + Application 编译成一个二进制、一个 Docker 镜像；企业版则可以按 Engine 或 Application 独立部署。
- **前端组织**：每个 Application 都是一个独立前端，只消费平台 API，不直接访问其他应用的数据。
- **能力复用**：Crawler、OCR、Embedding、RAG、Widget 等能力可以被 Agent 平台、知识平台和未来的新应用共同使用，而不会产生重复实现。
- **产品演进**：以后新增"合同助手""巡检助手""招投标助手"，本质只是增加一个新的 Application，它们复用已有 Engine，而不是继续扩大知识库系统本身。

这会让你的产品从一开始就具备**一体化部署**和**模块化演进**两种能力，而不需要等系统复杂之后再进行痛苦的拆分。

---

## 当前 CoreKG 模块分层现状

基于对 `apps/` 下 18 个模块和 `pkgs/` 下 15 个共享包的完整分析，按 Platform / Capability / Application 三层梳理如下。

### Platform 层（平台基础服务）

所有应用都依赖、不依赖任何业务模块的基础设施：

| 模块 | 说明 |
|---|---|
| **account** | 用户认证、组织/部门/角色/权限/岗位、API Key、OAuth2 连接、企业微信集成、私有化部署模式管理（~153 条路由） |
| **admin** | 内部运营后台：员工/公司/用户管理、License 管理、系统配置、Prompt 版本管理、公告、付款记录、Dashboard |
| **corekg** | 聚合单体入口，挂载 kecore+kechat+account+keapi+keparser+kesearch 全部路由，附加 License 校验 |
| **keinit** | 初始化 CLI：建表、ES 索引、MinIO bucket、LLM 配置、数据迁移（非长驻服务） |

**pkgs 中属于 Platform 的：**

| 包 | 说明 |
|---|---|
| `global` | 全局常量、错误码、API 前缀、License RSA 公钥、Redis key 模式（几乎每个 app 都引用） |
| `types` | Bool/Secret/SafeID/Money/StringArray 等自定义类型，GORM+JSON 序列化 |
| `utils` | 加密、HTTP、文件、验证、DB 连接、通知、S3 等通用工具集 |
| `task` | Redis Stream 分布式任务队列（Push/Pop/CallBack/健康检查），keparser+keworker 的核心 |
| `jobs` | DB-backed 分布式 Cron 调度器 |
| `mutex` | Redis/DB 分布式锁 + 分布式资源池 |
| `apis` | WebSocket Server 框架、响应码常量、企微 Webhook Handler |
| `queue` | RabbitMQ RPC 队列客户端（预留） |

### Capability 层（可复用能力引擎）

被多个应用共同消费，自身不绑定具体业务场景：

| 模块/包 | 说明 | 消费者 |
|---|---|---|
| **kesearch** | ES 混合检索 + Rerank + Chunk CRUD + 全局搜索 | corekg, keapi, kechat |
| **keparser** | 文档解析任务编排（任务分发 + 回调 + 队列监控） | keworker, corekg |
| **keworker** | 7 种 Worker（PDF提取/分块/转PDF/AI摘要/视频抽帧/ES写入/文件拷贝），纯 CLI | keparser |
| **kellm** | LLM 代理网关库，OpenAI 兼容路由到多上游 | kechat |
| **einonodes** | Eino 图节点库：IntentRecognizer/DataLoader/Planner/Reporter/Branch/Executor | kechat |
| **nodes** | Eino ChatModel 工厂（BaseModel/GetBaseModelFromSetting） | kechat |
| **einotools** | ReAct/Summary Agent + 工具注册 + 代码沙箱 + Prompt 模板 + SSE Printer | kechat, kecore, keapi, kesearch |
| **einodocument** | Excel Eino Parser | kecore |
| **latex2omml** | LaTeX→OMML/MathML Python 微服务 | corekg/kecore |
| **webfetch** | 独立 URL 抓取服务（SSRF校验 + Chromium 渲染回退） | API 市场、内部 |
| **websearch** | 独立多源搜索服务（百度/Bing/Brave/DDG + 游标分页 + 去重） | API 市场、内部 |
| **kesale** | 支付/订单库（微信支付 + 分布式锁 + 定时对账） | corekg, kecore |
| `agentclient` | OpenAI Chat Completions HTTP Client（流式+非流式） | kechat, kecore, kesearch, keworker |
| `connectors` | OAuth2 外部平台集成（Google/Slack/Confluence）+ Token 管理 | account, kechat, kesearch |
| `plugins` | 外部数据库插件框架（MySQL 自省） | kechat, kecore |
| `wecoms` | 企微数据模型 + 组织架构同步 + 消息类型 | account, admin, corekg |
| `wx` | 微信网页 OAuth | account, admin |

### Application 层（业务应用）

面向具体业务场景，只依赖 Platform 和 Capability，不互相依赖：

| 模块 | 说明 | 依赖的 Capability |
|---|---|---|
| **kecore** | 知识库管理核心：Forest CRUD、文件上传预览解析、知识图谱、QA 对、AI 写作、项目管理、同义词/热词、数据库知识库、收藏、支付下单（~301 条路由） | kesearch, keparser, kesale, einotools, plugins |
| **kechat** | 对话/Agent 核心：流式问答、Session 管理、Agent CRUD+版本、模型管理、Coze 插件、文件 QA、图表画布、问题扩展、Workflow 测试、LLM 透传 | kesearch, kellm, einonodes, nodes, einotools, agentclient |
| **keapi** | 外部 API 网关：API-Key 鉴权、Forest/File/Chat/Search REST + MCP Server（21 Tools） | kesearch, kecore(models), kechat(models), einotools |
| **workflow** | 可视化工作流引擎（基于 Coze Studio 二开，Hertz 框架，独立 DDD 架构） | 自有基础设施 |

### 当前依赖方向

```
kecore ──→ kesearch, keparser, kesale, einotools, plugins
kechat ──→ kesearch, kellm, einonodes, nodes, einotools
keapi  ──→ kesearch, kecore(models), kechat(models), einotools
workflow ──→ (自有 DDD，几乎无跨 app 依赖)
     ↓
kesearch, keparser, kellm, einonodes, einotools ...
     ↓
global, types, utils, task, jobs, mutex, apis
```

基本符合 `Application → Capability → Platform` 的分层原则。

### 待解决的耦合问题

- **🔴 kecore ↔ kechat 双向模型引用（首要阻塞项）**：kecore 引用 `kechat/models`，kechat 引用 `kecore/models/fs`，两者无法独立拆分或独立部署。这是当前最大的架构债务，必须在任何模块拆分之前解决。方案：抽取共享模型到 `pkgs/models/shared`，或将共享实体定义为 interface
- **corekg 聚合初始化**：聚合单体同时初始化所有子系统（ES、NebulaGraph、MinIO、Connectors），拆分时需将初始化逻辑下沉到各 Capability
- **kesale 嵌入 kecore**：支付能力作为库直接 import 嵌入 kecore，无独立 HTTP API，无 interface 抽象。需升级为"带 interface 的独立 Capability"，否则其他 Application 需要支付时会产生直接依赖
- **Storage 未统一封装**：对象存储能力散落在各服务的 `utils/s3util` 中，尚未成为独立 Capability

---

## 与目标架构的映射

下表说明当前模块与 CoreKG Platform 目标架构的对应关系：

| CoreKG Platform 模块 | 当前对应 | 说明 |
|---|---|---|
| Portal | account / admin | 用户认证、后台管理已独立为 account 和 admin 应用 |
| IAM | account + middleware | 认证鉴权通过 account 应用及各应用 middleware 实现 |
| Workspace | kecore (Forest) | 知识库/项目空间以 Forest 为核心实体 |
| Knowledge Engine | kecore + keparser | 知识管理、文档解析、入库流程 |
| Search Engine | kesearch | 独立搜索服务，ES 混合检索 + Rerank 管线 |
| Agent Engine | kechat (ChatWrapper + Eino) | ReAct/RAG Agent、GraphSearch、流式对话 |
| Tool Engine | keapi (MCP Server) | 21 个 Tool 通过 MCP 协议暴露，HTTP API 并行提供 |
| Workflow Engine | workflow (apps/workflow) | 基于 Coze Studio 的可视化工作流编排，独立 Hertz 框架，**已具备独立部署条件**，可作为拆分试点 |
| Widget Engine | 暂无独立模块 | 嵌入组件能力目前内嵌在 keapi 中，未来可抽取 |
| Web Fetch Engine | webfetch | 独立部署的 URL 抓取服务，已有 API-Key 鉴权 |
| Web Search Engine | websearch | 独立部署的多源搜索服务，已有 API-Key 鉴权 |
| Application Center | keworker + 业务定制 | 当前应用逻辑分散在各服务中，未来按应用独立拆分 |
| Developer Center | keapi (Swagger + MCP) | API 文档通过 swag 生成，MCP Server 已就绪 |
| Storage | pkgs/utils + OSS 配置 | 对象存储能力散落在各服务，尚未统一封装为独立 Capability |
| OCR / Crawler | keparser 内部 | 目前作为 keparser 子能力，未来可独立为 Capability |

### 演进路线图

基于当前实际，建议按以下阶段推进：

```
Phase 0  解耦 kecore ↔ kechat 共享模型
         → 抽取 pkgs/models/shared 或定义 interface
         → 消除双向 import，为后续拆分解除阻塞
         ← 首要任务，所有后续阶段依赖此步

Phase 1  workflow 独立部署试点
         → workflow 已具备独立部署条件（Hertz 框架、独立 DDD、零跨 app 依赖）
         → 验证拆分流程、部署管线、接口契约
         ← 风险最低的拆分试点

Phase 2  kesale 升级为独立 Capability
         → 从嵌入式库升级为带 interface + HTTP API 的独立服务
         → 解除 kecore 对支付的直接依赖

Phase 3  OCR / Crawler 从 keparser 抽取
         → 独立为 Capability，供 keworker / kecore / 未来 Application 共同消费
         → keparser 仅保留任务编排职责

Phase 4  Storage 统一封装
         → 将散落在 utils/s3util 中的对象存储能力收敛为独立 Capability
         → 统一上传/下载/签名 URL 接口

Phase 5  Portal 微前端 + BFF 层落地
         → Portal 搭建后同步引入 BFF
         → 前端按 Application 拆分为独立 Vue 项目

Phase 6  Application 层按业务场景独立拆分
         → 网站助手、维修助手、客服助手等从 kecore/kechat 中剥离
         → 每个 Application 只依赖 Platform + Capability
```

> **总结**：当前架构已具备模块化单体的雏形（apps/ 目录隔离、interface 调用、独立二进制构建、多库连接）。演进的核心原则是：**先解耦、后拆分**。Phase 0 解决 kecore↔kechat 双向依赖是所有后续工作的前提；Phase 1 以 workflow 为低风险试点验证拆分流程；后续逐步将嵌入库升级为独立 Capability，最终实现 Application 层的完全解耦。
