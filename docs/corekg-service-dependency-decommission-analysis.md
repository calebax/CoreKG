# CoreKG 服务依赖、独立性与裁撤分析

> 本文只讨论产品运行所需的代码服务、数据服务、算法服务和外部能力，不讨论接入层、编排层、调度或制品管理。MySQL、Redis、MinIO、Elasticsearch、NebulaGraph，以及 OpenCoze 使用的 NSQ、etcd、Milvus 等仍属于服务依赖：无论采用何种运行环境，产品都必须提供这些能力或兼容替代品。
>

## 1. 服务边界与插拔等级

本文所称“服务”包括：

- CoreKG 项目服务：`corekg`、`web`、`keinit`、知识处理 worker、转换服务、Agent 工具服务、OpenCoze、Editor/Docs 等。
- 数据服务：MySQL、Redis、MinIO/S3、Elasticsearch、NebulaGraph、Milvus、etcd、NSQ。
- 外部能力：Embedding、Rerank、LLM/VLM、PDF parser、`/split`、SMS、微信支付、企业微信 webhook 等。
- 兼容实现：只要满足本文列出的协议、数据和错误语义，可以替换具体产品或部署方式。

插拔等级：

| 等级 | 定义 |
|---|---|
| `H0 不可插拔` | 当前是启动或核心数据链硬依赖；删除会使主服务不可用 |
| `H1 冷替换` | 存在接口或存储抽象，但需迁移数据、改配置并重启 |
| `H2 功能级冷插拔` | 可关闭某项产品功能后删除，但必须同步改代码、配置、路由或任务 DAG |
| `H3 契约级可插拔` | 可按既有 HTTP、Task 或 OpenAI-compatible 契约替换实现；切换通常仍需重启 |
| `H4 热扩缩` | 同实现消费者可在线增加或减少实例；不代表可以删除最后一个实例 |

CoreKG 当前没有统一的动态插件框架。HTTP 调用或独立进程不等于热插拔；只有调用方能够在依赖消失时稳定降级，并且不需要重启或数据迁移，才属于真正热插拔。

## 2. 服务级依赖、独立性与裁撤分析

### 2.1 私有化服务数量统计

统计快照：2026-07-10。统计按“一个独立运行并提供业务、算法、数据或运维能力的组件计一个服务单元”，接入层和进程编排对象不重复计数。

统计口径：

- 接入层 `corekg-traefik` 不计入服务总数。
- 外部托管的 LLM/VLM、Embedding、Rerank、PDF parser、`/split`、通用 analyser、SMS/支付等不计入当前私有化服务数；若交付时也要本地化，应按实际服务拆分另行增加。
- OpenCoze 使用的 MySQL、Redis、Elasticsearch、MinIO 与 CoreKG 共享，只在共享数据服务中计数一次。
- 一次性初始化或控制任务、服务发现对象、进程副本和资源别名不单独计数。

| 分类 | 数量 | 计入的服务 |
|---|---:|---|
| CoreKG 项目与算法服务 | 17 | `corekg`、`web`、`keinit`、`ai-summary-worker`、`doc-analyzer`、`doc2pdf-worker`、`file-convert`、`graphrag-chunker`、`graphrag-graph`、`html2docx`、`mcp-chart-opton`、`ofd2pdf`、`rerank-emb`、`sandbox`、`web-docs`、`web-editor`、`word2pdf` |
| 共享数据服务 | 7 | `mysql`、`redis`、`minio`、`elasticsearch`、`nebula-graphd`、`nebula-metad`、`nebula-storaged` |
| OpenCoze 专用服务 | 6 | `opencoze-web`、`opencoze-server`、`opencoze-nsq`、`opencoze-nsqlookupd`、`opencoze-etcd`、`opencoze-milvus` |
| 纯运维观测服务 | 0 | 当前 `release-2.13` chart 未启用 Kibana；`templates/elasticsearch.yaml` 中 Kibana 资源为注释块 |
| **排除接入层后的全部服务单元** | **30** | 17 + 7 + 6 |

数量口径结论：

- **已确认产品服务为 30 个**：CoreKG 项目/算法 17 个 + 共享数据 7 个 + OpenCoze 专用 6 个。
- **业务相关服务为 30 个**：当前不再保留待确认业务服务候选。
- **排除接入层后的全部长期运行服务为 30 个**：当前 chart 中没有额外计入的纯运维观测服务。
- `corekg-chart` 的 `release-2.13` 模板只定义 `ofd2pdf` Deployment/Service，`knowledge:convert_to_pdf` 的 OFD URL 也指向 `http://ofd2pdf:8000/api/convert`；因此 `ofd2pdf-service` 属于过时名称，不计入服务清单。`rerank-emb` 虽未被当前 CoreKG Rerank CoreSetting 选中，仍作为实际运行的项目算法服务计入。
- 30 不是最小可运行服务数。Graph、OpenCoze、Editor、帮助站和 Rerank 等存在可选或待解耦能力；必须完成相应代码、配置和数据解耦后，才能形成不同功能组合的最小数量。

