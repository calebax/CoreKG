# 应用（Application）模块设计方案

## 1. 概述

在一级菜单"问AI"下方新增"应用"菜单项，作为与"知识库"平级的导航入口。应用是整个平台的核心聚合对象（Aggregate Root），所有能力（知识、Agent、RAG、Workflow 等）最终服务于应用。

本方案以 **网站助手（Website Assistant）** 作为第一个模板类型进行深度设计，完整实现从网站爬取、知识入库、对话测试到 Widget 发布的全链路，同时作为 AIOS Knowledge Engine 新架构的试点落地。

### 1.1 Phase 1 范围

**前端**：
- 侧边栏新增"应用"菜单项（i18n 全覆盖）
- 应用列表页（卡片展示，对接后端 API）
- 应用创建向导（3 步 Builder Wizard，对接后端 API）
- 应用详情页（6 Tab 布局，URL query 参数同步 Tab 状态，对接后端 API）
- 网站助手专属功能：页面管理、爬取规则配置、同步触发、FAQ 管理、对话测试、Widget 发布、Prompt 配置、运营数据面板

**后端**（独立 sub-app `keapp`）：
- Application CRUD API（创建、查询、列表、更新、删除）
- 完整权限体系（基于 `KeResourceScope`，user/department/company/public 四级 scope）
- 模板差异化数据结构（5 种模板类型各自独立 config schema）
- 复用 `dbutil.Knownow()` 数据库连接

**独立 Crawler 服务**（Node.js + Puppeteer）：
- 两层 Parser 架构（Container Parser + Content Parser）
- 通过现有 Redis 任务队列与 kecore/keapp 通信
- 输出统一 Resource Manifest
- 增量同步 + 版本控制（ETag/Content-Hash）
- 独立部署（Docker 容器）

### 1.2 架构定位：AIOS Knowledge Engine 试点

> **设计目标不是"如何做好 RAG"，而是"如何让任何知识资源，都能以最优方式被 Agent 检索和调用"。**

网站助手作为 AIOS Knowledge Engine 统一资源模型的首个落地场景，遵循以下核心原则：

- **统一资源模型**：网站页面视为 `Resource{ID, Type, Source, Metadata, Content}`，而非传统 Forest 扩展
- **两层 Parser**：Container Parser（Puppeteer 拆 DOM）+ Content Parser（理解文本/表格/图片）
- **多索引体系**：Vector Index + Keyword Index，预留 SQL/Vision/Graph Index 插件接口
- **Evidence 而非 Chunk**：检索结果封装为 `Evidence{resource_id, content_type, score, locator, snippet, payload}`
- **Index Builder 插件化**：`IndexBuilder.Build(manifest) → []IndexArtifact` 接口抽象

## 2. 长期愿景（Future Vision）

> **应用 = 产品**，其它（知识、Agent、RAG、OCR、Workflow）全部都是应用的基础设施。

### 2.1 Application 生命周期

```
创建 → 导入数据 → AI构建 → 测试 → 发布 → 运营 → 持续同步
```

### 2.2 Application 作为聚合根

```
Application
├── Resources（网站、PDF、FAQ、API、数据库）
├── Knowledge（解析后的多索引知识）
├── AI（Prompt、Agent、Workflow、Query Router）
├── Channels（Widget、API、微信）
├── Operations（日志、分析、健康度）
├── Members（负责人、协作者）
└── Deployments（发布实例）
```

### 2.3 应用模板类型

- **网站助手**（website）：网站URL、爬取规则、页面地图、FAQ、同步配置、Widget 发布、Prompt 配置、对话测试、运营数据
- **产品助手**（product）：产品名称、版本信息、PDF、图片、视频
- **售后助手**（aftersales）：服务范围、维修记录、故障案例、视频、零件库
- **培训助手**（training）：培训部门、课程、资料、考试题库
- **制度助手**（policy）：组织名称、制度文件、流程图、审批记录

每种模板类型在后端有独立的 config struct，通过 `AppConfig.Config` JSON 字段存储，运行时通过 `AsWebsite()` / `AsProduct()` 等方法进行类型断言。

### 2.4 AI 状态面板

```
知识: 542 | Prompt: OK | Workflow: OK | 模型: DeepSeek
Embedding: BGE | Rerank: 开启 | Query Router: 规则+LLM | Evidence: 启用
```

