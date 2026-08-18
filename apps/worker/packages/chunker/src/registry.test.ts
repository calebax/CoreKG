import { describe, it, expect, beforeEach } from "vitest";
import {
  registerStrategy,
  resolveStrategy,
  listStrategies,
  registerBuiltinStrategies,
} from "./registry.js";
import type { ChunkStrategy } from "./strategy.js";

describe("registry", () => {
  beforeEach(() => {
    for (const name of listStrategies()) {
      // clear by re-registering — but Map has no delete exposed,
      // so we just register builtins fresh
    }
  });

  it("registerBuiltinStrategies registers all 8 strategies", () => {
    registerBuiltinStrategies();
    const names = listStrategies();
    expect(names).toContain("basic");
    expect(names).toContain("smart");
    expect(names).toContain("title");
    expect(names).toContain("strict-regex");
    expect(names).toContain("slide");
    expect(names).toContain("resume");
    expect(names).toContain("llm");
    expect(names).toContain("auto");
  });

  it("resolveStrategy returns registered strategy", () => {
    registerBuiltinStrategies();
    const s = resolveStrategy("basic");
    expect(s.name).toBe("basic");
  });

  it("resolveStrategy throws for unknown strategy", () => {
    registerBuiltinStrategies();
    expect(() => resolveStrategy("nonexistent")).toThrow("Unknown chunk strategy: nonexistent");
  });

  it("registerStrategy allows custom strategies", () => {
    const custom: ChunkStrategy = {
      name: "custom-test",
      split() {
        return { chunks: [], metas: [] };
      },
    };
    registerStrategy(custom);
    expect(resolveStrategy("custom-test")).toBe(custom);
    expect(listStrategies()).toContain("custom-test");
  });
});
