# Admin 服务

## 服务定位

Admin 是 CoreKG 平台的**内部运营管理后台**，面向内部运营人员提供平台级管理能力。它是整个平台的控制面板，负责管理员工、角色权限、公司/租户、终端用户、License、Prompt 模板、系统公告、登录配置、运营设置和数据看板。

## 核心业务域

| 业务域 | 解决什么问题 | 关键能力 |
|--------|-------------|---------|
| 员工管理 | 内部运营账号管理 | 员工 CRUD、密码重置、微信绑定、头像上传 |
| RBAC 权限 | 精细化操作权限控制 | 职位管理、API 级权限 CRUD、员工-职位绑定 |
| 公司/租户管理 | 多租户运营 | 创建/修改公司、管理公司员工与角色 |
| 终端用户管理 | 平台用户账号运维 | 用户 CRUD、密码重置 |
| License 管理 | 私有化部署授权 | RSA 签名生成、分发、下载、应用 License |
| Prompt 模板 | LLM 提示词管理 | 版本化 Prompt CRUD、变量校验、预览渲染、版本切换 |
| 系统公告 | 版本发布通知 | 公告 CRUD |
| 登录配置 | 登录页定制 | 登录页配置管理 |
| 运营设置 | 灵活配置 | Key-Value 运营参数管理 |
| 数据看板 | 平台数据概览 | 跨库聚合统计（知识库/文档/对话/问答/Agent/图谱/文章/用户） |
| HTTP 代理 | 内部请求转发 | 白名单反向代理，支持 Auth Header 注入 |
| LKX 线索收集 | 销售线索录入 | 短信验证 + 企业微信 Webhook 通知 |

## 核心业务概念

- **Employee（运营员工）** — 内部运维人员，有独立的用户体系（区别于终端用户），通过用户名/密码登录
- **Position（职位）** — RBAC 角色，关联一组 API 权限
- **Privilege（权限）** — API 级别的操作权限，按 URI 路径 + action 定义，支持层级
- **License（许可证）** — RSA 签名的私有化部署授权文件，绑定机器 UID 和有效期
- **Prompt（提示词模板）** — LLM 提示词，支持多版本管理、变量 Key 校验、渲染预览
- **LoginSetting（登录配置）** — 可定制的登录页参数
- **AdminAnnouncement（系统公告）** — 版本更新说明

## 关键业务流程

### 员工权限校验链路

请求 → `InjectEmployeeLoginStatus` 解析 JWT → `RequireOpPrivilege` 按 API 路径检查 RBAC 权限 → handler 执行。系统管理员（sys_admin）跳过所有权限检查。

### License 生成流程

生成 RSA 密钥对 → 构造 License JSON 元数据 → PKCS1v15+SHA256 签名 → Base64 编码 → 存入数据库（原始字符串 + 密钥）→ 下载分发给客户 → 客户通过 `ApplyLicense` 注册到部署实例。

### 跨库数据看板

同时查询 Account（用户）、Knownow（知识库/文档/图谱）、Chat（会话/Agent）、Core（任务）、Elasticsearch（问答统计）等多个数据源，通过 `StatManager` 模式批量执行统计函数。

## API 路由分组

路由定义在 `internal/apis/apis.go`，所有路由使用 `admin.` 或 `lkxadmin.` action 前缀。

| 分组 | 代表接口 | 认证要求 |
|------|---------|---------|
| 登录 | LoginByPassword, LoginThird | 无需认证 |
| 员工自助 | ModifyMyUserInfo, ChangeMyWechat, GetMyAction | 需员工身份 |
| 员工管理 | ListEmployee, CreateEmployee, ModifyEmployeeInfo, DeleteEmployee | 员工 + 运营权限 |
| 职位/权限 | ListPosition, CreatePosition, ModifyPositionPrivilege, ListPrivilege | 员工 + 运营权限 |
| 运营设置 | ListSetting, CreateSetting, UpdateSetting | 员工 + 运营权限 |
| 登录配置 | ListLoginSetting, CreateLoginSetting, UpdateLoginSetting | 员工 + 运营权限 / 公开 |
| License | GenerateLicense, ListLicense, DownloadLicense, DistributeLicense | 员工 / API Key |
| 公司管理 | CreateCompany, ListCompany, CreateCompanyEmployee | 员工 + 运营权限 |
| 用户管理 | CreateUser, ListUser, ModifyUser, ModifyUserPassword | 员工 + 运营权限 |
| 数据看板 | GetDashboardData, GetDashboardOverview | 员工身份 |
| 公告 | ListAnnouncement, CreateAnnouncement, ModifyAnnouncement | 员工 + 运营权限 |
| Prompt 管理 | CreatePrompt, AddPromptVersion, SwitchPromptVersion, RenderPromptPreview | 员工 + 运营权限 |
| HTTP 代理 | ProxyHTTP | 员工 + 运营权限 |
| LKX 线索 | lkxadmin.SendVerifyCode, lkxadmin.VerifyCodeAndSave | 无需认证 |

