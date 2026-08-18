import { describe, it, expect } from "vitest";
import { createS3Provider } from "./s3-provider.js";

describe("createS3Provider", () => {
  it("creates provider with valid config", () => {
    const provider = createS3Provider({
      endpointUrl: "https://s3.example.com",
      accessKeyId: "ak",
      secretAccessKey: "sk",
      defaultBucket: "test",
      region: "us-east-1",
    });
    expect(typeof provider.downloadFile).toBe("function");
    expect(typeof provider.uploadFile).toBe("function");
    expect(typeof provider.uploadDirectory).toBe("function");
    expect(provider.getEndpoint()).toBe("https://s3.example.com");
  });
});
