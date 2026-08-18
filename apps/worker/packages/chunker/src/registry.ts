import type { ChunkStrategy, AsyncChunkStrategy } from "./strategy.js";
import { basicStrategy } from "./strategies/basic.js";
import { smartStrategy } from "./strategies/smart.js";
import { titleStrategy } from "./strategies/title.js";
import { strictRegexStrategy } from "./strategies/strict-regex.js";
import { slideStrategy } from "./strategies/slide.js";
import { resumeStrategy } from "./strategies/resume.js";
import { llmStrategy } from "./strategies/llm.js";
import { autoStrategy } from "./strategies/auto.js";

const strategies = new Map<string, ChunkStrategy | AsyncChunkStrategy>();

export function registerStrategy(s: ChunkStrategy | AsyncChunkStrategy): void {
  strategies.set(s.name, s);
}

export function resolveStrategy(name: string): ChunkStrategy | AsyncChunkStrategy {
  const s = strategies.get(name);
  if (!s) {
    throw new Error(
      `Unknown chunk strategy: ${name}. Available: ${[...strategies.keys()].join(", ")}`,
    );
  }
  return s;
}

export function listStrategies(): string[] {
  return [...strategies.keys()];
}

export function registerBuiltinStrategies(): void {
  registerStrategy(basicStrategy);
  registerStrategy(smartStrategy);
  registerStrategy(titleStrategy);
  registerStrategy(strictRegexStrategy);
  registerStrategy(slideStrategy);
  registerStrategy(resumeStrategy);
  registerStrategy(llmStrategy);
  registerStrategy(autoStrategy);
}
