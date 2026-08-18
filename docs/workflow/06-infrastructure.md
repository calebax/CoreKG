# 基础设施层

## 1. 概览

基础设施层位于 `apps/workflow/infra/`，为上层业务提供通用技术支撑。采用接口抽象 + 多实现模式，支持灵活切换底层组件。

```
infra/
├── cache/              # Redis 缓存
├── checkpoint/         # 工作流检查点
├── coderunner/         # 代码执行器
├── document/           # 文档处理管线
├── dynconf/            # 动态配置
├── embedding/          # 向量嵌入
├── es/                 # Elasticsearch
├── eventbus/           # 事件总线
├── idgen/              # ID 生成器
├── imagex/             # 图片服务
├── orm/                # MySQL ORM
├── rdb/                # 关系型数据库抽象
├── sqlparser/          # SQL 解析器
├── sse/                # Server-Sent Events
└── storage/            # 对象存储
```

## 2. 缓存（Redis）

Path: `infra/cache/`

- 客户端：标准 Redis 客户端
- 用途：
  - 工作流执行状态缓存
  - 流式输出 chunk 缓存
  - 取消信号标志
  - ID 生成器底层
  - 节点图标 URL 缓存
  - 测试运行最后执行 ID

## 3. 检查点（Checkpoint）

Path: `infra/checkpoint/`

工作流中断/恢复的核心存储，支持两种实现：

### 3.1 内存实现
- 单实例，调试用
- 不持久化，重启丢失

### 3.2 Redis 实现
- 生产环境使用
- 支持多实例共享
- 存储内容：
  - 全局状态（GlobalState）
  - 节点局部状态（NodeState）
  - 中断事件列表（InterruptEvents）
  - 复合节点迭代索引

## 4. 代码执行器（CodeRunner）

Path: `infra/coderunner/`

三种执行模式：

### 4.1 直接执行
- 在当前进程中执行代码
- 快速但隔离性差

### 4.2 沙箱执行
- 隔离环境执行
- 安全但需要沙箱服务

### 4.3 Python 脚本
- 通过 Python 解释器执行
- 支持 Python 生态库

## 5. 文档处理管线

Path: `infra/document/`

完整的文档处理流水线，用于知识库：

```
文档上传 → 格式解析 → OCR（图片） → 分块 → Embedding → ES 索引
```

### 5.1 文档解析器（parser/）

| 解析器 | 支持格式 |
|--------|----------|
| Builtin | CSV, JSON, Markdown, PDF, DOCX, XLSX, Text, Image |
| PPStructure | 表格/版面分析 |

### 5.2 OCR

| 实现 | 说明 |
|------|------|
| PPOCR | PaddleOCR 本地模型 |
| VEOCR | 云端 OCR 服务 |

### 5.3 Messages2Query

将对话消息转换为知识库检索查询。

### 5.4 NL2SQL

自然语言转 SQL，用于 Database 节点。

### 5.5 Rerank

RRF（Reciprocal Rank Fusion）重排序，用于知识库检索结果排序。

### 5.6 SearchStore

Elasticsearch 搜索存储，ES8 客户端。

### 5.7 进度条

文档处理进度追踪与查询。

## 6. 向量嵌入（Embedding）

Path: `infra/embedding/`

支持多种嵌入服务提供商：

| 提供商 | 实现 |
|--------|------|
| ARK | 字节跳动火山引擎 ARK |
| HTTP | 通用 HTTP API |
| OpenAI | OpenAI Embeddings API |
| Ollama | 本地 Ollama 模型 |
| Gemini | Google Gemini Embeddings |

每种实现遵循统一的 Embedding 接口，支持批量嵌入。

## 7. Elasticsearch

Path: `infra/es/`

- ES8 客户端
- 索引命名：`ke_{company_id}`
- 分词器：`ik_max_word`
- 用途：
  - 知识库文档检索
  - 资源搜索索引
  - 项目搜索

## 8. 事件总线（EventBus）

Path: `infra/eventbus/`

支持 5 种消息队列后端：

| 后端 | 说明 |
|------|------|
| NATS | 轻量级消息系统 |
| Pulsar | Apache Pulsar |
| RocketMQ | 阿里 RocketMQ |
| NSQ | 分布式消息平台 |
| Kafka | Apache Kafka |

用途：
- 资源事件发布（Created/Updated/Deleted → 搜索索引）
- 跨服务异步通知

## 9. ID 生成器

Path: `infra/idgen/`

