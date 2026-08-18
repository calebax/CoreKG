import { describe, it, expect, vi, beforeEach } from "vitest";
import { ChunkerService, type ChunkerOptions, type ChunkerDeps } from "./chunker-service.js";
import type { StorageProvider } from "@corekg/storage";
import type { LLMProvider, EmbeddingProvider } from "@corekg/ai";
import { writeFile, mkdir } from "node:fs/promises";
import { join } from "node:path";

function createMockDeps(mdContent: string): ChunkerDeps {
  const storage: StorageProvider = {
    async downloadFile(_url, destDir, _filename?) {
      await mkdir(destDir, { recursive: true });
      const dest = join(destDir, "test.md");
      await writeFile(dest, mdContent, "utf-8");
      return dest;
    },
    async uploadFile() {
      return { url: "", key: "" };
    },
    async uploadDirectory() {
      return [];
    },
    getEndpoint() {
      return "http://localhost";
    },
  };

  const llm: LLMProvider = {
    async chat() {
      return "smart";
    },
  };

  const embedding: EmbeddingProvider = {
    async embed() {
      return [0.1, 0.2, 0.3];
    },
    async embedBatch(texts) {
      return texts.map(() => [0.1, 0.2, 0.3]);
    },
  };

  return { storage, llm, embedding };
}

const defaultOptions: ChunkerOptions = {
  url: "http://example.com/test.md",
  forestId: "forest-1",
  companyId: "company-1",
  uin: "uin-1",
  fileId: "file-1",
  fileName: "test.md",
  fileExt: ".md",
  indexName: "test-index",
  removeEmail: false,
  removeUrl: false,
  removeEmptyLine: false,
  splitMode: "basic",
  chunkTokenNum: 1024,
  minChunkTokens: 10,
  splitLevel: 2,
  overlapRatio: 0,
  regexPattern: "",
  enableHeadingInContent: false,
  llmConcurrency: 3,
  embeddingConcurrency: 10,
};

describe("ChunkerService", () => {
  it("processes markdown into chunk documents", async () => {
    const content = "# Title\n\nSome paragraph text here.\n\nAnother paragraph.";
    const deps = createMockDeps(content);
    const service = new ChunkerService(deps);
    const docs = await service.process(defaultOptions);

    const entries = Object.entries(docs);
    expect(entries.length).toBeGreaterThan(0);
    for (const [, doc] of entries) {
      expect(doc.forest_id).toBe("forest-1");
      expect(doc.company_id).toBe("company-1");
      expect(doc.embedding).not.toBeNull();
      expect(doc.description_hash).toBeTruthy();
      expect(doc.version).toBe("V2.0");
    }
  });

  it("sets type=table for table chunks", async () => {
    const content = `| Name | Age |
| --- | --- |
| Alice | 30 |
| Bob | 25 |`;
    const deps = createMockDeps(content);
    const service = new ChunkerService(deps);
    const docs = await service.process(defaultOptions);
    const entries = Object.values(docs);
    const tableChunks = entries.filter((d) => d.type === "table");
    expect(tableChunks.length).toBeGreaterThan(0);
  });

  it("sets type=video for .mp4 files", async () => {
    const content = "Some content about a video.";
    const deps = createMockDeps(content);
    const service = new ChunkerService(deps);
    const docs = await service.process({ ...defaultOptions, fileExt: ".mp4" });
    const entries = Object.values(docs);
    expect(entries.every((d) => d.type === "video")).toBe(true);
  });

  it("skips empty chunks", async () => {
    const content = "   \n\n  \n\n   ";
    const deps = createMockDeps(content);
    const service = new ChunkerService(deps);
    const docs = await service.process(defaultOptions);
    expect(Object.keys(docs).length).toBe(0);
  });
});
