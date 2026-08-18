import type { TaskHandlerFn } from "../types.js";
import { algoSplit } from "../algo-client.js";

const handler: TaskHandlerFn = async (ctx, payload) => {
  const esIndex = payload.es_index || "";
  try {
    if (esIndex) {
      await ctx.es.vectorStore.deleteChunksByFileId(
        esIndex,
        String(payload.forest_id),
        String(payload.file_id),
        String(payload.company_id),
      );
    }

    const result = await algoSplit(ctx.algoUrl, {
      uin: String(payload.uin),
      companyId: String(payload.company_id),
      forestId: String(payload.forest_id),
      fileId: String(payload.file_id),
      content: payload.file_url,
      esIndex,
      fileExt: payload.file_ext || "",
    });

    await new Promise((r) => setTimeout(r, 2000));

    if (esIndex) {
      await ctx.es.search.queryChunkIdsByFileId(esIndex, String(payload.file_id));
    }

    return { status: "success", result };
  } catch (err) {
    return { status: "fail", error: err instanceof Error ? err.message : String(err) };
  }
};

export const splitTextChunkHandler = handler;
