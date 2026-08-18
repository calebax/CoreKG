import { describe, it, expect } from "vitest";
import { preprocessText } from "./preprocessor.js";

describe("preprocessText", () => {
  it("removes emails when enabled", () => {
    const result = preprocessText("Contact me at test@example.com for info", {
      removeEmail: true,
      removeUrl: false,
      removeEmptyLine: false,
    });
    expect(result).not.toContain("test@example.com");
  });

  it("removes bare URLs when enabled", () => {
    const result = preprocessText("Visit https://example.com today", {
      removeEmail: false,
      removeUrl: true,
      removeEmptyLine: false,
    });
    expect(result).not.toContain("https://example.com");
  });

  it("preserves markdown image URLs", () => {
    const result = preprocessText("![alt](https://img.example.com/pic.png)", {
      removeEmail: false,
      removeUrl: true,
      removeEmptyLine: false,
    });
    expect(result).toContain("![alt](https://img.example.com/pic.png)");
  });

  it("collapses multiple blank lines", () => {
    const result = preprocessText("line1\n\n\n\n\nline2", {
      removeEmail: false,
      removeUrl: false,
      removeEmptyLine: true,
    });
    expect(result).not.toMatch(/\n{3,}/);
  });

  it("handles all rules disabled", () => {
    const input = "test@example.com https://x.com\n\n\n\nend";
    const result = preprocessText(input, {
      removeEmail: false,
      removeUrl: false,
      removeEmptyLine: false,
    });
    expect(result).toBe(input);
  });
});
