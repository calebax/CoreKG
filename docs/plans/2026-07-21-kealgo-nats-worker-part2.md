# Node.js Worker + NATS JetStream 实施计划（Part 2）

> 本文档是 Part 1 的延续，包含 Task 4-28 的详细 TDD 步骤。

---

## Task 4: S3 Storage Provider (详细步骤)

**Files:**
- Create: `apps/apps/worker/src/storage/s3-provider.ts`
- Create: `apps/apps/worker/src/storage/s3-provider.test.ts`

**Interfaces:**
- Consumes: @aws-sdk/client-s3, @aws-sdk/lib-storage, S3ConfigSchema
- Produces: `StorageProvider`, `UploadResult`, `createS3Provider(config)`

- [ ] **Step 1: 安装依赖**

Run: `cd worker && pnpm add @aws-sdk/client-s3 @aws-sdk/lib-storage mime-types && pnpm add -D @types/mime-types`

- [ ] **Step 2: 写测试**

```typescript
// src/storage/s3-provider.test.ts
import { describe, it, expect } from "vitest";
import { createS3Provider } from "./s3-provider.js";

describe("createS3Provider", () => {
  it("creates provider with valid config", () => {
    const provider = createS3Provider({
      endpointUrl: "https://s3.example.com",
      accessKeyId: "ak",
      secretAccessKey: "sk",
      defaultBucket: "test",
    });
    expect(typeof provider.downloadFile).toBe("function");
    expect(typeof provider.uploadFile).toBe("function");
    expect(typeof provider.uploadDirectory).toBe("function");
    expect(provider.getEndpoint()).toBe("https://s3.example.com");
  });
});
```

- [ ] **Step 3: 运行测试确认失败**

Run: `cd worker && pnpm test -- src/storage/`
Expected: FAIL

- [ ] **Step 4: 实现**

```typescript
// src/storage/s3-provider.ts
import { S3Client } from "@aws-sdk/client-s3";
import { Upload } from "@aws-sdk/lib-storage";
import { createReadStream } from "node:fs";
import { readFile, readdir, stat, writeFile, mkdir } from "node:fs/promises";
import { join, relative, basename } from "node:path";
import mime from "mime-types";
import type { z } from "zod";
import type { S3ConfigSchema } from "../config/schema.js";

type S3Config = z.infer<typeof S3ConfigSchema>;

export interface UploadResult { url: string; key: string; }

export interface StorageProvider {
  downloadFile(url: string, destDir: string, filename?: string): Promise<string>;
  uploadFile(localPath: string, key: string, bucket?: string): Promise<UploadResult>;
  uploadDirectory(localDir: string, basePath: string, bucket?: string): Promise<UploadResult[]>;
  getEndpoint(): string;
}

export function createS3Provider(config: S3Config): StorageProvider {
  const client = new S3Client({
    endpoint: config.endpointUrl,
    credentials: { accessKeyId: config.accessKeyId, secretAccessKey: config.secretAccessKey },
    region: config.region,
    forcePathStyle: true,
  });
  const bucket = config.defaultBucket;
  const pubEndpoint = config.publicEndpointUrl || config.endpointUrl;

  return {
    getEndpoint() { return config.endpointUrl; },

    async downloadFile(url, destDir, filename?) {
      await mkdir(destDir, { recursive: true });
      const name = filename || basename(new URL(url).pathname);
      const dest = join(destDir, name);
      const resp = await fetch(url);
      if (!resp.ok) throw new Error(`Download failed: ${resp.status}`);
      await writeFile(dest, Buffer.from(await resp.arrayBuffer()));
      return dest;
    },

    async uploadFile(localPath, key, overrideBucket?) {
      const ct = mime.lookup(localPath) || "application/octet-stream";
      await new Upload({
        client,
        params: { Bucket: overrideBucket || bucket, Key: key, Body: createReadStream(localPath), ContentType: ct },
      }).done();
      return { url: `${pubEndpoint}/${overrideBucket || bucket}/${key}`, key };
    },

    async uploadDirectory(localDir, basePath, overrideBucket?) {
      const results: UploadResult[] = [];
      const entries = await readdir(localDir, { recursive: true });
      for (const entry of entries) {
        const full = join(localDir, entry as string);
        const s = await stat(full);
        if (!s.isFile()) continue;
        const key = `${basePath}/${relative(localDir, full)}`;
        results.push(await this.uploadFile(full, key, overrideBucket));
      }
      return results;
    },
  };
}
```

