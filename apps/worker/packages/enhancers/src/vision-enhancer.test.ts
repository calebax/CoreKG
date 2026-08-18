import { describe, it, expect } from "vitest";
import { VisionEnhancer, extractImageReferences } from "./vision-enhancer.js";
import type { LLMProvider } from "@corekg/ai";
import type { ChunkDocument } from "@corekg/search";

function makeDoc(desc: string): ChunkDocument {
  return {
    forest_id: "f",
    company_id: "c",
    uin: "u",
    file_id: "fi",
    version: "V2.0",
    file_name: null,
    type: "image",
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
    table: null,
    title_level_ids: null,
    title_level: null,
    references: null,
    graph_info: null,
    graph_status: null,
  };
}

describe("extractImageReferences", () => {
  it("extracts HTML img tags", () => {
    const refs = extractImageReferences('<img src="/img/photo.jpg" alt="test">');
    expect(refs).toHaveLength(1);
    expect(refs[0].url).toBe("/img/photo.jpg");
  });

  it("extracts markdown image syntax", () => {
    const refs = extractImageReferences("![alt text](https://example.com/img.png)");
    expect(refs).toHaveLength(1);
    expect(refs[0].url).toBe("https://example.com/img.png");
  });

  it("extracts multiple images", () => {
    const text = '<img src="a.jpg"> text ![b](b.png) <img src="c.gif">';
    const refs = extractImageReferences(text);
    expect(refs).toHaveLength(3);
  });

  it("returns empty for no images", () => {
    expect(extractImageReferences("no images here")).toHaveLength(0);
  });
});

describe("VisionEnhancer", () => {
  it("inserts descriptions for image chunks", async () => {
    const llm: LLMProvider = {
      async chat() {
        return "A beautiful sunset over the ocean";
      },
    };
    const enhancer = new VisionEnhancer(llm);
    const docs: Record<string, ChunkDocument> = {
      "id-1": makeDoc('Some text <img src="http://img.com/sunset.jpg"> more text'),
    };
    const result = await enhancer.enhance(docs);
    expect(result["id-1"].description).toContain("[图片描述: A beautiful sunset over the ocean]");
    expect(result["id-1"].image_url).toBe("http://img.com/sunset.jpg");
  });

  it("skips chunks without images", async () => {
    const llm: LLMProvider = {
      async chat() {
        return "should not be called";
      },
    };
    const enhancer = new VisionEnhancer(llm);
    const docs: Record<string, ChunkDocument> = {
      "id-1": makeDoc("No images in this chunk"),
    };
    const result = await enhancer.enhance(docs);
    expect(result["id-1"].description).toBe("No images in this chunk");
    expect(result["id-1"].image_url).toBeNull();
  });

  it("handles LLM failures gracefully", async () => {
    const llm: LLMProvider = {
      async chat() {
        throw new Error("LLM failed");
      },
    };
    const enhancer = new VisionEnhancer(llm);
    const docs: Record<string, ChunkDocument> = {
      "id-1": makeDoc('<img src="http://img.com/fail.jpg">'),
    };
    const result = await enhancer.enhance(docs);
    expect(result["id-1"].description).toBe('<img src="http://img.com/fail.jpg">');
  });
});
