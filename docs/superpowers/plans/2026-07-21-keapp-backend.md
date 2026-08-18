# keapp Go 后端实施计划

> **给智能体工作者：** 必需的子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 来逐任务实施本计划。步骤使用复选框（`- [ ]`）语法进行跟踪。

**目标：** 实现 `keapp` Go 子应用，提供 Application CRUD、Web 资源管理、爬取任务调度和证据搜索 API。

**架构：** 独立子应用，遵循 kecore 分层模式（model/types → DAO → service → DTO → handler → router）。复用 `dbutil.Knownow()` 数据库连接和 `KeResourceScope` 权限系统。通过 `Routers()` 和 `DoInitModels()` 集成到 corekg 聚合体中。

**技术栈：** Go, GORM, Gin (yg-go framework), MySQL, Redis, Elasticsearch

## 全局约束

- Go module 路径前缀：`github.com/insmtx/corekg/...`
- vendor 目录已提交；修改依赖需执行 `go mod tidy && go mod vendor`
- JSON struct tag：snake_case
- GORM model 嵌入 `gorm.Model`；表名常量定义在 `db.go` 中
- DAO 模式：`NewXxxDao()` + `WithTx(db)` + 嵌入 `BaseModel`
- Service：包级函数 + 哨兵错误
- Handler 签名：`(ctx *gin.Context, req *ReqDTO, resp *RespDTO)`
- Handler 校验：`req.Validity(resp); if resp.Code != errcode.CodeOK { return }`
- Router：`eng.PRequireLogin("keapp.Action", middleware, handler)`
- 权限：`perm.HasAct(ctx, uin, resourceID, ResourceTypeApp, action)` / `perm.HasManageAct(...)`
- Scope 创建：`forest.NewKeResourceScopeDao().WithTx(tx).Insert(ctx, &scope)`（不存在 `perm.SetScope`）
- 迁移脚本命名：`v2.17_X__action.sql`，全小写
- API 前缀：`keapp.` → `/v3/keapp.ActionName`

---

### 任务 1：Model 类型 — 枚举

**文件：**
- 新建：`apps/keapp/models/apptype/enums.go`

**接口：**
- 依赖：无
- 产出：`AppStatus`、`SyncStatus`、`AppTemplateType`、`CrawlRuleType`、`CrawlTaskType`、`CrawlTaskStatus`、`IndexStatus`

- [ ] **步骤 1：创建 enums.go**

```go
package apptype

type AppStatus string

const (
	AppStatusDraft  AppStatus = "draft"
	AppStatusOnline AppStatus = "online"
	AppStatusPaused AppStatus = "paused"
)

type SyncStatus string

const (
	SyncStatusSuccess SyncStatus = "success"
	SyncStatusFailed  SyncStatus = "failed"
	SyncStatusSyncing SyncStatus = "syncing"
)

type AppTemplateType string

const (
	AppTemplateWebsite    AppTemplateType = "website"
	AppTemplateProduct    AppTemplateType = "product"
	AppTemplateAftersales AppTemplateType = "aftersales"
	AppTemplateTraining   AppTemplateType = "training"
	AppTemplatePolicy     AppTemplateType = "policy"
)

type CrawlRuleType string

const (
	CrawlRuleInclude CrawlRuleType = "include"
	CrawlRuleExclude CrawlRuleType = "exclude"
)

type CrawlTaskType string

const (
	CrawlTaskFull        CrawlTaskType = "full"
	CrawlTaskIncremental CrawlTaskType = "incremental"
	CrawlTaskDeleted     CrawlTaskType = "deleted"
)

type CrawlTaskStatus string

const (
	CrawlTaskPending CrawlTaskStatus = "pending"
	CrawlTaskRunning CrawlTaskStatus = "running"
	CrawlTaskSuccess CrawlTaskStatus = "success"
	CrawlTaskFailed  CrawlTaskStatus = "failed"
)

type IndexStatus string

const (
	IndexStatusPending IndexStatus = "pending"
	IndexStatusIndexed IndexStatus = "indexed"
	IndexStatusFailed  IndexStatus = "failed"
)
```

- [ ] **步骤 2：验证编译**

运行：`go build ./apps/keapp/models/apptype/`
预期：通过

- [ ] **步骤 3：提交**

```bash
git add apps/keapp/models/apptype/enums.go && git commit -m "feat(keapp): add enum types"
```

---

### 任务 2：Model 类型 — 模板配置

**文件：**
- 新建：`apps/keapp/models/apptype/template.go`

