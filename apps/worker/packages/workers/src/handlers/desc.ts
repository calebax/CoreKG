import { readFile } from "node:fs/promises";
import type { TaskHandlerFn } from "../types.js";
import { agentRequest } from "../agent-client.js";
import { extractMarkdownTitles, extractCodeBlock, processEmbeddedUuid } from "../markdown-utils.js";
import { v4 as uuidv4 } from "uuid";

const handler: TaskHandlerFn = async (ctx, payload) => {
  const esIndex = payload.es_index || "";
  try {
    if (esIndex) {
      await ctx.es.vectorStore.deleteByType(
        esIndex,
        payload.forest_id,
        payload.file_id,
        "file_description",
      );
    }

    const tmpDir = `/tmp/yg_desc_${Date.now()}`;
    const fs = await import("node:fs/promises");
    await fs.mkdir(tmpDir, { recursive: true });
    const filePath = await ctx.storage.downloadFile(payload.file_url, tmpDir);
    const content = await readFile(filePath, "utf-8");

    const record: Record<string, unknown> = {
      forest_id: String(payload.forest_id),
      company_id: String(payload.company_id),
      uin: String(payload.uin),
      file_id: String(payload.file_id),
      type: "file_description",
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      version: "",
    };

    let abstractText = "";
    const abstractReady = new Promise<string>((resolve) => {
      const check = setInterval(() => {
        if (abstractText) {
          clearInterval(check);
          resolve(abstractText);
        }
      }, 100);
      setTimeout(() => { clearInterval(check); resolve(""); }, 600_000);
    });

    const doMindMap = async (): Promise<void> => {
      const titles = extractMarkdownTitles(content).join("\n");
      const resp = await agentRequest(
        { input1: titles },
        ctx.agentConfig,
        ctx.agentConfig.pool["mindmapMD"] || "default-model",
        ctx.agentConfig.pool["mindChunkMD"] || "",
        ctx.agentConfig.pool["mergeMindmapMD"] || "",
      );
      const jsonCode = extractCodeBlock("json", resp);
      record.mind_map = processEmbeddedUuid(jsonCode);
    };

    const doAbstract = async (): Promise<void> => {
      const resp = await agentRequest(
        { input1: content },
        ctx.agentConfig,
        ctx.agentConfig.pool["abstractMD"] || "default-model",
        ctx.agentConfig.pool["absChunkMD"] || "",
        ctx.agentConfig.pool["mergeAbstractMD"] || "",
      );
      abstractText = resp;
      record.abstract = resp;
    };

    const doDescription = async (): Promise<void> => {
      const abs = await abstractReady;
      if (!abs) return;
      const resp = await agentRequest(
        { input1: abs },
        ctx.agentConfig,
        ctx.agentConfig.pool["shortDescMD"] || "default-model",
      );
      record.description = resp;
      const emb = await ctx.embedding.embed(resp);
      record.embedding = emb || null;
    };

    await Promise.all([doMindMap(), doAbstract(), doDescription()]);

    if (esIndex) {
      const docId = uuidv4();
      await ctx.es.vectorStore.insertDocument(esIndex, docId, record);
    }

    const result = `${record.description || ""}\n${record.mind_map || ""}\n${record.abstract || ""}`;
    return { status: "success", result };
  } catch (err) {
    return { status: "fail", error: err instanceof Error ? err.message : String(err) };
  }
};

export const descHandler = handler;
