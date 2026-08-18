import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("node:fs/promises", () => ({
  readFile: vi.fn(),
}));
vi.mock("node:child_process", () => ({
  execSync: vi.fn(),
}));

import { readFile } from "node:fs/promises";
import { handleTextFile, handleCSVFile, handleJSONFile, handleVideoFile } from "./content-handlers.js";

describe("handleTextFile", () => {
  beforeEach(() => vi.clearAllMocks());

  it("should return file content as-is", async () => {
    vi.mocked(readFile).mockResolvedValue("hello world");
    const result = await handleTextFile("/tmp/test.txt");
    expect(result).toBe("hello world");
  });
});

describe("handleCSVFile", () => {
  beforeEach(() => vi.clearAllMocks());

  it("should convert CSV to markdown table", async () => {
    vi.mocked(readFile).mockResolvedValue("Name,Age\nAlice,30\nBob,25");
    const result = await handleCSVFile("/tmp/test.csv");
    expect(result).toContain("| Name | Age |");
    expect(result).toContain("| --- | --- |");
    expect(result).toContain("| Alice | 30 |");
    expect(result).toContain("| Bob | 25 |");
  });
});

describe("handleJSONFile", () => {
  beforeEach(() => vi.clearAllMocks());

  it("should convert JSON to markdown", async () => {
    vi.mocked(readFile).mockResolvedValue(JSON.stringify({ name: "test", value: 42 }));
    const result = await handleJSONFile("/tmp/test.json");
    expect(result).toContain("## JSON Data");
    expect(result).toContain("### name");
  });
});

describe("handleVideoFile", () => {
  it("should call ffmpeg", () => {
    const result = handleVideoFile("/tmp/test.mp4", "/tmp/output");
    expect(result).toBe("/tmp/output/images");
  });
});
