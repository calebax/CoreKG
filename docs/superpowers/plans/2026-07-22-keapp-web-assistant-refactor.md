# keapp Web Assistant 重构实施计划

> **智能体工作者须知：** 必须使用子技能：推荐使用 superpowers:subagent-driven-development 或 superpowers:executing-plans 逐任务执行本计划。步骤使用复选框（`- [ ]`）语法进行跟踪。

**目标：** 将 keapp 重构为完整的 Web Assistant 产品，包含基于 NATS JetStream 的爬虫 Worker、重新组织的代码结构以及完整的 CRUD API。

**架构：** keapp 是一个独立的网页爬取管理应用。爬取任务通过 NATS JetStream（`keapp.crawl.trigger`）入队，由进程内 Worker 消费。取消操作通过数据库状态驱动。未来规划：kecore 将新增 `forest_type=website`，keapp 会将爬取结果推送至其中；本计划仅涵盖 keapp 侧的工作。

**技术栈：** Go 1.22+、GORM、nats.go v1.34.1（JetStream）、gin（通过 yg-go）、cobra（CLI）、html-to-markdown

## 全局约束

- Go module 路径：`github.com/insmtx/corekg`
- 构建使用 vendor 模式：`go build -mod=vendor`
- 依赖已 vendor 化；如需新增依赖，执行 `go mod tidy && go mod vendor`
- 路由使用 `eng.PRequireLogin("action.String", handler, middleware...)` 模式
- 所有 struct tag：`json` = snake_case
- 除非明确要求，否则不添加注释
- SQL 迁移脚本命名：`v2.18_0__alter_keapp_tables.sql`
- 测试：`go test ./apps/keapp/...`（需要 MySQL）

---

## 文件结构

```
apps/keapp/
├── app.go                                    # 修改：添加 Worker 启动
├── cmd/
│   ├── main.go                               # 修改：移除 Redis，添加 NATS
│   └── init.go                               # 修改：移除 Redis，添加 NATS 初始化
├── conf/test/config.yaml                     # 修改：移除 redis，添加 nats
├── internal/
│   ├── apis/
│   │   ├── apis.go                           # 修改：导入 webctl
│   │   ├── appctl/                           # 保留
│   │   └── webctl/                           # 新增
│   │       ├── registry.go
│   │       ├── crawl.go
│   │       ├── resource.go
│   │       ├── rule.go
│   │       └── dto.go
│   └── middleware/
│       └── appmds.go                         # 新增（从 mds/ 重构而来）
├── models/
│   ├── app/                                  # 保留
│   ├── apptype/
│   │   ├── application.go                    # 修改：+ForestID, +CrawlConfig
│   │   ├── db.go                             # 修改：调用 web.InitDB
│   │   ├── enums.go                          # 修改：移除 web 相关枚举
│   │   ├── template.go                       # 保留
│   │   └── resource_manifest.go              # 保留
│   └── web/                                  # 新增
│       ├── resource.go
│       ├── crawl_rule.go
│       ├── crawl_task.go
│       ├── enums.go
│       └── db.go
├── services/
│   ├── svcapp/                               # 保留
│   └── svcweb/                               # 新增
│       ├── resource_api.go
│       ├── crawl_api.go
│       └── rule_api.go
└── worker/                                   # 新增
    ├── crawler.go
    ├── subscriber.go
    └── config.go
```

迁移完成后删除：`apps/keapp/website/`、`apps/keapp/mds/appmds.go`

---

### 任务 1：数据库迁移脚本

**文件：**
- 新建：`scripts/mysql/v2.18_0__alter_keapp_tables.sql`

- [ ] **步骤 1：编写迁移脚本**

