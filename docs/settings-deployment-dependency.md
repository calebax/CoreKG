# settings 包部署依赖说明

> 目的：说明 `vendor/github.com/ygpkg/yg-go/settings` 是什么、哪些应用依赖它、
> 部署需要满足哪些前置，以及 MCP / Sandbox / Excel 问答等特殊配置项在哪。
> 不记录任何真实密钥/Token，占位符统一用 `change-me`。

## 1. settings 包是什么（部署依赖的根因）

`settings` 是一个**运行时配置读写组件**：

- 落库：`dbtools.Core()` 指向的**核心 MySQL**，表名 `core_settings`
  （`settings/setting.go` `TableNameSettings`）。
- 缓存：**Redis**（`cache.Std()`，Key 形如 `core_setting::<group>::<key>`）。
- 读取顺序：先 Redis，miss 再 DB，命中后写回并设置 5 分钟 TTL。
- 子包 `settings/remote`（供框架 `config` 使用）从远端
  `yygu.cn/v2/cook.GetSettingContent` 拉取配置，由环境变量
  `YGCFG_AK / YGCFG_SK / YGCFG_GROUP` 控制，通常在生产链路使用。

**部署含义**：凡是依赖 settings 的服务，镜像都必须能连 **核心库（`core_settings` 表）+ Redis**。

## 2. 谁依赖 settings

### 2.1 直接使用（按应用）

| 应用 | 用途 | 写入/只读 |
|---|---|---|
| `corekg`（聚合单体） | 汇总 kechat+kecore+account+kesearch；自身写 `corekg/raw_license`、`core/website-info` | 读写 |
| `kechat` | excel 问答两条链（见 §4）、coze 地址、ES/LLM key | 只读 |
| `kecore` | excel 解析、森林文件存储、ES/nebula/embedding/coze 配置 | 只读 |
| `kesearch` | ES 索引、embedding、rerank（全部从 settings 读） | 只读 |
| `account` | `website-info`、JWT、aeskey、登录限流、存储 | 读写 |
| `keinit` | 初始化/迁移 `core_settings`、读取存储配置 | 读写（种子） |
| `workflow` | 封装读取 `corekg/corekg_url`、`corekg/coze_url` | 只读 |
| `keapi`（MCP Server） | 二进制经 yg-go 框架 link settings；21 个 Tool 为转发，不直接读 | 间接 |

### 2.2 必带 settings（框架/共享库间接）

所有 server 应用都会 link `ygpkg/yg-go/{config, cache, dbtools/redispool,
apis/runtime, storage}`，而这些包全部 import settings。因此下面这些即便不直接
调用也会把 settings 编进镜像：`keapp`、`ketask`、`webfetch`、`websearch`
以及 `keapi`、`corekg` 等。

其余非独立二进制：`kesale`/`kellm`/`nodes`（编译进其他二进制）、`apps/pipeline`（纯 Python，不涉及）。

## 3. 部署前置（核心清单）

部署任何核心服务前，确认：

1. **核心 MySQL** 可达，且存在 `core_settings` 表；
2. **Redis** 可达；
3. 使用 keinit 初始化系统设置（见 `docs/local-config-checklist.md` 第 3 节，
   会执行 `update-setting` / `init-mysql` 把 `core_setting.yaml` upsert 进 `core_settings`）。

可单独只更新系统设置：

```bash
./bundles/keinit -c apps/keinit/conf/test/config.yaml \
    --setting-file apps/keinit/conf/test/core_setting.yaml \
    update-setting
```

> `keinit` 读取的种子文件是 `apps/keinit/conf/test/core_setting.yaml`
> （gitignore，仅 `.example` 入库），`--setting-file` 默认 `./config/core_setting.yaml`。

## 4. Excel 问答与 MCP / Sandbox

kechat 的 excel 问答有三条落地路径，settings 配置不同：

### 4.1 ReactExcelChatMode（eino ReactAgent，含 **EChart MCP**）

`apps/kechat/chat/modes/excel.go` 注册 `ToolOptionCode/File/Chart`：

- **图表**：从 settings `corekg/agentenv` 读取 `mcp.chart.{mode,url}`，然后
  `einotools/tools/tool.go GetMcpTools` 连接外部 **MCP Server**
  （StreamableHTTP 或 SSE）拉取图表工具；`mode` 非 `sse` 即按 streamable 处理。
- **代码**：经 `sandbox` 执行（`local_command` 即地执行，或 `remote_http`
  调用独立 sandbox 服务 `POST /run`）。
- 图表输出落库到 `chat_type.ChatChart`（`saveChartFunc`）。

`corekg/agentenv` 配置样例（结构以 `apps/keinit/conf/test/core_setting.yaml.example` 为准）：

```yaml
sandbox:
  # 沙箱模式，可选值：auto, remote_http, local_command
  mode: remote_http        # 或 local_command（即地执行）
  http_base_url: https://change-me
  http_token: change-me
  timeout: 120
mcp:
  chart:
    # streamable（推荐）, sse
    mode: streamable
    url: https://change-me/mcp
```