**接口：**
- 依赖：任务 1 的 `AppTemplateType`
- 产出：`AppCapabilities`、`WebsiteConfig`、`ProductConfig`、`AftersalesConfig`、`TrainingConfig`、`PolicyConfig`、`AppConfig`，含 `Value()/Scan()` 和 `AsWebsite()/...` 方法

- [ ] **步骤 1：创建 template.go**

```go
package apptype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type AppCapabilities struct {
	AIAssistant bool `json:"ai_assistant"`
	Search      bool `json:"search"`
	FAQ         bool `json:"faq"`
	Widget      bool `json:"widget"`
}

type WebsiteConfig struct {
	URL           string          `json:"url"`
	SyncSchedule  string          `json:"sync_schedule"`
	MaxDepth      int             `json:"max_depth"`
	MaxPages      int             `json:"max_pages"`
	RespectRobots bool            `json:"respect_robots"`
	Capabilities  AppCapabilities `json:"capabilities"`
}

type ProductConfig struct {
	ProductName  string          `json:"product_name"`
	Capabilities AppCapabilities `json:"capabilities"`
}

type AftersalesConfig struct {
	ServiceScope string          `json:"service_scope"`
	Capabilities AppCapabilities `json:"capabilities"`
}

type TrainingConfig struct {
	Department   string          `json:"department"`
	Capabilities AppCapabilities `json:"capabilities"`
}

type PolicyConfig struct {
	OrgName      string          `json:"org_name"`
	Capabilities AppCapabilities `json:"capabilities"`
}

type AppConfig struct {
	Type   AppTemplateType `json:"type"`
	Config json.RawMessage `json:"config"`
}

func (c *AppConfig) Value() (driver.Value, error) {
	if c == nil { return nil, nil }
	return json.Marshal(c)
}

func (c *AppConfig) Scan(value interface{}) error {
	if value == nil { c.Type = ""; c.Config = nil; return nil }
	bytes, ok := value.([]byte)
	if !ok { return fmt.Errorf("failed to scan AppConfig: %T", value) }
	return json.Unmarshal(bytes, c)
}

func (c *AppConfig) AsWebsite() (*WebsiteConfig, error) {
	var cfg WebsiteConfig; err := json.Unmarshal(c.Config, &cfg); return &cfg, err
}
func (c *AppConfig) AsProduct() (*ProductConfig, error) {
	var cfg ProductConfig; err := json.Unmarshal(c.Config, &cfg); return &cfg, err
}
func (c *AppConfig) AsAftersales() (*AftersalesConfig, error) {
	var cfg AftersalesConfig; err := json.Unmarshal(c.Config, &cfg); return &cfg, err
}
func (c *AppConfig) AsTraining() (*TrainingConfig, error) {
	var cfg TrainingConfig; err := json.Unmarshal(c.Config, &cfg); return &cfg, err
}
func (c *AppConfig) AsPolicy() (*PolicyConfig, error) {
	var cfg PolicyConfig; err := json.Unmarshal(c.Config, &cfg); return &cfg, err
}
```

- [ ] **步骤 2：验证 + 提交**

运行：`go build ./apps/keapp/models/apptype/`
提交：`git add apps/keapp/models/apptype/template.go && git commit -m "feat(keapp): add template config structs"`

---

### 任务 3：Model 类型 — AIOS 资源清单

**文件：**
- 新建：`apps/keapp/models/apptype/resource_manifest.go`

**接口：**
- 产出：`ResourceManifest`、`ContentUnit`、`ContentLocator`、`Evidence`、`EvidenceLocator`、`IndexArtifact`、`IndexBuilder` 接口

- [ ] **步骤 1：创建 resource_manifest.go**

```go
package apptype

type ResourceManifest struct {
	ResourceID   string         `json:"resource_id"`
	ResourceType string         `json:"resource_type"`
	SourceURL    string         `json:"source_url"`
	ContentUnits []ContentUnit  `json:"content_units"`
	Metadata     map[string]any `json:"metadata"`
}

type ContentUnit struct {
	Type     string         `json:"type"`
	Location ContentLocator `json:"location"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata"`
}

type ContentLocator struct {
	Selector string `json:"selector,omitempty"`
	Index    int    `json:"index,omitempty"`
	Title    string `json:"title,omitempty"`
}

type Evidence struct {
	ResourceID  string          `json:"resource_id"`
	ContentType string          `json:"content_type"`
	Score       float64         `json:"score"`
	Locator     EvidenceLocator `json:"locator"`
	Snippet     string          `json:"snippet"`
	Payload     map[string]any  `json:"payload"`
}

type EvidenceLocator struct {
	URL      string `json:"url,omitempty"`
	Title    string `json:"title,omitempty"`
	Selector string `json:"selector,omitempty"`
}

