# KEInit

## 服务定位

KEInit 是 CoreKG 平台的**基础设施初始化与数据迁移工具**（CLI）。它不是常驻服务，而是在部署时运行，负责初始化整个平台的底层依赖：MySQL 表结构、Elasticsearch 索引映射、MinIO 存储桶、系统设置、AI 聊天模型配置和 API Key。可以理解为平台的"安装向导"或"bootstrap 工具"。

## 核心能力

| 能力 | 说明 |
|------|------|
| MySQL 表初始化 | 连接多库，执行所有模块的 InitDB（account/kecore/kechat/task/wecom/settings 等） |
| ES 索引创建 | 连接 ES，执行 DSL 文件创建索引映射 |
| MinIO Bucket 创建 | 创建对象存储桶（带重试，最多10次，间隔15秒） |
| 聊天模型配置 | 初始化/更新 LLM 模型配置 |
| API Key 配置 | 初始化/更新系统 API Key |
| 系统设置 | 更新系统配置项 |
| 数据迁移 | 执行自定义迁移脚本（Coze 模型同步、部门迁移、图谱节点迁移等） |
| 集群 ID 查询 | 返回 K8s 集群 UID（用于 License 绑定） |

## CLI 子命令

| 命令 | 说明 |
|------|------|
| `keinit`（无参数） | 执行完整初始化流程：initDB -> updateChatModel -> updateApiKey -> migrate -> initES -> createBucket |
| `keinit init-es` | 仅初始化 ES 索引 |
| `keinit migrator` | 仅执行数据迁移脚本 |
| `keinit update-setting` | 仅更新系统设置 |
| `keinit update-chatmodel` | 仅更新聊天模型配置 |
| `keinit update-api-key` | 仅更新 API Key |

## HTTP 端点

仅暴露两个轻量健康检查端点：

- `GET /v3/status.GetClusterID` - 返回 K8s 集群 UID
- `GET /v3/status.Ping` - 返回 "pong"

## 代码架构

```
apps/keinit/
├── cmd/
│   ├── main.go          # Cobra 主入口，定义 rootCmd 及子命令
│   ├── init.go          # initDatabase(): 多模块 DB 表结构初始化
│   ├── es.go            # initES(): 连接 ES + 执行 DSL 文件
│   ├── migrator.go      # runMigrator(): 自定义数据迁移
│   ├── mysql.go         # initMysqlEnv 子命令
│   ├── setting.go       # updateSetting 子命令
│   ├── chatmodel.go     # updateChatModel 子命令
│   ├── api_key.go       # updateApiKey 子命令
│   └── apis.go          # GetClusterID / Ping handler
├── migrator/
│   ├── registry.go      # 迁移注册
│   ├── cozemodelsync.go # Coze 模型同步
│   ├── department.go    # 部门数据迁移
│   └── ...              # 其他迁移脚本
├── models/
│   ├── helper/          # ENV 文件读取
│   ├── minio/           # MinIO Bucket 创建
│   └── es/              # ES DSL 执行
└── conf/                # 配置文件
```

## 与其他服务的关系

- 导入多个模块的 `InitDB` 函数：accounttype、chattype、foresttype、task、wecoms、settings
- 依赖 `apps/corekg/models/license` - 获取集群 UID
- 依赖 `pkgs/utils/dbutil` - 多库连接
- 依赖 `pkgs/wecoms` - 企业微信表初始化
