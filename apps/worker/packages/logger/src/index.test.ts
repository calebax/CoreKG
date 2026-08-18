import { describe, it, expect } from "vitest";
import { createLogger } from "./index.js";

describe("createLogger", () => {
  it("creates logger with name", () => {
    const l = createLogger("test");
    expect(typeof l.info).toBe("function");
    expect(l.level).toBe("info");
  });

  it("respects custom level", () => {
    expect(createLogger("t", "debug").level).toBe("debug");
  });
});
