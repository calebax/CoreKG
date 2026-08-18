import { readFile, rm, mkdir, writeFile } from "node:fs/promises";
import { join } from "node:path";
import type { TaskHandlerFn } from "../types.js";
import { agentRequest } from "../agent-client.js";
import { extractMarkdownTitles, extractCodeBlock, processEmbeddedUuid } from "../markdown-utils.js";

const handler: TaskHandlerFn = async (ctx, payload) => {
  const tmpDir = join("/tmp", `yg_mindmap_${Date.now()}`);
  try {
    await mkdir(tmpDir, { recursive: true });
    const filePath = await ctx.storage.downloadFile(payload.file_url, tmpDir);
    const content = await readFile(filePath, "utf-8");
    const titles = extractMarkdownTitles(content).join("\n");
    const resp = await agentRequest(
      { input1: titles },
      ctx.agentConfig,
      ctx.agentConfig.pool["mindmapMD"] || "default-model",
      ctx.agentConfig.pool["mindChunkMD"] || "",
      ctx.agentConfig.pool["mergeMindmapMD"] || "",
    );
    const jsonCode = extractCodeBlock("json", resp);
    const withUuids = processEmbeddedUuid(jsonCode);
    const uploadPath = payload.upload_path || "";
    const outPath = join(tmpDir, "mindmap.json");
    await writeFile(outPath, withUuids, "utf-8");
    await ctx.storage.uploadFile(outPath, uploadPath, payload.bucket);
    return { status: "success", result: uploadPath };
  } catch (err) {
    return { status: "fail", error: err instanceof Error ? err.message : String(err) };
  } finally {
    await rm(tmpDir, { recursive: true, force: true }).catch(() => {});
  }
};

export const mindmapHandler = handler;
