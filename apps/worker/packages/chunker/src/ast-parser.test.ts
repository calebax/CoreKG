import { describe, it, expect } from "vitest";
import { parseMarkdownTokens } from "./ast-parser.js";

describe("parseMarkdownTokens", () => {
  it("parses heading", () => {
    const tokens = parseMarkdownTokens("# Hello");
    expect(tokens.length).toBeGreaterThan(0);
    const heading = tokens.find((t: any) => t.type === "heading_open");
    expect(heading).toBeDefined();
    expect(heading!.tag).toBe("h1");
  });

  it("parses paragraph", () => {
    const tokens = parseMarkdownTokens("Some text here.");
    const paragraph = tokens.find((t: any) => t.type === "paragraph_open");
    expect(paragraph).toBeDefined();
  });

  it("returns empty array for empty input", () => {
    const tokens = parseMarkdownTokens("");
    expect(tokens).toEqual([]);
  });

  it("parses code block", () => {
    const tokens = parseMarkdownTokens("```python\nprint('hi')\n```");
    const fence = tokens.find((t: any) => t.type === "fence");
    expect(fence).toBeDefined();
  });
});
