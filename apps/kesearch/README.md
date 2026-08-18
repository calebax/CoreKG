# KESearch 服务

## 服务定位

KESearch 是 CoreKG 平台的**知识库检索与搜索服务**，是 RAG 系统的核心搜索后端。它提供知识库级别的语义搜索和全局搜索，支持文本/文档/图片/视频的多模态搜索、Chunk 增删改查、重排序搜索（Rerank），以及面向 Coze 等外部平台的搜索接口。

## 核心业务域

| 业务域 | 解决什么问题 | 关键能力 |
|--------|-------------|---------|
| 知识库搜索 | 指定知识库内的检索 | 文档/图片/视频分类搜索 |
| 全局搜索 | 跨知识库的统一检索 | 文档/图片/视频/Agent/知识库/外部数据搜索 |
| Chunk 管理 | 文本块的 CRUD | 列表查询、详情、更新、删除、禁用 |
| Rerank 搜索 | 二次排序提升精度 | 先粗排后精排的两阶段检索 |
| Coze 搜索 | 外部平台对接 | 无需登录的知识库搜索接口 |

## 核心业务概念

- **Chunk（文本块）** - 文档 RAG 解析后的最小检索单元，存储在 Elasticsearch
- **ForestSearch（知识库搜索）** - 限定在指定知识库 ID 列表内的搜索
- **GlobalSearch（全局搜索）** - 跨知识库的统一搜索，支持 Agent/知识库级别的搜索
- **RerankSearch（重排序搜索）** - 先召回后精排的两阶段管道
- **HighLight（高亮配置）** - 搜索结果中的关键词高亮

## API 路由分组

路由定义在 `internal/apis/apis.go`，大部分需要登录认证。

| 分组 | 代表接口 | 认证要求 |
|------|---------|---------|
| 知识库搜索 | ForestSearch, ForestSearchDoc/Image/Video | 需登录 |
| 全局搜索 | GlobalSearch, GlobalSearchDoc/Image/Video/Agent/Forest/ExternalData | 需登录 |
| Chunk 查询 | ListFileChunk, GetChunkByID, GetChunkBySequence, GetChunkDetail | 需登录 |
| Chunk 管理 | UpdateChunk, DeleteChunk, DisableFileChunk | 需登录 |
| Rerank | RerankSearchChunk | 需登录 |
| Coze 接口 | KnowledgeSearch | 无需登录 |
| 管理 | MigrateChunkFileName, ExcuteSql | 需登录 |
| 监控 | metrics | 无 |

## 代码架构

```
apps/kesearch/
├── app.go                   # Routers/Migrates/RunJob
├── cmd/
│   ├── main.go              # HTTP 服务 + ES/搜索/高亮/Connectors 初始化
│   └── init.go              # 数据库初始化
├── internal/
│   ├── apis/
│   │   ├── apis.go              # 路由注册
│   │   ├── globalsearchctl/     # 全局搜索 handler
│   │   ├── chunkctl/            # Chunk CRUD handler
│   │   ├── coze/                # Coze 搜索接口
│   │   └── reranksearch.go      # Rerank 搜索
│   ├── dto/                     # 请求/响应 DTO
│   ├── docs/                    # Swagger 自动生成
│   └── migrate/                 # 数据库迁移
├── services/
│   ├── svcchunk/               # Chunk 业务逻辑
│   ├── svcessearch/            # ES 搜索业务逻辑
│   └── svcreranksearch/        # Rerank 搜索逻辑
├── models/
│   ├── chunk/                  # Chunk 数据模型
│   ├── essearch/               # ES 搜索封装（被多个服务引用）
│   └── globalsearch/           # 全局搜索模型
└── pkg/                        # 内部工具包
```

## 技术要点

- **多 ES 客户端**：启动时初始化 essearch、chunk、highlight、chatquestion history 四个 ES 客户端
- **多模态搜索**：支持文本、文档、图片、视频四种搜索结果类型
- **Rerank 管线**：先粗排召回候选，再精排重排序提升精度
- **被广泛引用**：`models/essearch` 被 kecore、kechat、keworker 等多个服务导入
- **可被 corekg 内嵌**：实现标准 Routers/Migrates/RunJob 接口

## 本地开发

```bash
make local APP=kesearch
make run APP=kesearch ENV=test
make generate-docs APP=kesearch
```

## 与其他服务的关系

- 依赖 `apps/account/accountmds` - 登录认证中间件
- 依赖 `apps/kechat/models/chatquestion` - 聊天历史 ES 客户端
- 依赖 `pkgs/connectors` - 外部数据源连接器初始化
- 被 `apps/corekg` 聚合单体挂载
- `models/essearch` 被 `apps/kecore`、`apps/kechat`、`apps/keworker` 引用