### 2.2 CoreKG 项目核心服务

| 服务 | 职责与上游 | 下游依赖 | 故障影响 | 插拔/裁撤结论 |
|---|---|---|---|---|
| `corekg` | 聚合 `kecore`、`kechat`、`account`、`keapi`、`keparser`、`kesearch` 和自身 API | 业务 MySQL、Redis、MinIO、ES、Embedding、Highlight、条件性 Nebula、LLM、Coze、Sandbox 等 | 主 API、文件、搜索、Chat、账号等整体失效 | `H0`；只能先拆分聚合模块，不能直接删除 |
| `web` | `roc-web/ai` SPA | `corekg`、OpenCoze Web、web-editor、web-docs | Web UI 不可用；外部 API 可独立保留 | 构建 headless API 产品后可删除；本身不持有业务数据 |
| `keinit` | 初始化数据库结构/配置、ES mapping、MinIO bucket，并提供初始化状态 | MySQL、ES、MinIO | 初始化、重启后的服务可用性受影响 | 当前为启动链依赖；要裁撤需把 migration、seed、mapping、bucket 创建等职责迁移到可审计初始化服务或流程 |

`keinit` 的状态接口不是所有初始化副作用成功的充分证明：MinIO 普通 bucket/Coze bucket 创建重试耗尽后仍会继续启动状态服务。因此迁移 `keinit` 时不能只复制一个健康接口，必须逐项迁移初始化语义。

### 2.3 项目共享数据服务

| 服务 | 被谁依赖 | 必须保持的数据/协议语义 | 插拔结论 |
|---|---|---|---|
| MySQL | `corekg`、`keinit`、Task worker、OpenCoze | `core/account/knownow/chat/coze` 数据源、事务、索引、外键、Task claim 的 `FOR UPDATE SKIP LOCKED` | `H0/H1`；替换需完成 schema、数据、锁和事务语义迁移 |
| Redis | Task、SSE、cache、lock、token、OpenCoze | Redis Streams、consumer group、普通 KV、锁、过期语义 | `H0/H1`；初始化失败后继续运行不代表可选，后续功能可能报错或 panic |
| MinIO/S3 | 浏览器上传、Forest/Chat 文件、worker 输入输出、OpenCoze | S3 API、bucket/path、presigned PUT、CORS、public/signed URL、目录产物 | `H1`；需迁移对象和 URL 语义；MR !428 的附件 Task 还有 S3 bucket 硬约束 |
| Elasticsearch | Chunk、向量/关键词搜索、FQA、描述、Chat History、OpenCoze | index/mapping/analyzer/DSL、nested references、向量维度和相似度 | `H0/H1`；替换前需重建索引并验证召回语义 |
| NebulaGraph | Graph API、图检索、Graph/Chunk worker | Nebula 协议、space/schema、既有图数据 | `H2`；只有完整关闭 Graph 能力后才能删除 |

这些基础服务属于产品依赖而非部署细节。它们可以部署在同一主机、集群或云服务中，但接口和数据语义必须存在。

### 2.4 知识任务与算法服务

知识文件主任务链：

```text
ke.doc_to_pdf_task（按格式需要）
  -> ke.prase_pdf_task
  -> ke.knowledge_task
  -> ke.description_task
  -> ke.success_file_task
```

