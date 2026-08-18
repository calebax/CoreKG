import { describe, it, expect } from "vitest";
import { contentListToMarkdown } from "./json-to-md.js";

describe("contentListToMarkdown", () => {
  it("should convert text items with headings", () => {
    const input = [
      { type: "text", text: "Title", text_level: 1, page_idx: 0, bbox: [0, 0, 100, 50] },
      { type: "text", text: "Some paragraph", page_idx: 0, bbox: [0, 50, 100, 100] },
    ];
    const result = contentListToMarkdown(input);
    expect(result).toContain("# Title");
    expect(result).toContain("Some paragraph");
    expect(result).toContain("<!--yg_pos1,0,0,100,50yg_pos-->");
  });

  it("should convert image items", () => {
    const input = [
      { type: "image", img_path: "img1.jpg", image_caption: ["Figure 1"], page_idx: 1, bbox: [10, 20, 200, 300] },
    ];
    const result = contentListToMarkdown(input);
    expect(result).toContain("![Figure 1](images/img1.jpg)");
    expect(result).toContain("Figure 1");
  });

  it("should convert table items with HTML body", () => {
    const input = [
      { type: "table", table_body: "<table><tr><td>A</td></tr></table>", table_caption: ["Table 1"], page_idx: 0, bbox: [0, 0, 100, 100] },
    ];
    const result = contentListToMarkdown(input);
    expect(result).toContain("**Table 1**");
    expect(result).toContain("<table>");
  });

  it("should convert equation items", () => {
    const input = [
      { type: "equation", text: "E = mc^2", page_idx: 0, bbox: [0, 0, 100, 50] },
    ];
    const result = contentListToMarkdown(input);
    expect(result).toContain("$$\nE = mc^2\n$$");
  });

  it("should convert code items", () => {
    const input = [
      { type: "code", code_body: "print('hello')", sub_type: "code", page_idx: 0, bbox: [0, 0, 100, 100] },
    ];
    const result = contentListToMarkdown(input);
    expect(result).toContain("```python\nprint('hello')\n```");
  });

  it("should convert list items", () => {
    const input = [
      { type: "list", list_items: ["item 1", "item 2"], page_idx: 0, bbox: [0, 0, 100, 100] },
    ];
    const result = contentListToMarkdown(input);
    expect(result).toContain("- item 1");
    expect(result).toContain("- item 2");
  });

  it("should skip null and empty items", () => {
    const input = [null as any, {} as any, { type: "unknown" } as any];
    const result = contentListToMarkdown(input);
    expect(result).toBe("");
  });
});
