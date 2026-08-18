import { readFile, rm, mkdir, writeFile } from "node:fs/promises";
import { join } from "node:path";
import type { TaskHandlerFn } from "../types.js";
import { agentRequest } from "../agent-client.js";

const handler: TaskHandlerFn = async (ctx, payload) => {
  const tmpDir = join("/tmp", `yg_analysis_${Date.now()}`);
  try {
    await mkdir(tmpDir, { recursive: true });
    const filePath = await ctx.storage.downloadFile(payload.file_url, tmpDir);
    const content = await readFile(filePath, "utf-8");
    const resp = await agentRequest(
      { input1: content },
      ctx.agentConfig,
      ctx.agentConfig.pool["analysisMD"] || ctx.agentConfig.pool["abstractMD"] || "default-model",
    );
    const uploadPath = payload.upload_path || "";
    const outPath = join(tmpDir, "result.txt");
    await writeFile(outPath, resp, "utf-8");
    await ctx.storage.uploadFile(outPath, uploadPath, payload.bucket);
    return { status: "success", result: uploadPath };
  } catch (err) {
    return { status: "fail", error: err instanceof Error ? err.message : String(err) };
  } finally {
    await rm(tmpDir, { recursive: true, force: true }).catch(() => {});
  }
};

export const analysisHandler = handler;
