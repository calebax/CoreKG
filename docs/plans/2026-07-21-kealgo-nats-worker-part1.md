# Node.js Worker + NATS JetStream 实施计划（Part 1）

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 corekgtwo Go 单体仓库内创建独立 Node.js Worker（`apps/apps/worker/`），将 corekg-pipeline 的 Python 解析/切块逻辑迁移为 TypeScript，通过 NATS JetStream 与 Go 平台对接。

**Architecture:** Node.js Worker 作为独立子项目放在 `apps/apps/worker/`。通过 NATS JetStream Pull Consumer 消费任务（替代 HTTP GetPendingTask），处理后通过 NATS 发布回调（替代 HTTP TaskCallBack）。Go 侧在 `pkgs/task/` 新增 NATS 桥接发布器，将 Redis Streams 触发改为同时向 NATS 发布。Python corekg-pipeline 保留过渡，两套 Worker 可并行。

**Tech Stack:** Node.js 22 LTS, TypeScript 5.x, pnpm, Vitest, Pino, Zod, @nats-io/nats-core + @nats-io/jetstream, @aws-sdk/client-s3, @elastic/elasticsearch, ai (Vercel AI SDK), markdown-it, js-tiktoken, sharp

## Global Constraints

- Node.js >= 22 LTS，pnpm，Vitest，Pino，Zod
- NATS: @nats-io/nats-core + @nats-io/jetstream v3.x
- LLM SDK: Vercel AI SDK (`ai` + `@ai-sdk/openai-compatible`)
- Go 侧改动最小化：仅新增 NATS 发布器，不改现有 Redis/HTTP 逻辑
- 代码路径: `/root/go/src/corekgtwo/apps/worker/`
- 所有 TS 文件必须通过 `tsc --noEmit` 和 `vitest run`

## 总览：28 个 Task，7 个 Phase

| Phase | Tasks | 产出 |
|-------|-------|------|
| 1. 脚手架 | 1-3 | kealgo 子项目 + Logger + Config |
| 2. Provider Packages | 4-6 | S3 / ES / LLM+Embedding |
| 3. Chunker 核心 | 7-12 | Tokenizer + 11 种切块策略 + Registry |
| 4. Enhancers | 13-14 | Vision / Table 增强 |
| 5. Parser Worker | 15-17 | MinerU 调用 + json_to_md + 非 PDF |
| 6. NATS Worker 框架 | 18-22 | JetStream Consumer + Task Protocol + Signal + Bridge |
| 7. 集成 + 部署 | 23-28 | Pipeline 串联 + Dockerfile + Go 桥接 + 双跑验证 |

---

## Phase 1: 脚手架 + 共享 Packages

### Task 1: 初始化 worker 子项目

**Files:**
- Create: `apps/apps/worker/package.json`
- Create: `apps/apps/worker/tsconfig.json`
- Create: `apps/apps/worker/vitest.config.ts`
- Create: `apps/apps/worker/src/index.ts`

**Interfaces:**
- Produces: 可运行的空 Node.js 项目，`pnpm build && pnpm test && pnpm typecheck` 通过

- [ ] **Step 1: 创建 package.json**

```json
{
  "name": "@corekg/kealgo",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "build": "tsc",
    "dev": "tsx watch src/index.ts",
    "start": "node dist/index.js",
    "test": "vitest run",
    "test:watch": "vitest",
    "typecheck": "tsc --noEmit"
  },
  "engines": { "node": ">=22.0.0" }
}
```

- [ ] **Step 2: 安装基础依赖**

Run: `cd worker && pnpm add pino zod && pnpm add -D typescript vitest tsx @types/node`

