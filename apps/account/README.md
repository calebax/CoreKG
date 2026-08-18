# Account 服务

## 服务定位

Account 是 CoreKG 平台的**统一身份与组织管理服务**，解决"谁在用系统、属于哪个组织、能做什么"三个核心问题。

它是整个平台所有业务服务的鉴权基础——kecore、kechat、kesearch 等服务均依赖 account 提供的登录态和身份信息。该服务既可独立部署，也作为子模块被 `apps/corekg` 聚合单体挂载。

## 核心业务域

| 业务域 | 解决什么问题 | 关键能力 |
|--------|-------------|---------|
| 用户身份 | 平台用户的注册、登录、找回 | 邮箱密码登录、第三方 OAuth 登录、手机号验证码、忘记密码、UIN 统一身份标识 |
| 公司/租户 | 多租户隔离与公司认证 | 创建公司、公司实名认证、邀请码绑定公司、公司信息/Logo 管理 |
| 员工与组织 | 公司内部的人员管理与组织架构 | 员工 CRUD、部门树管理、员工归属部门、主部门设置、拖拽排序 |
| 权限控制 | 谁能做什么 | 系统角色（管理员/普通）、职位-权限 RBAC 模型、API 级别鉴权 |
| API Key | 程序化调用认证 | 用户 API Key、Agent API Key 的生命周期管理（创建/禁用/删除） |
| 外部平台绑定 | 连接微信等第三方账号 | OAuth2 Connect 流程、绑定/解绑、微信公众号消息处理 |
| 实名认证 | 个人/企业合规认证 | 个人实名认证提交与审核、公司认证提交与审核 |
| 站点配置 | 平台全局展示信息 | 网站名称/Logo、全局信息获取 |
| 私有化适配 | 私有部署场景的差异处理 | 独立登录入口、默认密码策略、管理员直接创建员工、密码修改提醒 |

## 核心业务概念

- **User** — 自然人，拥有手机号/邮箱/密码
- **UIN（User Identification）** — 用户在某个租户下的身份标识，同一用户在不同公司有不同 UIN
- **Employee** — UIN 在公司内的员工身份，关联系统角色和职位
- **Company** — 租户/公司实体，是数据隔离的基本单位
- **Department** — 公司下的组织单元，支持树形层级和排序
- **Position / Privilege** — RBAC 权限模型，职位关联一组 API 操作权限
- **API Key** — 程序化访问凭证，绑定到 UIN + Company
- **Connector** — 外部 OAuth2 平台（微信等）的绑定关系

## 关键业务流程

### 注册 → 创建公司 → 邀请员工

用户注册 → 创建/认证公司 → 通过邀请码或管理员添加员工 → 员工获得 UIN 和部门归属

### 登录鉴权链路

前端请求 → `InjectLoginStatus` 解析 JWT/API Key → 填充 UIN/CompanyID/EmployeeID 到上下文 → 业务 handler 使用

### 权限校验链路

- `RequireSysAdminRole` — 管理员才能操作
- `RequireOpPrivilege` — 按 API 路径检查 RBAC 权限
- `RequireRefreshToken` — 敏感操作需刷新令牌验证

### 私有化 vs SaaS 分支

密码策略、员工创建方式、Logo URL 生成等均根据 `version.DeployMode()` 走不同逻辑。

## API 路由分组

路由定义在 `internal/apis/apis.go`，按业务域组织：

| 分组 | 代表接口 | 认证要求 |
|------|---------|---------|
| 认证登录 | LoginByEmail, LoginByPassword, LoginThird, RegisterThird, ForgotPassword | 无需登录 |
| Forward Auth | Auth | 无需登录（供网关调用） |
| 用户中心 | Profile, UpdateUserInfo, UpdateAccountPassword, BindUserWechat | 需登录 |
| 实名认证 | PersonAuth, ReviewPersonAuth | 需登录 |
| 公司管理 | CreateCompany, CompanyAuth, BindCompany, GetBindCompanyKey | 需登录 |
| 员工管理 | ListEmployee, UpdateEmployee, DeleteEmployee | 需员工身份 + 运营权限/管理员 |
| 职位权限 | ListPosition, CreatePosition, ModifyPositionPrivilege, ListPrivilege | 需员工身份 + 运营权限 |
| API Key | ListAPIKey, CreateAPIKey, DeleteAPIKey | 需登录 |
| Agent API Key | CreateAgentApiKey, ListAgentAPIKey, SetAgentApiKeyStatus | 需登录 |
| 组织架构 | CreateDepartment, DeleteDepartment, MoveDepartment, GetDepartmentTree | 需登录 + 管理员 |
| 部门员工 | CreateDepartmentEmployee, EditDepartmentEmployee | 需登录 + 管理员 + 配额检查 |
| OAuth2 绑定 | PreConnect, Connect/:provider, Connect/callback, Bindings, Unbind | 需登录 |
| 站点/全局 | GetGlobalInfo, GetLoginSetting, UpdateWebsiteInfo, UploadWebSiteLogo | 视接口而定 |
| 私有化专属 | LoginByPasswordPrivate, CreateEmployee, EditEmployee, DeleteEmployeePrivate | 需登录 + 管理员 |
| 密码变更提醒 | ChangePasswordNotice, ChangeDefaultPassword | 需 RefreshToken |

## 代码架构

