# Task Worker NATS RPC 迁移实施计划

> **供 agentic worker 使用：** 必需的子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务执行此计划。步骤使用复选框（`- [ ]`）语法进行跟踪。

**目标：** 将 8 个任务从 `clients/task_worker`（Go HTTP 轮询）迁移到 `apps/worker`（TypeScript），worker 模式使用 NATS core Request/Reply RPC，CLI 模式使用本地直接执行。

**架构：** 每个任务拥有独立的 NATS subject（`core.task.rpc.{name}`）。新增 `@corekg/workers` 包承载纯业务逻辑。新增 `@corekg/rpc` 包封装 NATS `subscribe()`/`request()`。Worker 模式通过队列组订阅；CLI 模式直接调用 handler。

**技术栈：** TypeScript、pnpm monorepo、NATS v3（`@nats-io/transport-node`）、Zod、Commander.js、Vercel AI SDK、AWS S3 SDK、Elasticsearch v9

## 全局约束

- 所有新包必须设置 `"type": "module"`，所有 import 语句中使用 `.js` 扩展名
- 所有包使用 vitest 进行测试，使用 `tsc --noEmit` 进行类型检查
- 配置变量统一放入 `@corekg/config` 中的 `AppConfigSchema`，不使用独立环境变量
- `@corekg/nats/src/types.ts` 中的 `TaskPayload` schema 是规范的 payload 类型
- 任何地方都不包含 NebulaGraph 逻辑
- RPC 不使用 JetStream — 仅使用核心 NATS `subscribe()`/`request()`
- 遵循现有代码风格：除非复杂逻辑不添加注释，使用 pino logger，使用 Zod 验证

---

## 任务 1：扩展 Config，新增 Agent、Daemon、RPC Schema

### 需要修改的文件

- `apps/worker/packages/config/src/schema.ts`
- `apps/worker/packages/config/src/index.ts`
- `apps/worker/.env.example`

### 步骤

- [ ] **步骤 1： 在 schema.ts 中添加 AgentConfigSchema、DaemonConfigSchema、RPCConfigSchema**

在 `apps/worker/packages/config/src/schema.ts` 中，在 `AppConfigSchema` 之前添加以下 schema：

```typescript
export const AgentConfigSchema = z.object({
  apiUrl: z.string().url(),
  apiKey: z.string(),
  chunkSize: z.number().int().positive().default(60000),
  maxTokenSize: z.number().int().positive().default(120000),
  maxWorkers: z.number().int().positive().default(50),
  pool: z.record(z.string(), z.string()).default({}),
});

export const DaemonConfigSchema = z.object({
  url: z.string().url().default("http://localhost:5000/local.Run"),
});

export const RPCConfigSchema = z.object({
  queueGroup: z.string().default("core-task-workers"),
  timeoutMs: z.number().int().positive().default(300000),
});
```

- [ ] **步骤 2： 向 AppConfigSchema 添加 agent、daemon、rpc 字段**

更新 `apps/worker/packages/config/src/schema.ts` 中的 `AppConfigSchema`，包含新字段：

```typescript
export const AppConfigSchema = z.object({
  nats: NATSConfigSchema,
  es: ESConfigSchema,
  s3: S3ConfigSchema,
  llm: LLMConfigSchema,
  vllm: VLLMConfigSchema,
  embedding: EmbeddingConfigSchema,
  concurrency: ConcurrencyConfigSchema,
  agent: AgentConfigSchema,
  daemon: DaemonConfigSchema,
  rpc: RPCConfigSchema,
  workerId: z.string().optional(),
});
```

- [ ] **步骤 3： 向 LocalConfigSchema 添加 agent 字段**

更新 `apps/worker/packages/config/src/schema.ts` 中的 `LocalConfigSchema`：

```typescript
export const LocalConfigSchema = z.object({
  llm: LLMConfigSchema.optional(),
  embedding: EmbeddingConfigSchema.optional(),
  agent: AgentConfigSchema.optional(),
});
```

- [ ] **步骤 4： 更新 index.ts 中的 loadConfig() 以读取新环境变量**

在 `apps/worker/packages/config/src/index.ts` 中，更新 `loadConfig()` 以包含 `agent`、`daemon` 和 `rpc` 部分：

```typescript
export function loadConfig(): AppConfig {
  return AppConfigSchema.parse({
    nats: { url: process.env.NATS_URL, stream: process.env.NATS_STREAM },
    es: {
      host: process.env.ES_HOST,
      username: process.env.ES_USERNAME,
      password: process.env.ES_PASSWORD,
    },
    s3: {
      endpointUrl: process.env.S3_ENDPOINT_URL,
      accessKeyId: process.env.S3_ACCESS_KEY_ID,
      secretAccessKey: process.env.S3_SECRET_ACCESS_KEY,
      region: process.env.S3_REGION,
      defaultBucket: process.env.S3_DEFAULT_BUCKET,
      publicEndpointUrl: process.env.S3_PUBLIC_ENDPOINT_URL,
    },
    llm: {
      apiKey: process.env.LLM_API_KEY,
      baseUrl: process.env.LLM_BASE_URL,
      model: process.env.LLM_MODEL,
    },
    vllm: {
      apiKey: process.env.VLLM_API_KEY || process.env.LLM_API_KEY,
      baseUrl: process.env.VLLM_BASE_URL || process.env.LLM_BASE_URL,
      model: process.env.VLLM_MODEL,
    },
    embedding: {
      apiKey: process.env.EMBEDDING_API_KEY,
      baseUrl: process.env.EMBEDDING_BASE_URL,
      model: process.env.EMBEDDING_MODEL,
    },
    concurrency: {
      embeddingWorkers: Number(process.env.EB_WORKERS) || undefined,
      llmWorkers: Number(process.env.LLM_WORKERS) || undefined,
    },
    agent: {
      apiUrl: process.env.AGENT_API_URL,
      apiKey: process.env.AGENT_API_KEY,
      chunkSize: Number(process.env.AGENT_CHUNK_SIZE) || undefined,
      maxTokenSize: Number(process.env.AGENT_MAX_TOKEN_SIZE) || undefined,
      maxWorkers: Number(process.env.AGENT_MAX_WORKERS) || undefined,
    },
    daemon: {
      url: process.env.DAEMON_URL,
    },
    rpc: {
      queueGroup: process.env.RPC_QUEUE_GROUP,
      timeoutMs: Number(process.env.RPC_TIMEOUT_MS) || undefined,
    },
    workerId: process.env.WORKER_ID,
  });
}
```

- [ ] **步骤 5： 更新 index.ts 中的 loadLocalConfig() 以读取 agent 配置**

在 `apps/worker/packages/config/src/index.ts` 中，更新 `loadLocalConfig()`：

```typescript
export function loadLocalConfig(): LocalConfig {
  const raw: Record<string, unknown> = {};

  if (process.env.LLM_API_KEY || process.env.LLM_BASE_URL) {
    raw.llm = {
      apiKey: process.env.LLM_API_KEY,
      baseUrl: process.env.LLM_BASE_URL,
      model: process.env.LLM_MODEL,
    };
  }

  if (process.env.EMBEDDING_API_KEY || process.env.EMBEDDING_BASE_URL) {
    raw.embedding = {
      apiKey: process.env.EMBEDDING_API_KEY,
      baseUrl: process.env.EMBEDDING_BASE_URL,
      model: process.env.EMBEDDING_MODEL,
    };
  }

  if (process.env.AGENT_API_URL || process.env.AGENT_API_KEY) {
    raw.agent = {
      apiUrl: process.env.AGENT_API_URL,
      apiKey: process.env.AGENT_API_KEY,
      chunkSize: Number(process.env.AGENT_CHUNK_SIZE) || undefined,
      maxTokenSize: Number(process.env.AGENT_MAX_TOKEN_SIZE) || undefined,
      maxWorkers: Number(process.env.AGENT_MAX_WORKERS) || undefined,
    };
  }

  return LocalConfigSchema.parse(raw);
}
```

- [ ] **步骤 6： 从 index.ts 导出新的 Schema**

更新 `apps/worker/packages/config/src/index.ts` 中的导出块：

```typescript
export {
  AppConfigSchema,
  NATSConfigSchema,
  ESConfigSchema,
  S3ConfigSchema,
  LLMConfigSchema,
  EmbeddingConfigSchema,
  VLLMConfigSchema,
  ConcurrencyConfigSchema,
  AgentConfigSchema,
  DaemonConfigSchema,
  RPCConfigSchema,
  LocalConfigSchema,
} from "./schema.js";
export type { AppConfig, LocalConfig } from "./schema.js";
```

- [ ] **步骤 7： 更新 .env.example**

在 `apps/worker/.env.example` 末尾追加：

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

- [ ] **步骤 8： 类型检查**

```bash
cd apps/worker && pnpm --filter @corekg/config typecheck
```

- [ ] **步骤 9： 提交**

```bash
git add apps/worker/packages/config/ apps/worker/.env.example
git commit -m "feat(config): add agent, daemon, rpc config schemas"
```

---

## 任务 2：创建 @corekg/rpc 包

### 需要创建的文件

- `apps/worker/packages/rpc/package.json`
- `apps/worker/packages/rpc/tsconfig.json`
- `apps/worker/packages/rpc/src/subjects.ts`
- `apps/worker/packages/rpc/src/schema.ts`
- `apps/worker/packages/rpc/src/server.ts`
- `apps/worker/packages/rpc/src/client.ts`
- `apps/worker/packages/rpc/src/index.ts`

### 步骤

- [ ] **步骤 1： 创建 package.json**

创建 `apps/worker/packages/rpc/package.json`：

```json
{
  "name": "@corekg/rpc",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "scripts": {
    "build": "tsc",
    "test": "vitest run --passWithNoTests --dir src",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "@nats-io/transport-node": "^3.4.0",
    "zod": "^4.4.3",
    "@corekg/logger": "workspace:*",
    "@corekg/nats": "workspace:*"
  },
  "devDependencies": {
    "@types/node": "^26.1.1",
    "typescript": "^7.0.2",
    "vitest": "^4.1.10"
  }
}
```

- [ ] **步骤 2： 创建 tsconfig.json**

创建 `apps/worker/packages/rpc/tsconfig.json`：

```json
{
  "extends": "../../tsconfig.base.json",
  "compilerOptions": {
    "outDir": "dist",
    "rootDir": "src"
  },
  "include": ["src/**/*"]
}
```

- [ ] **步骤 3： 创建 src/subjects.ts**

创建 `apps/worker/packages/rpc/src/subjects.ts`：

