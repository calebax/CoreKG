# workflow 聚合进 corekg —— 设计文档（配置开关 + 同进程双 Server）

> 日期：2026-08-18
> 状态：**已实施（配置开关 + 同进程双 Server）** — 设计评审通过，核心代码已落地并通过编译。
> 决策前提：采用「方案 A（同进程双 Server）」聚合，并通过**配置开关**决定是否启用；
> §11 待确认项按「最小成本」原则已定夺。

---

## 1. 背景与目标

`apps/corekg/` 是 Go 单体仓库的**聚合单体**：其 `Routers()` 挂载 `kecore + kechat + account + keapi + ketask + kesearch + keapp`（均基于内部 **yg-go** 框架，注册到 `*server.Router`）。此前已盘点：`apps/workflow` 是仍独立部署、尚未聚合的应用。

**目标**：将 `apps/workflow` 纳入 corekg 单进程统一部署；且借助**配置开关**，允许在「启用 workflow」与「不启用 workflow」两种部署形态间平滑切换，**默认不启用**，开关关闭时 corekg 行为与现状完全一致。

---

## 2. 关键结论：workflow 与 corekg 聚合架构不兼容（必须先说清楚）

直接"新增一行 `workflow.Routers(eng)`"**不成立**：

| 维度 | corekg 已聚合应用（kecore/kechat/keapi/keapp…） | apps/workflow |
|------|------|------|
| Web 框架 | 内部 **yg-go**：`github.com/ygpkg/yg-go/apis/runtime/server` → `*server.Router`（gin 内核） | **CloudWeGo Hertz**：`github.com/cloudwego/hertz/pkg/app/server` → `server.Hertz` |
| 路由注册 | action 字符串（`eng.P("app.Action", h)`）由框架映射 URL | IDL 生成的原生路径分组（`/api/**`、`/v1/**`、`/v3/**`、`/admin/**`、`/` 静态页 + NoRoute） |
| 入口契约 | `app.go` 实现 `Routers(*server.Router)` / `Migrates(*gorm.DB)` / `RunJob` | 无 `app.go`；`cmd/main.go` 自建 Hertz server + 中间件链 |
| 配置 | yg-go `CoreConfig` + 各应用 yaml 段 | 自有 `conf/config.go`（`WorkflowConfig`）+ 全局 `conf.SetAppConfig/GetAppConfig` |
| 启动 | corekg `cmd/main.go` 统一初始化 DB/ES/Redis 后 `svr.Run(l)` | 独立 `yygudb.InitYyguDB()` + `application.Init(ctx)` + `appinfra.Init` |

**两个 HTTP 内核不共享路由注册对象**，因此在 corekg 进程内无法把 workflow 路由塞进 yg-go 的 `*server.Router`。聚合的正确形态是：**corekg 进程内同时跑两个 HTTP server — yg-go 主端口负责原有应用，Hertz 独立端口负责 workflow**，通过网关/代理对外统一。

---

## 3. 目标架构

```
corekg 进程（单镜像 / 单进程）
 ├── yg-go 内核 server（现有，主端口 :8080 等）
 │    挂载 kecore / kechat / keapi / keapp / ketask / kesearch / account
 │    行为与现状完全一致
 └── Hertz 内核 server（可选，端口由配置指定，如 :8899）
 │    承载 workflow 全部路由（/api、/v1、/v3、/admin、/ 静态）
 │    仅当  config.workflow.enabled = true 才拉起
 └── 对外：yg-go 主端口 + 可选 Hertz 端口，经网关/CDN 按路径路由分发
```

**工程价值**：默认 `enabled: false` 时，corekg 不 import workflow、不建第二个 server、不改原启动流程 —— **对现有部署零影响、零回归**。开启后只是多拉一个独立端口的 Hertz server。

---

## 4. 配置开关设计

### 4.1 配置文件（`apps/corekg/conf/<env>/config.yaml`）

在 corekg 的 yaml 中新增 `workflow:` 段（与 `main:`/`redis:` 平级）：

```yaml
workflow:
  enabled: false            # 总开关：是否在 corekg 进程内拉启 workflow Hertz server
  required: false           # 开启后启动失败的处理：false=仅告警不拖垮 corekg；true=启动失败则 corekg 整体退出
  http_addr: ":8899"        # workflow Hertz server 独立监听端口
  max_request_body_size: 1073741824

  # ---- workflow 业务运行所需的环境依赖（enabled=true 时才被读取）----
  log_level: debug
  server_host: "http://localhost:8899"
  admin_uins: ""
  ssl:
    enabled: false
    cert_file: ""
    key_file: ""
  redis:
    addr: "${REDIS_ADDR}"
    password: "${REDIS_PASSWORD}"
    db: 0
  elasticsearch:
    addr: "${ES_ADDR}"
    username: "${ES_USERNAME}"
    password: "${ES_PASSWORD}"
    version: v8
    number_of_shards: "1"
    number_of_replicas: "1"
  storage:
    type: minio
    # minio / tos / s3 …（同 workflow 现 config.yaml）
  mq:
    type: nats
    # nats …（同 workflow 现 config.yaml）
  upload:
    component_type: storage
    # imagex …
```