```sql
ALTER TABLE ke_application
    ADD COLUMN forest_id BIGINT UNSIGNED DEFAULT NULL AFTER id,
    ADD COLUMN crawl_config JSON DEFAULT NULL AFTER config;

ALTER TABLE ke_web_resource
    ADD COLUMN forest_file_id BIGINT UNSIGNED DEFAULT NULL AFTER id,
    ADD COLUMN raw_content LONGTEXT DEFAULT NULL AFTER metadata;

ALTER TABLE ke_crawl_task
    ADD COLUMN pages_new INT DEFAULT 0 AFTER pages_total,
    ADD COLUMN pages_updated INT DEFAULT 0 AFTER pages_new,
    ADD COLUMN pages_skipped INT DEFAULT 0 AFTER pages_updated;
```

- [ ] **步骤 2：提交**

```bash
git add scripts/mysql/v2.18_0__alter_keapp_tables.sql
git commit -m "feat(keapp): add v2.18 migration for web assistant refactor"
```

---

### 任务 2：重组 Models

**文件：**
- 新建：`apps/keapp/models/web/enums.go`
- 新建：`apps/keapp/models/web/resource.go`
- 新建：`apps/keapp/models/web/crawl_rule.go`
- 新建：`apps/keapp/models/web/crawl_task.go`
- 新建：`apps/keapp/models/web/db.go`
- 修改：`apps/keapp/models/apptype/application.go` — 添加 ForestID、CrawlConfig
- 修改：`apps/keapp/models/apptype/db.go` — 调用 web.InitDB
- 修改：`apps/keapp/models/apptype/enums.go` — 移除 web 相关枚举

**接口：**
- 产出：`web.NewWebResourceDao()`、`web.NewCrawlRuleDao()`、`web.NewCrawlTaskDao()`、`web.InitDB()`
- 产出：更新后的 `apptype.KeApplication`，包含 ForestID、CrawlConfig 字段

- [ ] **步骤 1：创建 models/web/enums.go**

将 CrawlRuleType、CrawlTaskType、CrawlTaskStatus、IndexStatus 从 apptype/enums.go 迁移过来。新增 `cancelled` 状态和 `single` 任务类型。

```go
package web

type CrawlRuleType string

const (
	CrawlRuleInclude CrawlRuleType = "include"
	CrawlRuleExclude CrawlRuleType = "exclude"
)

type CrawlTaskType string

const (
	CrawlTaskFull        CrawlTaskType = "full"
	CrawlTaskIncremental CrawlTaskType = "incremental"
	CrawlTaskSingle      CrawlTaskType = "single"
)

type CrawlTaskStatus string

const (
	CrawlTaskPending   CrawlTaskStatus = "pending"
	CrawlTaskRunning   CrawlTaskStatus = "running"
	CrawlTaskSuccess   CrawlTaskStatus = "success"
	CrawlTaskFailed    CrawlTaskStatus = "failed"
	CrawlTaskCancelled CrawlTaskStatus = "cancelled"
)

type IndexStatus string

const (
	IndexPending IndexStatus = "pending"
	IndexIndexed IndexStatus = "indexed"
	IndexFailed  IndexStatus = "failed"
)
```

- [ ] **步骤 2：从 apptype/enums.go 中移除 web 相关枚举**

删除 CrawlRuleType、CrawlTaskType、CrawlTaskStatus、IndexStatus 部分。保留 AppStatus、SyncStatus、AppTemplateType。

- [ ] **步骤 3：创建 models/web/resource.go**

KeWebResource struct（来自现有 website/models/types.go），新增字段 `ForestFileID *uint`、`RawContent string`。包含 WebResourceDao，方法有：Insert、GetByID、GetByURL、ListByAppID、Update、SoftDelete。遵循现有 DAO 模式，使用 `dbutil.Knownow()`。

- [ ] **步骤 4：创建 models/web/crawl_rule.go**

KeWebCrawlRule struct + CrawlRuleDao，方法有：Insert、ListByAppID、Update、Delete（硬删除）、GetByID。

- [ ] **步骤 5：创建 models/web/crawl_task.go**

