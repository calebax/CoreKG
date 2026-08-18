# CoreKG 聚合单体

## 服务定位

CoreKG 是平台的**聚合单体（All-in-One）部署模式**。它将六个子应用打包到一个 HTTP 服务进程中，提供完整的知识库/RAG 平台功能。适用于不需要微服务拆分的部署场景。

## 挂载的子应用

| 子应用 | 职责 |
|--------|------|
| **kecore** | 知识库管理、文档管理、知识图谱、写作空间 |
| **kechat** | AI 对话、RAG 问答、Agent 管理、模型管理 |
| **account** | 用户认证、公司/组织管理、API Key |
| **keapi** | 对外知识库 API + MCP Server |
| **keparser** | 文档解析任务调度 |
| **kesearch** | Elasticsearch 检索与搜索 |
| **wecom** | 企业微信 Webhook/回调（挂载在 /v2/account 和 /v3/account） |

此外，corekg 自身还提供 **License 管理** API（CheckLicense、GetLicenseInfo、RegisterLicense）。

## 核心业务概念

- **License 验证中间件** — RSA 签名的 License 校验，绑定 K8s 集群 UID 或物理机，72小时有效期窗口，每日定时验证
- **FinishedParseFileRoutine** — 后台协程，轮询已完成的文件解析任务并执行后处理
- **DeployMode（部署模式）** — on_premise / openpo / tencent_free，控制功能开关

## 启动初始化序列

```
1.  MySQL 多库连接 + Redis
2.  GORM 插件初始化
3.  任务队列初始化 + GraphFileTask 回调注册
4.  HTTP Router 创建（/v3/ 前缀）+ 可选 LicenseCheck 中间件
5.  文件存储初始化（COS/S3/MinIO）
6.  任务状态检查协程
7.  Elasticsearch 客户端（search + chunk + highlight + history）
8.  NebulaGraph 客户端（非 OpenPO 模式）
9.  外部 Provider Connectors
10. i18n 国际化
11. kecore 定时任务
12. 注册所有子应用路由
13. 启动 HTTP 服务
14. 后台协程（解析任务处理 + License 每日检查）
15. 优雅退出等待
```

## 代码架构

```
apps/corekg/
├── app.go                   # Routers/Migrates/RunJob — 组合所有子应用
├── cmd/
│   ├── main.go              # Cobra 入口，完整初始化链
│   └── init.go              # initDatabase/initTask
├── internal/
│   ├── apis/
│   │   ├── apis.go          # corekg 自身路由（License 端点）
│   │   └── licensectl/      # License API handler
│   ├── jobs/
│   │   ├── jobs.go          # RunRoutines: 后台协程入口
│   │   ├── task_finished_parse_file.go  # 解析完成处理
│   │   └── task_license_verify.go       # License 每日验证（Redis 分布式锁）
│   └── docs/                # Swagger 自动生成
├── mds/
│   └── license.go           # LicenseCheck 中间件（~20 个豁免 Action）
├── models/
│   └── license/
│       ├── crypto.go        # RSA 密钥解析/生成
│       ├── environment.go   # 环境接口（K8s UID/License 文件）
│       └── status.go        # 验证状态枚举
└── conf/test/               # 环境配置
```

## License 验证机制

- 启动时通过 `LicenseCheck` 中间件拦截所有请求
- 约 20 个 Action 豁免（登录、文档、License 注册等）
- 每日定时验证 License 有效性，多副本使用 Redis 分布式锁
- License 无效时所有 API 返回错误
- `tencent_free` 部署模式跳过 License 检查

## 技术要点

- **共享端口**：所有子应用运行在同一 HTTP 端口（默认 :8080）
- **Swagger 合并**：生成单一的合并 API 文档
- **Migrates 为空**：表结构通过 `--migrate-db` 启动参数触发
- **多库**：core、account、knownow、chat + Redis
- **优雅退出**：`lifecycle.Std().WaitExit()` 信号处理

## 本地开发

```bash
make local APP=corekg
make run APP=corekg ENV=test
make generate-docs APP=corekg
```

## 与其他服务的关系

- 直接导入并调用所有子应用的 `Routers()` 方法
- 依赖 `apps/admin/models` — License 检查、日志
- 依赖 `pkgs/apis/wecom` — 企业微信 Webhook 路由
- 依赖 `pkgs/task` — 任务队列系统
- 依赖 `pkgs/connectors` — 外部 Provider 初始化
- 依赖 `version` — 构建信息 + 部署模式检测