### 4.2 配置项语义

| 字段 | 默认 | 说明 |
|------|------|------|
| `workflow.enabled` | `false` | 总开关。`false` 时 corekg 完全不触达 workflow 包 |
| `workflow.required` | `false` | `true` 时若 workflow 初始化失败，corekg 整体 `Fatal` 退出；`false` 时仅 `Error` 并继续 |
| `workflow.http_addr` | `:8899` | Hertz server 监听地址，避免与 yg-go 主端口冲突 |
| `workflow.*`（其余） | — | workflow 业务依赖（Redis/ES/MQ/OSS/模型/插件…），`enabled=false` 时忽略 |

### 4.3 配置读取要点（风险点）

- yg-go 的 `config.LoadCoreConfig` 解析 `CoreConfig`，**可能丢弃未知的 `workflow:` 段** —— 因此 corekg 需**用 workflow 自己的 `conf.AppConfig` loader 对同一 configFile 做二次独立解析**（`ygpkg/yg-go/config.LoadYamlLocalFile(configFile, &wfAppCfg)`）。
- `main:` 段冲突：workflow 现 yaml 的 `main.app: workflow` 与 corekg 的 `main.app: corekg` 冲突。二次解析时**只取 workflow 需要的连接字段**（`database_conns.core` / `opencoze` 及 `http_addr`），不得覆盖 corekg 的 `main:`。
- 合并后需保证 `database_conns` 里 workflow 的 `core`/`opencoze` 连接名与 corekg 已有的 `account/chat/knownow/llm` 等**互不冲突**。

---

## 5. workflow 侧改造 —— 抽取可复用启动（保持独立二进制可用）

现状 `apps/workflow/cmd/main.go` 把「配置读取 → `yygudb.InitYyguDB` → `application.Init(ctx)` → 建 Hertz server + 中间件 → `router.GeneratedRegister` → `s.Spin()`」全部揉在 main 里。抽出为可复用入口，**独立 `make run APP=workflow` 仍可用**。

### 5.1 新增 `apps/workflow/startup` 包

```go
package startup

import (
    "context"
    "github.com/insmtx/corekg/apps/workflow/conf"
)

// Start 供 corekg 与 workflow 自身复用：
//   1. 注入已解析好的 workflow 配置
//   2. 初始化 ygudb / application
//   3. 建 Hertz server + 挂中间件 + 注册路由
//   4. 用 goroutine 阻塞式 Spin，返回后调用方可继续
func Start(ctx context.Context, appCfg *conf.AppConfig) error {
    conf.SetAppConfig(appCfg)
    if err := yygudb.InitYyguDB(); err != nil {
        return err
    }
    if err := application.Init(ctx); err != nil {
        return err
    }
    buildAndRunServer(appCfg) // 内部 go srv.Spin()
    return nil
}
```

**关键**：`s.Spin()` 必须放在 **goroutine** 里（`go srv.Spin()`），否则会阻塞 corekg 后续的 `svr.Run(l)`。原 `cmd/main.go` 顶层 `mainRun` 结束后程序依赖 `Spin` 阻塞保住存活；改造后需用别的方式阻塞（如 `select{}` / `lifecycle`），确保独立二进制不提前退出。

### 5.2 新增 `apps/workflow/app.go`（标准契约占位）

给 corekg 一个统一 import 入口，满足脚本/检查对标准应用符号的预期：

```go
package workflow

import "gorm.io/gorm"

func Migrates(db *gorm.DB) error { return nil }
func RunJob() error               { return nil }
```

> `Routers(*server.Router)` 不提供 —— 因为 Hertz 路由无法挂到 yg-go `*server.Router`。corekg 不再期望 workflow 走 `Routers` 契约，而是**显式调用 `workflow.Start(ctx, cfg)`**。

---

## 6. corekg 侧改造 —— 条件启动编排

改造点集中在 `apps/corekg/cmd/main.go` 主流程，**插入在 `kecore.RunJob(ctx)` 之后、`svr.Run(l)`（阻塞）之前**：

```go
// 决定是否拉启 workflow Hertz server
if wfEnabled, wfCfg := loadWorkflowConfig(configFile); wfEnabled {
    if err := workflow.Start(ctx, wfCfg); err != nil {
        if cfg.Workflow.Required {
            logs.FatalContextf(cmd.Context(), "[main] start workflow server failed (required): %v", err)
            return
        }
        logs.ErrorContextf(cmd.Context(), "[main] start workflow server failed (continue): %v", err)
    }
}

// …… 原有 svr.Run(l) 阻塞主端口逻辑保持不变
```

**顺序与阻塞约束**：
- `workflow.Start` **必须先于** `svr.Run(l)` 调用，否则其 `go srv.Spin()` 因 `svr.Run` 阻塞主协程，**永远没有机会执行**。
- `workflow.Start` 内部自身 `application.Init` 是同步的（会做 DB/Redis/ES/MQ 依赖校验），只有最后的 `s.Spin()` 是异步 goroutine —— 满足"先同步就绪、再异步服务"。

---

## 7. 灰度 / 故障降级语义

