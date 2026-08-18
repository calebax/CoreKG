import { describe, it, expect } from "vitest";
import { titleStrategy } from "./title.js";
import type { ChunkStrategyOptions } from "../strategy.js";

const defaultOptions: ChunkStrategyOptions = {
  chunkTokenNum: 256,
  minChunkTokens: 10,
  splitLevel: 2,
  overlapRatio: 0,
  regexPattern: "",
  delimiter: "\n!?。；！？",
  enableHeadingInContent: false,
};

describe("titleStrategy", () => {
  it("returns empty for empty input", () => {
    const result = titleStrategy.split("", defaultOptions);
    expect(result.chunks).toHaveLength(0);
  });

  it("splits at H2 boundaries by default", () => {
    const content = `# Main Title

Introduction paragraph text.

## Section One

Content for section one with enough detail to be meaningful.

## Section Two

Content for section two with enough detail to be meaningful.

## Section Three

Content for section three with enough detail to be meaningful.`;
    const result = titleStrategy.split(content, defaultOptions);
    expect(result.chunks.length).toBeGreaterThanOrEqual(3);
  });

  it("falls back to H1 when H2 produces only 1 chunk", () => {
    const content = `# Title A

Some content under title A with enough words to be meaningful for testing.

# Title B

Some content under title B with enough words to be meaningful for testing.`;
    const result = titleStrategy.split(content, { ...defaultOptions, splitLevel: 2 });
    expect(result.chunks.length).toBeGreaterThanOrEqual(2);
  });

  it("enableHeadingInContent adds parent headings", () => {
    const content = `# Main

Intro text here.

## Sub

Sub content here with enough words for a valid chunk in testing.`;
    const result = titleStrategy.split(content, {
      ...defaultOptions,
      enableHeadingInContent: true,
    });
    expect(result.metas.length).toBeGreaterThan(0);
  });

  it("has name 'title'", () => {
    expect(titleStrategy.name).toBe("title");
  });

  it("preserves heading metadata in metas", () => {
    const content = `# Chapter

Intro text.

## Part A

Part A content with enough words to form a proper test chunk.

## Part B

Part B content with enough words to form a proper test chunk.`;
    const result = titleStrategy.split(content, defaultOptions);
    const hasHeaders = result.metas.some((m) => m.headers && Object.keys(m.headers).length > 0);
    expect(hasHeaders).toBe(true);
  });
});
