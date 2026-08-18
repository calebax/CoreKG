# Task Worker NATS Request/Reply RPC 重构设计

## 1. 目标

将 `clients/task_worker`（Go，HTTP 轮询模式）中的 8 个任务迁移到 `apps/worker`（TypeScript pnpm monorepo），采用：

- **Worker 模式**：NATS core Request/Reply RPC 消费（持久订阅）
- **CLI 模式**：纯本地执行（复用现有 chunk/extract 模式，不走 NATS）
- **多实例安全**：Queue Group 保证同一请求只被一个 worker 实例处理

## 2. NATS Subject 体系

每个任务一个独立的 core NATS Request/Reply subject：

| 任务 | Subject | 核心依赖 |
|------|---------|---------|
| analysis | `core.task.rpc.analysis` | S3, LLM |
| copy | `core.task.rpc.copy` | S3 |
| desc | `core.task.rpc.desc` | S3, LLM, ES, Embedding |
| mindmap | `core.task.rpc.mindmap` | S3, LLM |
| pdf_extract | `core.task.rpc.pdf_extract` | S3, Daemon HTTP |
| video_extract | `core.task.rpc.video_extract` | S3, Daemon HTTP |
| split_text_chunk | `core.task.rpc.split_text_chunk` | ES, Algo HTTP |
| insert_index | `core.task.rpc.insert_index` | ES, Algo HTTP |

不使用 JetStream，使用 core NATS `subscribe()`/`request()`。

## 3. 新增包

### 3.1 `@corekg/workers` (`packages/workers/`)

纯业务逻辑包，无 NATS 依赖。

**文件结构：**

```
packages/workers/
├── package.json
├── tsconfig.json
└── src/
    ├── index.ts
    ├── types.ts
    ├── agent-client.ts
    ├── markdown-utils.ts
    ├── daemon-client.ts
    ├── algo-client.ts
    ├── handlers/
    │   ├── analysis.ts
    │   ├── copy.ts
    │   ├── desc.ts
    │   ├── mindmap.ts
    │   ├── pdf-extract.ts
    │   ├── video-extract.ts
    │   ├── split-text-chunk.ts
    │   └── insert-index.ts
    └── __tests__/
        └── ...
```

**核心接口：**

```ts
export interface TaskContext {
  s3: S3Provider;
  es: ESProvider;
  llm: ChatFn;
  embed: EmbedFn;
  agentClient: AgentClient;
  daemonUrl: string;
  algoUrl: string;
  logger: Logger;
}

export interface TaskHandler {
  name: string;
  subject: string;
  run(ctx: TaskContext, payload: TaskPayload): Promise<TaskResult>;
}

export interface TaskResult {
  status: "success" | "fail";
  result?: unknown;
  error?: string;
}
```

**各 handler 业务逻辑（从 Go task_worker 迁移）：**

| Handler | 输入 | 处理流程 | 输出 |
|---------|------|---------|------|
| analysis | FileURL, UploadPath, Bucket | 下载文件 → LLM agent → 上传 S3 | UploadPath |
| copy | FileURL, UploadPath, Bucket | 下载文件 → 上传 S3 (path+content.md) | UploadPath |
| desc | FileURL, FileID, ForestID, Uin, CompanyID, ESIndex | 删除旧 ES → 并行(mindmap+abstract+shortdesc+embed) → 写 ES FileDescription | desc\nmindmap\nabstract |
| mindmap | FileURL, UploadPath, Bucket | 提取标题 → LLM → UUID嵌入 → 上传 S3 | UploadPath |
| pdf_extract | FileURL, UploadPath, Bucket | 下载文件 → HTTP POST daemon → 上传目录到 S3 | S3文件列表(max 10) |
| video_extract | FileURL, UploadPath, Bucket | 下载视频 → HTTP POST daemon → 上传目录到 S3 | S3文件列表 |
| split_text_chunk | FileURL, Uin, CompanyID, ForestID, FileID, ESIndex, FileExt | 清理 ES chunks → HTTP POST algo/split → 查询 chunk IDs | algo响应 |
| insert_index | FileURL, Uin, CompanyID, ForestID, FileID, ESIndex | 查询 chunk IDs → HTTP POST algo/index | null |

**NebulaGraph 逻辑完全移除**（split_text_chunk 和 insert_index 中的图数据库清理/插入）。

### 3.2 `@corekg/rpc` (`packages/rpc/`)

NATS core Request/Reply 封装。

**文件结构：**

