import { describe, it, expect } from "vitest";
import { createEmbeddingProvider } from "./embedding-provider.js";

describe("createEmbeddingProvider", () => {
  it("creates provider with valid config", () => {
    const provider = createEmbeddingProvider({
      apiKey: "sk-test",
      baseUrl: "https://emb.example.com/v1",
      model: "Qwen3-Embedding-0.6B",
      timeoutMs: 30000,
    });
    expect(provider).toBeDefined();
    expect(typeof provider.embed).toBe("function");
    expect(typeof provider.embedBatch).toBe("function");
  });
});
