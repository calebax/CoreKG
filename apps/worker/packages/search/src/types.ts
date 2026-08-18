export interface ChunkDocument {
  forest_id: string;
  company_id: string;
  uin: string;
  file_id: string;
  version: string;
  file_name: string | null;
  type: "chunk" | "table" | "image" | "video" | "entity" | "file_description" | null;
  tokens: number;
  chunk_size: number;
  sequence: number;
  location: unknown | null;
  yg_location: unknown | null;
  description: string;
  description_hash: string;
  embedding: number[] | null;
  image_url: string | null;
  image_embedding: number[] | null;
  formula: string | null;
  table: string | null;
  title_level_ids: string[] | null;
  title_level: string[] | null;
  references: unknown | null;
  graph_info: unknown | null;
  graph_status: unknown | null;
}

export interface VectorStore {
  upsertChunks(index: string, docs: Record<string, ChunkDocument>): Promise<void>;
  deleteChunksByFileId(index: string, forestId: string, fileId: string, companyId: string): Promise<number>;
  insertDocument(index: string, id: string, doc: Record<string, unknown>): Promise<void>;
  deleteByType(index: string, forestId: string | number, fileId: string | number, type: string): Promise<number>;
}

export interface SearchProvider {
  getById(index: string, id: string): Promise<Record<string, unknown> | null>;
  query(index: string, body: Record<string, unknown>): Promise<Record<string, unknown>[]>;
  queryChunkIdsByFileId(index: string, fileId: string | number, limit?: number): Promise<string[]>;
}

export interface ESProvider {
  vectorStore: VectorStore;
  search: SearchProvider;
  close(): Promise<void>;
}