```
packages/rpc/
├── package.json
├── tsconfig.json
└── src/
    ├── index.ts
    ├── server.ts
    ├── client.ts
    ├── subjects.ts
    └── schema.ts
```

**RPCServer：**

```ts
export class RPCServer {
  constructor(nc: NatsConnection, handlers: TaskHandler[], ctx: TaskContext);
  start(queueGroup: string): Promise<void>;
  stop(): Promise<void>;
}
```

- 为每个 handler 的 subject 调用 `nc.subscribe(subject, {queue: queueGroup}, callback)`
- callback 内：JSON.parse → handler.run() → respond(JSON.stringify(RPCResponse))
- 异常时 respond({status: "fail", error: message})

**RPCClient：**

```ts
export class RPCClient {
  constructor(nc: NatsConnection);
  request(subject: string, payload: unknown, timeoutMs?: number): Promise<TaskResult>;
}
```

**RPCResponse Schema：**

```ts
z.object({
  status: z.enum(["success", "fail"]),
  result: z.unknown().optional(),
  error: z.string().optional(),
});
```

## 4. Worker 应用变更

### 4.1 新增 CLI 命令

8 个新 CLI 命令，全部纯本地执行（不走 NATS）：

| 命令 | 选项 | 说明 |
|------|------|------|
| `kealgo analysis` | `--file <path>`, `--upload-path <path>`, `--bucket <bucket>` | 本地 LLM analysis |
| `kealgo copy` | `--file-url <url>`, `--upload-path <path>`, `--bucket <bucket>` | 本地文件拷贝 |
| `kealgo desc` | `--file <path>`, `--file-id <id>`, `--forest-id <id>` | 生成本地描述 |
| `kealgo mindmap` | `--file <path>`, `--upload-path <path>`, `--bucket <bucket>` | 本地思维导图 |
| `kealgo pdf-extract` | `--file <path>`, `--daemon-url <url>`, `--upload-path <path>` | 本地 PDF 提取 |
| `kealgo video-extract` | `--file <path>`, `--daemon-url <url>`, `--upload-path <path>` | 本地视频提取 |
| `kealgo split` | `--file-url <url>`, `--algo-url <url>`, `--es-index <idx>` | 本地分片 |
| `kealgo index` | `--file-url <url>`, `--algo-url <url>`, `--es-index <idx>` | 本地索引 |

### 4.2 Worker 模式变更

在现有 JetStream chunker consumer 旁启动 RPCServer：

1. 初始化 providers (S3, ES, LLM, Embedding)
2. 初始化 AgentClient, DaemonClient, AlgoClient
3. 注册 8 个 TaskHandler
4. 启动 RPCServer(nc, handlers, ctx) with queue group "core-task-workers"
5. 启动现有 JetStream chunker consumer (保持不变)
6. 等待 SIGTERM/SIGINT → stop RPCServer + consumer → drain nc

## 5. 配置变更

在 `packages/config/src/schema.ts` 的 `AppConfigSchema` 中新增 3 个 section：

```ts
export const AgentConfigSchema = z.object({
  apiUrl: z.string().url(),
  apiKey: z.string(),
  chunkSize: z.number().int().positive().default(60000),
  maxTokenSize: z.number().int().positive().default(120000),
  maxWorkers: z.number().int().positive().default(50),
});

export const DaemonConfigSchema = z.object({
  url: z.string().url().default("http://localhost:5000/local.Run"),
});

export const RPCConfigSchema = z.object({
  queueGroup: z.string().default("core-task-workers"),
  timeoutMs: z.number().int().positive().default(300000),
});
```

`.env.example` 新增：

```
# Agent
AGENT_API_URL=https://yygu.cn/v3/llm.chat
AGENT_API_KEY=changeme
AGENT_CHUNK_SIZE=60000
AGENT_MAX_TOKEN_SIZE=120000
AGENT_MAX_WORKERS=50

# Daemon
DAEMON_URL=http://localhost:5000/local.Run

# RPC
RPC_QUEUE_GROUP=core-task-workers
RPC_TIMEOUT_MS=300000
```

## 6. 多实例安全

- 使用 NATS Queue Group：`nc.subscribe(subject, {queue: "core-task-workers"})`
- 同一 queue group 内 NATS 保证同一请求只被一个实例收到
- 不做额外并发上限控制
- 请求超时由调用方设置

## 7. 不在范围内

- NebulaGraph 操作（完全移除）
- JetStream work queue 模式（仅用 core NATS Request/Reply）
- 健康检查 / 心跳
- 异步回调发布（RPC 模式下响应直接返回给调用方）
