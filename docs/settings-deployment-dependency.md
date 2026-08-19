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

`corekg/agentenv` 配置样例：

```yaml
sandbox:
  mode: local_command        # 或 remote_http
  timeout: 120
  http_base_url: "http://change-me/run"
  http_token: ""
mcp:
  chart:
    mode: streamable | sse
    url: "http://change-me/mcp"
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
| `account/aeskey`、`employee_jwt`、`wechat_web_oauth` | 账号/加密/JWT | account |
| `chat/llm`、`knownow/llm_image_parse`、`knowledge/llm_image_parse` | LLM/多模态 | 各链 |

涵盖的完整配置样例以 `apps/keinit/conf/test/core_setting.yaml.example` 为准。

## 6. keinit 种子完整性与缺口

keinit 的种子来源有两处：`core_setting.yaml(.example)` 与
`scripts/mysql/v1.0_3__insert_setting.sql`（历史 SQL）。两者并集后，仍有以下
代码会读取但**未在种子中**的 Key（按部署影响分两类）：

### 6.1 部署必需，建议补进种子

| group/key | 缺失影响 |
|---|---|
| `knowledge/system_llm_api_key` | 内部 agent/chat 无法拿到系统 LLM key |
| `knowledge/mysql_excel_instance_readonly` | excel 问答只读连接缺失 |
| `knowledge/rerank`、`reranksearchcfg`、`graphsearchcfg` | 重排/图谱检索不生效 |
| `knowledge/convert_to_pdf` | 部分文件转换路径配置缺失 |
| `knowledge/deploy` | 部署模式判断默认空值 |
| `knowledge/mysql_corn_expression`、`notify_expiring_quotas_corn_expression` | 定时任务表达式默认缺失 |
| `corekg/agentenv` | ReactAgent 的 sandbox / **mcp.chart** 端点未配置 |
| `corekg/baidu_bce_api_key` | 搜索 Tool 无 key |

### 6.2 建议预留（多为运行期由运营/接口写入）

`corekg/raw_license`、`corekg/official_*wechat_webhook_url`、
`corekg/yg_api_analysis_file`、`corekg/core_file_convert`、
`corekg/llm_role_name`、`core/website-info`、`core/proxy_url`、
`account/aeskey`、`account/employee_jwt`、`account/wechat_web_oauth`。

> 结论：keinit 的 `core_setting.yaml(.example)` **不完整**——它与历史 SQL
> 并集后仍缺若干代码实际读取的 Key。其中直接卡功能的（§6.1）应补进种子；
> 运行期才下发的（§6.2）可随初始化或对应管理接口写入。