| 服务 | Task/调用契约 | 能否独立去除 | 必须同步处理 |
|:--|---|---|---|
| `doc-analyzer` | 消费 `ke.prase_pdf_task`；PDF/Office -> `content.md` 等产物 | 否，除非禁用全部 PDF/Office 知识入库和 Chat 附件 | 前端格式、Task 生成、MR !428 附件等待、在途任务 |
| PDF parser | `doc-analyzer` 调用；按约定目录产生 Markdown 等文件 | 可按 `H3` 替换，不能无替代删除 | HTTP payload、超时、业务状态、产物目录、错误语义 |
| `doc2pdf-worker` | 消费 `ke.doc_to_pdf_task` | 仅在禁用需转换格式后可删 | Task step、格式白名单、预览状态、失败提示 |
| `word2pdf` | Office/PPT -> PDF | 可按格式关闭或兼容 HTTP 契约替换 | `knowledge:convert_to_pdf`、转换内容完整性 |
| `ofd2pdf` | OFD -> PDF；当前 chart 中唯一 OFD 转换服务名 | 禁用 OFD 后可删或兼容替换 | `knowledge:convert_to_pdf` 的 OFD URL、文件类型、转换配置、失败提示 |
| `graphrag-chunker` | 消费 `ke.knowledge_task`；调用 `/split` 并写 ES | 主知识检索不可删除 | Task DAG、ES、Embedding、Nebula client、在途任务 |
| `/split` | `POST /split`；承担分块和索引 side effect | 可兼容替换，不能无替代删除 | 当前 worker 未转发 Task 中的 SplitConfig；必须固定真实默认参数和 ES 写入契约 |
| `ai-summary-worker` | 消费 `ke.description_task`；生成摘要、描述、推荐问题 | 不能直接删除，否则 DAG 停在 description step | 从 DAG 跳过该 step，并调整最终状态、数据字段和前端展示 |
| `graphrag-graph` | 消费 `ke.graph_file_task` | 关闭全部图能力后可删 | Task 生成、Graph API/UI、Nebula、既有任务和数据 |
| `file-convert` | `.xls -> .xlsx`、Strict OOXML `.xlsx -> Transitional .xlsx` | 禁用这两类输入后可删；普通 XLSX/CSV 不依赖 | 格式策略、调用方错误提示、产物兼容 |
| Embedding | OpenAI-compatible embeddings；向量契约与 ES mapping 绑定 | 不能无替代删除 | model、dimension、normalization、batch、error 兼容；更换模型通常需重建向量 |
| LLM/VLM | Chat、rewrite、summary、description、Graph 等 | 对相关 AI 功能不可删除 | SSE、tool/function calling、JSON、图片、多轮上下文、超时和取消语义 |
| Rerank | 检索候选重排 | 可先关闭功能，再删除或替换 | `enable_rerank=false` 后需重启加载配置，并验证 ES recall/fallback 质量 |
| `rerank-emb` | 本地 Rerank/Embedding 实现 | 当前 CoreKG CoreSetting 未选择该实现，不是当前主链直接依赖 | 只能列为候选；删除前仍需确认其他项目或仓库外调用方 |

worker 的契约级可插拔要求同时保持：Task type、Poll/Callback API、MySQL 状态机、Redis Stream、优先级、超时、redo、MinIO 路径、产物结构、ES/Nebula side effect 和幂等语义。满足时实现可以 `H3` 替换，同类实例可 `H4` 横向调整；否则只是重写服务，不是插件。

### 2.5 Agent、工具与 OpenCoze

| 服务 | 依赖关系 | 裁撤结论 |
|---|---|---|
| `sandbox` | Agent Code Tool、Forest Agent、React Excel Chat | `H2/H3`；对 React Excel Chat 是请求级硬依赖，需先移除该模式/工具 |
| `mcp-chart-opton` | Agent/Excel Chat 的 Chart MCP | `H2/H3`；从 `agentenv.mcp` 和已保存 Agent 工具配置中移除 |
| `opencoze-web` | `/agents` 嵌入 UI 和 `/coze` UI | 关闭或替换 Agent/Coze UI 与相关跳转后可删 |
| `opencoze-server` | Coze API、Chat、Plugin、Workflow、模型/Space 同步 | 解除 ChatModel/组织同步耦合后可删 |
| `opencoze-nsq`、`opencoze-nsqlookupd` | OpenCoze 异步消息 | 仅随 OpenCoze Server 整体裁撤；不是 CoreKG 独立依赖 |
| `opencoze-etcd` | OpenCoze/Milvus 元数据协调 | 仅随 OpenCoze/Milvus 整体裁撤 |
| `opencoze-milvus` | OpenCoze 向量存储 | 仅随 OpenCoze 整体裁撤；先确认向量数据保留需求 |
| OpenCoze 使用的 MySQL、Redis、ES、MinIO | 模型、业务数据、缓存/消息、搜索和对象存储 | 如果与 CoreKG 共享实例，裁撤 OpenCoze 时只能删除 OpenCoze schema/bucket/index/key，不能删除共享服务 |

OpenCoze 不是纯前端可选页。ChatModel Create/Update 会无条件调用 `SyncCozeModelInstance`；不支持 function call 且没有 Coze 绑定时函数会跳过，对支持 function call 或已有绑定的模型会发生 Coze DB 写入。Delete 仅在已有 `CozeModelID` 时操作 Coze。发生跨库写入时没有统一事务，因此裁撤前必须先解除该耦合。

### 2.6 辅助项目服务

| 服务 | 项目关系 | 裁撤结论 |
|---|---|---|
| `web-editor` | `/editor/:id` 使用的 AI 编辑器 | 取消编辑器路由、入口和数据兼容后可删；内部源码不在本次 `ai` 范围 |
| `html2docx` | HTML -> DOCX 导出能力 | 属于 Editor/导出功能依赖；应先核对 web-editor 调用契约，再按 `H2/H3` 裁撤或替换 |
| `web-docs` | 静态帮助文档 | 取消帮助入口后可直接裁撤，不影响业务数据 |