type IndexArtifact struct {
	IndexType  string `json:"index_type"`
	DocumentID string `json:"document_id"`
	Payload    any    `json:"payload"`
}

type IndexBuilder interface {
	Build(manifest ResourceManifest) ([]IndexArtifact, error)
	IndexType() string
}
```

- [ ] **步骤 2：验证 + 提交**

运行：`go build ./apps/keapp/models/apptype/`
提交：`git add apps/keapp/models/apptype/resource_manifest.go && git commit -m "feat(keapp): add AIOS resource manifest types"`

---

### 任务 4：Model 类型 — GORM Model + db.go

**文件：**
- 新建：`apps/keapp/models/apptype/application.go`
- 新建：`apps/keapp/models/apptype/web_resource.go`
- 新建：`apps/keapp/models/apptype/web_crawl_rule.go`
- 新建：`apps/keapp/models/apptype/crawl_task.go`
- 新建：`apps/keapp/models/apptype/db.go`

**接口：**
- 依赖：枚举（任务 1）、AppConfig（任务 2）
- 产出：所有 GORM model、`InitDB()`、表名常量

- [ ] **步骤 1：创建 application.go** — `KeApplication` + `KeApplicationList`（参见规格文档 4.2 节）
- [ ] **步骤 2：创建 web_resource.go** — `KeWebResource`（参见规格文档 4.2 节）
- [ ] **步骤 3：创建 web_crawl_rule.go** — `KeWebCrawlRule`（参见规格文档 4.2 节）
- [ ] **步骤 4：创建 crawl_task.go** — `KeCrawlTask`（参见规格文档 4.2 节）
- [ ] **步骤 5：创建 db.go**

```go
package apptype

