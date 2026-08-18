import type { LLMProvider } from "@corekg/ai";
import type { ChunkDocument } from "@corekg/search";

const IMG_HTML_RE = /<img\s+src="([^"]+)"[^>]*>/g;
const IMG_MD_RE = /!\[[^\]]*\]\(([^)]+)\)/g;

export interface ImageReference {
  fullMatch: string;
  url: string;
  index: number;
}

export function extractImageReferences(text: string): ImageReference[] {
  const refs: ImageReference[] = [];
  for (const m of text.matchAll(IMG_HTML_RE)) {
    refs.push({ fullMatch: m[0], url: m[1], index: m.index! });
  }
  for (const m of text.matchAll(IMG_MD_RE)) {
    refs.push({ fullMatch: m[0], url: m[1], index: m.index! });
  }
  refs.sort((a, b) => a.index - b.index);
  return refs;
}

export interface VisionEnhancerOptions {
  descriptionFormat?: string;
  maxConcurrency?: number;
}

export class VisionEnhancer {
  constructor(private vllm: LLMProvider, private options: VisionEnhancerOptions = {}) {}

  async enhance(
    docs: Record<string, ChunkDocument>,
  ): Promise<Record<string, ChunkDocument>> {
    const format = this.options.descriptionFormat ?? "[图片描述: {desc}]";
    const concurrency = this.options.maxConcurrency ?? 3;
    const sem = new Semaphore(concurrency);

    const imageChunks: Array<{ uid: string; refs: ImageReference[] }> = [];
    for (const [uid, doc] of Object.entries(docs)) {
      const refs = extractImageReferences(doc.description);
      if (refs.length > 0) {
        imageChunks.push({ uid, refs });
      }
    }

    if (imageChunks.length === 0) return docs;

    const uniqueUrls = new Set<string>();
    for (const { refs } of imageChunks) {
      for (const ref of refs) uniqueUrls.add(ref.url);
    }

    const descriptions = new Map<string, string>();
    const tasks: Array<Promise<void>> = [];

    for (const url of uniqueUrls) {
      tasks.push(
        sem.run(async () => {
          try {
            const desc = await this.vllm.chat("Describe this image in detail.", {
              contentParts: [
                { type: "image_url", image_url: { url } },
                { type: "text", text: "Describe this image in detail. Be concise." },
              ],
            });
            const cleaned = desc.replace(/\n/g, "").trim();
            if (cleaned) descriptions.set(url, cleaned);
          } catch {
            // skip failed images
          }
        }),
      );
    }

    await Promise.all(tasks);

    for (const { uid, refs } of imageChunks) {
      let enhanced = docs[uid].description;
      const urls: string[] = [];
      for (const ref of [...refs].reverse()) {
        const desc = descriptions.get(ref.url);
        if (desc) {
          const insertion = `<br>\n${format.replace("{desc}", desc)}\n`;
          const insertPos = ref.index + ref.fullMatch.length;
          enhanced = enhanced.slice(0, insertPos) + insertion + enhanced.slice(insertPos);
        }
        urls.push(ref.url);
      }
      docs[uid].description = enhanced;
      if (urls.length > 0) {
        docs[uid].image_url = urls.join("\n");
      }
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
