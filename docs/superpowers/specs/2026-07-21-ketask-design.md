# ketask: 合并 keparser + keworker

## 概述

将 `apps/keparser`（任务调度 HTTP API）和 `apps/keworker`（任务执行 CLI worker）合并为一个服务 `apps/ketask`。合并后的服务同时提供任务管理 HTTP API 和 7 个 worker 子命令（作为 cobra 子命令）。Worker 继续通过 HTTP 调用任务 API 来消费任务（进程内消费方式不变）。Models 包从 `apps/keparser/models/` 迁移到 `apps/ketask/models/`，并更新所有跨应用 import。`apps/corekg` 聚合单体更新为挂载 `ketask.Routers()` 以替代 `keparser.Routers()`。

## 决策

| 决策 | 选择 | 理由 |
|------|------|------|
| Models 位置 | `apps/ketask/models/` | 用户偏好；命名更清晰 |
| Worker 消费方式 | HTTP 调用（保持现状） | 用户偏好；避免改变已验证的行为 |
| corekg 聚合单体 | 将 `keparser.Routers()` 更新为 `ketask.Routers()` | 用户偏好；保持聚合单体一致性 |
| 配置格式 | 统一为标准 `CoreConfig` + worker 扩展 | keparser 使用 `CoreConfig`，keworker 使用自定义 YAML；合并为兼顾两者的统一配置 |

## 范围

**包含在内：**
- 创建 `apps/ketask/`，整合 keparser + keworker 的功能
- 将 `models/ragtask`、`models/ragtypes`、`models/algofilehandle` 迁移到 `apps/ketask/models/`
- 更新全部 30+ 处跨应用 import 路径
- 更新 `apps/corekg/app.go` 以使用 ketask
- 更新 `apps/corekg/internal/jobs/task_finished_parse_file.go` 的 import
- 更新 CI/CD 配置（`.gitlab-ci.yml`、`.github/workflows/*.yml`）
- 创建统一的 Dockerfile
- 创建统一的测试配置
- 删除 `apps/keparser/` 和 `apps/keworker/`

**不包含在内：**
- 将 worker 消费方式从 HTTP 改为进程内调用
- 修改 worker 业务逻辑
- 更新 `clients/task_worker/`（独立客户端二进制文件，如需要则单独更新）

## 受影响的文件

### 新增文件 (apps/ketask/)
- `app.go` — 基于 keparser 的 `app.go`，包名改为 `ketask`
- `cmd/main.go` — 合并：keparser HTTP 服务器 + keworker cobra 子命令 + 统一配置加载
- `cmd/init.go` — 来自 keparser（DB/Redis/Task 初始化）
- `cmd/flags.go` — 来自 keworker（S3/ES/Nebula/Agent 配置 + `AppConfig` 结构体）
- `cmd/task.go` — 来自 keworker（HTTP GetPendingTask/CallBackTask/DownloadFile）
- `cmd/biz.go` — 来自 keworker（Worker 接口、agent 辅助函数、MapReduce）
- `cmd/ping.go` — 来自 keworker（健康检查，当前已注释）
- `cmd/api_daemon.go` — 来自 keworker（守护进程子进程管理）
- `cmd/worker_pdf_extract.go` — 来自 keworker
- `cmd/worker_split_text_chunk.go` — 来自 keworker
- `cmd/worker_doc_to_pdf.go` — 来自 keworker
- `cmd/worker_description.go` — 来自 keworker
- `cmd/worker_video_extract.go` — 来自 keworker
- `cmd/worker_insert_index.go` — 来自 keworker
- `cmd/worker_copy.go` — 来自 keworker
- `cmd/biz_test.go` — 来自 keworker
- `cmd/worker_doc_to_pdf_test.go` — 来自 keworker
- `internal/apis/apis.go` — 来自 keparser
- `internal/docs/` — 来自 keparser（重新生成）
- `internal/jobs/jobs.go` — 来自 keparser
- `internal/jobs/task_finished_parse_file.go` — 来自 keparser
- `models/ragtask/payload.go` — 从 keparser 迁移
- `models/ragtask/pdf_extractext.go` — 从 keparser 迁移
- `models/ragtypes/chunk_type.go` — 从 keparser 迁移
- `models/ragtypes/chunks.go` — 从 keparser 迁移
- `models/algofilehandle/` — 从 keparser 迁移
- `conf/test/config.yaml` — 统一配置
- `script/Dockerfile` — 统一 Dockerfile

### 修改的文件
- `apps/corekg/app.go:13,44` — import + Routers 调用
- `apps/corekg/internal/jobs/task_finished_parse_file.go:14` — import 路径
- `.gitlab-ci.yml:23,25` — 应用名称列表
- `.github/workflows/deploy_service.yml:11,15` — 应用名称列表
- `.github/workflows/deploy_test.yml:11,15` — 应用名称列表
- `.github/workflows/deploy_service_priv.yml:13` — 应用名称列表
- kecore、kesearch、kechat、keapi 中 30+ 个 Go 文件 — import 路径更新

### 删除的目录
- `apps/keparser/`（整个目录）
- `apps/keworker/`（整个目录）
