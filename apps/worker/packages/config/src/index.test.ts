import { describe, it, expect } from "vitest";
import { AppConfigSchema } from "./schema.js";

describe("AppConfigSchema", () => {
  const validConfig = {
    nats: { url: "nats://localhost:4222" },
    es: { host: "https://es.example.com", username: "elastic", password: "pass" },
    s3: {
      endpointUrl: "https://s3.example.com",
      accessKeyId: "ak",
      secretAccessKey: "sk",
      defaultBucket: "test",
    },
    llm: { apiKey: "sk-xxx", baseUrl: "https://api.example.com/v1" },
    vllm: { apiKey: "sk-xxx", baseUrl: "https://api.example.com/v1" },
    embedding: { apiKey: "sk-xxx", baseUrl: "https://emb.example.com/v1" },
    concurrency: {},
  };

  it("parses valid config with defaults", () => {
    const result = AppConfigSchema.parse(validConfig);
    expect(result.nats.stream).toBe("CORE_TASKS");
    expect(result.llm.model).toBe("deepseek-v3");
    expect(result.concurrency.embeddingWorkers).toBe(30);
  });

  it("rejects invalid URL", () => {
    expect(() =>
      AppConfigSchema.parse({
        ...validConfig,
        es: { ...validConfig.es, host: "not-a-url" },
      }),
    ).toThrow();
  });
});
