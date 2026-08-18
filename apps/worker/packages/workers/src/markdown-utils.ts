import { v4 as uuidv4 } from "uuid";

export function extractMarkdownTitles(mdContent: string): string[] {
  const titles: string[] = [];
  const lines = mdContent.split("\n");

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const hashMatch = line.match(/^#{1,6}\s+(.+)$/);
    if (hashMatch) {
      titles.push(hashMatch[1].trim());
      continue;
    }

    if (i > 0 && /^[=-]{3,}$/.test(line.trim())) {
      const title = lines[i - 1].trim();
      if (title.length > 0) {
        titles.push(title);
      }
    }
  }

  return titles;
}

export function extractCodeBlock(codeType: string, text: string): string {
  const re = new RegExp("```" + codeType + "\\s*([\\s\\S]*?)\\s*```");
  const match = text.match(re);
  return match ? match[1] : text;
}

interface MindMapNode {
  id?: string;
  uuid?: string;
  children?: MindMapNode[];
}

export function processEmbeddedUuid(jsonStr: string): string {
  const node: MindMapNode = JSON.parse(jsonStr);
  assignUuids(node);
  return JSON.stringify(node);
}

function assignUuids(node: MindMapNode): void {
  node.uuid = uuidv4();
  if (node.children) {
    for (const child of node.children) {
      assignUuids(child);
    }
  }
}
