import markdownit from "markdown-it";
import { countTokens } from "../tokenizer.js";
import { basicStrategy } from "./basic.js";
import type { ChunkStrategy, ChunkStrategyOptions, ChunkSplitResult, ChunkMeta } from "../strategy.js";

const md = new markdownit({ html: true, breaks: true });
md.enable(["table"]);

interface MarkdownToken {
  type: string;
  tag: string;
  content: string;
  markup: string;
  info: string;
  attrs: Array<[string, string]> | null;
  children: MarkdownToken[] | null;
  attrGet(name: string): string | null;
}

function extractTextFromToken(token: MarkdownToken): string {
  if (token.content) return token.content;
  if (!token.children) return "";

  const parts: string[] = [];
  for (const child of token.children) {
    if (child.type === "text") {
      parts.push(child.content);
    } else if (child.type === "code_inline") {
      parts.push("`" + child.content + "`");
    } else if (child.type === "strong_open" || child.type === "strong_close") {
      parts.push("**");
    } else if (child.type === "em_open" || child.type === "em_close") {
      parts.push("*");
    } else if (child.type === "link_open") {
      const href = child.attrGet("href") || "";
      parts.push("[");
    } else if (child.type === "link_close") {
      parts.push("]()");
    } else {
      parts.push(extractTextFromToken(child));
    }
  }
  return parts.join("");
}

function extractTextFromTokens(tokens: MarkdownToken[]): string {
  const parts: string[] = [];
  for (const tok of tokens) {
    if (tok.type.endsWith("_open") || tok.type.endsWith("_close")) continue;
    if (tok.type === "inline") {
      parts.push(extractTextFromToken(tok));
    }
  }
  return parts.join("");
}

interface ContextEntry {
  level: number;
  title: string;
}

function updateContextStack(stack: ContextEntry[], level: number, title: string): void {
  while (stack.length > 0 && stack[stack.length - 1].level >= level) {
    stack.pop();
  }
  stack.push({ level, title });
}

interface ProcessedNode {
  content: string;
  shouldBreak: boolean;
}

function processNonHeadingNode(tokens: MarkdownToken[], startIdx: number, chunkTokenNum: number): ProcessedNode {
  const token = tokens[startIdx];

  if (token.type === "fence") {
    const info = token.info || "";
    const content = "```" + info + "\n" + token.content + "```";
    return { content, shouldBreak: false };
  }

  if (token.type === "hr") {
    return { content: "---", shouldBreak: true };
  }

  if (token.type === "table_open") {
    let endIdx = startIdx;
    while (endIdx < tokens.length && tokens[endIdx].type !== "table_close") {
      endIdx++;
    }
    const tableTokens = tokens.slice(startIdx, endIdx + 1);
    const mdLines: string[] = [];
    let currentCells: string[] = [];
    let headerDone = false;

    for (const t of tableTokens) {
      if (t.type === "tr_open") {
        currentCells = [];
      } else if (t.type === "tr_close") {
        mdLines.push("| " + currentCells.join(" | ") + " |");
        if (!headerDone) {
          mdLines.push("| " + currentCells.map(() => "---").join(" | ") + " |");
          headerDone = true;
        }
      } else if (t.type === "inline") {
        currentCells.push(t.content);
      }
    }

    const tableRendered = md.render(mdLines.join("\n"));
    const tokensCount = countTokens(tableRendered);
    return { content: tableRendered, shouldBreak: tokensCount > chunkTokenNum };
  }

  if (token.type === "blockquote_open") {
    let endIdx = startIdx;
    while (endIdx < tokens.length && tokens[endIdx].type !== "blockquote_close") {
      endIdx++;
    }
    const innerTokens = tokens.slice(startIdx + 1, endIdx);
    const content = "> " + extractTextFromTokens(innerTokens);
    return { content, shouldBreak: false };
  }

  if (token.type === "bullet_list_open" || token.type === "ordered_list_open") {
    const isOrdered = token.type === "ordered_list_open";
    let endIdx = startIdx;
    while (endIdx < tokens.length && tokens[endIdx].type !== token.type.replace("_open", "_close")) {
      endIdx++;
    }
    const items: string[] = [];
    let itemIdx = 0;
    for (let i = startIdx + 1; i < endIdx; i++) {
      if (tokens[i].type === "list_item_open") {
        const innerStart = i + 1;
        let innerEnd = i;
        while (innerEnd < endIdx && tokens[innerEnd].type !== "list_item_close") {
          innerEnd++;
        }
        const text = extractTextFromTokens(tokens.slice(innerStart, innerEnd));
        if (isOrdered) {
          items.push((itemIdx + 1) + ". " + text);
        } else {
          items.push("- " + text);
        }
        itemIdx++;
        i = innerEnd;
      }
    }
    return { content: items.join("\n"), shouldBreak: false };
  }

  if (token.type === "paragraph_open") {
    let endIdx = startIdx + 1;
    while (endIdx < tokens.length && tokens[endIdx].type !== "paragraph_close") {
      endIdx++;
    }
    const content = extractTextFromTokens(tokens.slice(startIdx, endIdx + 1));
    return { content, shouldBreak: false };
  }

  return { content: extractTextFromToken(token), shouldBreak: false };
}

function reconstructBlock(tokens: MarkdownToken[], start: number, end: number): string {
  let result = "";
  for (let i = start; i <= end; i++) {
    const t = tokens[i];
    if (t.type === "inline") {
      result += t.content + "\n";
    }
  }
  return result;
}

