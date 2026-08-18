import { smartStrategy } from "./smart.js";
import { basicStrategy } from "./basic.js";
import { titleStrategy } from "./title.js";
import { slideStrategy } from "./slide.js";
import { resumeStrategy } from "./resume.js";
import { strictRegexStrategy } from "./strict-regex.js";
import type { AsyncChunkStrategy, ChunkStrategyOptions, ChunkSplitResult } from "../strategy.js";
import type { LLMProvider } from "@corekg/ai";

const STRATEGY_MAP: Record<string, import("../strategy.js").ChunkStrategy> = {
  smart: smartStrategy,
  basic: basicStrategy,
  title: titleStrategy,
  slide: slideStrategy,
  resume: resumeStrategy,
  "strict-regex": strictRegexStrategy,
  strict_regex: strictRegexStrategy,
};

export const autoStrategy: AsyncChunkStrategy = {
  name: "auto",

  split(content: string, options: ChunkStrategyOptions): ChunkSplitResult {
    return smartStrategy.split(content, options);
  },

  async splitAsync(
    content: string,
    options: ChunkStrategyOptions,
    deps: { llm: LLMProvider },
  ): Promise<ChunkSplitResult> {
    if (!content || !content.trim()) {
      return { chunks: [], metas: [] };
    }

    const snippet = content.slice(0, 2000);
    const prompt = `Analyze this document snippet and choose the best chunking strategy from: smart, basic, title, slide, resume, strict-regex. Reply with ONLY the strategy name, nothing else.\n\n${snippet}`;

    let strategyName = "smart";
    try {
      const response = await deps.llm.chat(prompt);
      const normalized = response.trim().toLowerCase().replace(/[^a-z_-]/g, "");
      if (normalized in STRATEGY_MAP) {
        strategyName = normalized;
      }
    } catch {
      strategyName = "smart";
    }

    const strategy = STRATEGY_MAP[strategyName] || smartStrategy;
    return strategy.split(content, options);
  },
};
