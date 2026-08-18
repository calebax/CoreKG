import { z } from "zod";

export const NATSConfigSchema = z.object({
  url: z.string().default("nats://localhost:4222"),
  stream: z.string().default("CORE_TASKS"),
  subjects: z.object({
    analyser: z.string().default("core.task.ke.prase_pdf_task"),
    chunker: z.string().default("core.task.ke.knowledge_task"),
    callback: z.string().default("core.task.callback.*"),
  }).optional(),
});

export const ESConfigSchema = z.object({
  host: z.string().url(),
  username: z.string(),
  password: z.string(),
  poolSize: z.number().int().positive().default(10),
  requestTimeoutMs: z.number().int().positive().default(30000),
});

export const S3ConfigSchema = z.object({
  endpointUrl: z.string().url(),
  accessKeyId: z.string(),
  secretAccessKey: z.string(),
  region: z.string().default("us-east-1"),
  defaultBucket: z.string(),
  publicEndpointUrl: z.string().url().optional(),
});

export const LLMConfigSchema = z.object({
  apiKey: z.string(),
  baseUrl: z.string().url(),
  model: z.string().default("deepseek-v3"),
  timeoutMs: z.number().int().positive().default(300000),
});

export const EmbeddingConfigSchema = z.object({
  apiKey: z.string(),
  baseUrl: z.string().url(),
  model: z.string().default("Qwen3-Embedding-0.6B"),
  timeoutMs: z.number().int().positive().default(30000),
});

export const VLLMConfigSchema = z.object({
  apiKey: z.string(),
  baseUrl: z.string().url(),
  model: z.string().default("qwen3-vl-plus"),
  timeoutMs: z.number().int().positive().default(300000),
});

export const ConcurrencyConfigSchema = z.object({
  embeddingWorkers: z.number().int().positive().default(30),
  llmWorkers: z.number().int().positive().default(30),
});

export const AgentConfigSchema = z.object({
  apiUrl: z.string().url(),
  apiKey: z.string(),
  chunkSize: z.number().int().positive().default(60000),
  maxTokenSize: z.number().int().positive().default(120000),
  maxWorkers: z.number().int().positive().default(50),
  pool: z.record(z.string(), z.string()).default({}),
});

export type AgentConfig = z.infer<typeof AgentConfigSchema>;

export const DaemonConfigSchema = z.object({
  url: z.string().url().default("http://localhost:5000/local.Run"),
});

export type DaemonConfig = z.infer<typeof DaemonConfigSchema>;

export const RPCConfigSchema = z.object({
  queueGroup: z.string().default("core-task-workers"),
  timeoutMs: z.number().int().positive().default(300000),
});

export type RPCConfig = z.infer<typeof RPCConfigSchema>;

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

export type AppConfig = z.infer<typeof AppConfigSchema>;

export const LocalConfigSchema = z.object({
  llm: LLMConfigSchema.optional(),
  embedding: EmbeddingConfigSchema.optional(),
  agent: AgentConfigSchema.optional(),
});

export type LocalConfig = z.infer<typeof LocalConfigSchema>;