```typescript
export const RPC_SUBJECTS = {
  analysis: "core.task.rpc.analysis",
  copy: "core.task.rpc.copy",
  desc: "core.task.rpc.desc",
  mindmap: "core.task.rpc.mindmap",
  pdf_extract: "core.task.rpc.pdf_extract",
  video_extract: "core.task.rpc.video_extract",
  split_text_chunk: "core.task.rpc.split_text_chunk",
  insert_index: "core.task.rpc.insert_index",
} as const;

export type RPCSubject = (typeof RPC_SUBJECTS)[keyof typeof RPC_SUBJECTS];
```

- [ ] **步骤 4： 创建 src/schema.ts**

创建 `apps/worker/packages/rpc/src/schema.ts`：

```typescript
import { z } from "zod";

export const RPCResponseSchema = z.object({
  status: z.enum(["success", "fail"]),
  result: z.unknown().optional(),
  error: z.string().optional(),
});

export type RPCResponse = z.infer<typeof RPCResponseSchema>;
```

- [ ] **步骤 5： 创建 src/server.ts**

创建 `apps/worker/packages/rpc/src/server.ts`：

```typescript
import type { NatsConnection, Subscription } from "@nats-io/transport-node";
import { createLogger } from "@corekg/logger";
import { TaskPayloadSchema, type TaskPayload } from "@corekg/nats";
import type { RPCResponse } from "./schema.js";

const logger = createLogger("rpc-server");

export interface TaskHandlerEntry {
  subject: string;
  name: string;
  handler: (payload: TaskPayload) => Promise<{ status: "success" | "fail"; result?: unknown; error?: string }>;
}

export class RPCServer {
  private subscriptions: Subscription[] = [];

  constructor(
    private nc: NatsConnection,
    private handlers: TaskHandlerEntry[],
  ) {}

  async start(queueGroup: string): Promise<void> {
    for (const entry of this.handlers) {
      const sub = this.nc.subscribe(entry.subject, { queue: queueGroup });
      this.subscriptions.push(sub);
      logger.info({ subject: entry.subject, queueGroup }, "rpc handler registered");

      (async () => {
        for await (const msg of sub) {
          try {
            const raw = msg.json();
            const payload = TaskPayloadSchema.parse(raw);
            const result = await entry.handler(payload);
            const response: RPCResponse = {
              status: result.status,
              result: result.result,
              error: result.error,
            };
            msg.respond(JSON.stringify(response));
          } catch (err) {
            const errMsg = err instanceof Error ? err.message : String(err);
            logger.error({ subject: entry.subject, error: errMsg }, "rpc handler failed");
            const response: RPCResponse = { status: "fail", error: errMsg };
            msg.respond(JSON.stringify(response));
          }
        }
      })();
    }
  }

  async stop(): Promise<void> {
    for (const sub of this.subscriptions) {
      sub.unsubscribe();
    }
    this.subscriptions = [];
    logger.info("rpc server stopped");
  }
}
```

- [ ] **步骤 6： 创建 src/client.ts**

创建 `apps/worker/packages/rpc/src/client.ts`：

```typescript
import type { NatsConnection } from "@nats-io/transport-node";
import { RPCResponseSchema, type RPCResponse } from "./schema.js";

export interface RPCRequestOptions {
  timeoutMs?: number;
}

export class RPCClient {
  constructor(private nc: NatsConnection) {}

  async request(subject: string, payload: unknown, opts?: RPCRequestOptions): Promise<RPCResponse> {
    const timeout = opts?.timeoutMs ?? 300_000;
    const msg = await this.nc.request(subject, JSON.stringify(payload), { timeout });
    const raw = msg.json();
    return RPCResponseSchema.parse(raw);
  }
}
```

- [ ] **步骤 7： 创建 src/index.ts**

创建 `apps/worker/packages/rpc/src/index.ts`：

```typescript
export { RPC_SUBJECTS } from "./subjects.js";
export type { RPCSubject } from "./subjects.js";
export { RPCResponseSchema } from "./schema.js";
export type { RPCResponse } from "./schema.js";
export { RPCServer } from "./server.js";
export type { TaskHandlerEntry } from "./server.js";
export { RPCClient } from "./client.js";
export type { RPCRequestOptions } from "./client.js";
```

- [ ] **步骤 8： 安装依赖**

```bash
cd apps/worker && pnpm install
```

- [ ] **步骤 9： 类型检查**

```bash
cd apps/worker && pnpm --filter @corekg/rpc typecheck
```

- [ ] **步骤 10： 提交**

```bash
git add apps/worker/packages/rpc/
git commit -m "feat(rpc): add @corekg/rpc package with NATS request/reply"
```

---

## 任务 3：创建 @corekg/workers 包脚手架 + 类型定义

### 需要创建的文件

- `apps/worker/packages/workers/package.json`
- `apps/worker/packages/workers/tsconfig.json`
- `apps/worker/packages/workers/src/types.ts`
- `apps/worker/packages/workers/src/index.ts`

### 步骤

- [ ] **步骤 1： 创建 package.json**

创建 `apps/worker/packages/workers/package.json`：

```json
{
  "name": "@corekg/workers",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "scripts": {
    "build": "tsc",
    "test": "vitest run --passWithNoTests --dir src",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "uuid": "^14.0.1",
    "zod": "^4.4.3",
    "@corekg/config": "workspace:*",
    "@corekg/logger": "workspace:*",
    "@corekg/nats": "workspace:*",
    "@corekg/storage": "workspace:*",
    "@corekg/search": "workspace:*",
    "@corekg/ai": "workspace:*"
  },
  "devDependencies": {
    "@types/uuid": "^11.0.0",
    "@types/node": "^26.1.1",
    "typescript": "^7.0.2",
    "vitest": "^4.1.10"
  }
}
```

- [ ] **步骤 2： 创建 tsconfig.json**

创建 `apps/worker/packages/workers/tsconfig.json`：

```json
{
  "extends": "../../tsconfig.base.json",
  "compilerOptions": {
    "outDir": "dist",
    "rootDir": "src"
  },
  "include": ["src/**/*"]
}
```

- [ ] **步骤 3： 创建 src/types.ts**

创建 `apps/worker/packages/workers/src/types.ts`：

```typescript
import type { StorageProvider } from "@corekg/storage";
import type { ESProvider } from "@corekg/search";
import type { LLMProvider, EmbeddingProvider } from "@corekg/ai";
import type { TaskPayload } from "@corekg/nats";
import type pino from "pino";

export interface AgentClientConfig {
  apiUrl: string;
  apiKey: string;
  chunkSize: number;
  maxTokenSize: number;
  maxWorkers: number;
  pool: Record<string, string>;
}

export interface TaskContext {
  storage: StorageProvider;
  es: ESProvider;
  llm: LLMProvider;
  embedding: EmbeddingProvider;
  agentConfig: AgentClientConfig;
  daemonUrl: string;
  logger: pino.Logger;
}

export interface TaskHandlerResult {
  status: "success" | "fail";
  result?: unknown;
  error?: string;
}

export type TaskHandlerFn = (ctx: TaskContext, payload: TaskPayload) => Promise<TaskHandlerResult>;

export interface TaskHandlerDef {
  name: string;
  subject: string;
  handler: TaskHandlerFn;
}
```

- [ ] **步骤 4： 创建 src/index.ts (types only)**

创建 `apps/worker/packages/workers/src/index.ts`：

```typescript
export type {
  AgentClientConfig,
  TaskContext,
  TaskHandlerResult,
  TaskHandlerFn,
  TaskHandlerDef,
} from "./types.js";
```

- [ ] **步骤 5： 安装依赖**

```bash
cd apps/worker && pnpm install
```

- [ ] **步骤 6： 类型检查**

```bash
cd apps/worker && pnpm --filter @corekg/workers typecheck
```

- [ ] **步骤 7： 提交**

```bash
git add apps/worker/packages/workers/
git commit -m "feat(workers): add @corekg/workers package scaffold and types"
```

---

## 任务 4：添加 ES 辅助方法

### 需要修改的文件

- `apps/worker/packages/search/src/types.ts`
- `apps/worker/packages/search/src/es-provider.ts`

### 步骤

- [ ] **步骤 1： 在 types.ts 中扩展 VectorStore 和 SearchProvider 接口**

将 `apps/worker/packages/search/src/types.ts` 替换为：

```typescript
export interface ChunkDocument {
  forest_id: string;
  company_id: string;
  uin: string;
  file_id: string;
  version: string;
  file_name: string | null;
  type: "chunk" | "table" | "image" | "video" | "entity" | "file_description" | null;
  tokens: number;
  chunk_size: number;
  sequence: number;
  location: unknown | null;
  yg_location: unknown | null;
  description: string;
  description_hash: string;
  embedding: number[] | null;
  image_url: string | null;
  image_embedding: number[] | null;
  formula: string | null;
  table: string | null;
  title_level_ids: string[] | null;
  title_level: string[] | null;
  references: unknown | null;
  graph_info: unknown | null;
  graph_status: unknown | null;
}

export interface VectorStore {
  upsertChunks(index: string, docs: Record<string, ChunkDocument>): Promise<void>;
  deleteChunksByFileId(index: string, forestId: string, fileId: string, companyId: string): Promise<number>;
  insertDocument(index: string, id: string, doc: Record<string, unknown>): Promise<void>;
  deleteByType(index: string, forestId: string, fileId: string, type: string): Promise<number>;
}

export interface SearchProvider {
  getById(index: string, id: string): Promise<Record<string, unknown> | null>;
  query(index: string, body: Record<string, unknown>): Promise<Record<string, unknown>[]>;
  queryChunkIdsByFileId(index: string, fileId: string, limit?: number): Promise<string[]>;
}

export interface ESProvider {
  vectorStore: VectorStore;
  search: SearchProvider;
  close(): Promise<void>;
}
```

- [ ] **步骤 2： 在 es-provider.ts 中实现新方法**

将 `apps/worker/packages/search/src/es-provider.ts` 替换为：

