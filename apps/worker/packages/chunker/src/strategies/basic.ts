import markdownit from "markdown-it";
import { countTokens } from "../tokenizer.js";
import type { ChunkStrategy, ChunkStrategyOptions, ChunkSplitResult } from "../strategy.js";

const md = markdownit();

function extractTablesAndRemainder(txt: string): { remainder: string; tables: string[] } {
  const lines = txt.split("\n");
  const tables: string[] = [];
  const remainderLines: string[] = [];
  let inTable = false;
  let currentTable: string[] = [];

  for (let idx = 0; idx < lines.length; idx++) {
    const line = lines[idx];
    const strippedLine = line.trim();
    const isTableLine = strippedLine.startsWith("|") && strippedLine.endsWith("|");

    let isSeparatorLine = false;
    if (isTableLine && strippedLine.includes("-")) {
      const parts = strippedLine.slice(1, -1).split("|").map((p) => p.trim());
      if (parts.length > 0 && parts.every((p) => !p || /^[-:]+$/.test(p))) {
        isSeparatorLine = true;
      }
    }

    if (isTableLine || (inTable && strippedLine)) {
      if (!inTable && isTableLine && !isSeparatorLine) {
        let nextIsTable = false;
        if (idx + 1 < lines.length) {
          const nextStripped = lines[idx + 1].trim();
          if (nextStripped.startsWith("|") && nextStripped.endsWith("|") && nextStripped.includes("-")) {
            const partsNext = nextStripped.slice(1, -1).split("|").map((p) => p.trim());
            if (partsNext.length > 0 && partsNext.every((p) => !p || /^[-:]+$/.test(p))) {
              nextIsTable = true;
            }
          }
        }
        if (nextIsTable) {
          inTable = true;
          currentTable.push(line);
        } else {
          remainderLines.push(line);
        }
      } else if (inTable) {
        currentTable.push(line);
        if (!isTableLine && !strippedLine) {
          tables.push(currentTable.join("\n"));
          currentTable = [];
          inTable = false;
          remainderLines.push(line);
        }
      } else {
        remainderLines.push(line);
      }
    } else if (inTable && !strippedLine) {
      tables.push(currentTable.join("\n"));
      currentTable = [];
      inTable = false;
      remainderLines.push(line);
    } else if (inTable && !isTableLine) {
      tables.push(currentTable.join("\n"));
      currentTable = [];
      inTable = false;
      remainderLines.push(line);
    } else {
      remainderLines.push(line);
    }
  }

  if (currentTable.length > 0) {
    tables.push(currentTable.join("\n"));
  }

  return { remainder: remainderLines.join("\n"), tables };
}

function renderTableToHtml(tableMd: string): string {
  try {
    return md.render(tableMd);
  } catch {
    return tableMd;
  }
}

export const basicStrategy: ChunkStrategy = {
  name: "basic",

  split(content: string, options: ChunkStrategyOptions): ChunkSplitResult {
    if (!content || !content.trim()) {
      return { chunks: [], metas: [] };
    }

    const chunkTokenNum = options.chunkTokenNum;
    const { remainder, tables } = extractTablesAndRemainder(content);

    const processedChunks: string[] = [];
    for (const tableMd of tables) {
      if (tableMd.trim()) {
        processedChunks.push(renderTableToHtml(tableMd));
      }
    }

    const initialSections: string[] = [];
    if (remainder && remainder.trim()) {
      for (const line of remainder.split("\n")) {
        const lineContent = line.trim();
        if (!lineContent) {
          initialSections.push(line);
          continue;
        }
        if (countTokens(line) > 3 * chunkTokenNum) {
          const midPoint = Math.floor(line.length / 2);
          initialSections.push(line.slice(0, midPoint));
          initialSections.push(line.slice(midPoint));
        } else {
          initialSections.push(line);
        }
      }
    }

    const finalTextChunks: string[] = [];
    let currentChunkParts: string[] = [];
    let currentTokenCount = 0;

    for (const sectionText of initialSections) {
      const sectionTokenCount = countTokens(sectionText);

      if (!sectionText.trim() && currentChunkParts.length === 0) {
        continue;
      }

      if (currentTokenCount + sectionTokenCount <= chunkTokenNum) {
        currentChunkParts.push(sectionText);
        currentTokenCount += sectionTokenCount;
      } else {
        if (currentChunkParts.length > 0) {
          finalTextChunks.push(currentChunkParts.join("\n").trim());
        }

        if (sectionTokenCount > chunkTokenNum && sectionTokenCount <= 3 * chunkTokenNum) {
          finalTextChunks.push(sectionText.trim());
          currentChunkParts = [];
          currentTokenCount = 0;
        } else if (sectionTokenCount > 3 * chunkTokenNum) {
          const mid = Math.floor(sectionText.length / 2);
          finalTextChunks.push(sectionText.slice(0, mid).trim());
          finalTextChunks.push(sectionText.slice(mid).trim());
          currentChunkParts = [];
          currentTokenCount = 0;
        } else {
          currentChunkParts = [sectionText];
          currentTokenCount = sectionTokenCount;
        }
      }
    }

    if (currentChunkParts.length > 0) {
      finalTextChunks.push(currentChunkParts.join("\n").trim());
    }

    const allChunks: string[] = [];
    for (const chunk of processedChunks) {
      if (chunk.trim()) allChunks.push(chunk);
    }
    for (const chunk of finalTextChunks) {
      if (chunk.trim()) allChunks.push(chunk);
    }

    return {
      chunks: allChunks,
      metas: allChunks.map(() => ({ headers: {}, metadata: {} })),
    };
  },
};