## 3. 技术选型与架构决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| 后端归属 | 独立 `keapp` sub-app | 未来可独立部署/扩展 |
| DB 连接 | 复用 `dbutil.Knownow()` | 与 forest 同库 |
| 权限模型 | 完整四级 scope | 复用 `KeResourceScope` + `perm` 包 |
| Tab 状态 | URL query `?tab=xxx` | 支持刷新/分享/书签 |
| Crawler | Node.js + Puppeteer | JS 渲染支持，独立部署 |
| 爬取调度 | 复用 Redis 任务队列 | 与 pkgs/queue 一致 |
| 数据模型 | Resource 模型试点 | 不扩展 ForestType 枚举 |
| 检索结果 | Evidence 格式 | 统一 locator + payload |
| Index Builder | 插件接口 | Phase 1 实现 Vector + Keyword |

## 4. 后端设计

### 4.1 数据库表

#### ke_application（应用表）

```sql
CREATE TABLE IF NOT EXISTS `ke_application` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `uin` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `company_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `name` VARCHAR(255) NOT NULL DEFAULT '',
    `type` VARCHAR(64) NOT NULL DEFAULT '',
    `status` VARCHAR(32) NOT NULL DEFAULT 'draft',
    `description` VARCHAR(1024) NOT NULL DEFAULT '',
    `icon` VARCHAR(512) NOT NULL DEFAULT '',
    `color` VARCHAR(16) NOT NULL DEFAULT '#0C99FF',
    `knowledge_count` INT NOT NULL DEFAULT 0,
    `faq_count` INT NOT NULL DEFAULT 0,
    `sync_status` VARCHAR(32) NOT NULL DEFAULT 'success',
    `last_sync_at` DATETIME(3) DEFAULT NULL,
    `last_publish_at` DATETIME(3) DEFAULT NULL,
    `config` JSON DEFAULT NULL,
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at` DATETIME(3) DEFAULT NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_app_uin` (`uin`),
    INDEX `idx_app_company` (`company_id`),
    INDEX `idx_app_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='应用表';
```

#### ke_web_resource（网站资源表 — AIOS Resource 模型试点）

```sql
CREATE TABLE IF NOT EXISTS `ke_web_resource` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `app_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `company_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `source_url` VARCHAR(2048) NOT NULL DEFAULT '',
    `title` VARCHAR(512) NOT NULL DEFAULT '',
    `resource_type` VARCHAR(32) NOT NULL DEFAULT 'web',
    `content_hash` VARCHAR(64) NOT NULL DEFAULT '',
    `etag` VARCHAR(256) NOT NULL DEFAULT '',
    `last_modified` VARCHAR(64) NOT NULL DEFAULT '',
    `manifest` JSON DEFAULT NULL,
    `index_status` VARCHAR(32) NOT NULL DEFAULT 'pending',
    `indexed_at` DATETIME(3) DEFAULT NULL,
    `crawl_count` INT NOT NULL DEFAULT 0,
    `last_crawl_at` DATETIME(3) DEFAULT NULL,
    `crawl_error` TEXT,
    `metadata` JSON DEFAULT NULL,
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at` DATETIME(3) DEFAULT NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_web_res_app` (`app_id`),
    INDEX `idx_web_res_company` (`company_id`),
    INDEX `idx_web_res_hash` (`content_hash`),
    INDEX `idx_web_res_status` (`index_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网站资源表';
```

#### ke_web_crawl_rule（爬取规则表）

```sql
CREATE TABLE IF NOT EXISTS `ke_web_crawl_rule` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `app_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `rule_type` VARCHAR(32) NOT NULL DEFAULT 'include',
    `pattern` VARCHAR(512) NOT NULL DEFAULT '',
    `priority` INT NOT NULL DEFAULT 0,
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    INDEX `idx_crawl_rule_app` (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网站爬取规则表';
```

#### ke_crawl_task（爬取任务表）

