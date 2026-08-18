import { Client } from "@elastic/elasticsearch";
import type { z } from "zod";
import type { ESConfigSchema } from "@corekg/config";
import type { ChunkDocument, VectorStore, SearchProvider, ESProvider } from "./types.js";

export type { ChunkDocument, VectorStore, SearchProvider, ESProvider } from "./types.js";

type ESConfig = z.infer<typeof ESConfigSchema>;

export function createESProvider(config: ESConfig): ESProvider {
  const client = new Client({
    node: config.host,
    auth: { username: config.username, password: config.password },
    maxRetries: 3,
    requestTimeout: config.requestTimeoutMs,
    sniffOnStart: false,
  });

  const vectorStore: VectorStore = {
    async upsertChunks(index, documents) {
      const actions = Object.entries(documents).flatMap(([id, doc]) => [
        { index: { _index: index, _id: id } },
        doc,
      ]);
      await client.bulk({ operations: actions, refresh: "wait_for" });
    },

    async deleteChunksByFileId(index, forestId, fileId, companyId) {
      const result = await client.deleteByQuery({
        index,
        query: {
          bool: {
            filter: [
              { term: { forest_id: forestId } },
              { term: { file_id: fileId } },
              { term: { company_id: companyId } },
            ],
          },
        },
        conflicts: "proceed",
      });
      return (result.deleted ?? 0) as number;
    },

    async insertDocument(index, id, doc) {
      await client.index({ index, id, body: doc, refresh: "wait_for" });
    },

    async deleteByType(index, forestId, fileId, type) {
      const result = await client.deleteByQuery({
        index,
        query: {
          bool: {
            filter: [
              { term: { forest_id: forestId } },
              { term: { file_id: fileId } },
              { term: { type } },
            ],
          },
        },
        conflicts: "proceed",
      });
      return (result.deleted ?? 0) as number;
    },
  };

  const search: SearchProvider = {
    async getById(index, id) {
      try {
        const result = await client.get({ index, id });
        return result._source as Record<string, unknown> | null;
      } catch {
        return null;
      }
    },

    async query(index, body) {
      const result = await client.search({ index, ...body as any });
      return result.hits.hits.map((h) => h._source as Record<string, unknown>);
    },

    async queryChunkIdsByFileId(index, fileId, limit = 1000) {
      const result = await client.search({
        index,
        size: limit,
        _source: false,
        query: {
          bool: {
            filter: [
              { term: { file_id: fileId } },
            ],
          },
        },
      });
      return result.hits.hits.map((h) => h._id as string);
    },
  };

  return {
    vectorStore,
    search,
    async close() {
      await client.close();
    },
  };
}
