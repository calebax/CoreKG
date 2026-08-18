import { describe, it, expect } from "vitest";
import { autoStrategy } from "./auto.js";
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

describe("autoStrategy", () => {
  it("returns empty for empty input", async () => {
    const llm = createMockLLM("basic");
    const result = await autoStrategy.splitAsync("", defaultOptions, { llm });
    expect(result.chunks).toHaveLength(0);
  });

  it("delegates to smart when LLM says smart", async () => {
    const llm = createMockLLM("smart");
    const content = `# Heading\n\nSome paragraph text that has enough content for testing purposes here.`;
    const result = await autoStrategy.splitAsync(content, defaultOptions, { llm });
    expect(result.chunks.length).toBeGreaterThanOrEqual(1);
  });

  it("delegates to basic when LLM says basic", async () => {
    const llm = createMockLLM("basic");
    const content = "Some text.\n\nMore text.";
    const result = await autoStrategy.splitAsync(content, defaultOptions, { llm });
    expect(result.chunks.length).toBeGreaterThanOrEqual(1);
  });

  it("falls back to smart on unknown strategy name", async () => {
    const llm = createMockLLM("unknown_strategy_xyz");
    const content = "# Heading\n\nSome text.";
    const result = await autoStrategy.splitAsync(content, defaultOptions, { llm });
    expect(result.chunks.length).toBeGreaterThanOrEqual(1);
  });

  it("falls back to smart when LLM call fails", async () => {
    const llm: LLMProvider = { chat: async () => { throw new Error("LLM down"); } };
    const content = "# Heading\n\nSome text.";
    const result = await autoStrategy.splitAsync(content, defaultOptions, { llm });
    expect(result.chunks.length).toBeGreaterThanOrEqual(1);
  });

  it("has name 'auto'", () => {
    expect(autoStrategy.name).toBe("auto");
  });

  it("synchronous split falls back to smart", () => {
    const result = autoStrategy.split("# Heading\n\nSome text.", defaultOptions);
    expect(result.chunks.length).toBeGreaterThanOrEqual(1);
  });
});
