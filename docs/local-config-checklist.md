# 本地配置清单（初始化所需真实值）

> 归档本次初始化需补齐的环境变量/占位符清单，以及本地生效文件中的已填真实值。
> 涉及真实密钥的文件均被 `.gitignore` 忽略，**不会提交入库**；仓库内仅保留 `*.example` 占位模板。

## 三类配置来源

| 文件 | 是否入库 | 内容 |
|---|---|---|
| `apps/*/conf/*/config.yaml` | 否（gitignore） | 各 Go 服务运行配置（DB/Redis/JWT/企微/腾讯云/SMTP 等） |
| `apps/keinit/conf/test/core_setting.yaml` | 否（gitignore） | `core_settings` 表系统设置（Redis/ES/MinIO/模型/PDF 转换等） |
| `apps/keinit/conf/test/kinit.env.local` | 否（gitignore `*.local`） | keinit 初始化真实模型 key/地址（对话/视觉/Embedding） |

仓库内对应模板：`*.config.yaml.example`、`core_setting.yaml.example`、`kinit.env`（占位）。

---

## 一、基础中间件（docker-compose 提供）

按仓库根 `docker-compose.yml` 启动即可，默认值即本地值，通常无需额外提供：

```bash
docker compose up -d
```

| 服务 | 宿主机地址 | 账号 / 密码 |
|---|---|---|
| MySQL | localhost:3308 | `corekg` / `123456`（root 同为 `123456`） |
| Elasticsearch | localhost:9202 | `elastic` / `123456` |
| Redis | localhost:6381 | 无密码 |
| MinIO | localhost:9002 / 9003 | `minioadmin` / `minio123456` |
| NATS | localhost:4225 | 无认证 |

---

## 二、需提供/已填的真实值

### 1. 服务地址（`127.0.0.1`）
`apps/keinit/conf/test/kinit.env.local`：
```
BASE_URL = http://127.0.0.1:30000
BASE_HOST = 127.0.0.1
```

### 2. 对话 LLM
`kinit.env.local`（key 与域名已脱敏，真实值仅在本地 `.local` 文件）：
```
LLM_MODEL = deepseek/deepseek-v4-flash
LLM_MODEL_URL = <your-llm-host>/v3/llm.chat/chat/completions
LLM_MODEL_APIKEY = <your-llm-api-key>
```

### 3. 视觉 / 多模态模型
`kinit.env.local`：
```
LLM_VL_MODEL = qwen2.5-vl-72b-instruct
LLM_VL_MODEL_URL = <your-llm-host>/v3/llm.chat/chat/completions
LLM_VL_MODEL_APIKEY = <your-llm-api-key>
```
`core_setting.yaml` → `knowledge/llm_image_parse`（同上 api_key/base_url/model_name）。

### 4. Embedding 向量模型
`kinit.env.local`：
```
LLM_EMBEDDING_MODEL_URL = <your-embedding-host>/v1/embeddings
LLM_EMBEDDING_MODEL_APIKEY = <your-embedding-api-key>
LLM_EMBEDDING_MODEL_NAME = Qwen3-Embedding-0.6B
```
`core_setting.yaml` → `knowledge/embedding`（url/key/model_name 同上）。

### 5. PDF 转换服务
`core_setting.yaml` → `knowledge/convert_pdf`：
```
default.url = <your-pdf-convert-host>/forms/libreoffice/convert
ofd.url     = <your-pdf-convert-host>/forms/libreoffice/convert
```

### 6. JWT 密钥（随机定义）
`apps/corekg/conf/test/config.yaml`（已脱敏，本地生成后写入，勿提交以下示例值）：
```
jwt_secret: <your-random-jwt-secret>
plt_jwt_secret: <your-random-plt-jwt-secret>
```
若尚未生成，可用：
```bash
openssl rand -base64 32 | tr -d '/+=' | head -c 24
```

### 7. 未填（本次无需 / 生产环境再注入）
- 企微三应用密钥、腾讯云 SecretId/Key、SMTP 密码：`config.yaml` 中为历史生产凭据，本地无需启用；生产通过真实值替换。
- `LICENSE`：本次未启用（`kinit.env` 占位 `0`）。

---

## 三、初始化命令

> 注意：`keinit` 的 `--env-file` 只支持**单个**文件（`cmd/main.go` 中 `envFile` 为 `string`）。
> `ReadENV` 先读进程环境变量再读文件覆盖，因此把两个 env 都 source 到进程即可。

```bash
# 1) 编译 keinit
make local APP=keinit ENV=test

# 2) 准备本地真实配置
cp apps/keinit/conf/test/kinit.env apps/keinit/conf/test/kinit.env.local   # 若 .local 已存在可跳过
cp apps/keinit/conf/test/config.yaml.example apps/keinit/conf/test/config.yaml

# 3) 把默认 + 真实两个 env 都加载进进程环境（.local 覆盖同名占位）
set -a; . apps/keinit/conf/test/kinit.env; . apps/keinit/conf/test/kinit.env.local; set +a

# 4) 执行完整初始化（initDB -> chatmodel -> apikey -> migrate -> initES -> bucket -> core_settings）
./bundles/keinit -c apps/keinit/conf/test/config.yaml \
    --setting-file apps/keinit/conf/test/core_setting.yaml \
    --env-file apps/keinit/conf/test/kinit.env
```

仅更新系统设置（core_settings 表）可单独跑：
```bash
./bundles/keinit -c apps/keinit/conf/test/config.yaml \
    --setting-file apps/keinit/conf/test/core_setting.yaml \
    update-setting
```

---

## 四、初始管理员账号
由 `scripts/mysql/v1.0_2__insert_account.sql` 写入：
| 字段 | 值 |
|---|---|
| 登录账号 | `admin@admin.com` |
| 密码 | `admin123456`（如需改密请走账号体系接口，勿改已发布基线 SQL） |
