import { countTokens } from "../tokenizer.js";
import { basicStrategy } from "./basic.js";
import type { AsyncChunkStrategy, ChunkStrategyOptions, ChunkSplitResult } from "../strategy.js";
import type { LLMProvider } from "@corekg/ai";

const SENTENCE_SPLIT_PATTERN = /(?<=[。！？!?\n])(?=[^\s])/;

function splitIntoSentences(text: string): string[] {
  return text.split(SENTENCE_SPLIT_PATTERN).map((s) => s.trim()).filter(Boolean);
}

function buildSegmentList(sentences: string[], startIdx: number, maxTokens: number): { text: string; end: number } {
  const lines: string[] = [];
  let currentTokens = 0;
  let end = startIdx;

  for (let i = startIdx; i < sentences.length; i++) {
    const line = `[${i + 1}] ${sentences[i]}`;
    const lineTokens = countTokens(line);
    if (currentTokens + lineTokens > maxTokens && lines.length > 0) break;
    lines.push(line);
    currentTokens += lineTokens;
    end = i + 1;
  }

  return { text: lines.join("\n"), end };
}

function parseMergeResponse(response: string, startIdx: number, sentenceCount: number): Array<[number, number]> {
  const chunks: Array<[number, number]> = [];
  for (const rawLine of response.trim().split("\n")) {
    const line = rawLine.trim();
    if (!line || line.startsWith("#")) continue;

    let a: number, b: number;
    const rangeMatch = line.match(/^(\d+)\s*[-–]\s*(\d+)/);
    if (rangeMatch) {
      a = parseInt(rangeMatch[1], 10);
      b = parseInt(rangeMatch[2], 10);
    } else {
      const singleMatch = line.match(/^(\d+)/);
      if (!singleMatch) continue;
      a = b = parseInt(singleMatch[1], 10);
    }

    const absA = Math.max(0, Math.min(startIdx + a - 1, sentenceCount - 1));
    const absB = Math.max(absA, Math.min(startIdx + b - 1, sentenceCount - 1));
    chunks.push([absA, absB]);
  }
  return chunks;
}

export const llmStrategy: AsyncChunkStrategy = {
  name: "llm",

  split(content: string, options: ChunkStrategyOptions): ChunkSplitResult {
    return basicStrategy.split(content, options);
  },

  async splitAsync(
    content: string,
    options: ChunkStrategyOptions,
    deps: { llm: LLMProvider },
  ): Promise<ChunkSplitResult> {
    if (!content || !content.trim()) {
      return { chunks: [], metas: [] };
    }

    const sentences = splitIntoSentences(content);
    if (sentences.length <= 1) {
      return { chunks: [content.trim()], metas: [{ headers: {}, metadata: {} }] };
    }

    const batchMaxTokens = Math.floor(options.chunkTokenNum * 8);
    const rawChunks: string[] = [];
    let i = 0;
    let lastContext = "";

    while (i < sentences.length) {
      const { text: segmentsText, end } = buildSegmentList(sentences, i, batchMaxTokens);

      let prompt: string;
      if (lastContext) {
        prompt = `Context from previous batch: ${lastContext}\n\nPlease merge the following numbered sentences into semantic chunks. Output one chunk per line using format like "1-3" or "4":\n\n${segmentsText}`;
      } else {
        prompt = `Please merge the following numbered sentences into semantic chunks. Output one chunk per line using format like "1-3" or "4":\n\n${segmentsText}`;
      }

      try {
        const response = await deps.llm.chat(prompt);
        const chunkRanges = parseMergeResponse(response, i, sentences.length);

        for (const [absA, absB] of chunkRanges) {
          rawChunks.push(sentences.slice(absA, absB + 1).join("\n"));
        }

        if (chunkRanges.length > 0) {
          const [lastA] = chunkRanges[chunkRanges.length - 1];
          lastContext = sentences[lastA].slice(0, 200);
        } else {
          lastContext = "";
        }
      } catch {
        for (let j = i; j < end; j++) {
          rawChunks.push(sentences[j]);
        }
      }

      i = end;
    }

    const mergedChunks: string[] = [];
    let j = 0;
    while (j < rawChunks.length) {
      const cur = rawChunks[j];
      const curTokens = countTokens(cur);
      if (curTokens < options.minChunkTokens && j + 1 < rawChunks.length) {
        mergedChunks.push(cur + "\n" + rawChunks[j + 1]);
        j += 2;
      } else {
        mergedChunks.push(cur);
        j++;
      }
    }

    const filtered = mergedChunks.filter((c) => c.trim());
    return {
      chunks: filtered,
      metas: filtered.map(() => ({ headers: {}, metadata: {} })),
    };
  },
};