**部署含义**：启用 Chart 工具时，需要独立部署/连通 **EChart MCP Server** 与
（如用 remote_http）**Sandbox 服务**，端点配进 `corekg/agentenv`；
否则 `mcp.chart` 未配置时图表工具不注册，excel 问答仍可运行但无图。

### 4.2 旧 Chat-Agent 链（无 MCP）

`apps/kechat/models/excelchat/qa_excel.go` + `mysql_chat.go`：用
`sys_agent_excel_question_to_sql`、`sys_agent_excel_sql_result_analysis`
等 Chat Agent 生成 SQL → 跑外部 MySQL → 转答案。无 MCP。

### 4.3 einonodes 旧节点（无 MCP）

`apps/einonodes/qachatnodes/*` 的 `NodeMysqlGenerateECharts` /
`BatchGenerateEcharts` 同样走 `sys_agent_*` Chat Agent，不走 MCP。

## 5. settings 关键 Key 全景

| group/key | 作用 | 主要消费者 |
|---|---|---|
| `core/redis`、`knowledge/redis`、`core/loc_redis` | Redis 连接 | 各服务/框架 |
| `core/cos-ke`、`core/cos-{purpose}` | 对象存储（MinIO/S3） | kecore/account |
| `knowledge/es` | ES 连接 | kesearch/kecore/kechat |
| `knowledge/embedding` | 向量模型 | kesearch/kecore |
| `knowledge/nebula`、`nebulacount` | 图数据库 | kecore |
| `knowledge/system_llm_api_key` | 系统 LLM 调用 key | chatclient |
| `knowledge/mysql_excel_instance[_readonly]` | Excel 问答所用 MySQL 实例 | kechat `excel_chat.go` |
| `knowledge/rerank`、`reranksearchcfg`、`graphsearchcfg` | 重排/图谱检索 | kesearch |
| `knowledge/highlight` | 搜索高亮 | kesearch |
| `knowledge/deploy` | 部署模式标识 | kecore |
| `corekg/corekg_url`、`coze_url` | CoreKG/Workflow 互访地址 | workflow/coze |
| `corekg/agentenv` | **sandbox + mcp.chart** 配置 | einotools/tool.go |
| `corekg/baidu_bce_api_key` | 搜索 Tool | einotools/tool.go |
| `corekg/raw_license` | License | corekg licensectl |
| `corekg/llm_role_name` | Agent/问答角色名 | kechat/einotools |
| `core/website-info` | 站点信息 | account |
| `account/aeskey` | 账号敏感字段加密 AesKey | account（开源版保留占位） |
| `account/employee_jwt`、`wechat_web_oauth`（管理员后台/企微登录） | 账号/加密/JWT | account（仅需管理后台时启用，开源版种子已移除，如需要可自行补回） |
| `chat/llm`、`knownow/llm_image_parse`、`knowledge/llm_image_parse` | LLM/多模态 | 各链 |

涵盖的完整配置样例以 `apps/keinit/conf/test/core_setting.yaml.example` 为准。

## 6. keinit 种子完整性与缺口

keinit 的种子来源有两处：`core_setting.yaml(.example)` 与
`scripts/mysql/v1.0_3__insert_setting.sql`（历史 SQL）。两者并集后，仍有以下
代码会读取但**种子中曾有缺失**的 Key（按部署影响分两类）。**以下 §6.1 / §6.2
的 Key 已全部补进 `apps/keinit/conf/test/core_setting.yaml(.example)`。**

### 6.1 部署必需（已补齐）

| group/key | 说明 |
|---|---|
| `knowledge/system_llm_api_key` | 内部 agent/chat 系统 LLM key |
| `knowledge/rerank`、`reranksearchcfg`、`graphsearchcfg` | 重排/图谱检索配置 |
| `knowledge/convert_to_pdf` | 文件转换配置（decoupler 场景） |
| `knowledge/deploy` | 部署模式标识 |
| `knowledge/mysql_corn_expression`、`notify_expiring_quotas_corn_expression` | 定时任务表达式 |
| `corekg/agentenv` | ReactAgent sandbox + **mcp.chart**（EChart MCP 端点） |
| `corekg/baidu_bce_api_key` | 搜索 Tool API Key |

### 6.2 建议预留（已补齐，多为运行期由运营/接口写入）

`corekg/raw_license`、`corekg/yg_api_analysis_file`、`corekg/core_file_convert`、
`corekg/llm_role_name`、`core/website-info`、`account/aeskey`。

> 注：开源版按运行需要，以下管理/企业级 Key **已从种子移除（不再在
> `core_setting.yaml(.example)` 中）**：`account/employee_jwt`、`account/wechat_web_oauth`、
> `corekg/official_website_wechat_webhook_url`、`corekg/official_dotpen_website_wechat_webhook_url`；
> 需要对应功能时自行补回即可。

### 6.3 ⚠️ 勿用「TRUNCATE + 仅 update-setting」重置（会丢非 yaml 设置）

`core_settings` 表有两个种子来源（上文 §6 开头）：`core_setting.yaml(.example)`
与 `scripts/mysql/v1.0_3__insert_setting.sql`。**`core_setting.yaml` 并不包含全部
框架/agent 设置**，以下 Key 只由 SQL 迁移种子化，不在 yaml 里：

