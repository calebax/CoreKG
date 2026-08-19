# apps/pipeline 本地运行 + 与 apps/corekg 闭环排障指南

> 目标：把 `apps/pipeline`（Python 文档知识摄入 worker）在本地跑起来，并让它承接 `corekg`
> 上传文档后的「解析 → 切块 → 向量化 → 入库」处理，与 `apps/corekg` 形成闭环。
>
> 背景：**pipeline 代码不用改，但它在「拉任务 / 回报任务」依赖的 `knowledge.GetPendingTask` /
> `knowledge.TaskCallBack` 两个 HTTP 端点，在本仓库开源化时被移除了**（`roc` 参照仓库里的
> `apps/keparser`）。下面是补全这两端点 + NATS 派发（B 方案 + 只跑 corekg 聚合）的落地记录，
> 以及配置/启动与排障清单。

---

## 1. 架构定位（先理解，再动手）

### pipeline 是什么

`apps/pipeline` 是从其他项目 copy 过来的 Python 摄入管线，含两个 worker：

| Worker | 入口 | 角色 | 消费任务类型 |
|---|---|---|---|
| **Doc Analyser** | `doc_worker_main.py` | S3 下载文档 → MinerU 解析/转 Markdown → 上传回 S3 | `ke.prase_pdf_task` |
| **Chunk Splitter** | `chunk_worker_main.py` | 下载 Markdown → 清洗/切块 → 图片/表格增强 → Embedding → 写入 ES | `ke.knowledge_task` |

两个 worker 都以「**HTTP 轮询**」方式工作：定时 POST 到
`{api_url}/knowledge.GetPendingTask` 拉任务、处理完 POST 到 `{api_url}/knowledge.TaskCallBack` 汇报。

```
   corekg（上传文档后写 core_task 任务 + 分发）
        │
        ▼
  knowledge.GetPendingTask  ←────────  pipeline 轮询（HTTP，无服务端实现！）
        │  返回 task_id + payload
        ▼
  pipeline worker 处理（S3/MinerU/切块/ES）
        │
        ▼
  knowledge.TaskCallBack  ───────────  回报任务结果（HTTP，无服务端实现！）
```

### corekg 里的任务系统

- 任务实体为 MySQL `core_task` 表（`pkgs/task`），任务类型常量见
  `apps/kecore/models/coretask/generate_task.go`：`ke.copy_task` / `ke.doc_to_pdf_task` /
  `ke.prase_pdf_task` / `ke.knowledge_task` 等。**任务类型字符串与 pipeline 一致。**
- 任务分发在 `pkgs/task/task_queue.go` 的 `PushTaskQueue()`：把 pending 任务 publish 到
  **NATS JetStream** `core.task.dispatch.<short>`，结果从 `core.task.result.<short>` 回报。
- 本分支已在 `apps/corekg/cmd/init.go` 新增 `initNATS`，让 **corekg 聚合进程默认也建立 NATS bridge**（读 `NATS_URL`，默认 `nats://localhost:4225`），
  `PushTaskQueue` 因此能派发；NATS 不可用时仅告警不阻断启动。
  `PushTaskQueue` 用 `PeekOnePendingTask`（不预置 running），保证 HTTP worker 仍能通过 `GetPendingTask` 拉到任务。

---

## 2. corekg 侧的闭环缺口（关键）

### 缺口 A：`knowledge.GetPendingTask` / `knowledge.TaskCallBack` HTTP 服务端不存在

全仓库搜 `GetPendingTask` / `TaskCallBack`，只有：
- `apps/corekg/cmd/main.go` 的 **license 校验豁免白名单**里的字符串（`"knowledge.GetPendingTask"`、`"knowledge.TaskCallBack"`，
  注释还写着「keparser api」）；
- 客户端侧的调用方：`apps/ketask/cmd/task.go`、`clients/task_worker/task.go`。

**没有任何 controller 注册并实现这两个 action。** 而 pipeline 正是要 POST 这两个端点；
`clients/task_worker`（Go worker）也是同一套 HTTP 轮询风格，同样依赖它们。所以这两个端点
在开源化时（`git log` 首个 commit 是 "open-source CoreKG (initial import)"）被移除了一个「keparser」
任务调度应用，但 pipeline 和 Go worker 客户端仍然按旧协议轮询。

### 缺口 B：corekg 聚合进程不发 NATS 任务

如「架构定位」所述，`apps/corekg` 未设置 NATS bridge → `PushTaskQueue` 失败 → 上传文档后
任务停在 pending。要发 NATS，需要 `apps/ketask` 进程起来（它初始化 bridge + stream + result 消费者）。

