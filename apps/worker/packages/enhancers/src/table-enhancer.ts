import type { LLMProvider } from "@corekg/ai";
import type { ChunkDocument } from "@corekg/search";

export interface TableEnhancerOptions {
  maxConcurrency?: number;
}

export class TableEnhancer {
  constructor(private llm: LLMProvider, private options: TableEnhancerOptions = {}) {}

  async enhance(
    docs: Record<string, ChunkDocument>,
  ): Promise<Record<string, ChunkDocument>> {
    const concurrency = this.options.maxConcurrency ?? 3;
    const sem = new Semaphore(concurrency);

    const chunkIds = Object.keys(docs);
    const tableTasks: Array<{ uid: string; content: string; context: string | null }> = [];

    for (let idx = 0; idx < chunkIds.length; idx++) {
      const chunk = docs[chunkIds[idx]];
      if (chunk.type !== "table") continue;

      let context: string | null = null;
      for (let i = idx - 1; i >= 0; i--) {
        const prev = docs[chunkIds[i]];
        if (prev.type === "chunk" && prev.description.trim()) {
          context = prev.description.slice(-500);
          break;
        }
      }

      tableTasks.push({ uid: chunkIds[idx], content: chunk.description, context });
    }

    if (tableTasks.length === 0) return docs;

    const descriptions = new Map<string, string>();
    const tasks: Array<Promise<void>> = [];

    for (const task of tableTasks) {
      tasks.push(
        sem.run(async () => {
          try {
            const prompt = `Generate a brief summary for this table (max 100 words). Context: ${task.context || "none"}\n\nTable content:\n${task.content.slice(0, 3000)}`;
            const result = await this.llm.chat(prompt);
            const cleaned = result.trim();
            if (cleaned) descriptions.set(task.uid, cleaned);
          } catch {
            // skip failed
          }
        }),
      );
    }

    await Promise.all(tasks);

    for (const { uid } of tableTasks) {
      const desc = descriptions.get(uid);
      if (!desc) continue;
      const original = docs[uid].description;
      const tablePos = original.indexOf("<table");
      if (tablePos !== -1) {
        docs[uid].description =
          original.slice(0, tablePos) + `[表格摘要: ${desc}]\n` + original.slice(tablePos);
      } else {
        docs[uid].description = `[表格摘要: ${desc}]\n${original}`;
      }
      docs[uid].table = original;
    }

    return docs;
  }
}

class Semaphore {
  private _queue: Array<() => void> = [];
  private _active = 0;

  constructor(private max: number) {}

  run<T>(fn: () => Promise<T>): Promise<T> {
    return new Promise<T>((resolve, reject) => {
      const execute = () => {
        this._active++;
        fn()
          .then((v) => {
            this._active--;
            this._next();
            resolve(v);
          })
          .catch((e) => {
            this._active--;
            this._next();
            reject(e);
          });
      };

      if (this._active < this.max) {
        execute();
      } else {
        this._queue.push(execute);
      }
    });
  }

  private _next() {
    if (this._queue.length > 0 && this._active < this.max) {
      const next = this._queue.shift()!;
      next();
    }
  }
}
