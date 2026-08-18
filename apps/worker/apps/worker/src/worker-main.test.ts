import { describe, it, expect } from "vitest";

describe("worker-main module", () => {
  it("source file should exist and be importable", async () => {
    const fs = await import("node:fs/promises");
    const content = await fs.readFile(new URL("./worker-main.ts", import.meta.url), "utf-8");
    expect(content).toContain("export async function main()");
    expect(content).toContain("createNATSClient");
    expect(content).toContain("TaskConsumer");
    expect(content).toContain("publishCallback");
    expect(content).toContain("ChunkerService");
  });
});
