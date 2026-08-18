import { describe, it, expect } from "vitest";
import { slideStrategy } from "./slide.js";
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

describe("slideStrategy", () => {
  it("returns empty for empty input", () => {
    const result = slideStrategy.split("", defaultOptions);
    expect(result.chunks).toHaveLength(0);
  });

  it("returns single chunk when no yg_pos markers", () => {
    const content = "Just regular text without any markers.";
    const result = slideStrategy.split(content, defaultOptions);
    expect(result.chunks).toHaveLength(1);
    expect(result.chunks[0]).toBe(content);
  });

  it("splits by yg_pos page index", () => {
    const content = `<!--yg_pos1,0,0,100,200yg_pos-->Page one content here.
<!--yg_pos2,0,0,100,200yg_pos-->Page two content here.
<!--yg_pos3,0,0,100,200yg_pos-->Page three content here.`;
    const result = slideStrategy.split(content, defaultOptions);
    expect(result.chunks).toHaveLength(3);
    expect(result.chunks[0]).toContain("Page one");
    expect(result.chunks[1]).toContain("Page two");
    expect(result.chunks[2]).toContain("Page three");
  });

  it("strips yg_pos markers from output", () => {
    const content = `<!--yg_pos1,0,0,100,200yg_pos-->Hello world`;
    const result = slideStrategy.split(content, defaultOptions);
    expect(result.chunks[0]).not.toContain("yg_pos");
    expect(result.chunks[0]).toBe("Hello world");
  });

  it("groups multiple markers on same page", () => {
    const content = `<!--yg_pos1,0,0,50,100yg_pos-->First line.
<!--yg_pos1,0,50,100,200yg_pos-->Second line.
<!--yg_pos2,0,0,100,200yg_pos-->Page two.`;
    const result = slideStrategy.split(content, defaultOptions);
    expect(result.chunks).toHaveLength(2);
    expect(result.chunks[0]).toContain("First line");
    expect(result.chunks[0]).toContain("Second line");
  });

  it("has name 'slide'", () => {
    expect(slideStrategy.name).toBe("slide");
  });
});