```typescript
import { Client } from "@elastic/elasticsearch";
import type { z } from "zod";
import type { ESConfigSchema } from "@corekg/config";
import type { ChunkDocument, VectorStore, SearchProvider, ESProvider } from "./types.js";

export type { ChunkDocument, VectorStore, SearchProvider, ESProvider } from "./types.js";

type ESConfig = z.infer<typeof ESConfigSchema>;

export function createESProvider(config: ESConfig): ESProvider {
  const client = new Client({
    node: config.host,
    auth: { username: config.username, password: config.password },
    maxRetries: 3,
    requestTimeout: config.requestTimeoutMs,
    sniffOnStart: false,
  });

  const vectorStore: VectorStore = {
    async upsertChunks(index, documents) {
      const actions = Object.entries(documents).flatMap(([id, doc]) => [
        { index: { _index: index, _id: id } },
        doc,
      ]);
      await client.bulk({ operations: actions, refresh: "wait_for" });
    },

    async deleteChunksByFileId(index, forestId, fileId, companyId) {
      const result = await client.deleteByQuery({
        index,
        query: {
          bool: {
            filter: [
              { term: { forest_id: forestId } },
              { term: { file_id: fileId } },
              { term: { company_id: companyId } },
            ],
          },
        },
        conflicts: "proceed",
      });
      return (result.deleted ?? 0) as number;
    },

    async insertDocument(index, id, doc) {
      await client.index({ index, id, body: doc, refresh: "wait_for" });
    },

    async deleteByType(index, forestId, fileId, type) {
      const result = await client.deleteByQuery({
        index,
        query: {
          bool: {
            filter: [
              { term: { forest_id: forestId } },
              { term: { file_id: fileId } },
              { term: { type } },
            ],
          },
        },
        conflicts: "proceed",
      });
      return (result.deleted ?? 0) as number;
    },
  };

  const search: SearchProvider = {
    async getById(index, id) {
      try {
        const result = await client.get({ index, id });
        return result._source as Record<string, unknown> | null;
      } catch {
        return null;
      }
    },

    async query(index, body) {
      const result = await client.search({ index, ...body as any });
      return result.hits.hits.map((h) => h._source as Record<string, unknown>);
    },

    async queryChunkIdsByFileId(index, fileId, limit) {
      const result = await client.search({
        index,
        body: {
          query: {
            nested: {
              path: "references",
              query: {
                term: { "references.file_id": fileId },
              },
            },
          },
          _source: false,
          size: limit ?? 1000,
        },
      });
      return result.hits.hits.map((h) => h._id as string);
    },
  };

  return {
    vectorStore,
    search,
    async close() {
      await client.close();
    },
  };
}
```

- [ ] **步骤 3： 类型检查**

```bash
cd apps/worker && pnpm --filter @corekg/search typecheck
```

- [ ] **步骤 4： 提交**

```bash
git add apps/worker/packages/search/
git commit -m "feat(search): add insertDocument, deleteByType, queryChunkIdsByFileId"
```

---

## 任务 5：实现共享工具函数

### 需要创建的文件

- `apps/worker/packages/workers/src/agent-client.ts`
- `apps/worker/packages/workers/src/markdown-utils.ts`
- `apps/worker/packages/workers/src/daemon-client.ts`
- `apps/worker/packages/workers/src/algo-client.ts`

### 需要修改的文件

- `apps/worker/packages/workers/src/index.ts`

### 步骤

- [ ] **步骤 1： 创建 src/agent-client.ts**

创建 `apps/worker/packages/workers/src/agent-client.ts`：

```typescript
import type { TaskContext } from "./types.js";

interface AgentInput {
  name: string;
  value: string;
}

interface AgentRequest {
  model: string;
  chat_options: {
    input: AgentInput[];
  };
  stream: boolean;
}

interface AgentChoice {
  message: { content: string };
}

interface AgentResponse {
  choices: AgentChoice[];
}

export async function doAgentRequest(
  ctx: TaskContext,
  inputs: Record<string, string>,
  model: string,
  backups?: [string, string],
): Promise<string> {
  for (const [name, value] of Object.entries(inputs)) {
    if (value.length > ctx.agentConfig.maxTokenSize) {
      if (!backups) {
        throw new Error(`content exceeds maxTokenSize and no backup models provided`);
      }
      return doSplitMerge(ctx, value, backups[0], backups[1]);
    }
  }

  const inputList: AgentInput[] = Object.entries(inputs).map(([name, value]) => ({ name, value }));

  const body: AgentRequest = {
    model,
    chat_options: { input: inputList },
    stream: false,
  };

  let lastErr: Error | null = null;
  for (let i = 0; i < 20; i++) {
    try {
      const resp = await fetch(ctx.agentConfig.apiUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${ctx.agentConfig.apiKey}`,
        },
        body: JSON.stringify(body),
      });
      if (!resp.ok) throw new Error(`agent request failed: ${resp.status}`);
      const data = (await resp.json()) as AgentResponse;
      if (!data.choices?.length) throw new Error("no choices found");
      const content = data.choices[0]?.message?.content;
      if (!content) throw new Error("no content found");
      return content;
    } catch (err) {
      lastErr = err instanceof Error ? err : new Error(String(err));
      ctx.logger.warn({ attempt: i + 1, error: String(lastErr) }, "agent request retry");
    }
  }
  throw lastErr ?? new Error("agent request failed after 20 retries");
}

export function splitContent(content: string, chunkSize: number): string[] {
  if (content.length <= chunkSize) return [content];
  const paragraphs = content.split("\n");
  const chunks: string[] = [];
  let current = "";
  for (const para of paragraphs) {
    const newLen = current ? current.length + 1 + para.length : para.length;
    if (newLen <= chunkSize) {
      current = current ? current + "\n" + para : para;
    } else {
      if (current) chunks.push(current);
      current = para;
    }
  }
  if (current) chunks.push(current);
  return chunks;
}

export async function mapPhase(
  ctx: TaskContext,
  chunks: string[],
  chunkModel: string,
): Promise<string[]> {
  const results = new Array<string>(chunks.length);
  const semaphore = { count: 0, max: ctx.agentConfig.maxWorkers };

  const tasks = chunks.map((chunk, idx) =>
    (async () => {
      while (semaphore.count >= semaphore.max) {
        await new Promise((r) => setTimeout(r, 50));
      }
      semaphore.count++;
      try {
        const inputs = {
          input1: chunk,
          input2: `第${idx + 1}块，共${chunks.length}块`,
        };
        results[idx] = await doAgentRequest(ctx, inputs, chunkModel);
      } finally {
        semaphore.count--;
      }
    })(),
  );

  await Promise.all(tasks);
  return results;
}

export async function reducePhase(
  ctx: TaskContext,
  summaries: string[],
  chunkModel: string,
  mergeModel: string,
): Promise<string> {
  const combined = summaries.join("\n---\n");
  const inputs = {
    input1: combined,
    input2: `共${summaries.length}个分块的总结`,
  };
  return doAgentRequest(ctx, inputs, mergeModel, [chunkModel, mergeModel]);
}

export async function doSplitMerge(
  ctx: TaskContext,
  content: string,
  chunkModel: string,
  mergeModel: string,
): Promise<string> {
  const chunks = splitContent(content, ctx.agentConfig.chunkSize);
  ctx.logger.info({ chunkCount: chunks.length }, "split into chunks for merge");
  const summaries = await mapPhase(ctx, chunks, chunkModel);
  return reducePhase(ctx, summaries, chunkModel, mergeModel);
}
```

- [ ] **步骤 2： 创建 src/markdown-utils.ts**

创建 `apps/worker/packages/workers/src/markdown-utils.ts`：

```typescript
import { v4 as uuidv4 } from "uuid";

export function extractMarkdownTitles(md: string): string[] {
  const re = /(^#{1,6}\s+(.+)$)|(^(.+)\n[-=]{3,}$)/gm;
  const titles: string[] = [];
  let match: RegExpExecArray | null;
  while ((match = re.exec(md)) !== null) {
    if (match[2]) titles.push(match[2].trim());
    else if (match[4]) titles.push(match[4].trim());
  }
  return titles;
}

export function extractCodeBlock(codeType: string, text: string): string {
  const re = new RegExp(`\`\`\`${codeType}\\s*([\\s\\S]*?)\\s*\`\`\``, "m");
  const match = re.exec(text);
  return match ? match[1] : text;
}

interface Node {
  id: string;
  uuid?: string;
  children?: Node[];
}

export function processEmbeddedUuid(jsonStr: string): string {
  const node = JSON.parse(jsonStr) as Node;
  assignUuids(node);
  return JSON.stringify(node);
}

function assignUUIDs(node: Node): void {
  node.uuid = uuidv4();
  if (node.children) {
    for (const child of node.children) {
      assignUUIDs(child);
    }
  }
}
```

- [ ] **步骤 3： 创建 src/daemon-client.ts**

创建 `apps/worker/packages/workers/src/daemon-client.ts`：

```typescript
export async function daemonProcessPdf(
  daemonUrl: string,
  pdfFileName: string,
  targetPath: string,
  publicPath: string,
): Promise<void> {
  const resp = await fetch(daemonUrl, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      pdf_file_name: pdfFileName,
      target_path: targetPath,
      public_path: publicPath,
    }),
    signal: AbortSignal.timeout(600_000),
  });
  if (!resp.ok) {
    const body = await resp.text();
    throw new Error(`daemonProcessPdf failed: ${resp.status} ${body}`);
  }
}

export async function daemonProcessVideo(
  daemonUrl: string,
  videoPath: string,
  outputBaseDir: string,
  imagePrefix: string,
): Promise<void> {
  const resp = await fetch(daemonUrl, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      video_path: videoPath,
      output_base_dir: outputBaseDir,
      image_prefix: imagePrefix,
    }),
    signal: AbortSignal.timeout(600_000),
  });
  if (!resp.ok) {
    const body = await resp.text();
    throw new Error(`daemonProcessVideo failed: ${resp.status} ${body}`);
  }
}
```

- [ ] **步骤 4： 创建 src/algo-client.ts**

创建 `apps/worker/packages/workers/src/algo-client.ts`：

```typescript
export async function algoSplit(
  algoUrl: string,
  params: {
    uin: string;
    company_id: string;
    forest_id: string;
    file_id: string;
    content: string;
    es_index: string;
    file_ext: string;
  },
): Promise<string> {
  const form = new URLSearchParams(params);
  const resp = await fetch(`${algoUrl}/split`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: form.toString(),
  });
  if (!resp.ok) {
    const body = await resp.text();
    throw new Error(`algoSplit failed: ${resp.status} ${body}`);
  }
  return resp.text();
}

export async function algoIndex(
  algoUrl: string,
  params: {
    uin: string;
    company_id: string;
    forest_id: string;
    file_id: string;
    es_index: string;
  },
): Promise<void> {
  const form = new URLSearchParams(params);
  const resp = await fetch(`${algoUrl}/index`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: form.toString(),
  });
  if (!resp.ok) {
    const body = await resp.text();
    throw new Error(`algoIndex failed: ${resp.status} ${body}`);
  }
}
```

- [ ] **步骤 5： 更新 src/index.ts 以导出所有内容**

将 `apps/worker/packages/workers/src/index.ts` 替换为：

```typescript
export type {
  AgentClientConfig,
  TaskContext,
  TaskHandlerResult,
  TaskHandlerFn,
  TaskHandlerDef,
} from "./types.js";

export {
  doAgentRequest,
  splitContent,
  mapPhase,
  reducePhase,
  doSplitMerge,
} from "./agent-client.js";

export {
  extractMarkdownTitles,
  extractCodeBlock,
  processEmbeddedUuid,
} from "./markdown-utils.js";

export {
  daemonProcessPdf,
  daemonProcessVideo,
} from "./daemon-client.js";