```sql
CREATE TABLE IF NOT EXISTS `ke_crawl_task` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `app_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `resource_id` BIGINT UNSIGNED DEFAULT NULL,
    `source_url` VARCHAR(2048) NOT NULL DEFAULT '',
    `task_type` VARCHAR(32) NOT NULL DEFAULT 'full',
    `status` VARCHAR(32) NOT NULL DEFAULT 'pending',
    `redis_task_id` VARCHAR(128) NOT NULL DEFAULT '',
    `error_message` TEXT,
    `pages_crawled` INT NOT NULL DEFAULT 0,
    `pages_total` INT NOT NULL DEFAULT 0,
    `started_at` DATETIME(3) DEFAULT NULL,
    `finished_at` DATETIME(3) DEFAULT NULL,
    `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    INDEX `idx_crawl_task_app` (`app_id`),
    INDEX `idx_crawl_task_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='爬取任务表';
```

### 4.2 GORM Model 与类型定义

**包路径**：`apps/keapp/models/apptype/`

**枚举**（`enums.go`）：`AppStatus`（draft/online/paused）、`SyncStatus`、`AppTemplateType`（5种）、`CrawlRuleType`（include/exclude）、`CrawlTaskType`（full/incremental/deleted）、`CrawlTaskStatus`、`IndexStatus`（pending/indexed/failed）

**模板 Config**（`template.go`）：`WebsiteConfig`（url/sync_schedule/max_depth/max_pages/respect_robots/capabilities）、`ProductConfig`、`AftersalesConfig`、`TrainingConfig`、`PolicyConfig`、`AppCapabilities`、`AppConfig`（Value/Scan + AsWebsite/AsProduct 等类型断言方法）

**AIOS 类型**（`resource_manifest.go`）：
- `ResourceManifest{ResourceID, ResourceType, SourceURL, ContentUnits[], Metadata}`
- `ContentUnit{Type, Location, Content, Metadata}`
- `ContentLocator{Selector, Index, Title}`
- `Evidence{ResourceID, ContentType, Score, Locator{URL, Title, Selector}, Snippet, Payload}`
- `IndexArtifact{IndexType, DocumentID, Payload}`
- `IndexBuilder` 接口：`Build(ResourceManifest) → []IndexArtifact`

**GORM Models**：`KeApplication`、`KeWebResource`、`KeWebCrawlRule`、`KeCrawlTask`，均嵌入 `gorm.Model`，表名常量定义在 `db.go`。

### 4.3 DAO 层

遵循 `BaseModel` + `BaseCond` 模式，按模块拆分：

**公共层**（`models/app/`）：

| DAO | 核心方法 |
|-----|---------|
| ApplicationDao | Insert / GetByID / UpdateMap / SoftDelete / CheckNameExists / GetPageListByCond |

**网站助手**（`website/models/`）：

| DAO | 核心方法 |
|-----|---------|
| WebResourceDao | Insert / UpdateManifest / UpdateIndexStatus / ListByAppID / GetByURL |
| WebCrawlRuleDao | ListByAppID / Insert / DeleteRule |
| CrawlTaskDao | Insert / UpdateStatus / GetLatestByAppID |

### 4.4 Service 层

包级函数 + sentinel errors，按模块拆分：

**公共层**（`services/svcapp/`）：
- **app_api.go**：CRUD，创建时事务内同时插入 ke_resource_scope
- **permission_api.go**：复用 kecore perm 包

**网站助手**（`website/services/`）：
- **web_resource_api.go**：ListWebResources / AddCrawlRule / ListCrawlRules / DeleteCrawlRule
- **crawl_api.go**：TriggerFullCrawl → Redis 任务 → Crawler 消费 → HTTP 回调 HandleCrawlCallback → IndexBuilder → ES；TriggerIncrementalSync 对比 ETag/Hash
- **evidence_api.go**：SearchEvidence = FAQ 精确匹配 → ES Vector+BM25 并行 → Evidence 封装 → Rerank

未来新增具体应用（售后助手等）只需添加对应的 `services/` 目录，不改动公共层或其他应用代码。

### 4.5 DTO + Handler + Router

API action 保持 `keapp.*` 前缀，按模块拆分 handler：

**公共层**（`apis/appctl/`）：
```
CreateApplication / ListApplications / GetApplication / UpdateApplication / DeleteApplication / UpdateAppPermission
```

**网站助手**（`website/apis/`）：
```
ListWebResources / AddCrawlRule / ListCrawlRules / DeleteCrawlRule
TriggerCrawl / GetCrawlTaskStatus / CrawlCallback(internal) / SearchEvidence
```

Middleware：RequireAppCreatePerm / RequireAppViewPerm / RequireAppManagePerm

**路由聚合**：`apis/apis.go` 统一调用公共层和各具体应用的 `RegistryRouter`，未来新增应用只需在此处加一行注册。

### 4.6 后端文件结构

```
apps/keapp/
├── app.go                              # 聚合入口：Routers/Migrates/RunJob
├── cmd/{main.go, init.go}              # 精简版独立 bootstrap
├── conf/test/config.yaml
│
├── models/                             # 公共层 model
│   ├── apptype/                        # 公共类型（枚举、AppConfig、ResourceManifest、Evidence、KeApplication）
│   │   ├── enums.go
│   │   ├── template.go
│   │   ├── resource_manifest.go
│   │   ├── application.go              # KeApplication GORM model
│   │   └── db.go                       # 公共表 InitDB
│   └── app/                            # 公共 DAO（BaseModel + ApplicationDao）
│       ├── base.go
│       └── ke_application.go
│
├── services/                           # 公共层 service
│   └── svcapp/                         # Application CRUD + 权限
│       ├── app_api.go
│       └── permission_api.go
│
├── apis/                               # 公共层 handler + 路由聚合
│   ├── apis.go                         # 聚合 RegistryRouter（调用 appctl + websitectl + ...）
│   └── appctl/                         # Application CRUD handlers + DTOs
│       ├── app.go
│       └── dto.go
│
├── mds/                                # 公共权限中间件
│   └── appmds.go
│
├── website/                            # 网站助手（自包含 MVC）
│   ├── models/                         # KeWebResource + CrawlRule + CrawlTask + DAOs + InitDB
│   │   ├── types.go                    # 网站助手专属 GORM models
│   │   ├── dao.go                      # WebResourceDao + WebCrawlRuleDao + CrawlTaskDao
│   │   └── db.go                       # 网站助手表 InitDB
│   ├── services/                       # 爬取、Evidence、页面管理
│   │   ├── web_resource_api.go
│   │   ├── crawl_api.go
│   │   └── evidence_api.go
│   └── apis/                           # handlers + DTOs + RegistryRouter
│       ├── registry.go                 # 网站助手路由注册
│       ├── web_resource.go             # handlers
│       ├── crawl.go
│       ├── evidence.go
│       └── dto.go                      # 网站助手专属 DTOs
│
├── internal/docs/                      # swagger 自动生成
│
└── (未来: aftersales/, product/, training/, policy/ ...)
```

**设计原则**：
1. 根目录的 `models/` `services/` `apis/` `mds/` = 框架公共层，所有具体应用共享
2. 每个具体应用一个目录，内含完整 `models/` `services/` `apis/`，自包含
3. 新增具体应用 = 新增一个目录 + 在 `apis/apis.go` 加一行注册，零改动现有代码
4. 各模块有独立 `InitDB()`，corekg 逐个调用
5. 保留精简版 `cmd/` 用于独立构建/调试

## 5. Crawler 服务设计（Node.js + Puppeteer）

### 5.1 两层 Parser 架构

```
Redis Task Queue → Crawler Service
  ├── Container Parser (Puppeteer)
  │   DOM → Content Blocks (text/table/image/link)
  │   输出: Resource Manifest (content_units[])
  ├── Content Parser
  │   Text Normalize (HTML → Markdown, 去噪)
  │   Table Extract → Structured Table JSON
  │   (Image OCR/Caption: Phase 2)
  ├── Version Tracker
  │   ETag / Content-Hash / Last-Modified
  │   增量判断：hash 不变 → 跳过
  └── Callback Reporter
      POST keapp.CrawlCallback with ResourceManifest
