import { describe, it, expect } from "vitest";
import { createESProvider } from "./es-provider.js";

describe("createESProvider", () => {
  it("creates provider with valid config", () => {
    const provider = createESProvider({
      host: "https://es.example.com",
      username: "elastic",
      password: "pass",
      poolSize: 10,
      requestTimeoutMs: 30000,
    });
    expect(provider.vectorStore).toBeDefined();
    expect(provider.search).toBeDefined();
    expect(typeof provider.vectorStore.upsertChunks).toBe("function");
    expect(typeof provider.vectorStore.deleteChunksByFileId).toBe("function");
    expect(typeof provider.search.getById).toBe("function");
    expect(provider.close).toBeDefined();
  });
});
