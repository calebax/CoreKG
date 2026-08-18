import fs from "node:fs/promises";
import { resolve, extname, basename } from "node:path";
import { Command } from "commander";
import { createLogger, withContext } from "@corekg/logger";
import { createLLMProvider, createEmbeddingProvider } from "@corekg/ai";
import { registerBuiltinStrategies, preprocessText, resolveStrategy, countTokens } from "@corekg/chunker";
import type { AsyncChunkStrategy } from "@corekg/chunker";
import { v4 as uuidv4 } from "uuid";
import { createHash } from "node:crypto";
import { loadLocalConfig } from "@corekg/config";
import type { ChunkDocument } from "@corekg/search";

const logger = createLogger("chunk");

export function createChunkCommand(): Command {
  const cmd = new Command("chunk")
    .description("Local CLI tool for text chunking")
    .requiredOption("--file <path>", "local file path to chunk")
    .option("--output <path>", "output file path (default: stdout)")
    .option("--strategy <name>", "chunking strategy name", "smart")
    .option("--chunk-size <n>", "chunk token size", "512")
    .option("--min-tokens <n>", "minimum chunk tokens", "10")
    .option("--split-level <n>", "heading split level", "2")
    .option("--overlap <n>", "overlap ratio", "0")
    .option("--regex <pattern>", "regex split pattern", "")
    .option("--no-embedding", "skip embedding generation")
    .option("--no-preprocess", "skip text preprocessing")
    .action(async (opts) => {
      try {
        await runChunk(opts);
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        logger.error({ err: msg }, "chunk command failed");
        process.exit(1);
      }
    });
  return cmd;
}

interface ChunkOptions {
  file: string;
  output?: string;
  strategy: string;
  chunkSize: string;
  minTokens: string;
  splitLevel: string;
  overlap: string;
  regex: string;
  embedding: boolean;
  preprocess: boolean;
}

async function runChunk(opts: ChunkOptions) {
  const filePath = resolve(opts.file);
  const stat = await fs.stat(filePath).catch(() => null);
  if (!stat || !stat.isFile()) {
    throw new Error(`file not found: ${filePath}`);
  }

  const log = withContext(logger, { file: filePath, strategy: opts.strategy });

  registerBuiltinStrategies();

  let content = await fs.readFile(filePath, "utf-8");

  if (opts.preprocess) {
    content = preprocessText(content, {
      removeEmail: true,
      removeUrl: true,
      removeEmptyLine: true,
    });
  }

  const config = loadLocalConfig();
  const llm = config.llm ? createLLMProvider(config.llm) : null;
  const embedding = opts.embedding && config.embedding ? createEmbeddingProvider(config.embedding) : null;

  const chunkTokenNum = parseInt(opts.chunkSize, 10);
  const minChunkTokens = parseInt(opts.minTokens, 10);
  const splitLevel = parseInt(opts.splitLevel, 10);
  const overlapRatio = parseFloat(opts.overlap);

  const strategy = resolveStrategy(opts.strategy);
  const splitOptions = {
    chunkTokenNum,
    minChunkTokens,
    splitLevel,
    overlapRatio,
    regexPattern: opts.regex,
    delimiter: "\n!?。；！？",
    enableHeadingInContent: false,
  };

  let result;
  if ("splitAsync" in strategy) {
    if (!llm) {
      throw new Error(`strategy "${opts.strategy}" requires LLM, set LLM_API_KEY and LLM_BASE_URL`);
    }
    result = await (strategy as AsyncChunkStrategy).splitAsync(content, splitOptions, { llm });
  } else {
    result = strategy.split(content, splitOptions);
  }

  const fileName = basename(filePath);
  const fileExt = extname(filePath);
  const docs: Record<string, ChunkDocument> = {};
  let sequence = 1;

  for (let i = 0; i < result.chunks.length; i++) {
    const chunkText = result.chunks[i];
    if (!chunkText || chunkText.trim().length === 0) continue;

    const uid = uuidv4();
    const meta = result.metas[i] || {};
    const headers = meta.headers || {};
    const sortedKeys = Object.keys(headers).sort((a, b) => a.localeCompare(b));
    const titleLevel = sortedKeys.length > 0
      ? [sortedKeys.map((k) => headers[k]).join(" -> ")]
      : null;

    let type: ChunkDocument["type"] = "chunk";
    if (fileExt === ".mp4") type = "video";
    else if (/<table>.*<\/table>/s.test(chunkText)) type = "table";
    else if (/!\[.*?\]\(.*?\)/.test(chunkText)) type = "image";

    docs[uid] = {
      forest_id: "0",
      company_id: "0",
      uin: "0",
      file_id: "0",
      version: "V2.0",
      file_name: fileName,
      type,
      tokens: countTokens(chunkText),
      chunk_size: chunkText.length,
      sequence: sequence++,
      location: null,
      yg_location: null,
      description: chunkText,
      description_hash: createHash("sha256").update(chunkText).digest("hex"),
      embedding: null,
      image_url: null,
      image_embedding: null,
      formula: null,
      table: type === "table" ? chunkText : null,
      title_level_ids: null,
      title_level: titleLevel,
      references: null,
      graph_info: null,
      graph_status: null,
    };
  }

  if (embedding) {
    const entries = Object.entries(docs);
    const texts = entries.map(([, doc]) => doc.description);
    if (texts.length > 0) {
      log.info({ count: texts.length }, "generating embeddings");
      const embeddings = await embedding.embedBatch(texts, { concurrency: 10 });
      for (let i = 0; i < entries.length; i++) {
        docs[entries[i][0]].embedding = embeddings[i] ? [...embeddings[i]!] : null;
      }
    }
  }

  const output = JSON.stringify(Object.values(docs), null, 2);

  if (opts.output) {
    const outPath = resolve(opts.output);
    await fs.writeFile(outPath, output, "utf-8");
    log.info({ path: outPath, chunks: Object.keys(docs).length }, "chunk result written");
  } else {
    process.stdout.write(output + "\n");
  }
}
