import os from "node:os";
import { createLogger } from "@corekg/logger";
import { loadConfig } from "@corekg/config";
import { createNATSClient, TaskConsumer, TaskPayloadSchema, DispatchConsumer, ResultPublisher } from "@corekg/nats";
import { createS3Provider } from "@corekg/storage";
import { createESProvider } from "@corekg/search";
import { createLLMProvider, createEmbeddingProvider } from "@corekg/ai";
import { ChunkerService, registerBuiltinStrategies } from "@corekg/chunker";
import { handlerRegistry } from "@corekg/workers";
import type { TaskContext } from "@corekg/workers";

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

  const ctx: TaskContext = {
    storage,
    es,
    llm,
    embedding,
    agentConfig: config.agent,
    daemonUrl: config.daemon.url,
    algoUrl: process.env.ALGO_URL || config.daemon.url,
    logger: createLogger("task-handler"),
  };

  const { nc, js, jsm } = await createNATSClient(config);

  const resultPub = new ResultPublisher(js, jsm, workerId);
  await resultPub.ensureStream();

  const dispatchConsumers: DispatchConsumer<TaskContext>[] = [];
  for (const def of handlerRegistry) {
    const dc = new DispatchConsumer(
      js,
      jsm,
      {
        handlerName: def.name,
        dispatchSubject: def.dispatchSubject,
        resultSubject: def.resultSubject,
      },
      def.handler,
      ctx,
      resultPub,
      workerId,
    );
    dispatchConsumers.push(dc);
  }

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

      return { status: "success" as const, result: { chunk_count: Object.keys(docs).length } };
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      logger.error({ taskId, error: msg }, "chunk task failed");

      return { status: "fail" as const, error: msg };
    }
  });

  let stopping = false;
  const shutdown = async () => {
    if (stopping) {
      logger.warn("forced exit");
      process.exit(1);
    }
    stopping = true;
    logger.info("shutting down...");
    for (const dc of dispatchConsumers) {
      await dc.stop();
    }
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

  for (const def of handlerRegistry) {
    logger.info({ handler: def.name, subject: def.dispatchSubject }, "starting dispatch consumer");
  }
  await Promise.all(dispatchConsumers.map((dc) => dc.start()));
}
