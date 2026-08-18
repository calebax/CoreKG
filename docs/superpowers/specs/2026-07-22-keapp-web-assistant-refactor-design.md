# keapp 网页助手重构设计

> 日期: 2026-07-22
> 范围: keapp 完整产品重构（本次仅 keapp 侧，kecore website forest type 后续对接）

## 1. 整体架构与数据流

### 核心定位

keapp 是一个**网页数据采集+管理应用**，采集结果后续通过 kecore 的 website forest type 注入知识库索引管线。用户通过 keapp 管理爬取目标、规则、任务；通过 kecore 的知识库进行检索和对话。

### 数据流

```
用户配置 App（URL、规则）
        ↓
TriggerCrawl → 创建 KeCrawlTask (pending)
        ↓
Publish NATS message → "keapp.crawl.trigger"
        ↓
keapp 内部 NATS subscriber 消费任务
        ↓
爬取网页 → 提取内容 → 生成 markdown
        ↓
去重检查 (content_hash)
        ↓
写入/更新 ke_web_resource (元数据 + raw_content + manifest)
        ↓
更新 KeCrawlTask 统计 (pages_new/updated/skipped)
        ↓
[后续] 对接 kecore website forest type → parse → chunk → embed → ES
```

### 应用生命周期

- 创建 App 时初始化 `crawl_config`，`forest_id` 预留为 NULL
- 等 kecore 支持 website forest type 后，创建 App 时自动创建关联 Forest
- 删除 App 时软删除，关联资源保留

### 与 kecore 的关系

keapp 保持独立应用身份。kecore 后续新增 `forest_type=website`，keapp 爬取结果直接写入该类型 Forest 下的文件。本次重构不对接 kecore，仅预留 `forest_id` / `forest_file_id` 字段。

## 2. 数据库表结构

### ke_application（调整）

| 变更 | 类型 | 说明 |
|---|---|---|
| +`forest_id` | BIGINT UNSIGNED NULL | 预留，后续对接 kecore website forest 时回填 |
| +`crawl_config` | JSON | 爬取专用配置，从通用 config 拆出 |

`crawl_config` 结构（原 WebsiteConfig）:

```json
{
  "url": "https://example.com",
  "max_depth": 3,
  "max_pages": 500,
  "respect_robots": true,
  "sync_schedule": "0 2 * * *",
  "capabilities": {
    "ai_assistant": true,
    "search": true,
    "faq": false,
    "widget": false
  }
}
```

现有字段（name, type, status, description, icon, color, sync_status, last_sync_at 等）保留不变。

### ke_web_resource（调整）

| 变更 | 类型 | 说明 |
|---|---|---|
| +`forest_file_id` | BIGINT UNSIGNED NULL | 预留，后续对接 kecore 时回填 |
| +`raw_content` | LONGTEXT | 爬取后的 markdown 内容 |

保留 `manifest` JSON、`source_url`、`title`、`content_hash`、`etag`、`last_modified`、`index_status`、`crawl_count`、`last_crawl_at`、`crawl_error`、`metadata` 等现有字段。

### ke_web_crawl_rule（不变）

字段：id, app_id, rule_type, pattern, priority, created_at, updated_at。结构已合理。

### ke_crawl_task（调整）

| 变更 | 类型 | 说明 |
|---|---|---|
| +`pages_new` | INT | 本次新发现的页面数 |
| +`pages_updated` | INT | 内容变更的页面数 |
| +`pages_skipped` | INT | 跳过（规则排除或内容未变）的页面数 |

现有字段（status, task_type, source_url, redis_task_id→废弃, pages_crawled, pages_total, error_message, started_at, finished_at, created_by 等）保留。`redis_task_id` 字段废弃（不再使用 Redis）。

### 迁移脚本

新增 `scripts/mysql/v2.18_0__alter_keapp_tables.sql`，包含：
- ALTER ke_application ADD forest_id, crawl_config
- ALTER ke_web_resource ADD forest_file_id, raw_content
- ALTER ke_crawl_task ADD pages_new, pages_updated, pages_skipped
- UPDATE ke_crawl_task SET redis_task_id = NULL（清理废弃字段数据）

## 3. 接口设计

### 命名规范

- 通用 App 管理：`keapp.*`
- 网页助手子域：`keapp.web.*`
- 后续可扩展：`keapp.product.*`、`keapp.training.*` 等

### 通用 App 管理接口