KeCrawlTask struct，新增字段 PagesNew、PagesUpdated、PagesSkipped。移除 RedisTaskID。CrawlTaskDao 方法有：Insert、GetByID、UpdateStatus、UpdateProgress、CancelTask、ListByAppID、GetPendingAndRunning。

- [ ] **步骤 6：创建 models/web/db.go**

InitDB 自动迁移全部三个 web model。

- [ ] **步骤 7：修改 apptype/application.go**

在 KeApplication struct 中添加：`ForestID *uint` 和 `CrawlConfig string`（json 类型）。

- [ ] **步骤 8：修改 apptype/db.go**

导入 `models/web`，在 apptype 迁移之后调用 `web.InitDB(db)`。

- [ ] **步骤 9：验证编译**

执行：`go build -mod=vendor ./apps/keapp/models/...`
预期结果：构建成功

- [ ] **步骤 10：提交**

```bash
git add apps/keapp/models/
git commit -m "feat(keapp): reorganize models to models/web, add new fields"
```

---

### 任务 3：重组 Services

**文件：**
- 新建：`apps/keapp/services/svcweb/resource_api.go`
- 新建：`apps/keapp/services/svcweb/rule_api.go`
- 新建：`apps/keapp/services/svcweb/crawl_api.go`

**接口：**
- 消费：任务 2 中的 models/web DAO
- 产出：svcweb.TriggerCrawl、GetCrawlTask、ListCrawlTasks、CancelCrawlTask、RecrawlResource、ListWebResources、GetWebResource、DeleteWebResource、AddCrawlRule、ListCrawlRules、UpdateCrawlRule、DeleteCrawlRule

- [ ] **步骤 1：创建 svcweb/resource_api.go**

薄封装层：ListWebResources、GetWebResource、DeleteWebResource，委托给 web DAO 处理。

- [ ] **步骤 2：创建 svcweb/rule_api.go**

薄封装层：AddCrawlRule、ListCrawlRules、UpdateCrawlRule、DeleteCrawlRule。

- [ ] **步骤 3：创建 svcweb/crawl_api.go**

通过包级别 setter `SetNATSConn(*nats.Conn)` 获取 NATS 连接。`TriggerCrawl` 创建任务记录并发布 `keapp.crawl.trigger` 消息，内容为 `{task_id}`。`GetCrawlTask`、`ListCrawlTasks` 查询数据库。`CancelCrawlTask` 调用 `dao.CancelTask`。`RecrawlResource` 获取资源 URL 后以 CrawlTaskSingle 类型调用 TriggerCrawl。

- [ ] **步骤 4：验证编译**

执行：`go build -mod=vendor ./apps/keapp/services/...`
预期结果：构建成功

- [ ] **步骤 5：提交**

```bash
git add apps/keapp/services/svcweb/
git commit -m "feat(keapp): add svcweb services with NATS publish"
```

---

### 任务 4：重构权限中间件

**文件：**
- 新建：`apps/keapp/internal/middleware/appmds.go`

**接口：**
- 产出：middleware.AppContextMiddleware()、middleware.GetAppID(ctx)、middleware.RequireAppViewPerm()、middleware.RequireAppManagePerm()

- [ ] **步骤 1：创建 internal/middleware/appmds.go**

先阅读现有 `apps/keapp/mds/appmds.go` 以匹配 yg-go 运行时 API 的准确用法。重构为：
1. AppContextMiddleware：仅解析一次 body，提取 app_id，存入 context.WithValue
2. RequireAppViewPerm/ManagePerm：从 context 读取 app_id（不再重复解析 body）
3. 使用 svcapp.CheckAppPermission / CheckAppManagePermission 进行权限校验

- [ ] **步骤 2：验证编译**

执行：`go build -mod=vendor ./apps/keapp/internal/middleware/...`

- [ ] **步骤 3：提交**

```bash
git add apps/keapp/internal/middleware/appmds.go
git commit -m "refactor(keapp): unified AppContextMiddleware"
```

