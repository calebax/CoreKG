import { smartStrategy } from "./smart.js";
import type { ChunkStrategy, ChunkStrategyOptions, ChunkSplitResult } from "../strategy.js";

export const strictRegexStrategy: ChunkStrategy = {
  name: "strict-regex",

  split(content: string, options: ChunkStrategyOptions): ChunkSplitResult {
    if (!content || !content.trim()) {
      return { chunks: [], metas: [] };
    }

    const regexPattern = options.regexPattern;
    if (!regexPattern || !regexPattern.trim()) {
      const fallback = smartStrategy.split(content, options);
      return fallback;
    }

    try {
      const precisePattern = new RegExp("^\\s*" + regexPattern);
      const lines = content.split("\n");
      const chunks: string[] = [];
      let currentChunk: string[] = [];

      for (const line of lines) {
        if (precisePattern.test(line) && currentChunk.length > 0) {
          const chunkContent = currentChunk.join("\n").trim();
          if (chunkContent) chunks.push(chunkContent);
          currentChunk = [line];
        } else {
          currentChunk.push(line);
        }
      }

      if (currentChunk.length > 0) {
        const chunkContent = currentChunk.join("\n").trim();
        if (chunkContent) chunks.push(chunkContent);
      }

      const filtered = chunks.filter((c) => c.trim());
      return {
        chunks: filtered,
        metas: filtered.map(() => ({ headers: {}, metadata: {} })),
      };
    } catch {
      return smartStrategy.split(content, options);
    }
  },
};