RabbitMQ `rpcqueue` 虽在仓库保留 package，但未发现仓内调用者；当前知识任务使用 Redis Streams。`insert_index` worker 也未找到任务生成调用点。这两项属于代码遗留候选，不应被写成当前产品硬依赖；删除前仍需确认仓库外调用者。

## 3. CoreSetting 与外部服务边界

以下只记录依赖语义，不记录地址或凭据：

| Group:Key | 依赖 | 阶段 | 语义 |
|---|---|---|---|
| `core:redis` / `knowledge:redis` | Redis | 启动/运行 | Task、缓存、SSE 等使用不同逻辑 DB |
| `core:cos-ke` | Forest 文件存储 | 启动/运行 | MinIO/S3 语义 |
| `core:cos-yg-chat` | Chat 附件存储 | 请求 | MR !428 Attachment 使用 |
| `core:cos-cu-image`、`cos-company-logo` | 账号/Logo 图片 | 请求 | Account/品牌功能使用 |
| `knowledge:es` | Elasticsearch | 启动/运行 | Chunk、History、FQA、Search |
| `knowledge:embedding` | Embedding | 启动/运行 | OpenAI-compatible endpoint |
| `knowledge:rerank` | Rerank | 请求 | 当前使用外部 Rerank endpoint |
| `knowledge:reranksearchcfg` | Rerank 开关/fallback | 启动加载 | 修改后需重启相关 CoreKG 进程 |
| `knowledge:graphsearchcfg` | Graph Search 行为 | 启动加载/请求 | 图检索配置 |
| `knowledge:nebula` | NebulaGraph | 条件性启动/运行 | Graph 能力 |
| `knowledge:convert_to_pdf` | Word/OFD/PPT 转 PDF | 请求/任务 | 转换服务契约 |
| `knowledge:mysql_excel_instance`、`mysql_excel_instance_readonly` | Excel Forest database | 创建/查询 | 与租户动态 DB Knowledge 连接不同 |
| `knowledge:highlight` | Highlight | 启动 | 配置加载失败会阻止主服务启动 |
| `knowledge:sandbox` | Sandbox | Agent/Excel 请求 | Code Tool |
| `knowledge:proxy_llm_models`、`knowledge:system_llm_api_key` | LLM | 请求 | 外部模型服务 |
| `corekg:core_file_convert` | 文件转换 | 请求 | `.xls`/Strict XLSX 转换 |
| `corekg:agentenv` | Sandbox、MCP Tool | Agent 请求 | Agent 工具环境 |
| `corekg:coze_url`、`coze_ip` | OpenCoze | 请求 | Coze API/UI |
| `corekg:corekg_url` | Coze Plugin callback | 请求 | Coze 回调 CoreKG |
| `corekg:yg_api_analysis_file` | 其他附件分析 | 请求 | 非 PDF/Office/纯文本附件 analyser |
| `account:pkl_connect_providers` | OAuth/Connector | 启动 | 可为空，但初始化代码仍存在 |

`chat_model` 中的 LLM endpoint 是运行时外部服务依赖，不要求 LLM 与 CoreKG 部署在同一环境。模型管理可以动态修改 endpoint 元数据，但系统模型选择、缓存和调用兼容性仍需要单独验证。

## 4. 按服务/能力裁撤的代码清单

本章描述“决定去掉某项服务或能力时，需要同步改哪些代码和数据”，不是执行步骤。

### 4.1 去掉 OpenCoze

后端：

- `apps/kechat/models/coze`、Coze API/Service、Plugin/Workflow/External Chat。
- `apps/kecore/models/coze` 及组织/Space 同步。
- ChatModel Create/Update 中 `SyncCozeModelInstance` 的无条件调用，以及已有绑定模型的 Delete 分支。
- ChatModel/Agent 表中的 Coze ID 映射、迁移和兼容读取。
- `corekg:coze_url`、`coze_ip`、`corekg_url` 及相关配置。

前端：

- `/agents` 的嵌入 UI、`/coze` UI/相关跳转、Coze 登录/直接 API，以及仅 Coze 支持的编辑/发布能力。
- 根据产品选择恢复本地 Agent 页面或彻底下线 Agent。

数据与服务：

1. 先停止新写入并处理 ChatModel/Agent/Space 映射；
2. 再删除 OpenCoze Web/Server 业务；
3. 最后删除只被 OpenCoze 使用的 NSQ、Milvus、etcd 数据和服务；
4. MySQL、Redis、ES、MinIO 若与 CoreKG 共享，只清理 OpenCoze 专属数据，不删除共享服务。

### 4.2 去掉图谱/NebulaGraph

后端：

- `apps/corekg/cmd/main.go` Nebula 初始化。
- `kecore` Graph Router/Model/Service、`kesearch` Graph Search、文件删除时的图清理。
- `ke.graph_file_task` 生成与回调、worker Nebula PreRun。
- `knowledge:nebula`、Graph Search 配置和不完整的环境开关。