---

### 任务 5：创建 Web Handlers（webctl）

**文件：**
- 新建：`apps/keapp/internal/apis/webctl/dto.go`
- 新建：`apps/keapp/internal/apis/webctl/resource.go`
- 新建：`apps/keapp/internal/apis/webctl/rule.go`
- 新建：`apps/keapp/internal/apis/webctl/crawl.go`
- 新建：`apps/keapp/internal/apis/webctl/registry.go`

**接口：**
- 消费：svcweb 函数、middleware 权限函数
- 产出：webctl.RegistryRouter(eng)

- [ ] **步骤 1：创建 webctl/dto.go**

为全部 12 个端点定义 DTO。遵循现有 appctl/dto.go 模式，嵌入 BaseRequest/BaseResponse。

- [ ] **步骤 2：创建 webctl/resource.go**

Handler：ListResources、GetResource、DeleteResource、RecrawlResource。遵循 appctl/app.go handler 模式。

- [ ] **步骤 3：创建 webctl/rule.go**

Handler：AddCrawlRule、ListCrawlRules、UpdateCrawlRule、DeleteCrawlRule。

- [ ] **步骤 4：创建 webctl/crawl.go**

Handler：TriggerCrawl、GetCrawlTask、ListCrawlTasks、CancelCrawlTask。

- [ ] **步骤 5：创建 webctl/registry.go**

注册所有 `keapp.web.*` 路由及中间件。使用新的 middleware 包。

- [ ] **步骤 6：修改 internal/apis/apis.go**

导入 webctl，在现有注册逻辑旁添加 `webctl.RegistryRouter(eng)`。移除旧的 website 导入。

- [ ] **步骤 7：验证编译**

执行：`go build -mod=vendor ./apps/keapp/...`
预期结果：构建成功

- [ ] **步骤 8：提交**

```bash
git add apps/keapp/internal/apis/
git commit -m "feat(keapp): add webctl handlers and registry"
```

---

### 任务 6：NATS Worker

**文件：**
- 新建：`apps/keapp/worker/config.go`
- 新建：`apps/keapp/worker/crawler.go`
- 新建：`apps/keapp/worker/subscriber.go`

**接口：**
- 消费：models/web DAO、svcweb.SetNATSConn 模式
- 产出：worker.Start(ctx, nc, cfg) — 从 app.go RunJob 中调用

- [ ] **步骤 1：创建 worker/config.go**

Config struct：Concurrency int、CancelCheckInterval int、StreamName string、Subject string、ConsumerName string、MaxDeliver int、AckWait duration。从应用配置中加载。

- [ ] **步骤 2：创建 worker/crawler.go**

核心爬取逻辑：fetchPage(url) -> HTML，convertToMarkdown(html) -> string，computeHash(content) -> content_hash。BFS 遍历，遵守 max_depth、max_pages 限制。匹配爬取规则。对每个页面：检查取消状态（每 N 个页面通过数据库查询），创建/更新 KeWebResource。返回统计信息（new、updated、skipped）。

使用 `github.com/JohannesKaufmann/html-to-markdown` 进行转换（先检查是否已 vendor，若未包含则添加到 go.mod）。

- [ ] **步骤 3：创建 worker/subscriber.go**

Start(ng *nats.Conn, cfg Config)：如果 JetStream stream 不存在则创建，创建持久化 pull subscriber，启动 N 个 goroutine 拉取消息。每条消息：反序列化 CrawlTriggerMsg，从数据库查询任务，如已取消则跳过，更新状态为 running，调用 crawler，用结果/错误更新任务。

启动恢复：查询 pending/running 状态的任务，将每条任务发布到 trigger subject。

- [ ] **步骤 4：验证编译**

执行：`go build -mod=vendor ./apps/keapp/worker/...`
预期结果：构建成功

- [ ] **步骤 5：提交**