### 缺口 C：pipeline 的 ES 索引 / 索引名

- chunk worker 把结果写进 ES（`tools/es_storage.py`），索引名取自 `payload.es_index`，缺省用
  `config/chunk_config.yaml` 的 `ES.INDEX`。corekg 侧索引名来自 forest 的 `EsIndex()`（多为 `ke_0`…）。
- 切块写入的字段结构（`corekg_chunk/chunk.py` 的 `knowchunk_format_conversion`）与
  `apps/kesearch/models/chunk` 搜索时读取的字段需要能对上，否则 RAG/搜索查不到。这是**数据契约**层面
  的对接点，本地验证时建议先用 mock 或最小文档确认 ES 文档字段一致。

### 缺口 D：配置键名不一致（pipeline 内部）

- `corekg_chunk/task/work_chunk.py` 读的是 `config["ES"]["PASS"]` 和 `config["Work"]["API_URl"]`（注意大写眼尖），
  chunk 的 example 里写的是 `PASSWORD`、`api_url`——**直接 copy example 会 KeyError**。仓库内已给的
  `config/chunk_config.yaml`（本地版）用的就是代码实际读取的键，以它为基准即可。
- `tools/s3_minio.py` 读 `s3.endpoint_url / access_key_id / secret_access_key / region`，与 `analyser_config.yaml` 一致。

---

## 3. 已落地的实现（B 方案 + 只跑 corekg 聚合）

已在 `CoreKG-oss` 源码里补齐（参照 `roc` 仓库的 `keparser` + `pkgs/task/task_server.go`，但底层适配 NATS）：

| 改动 | 文件 | 说明 |
|---|---|---|
| HTTP 端点 DTO | `pkgs/task/biz.go` | `GetPendingTestRequest/Response`、`TaskCallBackRequest/Response` |
| 任务回调注册 | `pkgs/task/task.go` | `RegisterCallBack` / `GetCallBack` |
| 下一步推进 | `pkgs/task/crud.go` | `GetNextStepTask`；`SaveTask` 失败时按 `redo` 重入队；新增 `PeekOnePendingTask` |
| 端点实现 | `pkgs/task/task_server.go` | `GetPendingTask`（读 DB pending→置 running→返回 task_id+payload）、`TaskCallBack`（落库+回调） |
| NATS 派发修正 | `pkgs/task/task_queue.go` | `PushTaskQueue` 改用 `PeekOnePendingTask`，**不把任务预置 running**，保证 HTTP worker 仍能拉到 |
| 路由挂载 | `apps/corekg/internal/apis/apis.go` | 注册 `knowledge.GetPendingTask` / `knowledge.TaskCallBack`（挂在 `/v3/`） |
| NATS 桥接 | `apps/corekg/cmd/init.go` | 新增 `initNATS`：建 bridge + `EnsureStreams`；NATS 不可用仅告警不阻断 |
| 流转回调 | `apps/corekg/cmd/main.go` + `apps/corekg/internal/taskbiz/callbacks.go` | 注册 `copy→prase→knowledge` 链路的推进回调 + 更新文件状态 + `SuccessFile` |

### 已在本仓库 docker 环境验证

在运行中的 `corekg-net` 环境里用新构建二进制实测：

1. `GET /v3/knowledge.GetPendingTask` → 返回 `{task_id, payload}`（无可执行任务时 `code:404 暂无任务`）。
2. 插入一条 `ke.prase_pdf_task` pending 任务后，`GetPendingTask` 能取到并把任务置 `running`。
3. `TaskCallBack(status=success)` → 任务置 `success`，注册的回调**自动创建下一步 `ke.knowledge_task` pending 任务**。
4. NATS dispatch 流 `CORE_TASK_DISPATCH` 已建好，任务派发走 NATS。

> 用 txt/md 绕过 `doc_to_pdf` 阶段：直接用 `ke.prase_pdf_task`（不生成 `doc_to_pdf` 前置），见上文「缺口」下的说明。

---


### 方案 X（推荐，改动最小）：手工建任务 + 直读库回报

1. 不依赖任何 broker，**由你（或外围脚本）向 `core_task` 表写入一条 pending 任务**，
   再让 pipeline 的 `TaskQueue.get_task` 改为「读 `core_task` 表」、`callback` 改为「更新 `core_task` 表」。
   但这需要改 `apps/pipeline/.../task/*.py` 里的 `TaskQueue`，与「不改代码」矛盾，仅作兜底思路。

### 方案 Y（贴近原架构，推荐）：临时补一个「HTTP 调度」轻量服务

