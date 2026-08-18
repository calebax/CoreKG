# keapp 目录重组实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将现有 keapp 扁平结构重组为「公共层 + 具体应用模块」的分层结构，使每个具体应用（website 等）自包含完整 MVC。

**Architecture:** 根目录 models/services/apis/mds 为公共框架层；website/ 为第一个自包含具体应用模块。通过 git mv 移动文件 + 修改 package/import 路径 + 更新 apis.go 聚合。

**Tech Stack:** Go, GORM, Gin (yg-go)

## Global Constraints

- Go module: `github.com/insmtx/corekg/...`
- 不添加注释
- 不运行 `go mod tidy` / `go mod vendor`
- JSON tags: snake_case
- API 前缀保持 `keapp.*`
- 各模块独立 `InitDB()`
- 重组后 `make local APP=keapp` 和 `make local APP=corekg` 均必须通过

---

### Task 1: 移动网站助手 Model 到 website/models/

**操作：**
1. 创建 `apps/keapp/website/models/` 目录
2. 将 `apps/keapp/models/apptype/web_resource.go` → `apps/keapp/website/models/types.go`（与 web_crawl_rule.go + crawl_task.go 合并为一个 types.go）
3. 将 `apps/keapp/models/app/ke_web_resource.go` + `ke_web_crawl_rule.go` + `ke_crawl_task.go` → `apps/keapp/website/models/dao.go`（合并为一个 dao.go）
4. 创建 `apps/keapp/website/models/db.go`（网站助手专属 InitDB）
5. 修改所有文件的 `package` 声明和 import 路径
6. 从 `apps/keapp/models/apptype/db.go` 移除 KeWebResource/KeWebCrawlRule/KeCrawlTask 的 InitDB 注册
7. 将 `apps/keapp/models/app/base.go` 复制到 `apps/keapp/website/models/base.go`（website DAO 也需要 BaseModel）

**验证：** `go build ./apps/keapp/website/models/`

**Commit:** `git add -A && git commit -m "refactor(keapp): move website models to website/models/"`

---

### Task 2: 移动网站助手 Service 到 website/services/

**操作：**
1. 创建 `apps/keapp/website/services/` 目录
2. 移动 `apps/keapp/services/svcapp/web_resource_api.go` → `apps/keapp/website/services/web_resource_api.go`
3. 移动 `apps/keapp/services/svcapp/crawl_api.go` → `apps/keapp/website/services/crawl_api.go`
4. 移动 `apps/keapp/services/svcapp/evidence_api.go` → `apps/keapp/website/services/evidence_api.go`
5. 修改 package 为 `svcwebsite`，更新 import 指向新的 model 路径

**验证：** `go build ./apps/keapp/website/services/`

**Commit:** `git add -A && git commit -m "refactor(keapp): move website services to website/services/"`

---

### Task 3: 移动网站助手 Handler 到 website/apis/

**操作：**
1. 创建 `apps/keapp/website/apis/` 目录
2. 从 `apps/keapp/internal/apis/appctl/app.go` 中提取网站助手相关 handler（ListWebResources, AddCrawlRule, TriggerCrawl, SearchEvidence）→ `apps/keapp/website/apis/` 下对应文件
3. 从 `apps/keapp/internal/apis/appctl/dto.go` 中提取网站助手相关 DTO → `apps/keapp/website/apis/dto.go`
4. 创建 `apps/keapp/website/apis/registry.go` 导出 `RegistryRouter(eng)` 函数
5. 修改 `apis/appctl/app.go` 移除已迁出的 handler（仅保留 Application CRUD）
6. 修改 `apis/appctl/dto.go` 移除已迁出的 DTO
7. 更新 import 指向新的 service/model 路径

**验证：** `go build ./apps/keapp/website/apis/` 和 `go build ./apps/keapp/apis/...`

**Commit:** `git add -A && git commit -m "refactor(keapp): move website handlers to website/apis/"`

---

### Task 4: 更新 apis/apis.go 聚合 + corekg InitDB

**操作：**
1. 修改 `apps/keapp/apis/apis.go`：调用 `appctl.RegistryRouter(eng)` + `websitectl.RegistryRouter(eng)`
2. 创建 `apps/keapp/apis/appctl/registry.go` 导出 `RegistryRouter(eng)`（从原 apis.go 中提取公共路由）
3. 修改 `apps/corekg/cmd/init.go`：添加 `webtype.InitDB`（website/models 的 InitDB）到 DoInitModels

**验证：** `make local APP=keapp` 和 `make local APP=corekg`

**Commit:** `git add -A && git commit -m "refactor(keapp): update route aggregation and InitDB"`

---

### Task 5: 精简 cmd/ + 清理 + 重新生成 swagger

**操作：**
1. 精简 `apps/keapp/cmd/main.go`：移除不必要的初始化（如果 Task 9 加了多余内容）
2. 清理已移动文件的旧位置（确认 git mv 已处理）
3. 重新生成 swagger：`make generate-docs APP=keapp`
4. 最终验证：`make local APP=keapp && make local APP=corekg`

**Commit:** `git add -A && git commit -m "refactor(keapp): cleanup and regenerate docs"`

---

## 验证清单

- [ ] `make local APP=keapp` 构建通过
- [ ] `make local APP=corekg` 构建通过
- [ ] `go vet ./apps/keapp/...` 无错误
- [ ] 目录结构符合设计方案
- [ ] 各模块 InitDB 独立
- [ ] apis/apis.go 正确聚合所有模块路由