基于 Redis 的分布式 ID 生成器，用于：
- 工作流 ID
- 执行 ID
- 节点执行 ID
- 各种实体 ID

## 10. SQL 解析器

Path: `infra/sqlparser/`

解析和验证自定义 SQL 语句，确保 DatabaseCustomSQL 节点的安全性：
- SQL 语法验证
- 注入风险检测
- 权限范围控制

## 11. Server-Sent Events（SSE）

Path: `infra/sse/`

封装 SSE 协议，用于：
- 工作流流式执行输出
- ChatFlow 实时对话
- 节点执行状态推送

## 12. 对象存储

Path: `infra/storage/`

支持三种存储后端：

| 后端 | 说明 |
|------|------|
| MinIO | 自建对象存储 |
| S3 | AWS S3 / 兼容 S3 协议存储 |
| TOS | 腾讯云对象存储 |

用途：
- 文件上传存储
- 文档原始文件
- 工作流图片资源

## 13. 动态配置

Path: `infra/dynconf/`

从静态 JSON/YAML 文件加载配置：
- 模型模板配置（`conf/model/template/`，12+ YAML）
- 插件产品定义（`conf/plugin/pluginproduct/`，16+ YAML）
- 工作流节点配置（`conf/workflow/config.yaml`）

## 14. LLM 模型构建器

Path: `bizpkg/llm/modelbuilder/`

7 个 LLM 提供商构建器 + 1 个内置查找：

| 文件 | 提供商 |
|------|--------|
| openai.go | OpenAI API |
| claude.go | Anthropic Claude |
| gemini.go | Google Gemini |
| qwen.go | 阿里通义千问 |
| deepseek.go | DeepSeek |
| ark.go | 字节跳动 ARK（豆包） |
| ollama.go | Ollama 本地模型 |
| builtin.go | 从配置文件查找内置模型 |

模型模板配置位于 `conf/model/template/`：
- ARK Doubao（7 个变体）
- OpenAI GPT 系列
- Claude 系列
- Gemini 系列
- Qwen 系列
- DeepSeek 系列
- Ollama
- BytePlus Seed

## 15. 文档处理流程详解

### 15.1 知识库文档索引

```
1. 用户上传文档 → files 表 / 对象存储
2. 格式检测 → 选择对应 Parser
3. Parser 解析 → 提取文本 + 元数据
4. OCR（如果是图片/PDF扫描件） → 文字识别
5. 分块（Chunking） → 切分为固定大小片段
6. Embedding → 向量化
7. 写入 Elasticsearch 索引
8. 更新进度 → progressbar
9. 更新文档状态 → knowledge_document 表
```

### 15.2 知识库检索

```
1. 收到检索请求
2. Messages2Query → 将对话消息转为检索查询
3. Query → Elasticsearch 搜索
4. Rerank (RRF) → 结果重排序
5. 返回 top-K 结果
```

### 15.3 支持的文件格式

| 格式 | Parser |
|------|--------|
| CSV | Builtin |
| JSON | Builtin |
| Markdown | Builtin |
| PDF | Builtin + OCR |
| DOCX | Builtin |
| XLSX | Builtin |
| TXT | Builtin |
| Image | Builtin + OCR |
| PPTX | Builtin |

## 16. 配置结构

Path: `conf/config.go`

```go
type AppConfig struct {
    ygconfig.CoreConfig     // 远程配置（MySQL DSN, Redis, ES 等）
    Workflow WorkflowConfig
}

type WorkflowConfig struct {
    LogLevel          string
    MaxRequestBodySize int
    ServerHost        string
    AdminUins         string     // 管理后台 UIN 白名单
    SSL               SSLConfig
    Redis             RedisConfig
    Elasticsearch     ESConfig
    Storage           StorageConfig  // type: minio/tos/s3
    MQ                MQConfig       // type: rmq/pulsar/nats
    Upload            UploadConfig
}
```

## 17. 插件产品定义

Path: `conf/plugin/pluginproduct/`

16 个内置插件定义：

| 插件 | 说明 |
|------|------|
| Lark Sheet | 飞书电子表格 |
| Lark Docx | 飞书文档 |
| Lark Base | 飞书多维表格 |
| Lark Wiki | 飞书知识库 |
| Lark Task | 飞书任务 |
| Lark Message | 飞书消息 |
| Lark Calendar | 飞书日历 |
| Lark Auth | 飞书认证 |
| Gaode Maps | 高德地图 |
| Bocha Search | 博查搜索 |
| Wolfram Alpha | 数学计算引擎 |
| 其他 | 更多第三方服务集成 |
