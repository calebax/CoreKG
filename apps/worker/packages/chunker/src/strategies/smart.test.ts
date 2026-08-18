import { describe, it, expect } from "vitest";
import { smartStrategy } from "./smart.js";
import type { ChunkStrategyOptions } from "../strategy.js";

const defaultOptions: ChunkStrategyOptions = {
  chunkTokenNum: 100,
  minChunkTokens: 5,
  splitLevel: 2,
  overlapRatio: 0,
  regexPattern: "",
  delimiter: "\n!?。；！？",
  enableHeadingInContent: false,
};

describe("smartStrategy", () => {
  it("returns empty for empty input", () => {
    const result = smartStrategy.split("", defaultOptions);
    expect(result.chunks).toHaveLength(0);
  });

  it("splits at heading boundaries", () => {
    const content = `# Chapter 1

Some paragraph text here that is substantial enough to meet the minimum token threshold for chunking.

## Section 1.1

Another paragraph of text that should appear in a separate chunk from the previous section above.

## Section 1.2

Yet another paragraph of meaningful content that should form its own distinct chunk in output.`;
    const result = smartStrategy.split(content, defaultOptions);
    expect(result.chunks.length).toBeGreaterThanOrEqual(2);
  });

  it("preserves heading hierarchy in chunk metas", () => {
    const content = `# Chapter 1

Some intro text with enough words to meet minimum token threshold for chunking purposes.

## Section 1.1

Section content here that has enough words to form a valid chunk for testing purposes.`;
    const result = smartStrategy.split(content, defaultOptions);
    expect(result.chunks.length).toBeGreaterThanOrEqual(2);
    const hasHeaders = result.metas.some((m) => m.headers && Object.keys(m.headers).length > 0);
    expect(hasHeaders).toBe(true);
  });

  it("enableHeadingInContent adds parent headings to chunk text", () => {
    const content = `# Top Level

Intro text with enough words to exceed minimum token threshold for chunking here.

## Sub Section

Content here that also has enough words to form a proper chunk for our test.`;
    const result = smartStrategy.split(content, {
      ...defaultOptions,
      enableHeadingInContent: true,
    });
    expect(result.chunks.length).toBeGreaterThanOrEqual(1);
    const lastChunk = result.chunks[result.chunks.length - 1];
    expect(lastChunk).toContain("# Top Level");
  });

  it("keeps tables and code blocks intact", () => {
    const content = `# Heading

Some text before the table with enough words to count.

| A | B |
| --- | --- |
| 1 | 2 |

Some text after the table with enough words to count here.`;
    const result = smartStrategy.split(content, defaultOptions);
    const tableChunk = result.chunks.find((c) => c.includes("| A | B |") || c.includes("<table>") || c.includes("<th>"));
    expect(tableChunk).toBeDefined();
  });

  it("has name 'smart'", () => {
    expect(smartStrategy.name).toBe("smart");
  });
});
