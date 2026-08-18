import { describe, it, expect } from "vitest";
import { llmStrategy } from "./llm.js";
import type { ChunkStrategyOptions } from "../strategy.js";
import type { LLMProvider } from "@corekg/ai";

const defaultOptions: ChunkStrategyOptions = {
  chunkTokenNum: 256,
  minChunkTokens: 10,
  splitLevel: 2,
  overlapRatio: 0,
  regexPattern: "",
  delimiter: "\n!?。；！？",
  enableHeadingInContent: false,
};

function createMockLLM(response: string): LLMProvider {
  return { chat: async () => response };
}

describe("llmStrategy", () => {
  it("returns empty for empty input", async () => {
    const llm = createMockLLM("1");
    const result = await llmStrategy.splitAsync("", defaultOptions, { llm });
    expect(result.chunks).toHaveLength(0);
  });

  it("returns single chunk for single sentence", async () => {
    const llm = createMockLLM("1");
    const result = await llmStrategy.splitAsync("Just one sentence.", defaultOptions, { llm });
    expect(result.chunks).toHaveLength(1);
  });

  it("merges sentences based on LLM response", async () => {
    const content = "First sentence here with more words.\nSecond sentence here with more words.\nThird sentence here with more words.\nFourth sentence here with more words.\nFifth sentence here with more words.";
    const llm = createMockLLM("1-2\n3-5");
    const result = await llmStrategy.splitAsync(content, { ...defaultOptions, minChunkTokens: 2 }, { llm });
    expect(result.chunks.length).toBe(2);
    expect(result.chunks[0]).toContain("First sentence");
    expect(result.chunks[0]).toContain("Second sentence");
    expect(result.chunks[1]).toContain("Third sentence");
  });

  it("has name 'llm'", () => {
    expect(llmStrategy.name).toBe("llm");
  });

  it("synchronous split falls back to basic", () => {
    const result = llmStrategy.split("Some text here.\nMore text.\nEven more.", defaultOptions);
    expect(result.chunks.length).toBeGreaterThanOrEqual(1);
  });
});
