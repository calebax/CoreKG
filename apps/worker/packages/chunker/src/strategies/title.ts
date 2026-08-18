import markdownit from "markdown-it";
import { countTokens } from "../tokenizer.js";
import { smartStrategy } from "./smart.js";
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
  return token.children
    .map((c) => (c.type === "text" ? c.content : extractTextFromToken(c)))
    .join("");
}

function extractTextFromTokens(tokens: MarkdownToken[]): string {
  return tokens
    .filter((t) => t.type === "inline")
    .map((t) => extractTextFromToken(t))
    .join("");
}

interface NodeWithHeader {
  type: string;
  level?: number;
  title?: string;
  headers: Record<string, string>;
  isSplitBoundary: boolean;
  content: string;
}

function renderNodeContent(tokens: MarkdownToken[], idx: number): { content: string; endIdx: number } {
  const token = tokens[idx];

  if (token.type === "heading_open") {
    let end = idx + 1;
    while (end < tokens.length && tokens[end].type !== "heading_close") end++;
    const text = extractTextFromTokens(tokens.slice(idx, end + 1));
    const markup = token.markup || "#".repeat(parseInt(token.tag.replace("h", ""), 10));
    return { content: markup + " " + text, endIdx: end };
  }

  if (token.type === "fence") {
    const info = token.info || "";
    return { content: "```" + info + "\n" + token.content + "```", endIdx: idx };
  }

  if (token.type === "hr") {
    return { content: "---", endIdx: idx };
  }

  if (token.type === "table_open") {
    let end = idx;
    while (end < tokens.length && tokens[end].type !== "table_close") end++;
    const tableTokens = tokens.slice(idx, end + 1);
    const lines: string[] = [];
    let cells: string[] = [];
    let headerDone = false;
    for (const t of tableTokens) {
      if (t.type === "tr_open") cells = [];
      else if (t.type === "tr_close") {
        lines.push("| " + cells.join(" | ") + " |");
        if (!headerDone) {
          lines.push("| " + cells.map(() => "---").join(" | ") + " |");
          headerDone = true;
        }
      } else if (t.type === "inline") cells.push(t.content);
    }
    return { content: md.render(lines.join("\n")), endIdx: end };
  }

  if (token.type === "paragraph_open") {
    let end = idx + 1;
    while (end < tokens.length && tokens[end].type !== "paragraph_close") end++;
    const text = extractTextFromTokens(tokens.slice(idx, end + 1));
    return { content: text, endIdx: end };
  }

  if (token.type === "blockquote_open") {
    let end = idx;
    while (end < tokens.length && tokens[end].type !== "blockquote_close") end++;
    const text = extractTextFromTokens(tokens.slice(idx + 1, end));
    return { content: "> " + text, endIdx: end };
  }

  if (token.type === "bullet_list_open" || token.type === "ordered_list_open") {
    const isOrdered = token.type === "ordered_list_open";
    let end = idx;
    const closeType = isOrdered ? "ordered_list_close" : "bullet_list_close";
    while (end < tokens.length && tokens[end].type !== closeType) end++;
    const items: string[] = [];
    let itemIdx = 0;
    for (let j = idx + 1; j < end; j++) {
      if (tokens[j].type === "list_item_open") {
        let itemEnd = j;
        while (itemEnd < end && tokens[itemEnd].type !== "list_item_close") itemEnd++;
        const text = extractTextFromTokens(tokens.slice(j + 1, itemEnd));
        items.push(isOrdered ? (itemIdx + 1) + ". " + text : "- " + text);
        itemIdx++;
        j = itemEnd;
      }
    }
    return { content: items.join("\n"), endIdx: end };
  }

  return { content: extractTextFromToken(token), endIdx: idx };
}

function extractNodesWithHeaderInfo(tokens: MarkdownToken[], headersToSplitOn: number[]): NodeWithHeader[] {
  const nodes: NodeWithHeader[] = [];
  const currentHeaders: Record<string, string> = {};

  for (let i = 0; i < tokens.length; i++) {
    const token = tokens[i];

    if (token.type === "heading_open") {
      const level = parseInt(token.tag.replace("h", ""), 10);
      let end = i + 1;
      while (end < tokens.length && tokens[end].type !== "heading_close") end++;
      const title = extractTextFromTokens(tokens.slice(i, end + 1));
      const markup = token.markup || "#".repeat(level);

      for (const key of Object.keys(currentHeaders)) {
        if (Number(key) >= level) delete currentHeaders[key];
      }
      currentHeaders[String(level)] = title;

      nodes.push({
        type: "heading",
        level,
        title,
        headers: { ...currentHeaders },
        isSplitBoundary: headersToSplitOn.includes(level),
        content: markup + " " + title,
      });
      i = end;
      continue;
    }

    const { content, endIdx } = renderNodeContent(tokens, i);
    if (content.trim()) {
      nodes.push({
        type: token.type.replace("_open", ""),
        headers: { ...currentHeaders },
        isSplitBoundary: false,
        content,
      });
    }
    if (endIdx > i) i = endIdx;
  }

  return nodes;
}