前端：

- `/graph/**`、知识库 Graph/关系搜索/Graph 可视化入口。
- 旧 WordCloud/Knowledge Graph 页面及 mock fallback。

数据：先确定 Nebula space/schema 和既有图的导出、保留或删除策略。仅设置 `ENABLE_NEBULA_GRAPH=false` 不能完整裁撤。

### 4.3 去掉 Rerank

1. 把 `reranksearchcfg.enable_rerank=false`；
2. 重启相关 CoreKG 进程，使包级缓存重新加载；
3. 验证 ES recall、keyword/FQA fallback 的质量和延迟；
4. 再删除 Rerank client、CoreSetting 或替换 endpoint。

这是 `H2/H3` 能力，应该先关闭功能并验证 fallback，不应先删除服务制造故障。

### 4.4 去掉 Sandbox 或 Chart MCP

- 从 `corekg:agentenv` 移除对应工具 endpoint。
- 修改 `pkgs/einotools/tools` 的工具注册和 Agent prompt/权限。
- 处理已保存 Agent 的 tool 配置。
- 前端隐藏代码执行、Excel/图表工具能力。
- React Excel Chat 对 Sandbox 是硬依赖；必须先删除或重写该模式。

不要把 Keapi MCP Server 和 Agent MCP Client 当成同一项能力误删。

### 4.5 去掉 Chat 附件解析

- `apps/kechat/internal/apis/file_api.go`：上传响应和解析错误语义。
- `apps/kechat/services/svcfile/attachment_parse.go`：格式分流、Task 创建和同步等待。
- `apps/kechat/chat/modes/direct_model_chat.go`：`md_url`、内联内容和 file-analysis tool 分支。
- `pkgs/einotools/filecontent`：附件内容读取契约。
- 前端 `ai/src/components/dialog/AttachmentList`、DialogInput、Attachment DTO 和格式/状态提示。

产品语义必须明确选择：完全禁止附件、只保留纯文本，或保留上传但不分析。不能只停止 `doc-analyzer`。

### 4.6 去掉 AI 写作但保留文章

- 同时移除新旧 AI Write API，不能只删 deprecated 接口。
- 隐藏 `/write` 的 AI 生成入口；保留 Article CRUD 时继续保留编辑、历史和权限。
- 删除 `web-editor` 时同步删除 `/editor/:id` 入口和编辑器数据兼容。
- 保留 Project 时继续保持 Session `subject_id`；删除 Project 前先迁移会话关系。

### 4.7 去掉 Keapi MCP 或 REST

- MCP：取消 `keapimcp.RegistryRouter`、Tool 定义和 callback 配置；保留仍被 Agent 使用的 MCP Client。
- REST：取消 Forest/File/Chat/Search API、API Key 文档和外部调用方。
- 聚合模式当前没有调用 `InitMCPCfg`，默认 callback 指向 `localhost:8086`；无论保留还是裁撤都应先明确该边界。

### 4.8 去掉 Account

***Account 是重构项目，不是简单服务裁撤：***

- 替换登录 middleware、API Key、OAuth、员工/组织/角色/权限/配额。
- 为 Forest、File、Chat、Agent、Project、Article 建立新的 tenant/user identity。
- 迁移 `uin/company_id` 外键和对象存储路径。
- 重写前端 Auth、设置、组织、人员和权限组件。
- 重新定义审计和数据隔离。

这些工作完成前，Account 为 `H0`。

### 4.9 去掉特殊知识库

Excel/CSV：

- 删除 Excel 知识库类型、前端 Excel 页面、Project `react_excel_list` 映射和 `ListExcelSheet`。
- 改造 `svcforest.CreateForest` 的 Excel 自动建库、`svcforestfile` Excel 特判、`svcexcelforest`、新旧 Excel Chat mode。
- `ForestDB/ForestTable` 也被 DB Knowledge 共用，不能随 Excel 误删。
- 只有不再接收 `.xls`/Strict `.xlsx` 且无其他消费者时，才可删除 `file-convert`。

DB Knowledge：

- 删除前端 DB Knowledge 页面、API 和 Project mysql-table 分支。
- 删除 Forest DB Router、`forestctl/db.go`、`svcdbforest`、`SyncMysqlTable` cron、DB Chat session 分支和 `qachat/mysql_chat.go`。
- 目标数据库是租户动态配置的外部 MySQL，不是 CoreKG 固定数据服务。

QA Library：

- 删除 QA 页面/API/Project 入口和后端 qapair/FQA CRUD。
- 清理 FQA recall 前确认普通 Chat fallback 不再引用。
- 不影响标准文件 CoreTask，但会减少 ES/Embedding 使用场景。

CAD：先迁移既有 `forest_type=cad`，再删除 CAD 页面、route、enum 和校验。CAD 没有独立算法服务，既有 CAD 本质使用普通 PDF Task 链。

