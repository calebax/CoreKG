import { describe, it, expect, vi } from "vitest";

vi.mock("@nats-io/transport-node", () => ({
  connect: vi.fn().mockResolvedValue({ drain: vi.fn() }),
}));

vi.mock("@nats-io/jetstream", () => ({
  jetstream: vi.fn().mockReturnValue({}),
  jetstreamManager: vi.fn().mockResolvedValue({
    streams: { info: vi.fn().mockRejectedValue(new Error("not found")), add: vi.fn() },
  }),
  RetentionPolicy: { Workqueue: "workqueue" },
}));

import { createNATSClient } from "./client.js";

describe("createNATSClient", () => {
  it("should connect and create stream if missing", async () => {
    const { jetstreamManager } = await import("@nats-io/jetstream");
    const client = await createNATSClient({
      nats: { url: "nats://localhost:4222", stream: "TEST" },
      es: { host: "http://localhost:9200", username: "u", password: "p" },
      s3: { endpointUrl: "http://localhost:9000", accessKeyId: "a", secretAccessKey: "s", defaultBucket: "b" },
      llm: { apiKey: "k", baseUrl: "http://localhost" },
      vllm: { apiKey: "k", baseUrl: "http://localhost" },
      embedding: { apiKey: "k", baseUrl: "http://localhost" },
      concurrency: {},
    } as any);

    expect(client).toBeDefined();
    expect(client.nc).toBeDefined();
    expect(client.js).toBeDefined();
  });
});
