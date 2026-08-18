import { loadLocalConfig } from "@corekg/config";
import { createS3Provider, type StorageProvider } from "@corekg/storage";
import { createESProvider, type ESProvider } from "@corekg/search";
import { createLLMProvider, createEmbeddingProvider } from "@corekg/ai";
import type { LLMProvider, EmbeddingProvider } from "@corekg/ai";
import type { TaskContext } from "@corekg/workers";
import { createLogger } from "@corekg/logger";

export interface LocalContextOptions {
  esHost?: string;
  esUsername?: string;
  esPassword?: string;
  s3Endpoint?: string;
  s3AccessKey?: string;
  s3SecretKey?: string;
  s3Bucket?: string;
  daemonUrl?: string;
  algoUrl?: string;
}

export function buildLocalContext(opts: LocalContextOptions = {}): TaskContext {
  const cfg = loadLocalConfig();

  const esHost = opts.esHost || process.env.ES_HOST;
  const esUser = opts.esUsername || process.env.ES_USERNAME || "";
  const esPass = opts.esPassword || process.env.ES_PASSWORD || "";

  const s3Endpoint = opts.s3Endpoint || process.env.S3_ENDPOINT_URL || "";
  const s3AccessKey = opts.s3AccessKey || process.env.S3_ACCESS_KEY_ID || "";
  const s3SecretKey = opts.s3SecretKey || process.env.S3_SECRET_ACCESS_KEY || "";
  const s3Bucket = opts.s3Bucket || process.env.S3_DEFAULT_BUCKET || "";

  let es: ESProvider | null = null;
  if (esHost) {
    es = createESProvider({ host: esHost, username: esUser, password: esPass, poolSize: 10, requestTimeoutMs: 30000 });
  }

  let storage: StorageProvider | null = null;
  if (s3Endpoint && s3AccessKey) {
    storage = createS3Provider({
      endpointUrl: s3Endpoint,
      accessKeyId: s3AccessKey,
      secretAccessKey: s3SecretKey,
      region: process.env.S3_REGION || "us-east-1",
      defaultBucket: s3Bucket,
      publicEndpointUrl: process.env.S3_PUBLIC_ENDPOINT_URL,
    });
  }

  let llm: LLMProvider | null = null;
  if (cfg.llm) {
    llm = createLLMProvider(cfg.llm);
  }

  let embedding: EmbeddingProvider | null = null;
  if (cfg.embedding) {
    embedding = createEmbeddingProvider(cfg.embedding);
  }

  const agentCfg = cfg.agent || {
    apiUrl: process.env.AGENT_API_URL || "",
    apiKey: process.env.AGENT_API_KEY || "",
    chunkSize: Number(process.env.AGENT_CHUNK_SIZE) || 60000,
    maxTokenSize: Number(process.env.AGENT_MAX_TOKEN_SIZE) || 120000,
    maxWorkers: Number(process.env.AGENT_MAX_WORKERS) || 50,
    pool: process.env.AGENT_POOL ? JSON.parse(process.env.AGENT_POOL) : {},
  };

  return {
    storage: storage!,
    es: es!,
    llm: llm!,
    embedding: embedding!,
    agentConfig: agentCfg,
    daemonUrl: opts.daemonUrl || process.env.DAEMON_URL || "http://localhost:5000/local.Run",
    algoUrl: opts.algoUrl || process.env.ALGO_URL || "http://localhost:5000/local.Run",
    logger: createLogger("local-cli"),
  };
}
