import type { Logger } from "pino";
import type { StorageProvider } from "@corekg/storage";
import type { ESProvider } from "@corekg/search";
import type { LLMProvider, EmbeddingProvider } from "@corekg/ai";
import type { TaskPayload } from "@corekg/nats";

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
  algoUrl: string;
  logger: Logger;
}

export interface TaskHandlerResult {
  status: "success" | "fail";
  result?: unknown;
  error?: string;
}

export type TaskHandlerFn = (
  ctx: TaskContext,
  payload: TaskPayload,
) => Promise<TaskHandlerResult>;

export interface TaskResultMessage {
  task_id: number;
  worker_id: string;
  task_type: string;
  status: "success" | "fail";
  result?: string;
  error_message?: string;
}

export interface TaskHandlerDef {
  name: string;
  dispatchSubject: string;
  resultSubject: string;
  handler: TaskHandlerFn;
}