| Action | 说明 | 权限 |
|---|---|---|
| `keapp.CreateApplication` | 创建应用 + 初始化 crawl_config | login |
| `keapp.ListApplications` | 分页列表，支持 name/type/status 过滤 | login |
| `keapp.GetApplication` | 详情，含 crawl_config 和 forest_id | login + view |
| `keapp.UpdateApplication` | 部分更新，含 crawl_config | login + manage |
| `keapp.DeleteApplication` | 软删除 | login + manage |

### 网页助手接口（keapp.web.*）

#### 爬取规则

| Action | 说明 | 权限 |
|---|---|---|
| `keapp.web.AddCrawlRule` | 新增规则 | login + manage |
| `keapp.web.ListCrawlRules` | 列出某 app 下所有规则 | login + view |
| `keapp.web.UpdateCrawlRule` | 修改规则 pattern/priority/type | login + manage |
| `keapp.web.DeleteCrawlRule` | 删除规则 | login + manage |

#### 爬取任务

| Action | 说明 | 权限 |
|---|---|---|
| `keapp.web.TriggerCrawl` | 触发全量/增量/单页爬取 | login + manage |
| `keapp.web.GetCrawlTask` | 查询单个任务状态 | login + view |
| `keapp.web.ListCrawlTasks` | 分页列出某 app 的爬取任务历史 | login + view |
| `keapp.web.CancelCrawlTask` | 取消 pending/running 任务（DB 状态更新） | login + manage |

#### Web Resource

| Action | 说明 | 权限 |
|---|---|---|
| `keapp.web.ListResources` | 分页列表，支持 status/url 过滤 | login + view |
| `keapp.web.GetResource` | 单个资源详情（含 raw_content） | login + view |
| `keapp.web.DeleteResource` | 删除资源 | login + manage |
| `keapp.web.RecrawlResource` | 对单个 URL 重新爬取 | login + manage |

### 权限中间件优化

统一 `AppContextMiddleware`：
- 解析一次请求 body，提取 `app_id` 注入 gin.Context
- `RequireAppViewPerm` / `RequireAppManagePerm` 从 context 读取，不再重复解析 body
- `app_id == 0` 时跳过检查（pass-through）

## 4. NATS JetStream Worker

### 架构

```
TriggerCrawl API
    ↓
写入 ke_crawl_task (status=pending)
    ↓
Publish NATS message → subject "keapp.crawl.trigger" {task_id}
    ↓
keapp NATS subscriber (启动时创建 durable consumer)
    ↓
查询 task，校验 status（跳过 cancelled）
    ↓
更新 task status=running
    ↓
执行爬取逻辑
    ↓
写入/更新 ke_web_resource
    ↓
更新 task status=success/failed + 统计
```

### NATS 配置

| 参数 | 值 | 说明 |
|---|---|---|
| Subject | `keapp.crawl.trigger` | 爬取触发消息 |
| Stream | `KEAPP_CRAWL` | JetStream stream, retention=limits |
| Consumer | `keapp-crawl-worker` | Durable push consumer |
| MaxDeliver | 3 | 失败重试次数 |
| AckWait | 5min | 单任务最大处理耗时 |

### Worker 配置（config.yaml）

```yaml
worker:
  concurrency: 3
  poll_interval: 5s
  max_retries: 3
  cancel_check_interval: 5  # 每处理 N 个页面检查一次 DB cancel 状态
```

### 取消机制

纯 DB 状态驱动，无 NATS cancel 消息：

1. `CancelCrawlTask` API 更新 `ke_crawl_task.status = 'cancelled'`（WHERE status IN ('pending','running')）
2. Worker 每处理 N 个页面查询一次 DB 中该 task 的 status
3. 如果 status=cancelled，立即停止并返回
4. Worker 恢复 pending/running 任务时，从 NATS 拿到消息后先查 DB status，cancelled 则 ack 跳过

### 爬取流程

1. 读取 app 的 `crawl_config`（URL、max_depth、max_pages、rules）
2. BFS 遍历页面，对每个 URL：
   - 检查 cancel 状态（每 N 个页面）
   - 匹配 crawl rules（include/exclude pattern）
   - HTTP GET，记录 etag/last-modified
   - 计算 content_hash，与已有 resource 比对
   - 新内容：HTML → markdown → 写入 ke_web_resource
   - 内容未变：跳过，更新 crawl_count
   - 内容变更：更新 raw_content + content_hash
3. 更新 task 统计（pages_new, pages_updated, pages_skipped, pages_crawled）
4. 标记 task 完成

