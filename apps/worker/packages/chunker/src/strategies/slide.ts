import type { ChunkStrategy, ChunkStrategyOptions, ChunkSplitResult } from "../strategy.js";

const YG_POS_PATTERN = /<!--yg_pos(\d+),(\d+),(\d+),(\d+),(\d+)yg_pos-->/g;

export const slideStrategy: ChunkStrategy = {
  name: "slide",

  split(content: string, _options: ChunkStrategyOptions): ChunkSplitResult {
    if (!content || !content.trim()) {
      return { chunks: [], metas: [] };
    }

    const matches = [...content.matchAll(YG_POS_PATTERN)];
    if (matches.length === 0) {
      return {
        chunks: [content.trim()],
        metas: [{ headers: {}, metadata: { page: 0 } }],
      };
    }

    const markers = matches.map((m) => ({
      pos: m.index!,
      pageIdx: parseInt(m[1], 10),
    }));

    const rawChunks: string[] = [];
    const pages: number[] = [];
    let currentPage = markers[0].pageIdx;
    let pageStart = 0;

    for (const marker of markers) {
      if (marker.pageIdx !== currentPage) {
        const chunkText = content.slice(pageStart, marker.pos).trim();
        if (chunkText) {
          rawChunks.push(chunkText);
          pages.push(currentPage);
        }
        pageStart = marker.pos;
        currentPage = marker.pageIdx;
      }
    }

    const lastChunk = content.slice(pageStart).trim();
    if (lastChunk) {
      rawChunks.push(lastChunk);
      pages.push(currentPage);
    }

    const cleaned = rawChunks.map((c) => c.replace(YG_POS_PATTERN, "").trim());

    const resultChunks: string[] = [];
    const resultPages: number[] = [];
    for (let i = 0; i < cleaned.length; i++) {
      if (cleaned[i]) {
        resultChunks.push(cleaned[i]);
        resultPages.push(pages[i]);
      }
    }

    return {
      chunks: resultChunks,
      metas: resultPages.map((p) => ({ headers: {}, metadata: { page: p } })),
    };
  },
};
