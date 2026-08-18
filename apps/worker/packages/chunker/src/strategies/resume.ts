import type { ChunkStrategy, ChunkStrategyOptions, ChunkSplitResult } from "../strategy.js";

const YG_POS_PATTERN = /<!--yg_pos\d+,\d+,\d+,\d+,\d+yg_pos-->/g;

export const resumeStrategy: ChunkStrategy = {
  name: "resume",

  split(content: string, _options: ChunkStrategyOptions): ChunkSplitResult {
    if (!content || !content.trim()) {
      return { chunks: [], metas: [] };
    }
    const cleaned = content.replace(YG_POS_PATTERN, "").trim();
    if (!cleaned) {
      return { chunks: [], metas: [] };
    }
    return {
      chunks: [cleaned],
      metas: [{ headers: {}, metadata: {} }],
    };
  },
};
