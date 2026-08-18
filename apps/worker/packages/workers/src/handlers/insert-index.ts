import type { TaskHandlerFn } from "../types.js";
import { algoIndex } from "../algo-client.js";

const handler: TaskHandlerFn = async (ctx, payload) => {
  const esIndex = payload.es_index || "";
  try {
    if (esIndex) {
      const chunkIds = await ctx.es.search.queryChunkIdsByFileId(esIndex, String(payload.file_id));
      if (chunkIds.length === 0) {
        return { status: "success", result: null };
      }
    }

    await algoIndex(ctx.algoUrl, {
      uin: String(payload.uin),
      companyId: String(payload.company_id),
      forestId: String(payload.forest_id),
      fileId: String(payload.file_id),
      esIndex,
    });

    return { status: "success", result: null };
  } catch (err) {
    return { status: "fail", error: err instanceof Error ? err.message : String(err) };
  }
};

export const insertIndexHandler = handler;