export {
  algoSplit,
  algoIndex,
} from "./algo-client.js";
```

- [ ] **步骤 6： 类型检查**

```bash
cd apps/worker && pnpm --filter @corekg/workers typecheck
```

- [ ] **步骤 7： 提交**

```bash
git add apps/worker/packages/workers/
git commit -m "feat(workers): add shared utilities for agent, markdown, daemon, algo"
```

---

## 任务 6：实现 8 个任务处理器

### 需要创建的文件

- `apps/worker/packages/workers/src/handlers/analysis.ts`
- `apps/worker/packages/workers/src/handlers/copy.ts`
- `apps/worker/packages/workers/src/handlers/desc.ts`
- `apps/worker/packages/workers/src/handlers/mindmap.ts`
- `apps/worker/packages/workers/src/handlers/pdf-extract.ts`
- `apps/worker/packages/workers/src/handlers/video-extract.ts`
- `apps/worker/packages/workers/src/handlers/split-text-chunk.ts`
- `apps/worker/packages/workers/src/handlers/insert-index.ts`
- `apps/worker/packages/workers/src/handlers/registry.ts`

### 需要修改的文件

- `apps/worker/packages/workers/src/index.ts`

### 步骤

- [ ] **步骤 1： 创建 src/handlers/analysis.ts**

创建 `apps/worker/packages/workers/src/handlers/analysis.ts`：

```typescript
import { writeFile } from "node:fs/promises";
import { join } from "node:path";
import { tmpdir } from "node:os";
import { v4 as uuidv4 } from "uuid";
import { basename } from "node:path";
import type { TaskContext, TaskHandlerFn, TaskHandlerResult } from "../types.js";
import type { TaskPayload } from "@corekg/nats";
import { doAgentRequest } from "../agent-client.js";

export const handleAnalysis: TaskHandlerFn = async (ctx, payload) => {
  try {
    const tmpDir = join(tmpdir(), "task-analysis", uuidv4());
    await import("node:fs/promises").then((fs) => fs.mkdir(tmpDir, { recursive: true }));

    const tmpFile = join(tmpDir, basename(new URL(payload.file_url).pathname) || "input.md");
    await import("node:fs/promises").then(async (fs) => {
      const resp = await fetch(payload.file_url);
      if (!resp.ok) throw new Error(`Download failed: ${resp.status}`);
      await fs.writeFile(tmpFile, Buffer.from(await resp.arrayBuffer()));
    });

    const content = await import("node:fs/promises").then((fs) => fs.readFile(tmpFile, "utf-8"));

    const result = await doAgentRequest(ctx, { input1: content }, payload.split_config?.llm_model || "");

    const uploadPath = payload.upload_path || payload.storage_path || `analysis/${uuidv4()}.md`;
    const resultFile = join(tmpDir, "result.md");
    await import("node:fs/promises").then((fs) => fs.writeFile(resultFile, result, "utf-8"));
    await ctx.storage.uploadFile(resultFile, uploadPath);

    return { status: "success", result: uploadPath };
  } catch (err) {
    const error = err instanceof Error ? err.message : String(err);
    ctx.logger.error({ error, taskId: payload.file_id }, "analysis handler failed");
    return { status: "fail", error };
  }
};
```

- [ ] **步骤 2： 创建 src/handlers/copy.ts**

创建 `apps/worker/packages/workers/src/handlers/copy.ts`：

```typescript
import { writeFile, mkdir } from "node:fs/promises";
import { join, basename } from "node:path";
import { tmpdir } from "node:os";
import { v4 as uuidv4 } from "uuid";
import type { TaskHandlerFn } from "../types.js";
import type { TaskPayload } from "@corekg/nats";

export const handleCopy: TaskHandlerFn = async (ctx, payload) => {
  try {
    const tmpDir = join(tmpdir(), "task-copy", uuidv4());
    await mkdir(tmpDir, { recursive: true });

    const resp = await fetch(payload.file_url);
    if (!resp.ok) throw new Error(`Download failed: ${resp.status}`);
    const content = await resp.text();

    const tmpFile = join(tmpDir, "content.md");
    await writeFile(tmpFile, content, "utf-8");

    const uploadPath = payload.upload_path
      ? payload.upload_path + "content.md"
      : payload.storage_path
        ? payload.storage_path + "content.md"
        : `copy/${uuidv4()}/content.md`;

    await ctx.storage.uploadFile(tmpFile, uploadPath);

    return { status: "success", result: uploadPath };
  } catch (err) {
    const error = err instanceof Error ? err.message : String(err);
    ctx.logger.error({ error, taskId: payload.file_id }, "copy handler failed");
    return { status: "fail", error };
  }
};
```

- [ ] **步骤 3： 创建 src/handlers/mindmap.ts**

创建 `apps/worker/packages/workers/src/handlers/mindmap.ts`：

```typescript
import { writeFile, mkdir } from "node:fs/promises";
import { join } from "node:path";
import { tmpdir } from "node:os";
import { v4 as uuidv4 } from "uuid";
import type { TaskHandlerFn } from "../types.js";
import type { TaskPayload } from "@corekg/nats";
import { doAgentRequest } from "../agent-client.js";
import { extractMarkdownTitles, extractCodeBlock, processEmbeddedUuid } from "../markdown-utils.js";

export const handleMindmap: TaskHandlerFn = async (ctx, payload) => {
  try {
    const tmpDir = join(tmpdir(), "task-mindmap", uuidv4());
    await mkdir(tmpDir, { recursive: true });

    const resp = await fetch(payload.file_url);
    if (!resp.ok) throw new Error(`Download failed: ${resp.status}`);
    const content = await resp.text();

    const titles = extractMarkdownTitles(content).join("\n");

    const agentResult = await doAgentRequest(
      ctx,
      { input1: titles },
      payload.split_config?.llm_model || "",
    );

    const jsonCode = extractCodeBlock("json", agentResult);
    const withUuids = processEmbeddedUuid(jsonCode);

    const tmpFile = join(tmpDir, "mindmap.json");
    await writeFile(tmpFile, withUuids, "utf-8");

    const uploadPath = payload.upload_path || `mindmap/${uuidv4()}.json`;
    await ctx.storage.uploadFile(tmpFile, uploadPath);

    return { status: "success", result: uploadPath };
  } catch (err) {
    const error = err instanceof Error ? err.message : String(err);
    ctx.logger.error({ error, taskId: payload.file_id }, "mindmap handler failed");
    return { status: "fail", error };
  }
};
```

- [ ] **步骤 4： 创建 src/handlers/desc.ts**

创建 `apps/worker/packages/workers/src/handlers/desc.ts`：

```typescript
import { mkdir } from "node:fs/promises";
import { join } from "node:path";
import { tmpdir } from "node:os";
import { v4 as uuidv4 } from "uuid";
import type { TaskHandlerFn } from "../types.js";
import type { TaskPayload } from "@corekg/nats";
import { doAgentRequest } from "../agent-client.js";
import { extractMarkdownTitles, extractCodeBlock, processEmbeddedUuid } from "../markdown-utils.js";

export const handleDesc: TaskHandlerFn = async (ctx, payload) => {
  try {
    const esIndex = payload.es_index || payload.es?.index_name;
    if (!esIndex) throw new Error("es_index is required for desc task");

    const forestId = String(payload.forest_id);
    const fileId = String(payload.file_id);

    await ctx.es.vectorStore.deleteByType(esIndex, forestId, fileId, "file_description");

    const resp = await fetch(payload.file_url);
    if (!resp.ok) throw new Error(`Download failed: ${resp.status}`);
    const content = await resp.text();

    let abstract = "";
    let mindmap = "";
    let description = "";
    let embedding: number[] | null = null;

    const mindmapTask = doAgentRequest(
      ctx,
      { input1: extractMarkdownTitles(content).join("\n") },
      ctx.agentConfig.pool["mindmapMD"] || payload.split_config?.llm_model || "",
    );

    const abstractTask = doAgentRequest(
      ctx,
      { input1: content },
      ctx.agentConfig.pool["abstractMD"] || payload.split_config?.llm_model || "",
    );

    const results = await Promise.allSettled([mindmapTask, abstractTask]);
    if (results[0].status === "fulfilled") {
      const jsonCode = extractCodeBlock("json", results[0].value);
      mindmap = processEmbeddedUuid(jsonCode);
    }
    if (results[1].status === "fulfilled") {
      abstract = results[1].value;

      try {
        description = await doAgentRequest(
          ctx,
          { input1: abstract },
          ctx.agentConfig.pool["shortDescMD"] || payload.split_config?.llm_model || "",
        );
        embedding = await ctx.embedding.embed(description);
      } catch (e) {
        ctx.logger.warn({ error: String(e) }, "description/embed failed, continuing");
      }
    }

    const doc: Record<string, unknown> = {
      forest_id: forestId,
      company_id: String(payload.company_id),
      uin: String(payload.uin),
      file_id: fileId,
      type: "file_description",
      abstract,
      description,
      mind_map: mindmap,
      embedding: embedding ?? [],
    };

    await ctx.es.vectorStore.insertDocument(esIndex, `${fileId}-file_description`, doc);

    return { status: "success", result: `${description}\n${mindmap}\n${abstract}` };
  } catch (err) {
    const error = err instanceof Error ? err.message : String(err);
    ctx.logger.error({ error, taskId: payload.file_id }, "desc handler failed");
    return { status: "fail", error };
  }
};
```

- [ ] **步骤 5： 创建 src/handlers/pdf-extract.ts**

创建 `apps/worker/packages/workers/src/handlers/pdf-extract.ts`：

```typescript
import { mkdir } from "node:fs/promises";
import { join, extname } from "node:path";
import { tmpdir } from "node:os";
import { v4 as uuidv4 } from "uuid";
import type { TaskHandlerFn } from "../types.js";
import type { TaskPayload } from "@corekg/nats";
import { daemonProcessPdf } from "../daemon-client.js";

export const handlePdfExtract: TaskHandlerFn = async (ctx, payload) => {
  try {
    const tmpDir = join(tmpdir(), "task-pdf-extract", uuidv4());
    const storageDir = join(tmpDir, "storage");
    await mkdir(storageDir, { recursive: true });

    const fileName = payload.filename || payload.file_name || "input.pdf";
    const tmpFile = await ctx.storage.downloadFile(payload.file_url, tmpDir, fileName);

    const uploadPath = payload.upload_path || `pdf-extract/${uuidv4()}`;
    const publicPath = `${ctx.storage.getEndpoint()}/${uploadPath}`;

    await daemonProcessPdf(ctx.daemonUrl, tmpFile, storageDir, publicPath);

    const uploaded = await ctx.storage.uploadDirectory(storageDir, uploadPath);

    const files = uploaded.map((r) => r.key).slice(0, 10);
    return { status: "success", result: files };
  } catch (err) {
    const error = err instanceof Error ? err.message : String(err);
    ctx.logger.error({ error, taskId: payload.file_id }, "pdf-extract handler failed");
    return { status: "fail", error };
  }
};
```

- [ ] **步骤 6： 创建 src/handlers/video-extract.ts**

创建 `apps/worker/packages/workers/src/handlers/video-extract.ts`：

```typescript
import { mkdir } from "node:fs/promises";
import { join } from "node:path";
import { tmpdir } from "node:os";
import { v4 as uuidv4 } from "uuid";
import type { TaskHandlerFn } from "../types.js";
import type { TaskPayload } from "@corekg/nats";
import { daemonProcessVideo } from "../daemon-client.js";

