import { rm, mkdir } from "node:fs/promises";
import { join, extname } from "node:path";
import type { TaskHandlerFn } from "../types.js";
import { daemonProcessVideo } from "../daemon-client.js";

const handler: TaskHandlerFn = async (ctx, payload) => {
  const tmpDir = join("/tmp", `yg_video_${Date.now()}`);
  const storageDir = join(tmpDir, "storage");
  try {
    await mkdir(storageDir, { recursive: true });
    const ext = extname(new URL(payload.file_url).pathname) || ".mp4";
    const fileName = `origin${ext}`;
    const filePath = await ctx.storage.downloadFile(payload.file_url, tmpDir, fileName);
    const uploadPath = payload.upload_path || "";
    const pubEndpoint = ctx.storage.getEndpoint();
    const publicPath = `${pubEndpoint}/${payload.bucket || ""}/${uploadPath}`;

    await daemonProcessVideo(ctx.daemonUrl, {
      videoPath: filePath,
      outputDir: storageDir,
      imagePrefix: publicPath,
    });

    const results = await ctx.storage.uploadDirectory(storageDir, uploadPath, payload.bucket);
    const fileList = results.map((r: { key: string }) => r.key);
    return { status: "success", result: fileList };
  } catch (err) {
    return { status: "fail", error: err instanceof Error ? err.message : String(err) };
  } finally {
    await rm(tmpDir, { recursive: true, force: true }).catch(() => {});
  }
};

export const videoExtractHandler = handler;