- [ ] **Step 5: 运行测试确认通过** — Run: `cd worker && pnpm test -- src/storage/`
- [ ] **Step 6: Commit** — `feat(kealgo): add S3 storage provider`

---

## Task 5: ES VectorStore Provider (详细步骤)

- [ ] **Step 1: 安装** — `pnpm add @elastic/elasticsearch`

- [ ] **Step 2: 写 types.ts**

```typescript
// src/search/types.ts
export interface ChunkDocument {
  forest_id: string; company_id: string; uin: string; file_id: string;
  version: string; file_name: string | null;
  type: "chunk" | "table" | "image" | "video" | "entity" | "file_description" | null;
  tokens: number; chunk_size: number; sequence: number;
  location: unknown | null; yg_location: unknown | null;
  description: string; description_hash: string;
  embedding: number[] | null; image_url: string | null;
  image_embedding: number[] | null; formula: string | null;
  table: string | null; title_level_ids: string[] | null;
  title_level: string[] | null; references: unknown | null;
  graph_info: unknown | null; graph_status: unknown | null;
}
export interface VectorStore {
  upsertChunks(index: string, docs: Record<string, ChunkDocument>): Promise<void>;
  deleteChunksByFileId(index: string, forestId: string, fileId: string, companyId: string): Promise<number>;
}
export interface SearchProvider {
  getById(index: string, id: string): Promise<Record<string, unknown> | null>;
  query(index: string, body: Record<string, unknown>): Promise<Record<string, unknown>[]>;
}
export interface ESProvider { vectorStore: VectorStore; search: SearchProvider; close(): Promise<void>; }
```

- [ ] **Step 3: 写测试** — 验证 `createESProvider(config)` 返回 `{ vectorStore, search, close }`
- [ ] **Step 4: 实现** — ES `Client` + `bulk()` upsert + `deleteByQuery()` + `get()` + `search()`
- [ ] **Step 5: 运行测试确认通过**
- [ ] **Step 6: Commit** — `feat(kealgo): add ES provider`

---

## Task 6: LLM + Embedding Provider (详细步骤)

- [ ] **Step 1: 安装** — `pnpm add ai @ai-sdk/openai-compatible`

- [ ] **Step 2: 写 llm-provider.ts**

```typescript
// src/ai/llm-provider.ts
import { generateText, type CoreMessage } from "ai";
import { createOpenAICompatible } from "@ai-sdk/openai-compatible";
import type { z } from "zod";
import type { LLMConfigSchema } from "../config/schema.js";

type LLMConfig = z.infer<typeof LLMConfigSchema>;

export interface LLMChatOptions {
  model?: string; temperature?: number; systemPrompt?: string;
  history?: Array<{ role: "system" | "user" | "assistant"; content: string }>;
  contentParts?: Array<{ type: "text"; text: string } | { type: "image_url"; image_url: { url: string } }>;
}

export interface LLMProvider {
  chat(prompt: string, options?: LLMChatOptions): Promise<string>;
}

export function createLLMProvider(config: LLMConfig): LLMProvider {
  const provider = createOpenAICompatible({ name: "custom-llm", baseURL: config.baseUrl, apiKey: config.apiKey });
  return {
    async chat(prompt, opts = {}) {
      const messages: CoreMessage[] = [];
      if (opts.systemPrompt) messages.push({ role: "system", content: opts.systemPrompt });
      if (opts.history) messages.push(...opts.history);
      if (opts.contentParts) {
        messages.push({ role: "user", content: opts.contentParts as any });
      } else {
        messages.push({ role: "user", content: prompt });
      }
      const { text } = await generateText({
        model: provider(opts.model || config.model),
        messages, temperature: opts.temperature ?? 1,
        abortSignal: AbortSignal.timeout(config.timeoutMs),
      });
      return text;
    },
  };
}
```