WordCloud：如果保留新版 Graph，只删除旧 WordCloud API/页面和 `nbqueue` 查询；不要误删新版 Graph Router。若同时删除 Nebula，再执行 4.2 的完整裁撤。

### 4.10 去掉管理与辅助能力

- License：同步去掉前端启动 gating、License 页面、后端 API/middleware 和模块许可判断。
- 支付：删除订单/购买 UI、`pay.ts` 和 sale Router/Service。聚合 `corekg` 当前未调用 `kesale.Init`，Router 可见不代表支付服务已启用。
- `/version`：删除 SMS、Redis 验证码、Account DB 和企业微信 webhook 调用链。
- 同义词/行业术语：先删除 Forest Chat 热路径中的两次 replace，再清管理 UI、MySQL Service 和 Router。
- 标签/公告/消息：协调删除 UI/API/DAO，不影响其他服务。
- Personnel：除非已替代组织、权限和 Agent 人员选择，不能整体删除；企业微信 callback Router 可单独裁撤。
- 品牌/Logo：可删除定制 UI、API 和专用 storage purpose，但不能据此删除被知识文件共享的对象存储。

## 5. 迁移与替换契约

### 5.1 对象存储

替代服务至少兼容：

- 后端上传、下载、列举和删除。
- 浏览器 presigned PUT 和 CORS。
- public URL 或 signed URL 的读取时长。
- bucket 和既有对象路径。
- worker 目录上传以及 `content.md` 等算法产物。
- MR !428 Task payload 的 S3 bucket 读取。

迁移要包含对象全量复制、路径映射、双读或回源期、hash 校验、旧 URL 兼容和回滚窗口。

### 5.2 Elasticsearch

替代服务至少兼容：

- Chunk、History、FQA、Description 等 index。
- analyzer、nested reference、keyword/vector DSL。
- Embedding 维度和相似度语义。
- worker 写入与 `corekg` 查询的一致性。
- 数据重建、别名切换和抽样召回对比。

只迁移源文件而不重建索引，不等于完成 ES 迁移。

### 5.3 Embedding、Rerank、LLM

- Embedding：请求/响应、batch、向量维度、归一化和错误码必须兼容；更换模型通常要重建向量。
- Rerank：候选输入、score 方向、top-n、token 限制和 fallback 必须兼容。
- LLM：OpenAI-compatible 只是表层；还要验证 SSE、tool/function calling、JSON、图片、多轮上下文、超时和取消。

### 5.4 Worker 与算法服务

替换 worker、PDF parser、`/split` 等服务时必须冻结：

- Task type、priority、app_group、step/next step。
- Poll/Callback API 和鉴权。
- timeout、redo 和幂等键。
- 对象存储输入输出路径和产物完整性。
- ES/Nebula side effect。
- Task success 的业务判定。

当前部分 worker 只检查 HTTP 成功或弱校验产物。替换服务时应补充业务响应、非空产物和内容完整性校验。

### 5.5 MySQL 与 Redis

MySQL 替换必须验证：

- 多 datasource、事务隔离、字符集、索引和唯一约束。
- `FOR UPDATE SKIP LOCKED` 的 Task claim 行为。
- `core_task`、业务表和 OpenCoze schema 的一致迁移。
- 跨库写入无法由单库事务覆盖的补偿逻辑。

Redis 替换必须验证：

- Streams、consumer group、`NoAck` 和补推语义。
- SSE、lock、token、TTL 和 Redis key 前缀。
- CoreKG 与 OpenCoze 共用实例时的数据隔离。

### 5.6 OpenCoze 数据服务

如果保留 OpenCoze，替换其依赖时要分别保证：

- NSQ/nsqlookupd：消息发布、发现、消费确认和重试。
- etcd：元数据协调和一致性。
- Milvus：collection/schema、向量维度、index 和查询语义。
- MySQL/Redis/ES/MinIO：OpenCoze 数据模型与共享资源隔离。

如果删除 OpenCoze，必须先停止写入、导出或确认丢弃策略，再删除专属数据服务；不能先删除底层存储让上层产生部分成功。

## 6. 任务与服务运行态证据

本章只保留能够改变服务依赖判断的运行态证据，不讨论具体部署对象。

### 6.1 Task topic 与消费者活性

当前任务状态确认以下 Task topic 存在，并且证据窗口内都有活跃消费者：

- `ke.doc_to_pdf_task`
- `ke.prase_pdf_task`
- `ke.knowledge_task`
- `ke.description_task`
- `ke.graph_file_task`
- `ke.success_file_task`

同时存在大量历史 heartbeat key。`corekg` 使用 `task.InitTask(false)`，未启用 heartbeat 清理，因此不能把 key 总数当作活跃 worker 数。

### 6.2 CoreTask 状态

