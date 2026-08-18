import { describe, it, expect } from "vitest";
import { TaskPayloadSchema } from "./types.js";

describe("TaskPayloadSchema", () => {
  it("should parse valid payload", () => {
    const input = {
      task_type: "ke.knowledge_task",
      file_id: "123",
      file_url: "http://example.com/file.md",
      company_id: 1,
      forest_id: "f1",
      uin: "u1",
    };
    const result = TaskPayloadSchema.parse(input);
    expect(result.task_type).toBe("ke.knowledge_task");
    expect(result.file_id).toBe("123");
  });

  it("should parse payload with split_config", () => {
    const input = {
      task_type: "ke.knowledge_task",
      file_id: "123",
      file_url: "http://example.com/file.md",
      company_id: "c1",
      forest_id: "f1",
      uin: "u1",
      split_config: {
        split_mode: "title",
        chunk_token_num: 512,
        preprocessing_rules: { remove_email: false, remove_url: true, remove_empty_line: true },
      },
    };
    const result = TaskPayloadSchema.parse(input);
    expect(result.split_config?.split_mode).toBe("title");
    expect(result.split_config?.preprocessing_rules?.remove_email).toBe(false);
  });

  it("should reject missing required fields", () => {
    expect(() => TaskPayloadSchema.parse({})).toThrow();
  });
});