export const handleVideoExtract: TaskHandlerFn = async (ctx, payload) => {
  try {
    const tmpDir = join(tmpdir(), "task-video-extract", uuidv4());
    const storageDir = join(tmpDir, "storage");
    await mkdir(storageDir, { recursive: true });

    const fileName = payload.filename || payload.file_name || "input.mp4";
    const tmpFile = await ctx.storage.downloadFile(payload.file_url, tmpDir, fileName);

    const uploadPath = payload.upload_path || `video-extract/${uuidv4()}`;
    const publicPath = `${ctx.storage.getEndpoint()}/${uploadPath}`;

    await daemonProcessVideo(ctx.daemonUrl, tmpFile, storageDir, publicPath);

    const uploaded = await ctx.storage.uploadDirectory(storageDir, uploadPath);

    const files = uploaded.map((r) => r.key);
    return { status: "success", result: files };
  } catch (err) {
    const error = err instanceof Error ? err.message : String(err);
    ctx.logger.error({ error, taskId: payload.file_id }, "video-extract handler failed");
    return { status: "fail", error };
  }
};
```

- [ ] **步骤 7： 创建 src/handlers/split-text-chunk.ts**

创建 `apps/worker/packages/workers/src/handlers/split-text-chunk.ts`：

```typescript
import type { TaskHandlerFn } from "../types.js";
import type { TaskPayload } from "@corekg/nats";
import { algoSplit } from "../algo-client.js";

export const handleSplitTextChunk: TaskHandlerFn = async (ctx, payload) => {
  try {
    const esIndex = payload.es_index || payload.es?.index_name;
    if (!esIndex) throw new Error("es_index is required");

    const forestId = String(payload.forest_id);
    const fileId = String(payload.file_id);
    const companyId = String(payload.company_id);

    await ctx.es.vectorStore.deleteChunksByFileId(esIndex, forestId, fileId, companyId);

    const splitResult = await algoSplit(ctx.daemonUrl, {
      uin: String(payload.uin),
      company_id: companyId,
      forest_id: forestId,
      file_id: fileId,
      content: payload.file_url,
      es_index: esIndex,
      file_ext: payload.file_ext || "",
    });

    await new Promise((r) => setTimeout(r, 2000));

    const chunkIds = await ctx.es.search.queryChunkIdsByFileId(esIndex, fileId);
    ctx.logger.info({ chunkCount: chunkIds.length }, "split-text-chunk done");

    return { status: "success", result: splitResult };
  } catch (err) {
    const error = err instanceof Error ? err.message : String(err);
    ctx.logger.error({ error, taskId: payload.file_id }, "split-text-chunk handler failed");
    return { status: "fail", error };
  }
};
```

- [ ] **步骤 8： 创建 src/handlers/insert-index.ts**

创建 `apps/worker/packages/workers/src/handlers/insert-index.ts`：

```typescript
import type { TaskHandlerFn } from "../types.js";
import type { TaskPayload } from "@corekg/nats";
import { algoIndex } from "../algo-client.js";

export const handleInsertIndex: TaskHandlerFn = async (ctx, payload) => {
  try {
    const esIndex = payload.es_index || payload.es?.index_name;
    if (!esIndex) throw new Error("es_index is required");

    const forestId = String(payload.forest_id);
    const fileId = String(payload.file_id);
    const companyId = String(payload.company_id);

    const chunkIds = await ctx.es.search.queryChunkIdsByFileId(esIndex, fileId);
    if (chunkIds.length === 0) {
      return { status: "success", result: null };
    }

    await algoIndex(ctx.daemonUrl, {
      uin: String(payload.uin),
      company_id: companyId,
      forest_id: forestId,
      file_id: fileId,
      es_index: esIndex,
    });

    return { status: "success", result: null };
  } catch (err) {
    const error = err instanceof Error ? err.message : String(err);
    ctx.logger.error({ error, taskId: payload.file_id }, "insert-index handler failed");
    return { status: "fail", error };
  }
};
```

- [ ] **步骤 9： 创建 src/handlers/registry.ts**

创建 `apps/worker/packages/workers/src/handlers/registry.ts`：

```typescript
import type { TaskHandlerDef } from "../types.js";
import { RPC_SUBJECTS } from "@corekg/rpc";
import { handleAnalysis } from "./analysis.js";
import { handleCopy } from "./copy.js";
import { handleDesc } from "./desc.js";
import { handleMindmap } from "./mindmap.js";
import { handlePdfExtract } from "./pdf-extract.js";
import { handleVideoExtract } from "./video-extract.js";
import { handleSplitTextChunk } from "./split-text-chunk.js";
import { handleInsertIndex } from "./insert-index.js";

export const handlerRegistry: TaskHandlerDef[] = [
  { name: "analysis", subject: RPC_SUBJECTS.analysis, handler: handleAnalysis },
  { name: "copy", subject: RPC_SUBJECTS.copy, handler: handleCopy },
  { name: "desc", subject: RPC_SUBJECTS.desc, handler: handleDesc },
  { name: "mindmap", subject: RPC_SUBJECTS.mindmap, handler: handleMindmap },
  { name: "pdf_extract", subject: RPC_SUBJECTS.pdf_extract, handler: handlePdfExtract },
  { name: "video_extract", subject: RPC_SUBJECTS.video_extract, handler: handleVideoExtract },
  { name: "split_text_chunk", subject: RPC_SUBJECTS.split_text_chunk, handler: handleSplitTextChunk },
  { name: "insert_index", subject: RPC_SUBJECTS.insert_index, handler: handleInsertIndex },
];
```

- [ ] **步骤 10： 更新 src/index.ts 以导出 handler**

将 `apps/worker/packages/workers/src/index.ts` 替换为：

```typescript
export type {
  AgentClientConfig,
  TaskContext,
  TaskHandlerResult,
  TaskHandlerFn,
  TaskHandlerDef,
} from "./types.js";

export {
  doAgentRequest,
  splitContent,
  mapPhase,
  reducePhase,
  doSplitMerge,
} from "./agent-client.js";

export {
  extractMarkdownTitles,
  extractCodeBlock,
  processEmbeddedUuid,
} from "./markdown-utils.js";

export {
  daemonProcessPdf,
  daemonProcessVideo,
} from "./daemon-client.js";

export {
  algoSplit,
  algoIndex,
} from "./algo-client.js";

export { handlerRegistry } from "./handlers/registry.js";
```

- [ ] **步骤 11： 向 workers 包添加 @corekg/rpc 依赖**

更新 `apps/worker/packages/workers/package.json` — 添加 `"@corekg/rpc": "workspace:*"` 到 dependencies。

- [ ] **步骤 12： 安装依赖**

```bash
cd apps/worker && pnpm install
```

- [ ] **步骤 13： 类型检查**

```bash
cd apps/worker && pnpm --filter @corekg/workers typecheck
```

- [ ] **步骤 14： 提交**

```bash
git add apps/worker/packages/workers/
git commit -m "feat(workers): implement 8 task handlers with registry"
```

---

## 任务 7：添加 8 个 CLI 命令到 apps/worker

### 需要创建的文件

- `apps/worker/apps/worker/src/commands/analysis.ts`
- `apps/worker/apps/worker/src/commands/copy.ts`
- `apps/worker/apps/worker/src/commands/desc.ts`
- `apps/worker/apps/worker/src/commands/mindmap.ts`
- `apps/worker/apps/worker/src/commands/pdf-extract.ts`
- `apps/worker/apps/worker/src/commands/video-extract.ts`
- `apps/worker/apps/worker/src/commands/split.ts`
- `apps/worker/apps/worker/src/commands/index-cmd.ts`

### 需要修改的文件

- `apps/worker/apps/worker/src/index.ts`

### 步骤

- [ ] **步骤 1： 创建 src/commands/analysis.ts**

创建 `apps/worker/apps/worker/src/commands/analysis.ts`：

```typescript
import { Command } from "commander";
import { createLogger } from "@corekg/logger";
import { loadLocalConfig } from "@corekg/config";
import { createS3Provider } from "@corekg/storage";
import { createESProvider } from "@corekg/search";
import { createLLMProvider, createEmbeddingProvider } from "@corekg/ai";
import { type TaskContext, handleAnalysis } from "@corekg/workers";
import type { TaskPayload } from "@corekg/nats";

const logger = createLogger("cli-analysis");

export function createAnalysisCommand(): Command {
  return new Command("analysis")
    .description("Run analysis task locally")
    .requiredOption("--file-url <url>", "file URL to process")
    .optionalOption("--upload-path <path>", "S3 upload path")
    .optionalOption("--llm-model <model>", "LLM model name")
    .action(async (opts) => {
      try {
        const cfg = loadLocalConfig();
        if (!cfg.agent?.apiUrl) throw new Error("AGENT_API_URL required");
        const s3 = { endpointUrl: process.env.S3_ENDPOINT_URL!, accessKeyId: process.env.S3_ACCESS_KEY_ID!, secretAccessKey: process.env.S3_SECRET_ACCESS_KEY!, defaultBucket: process.env.S3_DEFAULT_BUCKET!, publicEndpointUrl: process.env.S3_PUBLIC_ENDPOINT_URL || undefined };
        const es = { host: process.env.ES_HOST!, username: process.env.ES_USERNAME!, password: process.env.ES_PASSWORD! };
        const ctx: TaskContext = {
          storage: createS3Provider(s3),
          es: createESProvider(es),
          llm: cfg.llm ? createLLMProvider(cfg.llm) : null as any,
          embedding: cfg.embedding ? createEmbeddingProvider(cfg.embedding) : null as any,
          agentConfig: { apiUrl: cfg.agent.apiUrl, apiKey: cfg.agent.apiKey, chunkSize: cfg.agent.chunkSize, maxTokenSize: cfg.agent.maxTokenSize, maxWorkers: cfg.agent.maxWorkers, pool: cfg.agent.pool },
          daemonUrl: process.env.DAEMON_URL || "http://localhost:5000/local.Run",
          logger,
        };
        const payload: TaskPayload = { task_type: "analysis", file_id: "0", file_url: opts.fileUrl, company_id: "0", forest_id: "0", uin: "0", upload_path: opts.uploadPath };
        if (opts.llmModel) payload.split_config = { llm_model: opts.llmModel };
        const result = await handleAnalysis(ctx, payload);
        console.log(JSON.stringify(result, null, 2));
        process.exit(result.status === "success" ? 0 : 1);
      } catch (err) {
        logger.error({ err }, "analysis command failed");
        process.exit(1);
      }
    });
}
```

- [ ] **步骤 2： 创建 src/commands/copy.ts**

创建 `apps/worker/apps/worker/src/commands/copy.ts`：

```typescript
import { Command } from "commander";
import { createLogger } from "@corekg/logger";
import { loadLocalConfig } from "@corekg/config";
import { createS3Provider } from "@corekg/storage";
import { createESProvider } from "@corekg/search";
import { createLLMProvider, createEmbeddingProvider } from "@corekg/ai";
import { type TaskContext, handleCopy } from "@corekg/workers";
import type { TaskPayload } from "@corekg/nats";

