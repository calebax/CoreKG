# CoreKG 本地基础环境搭建指南

本文档说明如何在本地快速搭建 CoreKG 所需的全部基础中间件（MySQL / Elasticsearch / Redis / MinIO / NATS），并初始化数据。所有服务均通过 Docker Compose 编排，使用 Docker Hub 官方 multi-arch 镜像（amd64 / arm64 均可用）。

## 1. 总览

| 中间件 | 镜像 | 容器内部端口 | 宿主机映射端口 | 账号 / 密码 |
|---|---|---|---|---|
| MySQL | mysql:8.4 | 3306 | **3308** | root / `123456`；corekg / `123456` |
| MySQL（opencoze 附加库） | - | - | - | 同一 MySQL |
| Elasticsearch | elasticsearch:8.18.1 | 9200 / 9300 | **9202** / **9302** | elastic / `123456` |
| Redis | redis:7 | 6379 | **6381** | 无密码 |
| MinIO | minio/minio:latest | 9000 / 9001 | **9002** / **9003** | minioadmin / `minio123456` |
| NATS | nats:2 | 4222 | **4225** | 无认证 |

> **端口设计说明**：容器内部保留各服务默认端口（互连时用服务名+内部端口）；宿主机端口统一在默认端口基础上 **+2**，以避免本地可能已安装的 MySQL/Redis/ES/MinIO 占用默认端口导致启动冲突。如某端口仍被占，可自行在 `docker-compose.yml` 中改映射端口，并同步修改对应 `config.yaml`。

> **凭据说明**：所有中间件明文密码统一为 `123456`（本地开发默认值，欢迎直接使用）。生产环境请勿使用本文件与默认密码。

## 2. 快速启动

```bash
# 1) 启动全部基础依赖（本仓库已提供固定默认值的 docker-compose.yml）
docker compose up -d

# 2) 等待 minio-init 完成 bucket 初始化后确认状态
docker compose ps
```

- 首次启动 MySQL 时，`scripts/mysql-docker-init.sh` 会自动额外创建 `opencoze` 数据库（供 kechat / keinit / workflow 使用），并向 `corekg` 用户授权。
- Minio 启动后 `minio-init` 会自动创建 `corekg-bucket`（幂等，重复执行不报错）。

### 各服务健康检查

```bash
# MySQL（本机通过映射端口 3308 访问）
mysql -h127.0.0.1 -P3308 -uroot -p123456 -e "SHOW DATABASES;"
# 应看到 corekg 与 opencoze

# Elasticsearch（注意带 Basic Auth，密码 123456）
curl -u elastic:123456 http://localhost:9202/

# Redis
redis-cli -p 6381 ping

# MinIO（Node 端口 9002；Web 控制台 9003，浏览器打开 http://localhost:9003）
# NATS
```

## 3. 各服务进程如何连接

- **容器间互连**：各 Docker 服务之间用 `服务名 + 容器内部端口`（如 `minio:9000`、`mysql:3306`）。
- **宿主机进程访问**：`make run` 启动的 Go 服务（keinit / corekg / keapi / ketask 等）跑在宿主机上，一律通过 **宿主机映射端口** 连接，因此 `config.yaml` 中连接的地址为 `localhost:3308`、`localhost:9202`、`localhost:6381`、`localhost:9002`、`nats://localhost:4225` 等。

### 连接字符串速查（写入各 config.yaml）

- **MySQL DSN**：`mysql://corekg:123456@localhost:3308/corekg?charset=utf8mb4&parseTime=true&loc=Local`
- **MySQL（opencoze）DSN**：`mysql://corekg:123456@localhost:3308/opencoze?charset=utf8mb4&parseTime=true&loc=Local`
- **Redis**：`addr: localhost:6381`
- **Elasticsearch**：`addresses: [http://localhost:9202]`，`username: elastic`，`password: 123456`
- **MinIO**：`end_point: http://localhost:9002`，`access_key_id: minioadmin`，`secret_access_key: minio123456`
  > ⚠️ **workflow 的 MinIO 连接是例外**：workflow 用 `minio-go` 客户端，其 `endpoint` 必须是**裸 host:port**（`localhost:9002`，不带 `://`），scheme 由 config 的 `storage.upload_http_scheme`（本地应填 `http`）决定。若写成 `http://localhost:9002` 会报 `Endpoint url cannot have fully qualified paths.`。凭证需与 docker-compose 的 `MINIO_ROOT_USER/MINIO_ROOT_PASSWORD`（`minioadmin` / `minio123456`）一致；`storage.bucket` 不存在时 workflow 会自动创建。
- **NATS**：`nats://localhost:4225`

