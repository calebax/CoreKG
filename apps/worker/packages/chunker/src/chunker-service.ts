import { v4 as uuidv4 } from "uuid";
import { createHash } from "node:crypto";
import { readFile, rm } from "node:fs/promises";
import { resolveStrategy } from "./registry.js";
import { countTokens } from "./tokenizer.js";
import { preprocessText } from "./preprocessor.js";
import { registerBuiltinStrategies } from "./registry.js";
import type { AsyncChunkStrategy } from "./strategy.js";
import type { StorageProvider } from "@corekg/storage";
import type { LLMProvider, EmbeddingProvider } from "@corekg/ai";
import type { ChunkDocument } from "@corekg/search";

registerBuiltinStrategies();

export interface ChunkerOptions {
  url: string;
  forestId: string;
  companyId: string;
  uin: string;
  fileId: string;
  fileName: string | null;
  fileExt: string | null;
  indexName: string | null;
  removeEmail: boolean;
  removeUrl: boolean;
  removeEmptyLine: boolean;
  splitMode: string;
  chunkTokenNum: number;
  minChunkTokens: number;
  splitLevel: number;
  overlapRatio: number;
  regexPattern: string;
  enableHeadingInContent: boolean;
  llmConcurrency: number;
  embeddingConcurrency: number;
}

export interface ChunkerDeps {
  storage: StorageProvider;
  llm: LLMProvider;
  vllm?: LLMProvider;
  embedding: EmbeddingProvider;
}

export class ChunkerService {
  constructor(private deps: ChunkerDeps) {}

  async process(options: ChunkerOptions): Promise<Record<string, ChunkDocument>> {
    const tmpDir = `/tmp/kealgo-${Date.now()}`;
    const mdPath = await this.deps.storage.downloadFile(options.url, tmpDir);
    let content = await readFile(mdPath, "utf-8");

    content = preprocessText(content, {
      removeEmail: options.removeEmail,
      removeUrl: options.removeUrl,
      removeEmptyLine: options.removeEmptyLine,
    });

    const strategy = resolveStrategy(options.splitMode);
    const splitOptions = {
      chunkTokenNum: options.chunkTokenNum,
      minChunkTokens: options.minChunkTokens,
      splitLevel: options.splitLevel,
      overlapRatio: options.overlapRatio,
      regexPattern: options.regexPattern,
      delimiter: "\n!?。；！？",
      enableHeadingInContent: options.enableHeadingInContent,
    };

    let result;
    if ("splitAsync" in strategy) {
      result = await (strategy as AsyncChunkStrategy).splitAsync(content, splitOptions, { llm: this.deps.llm });
    } else {
      result = strategy.split(content, splitOptions);
    }

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
      if (options.fileExt === ".mp4") type = "video";
      else if (/<table>.*<\/table>/s.test(chunkText)) type = "table";
      else if (/!\[.*?\]\(.*?\)/.test(chunkText)) type = "image";

      docs[uid] = {
        forest_id: options.forestId,
        company_id: options.companyId,
        uin: options.uin,
        file_id: options.fileId,
        version: "V2.0",
        file_name: options.fileName,
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

    const chunkEntries = Object.entries(docs);
    const texts = chunkEntries.map(([, doc]) => doc.description);
    if (texts.length > 0) {
      const embeddings = await this.deps.embedding.embedBatch(texts, {
        concurrency: options.embeddingConcurrency,
      });
      for (let i = 0; i < chunkEntries.length; i++) {
        docs[chunkEntries[i][0]].embedding = embeddings[i] ? [...embeddings[i]!] : null;
      }
    }

    await rm(tmpDir, { recursive: true, force: true }).catch(() => {});

    return docs;
  }
}