- [ ] **Step 3: 创建 tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2022", "module": "Node16", "moduleResolution": "Node16",
    "lib": ["ES2022"], "outDir": "dist", "rootDir": "src",
    "strict": true, "esModuleInterop": true, "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true, "resolveJsonModule": true,
    "declaration": true, "declarationMap": true, "sourceMap": true
  },
  "include": ["src/**/*"], "exclude": ["node_modules", "dist"]
}
```

- [ ] **Step 4: 创建 vitest.config.ts**

```typescript
import { defineConfig } from "vitest/config";
export default defineConfig({
  test: { globals: true, environment: "node", include: ["src/**/*.test.ts"] },
});
```

- [ ] **Step 5: 创建 src/index.ts**

```typescript
import pino from "pino";
const logger = pino({ level: "info" });
async function main() { logger.info("kealgo worker starting..."); }
main().catch((err) => { logger.error(err, "fatal"); process.exit(1); });
```

- [ ] **Step 6: 验证** — Run: `cd worker && pnpm build && pnpm test && pnpm typecheck`
- [ ] **Step 7: Commit** — `feat(kealgo): init Node.js worker subproject`

---

### Task 2: Logger Package

**Files:**
- Create: `apps/apps/worker/src/logger/index.ts`
- Create: `apps/apps/worker/src/logger/index.test.ts`

**Interfaces:**
- Produces: `createLogger(name: string, level?: string): pino.Logger`

- [ ] **Step 1: 写测试**

```typescript
import { describe, it, expect } from "vitest";
import { createLogger } from "./index.js";
describe("createLogger", () => {
  it("creates logger with name", () => {
    const l = createLogger("test");
    expect(typeof l.info).toBe("function");
    expect(l.level).toBe("info");
  });
  it("respects custom level", () => {
    expect(createLogger("t", "debug").level).toBe("debug");
  });
});
```

- [ ] **Step 2: 运行测试确认失败** — Run: `cd worker && pnpm test -- src/logger/`
- [ ] **Step 3: 实现**

```typescript
import pino from "pino";
export function createLogger(name: string, level = "info"): pino.Logger {
  return pino({ name, level });
}
```

- [ ] **Step 4: 运行测试确认通过** — Run: `cd worker && pnpm test -- src/logger/`
- [ ] **Step 5: Commit** — `feat(kealgo): add pino logger`

---

### Task 3: Config Package (Zod)

**Files:**
- Create: `apps/apps/worker/src/config/schema.ts`
- Create: `apps/apps/worker/src/config/index.ts`
- Create: `apps/apps/worker/src/config/index.test.ts`

**Interfaces:**
- Produces: `loadConfig(): AppConfig`, `AppConfig` type, all Zod schemas

- [ ] **Step 1: 写 schema.ts**

```typescript
import { z } from "zod";
export const NATSConfigSchema = z.object({
  url: z.string().default("nats://localhost:4222"),
  stream: z.string().default("CORE_TASKS"),
  subjects: z.object({
    analyser: z.string().default("core.task.ke.prase_pdf_task"),
    chunker: z.string().default("core.task.ke.knowledge_task"),
    callback: z.string().default("core.task.callback.*"),
  }),
});
export const ESConfigSchema = z.object({
  host: z.string().url(), username: z.string(), password: z.string(),
  poolSize: z.number().int().positive().default(10),
  requestTimeoutMs: z.number().int().positive().default(30000),
});
export const S3ConfigSchema = z.object({
  endpointUrl: z.string().url(), accessKeyId: z.string(),
  secretAccessKey: z.string(), region: z.string().default("us-east-1"),
  defaultBucket: z.string(), publicEndpointUrl: z.string().url().optional(),
});
export const LLMConfigSchema = z.object({
  apiKey: z.string(), baseUrl: z.string().url(),
  model: z.string().default("deepseek-v3"),
  timeoutMs: z.number().int().positive().default(300000),
});
export const EmbeddingConfigSchema = z.object({
  apiKey: z.string(), baseUrl: z.string().url(),
  model: z.string().default("Qwen3-Embedding-0.6B"),
  timeoutMs: z.number().int().positive().default(30000),
});
export const VLLMConfigSchema = z.object({
  apiKey: z.string(), baseUrl: z.string().url(),
  model: z.string().default("qwen3-vl-plus"),
  timeoutMs: z.number().int().positive().default(300000),
});
export const ConcurrencyConfigSchema = z.object({
  embeddingWorkers: z.number().int().positive().default(30),
  llmWorkers: z.number().int().positive().default(30),
});
export const AppConfigSchema = z.object({
  nats: NATSConfigSchema, es: ESConfigSchema, s3: S3ConfigSchema,
  llm: LLMConfigSchema, vllm: VLLMConfigSchema,
  embedding: EmbeddingConfigSchema, concurrency: ConcurrencyConfigSchema,
  workerId: z.string().optional(),
});
export type AppConfig = z.infer<typeof AppConfigSchema>;
```

- [ ] **Step 2: 写 index.ts** — `loadConfig()` 从 `process.env` 读取并 `AppConfigSchema.parse()`

```typescript
import { AppConfigSchema, type AppConfig } from "./schema.js";
export { AppConfigSchema } from "./schema.js";
export type { AppConfig } from "./schema.js";
export function loadConfig(): AppConfig {
  return AppConfigSchema.parse({
    nats: { url: process.env.NATS_URL, stream: process.env.NATS_STREAM },
    es: { host: process.env.ES_HOST, username: process.env.ES_USERNAME, password: process.env.ES_PASSWORD },
    s3: { endpointUrl: process.env.S3_ENDPOINT_URL, accessKeyId: process.env.S3_ACCESS_KEY_ID, secretAccessKey: process.env.S3_SECRET_ACCESS_KEY, region: process.env.S3_REGION, defaultBucket: process.env.S3_DEFAULT_BUCKET, publicEndpointUrl: process.env.S3_PUBLIC_ENDPOINT_URL },
    llm: { apiKey: process.env.LLM_API_KEY, baseUrl: process.env.LLM_BASE_URL, model: process.env.LLM_MODEL },
    vllm: { apiKey: process.env.VLLM_API_KEY || process.env.LLM_API_KEY, baseUrl: process.env.VLLM_BASE_URL || process.env.LLM_BASE_URL, model: process.env.VLLM_MODEL },
    embedding: { apiKey: process.env.EMBEDDING_API_KEY, baseUrl: process.env.EMBEDDING_BASE_URL, model: process.env.EMBEDDING_MODEL },
    concurrency: { embeddingWorkers: Number(process.env.EB_WORKERS) || undefined, llmWorkers: Number(process.env.LLM_WORKERS) || undefined },
    workerId: process.env.WORKER_ID,
  });
}
```

- [ ] **Step 3: 写测试** — 验证 valid config 解析 + defaults + invalid URL 拒绝
- [ ] **Step 4: 运行测试确认通过** — Run: `cd worker && pnpm test -- src/config/`
- [ ] **Step 5: Commit** — `feat(kealgo): add zod-validated config`

---

## Phase 2: Provider Packages

### Task 4: S3 Storage Provider

**Files:**
- Create: `apps/apps/worker/src/storage/s3-provider.ts`
- Create: `apps/apps/worker/src/storage/s3-provider.test.ts`

**Interfaces:**
- Consumes: @aws-sdk/client-s3, @aws-sdk/lib-storage
- Produces: `StorageProvider { downloadFile, uploadFile, uploadDirectory, getEndpoint }` + `createS3Provider(config)`

- [ ] **Step 1: 安装** — `pnpm add @aws-sdk/client-s3 @aws-sdk/lib-storage mime-types && pnpm add -D @types/mime-types`
- [ ] **Step 2: 写测试** — 验证 provider 创建 + 方法存在
- [ ] **Step 3: 运行确认失败**
- [ ] **Step 4: 实现** — `S3Client` + `Upload` + `fetch` download + `readdir` recursive upload (参考 Python `s3_minio.py`)
- [ ] **Step 5: 运行确认通过**
- [ ] **Step 6: Commit** — `feat(kealgo): add S3 storage provider`

### Task 5: ES VectorStore Provider

**Files:**
- Create: `apps/apps/worker/src/search/types.ts` — `ChunkDocument`, `VectorStore`, `SearchProvider` 接口
- Create: `apps/apps/worker/src/search/es-provider.ts` — `createESProvider(config): ESProvider`
- Create: `apps/apps/worker/src/search/es-provider.test.ts`

**Interfaces:**
- Produces: `ESProvider { vectorStore, search, close() }`

- [ ] **Step 1: 安装** — `pnpm add @elastic/elasticsearch`
- [ ] **Step 2-6: TDD 循环** — types → test → impl → verify → commit

### Task 6: LLM + Embedding Provider (Vercel AI SDK)

**Files:**
- Create: `apps/apps/worker/src/ai/llm-provider.ts` — `LLMProvider { chat() }` + `createLLMProvider(config)`
- Create: `apps/apps/worker/src/ai/embedding-provider.ts` — `EmbeddingProvider { embed(), embedBatch() }` + `createEmbeddingProvider(config)`
- Create: tests for both

**Interfaces:**
- Consumes: `ai`, `@ai-sdk/openai-compatible`

- [ ] **Step 1: 安装** — `pnpm add ai @ai-sdk/openai-compatible`
- [ ] **Step 2-8: TDD 循环** — 两个 provider 各写 test + impl

---

## Phase 3: Chunker 核心算法

### Task 7: Tokenizer (js-tiktoken)

- `countTokens(text: string): number` — 使用 `getEncoding("cl100k_base")` 单例
- 参考: Python `knowchunk.py:27-46`

### Task 8: Text Preprocessing

- `preprocessText(md, options): string` — remove email/url/empty lines
- 参考: Python `utils.py:simple_clean()`

### Task 9: markdown-it AST Parser

- `parseMarkdownAst(md: string): SyntaxTreeNode` — 返回 markdown-it token tree
- 参考: Python `knowchunk.py` 的 `MarkdownIt().render()` 调用
- markdown-it JS 是原版，API 同 Python 移植版

### Task 10: 切块策略 (11 种)

每种策略一个文件，实现统一接口：

```typescript
interface ChunkStrategy {
  readonly name: string;
  split(content: string, options: ChunkStrategyOptions): ChunkSplitResult;
}
interface AsyncChunkStrategy extends ChunkStrategy {
  splitAsync(content: string, options: ChunkStrategyOptions, deps: { llm: LLMProvider }): Promise<ChunkSplitResult>;
}
```

**Files (11 strategies):**

| # | File | Python Source | 复杂度 |
|---|------|---------------|--------|
| 10a | `strategies/basic.ts` | `split_markdown_to_chunks` ~75L | LOW |
| 10b | `strategies/smart.ts` | `split_markdown_to_chunks_smart` ~95L | MED |
| 10c | `strategies/advanced.ts` | `split_markdown_to_chunks_advanced` ~65L | MED |
| 10d | `strategies/title.ts` | `split_markdown_to_chunks_title` ~55L | LOW |
| 10e | `strategies/strict-regex.ts` | `split_markdown_to_chunks_strict_regex` ~45L | LOW |
| 10f | `strategies/slide.ts` | `split_markdown_to_chunks_slide` ~45L | LOW |
| 10g | `strategies/resume.ts` | `split_markdown_to_chunks_resume` ~10L | TRIVIAL |
| 10h | `strategies/paper.ts` | `split_markdown_to_chunks_paper` ~90L | MED |
| 10i | `strategies/laws.ts` | `split_markdown_to_chunks_laws` ~150L | HIGH |
| 10j | `strategies/llm.ts` | `split_markdown_to_chunks_llm` ~90L | MED (async) |
| 10k | `strategies/auto.ts` | auto mode dispatch ~40L | MED (async) |

每个策略 TDD: test → impl → verify → commit

### Task 11: ChunkStrategyRegistry

- `resolve(name): ChunkStrategy | AsyncChunkStrategy`
- `listStrategies(): string[]`

### Task 12: ChunkerService (Pipeline 编排)

- `ChunkerService.process(options): Promise<Record<string, ChunkDocument>>`
- 编排: download MD → preprocess → split → vision enhance → table enhance → embed → format ES docs
- 参考: Python `chunk.py:184-406`

---

## Phase 4: Enhancers

### Task 13: Vision Enhancer

- `VisionEnhancer.enhance(chunks, md, llm): Promise<EnhancedChunk[]>`
- 参考: Python `image_vision_enhancer.py` + `image_context_extractor.py`
- 使用 `LLMProvider.chat()` with vision contentParts

### Task 14: Table Enhancer

- `TableEnhancer.enhance(chunks, llm): Promise<EnhancedChunks>`
- 参考: Python `table_enhancer.py`

---

## Phase 5: Parser Worker

### Task 15: MinerU API Client

- `callMinerU(filePath, apiUrl): Promise<{ contentList, images, outputDir }>`
- HTTP POST multipart → ZIP 流 → 解压 → 解析 content_list.json
- 参考: Python `analyser_process.py`

### Task 16: JSON-to-Markdown Converter

- `convertContentListToMarkdown(contentList, options): string`
- 参考: Python `json_to_md.py` (~521L)

### Task 17: Non-PDF Handlers

- txt/md/csv/json 直接透传
- 视频关键帧: `ffmpeg` CLI 替代 OpenCV
- 参考: Python `others/content_process.py` + `video/video_process.py`

---

## Phase 6: NATS Worker 框架

### Task 18: NATS JetStream Client

- `createNATSClient(url): Promise<NATSConnection>`
- Stream 初始化: create/update `CORE_TASKS` stream (retention: WorkQueue)
- 安装: `pnpm add @nats-io/nats-core @nats-io/jetstream`

### Task 19: Task Consumer (Pull Subscribe)

- `TaskConsumer` class: `consume(subject, handler, options)` → long-pull fetch loop
- Graceful shutdown: `consumer.close()` + `drain()`
- 参考: Go `pkgs/task/task_queue.go` 的 Redis Streams 消费逻辑

### Task 20: Task Protocol Types

- Zod schemas matching Go `TaskPayload` JSON structure
- `TaskMessage`, `TaskCallback`, `SplitConfig`, `LLMModelConfig` 等
- 参考: Go `apps/keworker/internal/services/` + `pkgs/task/ragtask/payload.go`

### Task 21: Callback Publisher

- `publishCallback(nc, taskType, taskId, status, result)` → NATS publish to `core.task.callback.{taskType}`

### Task 22: Signal Handler + Worker Main

- `process.on('SIGTERM/SIGINT')` → stop signal → drain consumer → exit
- Worker entry: connect NATS → create consumers for analyser/chunker → start consume loops

---

## Phase 7: 集成 + 部署

### Task 23: ChunkWorker Pipeline

- 组装 NATS consumer + ChunkerService + S3/ES/LLM providers
- 消费 `core.task.ke.knowledge_task` → 处理 → callback

### Task 24: AnalyserWorker Pipeline

- 组装 NATS consumer + MinerU client + json_to_md + S3
- 消费 `core.task.ke.prase_pdf_task` → 处理 → callback

### Task 25: Dockerfile

- `apps/apps/worker/script/Dockerfile` — Node 22 build stage + runtime stage
- 参考: `frontend/corekg/Dockerfile.prod` 模式

### Task 26: Go 侧 NATS 桥接

- `pkgs/task/nats_bridge.go` — 在 `PushTaskQueue()` 中同时 `js.Publish()`
- Consumer 订阅 `core.task.callback.*` → 调用现有 `HandleTaskCallback()` 逻辑
- 最小改动：不改现有 Redis/HTTP 代码

### Task 27: 环境变量 + .env.example

- 完整的 `.env.example` 包含 NATS/ES/S3/LLM/Embedding/VLLM 全部变量

### Task 28: 双跑验证方案

- 文档说明：新旧 Worker 并行消费不同 NATS subject
- 对比脚本：相同文档 → 相同 split_config → 对比 chunk 输出一致性
- Go worker 的 `workerServerURL` 可切换指向 Node 服务

---

## 依赖安装汇总

```bash
# Task 1 基础
pnpm add pino zod && pnpm add -D typescript vitest tsx @types/node

# Task 4 S3
pnpm add @aws-sdk/client-s3 @aws-sdk/lib-storage mime-types && pnpm add -D @types/mime-types

# Task 5 ES
pnpm add @elastic/elasticsearch

# Task 6 AI
pnpm add ai @ai-sdk/openai-compatible

# Task 7 tiktoken
pnpm add js-tiktoken

# Task 9 markdown
pnpm add markdown-it && pnpm add -D @types/markdown-it

# Task 13 image
pnpm add sharp

# Task 18 NATS
pnpm add @nats-io/nats-core @nats-io/jetstream
```

---

详细任务 4-28 的完整 TDD 步骤见 [2026-07-21-kealgo-nats-worker-part2.md](./2026-07-21-kealgo-nats-worker-part2.md)。