## 代码架构

```
apps/admin/
├── cmd/main.go              # Cobra 入口，启动 HTTP 服务
├── app.go                   # Routers/Migrates/RunJob 标准三件套
├── adminmds/                # 中间件
│   └── auth_admin.go        # 员工 JWT 认证 + API 权限检查
├── internal/
│   ├── apis/                # Handler 层
│   │   ├── apis.go          # 统一路由注册
│   │   ├── employee_*.go    # 员工管理
│   │   ├── companyctl/      # 公司管理
│   │   ├── userctl/         # 终端用户管理
│   │   ├── loginctl/        # 登录
│   │   ├── license/         # License 管理
│   │   ├── promptctl/       # Prompt 模板管理
│   │   ├── dashboard/       # 数据看板
│   │   ├── lkxctl/          # LKX 线索收集
│   │   └── proxy_http.go   # HTTP 反向代理
│   ├── dto/                 # 请求/响应 DTO
│   └── docs/                # Swagger 自动生成（勿手动编辑）
├── models/                  # 数据访问层
│   ├── admintype/           # Employee/Position/Privilege/License 等类型定义
│   ├── employee/            # 员工查询与 RBAC
│   ├── company/             # 公司 CRUD
│   ├── license/             # License 生成
│   ├── login_setting/       # 登录配置
│   ├── prompt/              # Prompt 模型
│   ├── dashboard/           # 看板数据查询（跨库）
│   ├── lkx/                 # LKX 客户信息
│   └── user/                # 终端用户 CRUD
├── services/                # 业务逻辑层
│   ├── svcdashboard/        # 看板聚合统计（StatManager）
│   ├── svcprompt/           # Prompt 版本管理 + 渲染
│   ├── svcannouncement/     # 公告服务
│   ├── svcorder/            # 订单服务
│   └── svrlkx/              # LKX 短信验证 + 企微通知
└── conf/test/               # 环境配置
```

## 技术要点

- **多数据库**：同时连接 Account、Core、Knownow、Chat、Sale 五个数据库
- **RBAC**：API 级别的权限校验，通过 Position-Privilege 多对多映射实现
- **License 加密**：RSA 非对称签名（PKCS1v15+SHA256），绑定 K8s 集群 UID 或物理机
- **HTTP 代理**：白名单机制存储在 admin settings 中，支持 per-base Auth Header 注入
- **Prompt 版本管理**：支持多版本、切换激活版本、变量 Key 校验、渲染预览
- **WeCom 集成**：LKX 线索提交后通过企业微信机器人 Webhook 通知

## 本地开发

```bash
make local APP=admin
make run APP=admin ENV=test
make generate-docs APP=admin
```

## 与其他服务的关系

- 依赖 `apps/account/accountmds` — API Key 鉴权
- 依赖 `apps/account/models/accounttype` — 公司/部门/用户标识类型
- 依赖 `apps/corekg/models/license` — License 加密操作
- 依赖 `apps/kecore/models` — Forest/Graph/Task 类型（看板统计）
- 依赖 `apps/kechat/models/chattype` — ChatSession/ChatAgent（看板统计）
- 依赖 `apps/kesearch/models/essearch` — ES 客户端（问答统计）
- 依赖 `apps/kesale/models/sale` — Sale 数据库初始化
- 依赖 `pkgs/utils/dbutil` — 多库连接访问器
- 依赖 `pkgs/utils/notify/sms` — SMS 验证码发送
