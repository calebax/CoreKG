import { describe, it, expect } from "vitest";
import { TableEnhancer } from "./table-enhancer.js";
import type { LLMProvider } from "@corekg/ai";
import type { ChunkDocument } from "@corekg/search";

function makeDoc(desc: string, type: ChunkDocument["type"]): ChunkDocument {
  return {
    forest_id: "f",
    company_id: "c",
    uin: "u",
    file_id: "fi",
    version: "V2.0",
    file_name: null,
    type,
    tokens: 0,
    chunk_size: desc.length,
    sequence: 1,
    location: null,
    yg_location: null,
    description: desc,
    description_hash: "hash",
    embedding: null,
    image_url: null,
    image_embedding: null,
    formula: null,
    table: type === "table" ? desc : null,
    title_level_ids: null,
    title_level: null,
    references: null,
    graph_info: null,
    graph_status: null,
  };
}

describe("TableEnhancer", () => {
  it("inserts summary before <table> tag", async () => {
    const llm: LLMProvider = {
      async chat() {
        return "Sales data for Q1 2024";
      },
    };
    const enhancer = new TableEnhancer(llm);
    const tableContent = '<table><tr><td>A</td></tr></table>';
    const docs: Record<string, ChunkDocument> = {
      "id-1": makeDoc(tableContent, "table"),
    };
    const result = await enhancer.enhance(docs);
    expect(result["id-1"].description).toContain("[表格摘要: Sales data for Q1 2024]");
    expect(result["id-1"].description.indexOf("[表格摘要")).toBeLessThan(
      result["id-1"].description.indexOf("<table"),
    );
    expect(result["id-1"].table).toBe(tableContent);
  });

  it("skips non-table chunks", async () => {
    const llm: LLMProvider = {
      async chat() {
        throw new Error("should not be called");
      },
    };
    const enhancer = new TableEnhancer(llm);
    const docs: Record<string, ChunkDocument> = {
      "id-1": makeDoc("Just regular text", "chunk"),
    };
    const result = await enhancer.enhance(docs);
    expect(result["id-1"].description).toBe("Just regular text");
  });

  it("handles LLM failures gracefully", async () => {
    const llm: LLMProvider = {
      async chat() {
        throw new Error("LLM error");
      },
    };
    const enhancer = new TableEnhancer(llm);
    const docs: Record<string, ChunkDocument> = {
      "id-1": makeDoc('<table><tr><td>X</td></tr></table>', "table"),
    };
    const result = await enhancer.enhance(docs);
    expect(result["id-1"].description).toBe('<table><tr><td>X</td></tr></table>');
  });

  it("prepends summary when no <table> tag found", async () => {
    const llm: LLMProvider = {
      async chat() {
        return "Summary text";
      },
    };
    const enhancer = new TableEnhancer(llm);
    const docs: Record<string, ChunkDocument> = {
      "id-1": makeDoc("| A | B |\n|---|---|\n| 1 | 2 |", "table"),
    };
    const result = await enhancer.enhance(docs);
    expect(result["id-1"].description).toMatch(/^\[表格摘要: Summary text\]\n/);
  });
});