## 4. 初始化一次数据（keinit）

CoreKG 的数据初始化由 `keinit`（CLI / bootstrap 工具）完成，承担：MySQL 建表、SQL 种子数据、ES 索引、MinIO bucket、系统设置、聊天模型与 API Key。完整流程见上文对 `apps/keinit` 的说明。

```bash
# 准备 keinit 配置
cp apps/keinit/conf/test/config.yaml.example apps/keinit/conf/test/config.yaml
cp apps/keinit/conf/test/core_setting.yaml.example apps/keinit/conf/test/core_setting.yaml

# 编译并执行完整初始化（initDB -> chatmodel -> apikey -> migrate -> initES -> bucket）
make local APP=keinit ENV=test
./bundles/keinit -c apps/keinit/conf/test/config.yaml \
    --setting-file apps/keinit/conf/test/core_setting.yaml \
    --env-file apps/keinit/conf/test/kinit.env   # 可选：若需注入 LLM 模型/APIKEY 等变量
```

> `kechat` 等应用中的聊天模型（`chat_model`）与 `chat_agent` 种子数据依赖 `scripts/mysql` 的 SQL 脚本；其中 `@yg_*` 变量来自 `scripts/mysql/variable.sqltpl`（由 `--env-file` 或进程环境变量注入）。本地若暂时没有 LLM API Key，可先在 `variable.sqltpl` 对应 `@yg_llm_*` 填占位值并跳过模型连通性验证。

## 5. 初始管理员账号

初始化完成后，系统内置一个管理员账号，由 `scripts/mysql/v1.0_2__insert_account.sql` 写入：

| 字段 | 值 |
|---|---|
| 登录账号 | `admin@admin.com`（按邮箱登录） |
| 密码 | `admin123456` |
| identify | `admin` |

登录接口 `account.LoginByPassword`：当账号含 `@` 时按邮箱匹配 `user.email`，密码用 bcrypt 校验。

> 如需修改初始密码，请勿改已发布的基线 SQL（已写入 bcrypt 哈希）；应在系统运行后通过账号体系接口/后台修改。

## 6. 前端 / worker / pipeline 配置同步

- **前端**：`frontend/corekg/.env.development.example`、`.env.production.example`（API 地址指向对应后端应用映射端口）。
- **TS worker**：`apps/worker/.env.example`。
- **Python pipeline**：`apps/pipeline/config/*.yaml.example`。
- **workflow 应用**：其 `config.yaml`（`apps/workflow/conf/test/config.yaml`，以及聚合进 `apps/corekg/conf/test/config.yaml(.example)` 的 workflow 配置块）**不再依赖环境变量**，所有连接信息已收敛为与 `docker-compose.yml` 一致的**字面值**，直接运行即可：

  ```bash
  make run APP=workflow ENV=test          # 独立 workflow 服务
  make run APP=corekg  ENV=test           # 聚合服务（进程内拉启 workflow）
  ```

  各字段默认值如下（端口/凭据已按本方案统一偏移 +2、密码 `123456`；改动中间件时直接改这些字面值即可）：

  | 配置字段（`workflow:` 下） | 字面值 | 说明 |
  |---|---|---|
  | `main.database_conns.core` | `mysql://corekg:123456@localhost:3308/corekg` | core 库 DSN |
  | `main.database_conns.opencoze` | `mysql://corekg:123456@localhost:3308/opencoze` | opencoze 库 DSN |
  | `redis.addr` / `password` | `localhost:6381` / 空 | Redis（无密码） |
  | `elasticsearch.addr` / `username` / `password` | `http://localhost:9202` / `elastic` / `123456` | Elasticsearch |
  | `storage.minio.ak` / `sk` / `endpoint` / `region` / `api_host` | `minioadmin` / `minio123456` / `localhost:9002` / `us-east-1` / `localhost:9002` | MinIO 对象存储（`endpoint`/`api_host` 为**裸 host:port**，scheme 取自 `upload_http_scheme`） |
  | `mq.name_server` | `nats://localhost:4225` | NATS 消息总线 |

## 7. 常见问题

- **端口已被占用 / 容器启动失败**：因本方案已将宿主机端口统一 +2，绝大多数冲突已规避；若仍冲突，改 `docker-compose.yml` 端口映射并同步到对应 `config.yaml`。
- **MySQL 改了密码/清库后想重置**：删除数据卷后重新 `docker compose up -d`（首次初始化脚本会重建库与 opencoze）。
- **ES 鉴权失败**：确认使用 `-u elastic:123456`，且 `config.yaml` 中 `username/password` 为 `elastic/123456`。
