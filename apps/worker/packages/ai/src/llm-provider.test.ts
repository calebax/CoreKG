import { describe, it, expect } from "vitest";
import { createLLMProvider } from "./llm-provider.js";

describe("createLLMProvider", () => {
  it("creates provider with valid config", () => {
    const provider = createLLMProvider({
      apiKey: "sk-test",
      baseUrl: "https://api.example.com/v1",
      model: "deepseek-v3",
      timeoutMs: 30000,
    });
    expect(provider).toBeDefined();
    expect(typeof provider.chat).toBe("function");
  });
});
