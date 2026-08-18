import { AppConfigSchema, type AppConfig, LocalConfigSchema, type LocalConfig } from "./schema.js";

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
export type { AppConfig, AgentConfig, DaemonConfig, RPCConfig, LocalConfig } from "./schema.js";

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
    workerId: process.env.WORKER_ID,
    agent: {
      apiUrl: process.env.AGENT_API_URL,
      apiKey: process.env.AGENT_API_KEY,
      chunkSize: Number(process.env.AGENT_CHUNK_SIZE) || undefined,
      maxTokenSize: Number(process.env.AGENT_MAX_TOKEN_SIZE) || undefined,
      maxWorkers: Number(process.env.AGENT_MAX_WORKERS) || undefined,
      pool: process.env.AGENT_POOL ? JSON.parse(process.env.AGENT_POOL) : undefined,
    },
    daemon: {
      url: process.env.DAEMON_URL,
    },
    rpc: {
      queueGroup: process.env.RPC_QUEUE_GROUP,
      timeoutMs: Number(process.env.RPC_TIMEOUT_MS) || undefined,
    },
  });
}

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

  if (process.env.AGENT_API_KEY || process.env.AGENT_API_URL) {
    raw.agent = {
      apiUrl: process.env.AGENT_API_URL,
      apiKey: process.env.AGENT_API_KEY,
      chunkSize: Number(process.env.AGENT_CHUNK_SIZE) || undefined,
      maxTokenSize: Number(process.env.AGENT_MAX_TOKEN_SIZE) || undefined,
      maxWorkers: Number(process.env.AGENT_MAX_WORKERS) || undefined,
      pool: process.env.AGENT_POOL ? JSON.parse(process.env.AGENT_POOL) : undefined,
    };
  }

  return LocalConfigSchema.parse(raw);
}