- [ ] **Step 3: 写 embedding-provider.ts**

```typescript
// src/ai/embedding-provider.ts
import { embed, embedMany } from "ai";
import { createOpenAICompatible } from "@ai-sdk/openai-compatible";
import type { z } from "zod";
import type { EmbeddingConfigSchema } from "../config/schema.js";

type EmbeddingConfig = z.infer<typeof EmbeddingConfigSchema>;

export interface EmbeddingProvider {
  embed(text: string, model?: string): Promise<number[] | null>;
  embedBatch(texts: string[], options?: { concurrency?: number; model?: string }): Promise<Array<number[] | null>>;
}

export function createEmbeddingProvider(config: EmbeddingConfig): EmbeddingProvider {
  const provider = createOpenAICompatible({ name: "custom-emb", baseURL: config.baseUrl, apiKey: config.apiKey });
  return {
    async embed(text, modelName?) {
      if (!text.trim()) return null;
      const { embedding } = await embed({
        model: provider.embedding(modelName || config.model),
        value: text, abortSignal: AbortSignal.timeout(config.timeoutMs),
      });
      return embedding;
    },
    async embedBatch(texts, opts = {}) {
      const valid = texts.map((t, i) => ({ id: i, text: t })).filter(t => t.text.trim());
      if (!valid.length) return texts.map(() => null);
      const { embeddings } = await embedMany({
        model: provider.embedding(opts.model || config.model),
        values: valid.map(t => t.text), maxRetries: 2,
      });
      const results: Array<number[] | null> = new Array(texts.length).fill(null);
      for (let i = 0; i < valid.length; i++) results[valid[i].id] = embeddings[i];
      return results;
    },
  };
}
```

- [ ] **Step 4: 写两个 test 文件** — 验证 provider 创建 + 方法存在
- [ ] **Step 5: 运行测试确认通过**
- [ ] **Step 6: Commit** — `feat(kealgo): add LLM and embedding providers`

---

## Task 7: Tokenizer (详细步骤)

- [ ] **Step 1: 安装** — `pnpm add js-tiktoken`

- [ ] **Step 2: 写测试**

```typescript
// src/chunker/tokenizer.test.ts
import { describe, it, expect } from "vitest";
import { countTokens } from "./tokenizer.js";
describe("countTokens", () => {
  it("counts English", () => { expect(countTokens("Hello world")).toBeGreaterThan(0); });
  it("counts Chinese", () => { expect(countTokens("你好世界")).toBeGreaterThan(0); });
  it("empty = 0", () => { expect(countTokens("")).toBe(0); });
  it("matches Python tiktoken cl100k_base", () => {
    // 已知 Python countTokens("Hello world") == 2
    expect(countTokens("Hello world")).toBe(2);
  });
});
```

- [ ] **Step 3: 运行确认失败**
- [ ] **Step 4: 实现**

```typescript
// src/chunker/tokenizer.ts
import { getEncoding, type Tiktoken } from "js-tiktoken";
let encoder: Tiktoken | null = null;
function getEncoder(): Tiktoken {
  if (!encoder) encoder = getEncoding("cl100k_base");
  return encoder;
}
export function countTokens(text: string): number {
  if (!text) return 0;
  return getEncoder().encode(text).length;
}
```

- [ ] **Step 5: 运行确认通过**
- [ ] **Step 6: Commit** — `feat(kealgo): add tiktoken tokenizer`

---

## Task 8: Text Preprocessing

- [ ] **Step 1: 写测试** — 验证 removeEmail / removeUrl / removeEmptyLine
- [ ] **Step 2: 实现** — 正则替换，参考 Python `utils.py:simple_clean()`
- [ ] **Step 3: TDD 循环** — Commit: `feat(kealgo): add text preprocessor`

---

## Task 9: markdown-it AST Parser