```bash
git add apps/keapp/worker/
git commit -m "feat(keapp): add NATS JetStream crawl worker"
```

---

### 任务 7：串联应用启动流程

**文件：**
- 修改：`apps/keapp/cmd/init.go`
- 修改：`apps/keapp/cmd/main.go`
- 修改：`apps/keapp/app.go`
- 修改：`apps/keapp/conf/test/config.yaml`

**接口：**
- 消费：worker.Start、svcweb.SetNATSConn
- 产出：完整串联的应用

- [ ] **步骤 1：修改 conf/test/config.yaml**

移除 redis 配置段。添加 nats 配置段：`nats: { url: "nats://localhost:4222", stream: "KEAPP_CRAWL" }`。添加 worker 配置段。

- [ ] **步骤 2：修改 cmd/init.go**

移除 `redispool.InitRedis()` 调用。添加 NATS 连接初始化：`nats.Connect(cfg.NATS.URL)`。将 *nats.Conn 存储在包变量中或返回它。

- [ ] **步骤 3：修改 cmd/main.go**

将 NATS 连接传入初始化链。服务器启动后，在 goroutine 中调用 worker.Start。

- [ ] **步骤 4：修改 app.go**

在 RunJob 中：调用 `svcweb.SetNATSConn(nc)` 和 `worker.Start(ctx, nc, cfg)`。

- [ ] **步骤 5：验证完整构建**

执行：`go build -mod=vendor ./apps/keapp/...`
预期结果：构建成功

- [ ] **步骤 6：提交**

```bash
git add apps/keapp/cmd/ apps/keapp/app.go apps/keapp/conf/
git commit -m "feat(keapp): wire NATS + worker at startup, remove Redis"
```

---

### 任务 8：清理并删除旧文件

**文件：**
- 删除：`apps/keapp/website/`（整个目录）
- 删除：`apps/keapp/mds/appmds.go`

- [ ] **步骤 1：移除旧 website/ 目录**

```bash
rm -rf apps/keapp/website/
```

- [ ] **步骤 2：移除旧 mds/appmds.go**

```bash
rm apps/keapp/mds/appmds.go
rmdir apps/keapp/mds/ 2>/dev/null || true
```

- [ ] **步骤 3：更新所有仍引用旧路径的 import**

在整个代码库中搜索 `keapp/website`、`keapp/mds` 的导入（包括 apps/corekg，它也导入了 keapp）。更新为新路径。

- [ ] **步骤 4：验证完整构建**

执行：`go build -mod=vendor ./apps/keapp/...`
预期结果：构建成功

- [ ] **步骤 5：验证 corekg 单体构建**

执行：`go build -mod=vendor ./apps/corekg/...`
预期结果：构建成功

- [ ] **步骤 6：提交**

```bash
git add -A apps/keapp/
git commit -m "refactor(keapp): remove old website/ and mds/ directories"
```

---

### 任务 9：重新生成 Swagger 文档

- [ ] **步骤 1：运行 swagger 生成**

执行：`make generate-docs APP=keapp`
预期结果：更新 `apps/keapp/internal/docs/keapp_docs.go`

- [ ] **步骤 2：验证 swagger 重新生成后的构建**

执行：`go build -mod=vendor ./apps/keapp/...`

- [ ] **步骤 3：提交**

```bash
git add apps/keapp/internal/docs/
git commit -m "docs(keapp): regenerate swagger docs"
```

---

### 任务 10：Vendor 更新（仅在新增依赖时执行）

仅在添加了 html-to-markdown 或其他新依赖时执行：

- [ ] **步骤 1：更新 vendor**

```bash
go mod tidy && go mod vendor
```

- [ ] **步骤 2：验证构建**

执行：`go build -mod=vendor ./apps/keapp/...`

- [ ] **步骤 3：提交**

```bash
git add go.mod go.sum vendor/
git commit -m "chore(keapp): update vendor for new dependencies"
```