const logger = createLogger("cli-copy");

export function createCopyCommand(): Command {
  return new Command("copy")
    .description("Run copy task locally")
    .requiredOption("--file-url <url>", "file URL to copy")
    .optionalOption("--upload-path <path>", "S3 upload path")
    .action(async (opts) => {
      try {
        const cfg = loadLocalConfig();
        const s3 = { endpointUrl: process.env.S3_ENDPOINT_URL!, accessKeyId: process.env.S3_ACCESS_KEY_ID!, secretAccessKey: process.env.S3_SECRET_ACCESS_KEY!, defaultBucket: process.env.S3_DEFAULT_BUCKET!, publicEndpointUrl: process.env.S3_PUBLIC_ENDPOINT_URL || undefined };
        const es = { host: process.env.ES_HOST!, username: process.env.ES_USERNAME!, password: process.env.ES_PASSWORD! };
        const ctx: TaskContext = {
          storage: createS3Provider(s3),
          es: createESProvider(es),
          llm: cfg.llm ? createLLMProvider(cfg.llm) : null as any,
          embedding: cfg.embedding ? createEmbeddingProvider(cfg.embedding) : null as any,
          agentConfig: { apiUrl: cfg.agent?.apiUrl || "", apiKey: cfg.agent?.apiKey || "", chunkSize: cfg.agent?.chunkSize || 60000, maxTokenSize: cfg.agent?.maxTokenSize || 120000, maxWorkers: cfg.agent?.maxWorkers || 50, pool: cfg.agent?.pool || {} },
          daemonUrl: process.env.DAEMON_URL || "http://localhost:5000/local.Run",
          logger,
        };
        const payload: TaskPayload = { task_type: "copy", file_id: "0", file_url: opts.fileUrl, company_id: "0", forest_id: "0", uin: "0", upload_path: opts.uploadPath };
        const result = await handleCopy(ctx, payload);
        console.log(JSON.stringify(result, null, 2));
        process.exit(result.status === "success" ? 0 : 1);
      } catch (err) {
        logger.error({ err }, "copy command failed");
        process.exit(1);
      }
    });
}
```

- [ ] **步骤 3： 创建 src/commands/desc.ts**

创建 `apps/worker/apps/worker/src/commands/desc.ts`：

```typescript
import { Command } from "commander";
import { createLogger } from "@corekg/logger";
import { loadLocalConfig } from "@corekg/config";
import { createS3Provider } from "@corekg/storage";
import { createESProvider } from "@corekg/search";
import { createLLMProvider, createEmbeddingProvider } from "@corekg/ai";
import { type TaskContext, handleDesc } from "@corekg/workers";
import type { TaskPayload } from "@corekg/nats";

const logger = createLogger("cli-desc");

export function createDescCommand(): Command {
  return new Command("desc")
    .description("Run desc task locally")
    .requiredOption("--file-url <url>", "file URL")
    .requiredOption("--es-index <index>", "ES index name")
    .optionalOption("--forest-id <id>", "forest ID")
    .optionalOption("--file-id <id>", "file ID")
    .optionalOption("--company-id <id>", "company ID")
    .optionalOption("--uin <id>", "uin")
    .action(async (opts) => {
      try {
        const cfg = loadLocalConfig();
        if (!cfg.agent?.apiUrl) throw new Error("AGENT_API_URL required");
        const s3 = { endpointUrl: process.env.S3_ENDPOINT_URL!, accessKeyId: process.env.S3_ACCESS_KEY_ID!, secretAccessKey: process.env.S3_SECRET_ACCESS_KEY!, defaultBucket: process.env.S3_DEFAULT_BUCKET!, publicEndpointUrl: process.env.S3_PUBLIC_ENDPOINT_URL || undefined };
        const es = { host: process.env.ES_HOST!, username: process.env.ES_USERNAME!, password: process.env.ES_PASSWORD! };
        const ctx: TaskContext = {
          storage: createS3Provider(s3),
          es: createESProvider(es),
          llm: cfg.llm ? createLLMProvider(cfg.llm) : null as any,
          embedding: cfg.embedding ? createEmbeddingProvider(cfg.embedding) : null as any,
          agentConfig: { apiUrl: cfg.agent.apiUrl, apiKey: cfg.agent.apiKey, chunkSize: cfg.agent.chunkSize, maxTokenSize: cfg.agent.maxTokenSize, maxWorkers: cfg.agent.maxWorkers, pool: cfg.agent.pool },
          daemonUrl: process.env.DAEMON_URL || "http://localhost:5000/local.Run",
          logger,
        };
        const payload: TaskPayload = {
          task_type: "desc", file_id: opts.fileId || "0", file_url: opts.fileUrl,
          company_id: opts.companyId || "0", forest_id: opts.forestId || "0", uin: opts.uin || "0",
          es_index: opts.esIndex,
        };
        const result = await handleDesc(ctx, payload);
        console.log(JSON.stringify(result, null, 2));
        process.exit(result.status === "success" ? 0 : 1);
      } catch (err) {
        logger.error({ err }, "desc command failed");
        process.exit(1);
      }
    });
}
```

- [ ] **步骤 4： 创建 src/commands/mindmap.ts**

创建 `apps/worker/apps/worker/src/commands/mindmap.ts`：

```typescript
import { Command } from "commander";
import { createLogger } from "@corekg/logger";
import { loadLocalConfig } from "@corekg/config";
import { createS3Provider } from "@corekg/storage";
import { createESProvider } from "@corekg/search";
import { createLLMProvider, createEmbeddingProvider } from "@corekg/ai";
import { type TaskContext, handleMindmap } from "@corekg/workers";
import type { TaskPayload } from "@corekg/nats";

const logger = createLogger("cli-mindmap");

export function createMindmapCommand(): Command {
  return new Command("mindmap")
    .description("Run mindmap task locally")
    .requiredOption("--file-url <url>", "file URL")
    .optionalOption("--upload-path <path>", "S3 upload path")
    .optionalOption("--llm-model <model>", "LLM model name")
    .action(async (opts) => {
      try {
        const cfg = loadLocalConfig();
        const s3 = { endpointUrl: process.env.S3_ENDPOINT_URL!, accessKeyId: process.env.S3_ACCESS_KEY_ID!, secretAccessKey: process.env.S3_SECRET_ACCESS_KEY!, defaultBucket: process.env.S3_DEFAULT_BUCKET!, publicEndpointUrl: process.env.S3_PUBLIC_ENDPOINT_URL || undefined };
        const es = { host: process.env.ES_HOST!, username: process.env.ES_USERNAME!, password: process.env.ES_PASSWORD! };
        const ctx: TaskContext = {
          storage: createS3Provider(s3),
          es: createESProvider(es),
          llm: cfg.llm ? createLLMProvider(cfg.llm) : null as any,
          embedding: cfg.embedding ? createEmbeddingProvider(cfg.embedding) : null as any,
          agentConfig: { apiUrl: cfg.agent?.apiUrl || "", apiKey: cfg.agent?.apiKey || "", chunkSize: cfg.agent?.chunkSize || 60000, maxTokenSize: cfg.agent?.maxTokenSize || 120000, maxWorkers: cfg.agent?.maxWorkers || 50, pool: cfg.agent?.pool || {} },
          daemonUrl: process.env.DAEMON_URL || "http://localhost:5000/local.Run",
          logger,
        };
        const payload: TaskPayload = { task_type: "mindmap", file_id: "0", file_url: opts.fileUrl, company_id: "0", forest_id: "0", uin: "0", upload_path: opts.uploadPath };
        if (opts.llmModel) payload.split_config = { llm_model: opts.llmModel };
        const result = await handleMindmap(ctx, payload);
        console.log(JSON.stringify(result, null, 2));
        process.exit(result.status === "success" ? 0 : 1);
      } catch (err) {
        logger.error({ err }, "mindmap command failed");
        process.exit(1);
      }
    });
}
```

- [ ] **步骤 5： 创建 src/commands/pdf-extract.ts**

创建 `apps/worker/apps/worker/src/commands/pdf-extract.ts`：

```typescript
import { Command } from "commander";
import { createLogger } from "@corekg/logger";
import { loadLocalConfig } from "@corekg/config";
import { createS3Provider } from "@corekg/storage";
import { createESProvider } from "@corekg/search";
import { createLLMProvider, createEmbeddingProvider } from "@corekg/ai";
import { type TaskContext, handlePdfExtract } from "@corekg/workers";
import type { TaskPayload } from "@corekg/nats";

const logger = createLogger("cli-pdf-extract");