- [ ] **Step 1: 安装** — `pnpm add markdown-it && pnpm add -D @types/markdown-it`
- [ ] **Step 2: 写测试** — 验证 parseMarkdownAst 返回 tokens
- [ ] **Step 3: 实现**

```typescript
// src/chunker/ast-parser.ts
import MarkdownIt from "markdown-it";
const md = new MarkdownIt();
export function parseMarkdownTokens(content: string) {
  return md.parse(content, {});
}
```

- [ ] **Step 4: Commit** — `feat(kealgo): add markdown AST parser`

---

## Task 10: 切块策略 (11 种)

### 统一接口

```typescript
// src/chunker/strategy.ts
import type { LLMProvider } from "../ai/llm-provider.js";

export interface ChunkMeta {
  headers?: Record<string, string>;
  metadata?: Record<string, unknown>;
}

export interface ChunkSplitResult {
  chunks: string[];
  metas: ChunkMeta[];
}

export interface ChunkStrategyOptions {
  chunkTokenNum: number;
  minChunkTokens: number;
  splitLevel: number;
  overlapRatio: number;
  regexPattern: string;
  delimiter: string;
  enableHeadingInContent: boolean;
}

export interface ChunkStrategy {
  readonly name: string;
  split(content: string, options: ChunkStrategyOptions): ChunkSplitResult;
}

export interface AsyncChunkStrategy extends ChunkStrategy {
  splitAsync(content: string, options: ChunkStrategyOptions, deps: { llm: LLMProvider }): Promise<ChunkSplitResult>;
}
```

### 每种策略 TDD 模板

对每种策略 (10a-10k)：

- [ ] Step 1: 写单元测试 (用小 markdown 片段验证切块产出)
- [ ] Step 2: 运行确认失败
- [ ] Step 3: 从 Python 移植实现
- [ ] Step 4: 运行确认通过
- [ ] Step 5: Commit: `feat(kealgo): add {name} chunk strategy`

**优先级排序（先实现最常用的）：**
1. `basic` (10a) — 最简单，可快速验证框架
2. `smart` (10b) — 默认策略
3. `resume` (10g) — 最简单
4. `auto` (10k) — 依赖其他策略
5. 其余按需

---

## Task 11: ChunkStrategyRegistry

```typescript
// src/chunker/registry.ts
import type { ChunkStrategy, AsyncChunkStrategy } from "./strategy.js";

const strategies = new Map<string, ChunkStrategy | AsyncChunkStrategy>();

export function registerStrategy(s: ChunkStrategy | AsyncChunkStrategy) {
  strategies.set(s.name, s);
}

export function resolveStrategy(name: string): ChunkStrategy | AsyncChunkStrategy {
  const s = strategies.get(name);
  if (!s) throw new Error(`Unknown chunk strategy: ${name}`);
  return s;
}

export function listStrategies(): string[] { return [...strategies.keys()]; }
```

- [ ] TDD: test register/resolve/list → impl → commit

---

## Task 12: ChunkerService

- 编排完整管线: preprocess → split → vision enhance → table enhance → embed → format docs
- 参考 Python `chunk.py:chunk_process()`
- 注入所有 providers via 构造函数
- TDD: 单元测试用 mock providers
- Commit: `feat(kealgo): add chunker service pipeline`

---

## Task 13-14: Enhancers

### Task 13: VisionEnhancer

- 使用 `LLMProvider.chat()` with `contentParts` (base64 image)
- 参考 Python `image_vision_enhancer.py` + `image_context_extractor.py`
- Commit: `feat(kealgo): add vision enhancer`

### Task 14: TableEnhancer

- 使用 `LLMProvider.chat()` text-only
- 参考 Python `table_enhancer.py`
- Commit: `feat(kealgo): add table enhancer`

---

## Task 15-17: Parser Worker

### Task 15: MinerU API Client

- HTTP POST multipart/form-data to MinerU
- 接收 ZIP → 解压 → 解析 content_list.json
- 参考 Python `analyser_process.py:process_pdf_task_api()`
- Commit: `feat(kealgo): add MinerU API client`