任务统计显示 description/parse 存在耗尽重试的 fail，并有后续 pending；按领取条件没有可立即执行的 pending。这说明部分 pending 被前置失败 step 阻塞，不等于消费者停止工作。

删除一个 worker 后，后续 Task 会形成“记录仍在但 DAG 永远不推进”的业务阻塞；服务可用性必须用端到端任务验证，不能只检查进程存活。

### 6.3 当前配置选择

- CoreKG 当前使用外部 Rerank endpoint，而不是本地 Rerank 实现。关闭或替换 Rerank 应以 CoreSetting 和实际调用方为准。
- LLM 与 Embedding 是外部 endpoint，不要求与 CoreKG 同环境部署，但仍是产品功能依赖。
- OAuth/Connector provider 当前配置为空；这证明当前没有配置实例，不证明相关初始化代码可直接删除。
- OpenCoze 使用独立的业务组件，同时共享或依赖 MySQL、Redis、ES、MinIO 等数据能力。

## 7. 已知缺陷与风险

### 7.1 高优先级

1. **`keinit` 敏感信息日志。** `apps/keinit/cmd/main.go:39-55` 和 `apps/keinit/cmd/chatmodel.go:39-45` 会输出完整配置/环境集合。应改为字段白名单/脱敏并轮换已暴露凭据。
2. **Redis 假降级。** 初始化失败只记录日志，但后续 Task、SSE 等仍直接依赖 Redis，可能报错或 panic。
3. **ChatModel/OpenCoze 双写非原子。** Chat DB 与 Coze DB 没有统一事务，可能形成部分成功。
4. **MCP callback 地址缺口。** 聚合 `corekg` 未调用 `InitMCPCfg`，默认 callback 为 `localhost:8086`。
5. **MR !428 附件伪异步。** 无 worker 时上传同步等待最多 10 分钟；HTTP 超时不取消后台 Task。
6. **`keinit` 初始化 fail-open。** MinIO bucket 重试耗尽后仍可启动状态服务。
7. **支付 Router 与初始化脱节。** 聚合 `corekg` 暴露 sale API，却没有执行 `kesale.Init`。
8. **Excel 新上传伪成功。** 当前回调可能未实际分析/建表就把多个处理状态置 success。
9. **QA 导入格式契约不一致。** XLSX/CSV/XLS 的文案、前端校验和后端解析不一致。
10. **词云 mock 掩盖 Graph 故障。** API 空结果或异常时前端显示硬编码内容。
11. **同义词/行业术语增加 Chat 热路径调用。** 每个 Forest Chat 最多额外触发两次 LLM 关键词请求，且没有 Feature Flag。

### 7.2 Task 一致性风险

- Redis Stream `NoAck` + MySQL claim + 周期补推是最终对账，不是可靠消息确认。
- Task 先写 MySQL 再推 Redis；push 失败会留下 pending Task 并立即让请求失败。
- callback 无本地持久重试，running Task 可能长时间等待超时扫描。
- success 后 next-step push 失败不回滚。
- parser/`split` 部分调用只以 HTTP 成功为依据，可能出现 Task success 但产物不完整。
- MR !428 的 `checkOutput` 会拒绝非 2xx，但只读取 1 byte；空 2xx body 的 `io.EOF` 仍会被当作成功。
- `RetryParse` 只重置 DB，不立即推送队列，依赖周期对账。
- `InitTaskDBStauts` 有实现但没有调用，重启后的 running Task 恢复慢。

### 7.3 配置与边界风险

- 多数 CoreSetting 被加载为进程级单例；修改数据库不等于在线生效。
- `ENABLE_NEBULA_GRAPH` 不是完整 Feature Flag。
- Connector 配置为空不代表 Connector 代码可直接删除。
- 无源码算法服务只能确认调用契约和当前行为，不能保证内部处理、容量和数据保留语义。
- OpenCoze 与 CoreKG 共用数据服务时，裁撤 OpenCoze 不能直接删除共享 MySQL、Redis、ES 或 MinIO。

## 8. 建议的服务解耦与裁撤顺序

1. 引入统一 Feature Matrix，让 startup init、Router、前端菜单、Task DAG 和健康语义受同一开关控制。
2. 修复敏感日志、Redis 假降级、MCP callback 和 ChatModel/OpenCoze 双写。
3. 把 Chat 附件解析改为真正异步：上传立即返回 Task ID，状态查询/回调独立，支持取消或幂等重试。
4. 优先裁撤不持有核心数据的静态帮助、可选工具和明确有 fallback 的能力。
5. 再裁撤 Sandbox、Chart MCP、Rerank、Connector 等功能边界较清晰的服务。
6. 按产品决定裁撤 OpenCoze、Graph、AI 写作、Keapi MCP/REST；先解除代码和数据耦合，再去底层依赖。
7. 最后才考虑 Account、MySQL、Redis、MinIO、ES 或主知识 worker；这些属于重构或数据迁移项目。