| group/key | 说明 | 缺失影响 |
|---|---|---|
| `core/jwt-yygu` | 登录 JWT 签名密钥（`auth.GetJwtSetting("yygu")`） | **登录接口报 `获取登录设置失败` / code 500** |
| `chat/llm`、`knownow/llm_image_parse` | 聊天/视觉 LLM（旧 group，不同于 `knowledge/llm_image_parse`） | 对应链路缺 LLM 配置 |
| `knowledge/forest_prompt` | 知识库问答 prompt | 问答缺默认 prompt |
| `knowledge/agent_es_chat`、`agent_intention_recognition`、`agent_subquestion_chat` | es 问答/意图识别/子问题 agent | agent 链路缺配置 |
| `knowledge/nebulacount`、`preset_forest`、`system_config` | 图谱数量/预置知识库/全局配置 | 对应功能缺默认值 |

**正确做法**：不要只 `TRUNCATE` 后重跑 `update-setting`。需要重置时：
1. 重跑 keinit **完整初始化**，或
2. `TRUNCATE core_settings` 后，除 `update-setting` 外，**额外补回上述 key**
   （参考 `scripts/mysql/v1.0_3__insert_setting.sql` 里的 INSERT 值），否则
   登录等依赖框架 JWT 的链路会失效。

> 更新状态：`core_setting.yaml(.example)` 按 group 分组整理，模型类设置共用同一
> LLM 网关与 api_key（embedding / rerank 保留各自向量/重排端点），两个文件均通过
> `gopkg.in/yaml.v3` 解析校验（各 36 项、无重复 group/key）。其中地址/密钥占位为
> `change-me`/空，需按环境替换后执行 `keinit ... update-setting` 下发；运行期才下发的
> （§6.2）可保留默认值或由管理接口写入。
> 另：`knowledge/mysql_excel_instance_readonly` 与 `core/proxy_url` 已确认不再使用
> （分别对应 Excel 问答旧链路 `excel_chat.go`、已不活跃的 OAuth 代理），已从种子移除。

---

## 7. 配置主源结论：以 `core_settings` 为准

> 现状是出现了**双来源**（`core_settings` 表与各 app `config.yaml`），且二者 key 名
> 或语义重叠，易混乱。以下为对各配置域的**实际消费点核对结论**。唯一主源是
> **`core_settings` 表**（由 keinit 通过 `core_setting.yaml` 下发）。

### 7.1 各配置域「以什么为准」

| 配置域 | 唯一主源 | 说明 / config.yaml 里的对应块 |
|---|---|---|
| **Redis** | settings：`knowledge/redis`（corekg/kesearch/kechat/kecore/keapi）、`core/redis`（account/admin）、`ketask/redis`（ketask）、dev 用 `knowledge/loc_redis`/`core/loc_redis` | config.yaml 顶层 `redis:` **无 Go 消费**（yg-go `MainConfig` 只解析 `main:`/`logger:`）；只有 `workflow:` 块内嵌套 redis 属于 workflow 自身 |
| **JWT** | settings：`core/jwt-<issuer>`（框架 `auth.GetJwtSecret`）、`account/employee_jwt`（员工/管理后台登录，开源版种子未下发）、`core/jwt-yyguadmin` | config.yaml `account.jwt_secret/plt_jwt_secret/jwt_expire` **全仓库无解析**（死配置） |
| **对象存储** | settings：`core/cos-ke`、`core/cos-<purpose>`（知识库文件，kecore/account） | config.yaml `workflow.storage` 是 workflow 独立存储，作用域不同、非冲突；若共用同一 MinIO 才存在"同一凭据两处配" |
| **ES / LLM / Embedding / 图 / 转换** | settings（`knowledge/es`、`embedding`、`llm_image_parse`、`nebula`、`convert_pdf` 等） | 相关 config.yaml 块无消费 |
| **其它（wecom/notify/tencent_cloud/local_storage/zmrobot）** | — | config.yaml 顶层这些块**全仓库无 yaml tag 解析**（死配置） |

### 7.2 唯一有消费的 config.yaml 块
- `main:` — yg-go `MainConfig`（app/http/database_conns/env）。
- `logger:` — yg-go 日志。
- `workflow:` — corekg 通过 `wfconf.AppConfig` 对**同一文件二次解析**（corekg/cmd/init.go `loadWorkflowConfig`），只消费 workflow 嵌套的 redis/es/storage/mq 等。

### 7.3 清理方向（以 core_settings 为主源）
1. **删除/注明**各 app config.yaml 顶层 `redis:`、`account:`、`wecom:`、`notify:`、
   `tencent_cloud:`、`local_storage:`、`zmrobot:` 等死/冗余块；
   **保留** `main:`、`logger:`、`workflow:`。
2. Redis/JWT/存储统一只维护 settings；补充种子缺口：`ketask/redis`、dev 档 `knowledge/loc_redis`。
3. 重新初始化流程：停服务 → 删除数据库 → 重跑 keinit 全量初始化 → 验证接口。
