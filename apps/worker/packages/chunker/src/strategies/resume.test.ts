import { describe, it, expect } from "vitest";
import { resumeStrategy } from "./resume.js";
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

describe("resumeStrategy", () => {
  it("returns entire document as one chunk", () => {
    const content = "# Resume\n\nJohn Doe\n\nExperience: 5 years";
    const result = resumeStrategy.split(content, defaultOptions);
    expect(result.chunks).toHaveLength(1);
    expect(result.chunks[0]).toBe(content);
    expect(result.metas).toHaveLength(1);
  });

  it("strips yg_pos markers", () => {
    const content = "Hello <!--yg_pos1,0,0,100,200yg_pos--> World";
    const result = resumeStrategy.split(content, defaultOptions);
    expect(result.chunks).toHaveLength(1);
    expect(result.chunks[0]).toBe("Hello  World");
  });

  it("returns empty for empty input", () => {
    const result = resumeStrategy.split("", defaultOptions);
    expect(result.chunks).toHaveLength(0);
    expect(result.metas).toHaveLength(0);
  });

  it("returns empty for whitespace-only input", () => {
    const result = resumeStrategy.split("   \n  \n  ", defaultOptions);
    expect(result.chunks).toHaveLength(0);
  });

  it("has name 'resume'", () => {
    expect(resumeStrategy.name).toBe("resume");
  });
});