### 启动恢复

Worker 启动时查询 DB 中 `status IN ('pending', 'running')` 的 task，重新 publish 到 NATS subject，由 consumer 重新消费。

## 5. 代码结构

```
apps/keapp/
├── app.go
├── cmd/
│   ├── main.go
│   └── init.go              # MySQL + NATS 初始化（移除 Redis）
├── conf/test/config.yaml    # +NATS 配置，移除 redis
├── internal/
│   ├── apis/
│   │   ├── apis.go          # 注册所有子域路由
│   │   ├── appctl/          # 通用 App CRUD
│   │   │   ├── app.go
│   │   │   ├── dto.go
│   │   │   └── registry.go
│   │   └── webctl/          # 网页助手子域（原 website/apis/）
│   │       ├── crawl.go     # 爬取任务 handlers
│   │       ├── resource.go  # Web Resource handlers
│   │       ├── rule.go      # Crawl Rule handlers
│   │       ├── dto.go       # Web 子域 DTO
│   │       └── registry.go
│   ├── middleware/
│   │   └── appmds.go        # 统一 AppContextMiddleware（原 mds/appmds.go）
│   └── docs/
│       └── keapp_docs.go
├── models/
│   ├── app/                 # DAO（不变）
│   ├── apptype/
│   │   ├── application.go   # KeApplication struct
│   │   ├── db.go            # InitDB
│   │   └── enums.go         # 通用枚举
│   └── web/                 # 网页助手模型（原 website/models/）
│       ├── resource.go      # KeWebResource struct + DAO
│       ├── crawl_rule.go    # KeWebCrawlRule struct + DAO
│       ├── crawl_task.go    # KeCrawlTask struct + DAO
│       ├── enums.go         # Web 专属枚举
│       └── db.go            # Web 表 InitDB
├── services/
│   ├── svcapp/              # 通用 App 服务（不变）
│   │   ├── app_api.go
│   │   └── permission_api.go
│   └── svcweb/              # 网页助手服务（原 website/services/）
│       ├── resource_api.go
│       ├── crawl_api.go     # 实现 GetCrawlTask，TriggerCrawl 发 NATS
│       └── rule_api.go      # 补齐 UpdateCrawlRule
└── worker/                  # 新增：爬取 Worker
    ├── crawler.go           # 核心爬取逻辑（BFS、HTML→MD、去重）
    ├── subscriber.go        # NATS subscriber + cancel 检查
    └── config.go            # Worker 配置
```

### 关键变更

1. `website/` 目录消除，按职责拆分到 `internal/apis/webctl/`、`models/web/`、`services/svcweb/`
2. 新增 `worker/` 目录，爬取核心逻辑与 API 层分离
3. `mds/` 移入 `internal/middleware/`
4. `cmd/init.go` 移除 Redis 初始化，新增 NATS 连接
5. `conf/test/config.yaml` 移除 redis 配置，新增 nats 配置

## 6. 枚举与状态机

### 通用枚举（models/apptype/enums.go）

| 枚举 | 值 |
|---|---|
| `AppStatus` | draft, online, paused |
| `SyncStatus` | success, failed, syncing |
| `AppTemplateType` | website, product, aftersales, training, policy |

### Web 专属枚举（models/web/enums.go）

| 枚举 | 值 |
|---|---|
| `CrawlRuleType` | include, exclude |
| `CrawlTaskType` | full, incremental, single |
| `CrawlTaskStatus` | pending, running, success, failed, cancelled |
| `IndexStatus` | pending, indexed, failed |

### 爬取任务状态机

```
pending ──→ running ──→ success
    │           │
    │           └─────→ failed (可重试)
    │
    └─────────────────→ cancelled

running ──────────────→ cancelled
```

- `pending → cancelled`: Worker 尚未开始，API 直接更新 DB
- `running → cancelled`: Worker 定期轮询 DB 发现 cancelled，停止处理
- `failed` 的任务可通过 TriggerCrawl 重新触发（创建新 task）

## 7. 后续待办（本次不做）

- [ ] kecore 新增 `forest_type=website` 支持
- [ ] keapp 创建 App 时自动创建关联 Forest
- [ ] 爬取完成后将 raw_content 写入 kecore Forest File
- [ ] kecore task chain 支持 website 类型内容的 parse/chunk/embed
- [ ] Evidence 搜索功能（基于 kesearch 对 website forest 的检索）