在 corekg 聚合或一个独立小服务里，给 `knowledge.GetPendingTask` / `knowledge.TaskCallBack`
实现最小 handler，读写同一个 `core_task` 表：

- `GetPendingTask`：按 `task_type` 取一条 pending 任务返回 `{task_id, payload}`；可选地把状态置为 running。
- `TaskCallBack`：按 `task_id` 更新状态/结果；参考 `apps/ketask/internal/jobs/result_consumer.go`
  里的 `makeForestResultHandler`（把 NATS result 处理逻辑照搬到 HTTP handler 即可），
  顺带推进 `copy → prase → knowledge` 的下一步任务。

> 这样无需跑 NATS/ketask，pipeline 就能闭环。核心就是把 result_consumer 的 NATS 分段换成 HTTP handler。

### 方案 Z（原方案，需额外进程）：还原 NATS + ketask

1. 启动独立 `apps/ketask`（它初始化 NATS bridge/stream + result 消费者）。
2. 给 `knowledge.GetPendingTask`/`TaskCallBack` 补 HTTP handler（同上，因为端点本身也不存在）。
3. pipeline 照旧走 HTTP 轮询即可。

无论 X/Y/Z，**`knowledge.*` 两个 HTTP 端点都需要补服务端实现**——这是绕不过去的点。

### 用改动后的 corekg 跑起来（本地，只跑聚合）

```bash
# 1) 构建聚合二进制
make local APP=corekg ENV=test          # -> bundles/corekg
# 或：go build -o bundles/corekg ./apps/corekg/cmd

# 2) 启动依赖（MySQL/ES/Redis/MinIO/NATS）与配置
docker compose up -d
# 确保 NATS 宿主机端口 4225 可用；NATS_URL 缺省走 nats://localhost:4225

# 3) 运行 corekg 聚合（新二进制已注册 knowledge.* 端点 + NATS bridge + 流转回调）
export ENABLE_NEBULA_GRAPH=false
./bundles/corekg -c apps/corekg/conf/test/config.yaml

# 4) 拉起 pipeline 两个 worker（指向同一 corekg）
cd apps/pipeline && source .venv/bin/activate
python doc_worker_main.py    # 消费 ke.prase_pdf_task（先手动/经上传产生任务）
python chunk_worker_main.py  # 消费 ke.knowledge_task
```

> NATS 不可用时聚合进程仅告警不退出，`GetPendingTask`/`TaskCallBack` 等 HTTP 路径仍可用；
> 但「上传即派发」依赖 NATS 转发（供自有 worker），把 `NATS_URL` 配到可达地址即可。

---

## 4. pipeline 本地启动（依赖已就绪前提下）

### 依赖：基础中间件

```bash
docker compose up -d                # MySQL/ES/Redis/MinIO/NATS（端口均为 +2）
make local APP=keinit ENV=test      # 初始化 DB/ES 索引/bucket/账号（一次即可）
./bundles/keinit ...                # 见 docs/local-config-checklist.md
```

### 需要外部服务

| pipeline 依赖 | 说明 | 无真实服务时的替代 |
|---|---|---|
| S3（MinIO） | `analyser_config.yaml` 的 `s3` 段，下载+上传文档 | `docker compose` 自带 MinIO |
| MinerU REST API | `analyser_config.yaml` 的 `analyser_api_url`，PDF/图片转 Markdown | 缺则 PDF 解析失败（建议先用 txt/md 绕过） |
| LLM / VLLM / Embedding | `chunk_config.yaml`，表格/图片增强 + 向量化 | 无真实模型时把 `STATUS` 关掉，Embedding 仍需可用接口 |
| Elasticsearch | `chunk_config.yaml` 的 `ES` | `docker compose` 自带（首字节需 `ke_*` 索引，可用 keinit `es-init`） |

### 配置文件

仓库已给 `apps/pipeline/config/analyser_config.yaml` 和 `chunk_config.yaml`（本地版，端口已对齐
本机 +2：ES `localhost:9202`、MinIO `localhost:9002`、corekg `localhost:8080`）。检查/修改：

- `analyser_config.yaml`：
  - `api_url` → corekg（或临时调度服务）地址，如 `http://localhost:8080/v3`
  - `analyser_api_url` → MinerU 服务地址
  - `s3.*` → MinIO 凭证（`minioadmin/minio123456`）、`bucket` → `corekg-bucket`
- `chunk_config.yaml`：
  - `Work.API_URl` → corekg 地址；`Work.TASK_TYPE` → `ke.knowledge_task`
  - `ES.*` → ES 连接与索引（`HOST: http://localhost:9202`，`ACCOUNT/PASS`、`INDEX: ke_0`）
  - `LLM/VLLM/Embedding` → 真实模型，否则把 `STATUS` 置 `false`

