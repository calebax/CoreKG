import { embed, embedMany } from "ai";
import { createOpenAICompatible } from "@ai-sdk/openai-compatible";
import type { z } from "zod";
import type { EmbeddingConfigSchema } from "@corekg/config";

type EmbeddingConfig = z.infer<typeof EmbeddingConfigSchema>;

export interface EmbeddingProvider {
  embed(text: string, model?: string): Promise<number[] | null>;
  embedBatch(
    texts: string[],
    options?: { concurrency?: number; model?: string },
  ): Promise<Array<number[] | null>>;
}

export function createEmbeddingProvider(config: EmbeddingConfig): EmbeddingProvider {
  const provider = createOpenAICompatible({
    name: "custom-emb",
    baseURL: config.baseUrl,
    apiKey: config.apiKey,
  });

  return {
    async embed(text, modelName?) {
      if (!text.trim()) return null;
      const { embedding } = await embed({
        model: provider.embeddingModel(modelName || config.model),
        value: text,
        abortSignal: AbortSignal.timeout(config.timeoutMs),
      });
      return embedding;
    },

    async embedBatch(texts, opts = {}) {
      const valid = texts
        .map((t, i) => ({ id: i, text: t }))
        .filter((t) => t.text.trim());
      if (!valid.length) return texts.map(() => null);

      const { embeddings } = await embedMany({
        model: provider.embeddingModel(opts.model || config.model),
        values: valid.map((t) => t.text),
        maxRetries: 2,
      });

      const results: Array<number[] | null> = new Array(texts.length).fill(null);
      for (let i = 0; i < valid.length; i++) {
        results[valid[i].id] = embeddings[i];
      }
      return results;
    },
  };
}
