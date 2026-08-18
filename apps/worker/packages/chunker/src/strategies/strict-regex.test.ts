import { describe, it, expect } from "vitest";
import { strictRegexStrategy } from "./strict-regex.js";
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

describe("strictRegexStrategy", () => {
  it("returns empty for empty input", () => {
    const result = strictRegexStrategy.split("", defaultOptions);
    expect(result.chunks).toHaveLength(0);
  });

  it("falls back to smart when regex is empty", () => {
    const content = "# Heading\n\nSome text.";
    const result = strictRegexStrategy.split(content, defaultOptions);
    expect(result.chunks.length).toBeGreaterThanOrEqual(1);
  });

  it("splits by regex pattern at line start", () => {
    const content = `第一条 这是第一条的内容。
一些额外的文字说明。

第二条 这是第二条的内容。
更多文字。

第三条 这是第三条的内容。`;
    const result = strictRegexStrategy.split(content, {
      ...defaultOptions,
      regexPattern: "第[一二三四五六七八九十]+条",
    });
    expect(result.chunks.length).toBe(3);
    expect(result.chunks[0]).toContain("第一条");
    expect(result.chunks[1]).toContain("第二条");
    expect(result.chunks[2]).toContain("第三条");
  });

  it("does not split when regex never matches", () => {
    const content = "No matching lines here\nJust regular text\nMore text";
    const result = strictRegexStrategy.split(content, {
      ...defaultOptions,
      regexPattern: "XXXXX",
    });
    expect(result.chunks).toHaveLength(1);
  });

  it("has name 'strict-regex'", () => {
    expect(strictRegexStrategy.name).toBe("strict-regex");
  });

  it("falls back to smart on invalid regex", () => {
    const content = "# Some heading\n\nText content here.";
    const result = strictRegexStrategy.split(content, {
      ...defaultOptions,
      regexPattern: "[invalid",
    });
    expect(result.chunks.length).toBeGreaterThanOrEqual(1);
  });
});