export function createPdfExtractCommand(): Command {
  return new Command("pdf-extract")
    .description("Run PDF extract task locally")
    .requiredOption("--file-url <url>", "PDF file URL")
    .optionalOption("--upload-path <path>", "S3 upload path")
    .optionalOption("--daemon-url <url>", "daemon API URL")
    .action(async (opts) => {
      try {
        const cfg = loadLocalConfig();
        const s3 = { endpointUrl: process.env.S3_ENDPOINT_URL!, accessKeyId: process.env.S3_ACCESS_KEY_ID!, secretAccessKey: process.env.S3_SECRET_ACCESS_KEY!, defaultBucket: process.env.S3_DEFAULT_BUCKET!, publicEndpointUrl: process.env.S3_PUBLIC_ENDPOINT_URL || undefined };
        const es = { host: process.env.ES_HOST!, username: process.env.ES_USERNAME!, password: process.env.ES_PASSWORD! };
        const ctx: TaskContext = {
          storage: createS3Provider(s3),
          es: createESProvider(es),
          llm: cfg.llm ? createLLMProvider(cfg.llm) : null as any,
          embedding: cfg.embedding ? createEmbeddingProvider(cfg.embedding) : null as any,
          agentConfig: { apiUrl: cfg.agent?.apiUrl || "", apiKey: cfg.agent?.apiKey || "", chunkSize: cfg.agent?.chunkSize || 60000, maxTokenSize: cfg.agent?.maxTokenSize || 120000, maxWorkers: cfg.agent?.maxWorkers || 50, pool: cfg.agent?.pool || {} },
          daemonUrl: opts.daemonUrl || process.env.DAEMON_URL || "http://localhost:5000/local.Run",
          logger,
        };
        const payload: TaskPayload = { task_type: "pdf_extract", file_id: "0", file_url: opts.fileUrl, company_id: "0", forest_id: "0", uin: "0", upload_path: opts.uploadPath };
        const result = await handlePdfExtract(ctx, payload);
        console.log(JSON.stringify(result, null, 2));
        process.exit(result.status === "success" ? 0 : 1);
      } catch (err) {
        logger.error({ err }, "pdf-extract command failed");
        process.exit(1);
      }
    });
}
```

- [ ] **步骤 6： 创建 src/commands/video-extract.ts**

创建 `apps/worker/apps/worker/src/commands/video-extract.ts`：

```typescript
import { Command } from "commander";
import { createLogger } from "@corekg/logger";
import { loadLocalConfig } from "@corekg/config";
import { createS3Provider } from "@corekg/storage";
import { createESProvider } from "@corekg/search";
import { createLLMProvider, createEmbeddingProvider } from "@corekg/ai";
import { type TaskContext, handleVideoExtract } from "@corekg/workers";
import type { TaskPayload } from "@corekg/nats";

const logger = createLogger("cli-video-extract");

export function createVideoExtractCommand(): Command {
  return new Command("video-extract")
    .description("Run video extract task locally")
    .requiredOption("--file-url <url>", "video file URL")
    .optionalOption("--upload-path <path>", "S3 upload path")
    .optionalOption("--daemon-url <url>", "daemon API URL")
    .action(async (opts) => {
      try {
        const cfg = loadLocalConfig();
        const s3 = { endpointUrl: process.env.S3_ENDPOINT_URL!, accessKeyId: process.env.S3_ACCESS_KEY_ID!, secretAccessKey: process.env.S3_SECRET_ACCESS_KEY!, defaultBucket: process.env.S3_DEFAULT_BUCKET!, publicEndpointUrl: process.env.S3_PUBLIC_ENDPOINT_URL || undefined };
        const es = { host: process.env.ES_HOST!, username: process.env.ES_USERNAME!, password: process.env.ES_PASSWORD! };
        const ctx: TaskContext = {
          storage: createS3Provider(s3),
          es: createESProvider(es),
          llm: cfg.llm ? createLLMProvider(cfg.llm) : null as any,
          embedding: cfg.embedding ? createEmbeddingProvider(cfg.embedding) : null as any,
          agentConfig: { apiUrl: cfg.agent?.apiUrl || "", apiKey: cfg.agent?.apiKey || "", chunkSize: cfg.agent?.chunkSize || 60000, maxTokenSize: cfg.agent?.maxTokenSize || 120000, maxWorkers: cfg.agent?.maxWorkers || 50, pool: cfg.agent?.pool || {} },
          daemonUrl: opts.daemonUrl || process.env.DAEMON_URL || "http://localhost:5000/local.Run",
          logger,
        };
        const payload: TaskPayload = { task_type: "video_extract", file_id: "0", file_url: opts.fileUrl, company_id: "0", forest_id: "0", uin: "0", upload_path: opts.uploadPath };
        const result = await handleVideoExtract(ctx, payload);
        console.log(JSON.stringify(result, null, 2));
        process.exit(result.status === "success" ? 0 : 1);
      } catch (err) {
        logger.error({ err }, "video-extract command failed");
        process.exit(1);
      }
    });
}
```

- [ ] **步骤 7： 创建 src/commands/split.ts**

创建 `apps/worker/apps/worker/src/commands/split.ts`：

```typescript
import { Command } from "commander";
import { createLogger } from "@corekg/logger";
import { loadLocalConfig } from "@corekg/config";
import { createS3Provider } from "@corekg/storage";
import { createESProvider } from "@corekg/search";
import { createLLMProvider, createEmbeddingProvider } from "@corekg/ai";
import { type TaskContext, handleSplitTextChunk } from "@corekg/workers";
import type { TaskPayload } from "@corekg/nats";

const logger = createLogger("cli-split");

export function createSplitCommand(): Command {
  return new Command("split")
    .description("Run split-text-chunk task locally")
    .requiredOption("--file-url <url>", "file URL")
    .requiredOption("--es-index <index>", "ES index name")
    .optionalOption("--forest-id <id>", "forest ID")
    .optionalOption("--file-id <id>", "file ID")
    .optionalOption("--company-id <id>", "company ID")
    .optionalOption("--uin <id>", "uin")
    .optionalOption("--file-ext <ext>", "file extension")
    .optionalOption("--algo-url <url>", "algo service URL")
    .action(async (opts) => {
      try {
        const cfg = loadLocalConfig();
        const s3 = { endpointUrl: process.env.S3_ENDPOINT_URL!, accessKeyId: process.env.S3_ACCESS_KEY_ID!, secretAccessKey: process.env.S3_SECRET_ACCESS_KEY!, defaultBucket: process.env.S3_DEFAULT_BUCKET!, publicEndpointUrl: process.env.S3_PUBLIC_ENDPOINT_URL || undefined };
        const es = { host: process.env.ES_HOST!, username: process.env.ES_USERNAME!, password: process.env.ES_PASSWORD! };
        const ctx: TaskContext = {
          storage: createS3Provider(s3),
          es: createESProvider(es),
          llm: cfg.llm ? createLLMProvider(cfg.llm) : null as any,
          embedding: cfg.embedding ? createEmbeddingProvider(cfg.embedding) : null as any,
          agentConfig: { apiUrl: cfg.agent?.apiUrl || "", apiKey: cfg.agent?.apiKey || "", chunkSize: cfg.agent?.chunkSize || 60000, maxTokenSize: cfg.agent?.maxTokenSize || 120000, maxWorkers: cfg.agent?.maxWorkers || 50, pool: cfg.agent?.pool || {} },
          daemonUrl: opts.algoUrl || process.env.DAEMON_URL || "http://localhost:5000/local.Run",
          logger,
        };
        const payload: TaskPayload = {
          task_type: "split_text_chunk", file_id: opts.fileId || "0", file_url: opts.fileUrl,
          company_id: opts.companyId || "0", forest_id: opts.forestId || "0", uin: opts.uin || "0",
          es_index: opts.esIndex, file_ext: opts.fileExt,
        };
        const result = await handleSplitTextChunk(ctx, payload);
        console.log(JSON.stringify(result, null, 2));
        process.exit(result.status === "success" ? 0 : 1);
      } catch (err) {
        logger.error({ err }, "split command failed");
        process.exit(1);
      }
    });
}
```

- [ ] **步骤 8： 创建 src/commands/index-cmd.ts**

创建 `apps/worker/apps/worker/src/commands/index-cmd.ts`：

```typescript
import { Command } from "commander";
import { createLogger } from "@corekg/logger";
import { loadLocalConfig } from "@corekg/config";
import { createS3Provider } from "@corekg/storage";
import { createESProvider } from "@corekg/search";
import { createLLMProvider, createEmbeddingProvider } from "@corekg/ai";
import { type TaskContext, handleInsertIndex } from "@corekg/workers";
import type { TaskPayload } from "@corekg/nats";

const logger = createLogger("cli-index");

export function createIndexCommand(): Command {
  return new Command("index")
    .description("Run insert-index task locally")
    .requiredOption("--file-url <url>", "file URL")
    .requiredOption("--es-index <index>", "ES index name")
    .optionalOption("--forest-id <id>", "forest ID")
    .optionalOption("--file-id <id>", "file ID")
    .optionalOption("--company-id <id>", "company ID")
    .optionalOption("--uin <id>", "uin")
    .optionalOption("--algo-url <url>", "algo service URL")
    .action(async (opts) => {
      try {
        const cfg = loadLocalConfig();
        const s3 = { endpointUrl: process.env.S3_ENDPOINT_URL!, accessKeyId: process.env.S3_ACCESS_KEY_ID!, secretAccessKey: process.env.S3_SECRET_ACCESS_KEY!, defaultBucket: process.env.S3_DEFAULT_BUCKET!, publicEndpointUrl: process.env.S3_PUBLIC_ENDPOINT_URL || undefined };
        const es = { host: process.env.ES_HOST!, username: process.env.ES_USERNAME!, password: process.env.ES_PASSWORD! };
        const ctx: TaskContext = {
          storage: createS3Provider(s3),
          es: createESProvider(es),
          llm: cfg.llm ? createLLMProvider(cfg.llm) : null as any,
          embedding: cfg.embedding ? createEmbeddingProvider(cfg.embedding) : null as any,
          agentConfig: { apiUrl: cfg.agent?.apiUrl || "", apiKey: cfg.agent?.apiKey || "", chunkSize: cfg.agent?.chunkSize || 60000, maxTokenSize: cfg.agent?.maxTokenSize || 120000, maxWorkers: cfg.agent?.maxWorkers || 50, pool: cfg.agent?.pool || {} },
          daemonUrl: opts.algoUrl || process.env.DAEMON_URL || "http://localhost:5000/local.Run",
          logger,
        };
        const payload: TaskPayload = {
          task_type: "insert_index", file_id: opts.fileId || "0", file_url: opts.fileUrl,
          company_id: opts.companyId || "0", forest_id: opts.forestId || "0", uin: opts.uin || "0",
          es_index: opts.esIndex,
        };
        const result = await handleInsertIndex(ctx, payload);
        console.log(JSON.stringify(result, null, 2));
        process.exit(result.status === "success" ? 0 : 1);
      } catch (err) {
        logger.error({ err }, "index command failed");
        process.exit(1);
      }
    });
}
```

- [ ] **步骤 9： Update apps/worker apps/worker/src/index.ts**

将 `apps/worker/apps/worker/src/index.ts` 替换为：

```typescript
import { Command } from "commander";
import { createLogger } from "@corekg/logger";
import { createChunkCommand } from "./commands/chunk.js";
import { createExtractCommand } from "./commands/extract.js";
import { createAnalysisCommand } from "./commands/analysis.js";
import { createCopyCommand } from "./commands/copy.js";
import { createDescCommand } from "./commands/desc.js";
import { createMindmapCommand } from "./commands/mindmap.js";
import { createPdfExtractCommand } from "./commands/pdf-extract.js";
import { createVideoExtractCommand } from "./commands/video-extract.js";
import { createSplitCommand } from "./commands/split.js";
import { createIndexCommand } from "./commands/index-cmd.js";

const logger = createLogger("kealgo");

export { main } from "./worker-main.js";

