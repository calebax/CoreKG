# KECore 服务

## 服务定位

KECore 是 CoreKG 平台的**核心知识管理服务**，是整个知识库平台的中枢。它负责知识库（Forest）和文档的全生命周期管理、知识图谱、写作空间、权限控制、配额管理、搜索、标签分类、项目管理、消息通知和支付订单。

## 核心业务域

| 业务域 | 解决什么问题 | 关键能力 |
|--------|-------------|---------|
| 知识库管理 | 知识库生命周期 | CRUD、重命名、描述更新、启用/禁用 |
| 文档管理 | 文档上传与组织 | 预签名上传、分片回调、目录树、预览、移动、秒传 |
| 文档解析/AI分析 | 文档内容提取与智能分析 | 异步解析管线、思维导图、摘要、推荐问题 |
| 知识库问答 | 基于知识库的对话 | 单文档问答、多轮会话问答 |
| 知识图谱 | 结构化知识表示 | 图谱 Schema、节点/边 CRUD、图谱构建任务、RAG 图谱搜索 |
| 写作空间 | AI 辅助文章写作 | 文章 CRUD、AI 写作命令、模板 |
| 搜索 | 知识检索 | 知识库范围搜索、全局搜索、文档/图片/视频分类 |
| 权限控制 | 资源访问管理 | 资源级 RBAC（公开/公司/私有/自定义）、管理员、禁用列表 |
| 标签/分类 | 知识组织 | 标签组、标签树、资源标签关联、同义词/主关键词 |
| QA 对管理 | 手动维护问答 | CRUD、批量导入、提交到 ES |
| 数据库知识库 | MySQL/Excel 数据作为知识源 | DB 实例连接、表元数据、Excel Sheet |
| 项目管理 | 内容工作区 | 项目 CRUD、关联文章和资源 |
| 配额/订阅 | 资源用量管理 | 磁盘/QA/图谱/文章/员工配额、套餐、订单、微信支付 |
| 消息/公告 | 系统通知 | 消息中心、已读/未读、公告列表 |

## 核心业务概念

- **Forest（知识库）** - 核心实体，有类型（file/qa/cad/data）和数据源类型（standard/excel/db）
- **ForestFile（文档）** - 知识库内的文件，树形结构（ParentID/Depth），解析状态追踪
- **ResourceScope（资源权限）** - 通用权限模型，按资源类型+范围类型+动作定义
- **Graph（知识图谱）** - 关联知识库的图谱，含 Tags/Edges/Nodes，通过异步任务从文档构建
- **Article（文章）** - 写作空间内容，关联知识库，支持 AI 写作命令
- **CoreTask（异步任务）** - 文件处理管线：复制->文档转PDF->解析->索引->知识提取->图谱构建
- **QAPair（问答对）** - 手动维护的问答条目，索引到 ES
- **Quota（配额）** - 多维配额系统：磁盘/QA数/图谱数/文章数/员工数

## 异步文件处理管线

```
CopyTask -> Doc2PDFTask -> ParseTask -> [MindMapTask + AnalysisTask + DescriptionTask] -> KnowledgeTask -> [GraphFileTask] -> SuccessFileTask
```

每个步骤有优先级、超时（150分钟）和重试（3次）配置，通过 `pkgs/task` 队列系统调度。

## API 路由分组

路由定义在 `internal/apis/apis.go`，使用 `forest.` action 命名空间。

| 分组 | 代表接口 |
|------|---------|
| 知识库 CRUD | CreateForest, ListForest, GetForest, ModifyForest, DeleteForest |
| 文档管理 | ListFile, PreUploadFile, UploadFileCallBack, CreateDir, DeletePath, MovePath, PreviewFileByURL |
| 解析/AI分析 | GetContent, GetAnalysis, GetMindMap, GetFileDescription, GetRecommendQuestions |
| 单文档问答 | ListFileQA, FileChat, DeleteFileQA |
| 知识库问答 | ListSession, CreateForestSession, ForestQAChat, ListSessionQA |
| 搜索 | ForestSearch, GlobalSearch, GlobalSearchDoc/Image/Video/Agent/Forest |
| 知识图谱 | CreateGraph, CreateTag, CreateEdge, ListGraphNode, GetKnowledgeGraph, ParseGraph |
| 文章 | ListArticle, CreateArticle, EditArticle, AIWrite, ExecuteAIWriteCmd |
| 权限 | GetForestPermSet, ModifyForestPermSet, SetResourcePerm, SetResourceScope |
| QA 对 | CreateQAPair, ListQAPair, UploadQAPair, CommitQAPair |
| 数据库知识库 | CreateForestDBInstance, ListForestDB, ListForestTable |
| 项目 | CreateProject, ListProject, RenameProject |
| 配额/支付 | GetCompanyQuota, CreateOrder, QueryOrderStatus, HandleNotifyWX |
| 消息/公告 | ListMessage, SetMessageStatus, ListAnnouncement |