import (
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

const (
	TableNamePrefix         = "ke_"
	TableNameKeApplication  = TableNamePrefix + "application"
	TableNameKeWebResource  = TableNamePrefix + "web_resource"
	TableNameKeWebCrawlRule = TableNamePrefix + "web_crawl_rule"
	TableNameKeCrawlTask    = TableNamePrefix + "crawl_task"
)

func InitDB() error {
	return dbtools.InitModel(dbutil.Knownow(),
		&KeApplication{},
		&KeWebResource{},
		&KeWebCrawlRule{},
		&KeCrawlTask{},
	)
}
```

- [ ] **步骤 6：验证 + 提交**

运行：`go build ./apps/keapp/models/apptype/`
提交：`git add apps/keapp/models/apptype/ && git commit -m "feat(keapp): add GORM models and db.go"`

---

### 任务 5：DAO 层

**文件：**
- 新建：`apps/keapp/models/app/base.go`
- 新建：`apps/keapp/models/app/ke_application.go`
- 新建：`apps/keapp/models/app/ke_web_resource.go`
- 新建：`apps/keapp/models/app/ke_web_crawl_rule.go`
- 新建：`apps/keapp/models/app/ke_crawl_task.go`

**接口：**
- 依赖：任务 4 的所有 GORM model、`dbutil.Knownow()`
- 产出：`ApplicationDao`、`WebResourceDao`、`WebCrawlRuleDao`、`CrawlTaskDao`

- [ ] **步骤 1：创建 base.go** — 从 `apps/kecore/models/forest/base.go` 完整复制 `BaseModel` + `BaseCond` + `BuildBaseCondition`，默认使用 `dbutil.Knownow()`。
- [ ] **步骤 2：创建 ke_application.go** — `ApplicationDao`，包含：`NewApplicationDao()`、`TableName()`、`WithTx()`、`Insert()`、`GetByID()`、`UpdateByID()`、`UpdateMap()`、`SoftDelete()`、`CheckNameExists()`、`GetPageListByCond()`、`BuildCondition()`。完全遵循 `apps/kecore/models/forest/ke_forest.go` 的模式。
- [ ] **步骤 3：创建 ke_web_resource.go** — `WebResourceDao`，包含：`Insert()`、`GetByID()`、`UpdateManifest()`、`UpdateIndexStatus()`、`ListByAppID()`、`GetByURL()`。
- [ ] **步骤 4：创建 ke_web_crawl_rule.go** — `WebCrawlRuleDao`，包含：`ListByAppID()`、`UpsertRule()`、`DeleteRule()`。
- [ ] **步骤 5：创建 ke_crawl_task.go** — `CrawlTaskDao`，包含：`Insert()`、`UpdateStatus()`、`GetLatestByAppID()`。
- [ ] **步骤 6：验证 + 提交**

运行：`go build ./apps/keapp/models/app/`
提交：`git add apps/keapp/models/app/ && git commit -m "feat(keapp): add DAO layer"`

---

### 任务 6：Service 层

**文件：**
- 新建：`apps/keapp/services/svcapp/app_api.go`
- 新建：`apps/keapp/services/svcapp/web_resource_api.go`
- 新建：`apps/keapp/services/svcapp/crawl_api.go`
- 新建：`apps/keapp/services/svcapp/evidence_api.go`
- 新建：`apps/keapp/services/svcapp/permission_api.go`

**接口：**
- 依赖：DAO 层（任务 5）、`forest.NewKeResourceScopeDao()` 用于插入 scope、`perm.HasAct()/HasManageAct()` 用于权限校验
- 产出：任务 8 中 handler 消费的所有 service 函数

- [ ] **步骤 1：创建 app_api.go** — 哨兵错误 + `CreateApplication`（事务：插入 app + 通过 `forest.NewKeResourceScopeDao().WithTx(tx).Insert()` 插入 resource scope）、`GetApplication`、`ListApplications`、`UpdateApplication`（通过 map 部分更新，名称唯一性校验）、`DeleteApplication`（公司校验 + 软删除）。完全遵循 `apps/kecore/services/svcforest/forest_api.go` 的模式。
- [ ] **步骤 2：创建 web_resource_api.go** — `ListWebResources`、`GetWebResource`、`AddCrawlRule`、`ListCrawlRules`、`DeleteCrawlRule`。
- [ ] **步骤 3：创建 crawl_api.go** — `TriggerFullCrawl`（创建任务 + 发送 Redis 消息）、`TriggerIncrementalSync`（对比 ETag/Hash）、`GetCrawlTaskStatus`、`GetLatestCrawlTask`、`HandleCrawlCallback`（解析 manifest、upsert resource、调用 IndexBuilder、更新 ES）。
- [ ] **步骤 4：创建 evidence_api.go** — `SearchEvidence`（FAQ 匹配 → ES Vector+BM25 → Evidence 格式化 → Rerank）。
- [ ] **步骤 5：创建 permission_api.go** — `UpdateAppPermission`、`CheckAppPermission`（对 `perm.HasAct`/`perm.HasManageAct` 的封装）。
- [ ] **步骤 6：验证 + 提交**

运行：`go build ./apps/keapp/services/...`
提交：`git add apps/keapp/services/ && git commit -m "feat(keapp): add service layer"`

---

### 任务 7：DTO 层

**文件：**
- 新建：`apps/keapp/internal/dto/dtokeapp/app.go`
- 新建：`apps/keapp/internal/dto/dtokeapp/web_resource.go`
- 新建：`apps/keapp/internal/dto/dtokeapp/crawl.go`
- 新建：`apps/keapp/internal/dto/dtokeapp/evidence.go`

**接口：**
- 依赖：`apiobj.BaseRequest`/`BaseResponse`、`errcode`、`apptype` model
- 产出：任务 8 中 handler 消费的所有 DTO

完全遵循 `apps/kecore/internal/apis/forestctl/forest_biz.go` 的模式（DTO 定义在 handler 包中）— 或使用 `apps/keapi/internal/dto/dtokeapi/` 的独立 dto 包模式。由于规格要求 `internal/dto/dtokeapp/`，使用独立包。

- [ ] **步骤 1-4：创建所有 DTO 文件**，嵌入 `BaseRequest`/`BaseResponse`，包含 `Validity()` 方法，JSON tag 使用 snake_case。
- [ ] **步骤 5：验证 + 提交**

运行：`go build ./apps/keapp/internal/dto/...`
提交：`git add apps/keapp/internal/dto/ && git commit -m "feat(keapp): add DTO layer"`

---

### 任务 8：Handler + Router + 中间件

**文件：**
- 新建：`apps/keapp/internal/apis/appctl/app.go`
- 新建：`apps/keapp/internal/apis/appctl/web_resource.go`
- 新建：`apps/keapp/internal/apis/appctl/crawl.go`
- 新建：`apps/keapp/internal/apis/appctl/evidence.go`
- 新建：`apps/keapp/internal/apis/apis.go`
- 新建：`apps/keapp/mds/appmds.go`

**接口：**
- 依赖：DTO（任务 7）、Service 函数（任务 6）、`runtime.Uin()`/`runtime.CompanyID()`
- 产出：`RegistryRouter()` 函数，由 `app.go` 消费

- [ ] **步骤 1：创建中间件** — `appmds.go`，包含 `RequireAppCreatePerm`、`RequireAppViewPerm`（解析请求体 → 提取 app_id → `perm.HasAct`）、`RequireAppManagePerm`。
- [ ] **步骤 2：创建 handler** — 每个 handler 遵循：`Validity()` → `runtime.Uin/CompanyID` → 调用 service → 通过 `errors.Is` 映射错误 → 填充响应。添加 swagger 注解。
- [ ] **步骤 3：创建 apis.go** — 使用 `eng.PRequireLogin` 注册全部 14 条路由。
- [ ] **步骤 4：验证 + 提交**

运行：`go build ./apps/keapp/internal/apis/` 和 `go build ./apps/keapp/mds/`
提交：`git add apps/keapp/internal/apis/ apps/keapp/mds/ && git commit -m "feat(keapp): add handlers, router and middleware"`

---

### 任务 9：应用外壳 + 配置

**文件：**
- 新建：`apps/keapp/app.go`
- 新建：`apps/keapp/cmd/main.go`
- 新建：`apps/keapp/cmd/init.go`
- 新建：`apps/keapp/conf/test/config.yaml`

**接口：**
- 依赖：`apis.RegistryRouter`（任务 8）、`apptype.InitDB`（任务 4）
- 产出：`Routers()`、`Migrates()`、`RunJob()`，由 corekg 消费

- [ ] **步骤 1：创建 app.go** — 完全参照 `apps/kechat/app.go` 的模式。
- [ ] **步骤 2：创建 cmd/init.go** — 参照 `apps/kechat/cmd/init.go` 的模式，在 `DoInitModels` 中调用 `apptype.InitDB`。
- [ ] **步骤 3：创建 cmd/main.go** — 参照 `apps/kechat/cmd/main.go` 的模式，前缀 `/v3/`，调用 `keapp.Routers(svr)` + `keapp.RunJob()`。
- [ ] **步骤 4：创建 conf/test/config.yaml** — 参照 `apps/kechat/conf/test/config.yaml` 的模式，应用名 `keapp`，端口 `:8089`。
- [ ] **步骤 5：验证独立构建**

运行：`make local APP=keapp`
预期：生成 `bundles/keapp` 二进制文件

- [ ] **步骤 6：提交**

```bash
git add apps/keapp/app.go apps/keapp/cmd/ apps/keapp/conf/ && git commit -m "feat(keapp): add app shell and config"
```

---

### 任务 10：MySQL 迁移脚本

**文件：**
- 新建：`scripts/mysql/v2.17_0__create_keapp_tables.sql`

- [ ] **步骤 1：创建迁移脚本** — 规格文档 4.1 节中的全部 4 条 CREATE TABLE 语句。
- [ ] **步骤 2：提交**

```bash
git add scripts/mysql/v2.17_0__create_keapp_tables.sql && git commit -m "feat(keapp): add migration script for 4 tables"
```

---

### 任务 11：集成到 Corekg 聚合体

**文件：**
- 修改：`apps/corekg/app.go`（约第 14 行，添加 `keapp.Routers(eng)`）
- 修改：`apps/corekg/cmd/init.go`（约第 45 行，在 `DoInitModels` 中添加 `apptype.InitDB`）

- [ ] **步骤 1：在 corekg/app.go 中添加 import 和 Routers 调用**

Import：`"github.com/insmtx/corekg/apps/keapp"`
在 `Routers()` 中：在现有子应用调用之后添加 `keapp.Routers(eng)`。

- [ ] **步骤 2：在 corekg/cmd/init.go 中添加 InitDB**

Import：`"github.com/insmtx/corekg/apps/keapp/models/apptype"`
在 `DoInitModels` 中：将 `apptype.InitDB` 添加到列表中。

- [ ] **步骤 3：验证聚合体构建**

运行：`make local APP=corekg`
预期：生成 `bundles/corekg` 二进制文件，无错误。

- [ ] **步骤 4：提交**

```bash
git add apps/corekg/app.go apps/corekg/cmd/init.go && git commit -m "feat(corekg): integrate keapp sub-app into aggregate"
```

---

### 任务 12：Swagger 文档生成

- [ ] **步骤 1：生成文档**

运行：`make generate-docs APP=keapp`
预期：`apps/keapp/internal/docs/` 目录被填充。

- [ ] **步骤 2：提交**

```bash
git add apps/keapp/internal/docs/ && git commit -m "docs(keapp): generate swagger documentation"
```

---

## 验证清单

- [ ] `make local APP=keapp` — 独立二进制构建成功
- [ ] `make local APP=corekg` — 聚合体二进制构建成功，keapp 已集成
- [ ] `make generate-docs APP=keapp` — Swagger 文档已生成
- [ ] `go vet ./apps/keapp/...` — 无 vet 错误
- [ ] 迁移脚本 SQL 语法正确
