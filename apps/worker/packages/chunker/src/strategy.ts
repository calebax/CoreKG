import type { LLMProvider } from "@corekg/ai";

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
  splitAsync(
    content: string,
    options: ChunkStrategyOptions,
    deps: { llm: LLMProvider },
  ): Promise<ChunkSplitResult>;
}