### Task 16: JSON-to-Markdown Converter

- 纯数据转换，~521L Python → 等量 TS
- 参考 Python `json_to_md.py`
- Commit: `feat(kealgo): add json-to-markdown converter`

### Task 17: Non-PDF Handlers + Video

- txt/md/csv/json: 直接 readFile 返回
- video: `child_process.execSync('ffmpeg ...')` 替代 OpenCV
- Commit: `feat(kealgo): add non-PDF handlers and video extractor`

---

## Task 18: NATS JetStream Client

- [ ] **Step 1: 安装** — `pnpm add @nats-io/nats-core @nats-io/jetstream`

- [ ] **Step 2: 实现**

```typescript
// src/nats/client.ts
import { connect } from "@nats-io/nats-core";
import { jetstream } from "@nats-io/jetstream";
import type { AppConfig } from "../config/schema.js";

export async function createNATSClient(config: AppConfig) {
  const nc = await connect({ servers: config.nats.url });
  const js = jetstream(nc);

  // Ensure stream exists
  try {
    await js.streams.info(config.nats.stream);
  } catch {
    await js.streams.add({
      name: config.nats.stream,
      subjects: ["core.task.>"],
      retention: "workqueue",
    });
  }

  return { nc, js };
}
```

- [ ] **Step 3: 写测试** — 连接测试 (需 NATS server)
- [ ] **Step 4: Commit** — `feat(kealgo): add NATS JetStream client`

---

## Task 19: Task Consumer

```typescript
// src/nats/consumer.ts
import type { JetStreamClient } from "@nats-io/jetstream";
import { AckPolicy, DeliverPolicy } from "@nats-io/jetstream";
import { createLogger } from "../logger/index.js";

const logger = createLogger("nats-consumer");

export interface TaskHandler {
  (taskId: string, payload: unknown): Promise<{ status: "success" | "fail"; result?: unknown; error?: string }>;
}

export interface TaskConsumerOptions {
  stream: string;
  subject: string;
  durableName: string;
  maxAckPending?: number;
  ackWaitMs?: number;
}

export class TaskConsumer {
  private stopped = false;

  constructor(
    private js: JetStreamClient,
    private options: TaskConsumerOptions,
    private handler: TaskHandler,
  ) {}

  async start() {
    const consumer = await this.js.consumers.getOrAdd(this.options.stream, {
      durable_name: this.options.durableName,
      filter_subject: this.options.subject,
      ack_policy: AckPolicy.Explicit,
      ack_wait: (this.options.ackWaitMs ?? 300_000) * 1_000_000,
      max_ack_pending: this.options.maxAckPending ?? 2,
    });

    logger.info({ subject: this.options.subject }, "consumer started");

    while (!this.stopped) {
      const iter = await consumer.consume({ expires: 30_000, batch: 1 });
      for await (const msg of iter) {
        if (this.stopped) { msg.nak(); break; }
        try {
          const task = msg.json<{ task_id: string; payload: unknown }>();
          const result = await this.handler(task.task_id, task.payload);
          msg.ack();
          logger.info({ taskId: task.task_id, status: result.status }, "task completed");
        } catch (err) {
          logger.error({ err }, "task failed, nak");
          msg.nak();
        }
      }
    }
  }

  async stop() {
    this.stopped = true;
  }
}
```

- [ ] TDD cycle → commit: `feat(kealgo): add NATS task consumer`

---

## Task 20: Task Protocol Types

Zod schemas matching Go TaskPayload (from `apps/keworker/internal/services/` analysis):