```

### 5.2 通信协议

**任务消息格式**（Redis）：
```json
{
  "task_id": "ke_crawl_task_123",
  "app_id": 1,
  "source_url": "https://example.com",
  "task_type": "full",
  "max_depth": 3,
  "max_pages": 100,
  "respect_robots": true,
  "include_patterns": [".*"],
  "exclude_patterns": [".*\\.(pdf|zip)$"],
  "callback_url": "https://api.internal/v3/keapp.CrawlCallback"
}
```

**回调格式**（HTTP POST）：
```json
{
  "request": {
    "task_id": "ke_crawl_task_123",
    "status": "success",
    "manifest": {
      "resource_id": "web-page-001",
      "resource_type": "web",
      "source_url": "https://example.com/about",
      "content_units": [
        {"type": "text", "location": {"selector": "main", "title": "About Us"}, "content": "..."},
        {"type": "table", "location": {"selector": "table:nth(1)"}, "content": "[[...]]", "metadata": {"rows": 5, "cols": 3}}
      ],
      "metadata": {"title": "About - Example", "language": "zh-CN", "description": "..."}
    }
  }
}
```

### 5.3 增量同步策略

1. 请求 sitemap + 已知 URL 列表
2. 对每个 URL 发 HEAD 请求获取 ETag/Last-Modified
3. 与 `ke_web_resource.content_hash` / `etag` / `last_modified` 对比
4. 仅对变更 URL 发送爬取任务
5. 已删除页面标记为 deleted，保留历史版本

### 5.4 部署与配置

- 独立 Docker 镜像（`example.com/corekg/crawler:<tag>`）
- 配置文件：Redis 连接、回调 URL、Puppeteer 并发数、超时、User-Agent
- CI/CD：独立流水线，与 keapp/kecore 解耦

## 6. 前端设计

### 6.1 路由与 Tab 同步

已实现 3 条路由：`/apps`、`/apps/create`、`/apps/:id`

Tab 通过 URL query `?tab=overview` 同步，受控模式 `activeKey={activeTab}`，`setSearchParams({tab}, {replace: true})`。

### 6.2 网站助手专属 Tab 设计

| Tab | 内容 | Phase 1 |
|-----|------|---------|
| overview | 仪表盘：爬取覆盖率、索引健康度、内容新鲜度、知识/FAQ 统计、最近活动 | 完整 |
| data | **页面管理**：已爬取页面列表（URL/标题/状态/最近爬取时间）+ **爬取规则**：include/exclude 规则 CRUD + **同步控制**：手动触发全量/增量同步 + 任务进度 | 完整 |
| ai | Prompt 配置（系统 prompt/欢迎语/建议问题）+ 对话测试面板 + AI 状态 | Prompt 配置 + 对话测试 |
| publish | Widget 嵌入代码生成 + Iframe + API + 二维码 + 样式配置 | Widget + Iframe |
| analytics | 访问量/对话量/满意度/检索命中率（按 content_type 分维度） | 基础统计 |
| settings | 应用信息编辑 + 删除 + 权限管理 | 完整 |

### 6.3 API + Store

`src/api/application.ts`：applicationApi.list/get/create/update/delete + webResourceApi.list/addCrawlRule/listCrawlRules/deleteCrawlRule + crawlApi.trigger/getStatus + evidenceApi.search

`src/stores/application.ts`：Zustand store（loadList/loadDetail/createApp/updateApp/deleteApp + loadWebResources/addCrawlRule/triggerCrawl/searchEvidence）

### 6.4 i18n

4 语言全覆盖，`app.application.*` section 覆盖所有用户可见字符串。

## 7. 边界状态处理

| 场景 | 处理方式 |
|------|---------|
| 应用 ID 无效 | 404 + 返回列表 |
| 名称重复 | 后端 keapp_name_exists → message.error |
| 删除应用 | Popconfirm 二次确认 |
| Tab 参数非法 | 回退 overview |
| 爬取失败 | 任务列表显示错误信息，可重试 |
| 爬取中页面变更 | ETag/Hash 对比跳过未变更 |
| 回调超时 | 任务标记 failed，支持手动重试 |

## 8. 验证

### 后端
```bash
make local APP=keapp
make generate-docs APP=keapp
```

### 前端
```bash
cd frontend/corekg && pnpm lint && pnpm build-ts
```

### 端到端
- [ ] 创建网站助手 → 配置 URL → 触发全量爬取 → 页面列表显示爬取结果
- [ ] 修改爬取规则 → 重新同步 → 验证 include/exclude 生效
- [ ] 增量同步 → 仅重新爬取变更页面
- [ ] 对话测试 → Evidence 格式返回（含 url/title/selector locator）
- [ ] Widget 发布 → 嵌入代码可复制到外部网站
- [ ] Prompt 配置 → 保存后对话测试使用新 prompt
- [ ] 运营面板 → 显示爬取覆盖率/索引健康度/检索命中率
- [ ] 4 语言切换 → 所有文案正确
- [ ] 权限控制 → 非管理用户无法触发同步/修改规则
