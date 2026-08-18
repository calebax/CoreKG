import { describe, it, expect, vi, beforeEach } from "vitest";
import AdmZip from "adm-zip";

vi.mock("node:fs/promises", () => ({
  readFile: vi.fn(),
}));

vi.mock("mime-types", () => ({
  default: { lookup: vi.fn().mockReturnValue("application/pdf") },
}));

import { readFile } from "node:fs/promises";
import { processPdfWithMinerU } from "./mineru-client.js";

describe("processPdfWithMinerU", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should upload file and extract content_list.json from ZIP response", async () => {
    const zip = new AdmZip();
    const contentList = [
      { type: "text", text: "hello", page_idx: 0, bbox: [0, 0, 100, 100] },
    ];
    zip.addFile("output/content_list.json", Buffer.from(JSON.stringify(contentList)));
    const zipBuffer = zip.toBuffer();

    vi.mocked(readFile).mockImplementation((async (path: any) => {
      if (typeof path === "string" && path.endsWith("test.pdf")) {
        return Buffer.from("fake-pdf");
      }
      return JSON.stringify(contentList);
    }) as any);

    const originalFetch = globalThis.fetch;
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      arrayBuffer: () => Promise.resolve(zipBuffer.buffer.slice(
        zipBuffer.byteOffset,
        zipBuffer.byteOffset + zipBuffer.byteLength,
      )),
    }) as any;

    const result = await processPdfWithMinerU("/tmp/test.pdf", "/tmp/output", {
      apiUrl: "http://mineru:8080/api",
    });

    expect(result.contentList).toEqual(contentList);
    expect(result.outputDir).toBe("/tmp/output");

    globalThis.fetch = originalFetch;
  });

  it("should throw on HTTP error", async () => {
    vi.mocked(readFile).mockResolvedValue(Buffer.from("fake-pdf"));

    const originalFetch = globalThis.fetch;
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      statusText: "Internal Server Error",
    }) as any;

    await expect(
      processPdfWithMinerU("/tmp/test.pdf", "/tmp/output", {
        apiUrl: "http://mineru:8080/api",
      }),
    ).rejects.toThrow("MinerU API error");

    globalThis.fetch = originalFetch;
  });
});
