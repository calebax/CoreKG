import { describe, it, expect } from "vitest";
import { basicStrategy } from "./basic.js";
import type { ChunkStrategyOptions } from "../strategy.js";

const defaultOptions: ChunkStrategyOptions = {
  chunkTokenNum: 50,
  minChunkTokens: 10,
  splitLevel: 2,
  overlapRatio: 0,
  regexPattern: "",
  delimiter: "\n!?。；！？",
  enableHeadingInContent: false,
};

describe("basicStrategy", () => {
  it("returns empty for empty input", () => {
    const result = basicStrategy.split("", defaultOptions);
    expect(result.chunks).toHaveLength(0);
  });

  it("returns empty for whitespace-only input", () => {
    const result = basicStrategy.split("  \n  \n  ", defaultOptions);
    expect(result.chunks).toHaveLength(0);
  });

  it("short text stays as one chunk", () => {
    const result = basicStrategy.split("Hello world", defaultOptions);
    expect(result.chunks).toHaveLength(1);
    expect(result.chunks[0]).toBe("Hello world");
  });

  it("long text gets split into chunks under chunkTokenNum", () => {
    const lines = Array.from({ length: 20 }, (_, i) => `Line number ${i + 1} with some extra words to pad it out.`);
    const content = lines.join("\n");
    const result = basicStrategy.split(content, { ...defaultOptions, chunkTokenNum: 30 });
    expect(result.chunks.length).toBeGreaterThan(1);
  });

  it("extracts tables and handles them separately", () => {
    const content = `Some text before

| Name | Age |
| --- | --- |
| Alice | 30 |

Some text after`;
    const result = basicStrategy.split(content, defaultOptions);
    expect(result.chunks.length).toBeGreaterThanOrEqual(2);
    const hasHtml = result.chunks.some((c) => c.includes("<table>") || c.includes("<tr>"));
    expect(hasHtml).toBe(true);
  });

  it("has name 'basic'", () => {
    expect(basicStrategy.name).toBe("basic");
  });

  it("handles very long single line by splitting in half", () => {
    const longLine = "word ".repeat(200).trim();
    const result = basicStrategy.split(longLine, { ...defaultOptions, chunkTokenNum: 50 });
    expect(result.chunks.length).toBeGreaterThanOrEqual(2);
  });
});