由 `workflow.enabled` + `workflow.required` 组合给出三档：

| enabled | required | 语义 |
|---------|----------|------|
| `false` | — | 完全不拉起 workflow，corekg 行为与现状一致（**默认部署**） |
| `true`  | `false` | 拉起 workflow，但启动失败仅记日志，corekg 主服务照常运行（**推荐灰度**） |
| `true`  | `true`  | 拉起 workflow，启动失败则 corekg 整体退出（强一致部署） |

---

## 8. 构建 / 打包 / 健康检查

- **镜像资源**：corekg 镜像（Dockerfile）需额外 COPY workflow 的静态与配置资源：
  - `apps/workflow/resources/static/**`（前端页面、favicon）
  - `apps/workflow/resources/conf/**`（含 `admin/index.html`）
  - `apps/workflow/conf/plugin/**`、`apps/workflow/conf/model/**`、`apps/workflow/conf/prompt/**`
- **工作目录 / 静态路径**：workflow `staticFileRegister` 基于 `os.Getwd()` 定位 `resources/static`，需保证 corekg 进程工作目录/资源根路径与镜像内一致（或改为显式绝对路径）。
- **健康检查**：yg-go 主端口探活逻辑不变；新增 Hertz 端口需独立就绪探测。若经网关转发，网关侧对 `:8899` 加健康检查即可。
- **启动就绪**：确保 `application.Init`（含依赖校验）通过后、`go srv.Spin()` 已启动，再对 workflow 端口打就绪标记。

---

## 9. 改动文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `apps/workflow/startup/` | **新增** | 抽取 `Start(ctx, cfg)`（异步）与 `Run(ctx, cfg)`（阻塞）两入口，`Spin` 走 goroutine |
| `apps/workflow/cmd/main.go` | 修改 | 复用 `startup.Run`，独立二进制不回归 |
| `apps/workflow/app.go` | **新增** | 标准契约占位（`Migrates`/`RunJob`） |
| `apps/workflow/conf/config.go` | 修改 | `WorkflowConfig` 新增 `Enabled`/`Required`/`HttpAddr` 字段 |
| `apps/corekg/cmd/main.go` | 修改 | `kecore.RunJob` 与 `svr.Run` 之间插条件启动编排 |
| `apps/corekg/cmd/init.go`（如需要） | 修改 | 增加 `loadWorkflowConfig` 二次解析 helper |
| `apps/corekg/conf/<env>/config.yaml` | 修改 | 新增 `workflow:` 段，`enabled` 默认 `false` |
| corekg `Dockerfile` / 构建 | 修改 | 打包 workflow 静态与配置资源 |
| 网关 / CI 健康检查 | 可选 | `:8899` 独立探活 |

---

## 10. 实施状态

- [x] 1. `apps/workflow/startup` 抽取（`Start` 异步 + `Run` 阻塞两种入口）+ `cmd/main.go` 复用 —— **已完成**
- [x] 2. `apps/workflow/app.go` 契约占位（`Migrates`/`RunJob`）—— **已完成**
- [x] 3. corekg `conf` 新增 `workflow:` 段（默认 `enabled: false`）+ `loadWorkflowConfig` 二次解析 —— **已完成**
- [x] 4. corekg `cmd/main.go` 条件启动编排 `maybeStartWorkflow` —— **已完成，编译通过**
- [ ] 5. 运行时验证：`make run APP=workflow` 与聚合 `make run APP=corekg`（依赖本地 DB/Redis/ES，**沙箱环境无法联调**）—— **待部署环境验证**
- [ ] 6. 镜像资源打包（Dockerfile COPY workflow 静态/配置）+ `enabled: false` 回归验证 —— **待部署环境验证**

> **验证说明**：`apps/workflow` 与合并后的 `apps/corekg` 二进制均已编译通过；workflow 独立二进制实测可启动并完成配置解析（仅因沙箱无 MySQL 而止于连接阶段，属预期）。依赖真实基础设施的端到端联调需在本机/部署环境完成。

---

## 11. 待确认 / 已定夺（最小成本原则）

1. **workflow 对外 URL 形态** → 已定夺：**独立端口 `:8899` + 网关按路径转发**（不改路由、不迁移框架，成本最低；不做方案 B 的端口合并）。
2. **`appinfra.Init` 连接名/组件依赖** → 已定夺：**不预先深挖**；复用 workflow 现有 `AppConfig` 结构解析，配置合并时直接照抄 workflow 原 `config.yaml` 的 `workflow:` 段。若联调暴露出缺 `opencoze` 等连接，再在 `main.database_conns` 补充。
3. **灰度降级 `workflow.required=false`** → 已定夺：**保留**（默认 `false`，零成本）。`required: true` 时启动失败则 corekg 整体退出。

> 备注（联调前需注意）：workflow 的 `InitYyguDB` 要求 `main.database_conns.core`；其 `application.Init` 经 `appinfra.Init` 可能还需 `opencoze` 等连接名。corekg 现有 `main.database_conns` 含 `core` 但**不含 `opencoze`**，若启动 workflow 需在 corekg 配置中补该连接。