function splitByHeaderLevels(nodes: NodeWithHeader[], headersToSplitOn: number[]): Array<{ headers: Record<string, string>; nodes: NodeWithHeader[] }> {
  const chunks: Array<{ headers: Record<string, string>; nodes: NodeWithHeader[] }> = [];
  let currentChunk: { headers: Record<string, string>; nodes: NodeWithHeader[] } = { headers: {}, nodes: [] };

  let i = 0;
  while (i < nodes.length) {
    const nodeInfo = nodes[i];

    if (nodeInfo.isSplitBoundary) {
      if (nodeInfo.type === "heading") {
        let hasFollowingContent = false;
        for (let j = i + 1; j < nodes.length; j++) {
          if (nodes[j].type === "heading") continue;
          if (nodes[j].content.trim()) {
            hasFollowingContent = true;
            break;
          }
        }
        if (!hasFollowingContent) {
          currentChunk.nodes.push(nodeInfo);
          if (Object.keys(nodeInfo.headers).length > 0) {
            currentChunk.headers = { ...nodeInfo.headers };
          }
          i++;
          continue;
        }
      }

      if (currentChunk.nodes.length > 0 && currentChunk.nodes.some((n) => n.content.trim())) {
        chunks.push(currentChunk);
        currentChunk = { headers: {}, nodes: [] };
      }
    }

    if (Object.keys(nodeInfo.headers).length > 0) {
      currentChunk.headers = { ...nodeInfo.headers };
    }
    currentChunk.nodes.push(nodeInfo);
    i++;
  }

  if (currentChunk.nodes.length > 0 && currentChunk.nodes.some((n) => n.content.trim())) {
    chunks.push(currentChunk);
  }

  return chunks;
}

function renderHeaderChunk(chunkInfo: { headers: Record<string, string>; nodes: NodeWithHeader[] }): string {
  const parts: string[] = [];
  const hasHeader = chunkInfo.nodes.some((n) => n.type === "heading");

  if (!hasHeader && Object.keys(chunkInfo.headers).length > 0) {
    const levels = Object.keys(chunkInfo.headers).map(Number);
    const maxLevel = Math.max(...levels);
    const contextHeader = "#".repeat(maxLevel) + " " + chunkInfo.headers[String(maxLevel)];
    if (contextHeader) parts.push(contextHeader);
  }

  for (const node of chunkInfo.nodes) {
    if (node.content.trim()) parts.push(node.content);
  }

  return parts.join("\n\n").trim();
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
      missingLines.push("#".repeat(level) + " " + headers[String(level)]);
    }
  }

  if (missingLines.length > 0) {
    return missingLines.join("\n") + "\n\n" + chunkContent;
  }
  return chunkContent;
}

export const titleStrategy: ChunkStrategy = {
  name: "title",

  split(content: string, options: ChunkStrategyOptions): ChunkSplitResult {
    if (!content || !content.trim()) {
      return { chunks: [], metas: [] };
    }

    const splitLevel = options.splitLevel;
    const enableHeadingInContent = options.enableHeadingInContent;

    try {
      const tokens = md.parse(content, {});

      let currentLevel = splitLevel;
      let chunks: Array<{ headers: Record<string, string>; nodes: NodeWithHeader[] }> | null = null;

      while (currentLevel >= 1) {
        const headersToSplitOn = [currentLevel];
        const nodesWithHeaders = extractNodesWithHeaderInfo(tokens, headersToSplitOn);
        chunks = splitByHeaderLevels(nodesWithHeaders, headersToSplitOn);

        if (chunks.length > 1 || currentLevel === 1) break;
        currentLevel--;
      }

      const resultChunks: string[] = [];
      const resultMetas: ChunkMeta[] = [];

      for (const chunkInfo of chunks!) {
        let chunkContent = renderHeaderChunk(chunkInfo);
        if (!chunkContent.trim()) continue;

        if (enableHeadingInContent && Object.keys(chunkInfo.headers).length > 0) {
          chunkContent = addMissingParentHeadings(chunkContent, chunkInfo.headers);
        }

        resultChunks.push(chunkContent);
        const levels = Object.keys(chunkInfo.headers).map(Number);
        resultMetas.push({
          headers: chunkInfo.headers,
          metadata: { level: levels.length > 0 ? Math.max(...levels) : 0 },
        });
      }

      return { chunks: resultChunks, metas: resultMetas };
    } catch {
      return smartStrategy.split(content, options);
    }
  },
};
