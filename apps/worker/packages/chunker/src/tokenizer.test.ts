import { describe, it, expect } from "vitest";
import { countTokens } from "./tokenizer.js";

describe("countTokens", () => {
  it("counts English", () => {
    expect(countTokens("Hello world")).toBeGreaterThan(0);
  });

  it("counts Chinese", () => {
    expect(countTokens("你好世界")).toBeGreaterThan(0);
  });

  it("empty = 0", () => {
    expect(countTokens("")).toBe(0);
  });

  it("matches Python tiktoken cl100k_base", () => {
    expect(countTokens("Hello world")).toBe(2);
  });
});