```typescript
// src/nats/types.ts
import { z } from "zod";

export const LLMModelConfigSchema = z.object({
  api_key: z.string(), base_url: z.string(), model_name: z.string(),
  provider: z.string().optional(),
});

export const SplitConfigSchema = z.object({
  split_mode: z.enum(["auto", "rule", "smart", "basic", "advanced", "title",
    "strict_regex", "slide", "resume", "paper", "laws", "llm"]).default("smart"),
  chunk_size: z.number().optional(),
  chunk_token_num: z.number().optional(),
  split_mark: z.union([z.string(), z.array(z.string())]).optional(),
  regex_pattern: z.string().optional(),
  split_overlap: z.number().optional(),
  overlap_ratio: z.number().optional(),
  min_chunk_tokens: z.number().default(10),
  split_level: z.number().default(2),
  enable_heading_in_content: z.boolean().default(false),
  preprocessing_rules: z.object({
    remove_email: z.boolean().default(true),
    remove_url: z.boolean().default(true),
    remove_empty_line: z.boolean().default(true),
  }).default({}),
  llm_enabled: z.boolean().optional(),
  llm_model: z.string().optional(),
  vllm_enabled: z.boolean().optional(),
  vllm_model: z.string().optional(),
  image_width: z.number().optional(),
  eb_max_concurrency: z.number().optional(),
  llm_max_concurrency: z.number().optional(),
});

export const TaskPayloadSchema = z.object({
  task_type: z.string(),
  file_id: z.union([z.string(), z.number()]),
  file_url: z.string(),
  filename: z.string().optional(),
  file_name: z.string().optional(),
  file_ext: z.string().optional(),
  company_id: z.union([z.string(), z.number()]),
  forest_id: z.union([z.string(), z.number()]),
  uin: z.union([z.string(), z.number()]),
  es_index: z.string().optional(),
  bucket: z.string().optional(),
  upload_path: z.string().optional(),
  storage_path: z.string().optional(),
  llm: LLMModelConfigSchema.optional(),
  vllm: LLMModelConfigSchema.optional(),
  embedding: LLMModelConfigSchema.optional(),
  split_config: SplitConfigSchema.optional(),
  es: z.object({
    index_name: z.string(), addr: z.string(),
    username: z.string(), password: z.string(),
  }).optional(),
});

export type TaskPayload = z.infer<typeof TaskPayloadSchema>;

export const TaskCallbackSchema = z.object({
  task_id: z.union([z.string(), z.number()]),
  worker_id: z.string(),
  status: z.enum(["success", "fail", "cancel"]),
  result: z.string().optional(),
  error_message: z.string().optional(),
});
```

- [ ] TDD: parse valid/invalid payloads → commit: `feat(kealgo): add task protocol types`

---

## Task 21: Callback Publisher

```typescript
// src/nats/callback.ts
import type { NATSConnection } from "@nats-io/nats-core";
import { TaskCallbackSchema } from "./types.js";

export async function publishCallback(
  nc: NATSConnection,
  taskType: string,
  callback: z.infer<typeof TaskCallbackSchema>,
) {
  const validated = TaskCallbackSchema.parse(callback);
  const subject = `core.task.callback.${taskType}`;
  nc.publish(subject, JSON.stringify(validated));
}
```

- [ ] TDD → commit: `feat(kealgo): add callback publisher`

---

## Task 22: Signal Handler + Worker Main

```typescript
// src/worker-main.ts
import { createLogger } from "./logger/index.js";
import { loadConfig } from "./config/index.js";
import { createNATSClient } from "./nats/client.js";
import { TaskConsumer } from "./nats/consumer.js";
import { publishCallback } from "./nats/callback.js";
// ... providers

const logger = createLogger("kealgo");

async function main() {
  const config = loadConfig();
  const { nc, js } = await createNATSClient(config);
  const workerId = config.workerId || `${process.hostname}-${process.pid}`;

  // Create providers
  // Create services
  // Create consumers for analyser + chunker

  const chunkConsumer = new TaskConsumer(js, {
    stream: config.nats.stream,
    subject: config.nats.subjects.chunker,
    durableName: "kealgo-chunker",
  }, async (taskId, payload) => {
    // Parse payload, run ChunkerService, callback
    return { status: "success" };
  });

  const stopSignal = { triggered: false };
  process.on("SIGTERM", () => { stopSignal.triggered = true; });
  process.on("SIGINT", () => { stopSignal.triggered = true; });

  logger.info("starting consumers...");
  await chunkConsumer.start();

  await nc.drain();
  logger.info("worker stopped");
}

main().catch((err) => { logger.error(err, "fatal"); process.exit(1); });
```