## 代码架构

```
apps/kecore/
├── app.go                   # Routers/Migrates/RunJob
├── cmd/
│   ├── main.go              # Cobra 入口，完整初始化链
│   └── init.go              # 数据库初始化
├── internal/
│   ├── apis/                # Handler 层（按域分子目录）
│   │   ├── apis.go          # 统一路由注册（301行）
│   │   ├── forestctl/       # 知识库/文档 handler
│   │   ├── fileqactl/       # 单文档问答
│   │   ├── forestqactl/     # 知识库问答
│   │   ├── globalsearchctl/ # 搜索
│   │   ├── graphctl/        # 知识图谱
│   │   ├── accountctl/      # 权限/配额
│   │   ├── articlectl/      # 文章 AI 写作
│   │   └── ...
│   ├── dto/                 # 请求/响应 DTO
│   ├── docs/                # Swagger 自动生成
│   └── migrate/             # 数据库迁移
├── services/                # 业务逻辑层（33个包）
│   ├── svcforest/           # 知识库 CRUD
│   ├── svcforestfile/       # 文档操作
│   ├── svcgraph/            # 图谱管理
│   ├── svcarticle/          # 文章 CRUD
│   ├── svcperm/             # 权限管理
│   ├── membership/          # 配额管理器
│   ├── svchotwords/         # 热词生成
│   ├── graphragsearch/      # 图谱 RAG 搜索
│   └── ...
├── models/                  # 数据层（17个包）
│   ├── foresttype/          # 核心类型定义
│   ├── forest/              # DAO
│   ├── graph/               # 图谱 CRUD
│   ├── coretask/            # 异步任务管线
│   ├── nbgraph/             # NebulaGraph 集成
│   ├── perm/                # 权限模型
│   ├── fs/                  # 文件存储（COS/S3/MinIO）
│   └── ...
├── mds/
│   └── coremds.go           # 配额中间件（Disk/QA/Graph/Article）
├── jobs/
│   └── job.go               # 定时任务
└── conf/test/               # 环境配置
```

## 定时任务

| 任务 | 调度 | 说明 |
|------|------|------|
| SyncMysqlTable | 每日 00:00 | 同步数据库知识库的 MySQL 表元数据 |
| GenerateUsersHotWords | 每日 00:00 | 基于用户搜索模式自动生成热词 |
| PackageQuotaExpireNotify | 每日 09:00 | 通知即将过期的订阅套餐 |

## 技术要点

- **配额中间件**：所有配额检查作为 Gin 中间件在 handler 前执行，私有化部署跳过
- **权限中间件**：读取请求体提取资源 ID，验证用户访问权限后恢复 body
- **资源启用/禁用**：支持7种资源类型，Redis 分布式锁防并发
- **NebulaGraph**：两个 NebulaGraph 客户端实例（图谱存储 + RAG 搜索）
- **多库**：Knownow（知识库）、Core（核心）、Chat（对话）、Sale（销售）

## 本地开发

```bash
make local APP=kecore
make run APP=kecore ENV=test
make generate-docs APP=kecore
```

## 与其他服务的关系

- 依赖 `apps/account` - 认证中间件、公司信息（配额）
- 依赖 `apps/kechat` - QA 配额中间件、聊天历史 ES 客户端、模型选择
- 依赖 `apps/keparser` - 解析任务类型常量
- 依赖 `apps/kesearch` - ES 客户端初始化
- 依赖 `apps/kesale` - 微信支付回调、销售基础设施
- 依赖 `pkgs/task` - 异步任务队列
- 被 `apps/corekg` 聚合单体挂载