> ⚠️ 注意键名按上方「缺口 D」改准（`ES.PASS`、`Work.API_URl` 等）。

### 启动（需先解决第 3 节缺口）

```bash
cd apps/pipeline
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt -r requirements_analyser.txt

# 文档解析 worker
python doc_worker_main.py

# 切块 worker（另开一个终端）
python chunk_worker_main.py
```

docker 方式：`docker-compose.pipeline.yml` 中 `pipeline` 服务的构建上下文 `context: apps/pipeline` +
`dockerfile: Dockerfile`，但 **`apps/pipeline` 根目录没有 `Dockerfile`**（只有 `private/docker/*/Dockerfile`，
且是 Nuitka 加密构建）；该 compose 里 `scripts/mock_embedding.py` 也不存在。所以该 compose 文件当前
**无法原样 `docker compose up -d --build`**，需补一个普通 `Dockerfile`（直接 COPY 源码 + `requirements.txt`）。

---

## 5. clients/ 里的接口是否够用

`clients/task_worker`（Go，`clients/task_worker/`）与 `apps/ketask/cmd/*` 是**同一协议、同一风格**的
worker 客户端，都调用：

- `POST {base}/v3/knowledge.GetPendingTask`（`clients/task_worker/task.go` → `GetPendingTask`）
- `POST {base}/v3/knowledge.TaskCallBack`（`CallBackTask`）

它们**提供的是「消费端」，不提供「服务端端点」**——即它们和 pipeline 一样都是需要
`knowledge.*` 端点的调用方，不能反过来满足 pipeline 的需求。因此：

- ✅ 可参考它对 payload 的解释（`apps/ketask/models/ragtask/payload.go`）作为字段契约。
- ❌ 不能把它当作 pipeline 的「接口」；服务端 `knowledge.*` 端点仍需补齐（见第 3 节）。

---

### 5.1 MinerU（PDF/图片→Markdown）

`analyser` 的 PDF/图片解析走 `analyser_api_url`，代码要求（`analyser_process.py`）：
`POST multipart/form-data` 上传 `files` + `return_md/response_format_zip/return_images/...` 等 flag，
返回 **ZIP 流**，解压得到 markdown + 图片 + bbox。这正是官方 **MinerU `mineru-api`** 的 `/file_parse` 契约
（`mineru-api --host 0.0.0.0 --port 8000`，端点 `POST /file_parse`、`GET /health`）。

本地环境使用**已提供的 MinerU 地址**：
`analyser_config.yaml` 的 `analyser_api_url = https://mineru.test.i.yygu.cn:58081/file_parse`。
无需在本地 docker-compose 里部署 MinerU（已从 compose 移除）。

若以后要改为自部署（有网络/GPU）：
```bash
pip install "mineru"                      # CLI/API 一体，默认 CPU
mineru-api --host 0.0.0.0 --port 8000     # 提供 /file_parse
```
- 首次运行自动预下载模型（写默认模型目录），联网后无需人工干预。
- 把 `analyser_api_url` 指到对应地址即可；仅 txt/md 场景可留空跳过 MinerU。

> ⚠️ **性能/资源**：完整 PDF 解析较重（CPU 推理慢、内存占用高；有 GPU 加 `--gpus all`）。
> 本地验证「闭环」用 txt/md 最快；要验证 PDF 渲染再跑 MinerU。

---

## 6. 落地建议清单（最小动作）

1. 确认是要「跑通链路」还是「生产级闭环」；本地建议走**方案 Y**。
2. 补 `knowledge.GetPendingTask` / `knowledge.TaskCallBack` 两个 HTTP handler（读写 `core_task`，
   参考 `apps/ketask/internal/jobs/result_consumer.go` 的推进逻辑），或先写死一条任务验证 pipeline 单端。
3. 校验 `apps/pipeline/config/*.yaml` 键名与地址（缺口 D）。
4. 起中间件 + corekg，上传一篇 txt/md（绕过 MinerU），观察 pipeline 是否拉到任务、ES 是否写入。
5. 用 `apps/kesearch` / corekg 搜索接口确认 ES 文档可检索（缺口 C 的数据契约）。

> 排查线索汇总：
> - pipeline 一直「无任务」→ `knowledge.GetPendingTask` 没实现，或 corekg 没发 NATS（缺口 A/B）。
> - pipeline 报 KeyError → 配置键名不对（缺口 D）。
> - 解析 PDF 失败 → MinerU 未起（`analyser_api_url`）。
> - 入库成功但搜不到 → ES 索引名/字段契约不符（缺口 C）。