function finalizeAstChunk(
  chunkParts: string[],
  contextStack: ContextEntry[],
  enableHeadingInContent: boolean,
): { content: string; headingMetadata: { headers: Record<string, string>; level: number } } {
  let chunkContent = chunkParts.join("\n\n").trim();
  const headers: Record<string, string> = {};
  for (const item of contextStack) {
    headers[String(item.level)] = item.title;
  }

  if (enableHeadingInContent && Object.keys(headers).length > 0) {
    chunkContent = addMissingParentHeadings(chunkContent, headers);
  }

  return {
    content: chunkContent,
    headingMetadata: {
      headers,
      level: Object.keys(headers).length > 0 ? Math.max(...Object.keys(headers).map(Number)) : 0,
    },
  };
}

function addMissingParentHeadings(chunkContent: string, headers: Record<string, string>): string {
  const existingHeadings = new Set<string>();
  for (const line of chunkContent.split("\n")) {
    const trimmed = line.trim();
    if (trimmed.startsWith("#")) {
      const level = trimmed.length - trimmed.replace(/^#+/, "").length;
      if (level > 0 && level <= 6) {
        const headingText = trimmed.replace(/^#+\s*/, "");
        existingHeadings.add(level + ":" + headingText);
      }
    }
  }

  const missingLines: string[] = [];
  const sortedLevels = Object.keys(headers).map(Number).sort((a, b) => a - b);
  for (const level of sortedLevels) {
    const headingKey = level + ":" + headers[String(level)];
    if (!existingHeadings.has(headingKey)) {
      const prefix = "#".repeat(level);
      missingLines.push(prefix + " " + headers[String(level)]);
    }
  }

  if (missingLines.length > 0) {
    chunkContent = missingLines.join("\n") + "\n\n" + chunkContent;
  }

  return chunkContent;
}

export const smartStrategy: ChunkStrategy = {
  name: "smart",

  split(content: string, options: ChunkStrategyOptions): ChunkSplitResult {
    if (!content || !content.trim()) {
      return { chunks: [], metas: [] };
    }

    const chunkTokenNum = options.chunkTokenNum;
    const minChunkTokens = options.minChunkTokens;
    const enableHeadingInContent = options.enableHeadingInContent;

    try {
      const tokens = md.parse(content, {});

      const chunks: Array<{ content: string; headingMetadata: { headers: Record<string, string>; level: number } }> = [];
      let currentChunk: string[] = [];
      let currentTokens = 0;
      const contextStack: ContextEntry[] = [];

      let i = 0;
      while (i < tokens.length) {
        const token = tokens[i];

        if (token.type === "heading_open") {
          if (currentChunk.length > 0 && currentTokens >= minChunkTokens) {
            const chunkContent = finalizeAstChunk(currentChunk, contextStack, enableHeadingInContent);
            if (chunkContent.content.trim()) {
              chunks.push(chunkContent);
            }
            currentChunk = [];
            currentTokens = 0;
          }

          const level = parseInt(token.tag.replace("h", ""), 10);
          let endIdx = i + 1;
          while (endIdx < tokens.length && tokens[endIdx].type !== "heading_close") {
            endIdx++;
          }
          const titleText = extractTextFromTokens(tokens.slice(i, endIdx + 1));
          updateContextStack(contextStack, level, titleText);

          const markup = token.markup || "#".repeat(level);
          const chunkData = markup + " " + titleText;
          currentChunk.push(chunkData);
          currentTokens = countTokens(chunkData);
          i = endIdx + 1;
          continue;
        }

        if (token.type === "paragraph_open" || token.type === "fence" || token.type === "hr" ||
            token.type === "table_open" || token.type === "blockquote_open" ||
            token.type === "bullet_list_open" || token.type === "ordered_list_open") {

          let endIdx = i;
          let closeType: string | undefined;
          if (token.type === "paragraph_open") closeType = "paragraph_close";
          else if (token.type === "table_open") closeType = "table_close";
          else if (token.type === "blockquote_open") closeType = "blockquote_close";
          else if (token.type === "bullet_list_open") closeType = "bullet_list_close";
          else if (token.type === "ordered_list_open") closeType = "ordered_list_close";

          if (closeType) {
            while (endIdx < tokens.length && tokens[endIdx].type !== closeType) {
              endIdx++;
            }
          }

          const processed = processNonHeadingNode(tokens, i, chunkTokenNum);

          if (processed.content) {
            const chunkTokens = countTokens(processed.content);

            if (currentTokens + chunkTokens > chunkTokenNum && currentChunk.length > 0 && currentTokens >= minChunkTokens) {
              const chunkContent = finalizeAstChunk(currentChunk, contextStack, enableHeadingInContent);
              if (chunkContent.content.trim()) {
                chunks.push(chunkContent);
              }
              currentChunk = [];
              currentTokens = 0;
            }

            currentChunk.push(processed.content);
            currentTokens += chunkTokens;
          }

          if (closeType) {
            i = endIdx + 1;
          } else {
            i++;
          }
          continue;
        }

        i++;
      }

      if (currentChunk.length > 0) {
        const chunkContent = finalizeAstChunk(currentChunk, contextStack, enableHeadingInContent);
        if (chunkContent.content.trim()) {
          chunks.push(chunkContent);
        }
      }

      const resultChunks: string[] = [];
      const resultMetas: ChunkMeta[] = [];
      for (const chunk of chunks) {
        if (chunk.content.trim()) {
          resultChunks.push(chunk.content);
          resultMetas.push({ headers: chunk.headingMetadata.headers, metadata: { level: chunk.headingMetadata.level } });
        }
      }

      return { chunks: resultChunks, metas: resultMetas };
    } catch {
      return basicStrategy.split(content, options);
    }
  },
};
