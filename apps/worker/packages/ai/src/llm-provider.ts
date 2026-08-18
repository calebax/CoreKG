import { generateText, type ModelMessage } from "ai";
import { createOpenAICompatible } from "@ai-sdk/openai-compatible";
import type { z } from "zod";
import type { LLMConfigSchema } from "@corekg/config";

type LLMConfig = z.infer<typeof LLMConfigSchema>;

export interface LLMChatOptions {
  model?: string;
  temperature?: number;
  systemPrompt?: string;
  history?: Array<{ role: "system" | "user" | "assistant"; content: string }>;
  contentParts?: Array<
    | { type: "text"; text: string }
    | { type: "image_url"; image_url: { url: string } }
  >;
}

export interface LLMProvider {
  chat(prompt: string, options?: LLMChatOptions): Promise<string>;
}

export function createLLMProvider(config: LLMConfig): LLMProvider {
  const provider = createOpenAICompatible({
    name: "custom-llm",
    baseURL: config.baseUrl,
    apiKey: config.apiKey,
  });

  return {
    async chat(prompt, opts = {}) {
      const messages: ModelMessage[] = [];
      if (opts.systemPrompt) {
        messages.push({ role: "system", content: opts.systemPrompt });
      }
      if (opts.history) {
        messages.push(...opts.history);
      }

      if (opts.contentParts) {
        messages.push({ role: "user", content: opts.contentParts as any });
      } else {
        messages.push({ role: "user", content: prompt });
      }

      const { text } = await generateText({
        model: provider(opts.model || config.model),
        messages,
        temperature: opts.temperature ?? 1,
        abortSignal: AbortSignal.timeout(config.timeoutMs),
      });
      return text;
    },
  };
}