- [ ] TDD → commit: `feat(kealgo): add worker main with signal handling`

---

## Task 23-28: 集成 + 部署

### Task 23: ChunkWorker 完整管线

组装 ChunkerService + providers + NATS consumer，端到端测试
Commit: `feat(kealgo): wire chunk worker pipeline`

### Task 24: AnalyserWorker 完整管线

组装 MinerU + json_to_md + S3 + NATS consumer
Commit: `feat(kealgo): wire analyser worker pipeline`

### Task 25: Dockerfile

```dockerfile
# apps/apps/worker/script/Dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY apps/apps/worker/package.json apps/apps/worker/pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY apps/apps/worker/ ./
RUN pnpm build

FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./
CMD ["node", "dist/index.js"]
```

Commit: `chore(kealgo): add Dockerfile`

### Task 26: Go 侧 NATS 桥接

**Files:**
- Create: `pkgs/task/nats_bridge.go`

```go
// pkgs/task/nats_bridge.go
package task

import (
    "encoding/json"
    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

type NATSBridge struct {
    js      jetstream.JetStream
    stream  string
}

func NewNATSBridge(nc *nats.Conn, stream string) (*NATSBridge, error) {
    js, err := jetstream.New(nc)
    if err != nil { return nil, err }
    return &NATSBridge{js: js, stream: stream}, nil
}

func (b *NATSBridge) PublishTask(taskType string, payload interface{}) error {
    data, err := json.Marshal(payload)
    if err != nil { return err }
    subject := "core.task." + taskType
    _, err = b.js.Publish(subject, data)
    return err
}
```

修改 `pkgs/task/task_queue.go` 的 `PushTaskQueue()`：在现有 Redis XAdd 之后追加 `natsBridge.PublishTask()`

修改 keparser 的路由处理：新增 NATS consumer 订阅 `core.task.callback.*`，收到消息后调用现有 `HandleTaskCallback()` 逻辑

Commit: `feat(task): add NATS JetStream bridge alongside Redis`

### Task 27: 环境变量

Create: `apps/apps/worker/.env.example`

```env
NATS_URL=nats://localhost:4222
ES_HOST=https://example.com:58081
ES_USERNAME=elastic
ES_PASSWORD=xxx
S3_ENDPOINT_URL=https://example.com:58081
S3_ACCESS_KEY_ID=xxx
S3_SECRET_ACCESS_KEY=xxx
S3_DEFAULT_BUCKET=test-knownow
LLM_API_KEY=yg-xxx
LLM_BASE_URL=https://yygu.cn/v3/llm.chat
LLM_MODEL=deepseek-v3
VLLM_MODEL=qwen3-vl-plus
EMBEDDING_API_KEY=123
EMBEDDING_BASE_URL=http://example.com:58080/v1
EMBEDDING_MODEL=Qwen3-Embedding-0.6B
EB_WORKERS=30
LLM_WORKERS=30
```

Commit: `chore(kealgo): add .env.example`

### Task 28: 双跑验证方案

Create: `apps/apps/worker/docs/migration-verification.md`

- 新旧 Worker 并行消费不同 NATS subject
- 对比脚本: 相同文档 + split_config → diff chunks
- Go worker `workerServerURL` 切换说明
- Commit: `docs(kealgo): add migration verification plan`

---

## 实施建议

1. **并行开发**: Phase 1-4 可由一人完成，Phase 5-6 可分配不同人
2. **先跑通 basic 策略**: Task 10a 完成后立即接 Task 12 + Task 23，验证端到端
3. **Go 桥接最后做**: Task 26 风险最高（改动 Go 代码），放在最后
4. **Python 对比测试**: 每个策略完成后，用相同输入对比 Python/TS 输出
5. **渐进式切换**: 先在测试环境双跑，确认 chunk 一致性后再切换生产