```
apps/account/
├── cmd/                  # cobra 入口 + DB/Redis/Provider 初始化
│   ├── main.go           # 主入口，HTTP 服务启动
│   ├── init.go           # 数据库、Redis、OAuth2 Provider 初始化
│   └── reset_password.go # CLI 子命令：批量重置空密码用户
├── app.go                # 暴露 Routers/Migrates/RunJob，供独立部署和 corekg 聚合单体复用
├── accountmds/           # 中间件
│   ├── loginstatus.go    # 登录态注入（JWT / API Key）
│   ├── auth_sysadmin.go  # 系统管理员鉴权
│   ├── auth_opprivilege.go # RBAC 运营权限鉴权
│   ├── auth_api.go       # API Key 鉴权
│   ├── auth_encrypt.go   # 密码 MD5 解密
│   ├── auth_verify.go    # 验证码校验
│   ├── auth_agent_chat.go # Agent 聊天鉴权
│   └── quota.go          # 员工配额检查
├── internal/
│   ├── apis/             # 路由注册 + handler（按业务域拆文件）
│   │   ├── apis.go       # 统一路由注册入口
│   │   ├── login.go      # 登录相关
│   │   ├── register.go   # 注册相关
│   │   ├── profile.go    # 用户资料
│   │   ├── employee_manage.go / employee_profile.go / employee_perms.go / employee_biz.go
│   │   ├── auth_company.go / auth_company_biz.go  # 公司认证
│   │   ├── auth_user.go / auth_user_biz.go / user.go / user_biz.go
│   │   ├── organize.go   # 组织架构
│   │   ├── api_key.go / api_key_biz.go  # API Key
│   │   ├── apikey/       # Agent API Key
│   │   ├── connectoes.go / connectoes_biz.go  # OAuth2 绑定
│   │   ├── wechat_mp.go  # 微信公众号
│   │   ├── login_setting.go / login_setting_biz.go
│   │   ├── global.go / website.go
│   │   ├── upload_cos_image.go / upload_biz.go
│   │   ├── keauth.go / keauth_biz.go / obo_token.go
│   │   └── privite/      # 私有化专属接口
│   ├── dto/              # 请求/响应 DTO（按业务域分子包）
│   ├── docs/             # Swagger 自动生成（勿手动编辑）
│   └── migrate/          # 数据库迁移
├── services/             # 业务逻辑层
│   ├── svcuser/          # 用户相关（忘记密码等）
│   ├── svccompany/       # 公司创建与认证
│   ├── svcorganize/      # 组织架构（部门、员工、公司信息）
│   ├── svcglobal/        # 全局信息
│   └── svcwebsite/       # 站点配置
├── models/               # GORM 数据模型
│   ├── user/             # 用户、JWT、短信验证
│   ├── employee/         # 员工、职位、权限、JWT
│   ├── company/          # 公司、邀请码
│   ├── account/          # 部门、用户标识、API Key
│   ├── accounttype/      # 类型定义与公共模型
│   ├── apikey/           # API Key 业务逻辑
│   ├── connectors/       # 外部平台绑定
│   ├── perm/             # 权限
│   └── wechatmp/         # 微信公众号模板消息
├── conf/                 # 环境配置
│   └── test/config.yaml
└── script/               # Dockerfile + 迁移 SQL
```

## 技术要点

- **框架**：yg-go（内部框架，action-based 路由），底层 gin
- **数据库**：MySQL（GORM，多库连接 via `dbutil.Account()`），Redis（缓存 + 会话）
- **认证**：JWT（UIN 体系），API Key（Bearer），OAuth2 Connector
- **密码**：前端 MD5 加密传输 → `accountmds.DecryptMD` 中间件自动解密 → 后端 bcrypt 存储
- **权限模型**：SysRole（系统角色：sysadmin/普通）+ Position/Privilege（RBAC 职位-API 权限映射）
- **部署模式**：SaaS / 私有化，通过 `version.DeployMode()` 区分逻辑分支
- **i18n**：错误消息使用 i18n key，对应 `resource/locales/` 翻译文件
- **Swagger**：`make generate-docs APP=account` 自动生成到 `internal/docs/`，勿手动编辑

## 本地开发

```bash
# 构建
make local APP=account

# 构建并运行（使用 conf/test/config.yaml）
make run APP=account ENV=test

# 生成 Swagger 文档
make generate-docs APP=account

# 运行测试（需要真实 MySQL/Redis 连接）
go test ./apps/account/...
```

### CLI 子命令

| 命令 | 说明 |
|------|------|
| `account` | 启动 HTTP 服务 |
| `account reset-password` | 批量重置密码为空的用户（密码 = 邮箱 + "ABC"） |
| `account test-forward-auth` | 启动 :30010 端口调试 Forward Auth 请求头 |
| `account version` | 打印构建版本信息 |

## 与其他服务的关系

- **被 `apps/corekg/` 聚合单体挂载**：其 `Routers` 同时服务于独立部署和聚合模式
- **依赖 `pkgs/global`**：上下文 Key、常量、错误码
- **依赖 `pkgs/connectors`**：OAuth2 Provider 初始化与管理
- **依赖 `pkgs/utils/dbutil`**：多数据库连接池
- **依赖 `apps/admin/models/login_setting`**：登录配置（域名 → Issuer 映射）
- **调用 `apps/kecore/services/svccoze.SpaceSync`**：员工/公司信息变更后同步知识库空间