## 9. 服务删除前验收清单

任何服务删除都至少要回答：

- 是否有前端路由、菜单、iframe、SDK 或外部 API 客户？
- 是否有后端 Router、startup init、job/listener、defer side effect？
- 是否被 CoreSetting、环境变量、数据库行或动态模型配置引用？
- 是否有 Redis topic、MySQL pending/running Task 或周期 job？
- 是否直接或间接写 MySQL、Redis、MinIO、ES、Nebula、Milvus 或其他状态？
- 是否有 callback、webhook、MCP、OpenAI-compatible 或仓库外调用方？
- 依赖失败时能否稳定降级，还是会超时、panic、部分成功或阻塞 DAG？
- 替代实现是否保持数据、协议、幂等、超时和错误语义？
- 是否完成存量数据迁移、双读/回滚和一致性校验？
- 是否完成敏感配置清理和凭据轮换？

只有这些问题都有证据，才能把“未发现引用”升级为“可安全删除”。

## 10. 证据索引

后端关键文件：

- `apps/corekg/app.go`：聚合 app Router。
- `apps/corekg/cmd/main.go:60-184`：启动初始化和主服务入口。
- `apps/corekg/cmd/init.go:23-68`：MySQL、Redis、Task 初始化。
- `apps/kecore/models/coretask/generate_task.go:25-392`：知识文件 Task DAG。
- `apps/corekg/internal/jobs/task_finished_parse_file.go`：`ke.success_file_task` 的进程内消费者。
- `apps/kecore/services/svcforest/forest_api.go`、`svcforestfile/forest_file_upload.go`：Excel Forest 和上传双轨。
- `apps/kecore/services/svcdbforest`、`apps/kechat/models/qachat/mysql_chat.go`：DB Knowledge。
- `apps/kecore/internal/apis/qapair`：QA/FQA 的 ES/Embedding 链。
- `apps/kecore/services/devkeywords`、`apps/kechat/chat/modes/forest.go`：同义词/行业术语 Chat 热路径。
- `pkgs/task/task_queue.go`、`pkgs/task/crud.go`、`pkgs/task/task_server.go`：Task queue、claim、callback/next-step。
- `apps/keworker/cmd/worker_doc_to_pdf.go`、`worker_pdf_extract.go`、`worker_split_text_chunk.go`、`worker_description.go`：worker 契约。
- `apps/kechat/internal/apis/file_api.go`、`apps/kechat/services/svcfile/attachment_parse.go`：MR !428 附件链。
- `apps/kechat/services/svcmodel/llm_model.go`：ChatModel/OpenCoze 同步。
- `apps/keapi/internal/mcpcommon/common.go`、`apps/keapi/conf/config.go`：MCP callback。
- `apps/keinit/cmd/main.go`、`apps/keinit/cmd/chatmodel.go`：初始化和敏感日志风险。

前端关键文件：

- `roc-web/ai/src/router/index.tsx`：产品路由。
- `roc-web/ai/src/api`：Account、Knowledge、Graph、Project、Payment 等 API 边界。
- `roc-web/ai/src/components/dialog`：Chat、SSE、Attachment。
- `roc-web/ai/src/pages/app/docs`：普通和特殊知识库入口。
- `roc-web/ai/src/pages/graph`、`write`、`project`、`app/agents`：Graph、写作、Project、Agent。
- `roc-web/ai/src/pages/settings`、`profile`、`version`：管理与辅助能力。

运行态证据：

- MySQL `core_settings`、`chat_model`、`core_task` 数据事实。
- Redis Task topic、heartbeat 和 stream 状态事实。
- 现有服务配置中的调用 endpoint 和 Task 契约；敏感字段未记录。

## 11. 尚需服务所有者确认的边界

以下事项不影响已证实主链的判断，但会影响最终服务删除或替换决策：

1. 外部 PDF parser、`/split`、Embedding、LLM、Rerank、通用 analyser 的源码、SLA、数据保留和真实业务响应协议。
2. `web-editor` 对 `html2docx`、CoreKG 和对象存储的完整内部依赖。
3. Keapi REST/MCP 的外部客户、API Key 使用和调用流量。
4. OpenCoze、Nebula、ES、MinIO、Milvus 现有数据的保留期限和迁移责任人。
5. RabbitMQ `rpcqueue`、`insert_index` 等仓库遗留路径是否存在仓库外消费者。
6. 共享 MySQL、Redis、ES、MinIO 中 CoreKG 与 OpenCoze 数据的隔离规则和删除边界。
7. `rerank-emb` 是否有 CoreKG 之外的客户端。

在 Owner、调用方和数据证据补齐前，相应能力只能标记为“候选”，不能标记为“已确认可直接删除”。
