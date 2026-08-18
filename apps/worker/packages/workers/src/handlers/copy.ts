import { readFile, rm, mkdir, writeFile } from "node:fs/promises";
import { join } from "node:path";
import type { TaskHandlerFn } from "../types.js";

const handler: TaskHandlerFn = async (ctx, payload) => {
  const tmpDir = join("/tmp", `yg_copy_${Date.now()}`);
  try {
    await mkdir(tmpDir, { recursive: true });
    const filePath = await ctx.storage.downloadFile(payload.file_url, tmpDir);
    const content = await readFile(filePath, "utf-8");
    const uploadPath = (payload.upload_path || "") + "content.md";
    const outPath = join(tmpDir, "content.md");
    await writeFile(outPath, content, "utf-8");
    await ctx.storage.uploadFile(outPath, uploadPath, payload.bucket);
    return { status: "success", result: payload.upload_path };
  } catch (err) {
    return { status: "fail", error: err instanceof Error ? err.message : String(err) };
  } finally {
    await rm(tmpDir, { recursive: true, force: true }).catch(() => {});
  }
};

export const copyHandler = handler;