const program = new Command("kealgo")
  .description("kealgo worker and local CLI tools for document processing")
  .action(async () => {
    const mod = await import("./worker-main.js");
    await mod.main();
  });

program.addCommand(createChunkCommand());
program.addCommand(createExtractCommand());
program.addCommand(createAnalysisCommand());
program.addCommand(createCopyCommand());
program.addCommand(createDescCommand());
program.addCommand(createMindmapCommand());
program.addCommand(createPdfExtractCommand());
program.addCommand(createVideoExtractCommand());
program.addCommand(createSplitCommand());
program.addCommand(createIndexCommand());

program.parseAsync(process.argv).catch((err) => {
  logger.error(err, "fatal");
  process.exit(1);
});
```

- [ ] **步骤 10： 向 worker 应用 package.json 添加 @corekg/workers 依赖**

向 `apps/worker/apps/worker/package.json` 添加 `"@corekg/workers": "workspace:*"` 依赖。

- [ ] **步骤 11： 向 worker 应用 package.json 添加 @corekg/rpc 依赖**

向 `apps/worker/apps/worker/package.json` 添加 `"@corekg/rpc": "workspace:*"` 依赖。

- [ ] **步骤 12： 安装依赖并类型检查**

```bash
cd apps/worker && pnpm install && pnpm --filter @corekg/kealgo-worker typecheck
```

- [ ] **步骤 13： 提交**

```bash
git add apps/worker/apps/worker/
git commit -m "feat(worker): add 8 CLI commands for task handlers"
```

---

## 任务 8：将 RPCServer 集成到 worker-main.ts

### 需要修改的文件

- `apps/worker/apps/worker/src/worker-main.ts`

### 步骤

- [ ] **步骤 1： 更新 worker-main.ts 集成 RPC server**

将 `apps/worker/apps/worker/src/worker-main.ts` 替换为：

```typescript
import os from "node:os";
import { createLogger } from "@corekg/logger";
import { loadConfig } from "@corekg/config";
import { createNATSClient, TaskConsumer, publishCallback, TaskPayloadSchema } from "@corekg/nats";
import { createS3Provider } from "@corekg/storage";
import { createESProvider } from "@corekg/search";
import { createLLMProvider, createEmbeddingProvider } from "@corekg/ai";
import { ChunkerService, registerBuiltinStrategies } from "@corekg/chunker";
import { RPCServer } from "@corekg/rpc";
import { handlerRegistry, type TaskContext } from "@corekg/workers";

const logger = createLogger("kealgo");

export async function main() {
  const config = loadConfig();
  const workerId = config.workerId || `${os.hostname()}-${process.pid}`;

  logger.info({ workerId }, "initializing kealgo worker");

  registerBuiltinStrategies();

  const storage = createS3Provider(config.s3);
  const es = createESProvider(config.es);
  const llm = createLLMProvider(config.llm);
  const embedding = createEmbeddingProvider(config.embedding);

  const chunker = new ChunkerService({ storage, llm, embedding });

  const { nc, js, jsm } = await createNATSClient(config);

  const chunkConsumer = new TaskConsumer(js, jsm, {
    stream: config.nats.stream,
    subject: config.nats.subjects?.chunker ?? "core.task.ke.knowledge_task",
    durableName: "kealgo-chunker",
  }, async (taskId, rawPayload) => {
    try {
      const payload = TaskPayloadSchema.parse(rawPayload);
      const sc = payload.split_config;
      const pre = sc?.preprocessing_rules;

      const docs = await chunker.process({
        url: payload.file_url,
        forestId: String(payload.forest_id),
        companyId: String(payload.company_id),
        uin: String(payload.uin),
        fileId: String(payload.file_id),
        fileName: payload.file_name || payload.filename || null,
        fileExt: payload.file_ext || null,
        indexName: payload.es_index || payload.es?.index_name || null,
        removeEmail: pre?.remove_email ?? true,
        removeUrl: pre?.remove_url ?? true,
        removeEmptyLine: pre?.remove_empty_line ?? true,
        splitMode: sc?.split_mode ?? "smart",
        chunkTokenNum: sc?.chunk_token_num ?? sc?.chunk_size ?? 1024,
        minChunkTokens: sc?.min_chunk_tokens ?? 10,
        splitLevel: sc?.split_level ?? 2,
        overlapRatio: sc?.overlap_ratio ?? sc?.split_overlap ?? 0,
        regexPattern: sc?.regex_pattern ?? "",
        enableHeadingInContent: sc?.enable_heading_in_content ?? false,
        llmConcurrency: sc?.llm_max_concurrency ?? config.concurrency.llmWorkers,
        embeddingConcurrency: sc?.eb_max_concurrency ?? config.concurrency.embeddingWorkers,
      });

      const index = payload.es_index || payload.es?.index_name || "default";

      await es.vectorStore.deleteChunksByFileId(
        index,
        String(payload.forest_id),
        String(payload.file_id),
        String(payload.company_id),
      );

      await es.vectorStore.upsertChunks(index, docs);

      publishCallback(nc, payload.task_type, {
        task_id: taskId,
        worker_id: workerId,
        status: "success",
        result: JSON.stringify({ chunk_count: Object.keys(docs).length }),
      });

      return { status: "success" as const, result: { chunk_count: Object.keys(docs).length } };
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      logger.error({ taskId, error: msg }, "chunk task failed");

      publishCallback(nc, "ke.knowledge_task", {
        task_id: taskId,
        worker_id: workerId,
        status: "fail",
        error_message: msg,
      });

      return { status: "fail" as const, error: msg };
    }
  });

  const taskContext: TaskContext = {
    storage,
    es,
    llm,
    embedding,
    agentConfig: {
      apiUrl: config.agent.apiUrl,
      apiKey: config.agent.apiKey,
      chunkSize: config.agent.chunkSize,
      maxTokenSize: config.agent.maxTokenSize,
      maxWorkers: config.agent.maxWorkers,
      pool: config.agent.pool,
    },
    daemonUrl: config.daemon.url,
    logger,
  };

  const rpcHandlers = handlerRegistry.map((def) => ({
    subject: def.subject,
    name: def.name,
    handler: (payload: any) => def.handler(taskContext, payload),
  }));

  const rpcServer = new RPCServer(nc, rpcHandlers);
  await rpcServer.start(config.rpc.queueGroup);

  let stopping = false;
  const shutdown = async () => {
    if (stopping) {
      logger.warn("forced exit");
      process.exit(1);
    }
    stopping = true;
    logger.info("shutting down...");
    await rpcServer.stop();
    await chunkConsumer.stop();
    await nc.drain();
    await es.close();
    logger.info("worker stopped");
    process.exit(0);
  };

  process.on("SIGTERM", shutdown);
  process.on("SIGINT", shutdown);

  logger.info("starting chunk consumer");
  await chunkConsumer.start();
}
```

- [ ] **步骤 2： 类型检查**

```bash
cd apps/worker && pnpm --filter @corekg/kealgo-worker typecheck
```

- [ ] **步骤 3： 提交**

```bash
git add apps/worker/apps/worker/src/worker-main.ts
git commit -m "feat(worker): integrate RPCServer for NATS request/reply"
```

---

## 任务 9：将 @corekg/workers 和 @corekg/rpc 添加为 worker 应用的依赖

### 需要修改的文件

- `apps/worker/apps/worker/package.json`

### 步骤

- [ ] **步骤 1： 更新 package.json**

将以下内容添加到 `apps/worker/apps/worker/package.json` 依赖项（如果尚未从任务 7 中添加）：

```json
{
  "dependencies": {
    "@corekg/workers": "workspace:*",
    "@corekg/rpc": "workspace:*"
  }
}
```

- [ ] **步骤 2： 安装依赖**

```bash
cd apps/worker && pnpm install
```

- [ ] **步骤 3： 类型检查**

```bash
cd apps/worker && pnpm --filter @corekg/kealgo-worker typecheck
```

- [ ] **步骤 4： 提交**

```bash
git add apps/worker/apps/worker/package.json
git commit -m "chore(worker): add workers and rpc workspace deps"
```

---

## 任务 10：构建验证

### 步骤

- [ ] **步骤 1： 按顺序构建所有修改过的包**

```bash
cd apps/worker && pnpm --filter @corekg/config build
cd apps/worker && pnpm --filter @corekg/search build
cd apps/worker && pnpm --filter @corekg/rpc build
cd apps/worker && pnpm --filter @corekg/workers build
cd apps/worker && pnpm --filter @corekg/kealgo-worker build
```

- [ ] **步骤 2： 确认无 TypeScript 错误**

如果任何构建失败，修复错误后重新构建。

- [ ] **步骤 3： 运行可用的测试**

```bash
cd apps/worker && pnpm --filter @corekg/config test
cd apps/worker && pnpm --filter @corekg/search test
cd apps/worker && pnpm --filter @corekg/rpc test
cd apps/worker && pnpm --filter @corekg/workers test
```

- [ ] **步骤 4： 提交修复**

```bash
git add -A
git commit -m "fix: resolve build/type issues"
```

---

## 任务 11：完善 .env.example 中的所有环境变量

### 需要修改的文件

- `apps/worker/.env.example`

### 步骤

- [ ] **步骤 1： 确认 .env.example 内容完整**

确保 `apps/worker/.env.example` 包含所有必需的变量。该文件应该已在任务 1 中添加了 Agent/Daemon/RPC 变量。确认包含以下：

```
# NATS
NATS_URL=nats://localhost:4222

# Elasticsearch
ES_HOST=https://example.com:58081
ES_USERNAME=elastic
ES_PASSWORD=changeme
ES_POOL_SIZE=10
ES_TIMEOUT_MS=30000

# S3 / MinIO
S3_ENDPOINT_URL=https://example.com:58081
S3_ACCESS_KEY_ID=changeme
S3_SECRET_ACCESS_KEY=changeme
S3_REGION=cn-xian
S3_DEFAULT_BUCKET=test-knownow
S3_PUBLIC_ENDPOINT_URL=

# Text LLM
LLM_API_KEY=changeme
LLM_BASE_URL=https://yygu.cn/v3/llm.chat
LLM_MODEL=deepseek-v3

# Vision LLM
VLLM_API_KEY=changeme
VLLM_BASE_URL=https://yygu.cn/v3/llm.chat
VLLM_MODEL=qwen3-vl-plus

# Embedding
EMBEDDING_API_KEY=changeme
EMBEDDING_BASE_URL=http://example.com:58080/v1
EMBEDDING_MODEL=Qwen3-Embedding-0.6B

# Concurrency
EB_WORKERS=30
LLM_WORKERS=30

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

# Worker
WORKER_ID=
```

- [ ] **步骤 2： 如有修改则提交**

```bash
git add apps/worker/.env.example
git commit -m "docs: complete .env.example with all env vars"
```
